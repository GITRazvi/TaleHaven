package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/auth"
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxProfilePhotoBytes int64 = 5 * 1024 * 1024

// ProfileController exposes the authenticated user's editable presentation
// profile. Account security and reader preferences remain under /auth and
// /settings respectively.
func ProfileController(router *gin.Engine) {
	routes := router.Group("/profile")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.GET("", getProfile)
		routes.PATCH("", updateProfile)
		routes.PATCH("/theme", updateProfileTheme)
		routes.POST("/theme-image", uploadProfileThemeImage)
		routes.DELETE("/theme-image", removeProfileThemeImage)
		routes.POST("/photo", uploadProfilePhoto)
	}
}

func getProfile(c *gin.Context) {
	user, ok := profileUser(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme image is required", "message": "Choose a JPEG, PNG, or WebP image."})
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxProfilePhotoBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme image", "message": "Theme images must be 5 MB or smaller."})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme image", "message": "Unable to read the selected image."})
		return
	}
	defer file.Close()
	contents, err := io.ReadAll(io.LimitReader(file, maxProfilePhotoBytes+1))
	if err != nil || int64(len(contents)) > maxProfilePhotoBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme image", "message": "Theme images must be 5 MB or smaller."})
		return
	}
	extension := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[http.DetectContentType(contents)]
	if extension == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported theme image", "message": "Use a JPEG, PNG, or WebP image."})
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
