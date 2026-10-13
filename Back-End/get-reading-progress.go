package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"database/sql"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// ReadingProgressRecord represents a single reading progress entry
type ReadingProgressRecord struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	UserEmail   string  `json:"user_email"`
	MangaID     string  `json:"manga_id"`
	ChapterID   string  `json:"chapter_id"`
	CurrentPage int     `json:"current_page"`
	TotalPages  *int    `json:"total_pages"`
	LastReadAt  string  `json:"last_read_at"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// UserProgressSummary represents aggregated progress for a user
type UserProgressSummary struct {
	UserID         int                     `json:"user_id"`
	UserEmail      string                  `json:"user_email"`
	TotalChapters  int                     `json:"total_chapters_read"`
	UniqueManga    int                     `json:"unique_manga_count"`
	LastActivity   string                  `json:"last_activity"`
	ProgressItems  []ReadingProgressRecord `json:"progress_items"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Build connection string
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "talehaven")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✓ Connected to PostgreSQL database")
	fmt.Println("=====================================")
	fmt.Println()

	// Get command line argument for format
	format := "summary"
	if len(os.Args) > 1 {
		format = os.Args[1]
	}

	switch format {
	case "all":
		getAllProgress(db)
	case "summary":
		getUserSummaries(db)
	case "json":
		getAllProgressJSON(db)
	case "stats":
		getStatistics(db)
	default:
		fmt.Println("Usage: go run get-reading-progress.go [all|summary|json|stats]")
		fmt.Println("  all     - Show all progress records (detailed)")
		fmt.Println("  summary - Show summary per user (default)")
		fmt.Println("  json    - Export all data as JSON")
		fmt.Println("  stats   - Show aggregate statistics")
	}
}

func getAllProgress(db *sql.DB) {
	query := `
		SELECT 
			rp.id, rp.user_id, u.email, rp.manga_id, rp.chapter_id, 
			rp.current_page, rp.total_pages, rp.last_read_at, 
			rp.created_at, rp.updated_at
		FROM user_reading_progress rp
		JOIN users u ON rp.user_id = u.id
		ORDER BY rp.last_read_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query reading progress: %v", err)
	}
	defer rows.Close()

	fmt.Println("ALL READING PROGRESS")
	fmt.Println("=====================================")
	fmt.Printf("%-5s %-8s %-25s %-38s %-38s %-5s %-5s %-20s\n",
		"ID", "UserID", "Email", "MangaID", "ChapterID", "Page", "Total", "LastRead")
	fmt.Println("---------------------------------------------------------------------------------------------------------------------------------------------------------------------")

	count := 0
	for rows.Next() {
		var record ReadingProgressRecord
		var totalPages sql.NullInt64

		err := rows.Scan(
			&record.ID, &record.UserID, &record.UserEmail,
			&record.MangaID, &record.ChapterID, &record.CurrentPage,
			&totalPages, &record.LastReadAt, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		if totalPages.Valid {
			pages := int(totalPages.Int64)
			record.TotalPages = &pages
		}

		totalPagesStr := "N/A"
		if record.TotalPages != nil {
			totalPagesStr = fmt.Sprintf("%d", *record.TotalPages)
		}

		fmt.Printf("%-5d %-8d %-25s %-38s %-38s %-5d %-5s %-20s\n",
			record.ID, record.UserID, record.UserEmail,
			record.MangaID[:min(38, len(record.MangaID))],
			record.ChapterID[:min(38, len(record.ChapterID))],
			record.CurrentPage, totalPagesStr, record.LastReadAt[:19])

		count++
	}

	fmt.Println()
	fmt.Printf("Total records: %d\n", count)
}

func getUserSummaries(db *sql.DB) {
	query := `
		SELECT 
			u.id, u.email,
			COUNT(DISTINCT rp.chapter_id) as total_chapters,
			COUNT(DISTINCT rp.manga_id) as unique_manga,
			MAX(rp.last_read_at) as last_activity
		FROM users u
		LEFT JOIN user_reading_progress rp ON u.id = rp.user_id
		GROUP BY u.id, u.email
		ORDER BY u.id
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query user summaries: %v", err)
	}
	defer rows.Close()

	fmt.Println("USER READING PROGRESS SUMMARY")
	fmt.Println("=====================================")
	fmt.Printf("%-8s %-30s %-15s %-15s %-20s\n",
		"UserID", "Email", "Chapters Read", "Unique Manga", "Last Activity")
	fmt.Println("-----------------------------------------------------------------------------------------------------")

	totalUsers := 0
	totalChapters := 0
	totalManga := 0

	for rows.Next() {
		var userID int
		var email string
		var chapters int
		var uniqueManga int
		var lastActivity sql.NullString

		err := rows.Scan(&userID, &email, &chapters, &uniqueManga, &lastActivity)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		lastActivityStr := "Never"
		if lastActivity.Valid {
			lastActivityStr = lastActivity.String[:19]
		}

		fmt.Printf("%-8d %-30s %-15d %-15d %-20s\n",
			userID, email, chapters, uniqueManga, lastActivityStr)

		totalUsers++
		totalChapters += chapters
		totalManga += uniqueManga
	}

	fmt.Println()
	fmt.Printf("Total Users: %d\n", totalUsers)
	fmt.Printf("Total Chapters Read (across all users): %d\n", totalChapters)
	fmt.Printf("Total Unique Manga Read (across all users): %d\n", totalManga)
}

func getAllProgressJSON(db *sql.DB) {
	query := `
		SELECT 
			rp.id, rp.user_id, u.email, rp.manga_id, rp.chapter_id, 
			rp.current_page, rp.total_pages, rp.last_read_at, 
			rp.created_at, rp.updated_at
		FROM user_reading_progress rp
		JOIN users u ON rp.user_id = u.id
		ORDER BY rp.user_id, rp.last_read_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query reading progress: %v", err)
	}
	defer rows.Close()

	var records []ReadingProgressRecord

	for rows.Next() {
		var record ReadingProgressRecord
		var totalPages sql.NullInt64

		err := rows.Scan(
			&record.ID, &record.UserID, &record.UserEmail,
			&record.MangaID, &record.ChapterID, &record.CurrentPage,
			&totalPages, &record.LastReadAt, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		if totalPages.Valid {
			pages := int(totalPages.Int64)
			record.TotalPages = &pages
		}

		records = append(records, record)
	}

	jsonData, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}

func getStatistics(db *sql.DB) {
	fmt.Println("READING PROGRESS STATISTICS")
	fmt.Println("=====================================")
	fmt.Println()

	// Total records
	var totalRecords int
	db.QueryRow("SELECT COUNT(*) FROM user_reading_progress").Scan(&totalRecords)
	fmt.Printf("Total Progress Records: %d\n", totalRecords)

	// Total users with progress
	var usersWithProgress int
	db.QueryRow("SELECT COUNT(DISTINCT user_id) FROM user_reading_progress").Scan(&usersWithProgress)
	fmt.Printf("Users with Progress: %d\n", usersWithProgress)

	// Total unique manga
	var uniqueManga int
	db.QueryRow("SELECT COUNT(DISTINCT manga_id) FROM user_reading_progress").Scan(&uniqueManga)
	fmt.Printf("Unique Manga Read: %d\n", uniqueManga)

	// Total unique chapters
	var uniqueChapters int
	db.QueryRow("SELECT COUNT(DISTINCT chapter_id) FROM user_reading_progress").Scan(&uniqueChapters)
	fmt.Printf("Unique Chapters Read: %d\n", uniqueChapters)

	// Average pages per chapter
	var avgPages float64
	db.QueryRow("SELECT AVG(total_pages) FROM user_reading_progress WHERE total_pages IS NOT NULL").Scan(&avgPages)
	fmt.Printf("Average Pages per Chapter: %.2f\n", avgPages)

	// Most active user
	var mostActiveUser string
	var mostActiveCount int
	db.QueryRow(`
		SELECT u.email, COUNT(*) as count
		FROM user_reading_progress rp
		JOIN users u ON rp.user_id = u.id
		GROUP BY u.id, u.email
		ORDER BY count DESC
		LIMIT 1
	`).Scan(&mostActiveUser, &mostActiveCount)
	fmt.Printf("Most Active User: %s (%d chapters)\n", mostActiveUser, mostActiveCount)

	// Most read manga
	var mostReadManga string
	var mostReadCount int
	err := db.QueryRow(`
		SELECT manga_id, COUNT(*) as count
		FROM user_reading_progress
		GROUP BY manga_id
		ORDER BY count DESC
		LIMIT 1
	`).Scan(&mostReadManga, &mostReadCount)

	if err == nil {
		fmt.Printf("Most Read Manga: %s (%d times)\n", mostReadManga, mostReadCount)
	}

	// Recent activity
	var recentActivity string
	db.QueryRow("SELECT MAX(last_read_at) FROM user_reading_progress").Scan(&recentActivity)
	fmt.Printf("Most Recent Activity: %s\n", recentActivity)

	fmt.Println()
	fmt.Println("Top 5 Most Active Users:")
	fmt.Println("-------------------------------------")

	rows, err := db.Query(`
		SELECT u.email, COUNT(*) as chapters_read
		FROM user_reading_progress rp
		JOIN users u ON rp.user_id = u.id
		GROUP BY u.id, u.email
		ORDER BY chapters_read DESC
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		i := 1
		for rows.Next() {
			var email string
			var count int
			rows.Scan(&email, &count)
			fmt.Printf("%d. %s - %d chapters\n", i, email, count)
			i++
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

