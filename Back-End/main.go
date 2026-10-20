package main

import (
	"TaleHaven-Backend/controllers"
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/cache"
	"TaleHaven-Backend/modules/cleanup"
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Serverul ruleaza la adresele urmatoare:
// http://127.0.0.1:3000
// http://localhost:3000
func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Keep Recovery middleware to handle panics

	utils.LoadEnv()

	utils.InitLogger()
	defer utils.Sync()
	router.Use(middlewares.LoggerMiddleware())

	cache.InitRedis()

	database.InitPostgres()
	defer database.CloseDB()

	// Add CORS middleware
	router.Use(corsMiddleware())

	// Start manga cleanup service (auto-deletes manga not accessed in 2 weeks)
	cleanupService := cleanup.NewMangaCleanupService()
	cleanupService.Start()
	defer cleanupService.Stop()

	addRoutes(router)

	utils.Logger.Info("Server Running on Port: 3001", zap.String("SERVER_ADDRESS", "127.0.0.1:3001"))

	_ = router.Run(":3001")
}

// corsMiddleware adds the CORS headers to allow cross-origin requests.
func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://127.0.0.1:3000": true,
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// Handle preflight requests (OPTIONS)
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func addRoutes(router *gin.Engine) {
	// Google Books API disabled - use MangaDex instead
	// To re-enable: uncomment the line below
	// controllers.BooksController(router)

	// Authentication endpoints
	controllers.AuthController(router)

	controllers.MangaDexController(router)
	controllers.LibraryController(router)
	controllers.ReadingProgressController(router)

	router.NoRoute(func(c *gin.Context) {
		resp := gin.H{
			"error":   "Route not found",
			"message": "The requested URL was not found on the server",
		}

		utils.AddLogsToContext(c, "response", resp)

		c.JSON(http.StatusNotFound, resp)
	})
}
