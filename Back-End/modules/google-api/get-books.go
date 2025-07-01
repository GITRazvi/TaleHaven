package google_api

import (
	"TaleHaven-Backend/config"
	"TaleHaven-Backend/modules/cache"
	"TaleHaven-Backend/modules/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type googleVolumesInfo struct {
	Id         string           `json:"id"`
	VolumeInfo googleVolumeInfo `json:"volumeInfo"`
}

type googleVolumeInfo struct {
	Title         string               `json:"title"`
	Authors       []string             `json:"authors"`
	PublishedDate string               `json:"publishedDate"`
	Description   string               `json:"description"`
	AverageRating float32              `json:"averageRating"`
	RatingsCount  int                  `json:"ratingsCount"`
	Categories    []string             `json:"categories"`
	ImageLinks    googleThumbnailLinks `json:"imageLinks"`
}

type googleThumbnailLinks struct {
	Thumbnail string `json:"thumbnail"`
}

type volumeInfo struct {
	Title         string
	Authors       []string
	Id            string
	Description   string
	Categories    []string
	Thumbnail     string
	Rating        float32
	RatingCount   int
	PublishedDate string
}

// https://www.googleapis.com/books/v1/volumes/s1gVAAAAYAAJ
//http://127.0.0.1:3000/book/aHUUEQAAQBAJ

func GetBooks(bookId string) (volumeInfo, error) {
	body, err := getVolumeInfoFromCache(bookId)
	if err != nil {
		return volumeInfo{}, err
	}

	if len(body) == 0 {
		body, err = downloadBookInfo(bookId)
		if err != nil {
			return volumeInfo{}, err
		}

		if setError := setBookInformationInCache(body, bookId); setError != nil {
			return volumeInfo{}, err
		}

	}

	var parsedResponse googleVolumesInfo
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		return volumeInfo{}, err
	}

	bookInfo, err := extractVolumeDetails(parsedResponse)

	return bookInfo, err
}
func downloadBookInfo(bookID string) ([]byte, error) {
	params := url.Values{}
	params.Add("key", os.Getenv("GOOGLE_API_KEY"))
	requestURL := fmt.Sprintf("%s?%s", config.BaseUrl+"volumes/"+bookID, params.Encode())

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %v", resp.Status)
	}

	// Read the response body as raw bytes
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func setBookInformationInCache(body []byte, bookId string) error {
	redisError := cache.Client.Set(cache.Ctx, "google:responses:/get/book:"+bookId, body, 0).Err()
	if redisError != nil {
		utils.Logger.Error("Failed to set google response in redis", zap.Error(redisError))
		return redisError
	}

	return nil
}

func getVolumeInfoFromCache(bookID string) ([]byte, error) {

	responseFromCache, err := cache.Client.Get(cache.Ctx, "google:responses:/get/book:"+bookID).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		utils.Logger.Error("Failed to get google response from cache", zap.Error(err))
		return nil, err
	}

	cachedResponseBytes := []byte(responseFromCache)

	return cachedResponseBytes, nil
}

func extractVolumeDetails(googleParsedResponse googleVolumesInfo) (volumeInfo, error) {
	publishedDate := googleParsedResponse.VolumeInfo.PublishedDate
	if _, err := time.Parse("2006-01-02", googleParsedResponse.VolumeInfo.PublishedDate); err != nil {
		utils.Logger.Error("PublishedDate is not in valid format", zap.Error(err))
		publishedDate = ""
	}

	volumeInfo := volumeInfo{
		Id:            googleParsedResponse.Id,
		Title:         googleParsedResponse.VolumeInfo.Title,
		Authors:       googleParsedResponse.VolumeInfo.Authors,
		Description:   googleParsedResponse.VolumeInfo.Description,
		PublishedDate: publishedDate,
		Rating:        googleParsedResponse.VolumeInfo.AverageRating,
		RatingCount:   googleParsedResponse.VolumeInfo.RatingsCount,
		Categories:    googleParsedResponse.VolumeInfo.Categories,
		Thumbnail:     googleParsedResponse.VolumeInfo.ImageLinks.Thumbnail,
	}

	return volumeInfo, nil
}
