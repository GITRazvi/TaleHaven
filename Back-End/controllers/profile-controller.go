package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/auth"
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	maxProfilePhotoBytes   int64 = 5 * 1024 * 1024
	maxThemeImageBytes     int64 = 15 * 1024 * 1024
	maxSiteThemeImages           = 10
	maxProfileCommentRunes       = 1000
)

// PublicProfileResponse intentionally contains only the identity data a
// signed-in reader may view from a chatroom. It must never reuse UserResponse,
// which includes account and personal presentation settings.
type PublicProfileResponse struct {
	ID               int    `json:"id"`
	DisplayName      string `json:"displayName"`
	ProfileImagePath string `json:"profileImagePath"`
	CreatedAt        string `json:"createdAt"`
	CommentsEnabled  bool   `json:"commentsEnabled"`
}

// ProfileController exposes the authenticated user's editable presentation
// profile. Account security and reader preferences remain under /auth and
// /settings respectively.
func ProfileController(router *gin.Engine) {
	routes := router.Group("/profile")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.GET("", getProfile)
		routes.PATCH("", updateProfile)
		routes.GET("/comments-enabled", getProfileCommentsEnabled)
		routes.PATCH("/comments-enabled", updateProfileCommentsEnabled)
		routes.PATCH("/theme", updateProfileTheme)
		routes.PATCH("/site-theme", updateSiteTheme)
		routes.POST("/site-theme-images", uploadSiteThemeImages)
		routes.DELETE("/site-theme-images", removeSiteThemeImage)
		routes.POST("/theme-image", uploadProfileThemeImage)
		routes.DELETE("/theme-image", removeProfileThemeImage)
		routes.POST("/photo", uploadProfilePhoto)
	}
	publicRoutes := router.Group("/profiles")
	publicRoutes.Use(middlewares.AuthMiddleware())
	publicRoutes.GET("/:userID", getPublicProfile)
	publicRoutes.GET("/:userID/comments", listProfileComments)
	publicRoutes.POST("/:userID/comments", createProfileComment)
	publicRoutes.DELETE("/:userID/comments/:commentID", deleteProfileComment)
}

func getProfile(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
}

func getPublicProfile(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil || userID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID"})
		return
	}

	user, err := auth.GetUserByID(userID)
	if errors.Is(err, auth.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	if err != nil {
		utils.Logger.Error("Failed to load public profile", zap.Error(err), zap.Int("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load profile"})
		return
	}

	commentsEnabled, err := database.GetProfileCommentsEnabled(userID)
	if err != nil {
		utils.Logger.Error("Failed to load profile comment setting", zap.Error(err), zap.Int("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": toPublicProfileResponse(user, commentsEnabled)})
}

func toPublicProfileResponse(user *database.User, commentsEnabled bool) PublicProfileResponse {
	displayName := strings.TrimSpace(user.DisplayName)
	if displayName == "" {
		displayName = strings.Split(user.Email, "@")[0]
	}
	return PublicProfileResponse{
		ID:               user.ID,
		DisplayName:      displayName,
		ProfileImagePath: user.ProfileImagePath,
		CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CommentsEnabled:  commentsEnabled,
	}
}

func getProfileCommentsEnabled(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	enabled, err := database.GetProfileCommentsEnabled(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load profile comment setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "enabled": enabled})
}

func updateProfileCommentsEnabled(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	var request struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Enabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile comment setting"})
		return
	}
	if err := database.SetProfileCommentsEnabled(user.ID, *request.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save profile comment setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "enabled": *request.Enabled})
}

func listProfileComments(c *gin.Context) {
	if _, ok := profileUser(c); !ok {
		return
	}
	profileUserID, ok := profileCommentProfileUserID(c)
	if !ok {
		return
	}
	if !profileCommentsAreEnabled(c, profileUserID) {
		return
	}
	comments, err := database.GetProfileComments(profileUserID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load profile comments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": comments})
}

func createProfileComment(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	profileUserID, ok := profileCommentProfileUserID(c)
	if !ok {
		return
	}
	if !profileCommentsAreEnabled(c, profileUserID) {
		return
	}
	var request struct {
		Body string `json:"body"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile comment"})
		return
	}
	body := strings.TrimSpace(request.Body)
	if body == "" || utf8.RuneCountInString(body) > maxProfileCommentRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile comment", "message": "Comments must be between 1 and 1,000 characters."})
		return
	}
	comment, err := database.CreateProfileComment(profileUserID, user.ID, body)
	if errors.Is(err, database.ErrProfileCommentsDisabled) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profile comments are disabled"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to post profile comment"})
		return
	}
	comment.AuthorDisplayName = profileCommentDisplayName(user)
	comment.ProfileImagePath = user.ProfileImagePath
	c.JSON(http.StatusCreated, gin.H{"result": "ok", "comment": comment})
}

func deleteProfileComment(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	profileUserID, ok := profileCommentProfileUserID(c)
	if !ok {
		return
	}
	commentID, err := strconv.ParseInt(c.Param("commentID"), 10, 64)
	if err != nil || commentID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile comment ID"})
		return
	}
	if err := database.DeleteProfileComment(profileUserID, user.ID, commentID); errors.Is(err, database.ErrProfileCommentDenied) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profile comment deletion denied"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete profile comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "message": "Profile comment deleted"})
}

func profileCommentProfileUserID(c *gin.Context) (int, bool) {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil || userID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID"})
		return 0, false
	}
	return userID, true
}

func profileCommentsAreEnabled(c *gin.Context, profileUserID int) bool {
	enabled, err := database.GetProfileCommentsEnabled(profileUserID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load profile comments"})
		return false
	}
	if !enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profile comments are disabled"})
		return false
	}
	return true
}

func profileCommentDisplayName(user *database.User) string {
	if displayName := strings.TrimSpace(user.DisplayName); displayName != "" {
		return displayName
	}
	return strings.Split(user.Email, "@")[0]
}

func updateProfile(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	var request struct {
		DisplayName string `json:"displayName"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile update", "message": "Display name is required."})
		return
	}
	displayName := strings.TrimSpace(request.DisplayName)
	if displayName == "" || len([]rune(displayName)) > 80 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid display name", "message": "Display name must be between 1 and 80 characters."})
		return
	}
	updated, err := auth.UpdateUserDisplayName(user.ID, displayName)
	if err != nil {
		utils.Logger.Error("Failed to update display name", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile update failed", "message": "Unable to save your display name."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(updated)})
}

func updateProfileTheme(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	var request struct {
		DisplayTheme     string `json:"displayTheme"`
		CustomThemeColor string `json:"customThemeColor"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid display theme"})
		return
	}
	theme := strings.ToLower(strings.TrimSpace(request.DisplayTheme))
	if !database.IsDisplayTheme(theme) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid display theme", "message": "Choose one of the available display themes."})
		return
	}
	customThemeColor, valid := database.NormalizeCustomThemeColor(request.CustomThemeColor)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom theme color", "message": "Choose a six-digit hexadecimal color."})
		return
	}
	updated, err := auth.UpdateUserDisplayTheme(user.ID, theme, customThemeColor)
	if err != nil {
		utils.Logger.Error("Failed to update profile theme", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Display theme update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(updated)})
}

func updateSiteTheme(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	var request struct {
		SiteTheme                     string `json:"siteTheme"`
		SiteCustomColor               string `json:"siteCustomColor"`
		SiteThemeImageRotationMode    string `json:"siteThemeImageRotationMode"`
		SiteThemeImageRotationMinutes int    `json:"siteThemeImageRotationMinutes"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site theme"})
		return
	}
	theme := strings.ToLower(strings.TrimSpace(request.SiteTheme))
	if !database.IsDisplayTheme(theme) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site theme", "message": "Choose one of the available site themes."})
		return
	}
	customColor, valid := database.NormalizeCustomThemeColor(request.SiteCustomColor)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom site color", "message": "Choose a six-digit hexadecimal color."})
		return
	}
	rotationMode := strings.ToLower(strings.TrimSpace(request.SiteThemeImageRotationMode))
	if !database.IsSiteThemeRotationMode(rotationMode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site theme image rotation", "message": "Choose static, each refresh, per session, or a timed interval."})
		return
	}
	if rotationMode == database.SiteThemeRotationInterval && !database.IsSiteThemeRotationMinutes(request.SiteThemeImageRotationMinutes) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site theme image rotation interval", "message": "Choose between 1 and 1440 minutes."})
		return
	}
	rotationMinutes := request.SiteThemeImageRotationMinutes
	if rotationMinutes == 0 {
		rotationMinutes = 15
	}
	updated, err := auth.UpdateUserSiteTheme(user.ID, theme, customColor, rotationMode, rotationMinutes)
	if err != nil {
		utils.Logger.Error("Failed to update site theme", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Site theme update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(updated)})
}

func uploadSiteThemeImages(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site theme images", "message": "Choose one or more JPEG, PNG, WebP, or GIF images."})
		return
	}
	fileHeaders := form.File["images"]
	if len(fileHeaders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Site theme images are required", "message": "Choose one or more JPEG, PNG, WebP, or GIF images."})
		return
	}
	if len(user.SiteThemeImagePaths)+len(fileHeaders) > maxSiteThemeImages {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many site theme images", "message": fmt.Sprintf("Keep up to %d site theme images.", maxSiteThemeImages)})
		return
	}
	directory := filepath.Join("images", "site-themes")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		utils.Logger.Error("Failed to create site theme image directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Site theme image upload failed"})
		return
	}

	newPaths := make([]string, 0, len(fileHeaders))
	newFilePaths := make([]string, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		if fileHeader.Size <= 0 || fileHeader.Size > maxThemeImageBytes {
			removeStoredSiteThemeFiles(newFilePaths)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site theme image", "message": "Each image must be 15 MB or smaller."})
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			removeStoredSiteThemeFiles(newFilePaths)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site theme image", "message": "Unable to read a selected image."})
			return
		}
		contents, readErr := io.ReadAll(io.LimitReader(file, maxThemeImageBytes+1))
		_ = file.Close()
		if readErr != nil || int64(len(contents)) > maxThemeImageBytes {
			removeStoredSiteThemeFiles(newFilePaths)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site theme image", "message": "Each image must be 15 MB or smaller."})
			return
		}
		extension := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "image/gif": ".gif"}[http.DetectContentType(contents)]
		if extension == "" {
			removeStoredSiteThemeFiles(newFilePaths)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported site theme image", "message": "Use JPEG, PNG, WebP, or GIF images."})
			return
		}
		fileName := fmt.Sprintf("user-%d-%d-%d%s", user.ID, time.Now().UTC().UnixNano(), len(newPaths), extension)
		filePath := filepath.Join(directory, fileName)
		if err := os.WriteFile(filePath, contents, 0o644); err != nil {
			removeStoredSiteThemeFiles(newFilePaths)
			utils.Logger.Error("Failed to store site theme image", zap.Error(err), zap.Int("user_id", user.ID))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Site theme image upload failed"})
			return
		}
		newFilePaths = append(newFilePaths, filePath)
		newPaths = append(newPaths, "/images/site-themes/"+fileName)
	}

	updated, err := auth.UpdateUserSiteThemeImagePaths(user.ID, append(append([]string{}, user.SiteThemeImagePaths...), newPaths...))
	if err != nil {
		removeStoredSiteThemeFiles(newFilePaths)
		utils.Logger.Error("Failed to save site theme image paths", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Site theme image upload failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(updated)})
}

func removeSiteThemeImage(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	imagePath := strings.TrimSpace(c.Query("path"))
	if imagePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Site theme image path is required"})
		return
	}
	remaining := make([]string, 0, len(user.SiteThemeImagePaths))
	found := false
	for _, knownPath := range user.SiteThemeImagePaths {
		if knownPath == imagePath {
			found = true
			continue
		}
		remaining = append(remaining, knownPath)
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Site theme image not found"})
		return
	}
	updated, err := auth.UpdateUserSiteThemeImagePaths(user.ID, remaining)
	if err != nil {
		utils.Logger.Error("Failed to remove site theme image", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Site theme image removal failed"})
		return
	}
	removeStoredSiteThemeImage(imagePath)
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(updated)})
}

func uploadProfilePhoto(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profile photo is required", "message": "Choose a JPEG, PNG, or WebP image."})
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxProfilePhotoBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile photo", "message": "Profile photos must be 5 MB or smaller."})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile photo", "message": "Unable to read the selected image."})
		return
	}
	defer file.Close()
	contents, err := io.ReadAll(io.LimitReader(file, maxProfilePhotoBytes+1))
	if err != nil || int64(len(contents)) > maxProfilePhotoBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile photo", "message": "Profile photos must be 5 MB or smaller."})
		return
	}

	contentType := http.DetectContentType(contents)
	extension := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}[contentType]
	if extension == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported profile photo", "message": "Use a JPEG, PNG, or WebP image."})
		return
	}

	directory := filepath.Join("images", "profiles")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		utils.Logger.Error("Failed to create profile image directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile photo upload failed"})
		return
	}
	fileName := fmt.Sprintf("user-%d-%d%s", user.ID, time.Now().UTC().UnixNano(), extension)
	filePath := filepath.Join(directory, fileName)
	if err := os.WriteFile(filePath, contents, 0o644); err != nil {
		utils.Logger.Error("Failed to store profile image", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile photo upload failed"})
		return
	}

	imagePath := "/images/profiles/" + fileName
	updated, err := auth.UpdateUserProfileImagePath(user.ID, imagePath)
	if err != nil {
		_ = os.Remove(filePath)
		utils.Logger.Error("Failed to save profile image path", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile photo upload failed"})
		return
	}
	removeStoredProfilePhoto(user.ProfileImagePath)
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(updated)})
}

// uploadProfileThemeImage replaces only the signed-in user's profile
// background. It follows the same file validation as profile photos, while
// keeping it in a separate server-owned directory.
func uploadProfileThemeImage(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme image is required", "message": "Choose a JPEG, PNG, WebP, or GIF image."})
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxThemeImageBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme image", "message": "Theme images must be 15 MB or smaller."})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme image", "message": "Unable to read the selected image."})
		return
	}
	defer file.Close()
	contents, err := io.ReadAll(io.LimitReader(file, maxThemeImageBytes+1))
	if err != nil || int64(len(contents)) > maxThemeImageBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme image", "message": "Theme images must be 15 MB or smaller."})
		return
	}
	extension := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "image/gif": ".gif"}[http.DetectContentType(contents)]
	if extension == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported theme image", "message": "Use a JPEG, PNG, WebP, or GIF image."})
		return
	}
	directory := filepath.Join("images", "profile-themes")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		utils.Logger.Error("Failed to create profile theme image directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Theme image upload failed"})
		return
	}
	fileName := fmt.Sprintf("user-%d-%d%s", user.ID, time.Now().UTC().UnixNano(), extension)
	filePath := filepath.Join(directory, fileName)
	if err := os.WriteFile(filePath, contents, 0o644); err != nil {
		utils.Logger.Error("Failed to store profile theme image", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Theme image upload failed"})
		return
	}
	imagePath := "/images/profile-themes/" + fileName
	updated, err := auth.UpdateUserThemeImagePath(user.ID, imagePath)
	if err != nil {
		_ = os.Remove(filePath)
		utils.Logger.Error("Failed to save profile theme image", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Theme image upload failed"})
		return
	}
	removeStoredProfileThemeImage(user.ThemeImagePath)
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(updated)})
}

func removeProfileThemeImage(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	updated, err := auth.UpdateUserThemeImagePath(user.ID, "")
	if err != nil {
		utils.Logger.Error("Failed to clear profile theme image", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Theme image removal failed"})
		return
	}
	removeStoredProfileThemeImage(user.ThemeImagePath)
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(updated)})
}

func profileUser(c *gin.Context) (*database.User, bool) {
	value, ok := c.Get("user")
	user, valid := value.(*database.User)
	if !ok || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, false
	}
	return user, true
}

func removeStoredProfilePhoto(imagePath string) {
	const publicPrefix = "/images/profiles/"
	if !strings.HasPrefix(imagePath, publicPrefix) {
		return
	}
	fileName := strings.TrimPrefix(imagePath, publicPrefix)
	if fileName == "" || filepath.Base(fileName) != fileName {
		return
	}
	if err := os.Remove(filepath.Join("images", "profiles", fileName)); err != nil && !os.IsNotExist(err) {
		utils.Logger.Warn("Failed to remove replaced profile image", zap.Error(err))
	}
}

func removeStoredProfileThemeImage(imagePath string) {
	const publicPrefix = "/images/profile-themes/"
	if !strings.HasPrefix(imagePath, publicPrefix) {
		return
	}
	fileName := strings.TrimPrefix(imagePath, publicPrefix)
	if fileName == "" || filepath.Base(fileName) != fileName {
		return
	}
	if err := os.Remove(filepath.Join("images", "profile-themes", fileName)); err != nil && !os.IsNotExist(err) {
		utils.Logger.Warn("Failed to remove replaced profile theme image", zap.Error(err))
	}
}

func removeStoredSiteThemeImage(imagePath string) {
	const publicPrefix = "/images/site-themes/"
	if !strings.HasPrefix(imagePath, publicPrefix) {
		return
	}
	fileName := strings.TrimPrefix(imagePath, publicPrefix)
	if fileName == "" || filepath.Base(fileName) != fileName {
		return
	}
	if err := os.Remove(filepath.Join("images", "site-themes", fileName)); err != nil && !os.IsNotExist(err) {
		utils.Logger.Warn("Failed to remove site theme image", zap.Error(err))
	}
}

func removeStoredSiteThemeFiles(filePaths []string) {
	for _, filePath := range filePaths {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			utils.Logger.Warn("Failed to remove unpersisted site theme image", zap.Error(err))
		}
	}
}
