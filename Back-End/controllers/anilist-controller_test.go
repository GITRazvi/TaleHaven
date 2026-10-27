package controllers

import (
	anilist_api "TaleHaven-Backend/modules/anilist-api"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAniListTitleVariantsIncludesAllDistinctSynonyms(t *testing.T) {
	variants := aniListTitleVariants(anilist_api.MangaMetadata{
		Title: anilist_api.Title{English: "Main Title"},
		Synonyms: []string{
			"Alias One", "Alias Two", "Alias Three", "Alias Four",
			"Alias Five", "Alias Six", "main title",
		},
	})
	if len(variants) != 7 {
		t.Fatalf("variants = %#v, want main title and six distinct synonyms", variants)
	}
}

func TestAniListDiscoveryRejectsUnknownSectionBeforeCallingUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	AniListController(router)

	request := httptest.NewRequest(http.MethodGet, "/anilist/manga/unknown", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestAniListDiscoveryRejectsInvalidPaginationBeforeCallingUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	AniListController(router)

	request := httptest.NewRequest(http.MethodGet, "/anilist/manga/trending?perPage=26", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestAniListGenreRejectsInvalidInputBeforeCallingUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	AniListController(router)

	request := httptest.NewRequest(http.MethodGet, "/anilist/manga/genre/%20?page=0", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestAniListSearchRejectsMissingOrInvalidFiltersBeforeCallingUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	AniListController(router)

	for _, target := range []string{
		"/anilist/manga/search",
		"/anilist/manga/search?format=invalid",
		"/anilist/manga/search?year=1800",
		"/anilist/manga/search?year=not-a-year",
	} {
		request := httptest.NewRequest(http.MethodGet, target, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want %d", target, response.Code, http.StatusBadRequest)
		}
	}
}
