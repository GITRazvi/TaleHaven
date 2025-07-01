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

type googleSearchResponse struct {
	Items []googleBooksInfo `json:"items"`
}

type googleBooksInfo struct {
	Id         string         `json:"id"`
	VolumeInfo googleBookInfo `json:"volumeInfo"`
}

type googleBookInfo struct {
	Title         string           `json:"title"`
	Authors       []string         `json:"authors"`
	PublishedDate string           `json:"publishedDate"`
	Description   string           `json:"description"`
	AverageRating float32          `json:"averageRating"`
	RatingsCount  int              `json:"ratingsCount"`
	Categories    []string         `json:"categories"`
	ImageLinks    googleImageLinks `json:"imageLinks"`
}

type googleImageLinks struct {
	Thumbnail string `json:"thumbnail"`
}

type bookInfo struct {
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

// GET https://developers.google.com/books/docs/v1/getting_started
// SEARCH https://www.googleapis.com/books/v1/volumes?query=quilting

func SearchBooks(query string) ([]bookInfo, error) {
	body, err := getGoogleResponseFromCache(query)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		body, err = downloadBooksList(query)
		if err != nil {
			return nil, err
		}

		if setError := setGoogleResponseInCache(body, query); setError != nil {
			return nil, err
		}

	}

	var parsedResponse googleSearchResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		return nil, err
	}

	bookList, err := extractBooksListFromResponse(parsedResponse)

	return bookList, err
}

func downloadBooksList(query string) ([]byte, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("key", os.Getenv("GOOGLE_API_KEY"))
	requestURL := fmt.Sprintf("%s?%s", config.BaseUrl+"volumes", params.Encode())

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

func setGoogleResponseInCache(body []byte, query string) error {
	queryHash := utils.HashText(query)

	redisError := cache.Client.Set(cache.Ctx, "google:responses:/search/books:"+queryHash, body, 0).Err()
	if redisError != nil {
		utils.Logger.Error("Failed to set google response in redis", zap.Error(redisError))
		return redisError
	}

	return nil
}

func getGoogleResponseFromCache(query string) ([]byte, error) {
	queryHash := utils.HashText(query)

	responseFromCache, err := cache.Client.Get(cache.Ctx, "google:responses:/search/books:"+queryHash).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		utils.Logger.Error("Failed to get google response from cache", zap.Error(err))
		return nil, err
	}

	cachedResponseBytes := []byte(responseFromCache)

	return cachedResponseBytes, nil
}

func extractBooksListFromResponse(googleParsedResponse googleSearchResponse) ([]bookInfo, error) {
	var booksList []bookInfo

	//var err error
	for _, item := range googleParsedResponse.Items {
		publishedDate := item.VolumeInfo.PublishedDate
		if _, err := time.Parse("2006-01-02", item.VolumeInfo.PublishedDate); err != nil {
			utils.Logger.Error("PublishedDate is not in valid format", zap.Error(err))
			publishedDate = ""
		}

		bookList := bookInfo{
			Id:            item.Id,
			Title:         item.VolumeInfo.Title,
			Authors:       item.VolumeInfo.Authors,
			Description:   item.VolumeInfo.Description,
			PublishedDate: publishedDate,
			Rating:        item.VolumeInfo.AverageRating,
			RatingCount:   item.VolumeInfo.RatingsCount,
			Categories:    item.VolumeInfo.Categories,
			Thumbnail:     item.VolumeInfo.ImageLinks.Thumbnail,
		}

		booksList = append(booksList, bookList)

	}

	return booksList, nil
}
