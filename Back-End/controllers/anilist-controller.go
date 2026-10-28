package controllers

import (
	anilist_api "TaleHaven-Backend/modules/anilist-api"
	mangadex_api "TaleHaven-Backend/modules/mangadex-api"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var aniListClient = anilist_api.NewClient()

// AniListController exposes public, backend-owned AniList metadata routes.
func AniListController(router *gin.Engine) {
	router.GET("/anilist/character/details/:characterId", getAniListCharacterDetails)
	routes := router.Group("/anilist/manga")
	routes.GET("/search", searchAniListManga)
	routes.GET("/reading-source/:anilistId", getAniListReadingSource)
	routes.GET("/details/:anilistId", getAniListMangaDetails)
	routes.GET("/characters/:anilistId", getAniListMangaCharacters)
	routes.GET("/genre/:genre", getAniListMangaByGenre)
	routes.GET("/:section", getAniListDiscoverySection)
}

func searchAniListManga(c *gin.Context) {
	year, err := optionalIntQuery(c, "year")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search filters", "message": "year must be a whole number"})
		return
	}
	staffID, err := optionalIntQuery(c, "staffId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search filters", "message": "staffId must be a whole number"})
		return
	}
	filters, err := (anilist_api.MangaSearchFilters{
		Title:   c.Query("title"),
		Genre:   c.Query("genre"),
		Genres:  c.QueryArray("genres[]"),
		Tag:     c.Query("tag"),
		Year:    year,
		Format:  c.Query("format"),
		Status:  c.Query("status"),
		StaffID: staffID,
		Sort:    c.Query("sort"),
	}).Normalize()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search filters", "message": err.Error()})
		return
	}
	page, perPage := parseIntQuery(c, "page", 1), parseIntQuery(c, "perPage", 20)
	if page < 1 || perPage < 1 || perPage > 25 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination", "message": "page must be at least 1 and perPage must be between 1 and 25"})
		return
	}
	result, err := aniListClient.SearchManga(c.Request.Context(), filters, page, perPage)
	if err != nil {
		respondAniListUnavailable(c, err)
		return
	}
	respondAniListPage(c, result)
}

func optionalIntQuery(c *gin.Context, key string) (int, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func getAniListReadingSource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID", "message": "anilistId must be a positive integer"})
		return
	}
	manga, err := aniListClient.GetMangaByID(c.Request.Context(), id)
	if err != nil {
		respondAniListUnavailable(c, err)
		return
	}
	candidates := make([]mangadex_api.MangaData, 0, 30)
	seen := make(map[string]struct{})
	for _, title := range aniListTitleVariants(*manga) {
		response, searchErr := mangadex_api.SearchManga(title, 10, 0, nil, nil, nil)
		if searchErr != nil {
			continue
		}
		for _, result := range response.Data {
			if _, exists := seen[result.ID]; !exists {
				seen[result.ID] = struct{}{}
				candidates = append(candidates, result)
			}
		}
	}
	matched, found := mangadex_api.FindConfidentMangaMatchWithMetadata(candidates, mangadex_api.MatchMetadata{
		Titles:       aniListTitleVariants(*manga),
		Year:         manga.StartDate.Year,
		CreatorNames: aniListCreatorNames(*manga),
	})
	if !found {
		c.JSON(http.StatusOK, gin.H{"result": "ok", "available": false})
		return
	}
	chapters, chapterErr := mangadex_api.GetMangaChapters(matched.ID, 1, 0, nil)
	if chapterErr != nil || len(chapters.Data) == 0 {
		c.JSON(http.StatusOK, gin.H{"result": "ok", "available": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "available": true, "mangaId": matched.ID})
}

func aniListTitleVariants(manga anilist_api.MangaMetadata) []string {
	values := append([]string{manga.Title.English, manga.Title.Romaji, manga.Title.Native, manga.Title.UserPreferred}, manga.Synonyms...)
	result, seen := make([]string, 0, len(values)), make(map[string]struct{})
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value != "" {
			if _, exists := seen[key]; !exists {
				seen[key] = struct{}{}
				result = append(result, value)
			}
		}
	}
	return result
}

func aniListCreatorNames(manga anilist_api.MangaMetadata) []string {
	result, seen := make([]string, 0), make(map[string]struct{})
	for _, edge := range manga.Staff.Edges {
		role := strings.ToLower(edge.Role)
		if !strings.Contains(role, "story") && !strings.Contains(role, "art") && !strings.Contains(role, "author") && !strings.Contains(role, "creator") && !strings.Contains(role, "illustrat") {
			continue
		}
		for _, value := range []string{edge.Node.Name.Full, edge.Node.Name.UserPreferred, edge.Node.Name.Native} {
			value = strings.TrimSpace(value)
			key := strings.ToLower(value)
			if value == "" {
				continue
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func getAniListCharacterDetails(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("characterId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList character ID", "message": "characterId must be a positive integer"})
		return
	}
	character, err := aniListClient.GetCharacterByID(c.Request.Context(), id)
	if err != nil {
		respondAniListUnavailable(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": character})
}

func getAniListMangaDetails(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID", "message": "anilistId must be a positive integer"})
		return
	}
	manga, err := aniListClient.GetMangaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AniList metadata is unavailable", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": manga})
}

func getAniListMangaCharacters(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID", "message": "anilistId must be a positive integer"})
		return
	}
	page, perPage := parseIntQuery(c, "page", 1), parseIntQuery(c, "perPage", 10)
	if page < 1 || perPage < 1 || perPage > 25 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination", "message": "page must be at least 1 and perPage must be between 1 and 25"})
		return
	}
	characters, err := aniListClient.GetMangaCharacters(c.Request.Context(), id, page, perPage)
	if err != nil {
		respondAniListUnavailable(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": characters})
}

func getAniListMangaByGenre(c *gin.Context) {
	genre := strings.TrimSpace(c.Param("genre"))
	if genre == "" || len([]rune(genre)) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre", "message": "genre must be between 1 and 100 characters"})
		return
	}

	page := parseIntQuery(c, "page", 1)
	perPage := parseIntQuery(c, "perPage", 12)
	if page < 1 || perPage < 1 || perPage > 25 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid pagination",
			"message": "page must be at least 1 and perPage must be between 1 and 25",
		})
		return
	}

	result, err := aniListClient.GetMangaByGenre(c.Request.Context(), genre, page, perPage)
	if err != nil {
		respondAniListUnavailable(c, err)
		return
	}
	respondAniListPage(c, result)
}

func getAniListDiscoverySection(c *gin.Context) {
	section := anilist_api.DiscoverySection(c.Param("section"))
	if _, supported := section.Sort(); !supported {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid discovery section",
			"message": "section must be one of: trending, popular, top-rated, recent",
		})
		return
	}

	page := parseIntQuery(c, "page", 1)
	perPage := parseIntQuery(c, "perPage", 12)
	if page < 1 || perPage < 1 || perPage > 25 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid pagination",
			"message": "page must be at least 1 and perPage must be between 1 and 25",
		})
		return
	}

	result, err := aniListClient.GetDiscoveryManga(c.Request.Context(), section, page, perPage)
	if err != nil {
		respondAniListUnavailable(c, err)
		return
	}

	respondAniListPage(c, result)
}

func respondAniListUnavailable(c *gin.Context, err error) {
	status := http.StatusBadGateway
	var apiError *anilist_api.APIError
	if errors.As(err, &apiError) && apiError.Status == http.StatusTooManyRequests {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{
		"error":   "AniList metadata is unavailable",
		"message": err.Error(),
	})
}

func respondAniListPage(c *gin.Context, result *anilist_api.MangaSearchResult) {
	c.JSON(http.StatusOK, gin.H{
		"result":      "ok",
		"data":        result.Manga,
		"total":       result.PageInfo.Total,
		"page":        result.PageInfo.CurrentPage,
		"perPage":     result.PageInfo.PerPage,
		"hasNextPage": result.PageInfo.HasNextPage,
	})
}
