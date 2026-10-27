package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/database"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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
)

const (
	defaultChatMessageLimit       = 100
	maxChatMessageLimit           = 200
	maxChatImageBytes       int64 = 5 * 1024 * 1024
)

// ChatController exposes signed-in public and invite-only room workflows.
// Invite-only rooms use bearer-style random invite links, never passwords.
func ChatController(router *gin.Engine) {
	routes := router.Group("/chat")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.GET("/rooms", listChatRooms)
		routes.POST("/rooms", createChatRoom)
		routes.POST("/rooms/:roomID/join", joinPublicChatRoom)
		routes.PATCH("/rooms/:roomID/visibility", updateChatRoomVisibility)
		routes.PATCH("/rooms/:roomID/theme", updateChatRoomTheme)
		routes.POST("/rooms/:roomID/theme-image", uploadChatRoomThemeImage)
		routes.DELETE("/rooms/:roomID/theme-image", removeChatRoomThemeImage)
		routes.POST("/rooms/:roomID/image", uploadChatRoomImage)
		routes.GET("/rooms/:roomID", getChatRoom)
		routes.GET("/rooms/:roomID/messages", getChatMessages)
		routes.POST("/rooms/:roomID/messages", postChatMessage)
		routes.POST("/rooms/:roomID/messages/image", postChatImageMessage)
		routes.POST("/rooms/:roomID/invites", createChatRoomInvite)
		routes.POST("/invites/:token/join", joinChatRoomInvite)
	}
}

type createChatRoomRequest struct {
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

type postChatMessageRequest struct {
	Body string `json:"body"`
}

type updateChatRoomVisibilityRequest struct {
	Visibility string `json:"visibility"`
}

type updateChatRoomThemeRequest struct {
	Theme            string `json:"theme"`
	CustomThemeColor string `json:"customThemeColor"`
}

func listChatRooms(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	rooms, err := database.ListChatRoomsForUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load chat rooms"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": rooms})
}

func createChatRoom(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	var request createChatRoomRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat room", "message": "Name and visibility are required."})
		return
	}
	name, visibility, valid := normalizeChatRoom(request.Name, request.Visibility)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat room", "message": "Use a room name between 2 and 80 characters and choose public or private."})
		return
	}
	room, err := database.CreateChatRoom(userID, name, visibility)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create chat room"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"result": "ok", "room": room})
}

func joinPublicChatRoom(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	joined, err := database.JoinPublicChatRoom(roomID, userID)
	if errors.Is(err, database.ErrChatRoomNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat room not found"})
		return
	}
	if errors.Is(err, database.ErrChatInviteRequired) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invitation required", "message": "This room can only be joined through an invite link."})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to join chat room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "joined": joined})
}

func updateChatRoomVisibility(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	var request updateChatRoomVisibilityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room access setting"})
		return
	}
	visibility, valid := normalizeChatVisibility(request.Visibility)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room access setting", "message": "Choose public or private."})
		return
	}
	if err := database.UpdateChatRoomVisibility(roomID, userID, visibility); errors.Is(err, database.ErrChatRoomAccess) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change access"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update room access"})
		return
	}
	room, err := database.GetChatRoomForMember(roomID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load updated chat room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "room": room})
}

func updateChatRoomTheme(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	var request updateChatRoomThemeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chatroom theme"})
		return
	}
	theme := strings.ToLower(strings.TrimSpace(request.Theme))
	if !database.IsDisplayTheme(theme) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chatroom theme", "message": "Choose one of the available themes."})
		return
	}
	customThemeColor, valid := database.NormalizeCustomThemeColor(request.CustomThemeColor)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom theme color", "message": "Choose a six-digit hexadecimal color."})
		return
	}
	if err := database.UpdateChatRoomTheme(roomID, userID, theme, customThemeColor); errors.Is(err, database.ErrChatRoomAccess) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change the chatroom theme"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update chatroom theme"})
		return
	}
	room, err := database.GetChatRoomForMember(roomID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load updated chat room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "room": room})
}

func getChatRoom(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	room, err := database.GetChatRoomForMember(roomID, userID)
	if errors.Is(err, database.ErrChatRoomAccess) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chat room access denied", "message": "Join the public room or use an invite link first."})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load chat room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "room": room})
}

func getChatMessages(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	limit, valid := chatMessageLimit(c)
	if !valid {
		return
	}
	messages, err := database.GetChatMessages(roomID, userID, limit)
	if errors.Is(err, database.ErrChatRoomAccess) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chat room access denied"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load messages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": messages, "limit": limit})
}

func postChatMessage(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	var request postChatMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message", "message": "A message body is required."})
		return
	}
	body := strings.TrimSpace(request.Body)
	if body == "" || utf8.RuneCountInString(body) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message", "message": "Messages must be between 1 and 2,000 characters."})
		return
	}
	message, err := database.CreateChatMessage(roomID, userID, body, "")
	if errors.Is(err, database.ErrChatRoomAccess) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chat room access denied"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to send message"})
		return
	}
	message.AuthorDisplayName = chatDisplayName(c)
	c.JSON(http.StatusCreated, gin.H{"result": "ok", "message": message})
}

// uploadChatRoomImage lets only the room owner replace the room cover shown in
// chatroom lists and the room header. The image path is generated server-side.
func uploadChatRoomImage(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	room, err := database.GetChatRoomForMember(roomID, userID)
	if errors.Is(err, database.ErrChatRoomAccess) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change the room image"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load chat room"})
		return
	}
	if !room.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change the room image"})
		return
	}
	contents, extension, err := chatImageContents(c, "image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room image", "message": err.Error()})
		return
	}
	imagePath, filePath, err := storeChatImage(contents, extension, "rooms", fmt.Sprintf("room-%d", roomID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Room image upload failed"})
		return
	}
	if err := database.UpdateChatRoomImagePath(roomID, userID, imagePath); err != nil {
		_ = os.Remove(filePath)
		if errors.Is(err, database.ErrChatRoomAccess) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change the room image"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Room image upload failed"})
		return
	}
	removeStoredChatRoomImage(room.ImagePath)
	updatedRoom, err := database.GetChatRoomForMember(roomID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load updated chat room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "room": updatedRoom})
}

// uploadChatRoomThemeImage stores an owner-selected background image separately
// from the small room cover used in chatroom cards and headers.
func uploadChatRoomThemeImage(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	room, err := database.GetChatRoomForMember(roomID, userID)
	if errors.Is(err, database.ErrChatRoomAccess) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change the theme image"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load chat room"})
		return
	}
	if !room.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change the theme image"})
		return
	}
	contents, extension, err := chatImageContents(c, "image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme image", "message": err.Error()})
		return
	}
	imagePath, filePath, err := storeChatImage(contents, extension, "theme-backgrounds", fmt.Sprintf("room-%d", roomID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Theme image upload failed"})
		return
	}
	if err := database.UpdateChatRoomThemeImagePath(roomID, userID, imagePath); err != nil {
		_ = os.Remove(filePath)
		if errors.Is(err, database.ErrChatRoomAccess) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change the theme image"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Theme image upload failed"})
		return
	}
	removeStoredChatRoomThemeImage(room.ThemeImagePath)
	updatedRoom, err := database.GetChatRoomForMember(roomID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load updated chat room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "room": updatedRoom})
}

func removeChatRoomThemeImage(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	room, err := database.GetChatRoomForMember(roomID, userID)
	if errors.Is(err, database.ErrChatRoomAccess) || (err == nil && !room.IsOwner) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can remove the theme image"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load chat room"})
		return
	}
	if err := database.UpdateChatRoomThemeImagePath(roomID, userID, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Theme image removal failed"})
		return
	}
	removeStoredChatRoomThemeImage(room.ThemeImagePath)
	updatedRoom, err := database.GetChatRoomForMember(roomID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load updated chat room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "room": updatedRoom})
}

// postChatImageMessage accepts an optional text caption and one image from a
// current room member. Text-only messages continue to use the JSON endpoint.
func postChatImageMessage(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	body := strings.TrimSpace(c.PostForm("body"))
	if utf8.RuneCountInString(body) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message", "message": "Messages must be 2,000 characters or shorter."})
		return
	}
	contents, extension, err := chatImageContents(c, "image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message image", "message": err.Error()})
		return
	}
	imagePath, filePath, err := storeChatImage(contents, extension, "messages", fmt.Sprintf("message-%d-%d", roomID, userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Message image upload failed"})
		return
	}
	message, err := database.CreateChatMessage(roomID, userID, body, imagePath)
	if err != nil {
		_ = os.Remove(filePath)
		if errors.Is(err, database.ErrChatRoomAccess) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Chat room access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to send message"})
		return
	}
	message.AuthorDisplayName = chatDisplayName(c)
	c.JSON(http.StatusCreated, gin.H{"result": "ok", "message": message})
}

func createChatRoomInvite(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	roomID, valid := chatRoomID(c)
	if !valid {
		return
	}
	token, tokenHash, err := newChatInviteToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create invitation"})
		return
	}
	if err := database.CreateChatRoomInvite(roomID, userID, tokenHash); errors.Is(err, database.ErrChatRoomAccess) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can create invitations"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create invitation"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"result": "ok", "inviteToken": token})
}

func joinChatRoomInvite(c *gin.Context) {
	userID, ok := chatUserID(c)
	if !ok {
		return
	}
	token := strings.TrimSpace(c.Param("token"))
	if len(token) < 20 || len(token) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation"})
		return
	}
	roomID, joined, err := database.JoinChatRoomWithInvite(chatInviteTokenHash(token), userID)
	if errors.Is(err, database.ErrChatInviteInvalid) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found", "message": "This invite link is invalid."})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to join chat room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "roomId": roomID, "joined": joined})
}

func chatUserID(c *gin.Context) (int, bool) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return 0, false
	}
	return userID, true
}

func chatDisplayName(c *gin.Context) string {
	if user, exists := c.Get("user"); exists {
		if safeUser, ok := user.(*database.User); ok {
			if name := strings.TrimSpace(safeUser.DisplayName); name != "" {
				return name
			}
			return strings.Split(safeUser.Email, "@")[0]
		}
	}
	return "Reader"
}

func chatRoomID(c *gin.Context) (int64, bool) {
	roomID, err := strconv.ParseInt(c.Param("roomID"), 10, 64)
	if err != nil || roomID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat room"})
		return 0, false
	}
	return roomID, true
}

func chatMessageLimit(c *gin.Context) (int, bool) {
	limit := defaultChatMessageLimit
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxChatMessageLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message limit"})
			return 0, false
		}
		limit = parsed
	}
	return limit, true
}

func normalizeChatRoom(name, visibility string) (string, string, bool) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 80 {
		return "", "", false
	}
	visibility, valid := normalizeChatVisibility(visibility)
	if !valid {
		return "", "", false
	}
	return name, visibility, true
}

func normalizeChatVisibility(visibility string) (string, bool) {
	visibility = strings.ToLower(strings.TrimSpace(visibility))
	if visibility != database.ChatRoomPublic && visibility != database.ChatRoomInviteOnly {
		return "", false
	}
	return visibility, true
}

func chatImageContents(c *gin.Context, field string) ([]byte, string, error) {
	fileHeader, err := c.FormFile(field)
	if err != nil {
		return nil, "", fmt.Errorf("choose a JPEG, PNG, or WebP image")
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxChatImageBytes {
		return nil, "", fmt.Errorf("images must be 5 MB or smaller")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmt.Errorf("unable to read the selected image")
	}
	defer file.Close()
	contents, err := io.ReadAll(io.LimitReader(file, maxChatImageBytes+1))
	if err != nil || int64(len(contents)) > maxChatImageBytes {
		return nil, "", fmt.Errorf("images must be 5 MB or smaller")
	}
	extension := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}[http.DetectContentType(contents)]
	if extension == "" {
		return nil, "", fmt.Errorf("use a JPEG, PNG, or WebP image")
	}
	return contents, extension, nil
}

func storeChatImage(contents []byte, extension, category, prefix string) (string, string, error) {
	directory := filepath.Join("images", "chat", category)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", "", fmt.Errorf("create chat image directory: %w", err)
	}
	fileName := fmt.Sprintf("%s-%d%s", prefix, time.Now().UTC().UnixNano(), extension)
	filePath := filepath.Join(directory, fileName)
	if err := os.WriteFile(filePath, contents, 0o644); err != nil {
		return "", "", fmt.Errorf("store chat image: %w", err)
	}
	return "/images/chat/" + category + "/" + fileName, filePath, nil
}

func removeStoredChatRoomImage(imagePath string) {
	const publicPrefix = "/images/chat/rooms/"
	if !strings.HasPrefix(imagePath, publicPrefix) {
		return
	}
	fileName := strings.TrimPrefix(imagePath, publicPrefix)
	if fileName == "" || filepath.Base(fileName) != fileName {
		return
	}
	_ = os.Remove(filepath.Join("images", "chat", "rooms", fileName))
}

func removeStoredChatRoomThemeImage(imagePath string) {
	const publicPrefix = "/images/chat/theme-backgrounds/"
	if !strings.HasPrefix(imagePath, publicPrefix) {
		return
	}
	fileName := strings.TrimPrefix(imagePath, publicPrefix)
	if fileName == "" || filepath.Base(fileName) != fileName {
		return
	}
	_ = os.Remove(filepath.Join("images", "chat", "theme-backgrounds", fileName))
}

func newChatInviteToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return token, chatInviteTokenHash(token), nil
}

func chatInviteTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
