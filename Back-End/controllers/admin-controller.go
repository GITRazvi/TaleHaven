package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminController exposes aggregate-only administration information. Future
// moderation tools can extend this route group while retaining the same
// server-side role boundary.
func AdminController(router *gin.Engine) {
	routes := router.Group("/admin")
	routes.Use(middlewares.AuthMiddleware(), middlewares.RequireAdmin())
	{
		routes.GET("/dashboard", getAdminDashboard)
	}
}

func getAdminDashboard(c *gin.Context) {
	stats, err := database.GetAdminDashboardStats()
	if err != nil {
		utils.Logger.Error("Failed to retrieve administrator dashboard statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dashboard unavailable", "message": "Unable to load dashboard statistics. Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"stats":  stats,
	})
}
