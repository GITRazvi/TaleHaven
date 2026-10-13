package database

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // "-" prevents password from being serialized to JSON
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Manga represents manga metadata from MangaDex
type Manga struct {
	ID                   int            `json:"id" db:"id"`
	MangaDexID           string         `json:"mangadex_id" db:"mangadex_id"` // UUID from MangaDex
	Title                string         `json:"title" db:"title"`
	AlternativeTitles    pq.StringArray `json:"alternative_titles" db:"alternative_titles"`
	Description          sql.NullString `json:"description" db:"description"`
	Author               sql.NullString `json:"author" db:"author"`
	Artist               sql.NullString `json:"artist" db:"artist"`
	Tags                 pq.StringArray `json:"tags" db:"tags"` // Genre/theme tags
	Status               string         `json:"status" db:"status"` // ongoing, completed, hiatus, cancelled
	ContentRating        sql.NullString `json:"content_rating" db:"content_rating"` // safe, suggestive, erotica, pornographic
	Year                 sql.NullInt32  `json:"year" db:"year"`
	LastChapter          sql.NullString `json:"last_chapter" db:"last_chapter"`
	LastVolume           sql.NullString `json:"last_volume" db:"last_volume"`
	OriginalLanguage     sql.NullString `json:"original_language" db:"original_language"`
	AvailableLanguages   pq.StringArray `json:"available_languages" db:"available_languages"`
	CoverImagePath       sql.NullString `json:"cover_image_path" db:"cover_image_path"`
	CoverFileName        sql.NullString `json:"cover_filename" db:"cover_filename"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

// Chapter represents a manga chapter from MangaDex
type Chapter struct {
	ID                 int            `json:"id" db:"id"`
	MangaID            int            `json:"manga_id" db:"manga_id"`
	ChapterMangaDexID  string         `json:"chapter_mangadex_id" db:"chapter_mangadex_id"` // UUID from MangaDex
	ChapterNumber      sql.NullString `json:"chapter_number" db:"chapter_number"` // Can be string like "1", "1.5", etc.
	Title              sql.NullString `json:"title" db:"title"`
	Volume             sql.NullString `json:"volume" db:"volume"`
	TranslatedLanguage string         `json:"translated_language" db:"translated_language"`
	PageCount          int            `json:"page_count" db:"page_count"`
	FolderPath         string         `json:"folder_path" db:"folder_path"` // Local path where images are stored
	ScanlationGroup    sql.NullString `json:"scanlation_group" db:"scanlation_group"`
	PublishAt          sql.NullTime   `json:"publish_at" db:"publish_at"`
	Hash               sql.NullString `json:"hash" db:"hash"` // MangaDex chapter hash for image URLs
	Downloaded         bool           `json:"downloaded" db:"downloaded"` // Whether images have been downloaded
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
}

// ChapterPage represents individual pages of a chapter
type ChapterPage struct {
	ID          int       `json:"id" db:"id"`
	ChapterID   int       `json:"chapter_id" db:"chapter_id"`
	PageNumber  int       `json:"page_number" db:"page_number"`
	FileName    string    `json:"file_name" db:"file_name"` // Original filename from MangaDex
	LocalPath   string    `json:"local_path" db:"local_path"` // Local storage path
	Downloaded  bool      `json:"downloaded" db:"downloaded"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

