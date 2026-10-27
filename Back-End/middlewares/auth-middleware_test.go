package middlewares

import (
	"TaleHaven-Backend/modules/database"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTrustedRequestOrigins(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("FRONTEND_ORIGIN", "")
	if !IsTrustedRequestOrigin("http://localhost:3000") {
		t.Fatal("local frontend origin should be accepted during development")
	}
	if IsTrustedRequestOrigin("https://attacker.example") {
		t.Fatal("unconfigured origin must not be accepted")
	}

	t.Setenv("APP_ENV", "production")
	t.Setenv("FRONTEND_ORIGIN", "https://talehaven.example, https://www.talehaven.example")
	if !IsTrustedRequestOrigin("https://talehaven.example") {
		t.Fatal("configured production frontend origin should be accepted")
	}
	if IsTrustedRequestOrigin("http://localhost:3000") {
		t.Fatal("development fallback must not be accepted in production")
	}
}

func TestAuthCookieIsHttpOnlyAndSameSiteLax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("APP_ENV", "")
	t.Setenv("SESSION_COOKIE_SECURE", "false")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	SetAuthCookie(context, "opaque-session-token")

	response := &http.Response{Header: recorder.Header()}
	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != AuthCookieName || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected cookie settings: %#v", cookie)
	}
}

func TestRequireAdminRejectsNonAdminAndAllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, testCase := range []struct {
		name       string
		user       *database.User
		wantStatus int
	}{
		{name: "missing user", wantStatus: http.StatusForbidden},
		{name: "reader", user: &database.User{Role: "user"}, wantStatus: http.StatusForbidden},
		{name: "administrator", user: &database.User{Role: "admin"}, wantStatus: http.StatusNoContent},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if testCase.user != nil {
					c.Set("user", testCase.user)
				}
			})
			router.Use(RequireAdmin())
			router.GET("/admin/dashboard", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil))
			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, testCase.wantStatus)
			}
		})
	}
}
