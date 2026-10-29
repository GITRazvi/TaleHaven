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
		role VARCHAR(32) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
		display_name VARCHAR(80) NOT NULL DEFAULT '',
		profile_image_path VARCHAR(512) NOT NULL DEFAULT '',
		display_theme VARCHAR(24) NOT NULL DEFAULT 'midnight',
		custom_theme_color VARCHAR(7) NOT NULL DEFAULT '',
		theme_image_path VARCHAR(512) NOT NULL DEFAULT '',
		site_theme VARCHAR(24) NOT NULL DEFAULT 'midnight',
		site_custom_color VARCHAR(7) NOT NULL DEFAULT '',
		site_theme_image_paths TEXT[] NOT NULL DEFAULT '{}',
		site_theme_image_rotation_seconds INTEGER NOT NULL DEFAULT 0,
		site_theme_image_rotation_mode VARCHAR(16) NOT NULL DEFAULT 'static',
		site_theme_image_rotation_minutes INTEGER NOT NULL DEFAULT 15,
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

	addSecureSessionFields := `
	ALTER TABLE users
		ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'user'
		CHECK (role IN ('user', 'admin'));
	ALTER TABLE sessions
		ADD COLUMN IF NOT EXISTS verifier_hash VARCHAR(64),
		ADD COLUMN IF NOT EXISTS idle_expires_at TIMESTAMP,
		ADD COLUMN IF NOT EXISTS absolute_expires_at TIMESTAMP,
		ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMP;
	CREATE INDEX IF NOT EXISTS idx_sessions_active_lookup
		ON sessions(id, idle_expires_at, absolute_expires_at)
		WHERE revoked_at IS NULL;
	`

	addUserProfileFields := `
	ALTER TABLE users
		ADD COLUMN IF NOT EXISTS display_name VARCHAR(80) NOT NULL DEFAULT '',
		ADD COLUMN IF NOT EXISTS profile_image_path VARCHAR(512) NOT NULL DEFAULT '';
	UPDATE users
		SET display_name = split_part(email, '@', 1)
		WHERE display_name = '';
	`

	addDisplayThemeFields := `
	ALTER TABLE users
		ADD COLUMN IF NOT EXISTS display_theme VARCHAR(24) NOT NULL DEFAULT 'midnight',
		ADD COLUMN IF NOT EXISTS custom_theme_color VARCHAR(7) NOT NULL DEFAULT '',
		ADD COLUMN IF NOT EXISTS theme_image_path VARCHAR(512) NOT NULL DEFAULT '',
		ADD COLUMN IF NOT EXISTS site_theme VARCHAR(24) NOT NULL DEFAULT 'midnight',
		ADD COLUMN IF NOT EXISTS site_custom_color VARCHAR(7) NOT NULL DEFAULT '',
		ADD COLUMN IF NOT EXISTS site_theme_image_paths TEXT[] NOT NULL DEFAULT '{}',
		ADD COLUMN IF NOT EXISTS site_theme_image_rotation_seconds INTEGER NOT NULL DEFAULT 0;
	ALTER TABLE users
		ADD COLUMN IF NOT EXISTS site_theme_image_rotation_mode VARCHAR(16) NOT NULL DEFAULT 'static',
		ADD COLUMN IF NOT EXISTS site_theme_image_rotation_minutes INTEGER NOT NULL DEFAULT 15;
	UPDATE users SET display_theme = 'midnight'
		WHERE display_theme IS NULL OR display_theme = '';
	UPDATE users SET site_theme = 'midnight'
		WHERE site_theme IS NULL OR site_theme = '';
	UPDATE users SET site_theme_image_rotation_seconds = 0
		WHERE site_theme_image_rotation_seconds IS NULL
		   OR site_theme_image_rotation_seconds NOT IN (0, 15, 30, 60);
	UPDATE users SET site_theme_image_rotation_mode = 'static'
		WHERE site_theme_image_rotation_mode IS NULL
		   OR site_theme_image_rotation_mode NOT IN ('static', 'refresh', 'session', 'interval');
	UPDATE users SET site_theme_image_rotation_minutes = 15
		WHERE site_theme_image_rotation_minutes IS NULL
		   OR site_theme_image_rotation_minutes < 1
		   OR site_theme_image_rotation_minutes > 1440;
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

	// Create a per-user library. Manga metadata remains shared, while this table
	// records which titles a particular user chose to save.
	createUserLibraryTable := `
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
	`

	createAniListLibraryTable := `
	CREATE TABLE IF NOT EXISTS user_library_anilist_manga (
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		anilist_id INTEGER NOT NULL,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		reading_status VARCHAR(32) NOT NULL DEFAULT 'PLAN_TO_READ' CHECK (reading_status IN ('PLAN_TO_READ', 'CURRENTLY_READING', 'ON_HOLD', 'READ', 'DROPPED')),
		personal_rating SMALLINT CHECK (personal_rating BETWEEN 1 AND 10),
		PRIMARY KEY (user_id, anilist_id)
	);
	CREATE INDEX IF NOT EXISTS idx_user_library_anilist_user_added
		ON user_library_anilist_manga(user_id, added_at DESC);
	`

	addAniListLibraryReadingStatus := `
	ALTER TABLE user_library_anilist_manga
		ADD COLUMN IF NOT EXISTS reading_status VARCHAR(32) NOT NULL DEFAULT 'PLAN_TO_READ'
		CHECK (reading_status IN ('PLAN_TO_READ', 'CURRENTLY_READING', 'ON_HOLD', 'READ', 'DROPPED'));
	CREATE INDEX IF NOT EXISTS idx_user_library_anilist_user_status
		ON user_library_anilist_manga(user_id, reading_status);
	`

	addAniListLibraryPersonalRating := `
	ALTER TABLE user_library_anilist_manga
		ADD COLUMN IF NOT EXISTS personal_rating SMALLINT CHECK (personal_rating BETWEEN 1 AND 10);
	`

	addAniListLibraryCompletedAt := `
	ALTER TABLE user_library_anilist_manga
		ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP;
	UPDATE user_library_anilist_manga
		SET completed_at = added_at
		WHERE reading_status = 'READ' AND completed_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_user_library_anilist_user_completed
		ON user_library_anilist_manga(user_id, completed_at DESC)
		WHERE reading_status = 'READ';
	`

	createFavoriteTables := `
	CREATE TABLE IF NOT EXISTS user_favorite_anilist_manga (
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		anilist_id INTEGER NOT NULL,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, anilist_id)
	);
	CREATE INDEX IF NOT EXISTS idx_user_favorite_anilist_user_added
		ON user_favorite_anilist_manga(user_id, added_at DESC);
	CREATE TABLE IF NOT EXISTS user_favorite_manga (
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		manga_id INTEGER NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_id)
	);
	CREATE INDEX IF NOT EXISTS idx_user_favorite_manga_user_added
		ON user_favorite_manga(user_id, added_at DESC);
	`

	createAniListCharacterFavoriteTable := `
	CREATE TABLE IF NOT EXISTS user_favorite_anilist_character (
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		character_id INTEGER NOT NULL,
		source_anilist_manga_id INTEGER NOT NULL,
		role TEXT NOT NULL DEFAULT '',
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, character_id)
	);
	CREATE INDEX IF NOT EXISTS idx_user_favorite_anilist_character_user_added
		ON user_favorite_anilist_character(user_id, added_at DESC);
	`

	addPreferredTranslationLanguage := `
	ALTER TABLE users
		ADD COLUMN IF NOT EXISTS preferred_translation_language VARCHAR(16) NOT NULL DEFAULT 'all';
	`

	addPreferredReadingMode := `
	ALTER TABLE users
		ADD COLUMN IF NOT EXISTS preferred_reading_mode VARCHAR(16) NOT NULL DEFAULT 'vertical'
		CHECK (preferred_reading_mode IN ('vertical', 'ltr', 'rtl'));
	`

	createScanlationGroupPreferences := `
	CREATE TABLE IF NOT EXISTS user_manga_scanlation_group_preferences (
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		manga_mangadex_id VARCHAR(64) NOT NULL,
		language VARCHAR(16) NOT NULL,
		scanlation_group_id VARCHAR(64) NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_mangadex_id, language)
	);
	`

	createChatTables := `
	CREATE TABLE IF NOT EXISTS chat_rooms (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(80) NOT NULL,
		visibility VARCHAR(16) NOT NULL CHECK (visibility IN ('public', 'invite_only')),
		image_path VARCHAR(512) NOT NULL DEFAULT '',
		theme VARCHAR(24) NOT NULL DEFAULT 'midnight',
		custom_theme_color VARCHAR(7) NOT NULL DEFAULT '',
		theme_image_path VARCHAR(512) NOT NULL DEFAULT '',
		owner_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_chat_rooms_visibility_created
		ON chat_rooms(visibility, created_at DESC);
	CREATE TABLE IF NOT EXISTS chat_room_members (
		room_id BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (room_id, user_id)
	);
	CREATE INDEX IF NOT EXISTS idx_chat_room_members_user
		ON chat_room_members(user_id, joined_at DESC);
	CREATE TABLE IF NOT EXISTS chat_room_invites (
		id BIGSERIAL PRIMARY KEY,
		room_id BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
		created_by_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash VARCHAR(64) NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_chat_room_invites_room
		ON chat_room_invites(room_id, created_at DESC);
	CREATE TABLE IF NOT EXISTS chat_messages (
		id BIGSERIAL PRIMARY KEY,
		room_id BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		body VARCHAR(2000) NOT NULL,
		image_path VARCHAR(512) NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_chat_messages_room_id
		ON chat_messages(room_id, id DESC);
	ALTER TABLE chat_rooms
		ADD COLUMN IF NOT EXISTS image_path VARCHAR(512) NOT NULL DEFAULT '';
	ALTER TABLE chat_messages
		ADD COLUMN IF NOT EXISTS image_path VARCHAR(512) NOT NULL DEFAULT '';
	ALTER TABLE chat_rooms
		ADD COLUMN IF NOT EXISTS theme VARCHAR(24) NOT NULL DEFAULT 'midnight',
		ADD COLUMN IF NOT EXISTS custom_theme_color VARCHAR(7) NOT NULL DEFAULT '',
		ADD COLUMN IF NOT EXISTS theme_image_path VARCHAR(512) NOT NULL DEFAULT '';
	`

	createChatMessageReports := `
	CREATE TABLE IF NOT EXISTS chat_message_reports (
		id BIGSERIAL PRIMARY KEY,
		message_id BIGINT NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
		room_id BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
		reporter_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		reason VARCHAR(500) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (message_id, reporter_user_id)
	);
	CREATE INDEX IF NOT EXISTS idx_chat_message_reports_room_created
		ON chat_message_reports(room_id, created_at DESC);
	`

	addChatMessageReplyTarget := `
	ALTER TABLE chat_messages
		ADD COLUMN IF NOT EXISTS reply_to_message_id BIGINT NOT NULL DEFAULT 0;
	CREATE INDEX IF NOT EXISTS idx_chat_messages_reply_to
		ON chat_messages(reply_to_message_id)
		WHERE reply_to_message_id > 0;
	`

	createAniListMangaComments := `
	CREATE TABLE IF NOT EXISTS anilist_manga_comments (
		id BIGSERIAL PRIMARY KEY,
		anilist_id INTEGER NOT NULL CHECK (anilist_id > 0),
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		body VARCHAR(1000) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_anilist_manga_comments_manga_created
		ON anilist_manga_comments(anilist_id, id DESC);
	`

	createChatRoomModerationTables := `
	CREATE TABLE IF NOT EXISTS chat_room_moderators (
		room_id BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		assigned_by_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (room_id, user_id)
	);
	CREATE INDEX IF NOT EXISTS idx_chat_room_moderators_user
		ON chat_room_moderators(user_id, room_id);
	CREATE TABLE IF NOT EXISTS chat_room_bans (
		room_id BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		banned_by_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (room_id, user_id)
	);
	CREATE INDEX IF NOT EXISTS idx_chat_room_bans_user
		ON chat_room_bans(user_id, room_id);
	`

	createProfileComments := `
	ALTER TABLE users
		ADD COLUMN IF NOT EXISTS profile_comments_enabled BOOLEAN NOT NULL DEFAULT TRUE;
	CREATE TABLE IF NOT EXISTS profile_comments (
		id BIGSERIAL PRIMARY KEY,
		profile_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		body VARCHAR(1000) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_profile_comments_profile_created
		ON profile_comments(profile_user_id, id DESC);
	`
	createNotifications := `
		CREATE TABLE IF NOT EXISTS user_notifications (
			id BIGSERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			title VARCHAR(160) NOT NULL,
			body VARCHAR(500) NOT NULL DEFAULT '',
			target_path VARCHAR(500) NOT NULL DEFAULT '',
			read_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_user_notifications_recent
			ON user_notifications(user_id, id DESC);
		CREATE INDEX IF NOT EXISTS idx_user_notifications_unread
			ON user_notifications(user_id, id DESC) WHERE read_at IS NULL;
	`
	createChatRoomNotificationPreferences := `
		CREATE TABLE IF NOT EXISTS chat_room_notification_preferences (
			room_id BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			muted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (room_id, user_id)
		);
		CREATE INDEX IF NOT EXISTS idx_chat_room_notification_preferences_user
			ON chat_room_notification_preferences(user_id, room_id);
	`

	// Execute migrations
	migrations := []struct {
		name  string
		query string
	}{
		{"users_table", createUsersTable},
		{"sessions_table", createSessionsTable},
		{"secure_session_fields", addSecureSessionFields},
		{"user_profile_fields", addUserProfileFields},
		{"display_theme_fields", addDisplayThemeFields},
		{"manga_table", createMangaTable},
		{"chapters_table", createChaptersTable},
		{"chapter_pages_table", createChapterPagesTable},
		{"reading_progress_table", createReadingProgressTable},
		{"user_library_manga_table", createUserLibraryTable},
		{"user_library_anilist_manga_table", createAniListLibraryTable},
		{"user_library_anilist_reading_status", addAniListLibraryReadingStatus},
		{"user_library_anilist_personal_rating", addAniListLibraryPersonalRating},
		{"user_library_anilist_completed_at", addAniListLibraryCompletedAt},
		{"user_favorite_tables", createFavoriteTables},
		{"user_favorite_anilist_character_table", createAniListCharacterFavoriteTable},
		{"user_preferred_translation_language", addPreferredTranslationLanguage},
		{"user_preferred_reading_mode", addPreferredReadingMode},
		{"user_manga_scanlation_group_preferences", createScanlationGroupPreferences},
		{"chatrooms", createChatTables},
		{"chat_message_reports", createChatMessageReports},
		{"chat_message_reply_target", addChatMessageReplyTarget},
		{"anilist_manga_comments", createAniListMangaComments},
		{"chat_room_moderation", createChatRoomModerationTables},
		{"profile_comments", createProfileComments},
		{"user_notifications", createNotifications},
		{"chat_room_notification_preferences", createChatRoomNotificationPreferences},
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
