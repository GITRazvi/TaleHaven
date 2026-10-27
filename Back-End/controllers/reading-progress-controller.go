package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/sessions"
	"TaleHaven-Backend/modules/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Helper function to get user ID from context
func getUserID(c *gin.Context) (int, error) {
	// Try to get user_id from context (set by AuthMiddleware or OptionalAuthMiddleware)
	userID, exists := c.Get("user_id")
	if exists {
		if id, ok := userID.(int); ok {
			return id, nil
		}
	}

	return 0, errors.New("authentication required - please login")
}

// ReadingProgressController sets up routes for reading progress endpoints
func ReadingProgressController(router *gin.Engine) {
	progressRoutes := router.Group("/reading-progress")
	// Reading progress is private per user and must always have a server-side
	// authenticated session.
	progressRoutes.Use(middlewares.AuthMiddleware())
	{
		// Update reading progress
		progressRoutes.POST("", updateReadingProgress) // POST /reading-progress

		// Sync reading progress to PostgreSQL (important!)
		progressRoutes.POST("/sync", syncReadingProgress) // POST /reading-progress/sync

		// Get reading progress for a specific chapter
		progressRoutes.GET("/chapter/:chapterId", getChapterProgress) // GET /reading-progress/chapter/{chapterId}

		// Get all progress for a specific manga
		progressRoutes.GET("/manga/:mangaId", getMangaProgressHandler) // GET /reading-progress/manga/{mangaId}

		// Get all reading progress for the session
		progressRoutes.GET("", getAllProgressHandler) // GET /reading-progress

		// Get list of manga IDs that have been read
		progressRoutes.GET("/manga-list", getReadMangaListHandler) // GET /reading-progress/manga-list

		// Delete reading progress for a specific chapter
		progressRoutes.DELETE("/chapter/:chapterId", deleteChapterProgress) // DELETE /reading-progress/chapter/{chapterId}

		// Delete all progress for a specific manga
		progressRoutes.DELETE("/manga/:mangaId", deleteMangaProgressHandler) // DELETE /reading-progress/manga/{mangaId}

		// Delete all reading progress for the session
		progressRoutes.DELETE("", deleteAllProgress) // DELETE /reading-progress
	}

	utils.Logger.Info("Reading progress routes initialized")
}

// UpdateReadingProgressRequest represents the request body for updating reading progress
type UpdateReadingProgressRequest struct {
	MangaID     string `json:"mangaId" binding:"required"`
	ChapterID   string `json:"chapterId" binding:"required"`
	CurrentPage int    `json:"currentPage" binding:"required,min=1"`
	TotalPages  int    `json:"totalPages,omitempty"`
}

// SyncReadingProgressRequest represents the request body for syncing progress to PostgreSQL
type SyncReadingProgressRequest struct {
	ChapterID string `json:"chapterId" binding:"required"`
}

// updateReadingProgress handles updating reading progress for a chapter
func updateReadingProgress(c *gin.Context) {
	// Get user ID
	userID, err := getUserID(c)
	if err != nil {
		utils.Logger.Warn("User ID not found", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	var req UpdateReadingProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warn("Invalid request body for reading progress",
			zap.Int("userId", userID),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "Request body validation failed: " + err.Error(),
		})
		return
	}

	utils.Logger.Info("Updating reading progress",
		zap.Int("userId", userID),
		zap.String("mangaId", req.MangaID),
		zap.String("chapterId", req.ChapterID),
		zap.Int("currentPage", req.CurrentPage))

	err = sessions.SaveReadingProgress(userID, req.MangaID, req.ChapterID, req.CurrentPage, req.TotalPages)
	if err != nil {
		utils.Logger.Error("Failed to save reading progress",
			zap.Error(err),
			zap.Int("userId", userID),
			zap.String("mangaId", req.MangaID),
			zap.String("chapterId", req.ChapterID))

		// Check if it's a Redis-specific error
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "redis") || strings.Contains(errorMsg, "connection") {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service temporarily unavailable",
				"message": "Unable to save progress. Please try again later.",
				"details": "Redis storage error",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save progress",
				"message": "An error occurred while saving your progress",
				"details": errorMsg,
			})
		}
		return
	}

	utils.Logger.Info("Reading progress saved successfully",
		zap.Int("userId", userID),
		zap.String("mangaId", req.MangaID),
		zap.String("chapterId", req.ChapterID),
		zap.Int("currentPage", req.CurrentPage))

	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"message": "Reading progress updated successfully",
	})
}

// syncReadingProgress syncs Redis data to PostgreSQL (persistent storage)
// Call this when: user leaves chapter, closes tab, opens another chapter
func syncReadingProgress(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	var req SyncReadingProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	err = sessions.SyncReadingProgress(userID, req.ChapterID)
	if err != nil {
		utils.Logger.Error("Failed to sync reading progress",
			zap.Error(err),
			zap.Int("userId", userID),
			zap.String("chapterId", req.ChapterID))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Sync failed",
			"message": "Failed to sync progress to database",
		})
		return
	}

	utils.Logger.Info("Reading progress synced",
		zap.Int("userId", userID),
		zap.String("chapterId", req.ChapterID))

	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"message": "Reading progress synced successfully",
	})
}

// getChapterProgress retrieves reading progress for a specific chapter
func getChapterProgress(c *gin.Context) {
	// Get user ID
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	chapterID := c.Param("chapterId")
	mangaID := c.Query("mangaId")

	if chapterID == "" || mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameters",
			"message": "Both chapterId and mangaId query parameter are required",
		})
		return
	}

	progress, err := sessions.GetReadingProgress(userID, mangaID, chapterID)
	if err != nil {
		utils.Logger.Error("Failed to get reading progress",
			zap.Error(err),
			zap.Int("userId", userID),
			zap.String("mangaId", mangaID),
			zap.String("chapterId", chapterID))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve progress",
			"message": "An error occurred while retrieving your progress",
		})
		return
	}

	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"result":  "not_found",
			"message": "No progress found for this chapter",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   progress,
	})
}

// getMangaProgressHandler retrieves all reading progress for a specific manga
func getMangaProgressHandler(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'mangaId' is required",
		})
		return
	}

	progress, err := sessions.GetMangaProgress(userID, mangaID)
	if err != nil {
		utils.Logger.Error("Failed to get manga progress",
			zap.Error(err),
			zap.Int("userId", userID),
			zap.String("mangaId", mangaID))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve progress",
			"message": "An error occurred while retrieving manga progress",
		})
		return
	}

	if progress == nil {
		c.JSON(http.StatusOK, gin.H{
			"result": "ok",
			"data": gin.H{
				"mangaId":  mangaID,
				"chapters": map[string]interface{}{},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   progress,
	})
}

// getAllProgressHandler retrieves all reading progress for the user
func getAllProgressHandler(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	progress, err := sessions.GetAllReadingProgress(userID)
	if err != nil {
		utils.Logger.Error("Failed to get all reading progress",
			zap.Error(err),
			zap.Int("userId", userID))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve progress",
			"message": "An error occurred while retrieving your progress",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   progress,
	})
}

// getReadMangaListHandler retrieves the list of manga IDs that have been read
func getReadMangaListHandler(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	mangaIDs, err := sessions.GetReadMangaList(userID)
	if err != nil {
		utils.Logger.Error("Failed to get read manga list",
			zap.Error(err),
			zap.Int("userId", userID))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve manga list",
			"message": "An error occurred while retrieving the manga list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   mangaIDs,
		"count":  len(mangaIDs),
	})
}

// deleteChapterProgress removes reading progress for a specific chapter
func deleteChapterProgress(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	chapterID := c.Param("chapterId")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'chapterId' is required",
		})
		return
	}

	err = sessions.DeleteReadingProgress(userID, chapterID)
	if err != nil {
		utils.Logger.Error("Failed to delete reading progress",
			zap.Error(err),
			zap.Int("userId", userID),
			zap.String("chapterId", chapterID))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete progress",
			"message": "An error occurred while deleting your progress",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"message": "Reading progress deleted successfully",
	})
}

// deleteMangaProgressHandler removes all reading progress for a specific manga
func deleteMangaProgressHandler(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'mangaId' is required",
		})
		return
	}

	err = sessions.DeleteMangaProgress(userID, mangaID)
	if err != nil {
		utils.Logger.Error("Failed to delete manga progress",
			zap.Error(err),
			zap.Int("userId", userID),
			zap.String("mangaId", mangaID))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete progress",
			"message": "An error occurred while deleting manga progress",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"message": "Manga progress deleted successfully",
	})
}

// deleteAllProgress removes all reading progress for the user
func deleteAllProgress(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
		return
	}

	err = sessions.DeleteAllReadingProgress(userID)
	if err != nil {
		utils.Logger.Error("Failed to delete all reading progress",
			zap.Error(err),
			zap.Int("userId", userID))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete progress",
			"message": "An error occurred while deleting all progress",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"message": "All reading progress deleted successfully",
	})
}
