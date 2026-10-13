package mangadex_api

import (
	"TaleHaven-Backend/modules/utils"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"go.uber.org/zap"
)

// PrefetchNextChapter downloads the next chapter in the background to reduce lag
func PrefetchNextChapter(currentChapterID, mangaID string, quality string) {
	// Run in a goroutine to avoid blocking the current request
	go func() {
		utils.Logger.Info("Starting prefetch for next chapter",
			zap.String("currentChapterId", currentChapterID),
			zap.String("mangaId", mangaID))

		// Find the next chapter
		nextChapterID, err := FindNextChapter(mangaID, currentChapterID)
		if err != nil {
			utils.Logger.Warn("Could not find next chapter for prefetch",
				zap.String("currentChapterId", currentChapterID),
				zap.Error(err))
			return
		}

		if nextChapterID == "" {
			utils.Logger.Info("No next chapter to prefetch (this is the last chapter)",
				zap.String("currentChapterId", currentChapterID))
			return
		}

		// Check if already downloaded
		if IsChapterDownloaded(nextChapterID) {
			utils.Logger.Info("Next chapter already downloaded, skipping prefetch",
				zap.String("nextChapterId", nextChapterID))
			return
		}

		// Download the next chapter
		utils.Logger.Info("Prefetching next chapter",
			zap.String("nextChapterId", nextChapterID))

		_, err = DownloadAndGetChapterPages(nextChapterID, quality)
		if err != nil {
			utils.Logger.Error("Failed to prefetch next chapter",
				zap.String("nextChapterId", nextChapterID),
				zap.Error(err))
			return
		}

		utils.Logger.Info("Successfully prefetched next chapter",
			zap.String("nextChapterId", nextChapterID))
	}()
}

// FindNextChapter finds the next chapter in sequence based on chapter number
func FindNextChapter(mangaID, currentChapterID string) (string, error) {
	// Get current chapter details
	currentChapter, err := GetChapterByID(currentChapterID)
	if err != nil {
		return "", err
	}

	currentChapterNum := currentChapter.Data.Attributes.Chapter

	// Get all chapters for the manga
	// Use a reasonable limit, sort by chapter number
	chaptersResponse, err := GetMangaChapters(mangaID, 500, 0, []string{"en"})
	if err != nil {
		return "", err
	}

	// Parse and sort chapters
	type chapterInfo struct {
		ID     string
		Number float64
	}

	var chapters []chapterInfo
	for _, chapter := range chaptersResponse.Data {
		if chapter.Attributes.Chapter == "" {
			continue // Skip chapters without a number
		}

		num, err := strconv.ParseFloat(chapter.Attributes.Chapter, 64)
		if err != nil {
			utils.Logger.Warn("Could not parse chapter number",
				zap.String("chapterId", chapter.ID),
				zap.String("chapter", chapter.Attributes.Chapter))
			continue
		}

		chapters = append(chapters, chapterInfo{
			ID:     chapter.ID,
			Number: num,
		})
	}

	// Sort by chapter number
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Number < chapters[j].Number
	})

	// Parse current chapter number
	currentNum, err := strconv.ParseFloat(currentChapterNum, 64)
	if err != nil {
		utils.Logger.Warn("Could not parse current chapter number",
			zap.String("currentChapterId", currentChapterID),
			zap.String("chapter", currentChapterNum))
		return "", err
	}

	// Find the next chapter
	for i, chapter := range chapters {
		if chapter.ID == currentChapterID {
			// Found current chapter, return next one if it exists
			if i+1 < len(chapters) {
				nextChapter := chapters[i+1]
				utils.Logger.Info("Found next chapter",
					zap.String("currentChapter", currentChapterNum),
					zap.Float64("nextChapterNum", nextChapter.Number),
					zap.String("nextChapterId", nextChapter.ID))
				return nextChapter.ID, nil
			}
			// This is the last chapter
			return "", nil
		}
	}

	// Alternative: find by chapter number if ID not found
	for _, chapter := range chapters {
		if chapter.Number > currentNum {
			utils.Logger.Info("Found next chapter by number",
				zap.Float64("currentNum", currentNum),
				zap.Float64("nextNum", chapter.Number),
				zap.String("nextChapterId", chapter.ID))
			return chapter.ID, nil
		}
	}

	// No next chapter found
	return "", nil
}

// IsChapterDownloaded checks if a chapter's images are already downloaded
func IsChapterDownloaded(chapterID string) bool {
	// Try to get chapter details to find manga ID
	chapterResponse, err := GetChapterByID(chapterID)
	if err != nil {
		return false
	}

	// Extract manga ID
	var mangaID string
	for _, rel := range chapterResponse.Data.Relationships {
		if rel.Type == "manga" {
			mangaID = rel.ID
			break
		}
	}

	if mangaID == "" {
		return false
	}

	// Check if the chapter folder exists and has files
	chapterFolder := getChapterFolderPath(mangaID, chapterID)

	// Check if folder exists
	files, err := readDir(chapterFolder)
	if err != nil || len(files) == 0 {
		return false
	}

	// If folder has files, consider it downloaded
	return len(files) > 0
}

// Helper function to read directory (wrapper for os.ReadDir for easier testing)
func readDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// Helper to get chapter folder path (consistent with DownloadAndGetChapterPages)
func getChapterFolderPath(mangaID, chapterID string) string {
	return filepath.Join(ImagesBaseDir, MangaDir, mangaID, "chapters", chapterID)
}
