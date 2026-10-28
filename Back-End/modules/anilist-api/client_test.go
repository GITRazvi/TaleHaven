package anilist_api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestSearchMangaPostsVariablesAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", contentType)
		}

		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "query SearchManga") {
			t.Fatal("expected SearchManga query document")
		}
		if got := request.Variables["search"]; got != "Berserk" {
			t.Fatalf("search variable = %#v, want Berserk", got)
		}
		if got := request.Variables["genre"]; got != "Fantasy" {
			t.Fatalf("genre variable = %#v, want Fantasy", got)
		}
		if got := request.Variables["genres"]; !reflect.DeepEqual(got, []any{"Action", "Adventure"}) {
			t.Fatalf("genres variable = %#v, want Action and Adventure", got)
		}
		if got := request.Variables["tag"]; got != "Isekai" {
			t.Fatalf("tag variable = %#v, want Isekai", got)
		}
		if got := request.Variables["year"]; got != float64(1997) {
			t.Fatalf("year variable = %#v, want 1997", got)
		}
		if got := request.Variables["format"]; got != "MANGA" {
			t.Fatalf("format variable = %#v, want MANGA", got)
		}
		if got := request.Variables["status"]; got != "FINISHED" {
			t.Fatalf("status variable = %#v, want FINISHED", got)
		}
		if got := request.Variables["sort"]; !reflect.DeepEqual(got, []any{"SCORE_DESC"}) {
			t.Fatalf("sort variable = %#v, want SCORE_DESC", got)
		}
		if got := request.Variables["page"]; got != float64(2) {
			t.Fatalf("page variable = %#v, want 2", got)
		}
		if got := request.Variables["perPage"]; got != float64(12) {
			t.Fatalf("perPage variable = %#v, want 12", got)
		}

		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"currentPage":2,"hasNextPage":true,"perPage":12,"total":25},"media":[{"id":33,"idMal":2,"title":{"romaji":"Berserk","english":"Berserk"},"coverImage":{"large":"https://example.test/cover.jpg"},"genres":["Action"]}]}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	result, err := client.SearchManga(context.Background(), MangaSearchFilters{
		Title: "Berserk", Genre: "Fantasy", Genres: []string{"Action", "Adventure"}, Tag: "Isekai", Year: 1997, Format: "MANGA", Status: "FINISHED", Sort: "SCORE_DESC",
	}, 2, 12)
	if err != nil {
		t.Fatalf("SearchManga returned error: %v", err)
	}

	if result.PageInfo.Total != 25 || !result.PageInfo.HasNextPage {
		t.Fatalf("unexpected page info: %#v", result.PageInfo)
	}
	if len(result.Manga) != 1 || result.Manga[0].ID != 33 || result.Manga[0].CoverImage.Large != "https://example.test/cover.jpg" {
		t.Fatalf("unexpected manga: %#v", result.Manga)
	}
}

func TestSearchMangaByStaffIDUsesStaffMediaConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "query MangaByStaff") || !strings.Contains(request.Query, "staffMedia") {
			t.Fatalf("expected staff media query, got: %s", request.Query)
		}
		if got := request.Variables["staffId"]; got != float64(42) {
			t.Fatalf("staffId variable = %#v, want 42", got)
		}
		_, _ = w.Write([]byte(`{"data":{"Staff":{"staffMedia":{"pageInfo":{"currentPage":1,"hasNextPage":false,"perPage":12,"total":1},"nodes":[{"id":33,"title":{"romaji":"Berserk"}}]}}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	result, err := client.SearchManga(context.Background(), MangaSearchFilters{StaffID: 42}, 1, 12)
	if err != nil {
		t.Fatalf("SearchManga returned error: %v", err)
	}
	if len(result.Manga) != 1 || result.Manga[0].ID != 33 {
		t.Fatalf("unexpected manga: %#v", result.Manga)
	}
}

func TestSearchMangaOmitsUnsetFilterArguments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if strings.Contains(request.Query, "search:") || strings.Contains(request.Query, "$search") {
			t.Fatalf("query must omit empty search argument: %s", request.Query)
		}
		if !strings.Contains(request.Query, "genre: $genre") {
			t.Fatalf("query must include selected genre argument: %s", request.Query)
		}
		if _, exists := request.Variables["search"]; exists {
			t.Fatalf("variables must omit empty search: %#v", request.Variables)
		}
		if got := request.Variables["genre"]; got != "Action" {
			t.Fatalf("genre variable = %#v, want Action", got)
		}
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"currentPage":1,"hasNextPage":false,"perPage":3,"total":1},"media":[{"id":1,"title":{"romaji":"Action Story"}}]}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	result, err := client.SearchManga(context.Background(), MangaSearchFilters{Genre: "Action"}, 1, 3)
	if err != nil {
		t.Fatalf("SearchManga returned error: %v", err)
	}
	if result.PageInfo.Total != 1 || len(result.Manga) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestGetMangaByIDReturnsGraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":null,"errors":[{"message":"Too Many Requests.","status":429}]}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	_, err := client.GetMangaByID(context.Background(), 1)
	if err == nil {
		t.Fatal("GetMangaByID returned nil error")
	}

	var apiError *APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiError.Status != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", apiError.Status, http.StatusTooManyRequests)
	}
}

func TestGetMangaByIDRequestsAndDecodesCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "characters(page: 1, perPage: 10)") {
			t.Fatal("expected character connection in manga-details query")
		}
		if !strings.Contains(request.Query, "favourites") || !strings.Contains(request.Query, "rankings {") {
			t.Fatal("expected favourites and rankings in manga-details query")
		}
		_, _ = w.Write([]byte(`{"data":{"Media":{"id":1,"title":{"romaji":"Character Manga"},"favourites":4579,"rankings":[{"rank":802,"type":"RATED","allTime":true,"context":"highest rated all time"}],"characters":{"pageInfo":{"currentPage":1,"hasNextPage":false,"perPage":10,"total":1},"edges":[{"role":"MAIN","node":{"id":9,"name":{"full":"Main Hero"},"image":{"large":"https://example.test/hero.jpg"}}}]}}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	manga, err := client.GetMangaByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMangaByID returned error: %v", err)
	}
	if len(manga.Characters.Edges) != 1 || manga.Characters.Edges[0].Role != "MAIN" || manga.Characters.Edges[0].Node.Image.Large != "https://example.test/hero.jpg" {
		t.Fatalf("unexpected character data: %#v", manga.Characters)
	}
	if manga.Favourites != 4579 || len(manga.Rankings) != 1 || manga.Rankings[0].Rank != 802 {
		t.Fatalf("unexpected statistics data: favourites=%d rankings=%#v", manga.Favourites, manga.Rankings)
	}
}

func TestGetMangaCharactersRequestsAndDecodesPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "query MangaCharacters") {
			t.Fatal("expected character-page query")
		}
		if request.Variables["id"] != float64(1) || request.Variables["page"] != float64(2) || request.Variables["perPage"] != float64(10) {
			t.Fatalf("unexpected variables: %#v", request.Variables)
		}
		_, _ = w.Write([]byte(`{"data":{"Media":{"characters":{"pageInfo":{"currentPage":2,"hasNextPage":true,"perPage":10,"total":25},"edges":[{"role":"SUPPORTING","node":{"id":12,"name":{"full":"Second Page Hero"}}}]}}}}`))
	}))
	defer server.Close()

	characters, err := NewClientWithEndpoint(server.URL, server.Client()).GetMangaCharacters(context.Background(), 1, 2, 10)
	if err != nil {
		t.Fatalf("GetMangaCharacters returned error: %v", err)
	}
	if characters.PageInfo.CurrentPage != 2 || !characters.PageInfo.HasNextPage || len(characters.Edges) != 1 || characters.Edges[0].Node.ID != 12 {
		t.Fatalf("unexpected character page: %#v", characters)
	}
}

func TestGetMangaByIDRequestsAndDecodesRelatedManga(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "relations {") || !strings.Contains(request.Query, "relationType") {
			t.Fatal("expected related-media connection in manga-details query")
		}
		_, _ = w.Write([]byte(`{"data":{"Media":{"id":1,"title":{"romaji":"Origin"},"relations":{"edges":[{"relationType":"SEQUEL","node":{"id":2,"type":"MANGA","title":{"english":"Next Story"},"coverImage":{"medium":"https://example.test/next.jpg"},"format":"MANGA","status":"FINISHED","siteUrl":"https://anilist.co/manga/2"}}]}}}}`))
	}))
	defer server.Close()

	manga, err := NewClientWithEndpoint(server.URL, server.Client()).GetMangaByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMangaByID returned error: %v", err)
	}
	if len(manga.Relations.Edges) != 1 || manga.Relations.Edges[0].RelationType != "SEQUEL" || manga.Relations.Edges[0].Node.ID != 2 || manga.Relations.Edges[0].Node.Title.English != "Next Story" {
		t.Fatalf("unexpected related manga data: %#v", manga.Relations)
	}
}

func TestGetDiscoveryMangaUsesFixedSectionSort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "query MangaDiscovery") {
			t.Fatal("expected MangaDiscovery query document")
		}
		sort, ok := request.Variables["sort"].([]any)
		if !ok || len(sort) != 1 || sort[0] != "TRENDING_DESC" {
			t.Fatalf("sort variable = %#v, want [TRENDING_DESC]", request.Variables["sort"])
		}

		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"currentPage":1,"hasNextPage":false,"perPage":12,"total":1},"media":[{"id":33,"title":{"romaji":"Berserk"}}]}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	result, err := client.GetDiscoveryManga(context.Background(), DiscoveryTrending, 1, 12)
	if err != nil {
		t.Fatalf("GetDiscoveryManga returned error: %v", err)
	}
	if len(result.Manga) != 1 || result.Manga[0].Title.Romaji != "Berserk" {
		t.Fatalf("unexpected manga: %#v", result.Manga)
	}
}

func TestGetDiscoveryMangaRejectsUnknownSection(t *testing.T) {
	client := NewClientWithEndpoint("http://example.test", nil)
	_, err := client.GetDiscoveryManga(context.Background(), DiscoverySection("unknown"), 1, 12)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("error = %v, want unsupported section error", err)
	}
}

func TestGetMangaByGenrePostsGenreVariable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "query MangaGenre") {
			t.Fatal("expected MangaGenre query document")
		}
		if got := request.Variables["genre"]; got != "Fantasy" {
			t.Fatalf("genre variable = %#v, want Fantasy", got)
		}
		if got := request.Variables["page"]; got != float64(2) {
			t.Fatalf("page variable = %#v, want 2", got)
		}

		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"currentPage":2,"hasNextPage":false,"perPage":12,"total":1},"media":[{"id":34,"title":{"romaji":"Frieren"},"genres":["Adventure","Fantasy"]}]}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	result, err := client.GetMangaByGenre(context.Background(), " Fantasy ", 2, 12)
	if err != nil {
		t.Fatalf("GetMangaByGenre returned error: %v", err)
	}
	if len(result.Manga) != 1 || result.Manga[0].ID != 34 || result.Manga[0].Genres[1] != "Fantasy" {
		t.Fatalf("unexpected manga: %#v", result.Manga)
	}
}

func TestGetMangaByIDsBatchesLargeLibraryRequests(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		ids, ok := request.Variables["ids"].([]any)
		if !ok || len(ids) == 0 || len(ids) > 25 {
			t.Fatalf("ids variable = %#v, want between 1 and 25 IDs", request.Variables["ids"])
		}
		requests++
		_, _ = w.Write([]byte(`{"data":{"Page":{"media":[{"id":1,"title":{"romaji":"Berserk"}}]}}}`))
	}))
	defer server.Close()

	ids := make([]int, 26)
	for index := range ids {
		ids[index] = index + 1
	}
	client := NewClientWithEndpoint(server.URL, server.Client())
	manga, err := client.GetMangaByIDs(context.Background(), ids)
	if err != nil {
		t.Fatalf("GetMangaByIDs returned error: %v", err)
	}
	if requests != 2 || len(manga) != 2 {
		t.Fatalf("GetMangaByIDs requests/results = %d/%d, want 2/2", requests, len(manga))
	}
}

func TestGetCharactersByIDsRequestsAndDecodesCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "query CharactersByIDs") {
			t.Fatal("expected CharactersByIDs query document")
		}
		_, _ = w.Write([]byte(`{"data":{"Page":{"characters":[{"id":7,"name":{"full":"Favorite Hero"},"image":{"large":"https://example.test/hero.jpg"}}]}}}`))
	}))
	defer server.Close()

	characters, err := NewClientWithEndpoint(server.URL, server.Client()).GetCharactersByIDs(context.Background(), []int{7})
	if err != nil {
		t.Fatalf("GetCharactersByIDs returned error: %v", err)
	}
	if len(characters) != 1 || characters[0].ID != 7 || characters[0].Name.Full != "Favorite Hero" || characters[0].Image.Large != "https://example.test/hero.jpg" {
		t.Fatalf("unexpected characters: %#v", characters)
	}
}

func TestGetCharacterByIDRequestsAndDecodesDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(request.Query, "query CharacterDetails") {
			t.Fatal("expected CharacterDetails query document")
		}
		_, _ = w.Write([]byte(`{"data":{"Character":{"id":7,"name":{"full":"Hero"},"description":"A protagonist.","gender":"Male","media":{"pageInfo":{"currentPage":1,"hasNextPage":false,"perPage":25,"total":1},"edges":[{"characterRole":"MAIN","node":{"id":9,"type":"MANGA","title":{"english":"Origin Manga"}}}]}}}}`))
	}))
	defer server.Close()

	character, err := NewClientWithEndpoint(server.URL, server.Client()).GetCharacterByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetCharacterByID returned error: %v", err)
	}
	if character.ID != 7 || character.Description != "A protagonist." || len(character.Media.Edges) != 1 || character.Media.Edges[0].Node.Title.English != "Origin Manga" {
		t.Fatalf("unexpected character details: %#v", character)
	}
}

func TestSearchMangaValidatesFilters(t *testing.T) {
	client := NewClientWithEndpoint("http://example.test", nil)
	_, err := client.SearchManga(context.Background(), MangaSearchFilters{}, 1, 10)
	if err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Fatalf("error = %v, want missing-filter validation error", err)
	}
	_, err = client.SearchManga(context.Background(), MangaSearchFilters{Format: "INVALID"}, 1, 10)
	if err == nil || !strings.Contains(err.Error(), "format") {
		t.Fatalf("error = %v, want format validation error", err)
	}
}

func TestClientCachesEquivalentRequests(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"currentPage":1,"hasNextPage":false,"perPage":12,"total":1},"media":[{"id":33,"title":{"romaji":"Berserk"}}]}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	if _, err := client.GetDiscoveryManga(context.Background(), DiscoveryTrending, 1, 12); err != nil {
		t.Fatalf("first request: %v", err)
	}
	if _, err := client.GetDiscoveryManga(context.Background(), DiscoveryTrending, 1, 12); err != nil {
		t.Fatalf("cached request: %v", err)
	}
	if requests != 1 {
		t.Fatalf("upstream requests = %d, want 1", requests)
	}
}

func TestClientCoalescesConcurrentEquivalentRequests(t *testing.T) {
	requests := 0
	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		close(started)
		<-release
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"currentPage":1,"hasNextPage":false,"perPage":12,"total":1},"media":[{"id":33,"title":{"romaji":"Berserk"}}]}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	var group sync.WaitGroup
	group.Add(2)
	for range 2 {
		go func() {
			defer group.Done()
			if _, err := client.GetDiscoveryManga(context.Background(), DiscoveryTrending, 1, 12); err != nil {
				t.Errorf("discovery request: %v", err)
			}
		}()
	}
	<-started
	close(release)
	group.Wait()
	if requests != 1 {
		t.Fatalf("upstream requests = %d, want 1", requests)
	}
}

func TestClientRetriesOneRateLimitedRequest(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"data":null,"errors":[{"message":"Too Many Requests.","status":429}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"currentPage":1,"hasNextPage":false,"perPage":12,"total":1},"media":[{"id":33,"title":{"romaji":"Berserk"}}]}}}`))
	}))
	defer server.Close()

	client := NewClientWithEndpoint(server.URL, server.Client())
	client.retryDelay = 0
	if _, err := client.GetDiscoveryManga(context.Background(), DiscoveryTrending, 1, 12); err != nil {
		t.Fatalf("discovery request after retry: %v", err)
	}
	if requests != 2 {
		t.Fatalf("upstream requests = %d, want 2", requests)
	}
	if client.cacheTTL != metadataCacheTTL {
		t.Fatalf("cache TTL = %s, want %s", client.cacheTTL, metadataCacheTTL)
	}
}
