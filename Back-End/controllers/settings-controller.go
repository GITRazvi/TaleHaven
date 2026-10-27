package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/database"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var translationLanguagePattern = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})?$`)

// SettingsController exposes authenticated reader preferences.
func SettingsController(router *gin.Engine) {
	routes := router.Group("/settings")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.GET("/reading", getReadingSettings)
		routes.PATCH("/reading", updateReadingSettings)
		routes.GET("/reading/manga/:mangaId/scanlation-group", getPreferredScanlationGroup)
		routes.PATCH("/reading/manga/:mangaId/scanlation-group", updatePreferredScanlationGroup)
	}
}

type readingSettingsRequest struct {
	PreferredTranslationLanguage string `json:"preferredTranslationLanguage"`
	PreferredReadingMode         string `json:"preferredReadingMode"`
}

type scanlationGroupPreferenceRequest struct {
	Language        string `json:"language" binding:"required"`
	ScanlationGroup string `json:"scanlationGroup" binding:"required"`
}

func getReadingSettings(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	language, err := database.GetPreferredTranslationLanguage(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load reading settings"})
		return
	}
	mode, err := database.GetPreferredReadingMode(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load reading settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "preferredTranslationLanguage": language, "preferredReadingMode": mode})
}

func updateReadingSettings(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	var request readingSettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reading settings"})
		return
	}
	language := strings.ToLower(strings.TrimSpace(request.PreferredTranslationLanguage))
	mode := strings.ToLower(strings.TrimSpace(request.PreferredReadingMode))
	if language == "" {
		language, err = database.GetPreferredTranslationLanguage(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load reading settings"})
			return
		}
	}
	if mode == "" {
		mode, err = database.GetPreferredReadingMode(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load reading settings"})
			return
		}
	}
	if language != "all" && !translationLanguagePattern.MatchString(language) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid translation language"})
		return
	}
	if mode != "vertical" && mode != "ltr" && mode != "rtl" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reading mode"})
		return
	}
	if err := database.SetPreferredTranslationLanguage(userID, language); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update reading settings"})
		return
	}
	if err := database.SetPreferredReadingMode(userID, mode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update reading settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "preferredTranslationLanguage": language, "preferredReadingMode": mode})
}

func normalizeScanlationGroupPreference(c *gin.Context, language, group string) (string, string, bool) {
	mangaID := strings.TrimSpace(c.Param("mangaId"))
	language = strings.ToLower(strings.TrimSpace(language))
	group = strings.TrimSpace(group)
	if mangaID == "" || len(mangaID) > 64 || language == "" || (language != "all" && !translationLanguagePattern.MatchString(language)) || group == "" || len(group) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scanlation-group preference"})
		return "", "", false
	}
	if language == "all" && group != "all" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Choose one language before selecting a scanlation group"})
		return "", "", false
	}
	return language, group, true
}

func getPreferredScanlationGroup(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	language, _, valid := normalizeScanlationGroupPreference(c, c.DefaultQuery("language", "all"), "all")
	if !valid {
		return
	}
	group, err := database.GetPreferredScanlationGroup(userID, c.Param("mangaId"), language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load scanlation-group preference"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "mangaId": c.Param("mangaId"), "language": language, "scanlationGroup": group})
}

func updatePreferredScanlationGroup(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	var request scanlationGroupPreferenceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scanlation-group preference"})
		return
	}
	language, group, valid := normalizeScanlationGroupPreference(c, request.Language, request.ScanlationGroup)
	if !valid {
		return
	}
	if err := database.SetPreferredScanlationGroup(userID, c.Param("mangaId"), language, group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update scanlation-group preference"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "mangaId": c.Param("mangaId"), "language": language, "scanlationGroup": group})
}
