package main

import (
	"TaleHaven-Backend/controllers"
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/cache"
	"TaleHaven-Backend/modules/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
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

	// Add CORS middleware
	router.Use(corsMiddleware())

	addRoutes(router)

	utils.Logger.Info("Server Running on Port: 3001", zap.String("SERVER_ADDRESS", "127.0.0.1:3001"))

	_ = router.Run(":3001")
}

// corsMiddleware adds the CORS headers to allow cross-origin requests.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// Handle preflight requests (OPTIONS)
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func addRoutes(router *gin.Engine) {
	controllers.BooksController(router)

	router.NoRoute(func(c *gin.Context) {
		resp := gin.H{
			"error":   "Route not found",
			"message": "The requested URL was not found on the server",
		}

		utils.AddLogsToContext(c, "response", resp)

		c.JSON(http.StatusNotFound, resp)
	})
}
