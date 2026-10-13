package main

import (
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"encoding/json"
	"fmt"
)

func main() {
	utils.LoadEnv()
	utils.InitLogger()
	defer utils.Sync()

	database.InitPostgres()
	defer database.CloseDB()

	fmt.Println("\n=== Testing GET /manga/library endpoint ===\n")

	// Get all manga in database
	var allManga []struct {
		ID             int
		MangaDexID     string
		Title          string
		CoverImagePath string
	}

	query := `SELECT id, mangadex_id, title, COALESCE(cover_image_path, 'NULL') as cover_image_path FROM manga ORDER BY title`
	rows, err := database.DB.Query(query)
	if err != nil {
		fmt.Printf("Error querying manga: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m struct {
			ID             int
			MangaDexID     string
			Title          string
			CoverImagePath string
		}
		if err := rows.Scan(&m.ID, &m.MangaDexID, &m.Title, &m.CoverImagePath); err != nil {
			fmt.Printf("Error scanning: %v\n", err)
			continue
		}
		allManga = append(allManga, m)
	}

	fmt.Printf("Total manga in database: %d\n\n", len(allManga))
	for _, m := range allManga {
		fmt.Printf("Manga: %s\n", m.Title)
		fmt.Printf("  ID: %d\n", m.ID)
		fmt.Printf("  MangaDex ID: %s\n", m.MangaDexID)
		fmt.Printf("  Cover Path: %s\n", m.CoverImagePath)

		// Check chapters
		var chapterCount, downloadedCount int
		database.DB.QueryRow(`SELECT COUNT(*), SUM(CASE WHEN downloaded THEN 1 ELSE 0 END) FROM chapters WHERE manga_id = $1`, m.ID).Scan(&chapterCount, &downloadedCount)
		fmt.Printf("  Chapters: %d total, %d downloaded\n\n", chapterCount, downloadedCount)
	}

	// Now test the actual function
	fmt.Println("=== Testing GetDownloadedMangaWithCovers() function ===\n")
	mangaList, err := database.GetDownloadedMangaWithCovers()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Result count: %d\n", len(mangaList))
	if len(mangaList) == 0 {
		fmt.Println("No manga returned (empty library)")
		fmt.Println("\nPossible reasons:")
		fmt.Println("1. No manga have downloaded chapters (downloaded = true)")
		fmt.Println("2. No manga have cover images (cover_image_path is NULL or empty)")
	} else {
		fmt.Println("\nReturned manga:")
		jsonData, _ := json.MarshalIndent(mangaList, "", "  ")
		fmt.Println(string(jsonData))
	}
}

