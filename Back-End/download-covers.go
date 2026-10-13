package main

import (
	"TaleHaven-Backend/modules/cache"
	"TaleHaven-Backend/modules/database"
	mangadex_api "TaleHaven-Backend/modules/mangadex-api"
	"TaleHaven-Backend/modules/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	utils.LoadEnv()
	utils.InitLogger()
	defer utils.Sync()

	cache.InitRedis()
	database.InitPostgres()
	defer database.CloseDB()

	fmt.Println("\n=== Downloading Cover Images ===\n")

	// Get all manga from database
	query := `SELECT mangadex_id, title FROM manga ORDER BY title`
	rows, err := database.DB.Query(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer rows.Close()

	successCount := 0
	failCount := 0

	for rows.Next() {
		var mangaID, title string
		if err := rows.Scan(&mangaID, &title); err != nil {
			continue
		}

		fmt.Printf("\nProcessing: %s\n", title)
		fmt.Printf("  Manga ID: %s\n", mangaID)

		// Get manga details from API
		mangaResponse, err := mangadex_api.GetMangaByID(mangaID)
		if err != nil {
			fmt.Printf("  ❌ Failed to fetch manga: %v\n", err)
			failCount++
			continue
		}

		// Find cover art relationship
		var coverFileName string
		for _, rel := range mangaResponse.Data.Relationships {
			if rel.Type == "cover_art" {
				if fileName, ok := rel.Attributes["fileName"].(string); ok {
					coverFileName = fileName
					break
				}
			}
		}

		if coverFileName == "" {
			fmt.Printf("  ❌ No cover art found\n")
			failCount++
			continue
		}

		fmt.Printf("  Cover file: %s\n", coverFileName)

		// Build cover URL - MangaDex covers format: https://uploads.mangadex.org/covers/{mangaID}/{fileName}
		coverURL := fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", mangaID, coverFileName)
		fmt.Printf("  Cover URL: %s\n", coverURL)

		// Create manga-specific folder
		mangaDir := filepath.Join("images", "manga", mangaID)
		if err := os.MkdirAll(mangaDir, 0755); err != nil {
			fmt.Printf("  ❌ Failed to create directory: %v\n", err)
			failCount++
			continue
		}

		// Local path for cover
		ext := filepath.Ext(coverFileName)
		if ext == "" {
			ext = ".jpg"
		}
		localPath := filepath.Join(mangaDir, "cover"+ext)

		// Download the cover
		err = downloadImage(coverURL, localPath)
		if err != nil {
			fmt.Printf("  ❌ Download failed: %v\n", err)
			failCount++
			continue
		}

		fmt.Printf("  ✓ Cover saved to: %s\n", localPath)

		// Update database with cover path - make it web-accessible
		webPath := strings.ReplaceAll(localPath, "\\", "/")
		updateQuery := `UPDATE manga SET cover_image_path = $1, cover_filename = $2, updated_at = CURRENT_TIMESTAMP WHERE mangadex_id = $3`
		_, err = database.DB.Exec(updateQuery, webPath, coverFileName, mangaID)
		if err != nil {
			fmt.Printf("  ⚠ Failed to update database: %v\n", err)
		} else {
			fmt.Printf("  ✓ Database updated\n")
		}

		successCount++
		time.Sleep(500 * time.Millisecond) // Rate limiting
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("✓ Successfully downloaded: %d covers\n", successCount)
	fmt.Printf("❌ Failed: %d covers\n", failCount)
}

func downloadImage(url string, localPath string) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "TaleHaven-Backend/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(localPath)
		return err
	}

	return nil
}

