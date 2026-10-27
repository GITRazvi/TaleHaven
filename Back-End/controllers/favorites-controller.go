package controllers

import (
	"TaleHaven-Backend/middlewares"
	anilistapi "TaleHaven-Backend/modules/anilist-api"
	"TaleHaven-Backend/modules/database"
	mangadexapi "TaleHaven-Backend/modules/mangadex-api"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FavoritesController exposes a user's AniList and MangaDex favorites.
// A title must already be saved in My Library before it can be favorited.
func FavoritesController(router *gin.Engine) {
	routes := router.Group("/favorites")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.GET("/anilist/characters", getAniListCharacterFavorites)
		routes.POST("/anilist/characters/:characterId", addAniListCharacterFavorite)
		routes.DELETE("/anilist/characters/:characterId", removeAniListCharacterFavorite)
		routes.GET("/anilist", getAniListFavorites)
		routes.POST("/anilist/:anilistId", addAniListFavorite)
		routes.DELETE("/anilist/:anilistId", removeAniListFavorite)
		routes.GET("/mangadex", getMangaDexFavorites)
		routes.POST("/mangadex/:mangaId", addMangaDexFavorite)
		routes.DELETE("/mangadex/:mangaId", removeMangaDexFavorite)
	}
}

func getAniListCharacterFavorites(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	favorites, err := database.GetAniListFavoriteCharacters(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve character favorites"})
		return
	}
	ids := make([]int, 0, len(favorites))
	for _, favorite := range favorites {
		ids = append(ids, favorite.CharacterID)
	}
	characters, err := aniListClient.GetCharactersByIDs(c.Request.Context(), ids)
	if err != nil {
		respondAniListUnavailable(c, err)
		return
	}
	charactersByID := make(map[int]anilistapi.Character, len(characters))
	for _, character := range characters {
		charactersByID[character.ID] = character
	}
	data := make([]gin.H, 0, len(favorites))
	for _, favorite := range favorites {
		character, found := charactersByID[favorite.CharacterID]
		if !found {
			continue
		}
		data = append(data, gin.H{
			"characterId": favorite.CharacterID,
			"mangaId":     favorite.SourceAniListMangaID,
			"role":        favorite.Role,
			"name":        character.Name,
			"image":       character.Image,
		})
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": data})
}

func addAniListCharacterFavorite(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	characterID, err := strconv.Atoi(c.Param("characterId"))
	if err != nil || characterID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList character ID"})
		return
	}
	var body struct {
		MangaID int    `json:"mangaId"`
		Role    string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.MangaID < 1 || len([]rune(body.Role)) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character favorite"})
		return
	}
	added, err := database.AddAniListCharacterToFavorites(userID, characterID, body.MangaID, body.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add character favorite"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "added": added})
}

func removeAniListCharacterFavorite(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	characterID, err := strconv.Atoi(c.Param("characterId"))
	if err != nil || characterID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList character ID"})
		return
	}
	removed, err := database.RemoveAniListCharacterFromFavorites(userID, characterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove character favorite"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "removed": removed})
}

func getFavoriteUserID(c *gin.Context) (int, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required", "message": "Please log in to manage favorites"})
		return 0, false
	}
	userID, ok := value.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authentication context", "message": "Could not identify the signed-in user"})
		return 0, false
	}
	return userID, true
}

func getAniListFavorites(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	ids, err := database.GetAniListFavoriteIDs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve favorites"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": ids})
}

func addAniListFavorite(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID"})
		return
	}
	saved, err := database.IsAniListMangaInUserLibrary(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify library membership"})
		return
	}
	if !saved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Add this manga to My Library before favoriting it"})
		return
	}
	added, err := database.AddAniListMangaToFavorites(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add favorite"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "added": added})
}

func removeAniListFavorite(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("anilistId"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AniList ID"})
		return
	}
	removed, err := database.RemoveAniListMangaFromFavorites(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove favorite"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "removed": removed})
}

func getMangaDexFavorites(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	ids, err := database.GetMangaDexFavoriteIDs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve favorites"})
		return
	}
	favorites, err := database.GetMangaDexFavoriteManga(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve favorites"})
		return
	}

	manga := make([]gin.H, 0, len(favorites))
	for _, favorite := range favorites {
		coverURL := ""
		if favorite.CoverFileName != "" {
			coverURL = mangadexapi.BuildCoverURL(favorite.MangaDexID, favorite.CoverFileName, "512")
		}
		manga = append(manga, gin.H{
			"mangaId":  favorite.MangaDexID,
			"title":    favorite.Title,
			"coverUrl": coverURL,
		})
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "data": ids, "manga": manga})
}

func addMangaDexFavorite(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing manga ID"})
		return
	}
	saved, err := database.IsMangaDexMangaInUserLibrary(userID, mangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify library membership"})
		return
	}
	if !saved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Add this manga to My Library before favoriting it"})
		return
	}
	response, err := mangadexapi.GetMangaByID(mangaID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to retrieve manga"})
		return
	}
	manga, err := mangadexapi.SaveMangaToDB(response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save manga"})
		return
	}
	added, err := database.AddMangaToFavorites(userID, manga.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add favorite"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "added": added})
}

func removeMangaDexFavorite(c *gin.Context) {
	userID, ok := getFavoriteUserID(c)
	if !ok {
		return
	}
	mangaID := c.Param("mangaId")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing manga ID"})
		return
	}
	removed, err := database.RemoveMangaFromFavorites(userID, mangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove favorite"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok", "removed": removed})
}
