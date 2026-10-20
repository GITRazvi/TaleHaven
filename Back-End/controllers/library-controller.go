package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/database"
	mangadexapi "TaleHaven-Backend/modules/mangadex-api"
	"TaleHaven-Backend/modules/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	defaultLibraryPageSize = 24
	maxLibraryPageSize     = 100
)

// LibraryController exposes the signed-in user's saved manga library.
func LibraryController(router *gin.Engine) {
	libraryRoutes := router.Group("/library")
	libraryRoutes.Use(middlewares.AuthMiddleware())
	{
		libraryRoutes.GET("", getUserLibrary)
		libraryRoutes.POST("/:mangaId", addMangaToLibrary)
		libraryRoutes.DELETE("/:mangaId", removeMangaFromLibrary)
	}
}

func getLibraryUserID(c *gin.Context) (int, bool) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required", "message": "Please log in to access your library"})
		return 0, false
	}

	id, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authentication context", "message": "Could not identify the signed-in user"})
		return 0, false
	}
	return id, true
}

func getUserLibrary(c *gin.Context) {
	userID, ok := getLibraryUserID(c)
	if !ok {
		return
	}

	limit, offset, ok := getLibraryPagination(c)
	if !ok {
		return
	}

	library, total, err := database.GetUserLibraryPage(userID, limit, offset)
	if err != nil {
		utils.Logger.Error("Failed to retrieve user library", zap.Int("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve library", "message": "Please try again later"})
		return
	}

	items := make([]gin.H, 0, len(library))
	for _, manga := range library {
		coverURL := ""
		if manga.CoverFileName != "" {
			coverURL = mangadexapi.BuildCoverURL(manga.MangaDexID, manga.CoverFileName, "512")
		}
		items = append(items, gin.H{
			"mangaId":  manga.MangaDexID,
			"title":    manga.Title,
			"coverUrl": coverURL,
			"addedAt":  manga.AddedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"count":  len(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"data":   items,
	})
}

func getLibraryPagination(c *gin.Context) (int, int, bool) {
	limit := defaultLibraryPageSize
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 || parsedLimit > maxLibraryPageSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit", "message": "Query parameter 'limit' must be between 1 and 100"})
			return 0, 0, false
		}
		limit = parsedLimit
	}

	offset := 0
	if rawOffset := c.Query("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil || parsedOffset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset", "message": "Query parameter 'offset' must be zero or greater"})
			return 0, 0, false
		}
		offset = parsedOffset
	}

	return limit, offset, true
}

func addMangaToLibrary(c *gin.Context) {
	userID, ok := getLibraryUserID(c)
	if !ok {
		return
	}

	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameter", "message": "Path parameter 'mangaId' is required"})
		return
	}

	mangaResponse, err := mangadexapi.GetMangaByID(mangaID)
	if err != nil {
		utils.Logger.Error("Failed to retrieve manga before adding to library", zap.String("manga_id", mangaID), zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to retrieve manga", "message": "The title could not be added to your library"})
		return
	}

	manga, err := mangadexapi.SaveMangaToDB(mangaResponse)
	if err != nil {
		utils.Logger.Error("Failed to save manga before adding to library", zap.String("manga_id", mangaID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save manga", "message": "The title could not be added to your library"})
		return
	}

	added, err := database.AddMangaToUserLibrary(userID, manga.ID)
	if err != nil {
		utils.Logger.Error("Failed to add manga to user library", zap.Int("user_id", userID), zap.String("manga_id", mangaID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add manga", "message": "Please try again later"})
		return
	}

	message := "Manga added to your library"
	if !added {
		message = "Manga is already in your library"
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "added": added, "message": message})
}

func removeMangaFromLibrary(c *gin.Context) {
	userID, ok := getLibraryUserID(c)
	if !ok {
		return
	}

	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameter", "message": "Path parameter 'mangaId' is required"})
		return
	}

	removed, err := database.RemoveMangaFromUserLibrary(userID, mangaID)
	if err != nil {
		utils.Logger.Error("Failed to remove manga from user library", zap.Int("user_id", userID), zap.String("manga_id", mangaID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove manga", "message": "Please try again later"})
		return
	}

	message := "Manga removed from your library"
	if !removed {
		message = "Manga was not in your library"
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "removed": removed, "message": message})
}
