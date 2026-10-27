package controllers

import (
	"TaleHaven-Backend/middlewares"
	anilistapi "TaleHaven-Backend/modules/anilist-api"
	"TaleHaven-Backend/modules/database"
	mangadexapi "TaleHaven-Backend/modules/mangadex-api"
	"TaleHaven-Backend/modules/sessions"
	"TaleHaven-Backend/modules/utils"
	"net/http"
	"sort"
	"strconv"
	"strings"

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
		libraryRoutes.GET("/anilist", getAniListUserLibrary)
		libraryRoutes.POST("/anilist/:anilistId", addAniListMangaToLibrary)
		libraryRoutes.PATCH("/anilist/:anilistId/status", updateAniListMangaReadingStatus)
		libraryRoutes.PATCH("/anilist/:anilistId/rating", updateAniListMangaPersonalRating)
		libraryRoutes.DELETE("/anilist/:anilistId", removeAniListMangaFromLibrary)
		libraryRoutes.POST("/:mangaId", addMangaToLibrary)
		libraryRoutes.DELETE("/:mangaId", removeMangaFromLibrary)
	}
}

func getAniListUserLibrary(c *gin.Context) {
	userID, ok := getLibraryUserID(c)
	if !ok {
		return
	}
	limit, offset, ok := getLibraryPagination(c)
	if !ok {
		return
	}
	items, total, err := database.GetAniListUserLibraryPage(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve library", "message": "Please try again later"})
		return
	}
	statusCounts, err := database.GetAniListUserLibraryStatusCounts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profile statistics", "message": "Please try again later"})
		return
	}
	chaptersRead, _, err := sessions.GetReadingStatistics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapter statistics", "message": "Please try again later"})
		return
	}
	lastCompletedID, hasLastCompleted, err := database.GetLastCompletedAniListMangaID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve last completed manga", "message": "Please try again later"})
		return
	}
	topGenres, err := getTopCompletedGenres(c, userID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AniList metadata is unavailable", "message": err.Error()})
		return
	}
	ids := make([]int, 0, len(items)+1)
	knownIDs := make(map[int]bool, len(items)+1)
	for _, item := range items {
		if !knownIDs[item.AniListID] {
			ids = append(ids, item.AniListID)
			knownIDs[item.AniListID] = true
		}
	}
	if hasLastCompleted && !knownIDs[lastCompletedID] {
		ids = append(ids, lastCompletedID)
	}
	metadata, err := aniListClient.GetMangaByIDs(c.Request.Context(), ids)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AniList metadata is unavailable", "message": err.Error()})
		return
	}
	byID := make(map[int]anilistapi.MangaMetadata, len(metadata))
	for _, manga := range metadata {
		byID[manga.ID] = manga
	}
	var lastMangaCompleted gin.H
	if hasLastCompleted {
		if manga, found := byID[lastCompletedID]; found {
			title := manga.Title.English
			if title == "" {
				title = manga.Title.UserPreferred
			}
			if title == "" {
				title = manga.Title.Romaji
			}
			if title == "" {
				title = "Completed manga"
			}
			coverURL := manga.CoverImage.Large
			if coverURL == "" {
				coverURL = manga.CoverImage.Medium
			}
			lastMangaCompleted = gin.H{"anilistId": lastCompletedID, "title": title, "coverUrl": coverURL}
		}
	}
	data := make([]gin.H, 0, len(items))
	for _, item := range items {
		if manga, found := byID[item.AniListID]; found {
			data = append(data, gin.H{"anilistId": item.AniListID, "addedAt": item.AddedAt, "readingStatus": item.ReadingStatus, "personalRating": item.PersonalRating, "manga": manga})
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"result": "ok", "count": len(data), "total": total, "limit": limit, "offset": offset, "data": data,
		"stats": gin.H{
			"total":              total,
			"byStatus":           statusCounts,
			"topGenres":          topGenres,
			"chaptersRead":       chaptersRead,
			"lastMangaCompleted": lastMangaCompleted,
		},
	})
}

// getTopCompletedGenres returns every tied most-frequent genre from the user's
// completed AniList titles. A profile with no completed titles has no top genre.
func getTopCompletedGenres(c *gin.Context, userID int) ([]string, error) {
	ids, err := database.GetAniListUserLibraryIDsByStatus(userID, database.AniListRead)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []string{}, nil
	}

	manga, err := aniListClient.GetMangaByIDs(c.Request.Context(), ids)
	if err != nil {
		return nil, err
	}

	genreCounts := make(map[string]int)
	for _, title := range manga {
		for _, genre := range title.Genres {
			genreCounts[genre]++
		}
	}

	mostReadCount := 0
	for _, count := range genreCounts {
		if count > mostReadCount {
			mostReadCount = count
		}
	}
	if mostReadCount == 0 {
		return []string{}, nil
	}

	topGenres := make([]string, 0)
	for genre, count := range genreCounts {
		if count == mostReadCount {
			topGenres = append(topGenres, genre)
		}
	}
	sort.Strings(topGenres)
	return topGenres, nil
}

type updateAniListMangaReadingStatusRequest struct {
	Status database.AniListReadingStatus `json:"status"`
}

type updateAniListMangaPersonalRatingRequest struct {
	PersonalRating *int `json:"personalRating"`
}

func updateAniListMangaReadingStatus(c *gin.Context) {
	userID, ok := getLibraryUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID", "message": "anilistId must be a positive integer"})
		return
	}
	var request updateAniListMangaReadingStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status update", "message": "A JSON status is required"})
		return
	}
	request.Status = database.AniListReadingStatus(strings.TrimSpace(string(request.Status)))
	if !database.IsValidAniListReadingStatus(request.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reading status", "message": "status must be PLAN_TO_READ, CURRENTLY_READING, ON_HOLD, READ, or DROPPED"})
		return
	}
	updated, err := database.SetAniListMangaReadingStatus(userID, id, request.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reading status", "message": "Please try again later"})
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"error": "Library title not found", "message": "Add this manga to your library before setting its status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "readingStatus": request.Status, "message": "Reading status updated"})
}

func updateAniListMangaPersonalRating(c *gin.Context) {
	userID, ok := getLibraryUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID", "message": "anilistId must be a positive integer"})
		return
	}
	var request updateAniListMangaPersonalRatingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid personal rating update", "message": "A JSON object is required"})
		return
	}
	if request.PersonalRating != nil && (*request.PersonalRating < 1 || *request.PersonalRating > 10) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid personal rating", "message": "personalRating must be between 1 and 10, or null"})
		return
	}
	updated, err := database.SetAniListMangaPersonalRating(userID, id, request.PersonalRating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update personal rating", "message": "Please try again later"})
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"error": "Library title not found", "message": "Add this manga to your library before rating it"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "personalRating": request.PersonalRating, "message": "Personal rating updated"})
}

func addAniListMangaToLibrary(c *gin.Context) {
	userID, ok := getLibraryUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID", "message": "anilistId must be a positive integer"})
		return
	}
	added, err := database.AddAniListMangaToUserLibrary(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add manga", "message": "Please try again later"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "added": added, "message": "Manga added to your library"})
}

func removeAniListMangaFromLibrary(c *gin.Context) {
	userID, ok := getLibraryUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID", "message": "anilistId must be a positive integer"})
		return
	}
	removed, err := database.RemoveAniListMangaFromUserLibrary(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove manga", "message": "Please try again later"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "removed": removed, "message": "Manga removed from your library"})
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
