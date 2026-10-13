package sessions

import (
	"TaleHaven-Backend/modules/cache"
	"TaleHaven-Backend/modules/utils"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	MangaAccessTTL = 14 * 24 * time.Hour // 2 weeks
)

// TrackMangaAccess records when a manga/chapter is accessed
// This is used to determine which manga can be auto-deleted after 2 weeks of inactivity
func TrackMangaAccess(mangaID, chapterID string) error {
	if mangaID == "" {
		return fmt.Errorf("mangaID is required")
	}

	now := time.Now().Unix()

	// Track manga-level access
	mangaKey := fmt.Sprintf("manga:access:%s", mangaID)
	err := cache.Client.Set(cache.Ctx, mangaKey, now, MangaAccessTTL).Err()
	if err != nil {
		utils.Logger.Error("Failed to track manga access",
			zap.String("mangaId", mangaID),
			zap.Error(err))
		return fmt.Errorf("failed to track manga access: %w", err)
	}

	// Track chapter-level access if provided
	if chapterID != "" {
		chapterKey := fmt.Sprintf("manga:access:%s:chapter:%s", mangaID, chapterID)
		err = cache.Client.Set(cache.Ctx, chapterKey, now, MangaAccessTTL).Err()
		if err != nil {
			utils.Logger.Warn("Failed to track chapter access",
				zap.String("mangaId", mangaID),
				zap.String("chapterId", chapterID),
				zap.Error(err))
			// Don't return error for chapter tracking failure
		}
	}

	// Add to active manga set for quick lookup
	activeMangaKey := "manga:active_set"
	err = cache.Client.SAdd(cache.Ctx, activeMangaKey, mangaID).Err()
	if err != nil {
		utils.Logger.Warn("Failed to add manga to active set",
			zap.String("mangaId", mangaID),
			zap.Error(err))
	}

	utils.Logger.Debug("Manga access tracked",
		zap.String("mangaId", mangaID),
		zap.String("chapterId", chapterID))

	return nil
}

// GetMangaLastAccess returns the timestamp of when a manga was last accessed
func GetMangaLastAccess(mangaID string) (*time.Time, error) {
	if mangaID == "" {
		return nil, fmt.Errorf("mangaID is required")
	}

	mangaKey := fmt.Sprintf("manga:access:%s", mangaID)
	timestamp, err := cache.Client.Get(cache.Ctx, mangaKey).Int64()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get manga last access: %w", err)
	}

	lastAccess := time.Unix(timestamp, 0)
	return &lastAccess, nil
}

// GetInactiveManga returns list of manga IDs that haven't been accessed in 2 weeks
func GetInactiveManga() ([]string, error) {
	// Get all manga in active set
	activeMangaKey := "manga:active_set"
	mangaIDs, err := cache.Client.SMembers(cache.Ctx, activeMangaKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active manga set: %w", err)
	}

	var inactiveManga []string
	twoWeeksAgo := time.Now().Add(-MangaAccessTTL)

	for _, mangaID := range mangaIDs {
		lastAccess, err := GetMangaLastAccess(mangaID)
		if err != nil {
			utils.Logger.Warn("Failed to get last access time",
				zap.String("mangaId", mangaID),
				zap.Error(err))
			continue
		}

		// If no access record or older than 2 weeks, mark as inactive
		if lastAccess == nil || lastAccess.Before(twoWeeksAgo) {
			inactiveManga = append(inactiveManga, mangaID)
		}
	}

	utils.Logger.Info("Found inactive manga",
		zap.Int("count", len(inactiveManga)),
		zap.Int("total", len(mangaIDs)))

	return inactiveManga, nil
}

// RemoveMangaFromActiveSet removes a manga from the active tracking set
func RemoveMangaFromActiveSet(mangaID string) error {
	if mangaID == "" {
		return fmt.Errorf("mangaID is required")
	}

	activeMangaKey := "manga:active_set"
	err := cache.Client.SRem(cache.Ctx, activeMangaKey, mangaID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove manga from active set: %w", err)
	}

	// Clean up access tracking keys
	mangaKey := fmt.Sprintf("manga:access:%s", mangaID)
	err = cache.Client.Del(cache.Ctx, mangaKey).Err()
	if err != nil {
		utils.Logger.Warn("Failed to delete manga access key",
			zap.String("mangaId", mangaID),
			zap.Error(err))
	}

	// Clean up chapter access keys
	pattern := fmt.Sprintf("manga:access:%s:chapter:*", mangaID)
	keys, err := cache.Client.Keys(cache.Ctx, pattern).Result()
	if err == nil && len(keys) > 0 {
		err = cache.Client.Del(cache.Ctx, keys...).Err()
		if err != nil {
			utils.Logger.Warn("Failed to delete chapter access keys",
				zap.String("mangaId", mangaID),
				zap.Error(err))
		}
	}

	utils.Logger.Info("Manga removed from active tracking",
		zap.String("mangaId", mangaID))

	return nil
}

