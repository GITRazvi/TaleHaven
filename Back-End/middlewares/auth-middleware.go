package middlewares

import (
	"TaleHaven-Backend/modules/auth"
	jwtModule "TaleHaven-Backend/modules/jwt"
	"TaleHaven-Backend/modules/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	AuthCookieName = "talehaven_auth_token"
)

// AuthMiddleware checks if the user is authenticated via JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get JWT token from cookie
		tokenString, err := c.Cookie(AuthCookieName)
		if err != nil {
			utils.Logger.Info("No auth token found",
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "You must be logged in to access this resource",
			})
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := jwtModule.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, jwtModule.ErrExpiredToken) {
				utils.Logger.Info("Expired token",
					zap.String("path", c.Request.URL.Path))
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Token expired",
					"message": "Your session has expired. Please log in again",
				})
			} else {
				utils.Logger.Warn("Invalid token",
					zap.Error(err),
					zap.String("path", c.Request.URL.Path))
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Invalid token",
					"message": "Your session is invalid. Please log in again",
				})
			}
			c.Abort()
			return
		}

		// Load full user data
		user, err := auth.GetUserByID(claims.UserID)
		if err != nil {
			utils.Logger.Error("Failed to load user from token",
				zap.Error(err),
				zap.Int("user_id", claims.UserID))

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "Your session is invalid. Please log in again",
			})
			c.Abort()
			return
		}

		// Add user to context for use in handlers
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("jwt_claims", claims)

		utils.Logger.Debug("User authenticated successfully via JWT",
			zap.Int("user_id", user.ID),
			zap.String("email", user.Email),
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}

// OptionalAuthMiddleware loads user data if authenticated but doesn't require it
// Useful for endpoints that behave differently for authenticated vs anonymous users
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get JWT token from cookie
		tokenString, err := c.Cookie(AuthCookieName)
		if err != nil {
			c.Next()
			return
		}

		// Validate JWT token
		claims, err := jwtModule.ValidateToken(tokenString)
		if err != nil {
			utils.Logger.Debug("Optional auth failed - continuing without auth",
				zap.Error(err))
			c.Next()
			return
		}

		// Load user data
		user, err := auth.GetUserByID(claims.UserID)
		if err != nil {
			utils.Logger.Warn("Failed to load user from token in optional auth",
				zap.Error(err),
				zap.Int("user_id", claims.UserID))
			c.Next()
			return
		}

		// Add user to context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

