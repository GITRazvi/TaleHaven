package sessions

import (
	"TaleHaven-Backend/modules/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	SessionCookieName = "talehaven_session"
)

// SessionMiddleware manages user sessions automatically
func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get session from cookie
		sessionID, err := c.Cookie(SessionCookieName)

		var session *Session

		if err == nil && sessionID != "" {
			// Validate existing session
			session, err = GetSession(sessionID)
			if err != nil {
				utils.Logger.Error("Failed to get session", zap.Error(err))
			}

			// Check if session expired
			if session != nil && session.IsExpired() {
				// Delete expired session
				DeleteSession(session.ID)
				session = nil
			}
		}

		// Create new session if needed
		if session == nil {
			ipAddress := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")

			session, err = CreateSession(ipAddress, userAgent)
			if err != nil {
				utils.Logger.Error("Failed to create session", zap.Error(err))
				c.Next()
				return
			}

			// Set cookie (30 days)
			c.SetCookie(
				SessionCookieName,
				session.ID,
				int(SessionDuration.Seconds()),
				"/",
				"",
				false, // Set to true in production with HTTPS
				true,  // HttpOnly
			)

			utils.Logger.Info("New session created",
				zap.String("sessionId", session.ID),
				zap.String("ip", ipAddress))
		} else {
			// Update session activity
			err = UpdateSessionActivity(session.ID)
			if err != nil {
				utils.Logger.Error("Failed to update session activity", zap.Error(err))
			}
		}

		// Store session in context
		c.Set("session", session)
		c.Set("sessionId", session.ID)

		c.Next()
	}
}

// GetSessionFromContext retrieves the session from gin context
func GetSessionFromContext(c *gin.Context) *Session {
	if session, exists := c.Get("session"); exists {
		return session.(*Session)
	}
	return nil
}

// GetSessionIDFromContext retrieves the session ID from gin context
func GetSessionIDFromContext(c *gin.Context) string {
	if sessionID, exists := c.Get("sessionId"); exists {
		return sessionID.(string)
	}
	return ""
}

// RequireSession middleware ensures a session exists
func RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSessionFromContext(c)
		if session == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "No session",
				"message": "Session required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

