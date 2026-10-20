-- TaleHaven Database Schema
-- This file is for reference only. Migrations are automatically run by the Go application.

-- Users Table
-- Stores user authentication and profile information
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Sessions Table
-- Stores user sessions for tracking and authentication
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(64) PRIMARY KEY,  -- Secure random session ID
    user_id INTEGER,  -- NULL for anonymous sessions, FK for logged-in users
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),  -- IPv4 or IPv6
    user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Manga Table
-- Stores manga metadata from MangaDex API
CREATE TABLE IF NOT EXISTS manga (
    id SERIAL PRIMARY KEY,
    mangadex_id VARCHAR(255) UNIQUE NOT NULL,  -- UUID from MangaDex
    title VARCHAR(500) NOT NULL,
    alternative_titles TEXT[],  -- Array of alternative titles
    description TEXT,
    author VARCHAR(255),
    artist VARCHAR(255),
    tags TEXT[],  -- Genre/theme tags (e.g., Action, Romance, Isekai)
    status VARCHAR(50) DEFAULT 'ongoing',  -- 'ongoing', 'completed', 'hiatus', 'cancelled'
    content_rating VARCHAR(50),  -- 'safe', 'suggestive', 'erotica', 'pornographic'
    year INTEGER,
    last_chapter VARCHAR(50),
    last_volume VARCHAR(50),
    original_language VARCHAR(10),
    available_languages TEXT[],  -- Array of available translation languages
    cover_image_path VARCHAR(1000),  -- Local path to downloaded cover image
    cover_filename VARCHAR(255),  -- Original filename from MangaDex
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_manga_title ON manga(title);
CREATE INDEX IF NOT EXISTS idx_manga_mangadex_id ON manga(mangadex_id);

-- User Library Table
-- Links a user to the manga titles they explicitly save in their library.
CREATE TABLE IF NOT EXISTS user_library_manga (
    user_id INTEGER NOT NULL,
    manga_id INTEGER NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, manga_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_library_manga_user_added
    ON user_library_manga(user_id, added_at DESC);

-- Chapters Table
-- Stores chapter metadata with references to downloaded images
CREATE TABLE IF NOT EXISTS chapters (
    id SERIAL PRIMARY KEY,
    manga_id INTEGER NOT NULL,
    chapter_mangadex_id VARCHAR(255) UNIQUE NOT NULL,  -- UUID from MangaDex
    chapter_number VARCHAR(50),  -- Can be "1", "1.5", "extra", etc.
    title VARCHAR(500),
    volume VARCHAR(50),
    translated_language VARCHAR(10) NOT NULL,  -- ISO language code (e.g., 'en', 'es')
    page_count INTEGER DEFAULT 0,
    folder_path VARCHAR(1000) NOT NULL,  -- Local folder path containing chapter images
    scanlation_group VARCHAR(255),  -- Name of the translation group
    publish_at TIMESTAMP,
    hash VARCHAR(255),  -- MangaDex chapter hash for image retrieval
    downloaded BOOLEAN DEFAULT FALSE,  -- Whether images have been downloaded
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chapters_manga_id ON chapters(manga_id);
CREATE INDEX IF NOT EXISTS idx_chapters_mangadex_id ON chapters(chapter_mangadex_id);
CREATE INDEX IF NOT EXISTS idx_chapters_downloaded ON chapters(downloaded);

-- Chapter Pages Table
-- Tracks individual pages within a chapter
CREATE TABLE IF NOT EXISTS chapter_pages (
    id SERIAL PRIMARY KEY,
    chapter_id INTEGER NOT NULL,
    page_number INTEGER NOT NULL,
    file_name VARCHAR(255) NOT NULL,  -- Original filename from MangaDex
    local_path VARCHAR(1000) NOT NULL,  -- Full local path to the downloaded image
    downloaded BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    UNIQUE(chapter_id, page_number)  -- Prevent duplicate pages
);

CREATE INDEX IF NOT EXISTS idx_chapter_pages_chapter_id ON chapter_pages(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_pages_downloaded ON chapter_pages(downloaded);

-- Example Queries:

-- Insert a new user
-- INSERT INTO users (email, password_hash) VALUES ('user@example.com', 'hashed_password_here');

-- Insert a new manga from MangaDex
-- INSERT INTO manga (mangadex_id, title, description, author, artist, tags, status, content_rating, year, cover_image_path)
-- VALUES ('a1c7c817-4e59-43b7-9365-09675a149a6f', 'Naruto', 'A ninja story', 'Kishimoto Masashi', 'Kishimoto Masashi',
--         ARRAY['Action', 'Adventure', 'Shounen'], 'completed', 'safe', 1999, '/images/covers/naruto.jpg');

-- Insert a new chapter
-- INSERT INTO chapters (manga_id, chapter_mangadex_id, chapter_number, title, volume, translated_language, page_count, folder_path, downloaded)
-- VALUES (1, '015979c8-ffa4-4afa-b48e-3da5d0ac87c3', '1', 'Uzumaki Naruto', '1', 'en', 53, '/images/manga/naruto/chapter-1/', true);

-- Insert chapter pages
-- INSERT INTO chapter_pages (chapter_id, page_number, file_name, local_path, downloaded)
-- VALUES (1, 1, 'x1-abc123.jpg', '/images/manga/naruto/chapter-1/page-001.jpg', true);

-- Get all chapters for a manga
-- SELECT * FROM chapters WHERE manga_id = 1 ORDER BY chapter_number;

-- Get all pages for a chapter
-- SELECT * FROM chapter_pages WHERE chapter_id = 1 ORDER BY page_number;

-- Get manga with specific tags
-- SELECT * FROM manga WHERE 'Action' = ANY(tags);

-- Get chapters that haven't been downloaded yet
-- SELECT * FROM chapters WHERE downloaded = FALSE;

-- Get download progress for a manga
-- SELECT m.title, COUNT(c.id) as total_chapters, SUM(CASE WHEN c.downloaded THEN 1 ELSE 0 END) as downloaded_chapters
-- FROM manga m
-- LEFT JOIN chapters c ON m.id = c.manga_id
-- WHERE m.id = 1
-- GROUP BY m.id, m.title;


