package controllers

import (
	"TaleHaven-Backend/modules/database"
	mangadex_api "TaleHaven-Backend/modules/mangadex-api"
	"TaleHaven-Backend/modules/sessions"
	"TaleHaven-Backend/modules/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MangaDexController sets up routes for MangaDex API endpoints
func MangaDexController(router *gin.Engine) {
	mangaRoutes := router.Group("/manga")
	{
		mangaRoutes.GET("/search", searchManga)                          // GET /manga/search?title=naruto&limit=10&offset=0
		mangaRoutes.GET("/popular", getPopularManga)                     // GET /manga/popular?limit=20&offset=0&orderBy=followedCount
		mangaRoutes.GET("/library", getDownloadedMangaLibrary)           // GET /manga/library - Get all downloaded manga with covers
		mangaRoutes.GET("/:mangaId", getMangaByID)                       // GET /manga/{mangaId}
		mangaRoutes.GET("/:mangaId/chapters", getChapters)               // GET /manga/{mangaId}/chapters
		mangaRoutes.GET("/:mangaId/cover", getMangaCover)                // GET /manga/{mangaId}/cover
		mangaRoutes.POST("/:mangaId/download-cover", downloadMangaCover) // POST /manga/{mangaId}/download-cover
	}

	chapterRoutes := router.Group("/chapter")
	{
		chapterRoutes.GET("/:chapterId", getChapterByID)                  // GET /chapter/{chapterId}
		chapterRoutes.GET("/:chapterId/pages", getChapterPages)           // GET /chapter/{chapterId}/pages
		chapterRoutes.GET("/:chapterId/pages/stream", streamChapterPages) // GET /chapter/{chapterId}/pages/stream
		chapterRoutes.POST("/:chapterId/download", downloadChapter)       // POST /chapter/{chapterId}/download
	}

	// Serve downloaded images as static files
	// Access pattern: /images/manga/{mangaId}/chapters/{chapterId}/1.jpg
	router.Static("/images", "./images")
	utils.Logger.Info("Static file serving enabled for /images")
}

// getPopularManga handles popular manga requests with cover images
func getPopularManga(c *gin.Context) {
	// Parse optional parameters
	limit := parseIntQuery(c, "limit", 20)
	offset := parseIntQuery(c, "offset", 0)
	orderBy := c.DefaultQuery("orderBy", "followedCount") // followedCount or rating
	includedTags := c.QueryArray("includedTags[]")
	excludedTags := c.QueryArray("excludedTags[]")
	// Parse language filter - defaults to English
	languages := c.QueryArray("availableLanguage[]")
	if len(languages) == 0 {
		languages = []string{"en"} // Default to English
	}

	utils.Logger.Info("Getting popular manga",
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("orderBy", orderBy),
		zap.Int("includedTagsCount", len(includedTags)),
		zap.Int("excludedTagsCount", len(excludedTags)),
		zap.Strings("availableLanguages", languages))

	response, err := mangadex_api.GetPopularManga(limit, offset, orderBy, includedTags, excludedTags, languages)
	if err != nil {
		utils.Logger.Error("Failed to get popular manga", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve popular manga",
			"message": err.Error(),
		})
		return
	}

	// Transform to simplified format with cover URLs
	simplifiedManga := make([]mangadex_api.SimpleManga, 0)
	for _, manga := range response.Data {
		simplified := mangadex_api.TransformToSimpleManga(manga)
		// Build full cover URL using the proper helper function
		coverURL := mangadex_api.BuildFullCoverURL(manga.ID, manga.Relationships, "512")
		if coverURL != "" {
			simplified.CoverUrl = coverURL
		}
		simplifiedManga = append(simplifiedManga, simplified)
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   simplifiedManga,
		"total":  response.Total,
		"limit":  response.Limit,
		"offset": response.Offset,
	})
}

// searchManga handles manga search requests
// Supports both title search and filter-only search
func searchManga(c *gin.Context) {
	title := c.Query("title") // Optional - can search by filters only

	// Parse optional parameters
	limit := parseIntQuery(c, "limit", 10)
	offset := parseIntQuery(c, "offset", 0)
	includedTags := c.QueryArray("includedTags[]")
	excludedTags := c.QueryArray("excludedTags[]")
	// Parse language filter - defaults to English
	languages := c.QueryArray("availableLanguage[]")
	if len(languages) == 0 {
		languages = []string{"en"} // Default to English
	}

	utils.Logger.Info("Searching manga",
		zap.String("title", title),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.Int("includedTagsCount", len(includedTags)),
		zap.Int("excludedTagsCount", len(excludedTags)),
		zap.Strings("availableLanguages", languages))

	response, err := mangadex_api.SearchManga(title, limit, offset, includedTags, excludedTags, languages)
	if err != nil {
		utils.Logger.Error("Failed to search manga", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"message": err.Error(),
		})
		return
	}

	// Transform to simplified format
	simplifiedManga := make([]mangadex_api.SimpleManga, 0)
	for _, manga := range response.Data {
		simplified := mangadex_api.TransformToSimpleManga(manga)
		// Build full cover URL using the proper helper function
		coverURL := mangadex_api.BuildFullCoverURL(manga.ID, manga.Relationships, "512")
		if coverURL != "" {
			simplified.CoverUrl = coverURL
		}
		simplifiedManga = append(simplifiedManga, simplified)
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   simplifiedManga,
		"total":  response.Total,
		"limit":  response.Limit,
		"offset": response.Offset,
	})
}

// getDownloadedMangaLibrary returns all manga that have been downloaded with cover images
func getDownloadedMangaLibrary(c *gin.Context) {
	utils.Logger.Info("Getting downloaded manga library")

	mangaList, err := database.GetDownloadedMangaWithCovers()
	if err != nil {
		utils.Logger.Error("Failed to get downloaded manga library", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve library",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   mangaList,
		"count":  len(mangaList),
	})
}

// getMangaByID retrieves a specific manga by ID
func getMangaByID(c *gin.Context) {
	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'mangaId' is required",
		})
		return
	}

	utils.Logger.Info("Getting manga by ID", zap.String("mangaId", mangaID))

	response, err := mangadex_api.GetMangaByID(mangaID)
	if err != nil {
		utils.Logger.Error("Failed to get manga", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve manga",
			"message": err.Error(),
		})
		return
	}

	// Save to database asynchronously (user clicked on manga)
	go func() {
		err := mangadex_api.SaveMangaToDBIfNotExists(mangaID, response)
		if err != nil {
			utils.Logger.Error("Failed to save manga to database",
				zap.Error(err),
				zap.String("mangaId", mangaID))
		} else {
			utils.Logger.Info("Manga saved to database", zap.String("mangaId", mangaID))
		}
	}()

	// Transform to simplified format
	simplified := mangadex_api.TransformToSimpleManga(response.Data)
	// Build full cover URL using the proper helper function
	coverURL := mangadex_api.BuildFullCoverURL(response.Data.ID, response.Data.Relationships, "512")
	if coverURL != "" {
		simplified.CoverUrl = coverURL
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   simplified,
	})
}

// getChapters retrieves chapters for a manga
func getChapters(c *gin.Context) {
	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'mangaId' is required",
		})
		return
	}

	// Parse optional parameters
	limit := parseIntQuery(c, "limit", 100)
	offset := parseIntQuery(c, "offset", 0)
	languages := c.QueryArray("translatedLanguage[]")
	if len(languages) == 0 {
		languages = []string{"en"} // Default to English
	}

	utils.Logger.Info("Getting manga chapters",
		zap.String("mangaId", mangaID),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.Strings("languages", languages))

	response, err := mangadex_api.GetMangaChapters(mangaID, limit, offset, languages)
	if err != nil {
		utils.Logger.Error("Failed to get chapters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve chapters",
			"message": err.Error(),
		})
		return
	}

	// Transform to simplified format
	simplifiedChapters := make([]mangadex_api.SimpleChapter, 0)
	for _, chapter := range response.Data {
		simplifiedChapters = append(simplifiedChapters, mangadex_api.TransformToSimpleChapter(chapter))
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   simplifiedChapters,
		"total":  response.Total,
		"limit":  response.Limit,
		"offset": response.Offset,
	})
}

// getChapterByID retrieves a specific chapter by ID
func getChapterByID(c *gin.Context) {
	chapterID := c.Param("chapterId")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'chapterId' is required",
		})
		return
	}

	utils.Logger.Info("Getting chapter by ID", zap.String("chapterId", chapterID))

	response, err := mangadex_api.GetChapterByID(chapterID)
	if err != nil {
		utils.Logger.Error("Failed to get chapter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve chapter",
			"message": err.Error(),
		})
		return
	}

	simplified := mangadex_api.TransformToSimpleChapter(response.Data)

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"data":   simplified,
	})
}

// getChapterPages retrieves page URLs for a chapter and downloads images
func getChapterPages(c *gin.Context) {
	chapterID := c.Param("chapterId")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'chapterId' is required",
		})
		return
	}

	quality := c.DefaultQuery("quality", "data") // "data" or "saver"

	utils.Logger.Info("Getting chapter pages",
		zap.String("chapterId", chapterID),
		zap.String("quality", quality))

	// Download chapter images and get local paths
	result, err := mangadex_api.DownloadAndGetChapterPages(chapterID, quality)
	if err != nil {
		utils.Logger.Error("Failed to get/download chapter pages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve chapter pages",
			"message": err.Error(),
		})
		return
	}

	// Track manga access for auto-cleanup (manga not accessed in 2 weeks will be deleted)
	if result.MangaID != "" {
		err = sessions.TrackMangaAccess(result.MangaID, chapterID)
		if err != nil {
			utils.Logger.Warn("Failed to track manga access",
				zap.String("mangaId", result.MangaID),
				zap.String("chapterId", chapterID),
				zap.Error(err))
			// Don't fail the request if tracking fails
		}

		// Prefetch next chapter to reduce lag when user continues reading
		// This runs in the background and won't block the response
		utils.Logger.Info("Initiating prefetch for next chapter",
			zap.String("currentChapterId", chapterID),
			zap.String("mangaId", result.MangaID))
		mangadex_api.PrefetchNextChapter(chapterID, result.MangaID, quality)
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    "ok",
		"mangaId":   result.MangaID,
		"chapterId": chapterID,
		"pages":     result.LocalPaths,
		"pageCount": len(result.LocalPaths),
	})
}

// streamChapterPages streams pages progressively via SSE while downloading in order.
func streamChapterPages(c *gin.Context) {
	chapterID := c.Param("chapterId")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'chapterId' is required",
		})
		return
	}

	quality := c.DefaultQuery("quality", "data")

	// SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	// Announce stream start quickly so frontend can show loader immediately.
	c.SSEvent("start", gin.H{
		"result":    "ok",
		"chapterId": chapterID,
		"message":   "Starting progressive download",
	})
	c.Writer.Flush()

	result, err := mangadex_api.DownloadChapterPagesProgressively(chapterID, quality, func(event mangadex_api.PageDownloadEvent) error {
		select {
		case <-c.Request.Context().Done():
			return fmt.Errorf("client disconnected")
		default:
		}

		if event.Error != "" {
			c.SSEvent("page_error", gin.H{
				"chapterId":  event.ChapterID,
				"mangaId":    event.MangaID,
				"pageNumber": event.PageNumber,
				"totalPages": event.TotalPages,
				"error":      event.Error,
			})
		} else {
			c.SSEvent("page", gin.H{
				"chapterId":  event.ChapterID,
				"mangaId":    event.MangaID,
				"pageNumber": event.PageNumber,
				"totalPages": event.TotalPages,
				"pageUrl":    event.PublicURL,
				"downloaded": event.Downloaded,
			})
		}

		c.Writer.Flush()
		return nil
	})

	if err != nil {
		utils.Logger.Error("Progressive page streaming failed",
			zap.String("chapterId", chapterID),
			zap.Error(err))

		c.SSEvent("fatal_error", gin.H{
			"error":   "Failed to stream chapter pages",
			"message": err.Error(),
		})
		c.Writer.Flush()
		return
	}

	// Track access and prefetch next chapter after current stream completes.
	if result.MangaID != "" {
		if err := sessions.TrackMangaAccess(result.MangaID, chapterID); err != nil {
			utils.Logger.Warn("Failed to track manga access after stream",
				zap.String("mangaId", result.MangaID),
				zap.String("chapterId", chapterID),
				zap.Error(err))
		}

		mangadex_api.PrefetchNextChapter(chapterID, result.MangaID, quality)
	}

	c.SSEvent("complete", gin.H{
		"result":       "ok",
		"chapterId":    result.ChapterID,
		"mangaId":      result.MangaID,
		"totalPages":   result.TotalPages,
		"successCount": result.SuccessCount,
	})
	c.Writer.Flush()
}

// getMangaCover retrieves the cover URL for a manga
func getMangaCover(c *gin.Context) {
	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'mangaId' is required",
		})
		return
	}

	quality := c.DefaultQuery("quality", "512") // Default 512px

	utils.Logger.Info("Getting manga cover", zap.String("mangaId", mangaID))

	// First get manga details to find cover relationship
	mangaResponse, err := mangadex_api.GetMangaByID(mangaID)
	if err != nil {
		utils.Logger.Error("Failed to get manga for cover", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve cover",
			"message": err.Error(),
		})
		return
	}

	coverURL := mangadex_api.BuildFullCoverURL(mangaID, mangaResponse.Data.Relationships, quality)
	if coverURL == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Cover not found",
			"message": "No cover art available for this manga",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":   "ok",
		"coverUrl": coverURL,
	})
}

// downloadChapter downloads all images for a chapter
func downloadChapter(c *gin.Context) {
	chapterID := c.Param("chapterId")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'chapterId' is required",
		})
		return
	}

	quality := c.DefaultQuery("quality", "data") // "data" or "saver"

	utils.Logger.Info("Starting chapter download",
		zap.String("chapterId", chapterID),
		zap.String("quality", quality))

	// Start download in background (async)
	go func() {
		err := mangadex_api.DownloadChapterImages(chapterID, quality)
		if err != nil {
			utils.Logger.Error("Failed to download chapter images",
				zap.Error(err),
				zap.String("chapterId", chapterID))
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"result":    "ok",
		"message":   "Chapter download started",
		"chapterId": chapterID,
	})
}

// downloadMangaCover downloads the cover image for a manga
func downloadMangaCover(c *gin.Context) {
	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing parameter",
			"message": "Path parameter 'mangaId' is required",
		})
		return
	}

	quality := c.DefaultQuery("quality", "512")

	utils.Logger.Info("Starting cover download",
		zap.String("mangaId", mangaID),
		zap.String("quality", quality))

	err := mangadex_api.DownloadCoverImage(mangaID, quality)
	if err != nil {
		utils.Logger.Error("Failed to download cover", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Download failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"message": "Cover downloaded successfully",
		"mangaId": mangaID,
	})
}

// Helper function to parse integer query parameters
func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
