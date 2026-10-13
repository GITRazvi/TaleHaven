package database

import (
	"TaleHaven-Backend/modules/utils"

	"go.uber.org/zap"
)

// RunMigrations executes all database migrations
func RunMigrations() {
	utils.Logger.Info("Running database migrations...")

	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	// Create sessions table
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(64) PRIMARY KEY,
		user_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		ip_address VARCHAR(45),
		user_agent TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`

	// Create manga table with MangaDex-specific fields
	createMangaTable := `
	CREATE TABLE IF NOT EXISTS manga (
		id SERIAL PRIMARY KEY,
		mangadex_id VARCHAR(255) UNIQUE NOT NULL,
		title VARCHAR(500) NOT NULL,
		alternative_titles TEXT[],
		description TEXT,
		author VARCHAR(255),
		artist VARCHAR(255),
		tags TEXT[],
		status VARCHAR(50) DEFAULT 'ongoing',
		content_rating VARCHAR(50),
		year INTEGER,
		last_chapter VARCHAR(50),
		last_volume VARCHAR(50),
		original_language VARCHAR(10),
		available_languages TEXT[],
		cover_image_path VARCHAR(1000),
		cover_filename VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_manga_title ON manga(title);
	CREATE INDEX IF NOT EXISTS idx_manga_mangadex_id ON manga(mangadex_id);
	`

	// Create chapters table with MangaDex-specific fields
	createChaptersTable := `
	CREATE TABLE IF NOT EXISTS chapters (
		id SERIAL PRIMARY KEY,
		manga_id INTEGER NOT NULL,
		chapter_mangadex_id VARCHAR(255) UNIQUE NOT NULL,
		chapter_number VARCHAR(50),
		title VARCHAR(500),
		volume VARCHAR(50),
		translated_language VARCHAR(10) NOT NULL,
		page_count INTEGER DEFAULT 0,
		folder_path VARCHAR(1000) NOT NULL,
		scanlation_group VARCHAR(255),
		publish_at TIMESTAMP,
		hash VARCHAR(255),
		downloaded BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_chapters_manga_id ON chapters(manga_id);
	CREATE INDEX IF NOT EXISTS idx_chapters_mangadex_id ON chapters(chapter_mangadex_id);
	CREATE INDEX IF NOT EXISTS idx_chapters_downloaded ON chapters(downloaded);
	`

	// Create chapter_pages table for individual page tracking
	createChapterPagesTable := `
	CREATE TABLE IF NOT EXISTS chapter_pages (
		id SERIAL PRIMARY KEY,
		chapter_id INTEGER NOT NULL,
		page_number INTEGER NOT NULL,
		file_name VARCHAR(255) NOT NULL,
		local_path VARCHAR(1000) NOT NULL,
		downloaded BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
		UNIQUE(chapter_id, page_number)
	);
	CREATE INDEX IF NOT EXISTS idx_chapter_pages_chapter_id ON chapter_pages(chapter_id);
	CREATE INDEX IF NOT EXISTS idx_chapter_pages_downloaded ON chapter_pages(downloaded);
	`

	// Create user_reading_progress table for persistent reading progress
	createReadingProgressTable := `
	CREATE TABLE IF NOT EXISTS user_reading_progress (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		manga_id VARCHAR(255) NOT NULL,
		chapter_id VARCHAR(255) NOT NULL,
		current_page INTEGER NOT NULL DEFAULT 1,
		total_pages INTEGER,
		last_read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, chapter_id)
	);
	CREATE INDEX IF NOT EXISTS idx_reading_progress_user_id ON user_reading_progress(user_id);
	CREATE INDEX IF NOT EXISTS idx_reading_progress_manga_id ON user_reading_progress(manga_id);
	CREATE INDEX IF NOT EXISTS idx_reading_progress_user_manga ON user_reading_progress(user_id, manga_id);
	CREATE INDEX IF NOT EXISTS idx_reading_progress_last_read ON user_reading_progress(last_read_at);
	`

	// Execute migrations
	migrations := []struct {
		name  string
		query string
	}{
		{"users_table", createUsersTable},
		{"sessions_table", createSessionsTable},
		{"manga_table", createMangaTable},
		{"chapters_table", createChaptersTable},
		{"chapter_pages_table", createChapterPagesTable},
		{"reading_progress_table", createReadingProgressTable},
	}

	for _, migration := range migrations {
		_, err := DB.Exec(migration.query)
		if err != nil {
			utils.Logger.Fatal("Failed to run migration",
				zap.String("migration", migration.name),
				zap.Error(err))
		}
		utils.Logger.Info("Migration executed successfully",
			zap.String("migration", migration.name))
	}

	utils.Logger.Info("All database migrations completed successfully")
}

