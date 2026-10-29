package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMangaDexBrowseRejectsInvalidPaginationBeforeCallingUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/manga/browse", browseMangaAlphabetically)

	for _, target := range []string{
		"/manga/browse?limit=0",
		"/manga/browse?limit=101",
		"/manga/browse?offset=-1",
	} {
		request := httptest.NewRequest(http.MethodGet, target, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want %d", target, response.Code, http.StatusBadRequest)
		}
	}
}
