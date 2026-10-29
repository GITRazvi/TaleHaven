package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/database"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const (
	defaultAniListMangaCommentLimit = 100
	maxAniListMangaCommentRunes     = 1000
)

type createAniListMangaCommentRequest struct {
	Body string `json:"body"`
}

// AniListMangaCommentsController owns the signed-in discussion endpoints.
// Metadata routes remain public; comments are deliberately session-bound.
func AniListMangaCommentsController(router *gin.Engine) {
	routes := router.Group("/anilist/manga")
	routes.Use(middlewares.AuthMiddleware())
	// The :section parameter name intentionally matches AniListController's
	// existing single-segment discovery route. Gin shares that route-tree node;
	// the resulting public URL remains /anilist/manga/:anilistId/comments.
	routes.GET("/:section/comments", listAniListMangaComments)
	routes.POST("/:section/comments", createAniListMangaComment)
	routes.DELETE("/:section/comments/:commentID", deleteAniListMangaComment)
}

func listAniListMangaComments(c *gin.Context) {
	if _, ok := aniListCommentUserID(c); !ok {
		return
	}
	aniListID, ok := aniListCommentMangaID(c)
	if !ok {
		return
	}
	comments, err := database.GetAniListMangaComments(aniListID, defaultAniListMangaCommentLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load manga comments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": comments})
}

func createAniListMangaComment(c *gin.Context) {
	userID, ok := aniListCommentUserID(c)
	if !ok {
		return
	}
	aniListID, ok := aniListCommentMangaID(c)
	if !ok {
		return
	}
	var request createAniListMangaCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment", "message": "A comment body is required."})
		return
	}
	body, valid := normalizeAniListMangaComment(request.Body)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment", "message": "Comments must be between 1 and 1,000 characters."})
		return
	}
	comment, err := database.CreateAniListMangaComment(aniListID, userID, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to post manga comment"})
		return
	}
	comment.AuthorDisplayName = chatDisplayName(c)
	comment.ProfileImagePath = chatProfileImagePath(c)
	c.JSON(http.StatusCreated, gin.H{"result": "ok", "comment": comment})
}

func deleteAniListMangaComment(c *gin.Context) {
	userID, ok := aniListCommentUserID(c)
	if !ok {
		return
	}
	aniListID, ok := aniListCommentMangaID(c)
	if !ok {
		return
	}
	commentID, err := strconv.ParseInt(c.Param("commentID"), 10, 64)
	if err != nil || commentID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	if err := database.DeleteAniListMangaComment(aniListID, userID, commentID); errors.Is(err, database.ErrAniListMangaCommentDenied) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Comment deletion denied"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete manga comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "message": "Comment deleted"})
}

func aniListCommentUserID(c *gin.Context) (int, bool) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return 0, false
	}
	return userID, true
}

func aniListCommentMangaID(c *gin.Context) (int, bool) {
	aniListID, err := strconv.Atoi(c.Param("section"))
	if err != nil || aniListID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID"})
		return 0, false
	}
	return aniListID, true
}

func normalizeAniListMangaComment(value string) (string, bool) {
	value = strings.TrimSpace(value)
	return value, value != "" && utf8.RuneCountInString(value) <= maxAniListMangaCommentRunes
}
