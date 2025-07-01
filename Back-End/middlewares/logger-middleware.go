package middlewares

import (
	"TaleHaven-Backend/modules/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

// LoggerMiddleware adds a logger to the Gin context
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		contextLogs := make(map[string]any)
		c.Set("contextLogs", contextLogs)

		// Clone the logger with request-specific fields
		logger := utils.Logger.With(
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)

		// Store the logger in Gin's context
		c.Set("logger", logger)

		// Proceed with the request
		c.Next()

		// Log the end of the request
		duration := time.Since(startTime)
		status := c.Writer.Status()
		logger.Info("REQUEST_COMPLETED",
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.Any("contextLogs", contextLogs),
		)
	}
}
