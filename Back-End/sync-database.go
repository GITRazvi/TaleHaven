package main

import (
	"TaleHaven-Backend/modules/cache"
	"TaleHaven-Backend/modules/database"
	mangadex_api "TaleHaven-Backend/modules/mangadex-api"
	"TaleHaven-Backend/modules/utils"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	utils.LoadEnv()
	utils.InitLogger()
	defer utils.Sync()

	cache.InitRedis()

	database.InitPostgres()
	defer database.CloseDB()

	fmt.Println("\n=== Syncing Database with Downloaded Files ===\n")

	// Get all manga folders
	mangaBaseDir := filepath.Join("images", "manga")
	mangaDirs, err := os.ReadDir(mangaBaseDir)
	if err != nil {
		fmt.Printf("Error reading manga directory: %v\n", err)
		return
	}

	totalManga := 0
	totalChapters := 0
	totalCoversDownloaded := 0

	for _, mangaDir := range mangaDirs {
		if !mangaDir.IsDir() {
			continue
		}

		mangaID := mangaDir.Name()
		fmt.Printf("\nProcessing manga: %s\n", mangaID)

		// Fetch manga from API and save to database
		mangaResponse, err := mangadex_api.GetMangaByID(mangaID)
		if err != nil {
			fmt.Printf("  ❌ Failed to fetch manga from API: %v\n", err)
			continue
		}

		// Save/update manga using direct database operation
		manga := convertMangaResponseToDBModel(mangaResponse)
		mangaDBID, err := database.SaveManga(manga)
		if err != nil {
			fmt.Printf("  ❌ Failed to save manga: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Manga in database: %s (ID: %d)\n", manga.Title, mangaDBID)
		totalManga++

		// Download cover image
		fmt.Printf("  Downloading cover image...\n")
		err = mangadex_api.DownloadCoverImage(mangaID, "512")
		if err != nil {
			fmt.Printf("  ⚠ Failed to download cover: %v\n", err)
		} else {
			fmt.Printf("  ✓ Cover downloaded\n")
			totalCoversDownloaded++
		}

		// Check for chapter folders
		chaptersDir := filepath.Join(mangaBaseDir, mangaID, "chapters")
		chapterDirs, err := os.ReadDir(chaptersDir)
		if err != nil {
			fmt.Printf("  No chapters folder found\n")
			continue
		}

		fmt.Printf("  Found %d chapter folders\n", len(chapterDirs))

		for _, chapterDir := range chapterDirs {
			if !chapterDir.IsDir() {
				continue
			}

			chapterID := chapterDir.Name()

			// Check if chapter exists in database
			dbChapter, err := database.GetChapterByMangaDexID(chapterID)
			if err != nil {
				fmt.Printf("    ⚠ Error checking chapter %s: %v\n", chapterID, err)
				continue
			}

			if dbChapter != nil {
				// Chapter exists, mark as downloaded
				if !dbChapter.Downloaded {
					err = database.UpdateChapterDownloadStatus(dbChapter.ID, true)
					if err != nil {
						fmt.Printf("    ⚠ Failed to mark chapter as downloaded: %v\n", err)
					} else {
						fmt.Printf("    ✓ Marked chapter %s as downloaded\n", chapterID)
						totalChapters++
					}
				} else {
					fmt.Printf("    ✓ Chapter %s already marked as downloaded\n", chapterID)
					totalChapters++
				}
				continue
			}

			// Chapter doesn't exist, fetch and save it
			fmt.Printf("    Fetching chapter %s from API...\n", chapterID)
			chapterResponse, err := mangadex_api.GetChapterByID(chapterID)
			if err != nil {
				fmt.Printf("    ⚠ Failed to fetch chapter: %v\n", err)
				continue
			}

			// Save chapter to database
			chapter := &database.Chapter{
				MangaID:            mangaDBID,
				ChapterMangaDexID:  chapterID,
				ChapterNumber:      sql.NullString{String: chapterResponse.Data.Attributes.Chapter, Valid: true},
				Title:              sql.NullString{String: chapterResponse.Data.Attributes.Title, Valid: true},
				Volume:             sql.NullString{String: chapterResponse.Data.Attributes.Volume, Valid: true},
				TranslatedLanguage: chapterResponse.Data.Attributes.TranslatedLanguage,
				PageCount:          chapterResponse.Data.Attributes.Pages,
				FolderPath:         filepath.Join(chaptersDir, chapterID),
				Downloaded:         true, // Mark as downloaded since files exist
			}

			_, err = database.SaveChapter(chapter)
			if err != nil {
				fmt.Printf("    ⚠ Failed to save chapter: %v\n", err)
				continue
			}

			fmt.Printf("    ✓ Saved and marked chapter %s as downloaded\n", chapterID)
			totalChapters++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total manga synced: %d\n", totalManga)
	fmt.Printf("Total chapters synced: %d\n", totalChapters)
	fmt.Printf("Total covers downloaded: %d\n", totalCoversDownloaded)
	fmt.Printf("\nDatabase is now in sync with downloaded files!\n")
}

func convertMangaResponseToDBModel(mangaResponse *mangadex_api.MangaResponse) *database.Manga {
	manga := &database.Manga{
		MangaDexID:  mangaResponse.Data.ID,
		Title:       getPreferredTitle(mangaResponse.Data.Attributes.Title),
		Description: sql.NullString{String: getPreferredDescription(mangaResponse.Data.Attributes.Description), Valid: true},
		Status:      mangaResponse.Data.Attributes.Status,
	}

	// Extract tags
	tags := make([]string, 0)
	for _, tag := range mangaResponse.Data.Attributes.Tags {
		if tagName, ok := tag.Attributes.Name["en"]; ok {
			tags = append(tags, tagName)
		}
	}
	manga.Tags = tags

	// Extract author and artist from relationships
	for _, rel := range mangaResponse.Data.Relationships {
		switch rel.Type {
		case "author":
			if name, ok := rel.Attributes["name"].(string); ok {
				manga.Author = sql.NullString{String: name, Valid: true}
			}
		case "artist":
			if name, ok := rel.Attributes["name"].(string); ok {
				manga.Artist = sql.NullString{String: name, Valid: true}
			}
		}
	}

	// Set other fields
	if mangaResponse.Data.Attributes.Year > 0 {
		manga.Year = sql.NullInt32{Int32: int32(mangaResponse.Data.Attributes.Year), Valid: true}
	}
	manga.ContentRating = sql.NullString{String: mangaResponse.Data.Attributes.ContentRating, Valid: true}
	manga.LastChapter = sql.NullString{String: mangaResponse.Data.Attributes.LastChapter, Valid: true}
	manga.LastVolume = sql.NullString{String: mangaResponse.Data.Attributes.LastVolume, Valid: true}
	manga.OriginalLanguage = sql.NullString{String: mangaResponse.Data.Attributes.OriginalLanguage, Valid: true}
	manga.AvailableLanguages = mangaResponse.Data.Attributes.AvailableTranslatedLanguages

	return manga
}

func getPreferredTitle(titles map[string]string) string {
	// Prefer English, then romaji, then any available
	if title, ok := titles["en"]; ok {
		return title
	}
	if title, ok := titles["ja-ro"]; ok {
		return title
	}
	for _, title := range titles {
		return title
	}
	return "Unknown Title"
}

func getPreferredDescription(descriptions map[string]string) string {
	if desc, ok := descriptions["en"]; ok {
		return desc
	}
	for _, desc := range descriptions {
		return desc
	}
	return ""
}

