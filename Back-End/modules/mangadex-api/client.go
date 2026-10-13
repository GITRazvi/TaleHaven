package mangadex_api

import (
	"TaleHaven-Backend/modules/cache"
	"TaleHaven-Backend/modules/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	BaseURL         = "https://api.mangadex.org"
	CoverBaseURL    = "https://uploads.mangadex.org/covers"
	CachePrefix     = "mangadex"
	CacheTTL        = 1 * time.Hour        // General cache TTL
	SearchCacheTTL  = 15 * time.Minute     // Search cache TTL (shorter for freshness)
)

// SearchManga searches for manga by title
func SearchManga(title string, limit int, offset int, includedTags []string, excludedTags []string, availableLanguages []string) (*MangaListResponse, error) {
	params := url.Values{}
	params.Add("title", title)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	params.Add("includes[]", "cover_art")
	params.Add("includes[]", "author")
	params.Add("includes[]", "artist")
	params.Add("contentRating[]", "safe")
	params.Add("contentRating[]", "suggestive")
	params.Add("contentRating[]", "erotica")

	for _, tag := range includedTags {
		params.Add("includedTags[]", tag)
	}
	for _, tag := range excludedTags {
		params.Add("excludedTags[]", tag)
	}
	
	// Filter by available translated languages
	for _, lang := range availableLanguages {
		params.Add("availableTranslatedLanguage[]", lang)
	}

	cacheKey := fmt.Sprintf("%s:search:%s", CachePrefix, utils.HashText(params.Encode()))

	// Try cache first
	cachedData, err := getFromCache(cacheKey)
	if err == nil && len(cachedData) > 0 {
		var response MangaListResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			utils.Logger.Info("MangaDex search cache hit", zap.String("title", title))
			return &response, nil
		}
	}

	// Make API request
	requestURL := fmt.Sprintf("%s/manga?%s", BaseURL, params.Encode())
	body, err := makeRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var response MangaListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		utils.Logger.Error("Failed to parse MangaDex search response", zap.Error(err))
		return nil, err
	}

	// Cache the response with 15 minute TTL for search results
	_ = setInCache(cacheKey, body, SearchCacheTTL)

	return &response, nil
}

// GetPopularManga retrieves popular manga ordered by follower count or rating
func GetPopularManga(limit int, offset int, orderBy string, includedTags []string, excludedTags []string, availableLanguages []string) (*MangaListResponse, error) {
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	params.Add("includes[]", "cover_art")
	params.Add("includes[]", "author")
	params.Add("includes[]", "artist")
	params.Add("contentRating[]", "safe")
	params.Add("contentRating[]", "suggestive")
	params.Add("contentRating[]", "erotica")

	// Order by popularity (followedCount) or rating
	if orderBy == "" || orderBy == "followedCount" {
		params.Add("order[followedCount]", "desc")
	} else if orderBy == "rating" {
		params.Add("order[rating]", "desc")
	} else {
		params.Add("order[followedCount]", "desc") // Default to followedCount
	}

	for _, tag := range includedTags {
		params.Add("includedTags[]", tag)
	}
	for _, tag := range excludedTags {
		params.Add("excludedTags[]", tag)
	}
	
	// Filter by available translated languages
	for _, lang := range availableLanguages {
		params.Add("availableTranslatedLanguage[]", lang)
	}

	cacheKey := fmt.Sprintf("%s:popular:%s", CachePrefix, utils.HashText(params.Encode()))

	// Try cache first
	cachedData, err := getFromCache(cacheKey)
	if err == nil && len(cachedData) > 0 {
		var response MangaListResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			utils.Logger.Info("MangaDex popular manga cache hit", zap.String("orderBy", orderBy))
			return &response, nil
		}
	}

	// Make API request
	requestURL := fmt.Sprintf("%s/manga?%s", BaseURL, params.Encode())
	body, err := makeRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var response MangaListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		utils.Logger.Error("Failed to parse MangaDex popular manga response", zap.Error(err))
		return nil, err
	}

	// Cache the response with 15 minute TTL
	_ = setInCache(cacheKey, body, SearchCacheTTL)

	return &response, nil
}

// GetMangaByID retrieves a specific manga by ID
func GetMangaByID(mangaID string) (*MangaResponse, error) {
	cacheKey := fmt.Sprintf("%s:manga:%s", CachePrefix, mangaID)

	// Try cache first
	cachedData, err := getFromCache(cacheKey)
	if err == nil && len(cachedData) > 0 {
		var response MangaResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			utils.Logger.Info("MangaDex manga cache hit", zap.String("mangaId", mangaID))
			return &response, nil
		}
	}

	// Make API request
	requestURL := fmt.Sprintf("%s/manga/%s?includes[]=cover_art&includes[]=author&includes[]=artist", BaseURL, mangaID)
	body, err := makeRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var response MangaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		utils.Logger.Error("Failed to parse MangaDex manga response", zap.Error(err))
		return nil, err
	}

	// Cache the response
	_ = setInCache(cacheKey, body, CacheTTL)

	return &response, nil
}

// GetMangaChapters retrieves chapters for a specific manga
func GetMangaChapters(mangaID string, limit int, offset int, translatedLanguage []string) (*ChapterListResponse, error) {
	params := url.Values{}
	params.Add("manga", mangaID)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	params.Add("includes[]", "scanlation_group")
	params.Add("order[chapter]", "asc")

	for _, lang := range translatedLanguage {
		params.Add("translatedLanguage[]", lang)
	}

	cacheKey := fmt.Sprintf("%s:chapters:%s:%s", CachePrefix, mangaID, utils.HashText(params.Encode()))

	// Try cache first
	cachedData, err := getFromCache(cacheKey)
	if err == nil && len(cachedData) > 0 {
		var response ChapterListResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			utils.Logger.Info("MangaDex chapters cache hit", zap.String("mangaId", mangaID))
			return &response, nil
		}
	}

	// Make API request
	requestURL := fmt.Sprintf("%s/chapter?%s", BaseURL, params.Encode())
	body, err := makeRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var response ChapterListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		utils.Logger.Error("Failed to parse MangaDex chapters response", zap.Error(err))
		return nil, err
	}

	// Cache the response
	_ = setInCache(cacheKey, body, CacheTTL)

	return &response, nil
}

// GetChapterByID retrieves a specific chapter by ID
func GetChapterByID(chapterID string) (*ChapterResponse, error) {
	cacheKey := fmt.Sprintf("%s:chapter:%s", CachePrefix, chapterID)

	// Try cache first
	cachedData, err := getFromCache(cacheKey)
	if err == nil && len(cachedData) > 0 {
		var response ChapterResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			utils.Logger.Info("MangaDex chapter cache hit", zap.String("chapterId", chapterID))
			return &response, nil
		}
	}

	// Make API request
	requestURL := fmt.Sprintf("%s/chapter/%s?includes[]=scanlation_group&includes[]=manga", BaseURL, chapterID)
	body, err := makeRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var response ChapterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		utils.Logger.Error("Failed to parse MangaDex chapter response", zap.Error(err))
		return nil, err
	}

	// Cache the response
	_ = setInCache(cacheKey, body, CacheTTL)

	return &response, nil
}

// GetChapterPages retrieves the page URLs for a specific chapter
func GetChapterPages(chapterID string) (*ChapterPagesResponse, error) {
	cacheKey := fmt.Sprintf("%s:chapter-pages:%s", CachePrefix, chapterID)

	// Try cache first
	cachedData, err := getFromCache(cacheKey)
	if err == nil && len(cachedData) > 0 {
		var response ChapterPagesResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			utils.Logger.Info("MangaDex chapter pages cache hit", zap.String("chapterId", chapterID))
			return &response, nil
		}
	}

	// Make API request
	requestURL := fmt.Sprintf("%s/at-home/server/%s", BaseURL, chapterID)
	body, err := makeRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var response ChapterPagesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		utils.Logger.Error("Failed to parse MangaDex chapter pages response", zap.Error(err))
		return nil, err
	}

	// Cache the response (shorter TTL for page URLs)
	_ = setInCache(cacheKey, body, 15*time.Minute)

	return &response, nil
}

// GetCoverArt retrieves cover art for a manga
func GetCoverArt(coverID string) (*CoverData, error) {
	cacheKey := fmt.Sprintf("%s:cover:%s", CachePrefix, coverID)

	// Try cache first
	cachedData, err := getFromCache(cacheKey)
	if err == nil && len(cachedData) > 0 {
		var coverData CoverData
		if err := json.Unmarshal(cachedData, &coverData); err == nil {
			utils.Logger.Info("MangaDex cover cache hit", zap.String("coverId", coverID))
			return &coverData, nil
		}
	}

	// Make API request
	requestURL := fmt.Sprintf("%s/cover/%s", BaseURL, coverID)
	body, err := makeRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data CoverData `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		utils.Logger.Error("Failed to parse MangaDex cover response", zap.Error(err))
		return nil, err
	}

	// Cache the response
	_ = setInCache(cacheKey, body, CacheTTL)

	return &response.Data, nil
}

// Helper function to make HTTP requests
func makeRequest(requestURL string) ([]byte, error) {
	utils.Logger.Info("Making MangaDex API request", zap.String("url", requestURL))

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "TaleHaven-Backend/1.0")

	resp, err := client.Do(req)
	if err != nil {
		utils.Logger.Error("MangaDex API request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.Logger.Error("MangaDex API returned non-200 status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("MangaDex API error: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Logger.Error("Failed to read MangaDex response body", zap.Error(err))
		return nil, err
	}

	return body, nil
}

// Cache helpers
func getFromCache(key string) ([]byte, error) {
	responseFromCache, err := cache.Client.Get(cache.Ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		utils.Logger.Error("Failed to get response from cache", zap.Error(err), zap.String("key", key))
		return nil, err
	}
	return []byte(responseFromCache), nil
}

func setInCache(key string, data []byte, ttl time.Duration) error {
	redisError := cache.Client.Set(cache.Ctx, key, data, ttl).Err()
	if redisError != nil {
		utils.Logger.Error("Failed to set response in redis", zap.Error(redisError), zap.String("key", key))
		return redisError
	}
	return nil
}

// BuildCoverURL constructs the full URL for a manga cover
// fileName should include the extension (e.g., "abc123.jpg")
// MangaDex format: covers/{mangaId}/{fileName}.{quality}.jpg
// Example: abc123.jpg becomes abc123.jpg.512.jpg for 512px quality
func BuildCoverURL(mangaID string, fileName string, quality string) string {
	if quality == "" {
		quality = "512" // Default to 512px
	}
	// fileName already has extension, just append .{quality}.jpg
	return fmt.Sprintf("%s/%s/%s.%s.jpg", CoverBaseURL, mangaID, fileName, quality)
}
