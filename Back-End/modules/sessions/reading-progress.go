package sessions

import (
	"TaleHaven-Backend/modules/cache"
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	// Redis TTL for reading progress (30 minutes of inactivity)
	ReadingProgressTTL = 30 * time.Minute
)

// ReadingProgress represents a user's reading progress for a manga chapter
type ReadingProgress struct {
	UserID      int       `json:"userId"`
	MangaID     string    `json:"mangaId"`
	ChapterID   string    `json:"chapterId"`
	CurrentPage int       `json:"currentPage"`
	TotalPages  int       `json:"totalPages,omitempty"`
	LastReadAt  time.Time `json:"lastReadAt"`
}

// MangaProgress represents all reading progress for a specific manga
type MangaProgress struct {
	MangaID   string                      `json:"mangaId"`
	Chapters  map[string]*ReadingProgress `json:"chapters"` // key: chapterID
	UpdatedAt time.Time                   `json:"updatedAt"`
}

// UserReadingData represents all reading progress for a user
type UserReadingData struct {
	UserID    int                       `json:"userId"`
	Mangas    map[string]*MangaProgress `json:"mangas"` // key: mangaID
	UpdatedAt time.Time                 `json:"updatedAt"`
}

// ============================================================================
// MAIN API - Used by controllers
// ============================================================================

// GetReadingProgress retrieves reading progress for a specific chapter
// Flow: Redis first → PostgreSQL fallback → warm Redis
func GetReadingProgress(userID int, mangaID, chapterID string) (*ReadingProgress, error) {
	if userID <= 0 || mangaID == "" || chapterID == "" {
		return nil, fmt.Errorf("userID, mangaID, and chapterID are required")
	}

	// Try Redis first (hot cache)
	progress, err := getFromRedis(userID, chapterID)
	if err == nil && progress != nil {
		utils.Logger.Debug("Reading progress found in Redis",
			zap.Int("userId", userID),
			zap.String("chapterId", chapterID))
		return progress, nil
	}

	// Redis miss - check PostgreSQL (persistent storage)
	progress, err = getFromPostgres(userID, chapterID)
	if err != nil {
		return nil, err
	}

	if progress != nil {
		// Warm Redis cache for future requests
		_ = saveToRedis(progress)
		utils.Logger.Debug("Reading progress loaded from PostgreSQL and cached",
			zap.Int("userId", userID),
			zap.String("chapterId", chapterID))
		return progress, nil
	}

	// Not found anywhere - user hasn't read this chapter
	return nil, nil
}

// SaveReadingProgress saves reading progress (Redis only - fast writes)
// PostgreSQL sync happens separately via SyncReadingProgress
func SaveReadingProgress(userID int, mangaID, chapterID string, currentPage, totalPages int) error {
	if userID <= 0 || mangaID == "" || chapterID == "" {
		return fmt.Errorf("userID, mangaID, and chapterID are required")
	}

	if currentPage < 1 {
		return fmt.Errorf("currentPage must be at least 1")
	}

	progress := &ReadingProgress{
		UserID:      userID,
		MangaID:     mangaID,
		ChapterID:   chapterID,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		LastReadAt:  time.Now(),
	}

	// Save to Redis only (fast write)
	err := saveToRedis(progress)
	if err != nil {
		return fmt.Errorf("failed to save reading progress to Redis: %w", err)
	}

	utils.Logger.Debug("Reading progress saved to Redis",
		zap.Int("userId", userID),
		zap.String("mangaId", mangaID),
		zap.String("chapterId", chapterID),
		zap.Int("currentPage", currentPage))

	return nil
}

// SyncReadingProgress syncs Redis data to PostgreSQL (persistent storage)
// Call this when: user leaves chapter, closes tab, opens another chapter
func SyncReadingProgress(userID int, chapterID string) error {
	if userID <= 0 || chapterID == "" {
		return fmt.Errorf("userID and chapterID are required")
	}

	// Get current state from Redis
	progress, err := getFromRedis(userID, chapterID)
	if err != nil || progress == nil {
		// Nothing to sync (Redis empty or error)
		return nil
	}

	// Upsert to PostgreSQL
	err = saveToPostgres(progress)
	if err != nil {
		return fmt.Errorf("failed to sync reading progress to PostgreSQL: %w", err)
	}

	utils.Logger.Info("Reading progress synced to PostgreSQL",
		zap.Int("userId", userID),
		zap.String("mangaId", progress.MangaID),
		zap.String("chapterId", chapterID),
		zap.Int("currentPage", progress.CurrentPage))

	return nil
}

// ============================================================================
// BATCH OPERATIONS
// ============================================================================

// GetMangaProgress retrieves all reading progress for a specific manga
func GetMangaProgress(userID int, mangaID string) (*MangaProgress, error) {
	if userID <= 0 || mangaID == "" {
		return nil, fmt.Errorf("userID and mangaID are required")
	}

	// Get all progress from PostgreSQL (persistent storage)
	query := `
		SELECT chapter_id, current_page, total_pages, last_read_at
		FROM user_reading_progress
		WHERE user_id = $1 AND manga_id = $2
		ORDER BY last_read_at DESC
	`

	rows, err := database.DB.Query(query, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manga progress: %w", err)
	}
	defer rows.Close()

	mangaProgress := &MangaProgress{
		MangaID:   mangaID,
		Chapters:  make(map[string]*ReadingProgress),
		UpdatedAt: time.Now(),
	}

	for rows.Next() {
		var chapterID string
		var currentPage, totalPages int
		var lastReadAt time.Time

		err := rows.Scan(&chapterID, &currentPage, &totalPages, &lastReadAt)
		if err != nil {
			utils.Logger.Warn("Failed to scan chapter progress", zap.Error(err))
			continue
		}

		progress := &ReadingProgress{
			UserID:      userID,
			MangaID:     mangaID,
			ChapterID:   chapterID,
			CurrentPage: currentPage,
			TotalPages:  totalPages,
			LastReadAt:  lastReadAt,
		}

		mangaProgress.Chapters[chapterID] = progress

		if lastReadAt.After(mangaProgress.UpdatedAt) {
			mangaProgress.UpdatedAt = lastReadAt
		}
	}

	if len(mangaProgress.Chapters) == 0 {
		return nil, nil
	}

	return mangaProgress, nil
}

// GetAllReadingProgress retrieves all reading progress for a user
func GetAllReadingProgress(userID int) (*UserReadingData, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("userID is required")
	}

	query := `
		SELECT manga_id, chapter_id, current_page, total_pages, last_read_at
		FROM user_reading_progress
		WHERE user_id = $1
		ORDER BY last_read_at DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all reading progress: %w", err)
	}
	defer rows.Close()

	userData := &UserReadingData{
		UserID:    userID,
		Mangas:    make(map[string]*MangaProgress),
		UpdatedAt: time.Now(),
	}

	for rows.Next() {
		var mangaID, chapterID string
		var currentPage, totalPages int
		var lastReadAt time.Time

		err := rows.Scan(&mangaID, &chapterID, &currentPage, &totalPages, &lastReadAt)
		if err != nil {
			utils.Logger.Warn("Failed to scan reading progress", zap.Error(err))
			continue
		}

		progress := &ReadingProgress{
			UserID:      userID,
			MangaID:     mangaID,
			ChapterID:   chapterID,
			CurrentPage: currentPage,
			TotalPages:  totalPages,
			LastReadAt:  lastReadAt,
		}

		// Initialize manga progress if not exists
		if _, exists := userData.Mangas[mangaID]; !exists {
			userData.Mangas[mangaID] = &MangaProgress{
				MangaID:   mangaID,
				Chapters:  make(map[string]*ReadingProgress),
				UpdatedAt: lastReadAt,
			}
		}

		userData.Mangas[mangaID].Chapters[chapterID] = progress

		if lastReadAt.After(userData.Mangas[mangaID].UpdatedAt) {
			userData.Mangas[mangaID].UpdatedAt = lastReadAt
		}
		if lastReadAt.After(userData.UpdatedAt) {
			userData.UpdatedAt = lastReadAt
		}
	}

	return userData, nil
}

// GetReadMangaList retrieves the list of manga IDs that the user has read
func GetReadMangaList(userID int) ([]string, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("userID is required")
	}

	query := `
		SELECT DISTINCT manga_id
		FROM user_reading_progress
		WHERE user_id = $1
		ORDER BY manga_id
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get read manga list: %w", err)
	}
	defer rows.Close()

	var mangaIDs []string
	for rows.Next() {
		var mangaID string
		if err := rows.Scan(&mangaID); err != nil {
			continue
		}
		mangaIDs = append(mangaIDs, mangaID)
	}

	return mangaIDs, nil
}

// ============================================================================
// DELETE OPERATIONS
// ============================================================================

// DeleteReadingProgress removes reading progress for a specific chapter
func DeleteReadingProgress(userID int, chapterID string) error {
	if userID <= 0 || chapterID == "" {
		return fmt.Errorf("userID and chapterID are required")
	}

	// Delete from Redis
	redisKey := getRedisKey(userID, chapterID)
	_ = cache.Client.Del(cache.Ctx, redisKey).Err()

	// Delete from PostgreSQL
	query := `DELETE FROM user_reading_progress WHERE user_id = $1 AND chapter_id = $2`
	_, err := database.DB.Exec(query, userID, chapterID)
	if err != nil {
		return fmt.Errorf("failed to delete reading progress: %w", err)
	}

	utils.Logger.Info("Reading progress deleted",
		zap.Int("userId", userID),
		zap.String("chapterId", chapterID))

	return nil
}

// DeleteMangaProgress removes all reading progress for a specific manga
func DeleteMangaProgress(userID int, mangaID string) error {
	if userID <= 0 || mangaID == "" {
		return fmt.Errorf("userID and mangaID are required")
	}

	// Delete from PostgreSQL first to get chapter IDs
	query := `DELETE FROM user_reading_progress WHERE user_id = $1 AND manga_id = $2 RETURNING chapter_id`
	rows, err := database.DB.Query(query, userID, mangaID)
	if err != nil {
		return fmt.Errorf("failed to delete manga progress: %w", err)
	}
	defer rows.Close()

	// Delete from Redis
	var chapterIDs []string
	for rows.Next() {
		var chapterID string
		if rows.Scan(&chapterID) == nil {
			chapterIDs = append(chapterIDs, chapterID)
		}
	}

	for _, chapterID := range chapterIDs {
		redisKey := getRedisKey(userID, chapterID)
		_ = cache.Client.Del(cache.Ctx, redisKey).Err()
	}

	utils.Logger.Info("Manga progress deleted",
		zap.Int("userId", userID),
		zap.String("mangaId", mangaID),
		zap.Int("chaptersDeleted", len(chapterIDs)))

	return nil
}

// DeleteAllReadingProgress removes all reading progress for a user
func DeleteAllReadingProgress(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("userID is required")
	}

	// Delete from PostgreSQL first to get chapter IDs
	query := `DELETE FROM user_reading_progress WHERE user_id = $1 RETURNING chapter_id`
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all reading progress: %w", err)
	}
	defer rows.Close()

	// Delete from Redis
	var chapterIDs []string
	for rows.Next() {
		var chapterID string
		if rows.Scan(&chapterID) == nil {
			chapterIDs = append(chapterIDs, chapterID)
		}
	}

	for _, chapterID := range chapterIDs {
		redisKey := getRedisKey(userID, chapterID)
		_ = cache.Client.Del(cache.Ctx, redisKey).Err()
	}

	utils.Logger.Info("All reading progress deleted",
		zap.Int("userId", userID),
		zap.Int("chaptersDeleted", len(chapterIDs)))

	return nil
}

// ============================================================================
// INTERNAL HELPERS - Redis Operations
// ============================================================================

func getRedisKey(userID int, chapterID string) string {
	return fmt.Sprintf("progress:%d:%s", userID, chapterID)
}

func getFromRedis(userID int, chapterID string) (*ReadingProgress, error) {
	key := getRedisKey(userID, chapterID)

	data, err := cache.Client.Get(cache.Ctx, key).Bytes()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var progress ReadingProgress
	err = json.Unmarshal(data, &progress)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal progress: %w", err)
	}

	// Refresh TTL on read
	_ = cache.Client.Expire(cache.Ctx, key, ReadingProgressTTL).Err()

	return &progress, nil
}

func saveToRedis(progress *ReadingProgress) error {
	key := getRedisKey(progress.UserID, progress.ChapterID)

	data, err := json.Marshal(progress)
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	err = cache.Client.Set(cache.Ctx, key, data, ReadingProgressTTL).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

// ============================================================================
// INTERNAL HELPERS - PostgreSQL Operations
// ============================================================================

func getFromPostgres(userID int, chapterID string) (*ReadingProgress, error) {
	query := `
		SELECT manga_id, current_page, total_pages, last_read_at
		FROM user_reading_progress
		WHERE user_id = $1 AND chapter_id = $2
	`

	var mangaID string
	var currentPage, totalPages int
	var lastReadAt time.Time

	err := database.DB.QueryRow(query, userID, chapterID).Scan(&mangaID, &currentPage, &totalPages, &lastReadAt)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("postgres query error: %w", err)
	}

	return &ReadingProgress{
		UserID:      userID,
		MangaID:     mangaID,
		ChapterID:   chapterID,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		LastReadAt:  lastReadAt,
	}, nil
}

func saveToPostgres(progress *ReadingProgress) error {
	query := `
		INSERT INTO user_reading_progress (user_id, manga_id, chapter_id, current_page, total_pages, last_read_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, chapter_id)
		DO UPDATE SET
			current_page = EXCLUDED.current_page,
			total_pages = EXCLUDED.total_pages,
			last_read_at = EXCLUDED.last_read_at,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := database.DB.Exec(query,
		progress.UserID,
		progress.MangaID,
		progress.ChapterID,
		progress.CurrentPage,
		progress.TotalPages,
		progress.LastReadAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("postgres upsert error: %w", err)
	}

	return nil
}

