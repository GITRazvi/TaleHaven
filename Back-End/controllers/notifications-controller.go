package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/database"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func NotificationsController(router *gin.Engine) {
	routes := router.Group("/notifications")
	routes.Use(middlewares.AuthMiddleware())
	routes.GET("", listNotifications)
	routes.GET("/unread-count", notificationUnreadCount)
	routes.PATCH("/:notificationID/read", markNotificationRead)
	routes.PATCH("/read-all", markAllNotificationsRead)
}

func notificationUserID(c *gin.Context) (int, bool) {
	value, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return 0, false
	}
	user, ok := value.(*database.User)
	if !ok || user.ID < 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return 0, false
	}
	return user.ID, true
}

func listNotifications(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}
	items, unread, err := database.GetNotifications(userID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": items, "unread": unread})
}

func notificationUnreadCount(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}
	_, unread, err := database.GetNotifications(userID, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "unread": unread})
}

func markNotificationRead(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("notificationID"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification"})
		return
	}
	if err := database.MarkNotificationRead(userID, id); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update notification"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func markAllNotificationsRead(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}
	if err := database.MarkAllNotificationsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}
