package cleanup

import (
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/sessions"
	"TaleHaven-Backend/modules/utils"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// MangaCleanupService handles automatic deletion of inactive manga
type MangaCleanupService struct {
	ticker *time.Ticker
	done   chan bool
}

// NewMangaCleanupService creates a new cleanup service
func NewMangaCleanupService() *MangaCleanupService {
	return &MangaCleanupService{
		done: make(chan bool),
	}
}

// Start begins the cleanup service (runs once per day)
func (s *MangaCleanupService) Start() {
	// Run immediately on start
	go s.performCleanup()

	// Then run once per day
	s.ticker = time.NewTicker(24 * time.Hour)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.performCleanup()
			case <-s.done:
				s.ticker.Stop()
				return
			}
		}
	}()

	utils.Logger.Info("Manga cleanup service started (runs daily)")
}

// Stop stops the cleanup service
func (s *MangaCleanupService) Stop() {
	s.done <- true
	utils.Logger.Info("Manga cleanup service stopped")
}

// performCleanup executes the cleanup process
func (s *MangaCleanupService) performCleanup() {
	utils.Logger.Info("Starting manga cleanup process")

	// Get inactive manga (not accessed in 2 weeks)
	inactiveManga, err := sessions.GetInactiveManga()
	if err != nil {
		utils.Logger.Error("Failed to get inactive manga", zap.Error(err))
		return
	}

	if len(inactiveManga) == 0 {
		utils.Logger.Info("No inactive manga found to clean up")
		return
	}

	utils.Logger.Info("Found inactive manga to clean up",
		zap.Int("count", len(inactiveManga)))

	var deletedCount int
	var failedCount int

	for _, mangaID := range inactiveManga {
		err := s.cleanupManga(mangaID)
		if err != nil {
			utils.Logger.Error("Failed to cleanup manga",
				zap.String("mangaId", mangaID),
				zap.Error(err))
			failedCount++
		} else {
			deletedCount++
		}
	}

	utils.Logger.Info("Manga cleanup completed",
		zap.Int("deleted", deletedCount),
		zap.Int("failed", failedCount),
		zap.Int("total", len(inactiveManga)))
}

// cleanupManga deletes all data for a specific manga
func (s *MangaCleanupService) cleanupManga(mangaID string) error {
	utils.Logger.Info("Cleaning up manga", zap.String("mangaId", mangaID))

	// 1. Delete downloaded images
	err := s.deleteMangaImages(mangaID)
	if err != nil {
		utils.Logger.Warn("Failed to delete manga images",
			zap.String("mangaId", mangaID),
			zap.Error(err))
		// Continue even if image deletion fails
	}

	// 2. Delete from database (optional - you might want to keep metadata)
	err = s.deleteMangaFromDB(mangaID)
	if err != nil {
		utils.Logger.Warn("Failed to delete manga from database",
			zap.String("mangaId", mangaID),
			zap.Error(err))
		// Continue even if DB deletion fails
	}

	// 3. Remove from active tracking
	err = sessions.RemoveMangaFromActiveSet(mangaID)
	if err != nil {
		utils.Logger.Warn("Failed to remove manga from active tracking",
			zap.String("mangaId", mangaID),
			zap.Error(err))
	}

	utils.Logger.Info("Manga cleanup completed successfully",
		zap.String("mangaId", mangaID))

	return nil
}

// deleteMangaImages removes all downloaded images for a manga
func (s *MangaCleanupService) deleteMangaImages(mangaID string) error {
	// Path to manga images: ./images/manga/{mangaID}/
	imagesPath := filepath.Join("images", "manga", mangaID)

	// Check if directory exists
	if _, err := os.Stat(imagesPath); os.IsNotExist(err) {
		utils.Logger.Debug("Manga images directory doesn't exist",
			zap.String("mangaId", mangaID),
			zap.String("path", imagesPath))
		return nil // Not an error if directory doesn't exist
	}

	// Delete the entire manga directory
	err := os.RemoveAll(imagesPath)
	if err != nil {
		return err
	}

	utils.Logger.Info("Deleted manga images",
		zap.String("mangaId", mangaID),
		zap.String("path", imagesPath))

	return nil
}

// deleteMangaFromDB removes manga entries from the database
// This is optional - you might want to keep the metadata even after deleting images
func (s *MangaCleanupService) deleteMangaFromDB(mangaID string) error {
	if database.DB == nil {
		return nil // Skip if DB not initialized
	}

	// Delete chapter pages
	_, err := database.DB.Exec(`DELETE FROM chapter_pages WHERE chapter_id IN 
		(SELECT id FROM chapters WHERE manga_id = $1)`, mangaID)
	if err != nil {
		return err
	}

	// Delete chapters
	_, err = database.DB.Exec(`DELETE FROM chapters WHERE manga_id = $1`, mangaID)
	if err != nil {
		return err
	}

	// Delete manga (optional - might want to keep for history)
	// _, err = database.DB.Exec(`DELETE FROM manga WHERE id = $1`, mangaID)
	// if err != nil {
	// 	return err
	// }

	utils.Logger.Info("Deleted manga data from database",
		zap.String("mangaId", mangaID))

	return nil
}

// ForceCleanupManga manually triggers cleanup for a specific manga (for testing/admin)
func ForceCleanupManga(mangaID string) error {
	service := NewMangaCleanupService()
	return service.cleanupManga(mangaID)
}

