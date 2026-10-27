package middlewares

import (
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/sessions"
	"TaleHaven-Backend/modules/utils"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	AuthCookieName = "talehaven_session"
)

// AuthMiddleware validates a revocable server-side cookie session and makes
// the authenticated user available to handlers.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isUnsafeMethod(c.Request.Method) && !IsTrustedRequestOrigin(c.Request.Header.Get("Origin")) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid request origin", "message": "This request origin is not allowed"})
			c.Abort()
			return
		}

		token, err := c.Cookie(AuthCookieName)
		if err != nil {
			unauthorized(c, "Authentication required", "You must be logged in to access this resource")
			return
		}

		session, user, err := sessions.AuthenticateSession(token)
		if err != nil {
			ClearAuthCookie(c)
			if errors.Is(err, sessions.ErrExpiredAuthenticatedSession) {
				unauthorized(c, "Session expired", "Your session has expired. Please log in again")
			} else {
				utils.Logger.Warn("Invalid session", zap.Error(err), zap.String("path", c.Request.URL.Path))
				unauthorized(c, "Invalid session", "Your session is invalid. Please log in again")
			}
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("session_id", session.ID)

		c.Next()
	}
}

// OptionalAuthMiddleware loads user data if authenticated but doesn't require it
// Useful for endpoints that behave differently for authenticated vs anonymous users
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(AuthCookieName)
		if err != nil {
			c.Next()
			return
		}

		session, user, err := sessions.AuthenticateSession(token)
		if err != nil {
			ClearAuthCookie(c)
			c.Next()
			return
		}
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("session_id", session.ID)

		c.Next()
	}
}

// RequireAdmin ensures the authenticated session belongs to an administrator.
// It must be placed after AuthMiddleware so the trusted user record is already
// present in the request context.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("user")
		user, ok := value.(*database.User)
		if !exists || !ok || user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Administrator access required", "message": "You do not have permission to access this resource"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireTrustedOrigin protects state-changing browser requests that use
// cookies. A missing origin is rejected rather than guessed.
func RequireTrustedOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isUnsafeMethod(c.Request.Method) && !IsTrustedRequestOrigin(c.Request.Header.Get("Origin")) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid request origin", "message": "This request origin is not allowed"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func isUnsafeMethod(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}

// IsTrustedRequestOrigin accepts only configured UI origins. In production,
// FRONTEND_ORIGIN must be configured; local origins are development-only.
func IsTrustedRequestOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	allowed := map[string]struct{}{}
	if configured := os.Getenv("FRONTEND_ORIGIN"); configured != "" {
		for _, value := range strings.Split(configured, ",") {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				allowed[trimmed] = struct{}{}
			}
		}
	} else if os.Getenv("APP_ENV") != "production" {
		allowed["http://localhost:3000"] = struct{}{}
		allowed["http://127.0.0.1:3000"] = struct{}{}
	}
	_, ok := allowed[origin]
	return ok
}

func SetAuthCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AuthCookieName, token, int(sessions.AuthenticatedSessionAbsoluteDuration.Seconds()), "/", "", sessionCookieSecure(), true)
}

func ClearAuthCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AuthCookieName, "", -1, "/", "", sessionCookieSecure(), true)
}

func sessionCookieSecure() bool {
	if configured := os.Getenv("SESSION_COOKIE_SECURE"); configured != "" {
		return strings.EqualFold(configured, "true")
	}
	return os.Getenv("APP_ENV") == "production"
}

func unauthorized(c *gin.Context, errorText, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": errorText, "message": message})
	c.Abort()
}
