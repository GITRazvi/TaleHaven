package anilist_api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultEndpoint     = "https://graphql.anilist.co"
	defaultTimeout      = 15 * time.Second
	maxResponseSize     = 2 << 20
	metadataCacheTTL    = 2 * time.Minute
	requestInterval     = 750 * time.Millisecond
	rateLimitBackoff    = time.Second
	maxRetryDelay       = 3 * time.Second
	maxRateLimitRetries = 1
)

// Client retrieves public AniList metadata through the GraphQL API.
// It owns no credentials because public manga queries do not require OAuth.
type Client struct {
	endpoint        string
	httpClient      *http.Client
	cacheTTL        time.Duration
	requestInterval time.Duration
	retryDelay      time.Duration

	mutex         sync.Mutex
	cache         map[string]cachedResponse
	inFlight      map[string]*inFlightRequest
	nextRequestAt time.Time
}

type cachedResponse struct {
	data      json.RawMessage
	expiresAt time.Time
}

type inFlightRequest struct {
	done chan struct{}
	data json.RawMessage
	err  error
}

// NewClient returns a client configured for AniList's public GraphQL endpoint.
func NewClient() *Client {
	client := NewClientWithEndpoint(DefaultEndpoint, &http.Client{Timeout: defaultTimeout})
	client.requestInterval = requestInterval
	client.retryDelay = rateLimitBackoff
	return client
}

// NewClientWithEndpoint allows tests to provide a local GraphQL endpoint.
func NewClientWithEndpoint(endpoint string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	return &Client{
		endpoint:   endpoint,
		httpClient: httpClient,
		cacheTTL:   metadataCacheTTL,
		cache:      make(map[string]cachedResponse),
		inFlight:   make(map[string]*inFlightRequest),
	}
}

// SearchManga finds public manga metadata using the supplied catalog filters.
func (c *Client) SearchManga(ctx context.Context, filters MangaSearchFilters, page, perPage int) (*MangaSearchResult, error) {
	filters, err := filters.Normalize()
	if err != nil {
		return nil, err
	}
	if page < 1 {
		return nil, fmt.Errorf("AniList page must be at least 1")
	}
	if perPage < 1 || perPage > 50 {
		return nil, fmt.Errorf("AniList perPage must be between 1 and 50")
	}
	if filters.StaffID > 0 {
		return c.GetMangaByStaffID(ctx, filters.StaffID, page, perPage)
	}

	var data struct {
		Page struct {
			PageInfo Pagination      `json:"pageInfo"`
			Media    []MangaMetadata `json:"media"`
		} `json:"Page"`
	}
	variables := map[string]any{
		"sort":    []string{filters.Sort},
		"page":    page,
		"perPage": perPage,
	}
	if filters.Title != "" {
		variables["search"] = filters.Title
	}
	if filters.Genre != "" {
		variables["genre"] = filters.Genre
	}
	if len(filters.Genres) > 0 {
		variables["genres"] = filters.Genres
	}
	if filters.Tag != "" {
		variables["tag"] = filters.Tag
	}
	if filters.Year != 0 {
		variables["year"] = filters.Year
	}
	if filters.Format != "" {
		variables["format"] = filters.Format
	}
	if filters.Status != "" {
		variables["status"] = filters.Status
	}
	if err := c.execute(ctx, SearchMangaQuery(filters), variables, &data); err != nil {
		return nil, err
	}

	return &MangaSearchResult{PageInfo: data.Page.PageInfo, Manga: data.Page.Media}, nil
}

// GetMangaByStaffID retrieves manga credited to one AniList staff member.
func (c *Client) GetMangaByStaffID(ctx context.Context, staffID, page, perPage int) (*MangaSearchResult, error) {
	if staffID < 1 {
		return nil, fmt.Errorf("AniList staff ID must be positive")
	}
	var data struct {
		Staff struct {
			StaffMedia struct {
				PageInfo Pagination      `json:"pageInfo"`
				Nodes    []MangaMetadata `json:"nodes"`
			} `json:"staffMedia"`
		} `json:"Staff"`
	}
	variables := map[string]any{"staffId": staffID, "page": page, "perPage": perPage}
	if err := c.execute(ctx, MangaByStaffQuery, variables, &data); err != nil {
		return nil, err
	}
	return &MangaSearchResult{PageInfo: data.Staff.StaffMedia.PageInfo, Manga: data.Staff.StaffMedia.Nodes}, nil
}

// GetMangaByID retrieves public metadata for one AniList manga media ID.
func (c *Client) GetMangaByID(ctx context.Context, id int) (*MangaMetadata, error) {
	if id < 1 {
		return nil, fmt.Errorf("AniList manga ID must be positive")
	}

	var data struct {
		Media MangaMetadata `json:"Media"`
	}
	if err := c.execute(ctx, MangaDetailsQuery, map[string]any{"id": id}, &data); err != nil {
		return nil, err
	}

	return &data.Media, nil
}

// GetMangaCharacters retrieves one paged set of story characters for a manga.
func (c *Client) GetMangaCharacters(ctx context.Context, id, page, perPage int) (*CharacterConnection, error) {
	if id < 1 {
		return nil, fmt.Errorf("AniList manga ID must be positive")
	}
	if page < 1 {
		return nil, fmt.Errorf("AniList character page must be at least 1")
	}
	if perPage < 1 || perPage > 25 {
		return nil, fmt.Errorf("AniList character perPage must be between 1 and 25")
	}

	var data struct {
		Media struct {
			Characters CharacterConnection `json:"characters"`
		} `json:"Media"`
	}
	if err := c.execute(ctx, MangaCharactersQuery, map[string]any{
		"id": id, "page": page, "perPage": perPage,
	}, &data); err != nil {
		return nil, err
	}
	return &data.Media.Characters, nil
}

// GetDiscoveryManga retrieves one paged AniList manga discovery section.
func (c *Client) GetDiscoveryManga(ctx context.Context, section DiscoverySection, page, perPage int) (*MangaSearchResult, error) {
	sort, supported := section.Sort()
	if !supported {
		return nil, fmt.Errorf("unsupported AniList discovery section %q", section)
	}
	if page < 1 {
		return nil, fmt.Errorf("AniList page must be at least 1")
	}
	if perPage < 1 || perPage > 25 {
		return nil, fmt.Errorf("AniList perPage must be between 1 and 25")
	}

	var data struct {
		Page struct {
			PageInfo Pagination      `json:"pageInfo"`
			Media    []MangaMetadata `json:"media"`
		} `json:"Page"`
	}
	if err := c.execute(ctx, MangaDiscoveryQuery, map[string]any{
		"page":    page,
		"perPage": perPage,
		"sort":    sort,
	}, &data); err != nil {
		return nil, err
	}

	return &MangaSearchResult{PageInfo: data.Page.PageInfo, Manga: data.Page.Media}, nil
}

// GetMangaByGenre retrieves public manga metadata for one AniList genre,
// ordered by AniList popularity.
func (c *Client) GetMangaByGenre(ctx context.Context, genre string, page, perPage int) (*MangaSearchResult, error) {
	genre = strings.TrimSpace(genre)
	if genre == "" {
		return nil, fmt.Errorf("AniList genre is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("AniList page must be at least 1")
	}
	if perPage < 1 || perPage > 25 {
		return nil, fmt.Errorf("AniList perPage must be between 1 and 25")
	}

	var data struct {
		Page struct {
			PageInfo Pagination      `json:"pageInfo"`
			Media    []MangaMetadata `json:"media"`
		} `json:"Page"`
	}
	if err := c.execute(ctx, MangaGenreQuery, map[string]any{
		"genre":   genre,
		"page":    page,
		"perPage": perPage,
	}, &data); err != nil {
		return nil, err
	}

	return &MangaSearchResult{PageInfo: data.Page.PageInfo, Manga: data.Page.Media}, nil
}

func (c *Client) GetMangaByIDs(ctx context.Context, ids []int) ([]MangaMetadata, error) {
	if len(ids) == 0 {
		return []MangaMetadata{}, nil
	}
	allManga := make([]MangaMetadata, 0, len(ids))
	for start := 0; start < len(ids); start += 25 {
		end := start + 25
		if end > len(ids) {
			end = len(ids)
		}
		manga, err := c.getMangaByIDBatch(ctx, ids[start:end])
		if err != nil {
			return nil, err
		}
		allManga = append(allManga, manga...)
	}
	return allManga, nil
}

func (c *Client) getMangaByIDBatch(ctx context.Context, ids []int) ([]MangaMetadata, error) {
	var data struct {
		Page struct {
			Media []MangaMetadata `json:"media"`
		} `json:"Page"`
	}
	if err := c.execute(ctx, MangaByIDsQuery, map[string]any{"ids": ids}, &data); err != nil {
		return nil, err
	}
	return data.Page.Media, nil
}

// GetCharactersByIDs retrieves public display metadata for saved AniList
// characters. Requests are batched to keep profile loading bounded.
func (c *Client) GetCharactersByIDs(ctx context.Context, ids []int) ([]Character, error) {
	if len(ids) == 0 {
		return []Character{}, nil
	}
	allCharacters := make([]Character, 0, len(ids))
	for start := 0; start < len(ids); start += 50 {
		end := start + 50
		if end > len(ids) {
			end = len(ids)
		}
		var data struct {
			Page struct {
				Characters []Character `json:"characters"`
			} `json:"Page"`
		}
		if err := c.execute(ctx, CharactersByIDsQuery, map[string]any{"ids": ids[start:end]}, &data); err != nil {
			return nil, err
		}
		allCharacters = append(allCharacters, data.Page.Characters...)
	}
	return allCharacters, nil
}

// GetCharacterByID retrieves the fuller public profile displayed after a
// reader selects a character.
func (c *Client) GetCharacterByID(ctx context.Context, id int) (*Character, error) {
	if id < 1 {
		return nil, fmt.Errorf("AniList character ID must be positive")
	}
	var data struct {
		Character Character `json:"Character"`
	}
	if err := c.execute(ctx, CharacterDetailsQuery, map[string]any{"id": id}, &data); err != nil {
		return nil, err
	}
	return &data.Character, nil
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

// GraphQLError is an upstream GraphQL error reported in a response body.
type GraphQLError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// APIError contains the upstream HTTP status and any GraphQL errors returned
// by AniList. It lets a future TaleHaven controller distinguish rate limits.
type APIError struct {
	Status int
	Errors []GraphQLError
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 && e.Errors[0].Message != "" {
		return fmt.Sprintf("AniList API error (status %d): %s", e.Status, e.Errors[0].Message)
	}
	return fmt.Sprintf("AniList API error (status %d)", e.Status)
}

func (c *Client) execute(ctx context.Context, query string, variables map[string]any, target any) error {
	payload, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("encode AniList GraphQL request: %w", err)
	}
	key := string(payload)
	data, err := c.getResponseData(ctx, key, payload)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode AniList GraphQL data: %w", err)
	}
	return nil
}

func (c *Client) getResponseData(ctx context.Context, key string, payload []byte) (json.RawMessage, error) {
	now := time.Now()
	c.mutex.Lock()
	if cached, found := c.cache[key]; found {
		if now.Before(cached.expiresAt) {
			data := append(json.RawMessage(nil), cached.data...)
			c.mutex.Unlock()
			return data, nil
		}
		delete(c.cache, key)
	}
	if request, found := c.inFlight[key]; found {
		done := request.done
		c.mutex.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
			if request.err != nil {
				return nil, request.err
			}
			return append(json.RawMessage(nil), request.data...), nil
		}
	}
	request := &inFlightRequest{done: make(chan struct{})}
	c.inFlight[key] = request
	c.mutex.Unlock()

	data, requestErr := c.requestWithRetry(ctx, payload)

	c.mutex.Lock()
	request.data = data
	request.err = requestErr
	if requestErr == nil {
		c.cache[key] = cachedResponse{
			data:      append(json.RawMessage(nil), data...),
			expiresAt: time.Now().Add(c.cacheTTL),
		}
	}
	delete(c.inFlight, key)
	close(request.done)
	c.mutex.Unlock()

	if requestErr != nil {
		return nil, requestErr
	}
	return data, nil
}

func (c *Client) requestWithRetry(ctx context.Context, payload []byte) (json.RawMessage, error) {
	for attempt := 0; ; attempt++ {
		data, retryAfter, err := c.request(ctx, payload)
		if !isRateLimitError(err) || attempt >= maxRateLimitRetries {
			return data, err
		}
		if retryAfter <= 0 {
			retryAfter = c.retryDelay
		}
		if retryAfter > maxRetryDelay {
			retryAfter = maxRetryDelay
		}
		if err := waitFor(ctx, retryAfter); err != nil {
			return nil, err
		}
	}
}

func (c *Client) request(ctx context.Context, payload []byte) (json.RawMessage, time.Duration, error) {
	if err := c.waitForRequestSlot(ctx); err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("create AniList GraphQL request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TaleHaven-Backend/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("send AniList GraphQL request: %w", err)
	}
	defer resp.Body.Close()
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, retryAfter, fmt.Errorf("read AniList GraphQL response: %w", err)
	}

	var response graphQLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, retryAfter, fmt.Errorf("decode AniList GraphQL response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices || len(response.Errors) > 0 {
		status := resp.StatusCode
		for _, graphQLError := range response.Errors {
			if graphQLError.Status != 0 {
				status = graphQLError.Status
				break
			}
		}
		return nil, retryAfter, &APIError{Status: status, Errors: response.Errors}
	}
	return response.Data, retryAfter, nil
}

func (c *Client) waitForRequestSlot(ctx context.Context) error {
	if c.requestInterval <= 0 {
		return nil
	}
	c.mutex.Lock()
	now := time.Now()
	requestAt := now
	if c.nextRequestAt.After(now) {
		requestAt = c.nextRequestAt
	}
	c.nextRequestAt = requestAt.Add(c.requestInterval)
	c.mutex.Unlock()
	return waitFor(ctx, time.Until(requestAt))
}

func waitFor(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isRateLimitError(err error) bool {
	var apiError *APIError
	return errors.As(err, &apiError) && apiError.Status == http.StatusTooManyRequests
}

func parseRetryAfter(value string) time.Duration {
	seconds, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}
