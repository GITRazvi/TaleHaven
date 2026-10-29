package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeChatRoom(t *testing.T) {
	name, visibility, valid := normalizeChatRoom("  Tavern  ", "INVITE_ONLY")
	if !valid || name != "Tavern" || visibility != "invite_only" {
		t.Fatalf("normalize valid room = (%q, %q, %t)", name, visibility, valid)
	}
	for _, testCase := range []struct{ name, visibility string }{
		{name: "x", visibility: "public"},
		{name: "Valid", visibility: "password"},
	} {
		if _, _, valid := normalizeChatRoom(testCase.name, testCase.visibility); valid {
			t.Fatalf("normalize invalid room (%q, %q) succeeded", testCase.name, testCase.visibility)
		}
	}
}

func TestNormalizeChatVisibility(t *testing.T) {
	visibility, valid := normalizeChatVisibility(" PUBLIC ")
	if !valid || visibility != "public" {
		t.Fatalf("normalize public visibility = (%q, %t)", visibility, valid)
	}
	if _, valid := normalizeChatVisibility("password"); valid {
		t.Fatal("password visibility should be rejected")
	}
}

func TestChatInviteTokenHashIsStableAndSecretIsRandom(t *testing.T) {
	firstToken, firstHash, err := newChatInviteToken()
	if err != nil || len(firstToken) < 20 || firstHash != chatInviteTokenHash(firstToken) {
		t.Fatalf("first invite token = (%q, %q, %v)", firstToken, firstHash, err)
	}
	secondToken, _, err := newChatInviteToken()
	if err != nil || firstToken == secondToken {
		t.Fatalf("invite tokens should be distinct: %q and %q, err %v", firstToken, secondToken, err)
	}
}

func TestChatThemeBackgroundAllowsGIFWithoutExtendingOtherImageUploads(t *testing.T) {
	gifContents := []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff!\xf9\x04\x01\x00\x00\x00\x00,")
	if extension := chatImageExtension(gifContents, true); extension != ".gif" {
		t.Fatalf("GIF theme extension = %q, want .gif", extension)
	}
	if extension := chatImageExtension(gifContents, false); extension != "" {
		t.Fatalf("GIF should remain unavailable for non-theme uploads, got %q", extension)
	}
}

func TestThemeImageLimitIsFifteenMegabytes(t *testing.T) {
	if maxThemeImageBytes != 15*1024*1024 {
		t.Fatalf("theme image limit = %d, want 15 MB", maxThemeImageBytes)
	}
	if maxProfilePhotoBytes != 5*1024*1024 || maxChatImageBytes != 5*1024*1024 {
		t.Fatal("non-theme image limits must remain 5 MB")
	}
	if message := imageSizeLimitError(maxThemeImageBytes).Error(); message != "images must be 15 MB or smaller" {
		t.Fatalf("theme size message = %q", message)
	}
}

func TestChatModerationRoutesRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ChatController(router)
	request := httptest.NewRequest(http.MethodGet, "/chat/rooms/1/members", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
