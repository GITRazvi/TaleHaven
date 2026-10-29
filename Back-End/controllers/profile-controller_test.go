package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"TaleHaven-Backend/modules/database"
	"github.com/gin-gonic/gin"
)

func TestPublicProfileRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/profiles/not-a-number", nil)
	context.Params = []gin.Param{{Key: "userID", Value: "not-a-number"}}

	getPublicProfile(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestPublicProfileResponseExcludesAccountFields(t *testing.T) {
	response := toPublicProfileResponse(&database.User{
		ID:               7,
		Email:            "private@example.test",
		Role:             "admin",
		DisplayName:      "Chat Friend",
		ProfileImagePath: "/images/profiles/friend.jpg",
		PasswordHash:     "secret-hash",
		CreatedAt:        time.Date(2026, time.October, 21, 0, 0, 0, 0, time.UTC),
	}, true)

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal public profile: %v", err)
	}
	for _, forbidden := range []string{"private@example.test", "admin", "secret-hash", "password"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("public profile leaked %q: %s", forbidden, encoded)
		}
	}
}
