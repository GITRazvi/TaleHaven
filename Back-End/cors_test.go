package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSPatchPreflightIsAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(corsMiddleware())
	router.OPTIONS("/library/anilist/:anilistId/status", func(c *gin.Context) {})

	request := httptest.NewRequest(http.MethodOptions, "/library/anilist/105398/status", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if !strings.Contains(response.Header().Get("Access-Control-Allow-Methods"), http.MethodPatch) {
		t.Fatalf("PATCH missing from CORS methods: %q", response.Header().Get("Access-Control-Allow-Methods"))
	}
}
