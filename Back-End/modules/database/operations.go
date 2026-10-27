package database

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"TaleHaven-Backend/modules/utils"
)

// GetPreferredTranslationLanguage returns a user's saved MangaDex language
// preference. Existing users are treated as preferring all languages.
func GetPreferredTranslationLanguage(userID int) (string, error) {
	var language string
	err := DB.QueryRow(`
		SELECT COALESCE(NULLIF(preferred_translation_language, ''), 'all')
		FROM users WHERE id = $1
	`, userID).Scan(&language)
	if err != nil {
		return "", err
	}
	return language, nil
}

// SetPreferredTranslationLanguage updates a user's saved MangaDex language
// preference. Validation belongs to the authenticated settings controller.
func SetPreferredTranslationLanguage(userID int, language string) error {
	_, err := DB.Exec(`
		UPDATE users
		SET preferred_translation_language = $1
		WHERE id = $2
	`, language, userID)
	return err
}

// GetPreferredReadingMode returns the user's saved reader display mode.
// Existing users keep continuous vertical reading until they choose otherwise.
func GetPreferredReadingMode(userID int) (string, error) {
	var mode string
	err := DB.QueryRow(`
		SELECT COALESCE(NULLIF(preferred_reading_mode, ''), 'vertical')
		FROM users WHERE id = $1
	`, userID).Scan(&mode)
	if err != nil {
		return "", err
	}
	return mode, nil
}

// SetPreferredReadingMode updates a user's saved display mode. The controller
// validates the allowed values before this database operation is called.
func SetPreferredReadingMode(userID int, mode string) error {
	_, err := DB.Exec(`
		UPDATE users
		SET preferred_reading_mode = $1
		WHERE id = $2
	`, mode, userID)
	return err
}

// GetPreferredScanlationGroup returns a user's saved group choice for one
// MangaDex manga and translation language. The absence of a choice is All
// groups, which retains the existing deduplicated chapter-list behavior.
func GetPreferredScanlationGroup(userID int, mangaID, language string) (string, error) {
	var groupID string
	err := DB.QueryRow(`
		SELECT scanlation_group_id
		FROM user_manga_scanlation_group_preferences
		WHERE user_id = $1 AND manga_mangadex_id = $2 AND language = $3
	`, userID, mangaID, language).Scan(&groupID)
	if err == sql.ErrNoRows {
		return "all", nil
	}
	if err != nil {
		return "", err
	}
	return groupID, nil
}

// SetPreferredScanlationGroup stores a user's group choice for one manga and
// language. Saving All groups removes any old explicit choice.
func SetPreferredScanlationGroup(userID int, mangaID, language, groupID string) error {
	if groupID == "all" {
		_, err := DB.Exec(`
			DELETE FROM user_manga_scanlation_group_preferences
			WHERE user_id = $1 AND manga_mangadex_id = $2 AND language = $3
		`, userID, mangaID, language)
		return err
	}

	_, err := DB.Exec(`
		INSERT INTO user_manga_scanlation_group_preferences
			(user_id, manga_mangadex_id, language, scanlation_group_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, manga_mangadex_id, language)
		DO UPDATE SET scanlation_group_id = EXCLUDED.scanlation_group_id,
			updated_at = CURRENT_TIMESTAMP
	`, userID, mangaID, language, groupID)
	return err
}

// SaveManga saves or updates a manga in the database
func SaveManga(manga *Manga) (int, error) {
	query := `
		INSERT INTO manga (
			mangadex_id, title, alternative_titles, description, author, artist,
			tags, status, content_rating, year, last_chapter, last_volume,
			original_language, available_languages, cover_image_path, cover_filename,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, CURRENT_TIMESTAMP)
		ON CONFLICT (mangadex_id) 
		DO UPDATE SET
			title = EXCLUDED.title,
			alternative_titles = EXCLUDED.alternative_titles,
			description = EXCLUDED.description,
			author = EXCLUDED.author,
			artist = EXCLUDED.artist,
			tags = EXCLUDED.tags,
			status = EXCLUDED.status,
			content_rating = EXCLUDED.content_rating,
			year = EXCLUDED.year,
			last_chapter = EXCLUDED.last_chapter,
			last_volume = EXCLUDED.last_volume,
			original_language = EXCLUDED.original_language,
			available_languages = EXCLUDED.available_languages,
			cover_image_path = EXCLUDED.cover_image_path,
			cover_filename = EXCLUDED.cover_filename,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`

	var id int
	err := DB.QueryRow(
		query,
		manga.MangaDexID,
		manga.Title,
		pq.Array(manga.AlternativeTitles),
		manga.Description,
		manga.Author,
		manga.Artist,
		pq.Array(manga.Tags),
		manga.Status,
		manga.ContentRating,
		manga.Year,
		manga.LastChapter,
		manga.LastVolume,
		manga.OriginalLanguage,
		pq.Array(manga.AvailableLanguages),
		manga.CoverImagePath,
		manga.CoverFileName,
	).Scan(&id)

	if err != nil {
		utils.Logger.Error("Failed to save manga", zap.Error(err), zap.String("mangadex_id", manga.MangaDexID))
		return 0, err
	}

	utils.Logger.Info("Manga saved successfully", zap.Int("id", id), zap.String("title", manga.Title))
	return id, nil
}

// GetMangaByMangaDexID retrieves a manga by its MangaDex ID
func GetMangaByMangaDexID(mangadexID string) (*Manga, error) {
	query := `
		SELECT id, mangadex_id, title, alternative_titles, description, author, artist,
		       tags, status, content_rating, year, last_chapter, last_volume,
		       original_language, available_languages, cover_image_path, cover_filename,
		       created_at, updated_at
		FROM manga
		WHERE mangadex_id = $1
	`

	manga := &Manga{}
	err := DB.QueryRow(query, mangadexID).Scan(
		&manga.ID,
		&manga.MangaDexID,
		&manga.Title,
		&manga.AlternativeTitles,
		&manga.Description,
		&manga.Author,
		&manga.Artist,
		&manga.Tags,
		&manga.Status,
		&manga.ContentRating,
		&manga.Year,
		&manga.LastChapter,
		&manga.LastVolume,
		&manga.OriginalLanguage,
		&manga.AvailableLanguages,
		&manga.CoverImagePath,
		&manga.CoverFileName,
		&manga.CreatedAt,
		&manga.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		utils.Logger.Error("Failed to get manga", zap.Error(err), zap.String("mangadex_id", mangadexID))
		return nil, err
	}

	return manga, nil
}

// SaveChapter saves or updates a chapter in the database
func SaveChapter(chapter *Chapter) (int, error) {
	query := `
		INSERT INTO chapters (
			manga_id, chapter_mangadex_id, chapter_number, title, volume,
			translated_language, page_count, folder_path, scanlation_group,
			publish_at, hash, downloaded, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP)
		ON CONFLICT (chapter_mangadex_id)
		DO UPDATE SET
			chapter_number = EXCLUDED.chapter_number,
			title = EXCLUDED.title,
			volume = EXCLUDED.volume,
			page_count = EXCLUDED.page_count,
			folder_path = EXCLUDED.folder_path,
			scanlation_group = EXCLUDED.scanlation_group,
			publish_at = EXCLUDED.publish_at,
			hash = EXCLUDED.hash,
			downloaded = EXCLUDED.downloaded,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`

	var id int
	err := DB.QueryRow(
		query,
		chapter.MangaID,
		chapter.ChapterMangaDexID,
		chapter.ChapterNumber,
		chapter.Title,
		chapter.Volume,
		chapter.TranslatedLanguage,
		chapter.PageCount,
		chapter.FolderPath,
		chapter.ScanlationGroup,
		chapter.PublishAt,
		chapter.Hash,
		chapter.Downloaded,
	).Scan(&id)

	if err != nil {
		utils.Logger.Error("Failed to save chapter", zap.Error(err), zap.String("chapter_id", chapter.ChapterMangaDexID))
		return 0, err
	}

	utils.Logger.Info("Chapter saved successfully", zap.Int("id", id))
	return id, nil
}

// GetChapterByMangaDexID retrieves a chapter by its MangaDex ID
func GetChapterByMangaDexID(mangadexID string) (*Chapter, error) {
	query := `
		SELECT id, manga_id, chapter_mangadex_id, chapter_number, title, volume,
		       translated_language, page_count, folder_path, scanlation_group,
		       publish_at, hash, downloaded, created_at, updated_at
		FROM chapters
		WHERE chapter_mangadex_id = $1
	`

	chapter := &Chapter{}
	err := DB.QueryRow(query, mangadexID).Scan(
		&chapter.ID,
		&chapter.MangaID,
		&chapter.ChapterMangaDexID,
		&chapter.ChapterNumber,
		&chapter.Title,
		&chapter.Volume,
		&chapter.TranslatedLanguage,
		&chapter.PageCount,
		&chapter.FolderPath,
		&chapter.ScanlationGroup,
		&chapter.PublishAt,
		&chapter.Hash,
		&chapter.Downloaded,
		&chapter.CreatedAt,
		&chapter.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		utils.Logger.Error("Failed to get chapter", zap.Error(err), zap.String("mangadex_id", mangadexID))
		return nil, err
	}

	return chapter, nil
}

// SaveChapterPage saves a chapter page record
func SaveChapterPage(page *ChapterPage) error {
	query := `
		INSERT INTO chapter_pages (chapter_id, page_number, file_name, local_path, downloaded)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (chapter_id, page_number)
		DO UPDATE SET
			file_name = EXCLUDED.file_name,
			local_path = EXCLUDED.local_path,
			downloaded = EXCLUDED.downloaded
	`

	_, err := DB.Exec(query, page.ChapterID, page.PageNumber, page.FileName, page.LocalPath, page.Downloaded)
	if err != nil {
		utils.Logger.Error("Failed to save chapter page", zap.Error(err), zap.Int("chapter_id", page.ChapterID))
		return err
	}

	return nil
}

// GetChapterPages retrieves all pages for a chapter
func GetChapterPages(chapterID int) ([]ChapterPage, error) {
	query := `
		SELECT id, chapter_id, page_number, file_name, local_path, downloaded, created_at
		FROM chapter_pages
		WHERE chapter_id = $1
		ORDER BY page_number
	`

	rows, err := DB.Query(query, chapterID)
	if err != nil {
		utils.Logger.Error("Failed to get chapter pages", zap.Error(err), zap.Int("chapter_id", chapterID))
		return nil, err
	}
	defer rows.Close()

	var pages []ChapterPage
	for rows.Next() {
		var page ChapterPage
		err := rows.Scan(
			&page.ID,
			&page.ChapterID,
			&page.PageNumber,
			&page.FileName,
			&page.LocalPath,
			&page.Downloaded,
			&page.CreatedAt,
		)
		if err != nil {
			utils.Logger.Error("Failed to scan chapter page", zap.Error(err))
			continue
		}
		pages = append(pages, page)
	}

	return pages, nil
}

// UpdateChapterDownloadStatus updates the downloaded status of a chapter
func UpdateChapterDownloadStatus(chapterID int, downloaded bool) error {
	query := `UPDATE chapters SET downloaded = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	_, err := DB.Exec(query, downloaded, chapterID)
	if err != nil {
		utils.Logger.Error("Failed to update chapter download status", zap.Error(err), zap.Int("chapter_id", chapterID))
		return err
	}

	return nil
}

// GetChaptersByMangaID retrieves all chapters for a manga
func GetChaptersByMangaID(mangaID int) ([]Chapter, error) {
	query := `
		SELECT id, manga_id, chapter_mangadex_id, chapter_number, title, volume,
		       translated_language, page_count, folder_path, scanlation_group,
		       publish_at, hash, downloaded, created_at, updated_at
		FROM chapters
		WHERE manga_id = $1
		ORDER BY chapter_number
	`

	rows, err := DB.Query(query, mangaID)
	if err != nil {
		utils.Logger.Error("Failed to get chapters", zap.Error(err), zap.Int("manga_id", mangaID))
		return nil, err
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var chapter Chapter
		err := rows.Scan(
			&chapter.ID,
			&chapter.MangaID,
			&chapter.ChapterMangaDexID,
			&chapter.ChapterNumber,
			&chapter.Title,
			&chapter.Volume,
			&chapter.TranslatedLanguage,
			&chapter.PageCount,
			&chapter.FolderPath,
			&chapter.ScanlationGroup,
			&chapter.PublishAt,
			&chapter.Hash,
			&chapter.Downloaded,
			&chapter.CreatedAt,
			&chapter.UpdatedAt,
		)
		if err != nil {
			utils.Logger.Error("Failed to scan chapter", zap.Error(err))
			continue
		}
		chapters = append(chapters, chapter)
	}

	return chapters, nil
}

// GetDownloadProgress returns download statistics for a manga
func GetDownloadProgress(mangaID int) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_chapters,
			SUM(CASE WHEN downloaded THEN 1 ELSE 0 END) as downloaded_chapters,
			SUM(page_count) as total_pages
		FROM chapters
		WHERE manga_id = $1
	`

	var totalChapters, downloadedChapters, totalPages int
	err := DB.QueryRow(query, mangaID).Scan(&totalChapters, &downloadedChapters, &totalPages)
	if err != nil {
		return nil, err
	}

	progress := map[string]interface{}{
		"total_chapters":      totalChapters,
		"downloaded_chapters": downloadedChapters,
		"total_pages":         totalPages,
		"percentage":          0,
	}

	if totalChapters > 0 {
		progress["percentage"] = fmt.Sprintf("%.2f", float64(downloadedChapters)/float64(totalChapters)*100)
	}

	return progress, nil
}

// DownloadedMangaInfo represents minimal manga info for library display
type DownloadedMangaInfo struct {
	MangaDexID     string `json:"mangadex_id" db:"mangadex_id"`
	Title          string `json:"title" db:"title"`
	CoverImagePath string `json:"cover_image_path" db:"cover_image_path"`
}

// UserLibraryMangaInfo is the data needed to render a user's saved manga.
type UserLibraryMangaInfo struct {
	MangaDexID     string `json:"mangadexId"`
	Title          string `json:"title"`
	CoverFileName  string `json:"coverFileName"`
	CoverImagePath string `json:"coverImagePath"`
	AddedAt        string `json:"addedAt"`
}

type UserLibraryAniListInfo struct {
	AniListID      int                  `json:"anilistId"`
	AddedAt        string               `json:"addedAt"`
	ReadingStatus  AniListReadingStatus `json:"readingStatus"`
	PersonalRating *int                 `json:"personalRating"`
}

// AniListReadingStatus is a user's book-level status for a saved AniList
// title. It deliberately does not represent or modify chapter progress.
type AniListReadingStatus string

const (
	AniListPlanToRead       AniListReadingStatus = "PLAN_TO_READ"
	AniListCurrentlyReading AniListReadingStatus = "CURRENTLY_READING"
	AniListOnHold           AniListReadingStatus = "ON_HOLD"
	AniListRead             AniListReadingStatus = "READ"
	AniListDropped          AniListReadingStatus = "DROPPED"
)

func IsValidAniListReadingStatus(status AniListReadingStatus) bool {
	switch status {
	case AniListPlanToRead, AniListCurrentlyReading, AniListOnHold, AniListRead, AniListDropped:
		return true
	default:
		return false
	}
}

func AddAniListMangaToUserLibrary(userID, aniListID int) (bool, error) {
	result, err := DB.Exec(`INSERT INTO user_library_anilist_manga (user_id, anilist_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, aniListID)
	if err != nil {
		return false, fmt.Errorf("failed to add AniList manga to user library: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm AniList library insert: %w", err)
	}
	return rows > 0, nil
}

func RemoveAniListMangaFromUserLibrary(userID, aniListID int) (bool, error) {
	result, err := DB.Exec(`DELETE FROM user_library_anilist_manga WHERE user_id = $1 AND anilist_id = $2`, userID, aniListID)
	if err != nil {
		return false, fmt.Errorf("failed to remove AniList manga from user library: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm AniList library deletion: %w", err)
	}
	return rows > 0, nil
}

// IsAniListMangaInUserLibrary reports whether a reader has saved an AniList
// title, which is required before the title can be favorited.
func IsAniListMangaInUserLibrary(userID, aniListID int) (bool, error) {
	var saved bool
	if err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM user_library_anilist_manga WHERE user_id = $1 AND anilist_id = $2)`, userID, aniListID).Scan(&saved); err != nil {
		return false, fmt.Errorf("failed to check AniList library membership: %w", err)
	}
	return saved, nil
}

func GetAniListFavoriteIDs(userID int) ([]int, error) {
	rows, err := DB.Query(`
		SELECT favorite.anilist_id
		FROM user_favorite_anilist_manga favorite
		INNER JOIN user_library_anilist_manga library
			ON library.user_id = favorite.user_id
			AND library.anilist_id = favorite.anilist_id
		WHERE favorite.user_id = $1
		ORDER BY favorite.added_at DESC, favorite.anilist_id ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve AniList favorites: %w", err)
	}
	defer rows.Close()
	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to read AniList favorite: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while retrieving AniList favorites: %w", err)
	}
	return ids, nil
}

func AddAniListMangaToFavorites(userID, aniListID int) (bool, error) {
	result, err := DB.Exec(`INSERT INTO user_favorite_anilist_manga (user_id, anilist_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, aniListID)
	if err != nil {
		return false, fmt.Errorf("failed to add AniList manga to favorites: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm AniList favorite insert: %w", err)
	}
	return rows > 0, nil
}

func RemoveAniListMangaFromFavorites(userID, aniListID int) (bool, error) {
	result, err := DB.Exec(`DELETE FROM user_favorite_anilist_manga WHERE user_id = $1 AND anilist_id = $2`, userID, aniListID)
	if err != nil {
		return false, fmt.Errorf("failed to remove AniList manga from favorites: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm AniList favorite deletion: %w", err)
	}
	return rows > 0, nil
}

// FavoriteAniListCharacter is a reader's saved character plus the manga role
// that was shown when they favorited it.
type FavoriteAniListCharacter struct {
	CharacterID          int
	SourceAniListMangaID int
	Role                 string
}

func GetAniListFavoriteCharacters(userID int) ([]FavoriteAniListCharacter, error) {
	rows, err := DB.Query(`
		SELECT character_id, source_anilist_manga_id, role
		FROM user_favorite_anilist_character
		WHERE user_id = $1
		ORDER BY added_at DESC, character_id ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve AniList character favorites: %w", err)
	}
	defer rows.Close()

	favorites := make([]FavoriteAniListCharacter, 0)
	for rows.Next() {
		var favorite FavoriteAniListCharacter
		if err := rows.Scan(&favorite.CharacterID, &favorite.SourceAniListMangaID, &favorite.Role); err != nil {
			return nil, fmt.Errorf("failed to read AniList character favorite: %w", err)
		}
		favorites = append(favorites, favorite)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while retrieving AniList character favorites: %w", err)
	}
	return favorites, nil
}

func AddAniListCharacterToFavorites(userID, characterID, sourceAniListMangaID int, role string) (bool, error) {
	result, err := DB.Exec(`
		INSERT INTO user_favorite_anilist_character (user_id, character_id, source_anilist_manga_id, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`, userID, characterID, sourceAniListMangaID, role)
	if err != nil {
		return false, fmt.Errorf("failed to add AniList character to favorites: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm AniList character favorite insert: %w", err)
	}
	return rows > 0, nil
}

func RemoveAniListCharacterFromFavorites(userID, characterID int) (bool, error) {
	result, err := DB.Exec(`DELETE FROM user_favorite_anilist_character WHERE user_id = $1 AND character_id = $2`, userID, characterID)
	if err != nil {
		return false, fmt.Errorf("failed to remove AniList character from favorites: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm AniList character favorite deletion: %w", err)
	}
	return rows > 0, nil
}

func GetMangaDexFavoriteIDs(userID int) ([]string, error) {
	rows, err := DB.Query(`
		SELECT manga.mangadex_id
		FROM user_favorite_manga favorite
		JOIN manga ON manga.id = favorite.manga_id
		INNER JOIN user_library_manga library
			ON library.user_id = favorite.user_id
			AND library.manga_id = favorite.manga_id
		WHERE favorite.user_id = $1
		ORDER BY favorite.added_at DESC, manga.mangadex_id ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve MangaDex favorites: %w", err)
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to read MangaDex favorite: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while retrieving MangaDex favorites: %w", err)
	}
	return ids, nil
}

// GetMangaDexFavoriteManga returns the stored display metadata for one user's
// MangaDex favorites, newest first.
func GetMangaDexFavoriteManga(userID int) ([]UserLibraryMangaInfo, error) {
	rows, err := DB.Query(`
		SELECT m.mangadex_id,
		       m.title,
		       COALESCE(m.cover_filename, ''),
		       COALESCE(m.cover_image_path, ''),
		       TO_CHAR(ufm.added_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM user_favorite_manga ufm
		INNER JOIN manga m ON m.id = ufm.manga_id
		INNER JOIN user_library_manga ulm
			ON ulm.user_id = ufm.user_id
			AND ulm.manga_id = ufm.manga_id
		WHERE ufm.user_id = $1
		ORDER BY ufm.added_at DESC, m.title ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve MangaDex favorite manga: %w", err)
	}
	defer rows.Close()

	favorites := make([]UserLibraryMangaInfo, 0)
	for rows.Next() {
		var manga UserLibraryMangaInfo
		if err := rows.Scan(&manga.MangaDexID, &manga.Title, &manga.CoverFileName, &manga.CoverImagePath, &manga.AddedAt); err != nil {
			return nil, fmt.Errorf("failed to read MangaDex favorite manga: %w", err)
		}
		favorites = append(favorites, manga)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading MangaDex favorite manga: %w", err)
	}

	return favorites, nil
}

func AddMangaToFavorites(userID, mangaID int) (bool, error) {
	result, err := DB.Exec(`INSERT INTO user_favorite_manga (user_id, manga_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, mangaID)
	if err != nil {
		return false, fmt.Errorf("failed to add manga to favorites: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm manga favorite insert: %w", err)
	}
	return rows > 0, nil
}

// IsMangaDexMangaInUserLibrary reports whether a reader has saved a MangaDex
// title, which is required before the title can be favorited.
func IsMangaDexMangaInUserLibrary(userID int, mangaDexID string) (bool, error) {
	var saved bool
	if err := DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM user_library_manga library
			INNER JOIN manga ON manga.id = library.manga_id
			WHERE library.user_id = $1 AND manga.mangadex_id = $2
		)
	`, userID, mangaDexID).Scan(&saved); err != nil {
		return false, fmt.Errorf("failed to check MangaDex library membership: %w", err)
	}
	return saved, nil
}

func RemoveMangaFromFavorites(userID int, mangaDexID string) (bool, error) {
	result, err := DB.Exec(`
		DELETE FROM user_favorite_manga
		USING manga
		WHERE user_favorite_manga.manga_id = manga.id
			AND user_favorite_manga.user_id = $1
			AND manga.mangadex_id = $2
	`, userID, mangaDexID)
	if err != nil {
		return false, fmt.Errorf("failed to remove manga from favorites: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm manga favorite deletion: %w", err)
	}
	return rows > 0, nil
}

func GetAniListUserLibraryPage(userID, limit, offset int) ([]UserLibraryAniListInfo, int, error) {
	var total int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM user_library_anilist_manga WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count AniList library: %w", err)
	}
	rows, err := DB.Query(`SELECT anilist_id, TO_CHAR(added_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), reading_status, personal_rating FROM user_library_anilist_manga WHERE user_id = $1 ORDER BY added_at DESC, anilist_id ASC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve AniList library: %w", err)
	}
	defer rows.Close()
	items := make([]UserLibraryAniListInfo, 0, limit)
	for rows.Next() {
		var item UserLibraryAniListInfo
		if err := rows.Scan(&item.AniListID, &item.AddedAt, &item.ReadingStatus, &item.PersonalRating); err != nil {
			return nil, 0, fmt.Errorf("failed to read AniList library: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed while reading AniList library: %w", err)
	}
	return items, total, nil
}

// GetAniListUserLibraryIDsByStatus returns IDs for one user's saved titles in
// a book-level status. It does not interact with chapter progress.
func GetAniListUserLibraryIDsByStatus(userID int, status AniListReadingStatus) ([]int, error) {
	if !IsValidAniListReadingStatus(status) {
		return nil, fmt.Errorf("invalid AniList reading status %q", status)
	}
	rows, err := DB.Query(`SELECT anilist_id FROM user_library_anilist_manga WHERE user_id = $1 AND reading_status = $2 ORDER BY anilist_id ASC`, userID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve AniList library IDs by status: %w", err)
	}
	defer rows.Close()
	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to read AniList library ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while retrieving AniList library IDs: %w", err)
	}
	return ids, nil
}

// GetAniListUserLibraryStatusCounts returns book-level reading-status totals
// for one user. It is deliberately separate from chapter progress.
func GetAniListUserLibraryStatusCounts(userID int) (map[AniListReadingStatus]int, error) {
	counts := map[AniListReadingStatus]int{
		AniListPlanToRead:       0,
		AniListCurrentlyReading: 0,
		AniListOnHold:           0,
		AniListRead:             0,
		AniListDropped:          0,
	}
	rows, err := DB.Query(`SELECT reading_status, COUNT(*) FROM user_library_anilist_manga WHERE user_id = $1 GROUP BY reading_status`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count AniList library statuses: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status AniListReadingStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to read AniList library status count: %w", err)
		}
		if IsValidAniListReadingStatus(status) {
			counts[status] = count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while counting AniList library statuses: %w", err)
	}
	return counts, nil
}

// GetLastCompletedAniListMangaID returns the saved title most recently marked
// Completed by the user. The boolean is false when no completed title exists.
func GetLastCompletedAniListMangaID(userID int) (int, bool, error) {
	var aniListID int
	err := DB.QueryRow(`
		SELECT anilist_id
		FROM user_library_anilist_manga
		WHERE user_id = $1 AND reading_status = $2 AND completed_at IS NOT NULL
		ORDER BY completed_at DESC, anilist_id ASC
		LIMIT 1
	`, userID, AniListRead).Scan(&aniListID)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to retrieve last completed AniList manga: %w", err)
	}
	return aniListID, true, nil
}

// SetAniListMangaReadingStatus changes the book-level state of a saved title
// for one user. The boolean is false when that user does not own the title.
func SetAniListMangaReadingStatus(userID, aniListID int, status AniListReadingStatus) (bool, error) {
	if !IsValidAniListReadingStatus(status) {
		return false, fmt.Errorf("invalid AniList reading status %q", status)
	}
	result, err := DB.Exec(`
		UPDATE user_library_anilist_manga
		SET reading_status = $3,
			completed_at = CASE
				WHEN $3::varchar = 'READ' AND reading_status <> 'READ' THEN CURRENT_TIMESTAMP
				WHEN $3::varchar <> 'READ' THEN NULL
				ELSE completed_at
			END
		WHERE user_id = $1 AND anilist_id = $2
	`, userID, aniListID, status)
	if err != nil {
		return false, fmt.Errorf("failed to update AniList manga reading status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm AniList manga reading status update: %w", err)
	}
	return rows > 0, nil
}

// SetAniListMangaPersonalRating changes a user's optional 1–10 score for a
// saved title. A nil rating clears the user's score.
func SetAniListMangaPersonalRating(userID, aniListID int, rating *int) (bool, error) {
	if rating != nil && (*rating < 1 || *rating > 10) {
		return false, fmt.Errorf("personal rating must be between 1 and 10")
	}
	result, err := DB.Exec(`UPDATE user_library_anilist_manga SET personal_rating = $3 WHERE user_id = $1 AND anilist_id = $2`, userID, aniListID, rating)
	if err != nil {
		return false, fmt.Errorf("failed to update AniList manga personal rating: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm AniList manga personal rating update: %w", err)
	}
	return rows > 0, nil
}

// AddMangaToUserLibrary saves a manga to a user's library. The operation is
// idempotent so repeated clicks cannot create duplicate entries.
func AddMangaToUserLibrary(userID, mangaID int) (bool, error) {
	result, err := DB.Exec(`
		INSERT INTO user_library_manga (user_id, manga_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, manga_id) DO NOTHING
	`, userID, mangaID)
	if err != nil {
		return false, fmt.Errorf("failed to add manga to user library: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm library insert: %w", err)
	}

	return rowsAffected > 0, nil
}

// RemoveMangaFromUserLibrary removes a saved manga. It is idempotent so the
// desired end state is the same even when a client retries the request.
func RemoveMangaFromUserLibrary(userID int, mangaDexID string) (bool, error) {
	result, err := DB.Exec(`
		DELETE FROM user_library_manga ulm
		USING manga m
		WHERE ulm.manga_id = m.id
		  AND ulm.user_id = $1
		  AND m.mangadex_id = $2
	`, userID, mangaDexID)
	if err != nil {
		return false, fmt.Errorf("failed to remove manga from user library: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm library deletion: %w", err)
	}

	return rowsAffected > 0, nil
}

// GetUserLibraryPage returns a page of manga explicitly saved by one user,
// newest first, together with the full saved-title count.
func GetUserLibraryPage(userID, limit, offset int) ([]UserLibraryMangaInfo, int, error) {
	var total int
	if err := DB.QueryRow("SELECT COUNT(*) FROM user_library_manga WHERE user_id = $1", userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count user library: %w", err)
	}

	rows, err := DB.Query(`
		SELECT m.mangadex_id,
		       m.title,
		       COALESCE(m.cover_filename, ''),
		       COALESCE(m.cover_image_path, ''),
		       TO_CHAR(ulm.added_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM user_library_manga ulm
		INNER JOIN manga m ON m.id = ulm.manga_id
		WHERE ulm.user_id = $1
		ORDER BY ulm.added_at DESC, m.title ASC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve user library: %w", err)
	}
	defer rows.Close()

	library := make([]UserLibraryMangaInfo, 0)
	for rows.Next() {
		var manga UserLibraryMangaInfo
		if err := rows.Scan(&manga.MangaDexID, &manga.Title, &manga.CoverFileName, &manga.CoverImagePath, &manga.AddedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to read user library manga: %w", err)
		}
		library = append(library, manga)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed while reading user library: %w", err)
	}

	return library, total, nil
}

// GetAdminDashboardStats returns aggregate site counts for the administrator
// dashboard. All values are intentionally global counts rather than lists of
// people, credentials, sessions, or other sensitive records.
func GetAdminDashboardStats() (AdminDashboardStats, error) {
	var stats AdminDashboardStats
	err := DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM user_library_anilist_manga),
			(SELECT COUNT(*) FROM user_library_manga),
			(
				(SELECT COUNT(*) FROM user_favorite_anilist_manga) +
				(SELECT COUNT(*) FROM user_favorite_manga) +
				(SELECT COUNT(*) FROM user_favorite_anilist_character)
			),
			(SELECT COUNT(*) FROM user_reading_progress)
	`).Scan(
		&stats.Users,
		&stats.AniListLibrary,
		&stats.MangaDexLibrary,
		&stats.Favorites,
		&stats.TrackedChapters,
	)
	if err != nil {
		return AdminDashboardStats{}, fmt.Errorf("failed to retrieve administrator dashboard statistics: %w", err)
	}
	return stats, nil
}

// CreateChatRoom creates a room and makes its creator the first member.
func CreateChatRoom(ownerUserID int, name, visibility string) (ChatRoomSummary, error) {
	tx, err := DB.Begin()
	if err != nil {
		return ChatRoomSummary{}, fmt.Errorf("begin chat room transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var room ChatRoomSummary
	err = tx.QueryRow(`
		INSERT INTO chat_rooms (name, visibility, owner_user_id)
		VALUES ($1, $2, $3)
		RETURNING id, owner_user_id, TO_CHAR(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	`, name, visibility, ownerUserID).Scan(&room.ID, &room.OwnerUserID, &room.CreatedAt)
	if err != nil {
		return ChatRoomSummary{}, fmt.Errorf("create chat room: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO chat_room_members (room_id, user_id) VALUES ($1, $2)`, room.ID, ownerUserID); err != nil {
		return ChatRoomSummary{}, fmt.Errorf("add chat room owner as member: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ChatRoomSummary{}, fmt.Errorf("commit chat room transaction: %w", err)
	}
	room.Name = name
	room.Visibility = visibility
	room.ImagePath = ""
	room.Theme = DisplayThemeMidnight
	room.CustomThemeColor = ""
	room.ThemeImagePath = ""
	room.MemberCount = 1
	room.IsMember = true
	room.IsOwner = true
	return room, nil
}

// ListChatRoomsForUser exposes all public rooms and private rooms the user has
// joined. Invite-only room names therefore never appear in public discovery.
func ListChatRoomsForUser(userID int) ([]ChatRoomSummary, error) {
	rows, err := DB.Query(`
		SELECT r.id, r.name, r.visibility, COALESCE(r.image_path, ''), COALESCE(r.theme, 'midnight'), COALESCE(r.custom_theme_color, ''), COALESCE(r.theme_image_path, ''), r.owner_user_id,
			COALESCE(NULLIF(owner.display_name, ''), split_part(owner.email, '@', 1)),
			COUNT(member.user_id),
			EXISTS(SELECT 1 FROM chat_room_members mine WHERE mine.room_id = r.id AND mine.user_id = $1),
			r.owner_user_id = $1,
			TO_CHAR(r.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM chat_rooms r
		INNER JOIN users owner ON owner.id = r.owner_user_id
		LEFT JOIN chat_room_members member ON member.room_id = r.id
		WHERE r.visibility = $2
			OR EXISTS(SELECT 1 FROM chat_room_members mine WHERE mine.room_id = r.id AND mine.user_id = $1)
		GROUP BY r.id, owner.id
		ORDER BY r.created_at DESC, r.id DESC
	`, userID, ChatRoomPublic)
	if err != nil {
		return nil, fmt.Errorf("list chat rooms: %w", err)
	}
	defer rows.Close()

	rooms := make([]ChatRoomSummary, 0)
	for rows.Next() {
		var room ChatRoomSummary
		if err := rows.Scan(&room.ID, &room.Name, &room.Visibility, &room.ImagePath, &room.Theme, &room.CustomThemeColor, &room.ThemeImagePath, &room.OwnerUserID, &room.OwnerDisplayName, &room.MemberCount, &room.IsMember, &room.IsOwner, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat room: %w", err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat rooms: %w", err)
	}
	return rooms, nil
}

// GetChatRoomForMember retrieves one room only if the caller is a member.
func GetChatRoomForMember(roomID int64, userID int) (ChatRoomSummary, error) {
	var room ChatRoomSummary
	err := DB.QueryRow(`
		SELECT r.id, r.name, r.visibility, COALESCE(r.image_path, ''), COALESCE(r.theme, 'midnight'), COALESCE(r.custom_theme_color, ''), COALESCE(r.theme_image_path, ''), r.owner_user_id,
			COALESCE(NULLIF(owner.display_name, ''), split_part(owner.email, '@', 1)),
			(SELECT COUNT(*) FROM chat_room_members member WHERE member.room_id = r.id),
			true, r.owner_user_id = $2,
			TO_CHAR(r.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM chat_rooms r
		INNER JOIN users owner ON owner.id = r.owner_user_id
		INNER JOIN chat_room_members mine ON mine.room_id = r.id AND mine.user_id = $2
		WHERE r.id = $1
	`, roomID, userID).Scan(&room.ID, &room.Name, &room.Visibility, &room.ImagePath, &room.Theme, &room.CustomThemeColor, &room.ThemeImagePath, &room.OwnerUserID, &room.OwnerDisplayName, &room.MemberCount, &room.IsMember, &room.IsOwner, &room.CreatedAt)
	if err == sql.ErrNoRows {
		return ChatRoomSummary{}, ErrChatRoomAccess
	}
	if err != nil {
		return ChatRoomSummary{}, fmt.Errorf("get chat room: %w", err)
	}
	return room, nil
}

// JoinPublicChatRoom adds a user to a public room. Invite-only rooms cannot be
// joined through this route even when their numeric ID is known.
func JoinPublicChatRoom(roomID int64, userID int) (bool, error) {
	var visibility string
	if err := DB.QueryRow(`SELECT visibility FROM chat_rooms WHERE id = $1`, roomID).Scan(&visibility); err != nil {
		if err == sql.ErrNoRows {
			return false, ErrChatRoomNotFound
		}
		return false, fmt.Errorf("load chat room visibility: %w", err)
	}
	if visibility != ChatRoomPublic {
		return false, ErrChatInviteRequired
	}
	result, err := DB.Exec(`INSERT INTO chat_room_members (room_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, roomID, userID)
	if err != nil {
		return false, fmt.Errorf("join public chat room: %w", err)
	}
	added, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read public chat join result: %w", err)
	}
	return added > 0, nil
}

// UpdateChatRoomVisibility changes discoverability for future members. It does
// not alter the membership table, so everyone already in the room retains
// their access when an owner makes it invite-only.
func UpdateChatRoomVisibility(roomID int64, ownerUserID int, visibility string) error {
	result, err := DB.Exec(`
		UPDATE chat_rooms
		SET visibility = $1
		WHERE id = $2 AND owner_user_id = $3
	`, visibility, roomID, ownerUserID)
	if err != nil {
		return fmt.Errorf("update chat room visibility: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read chat room visibility result: %w", err)
	}
	if updated == 0 {
		return ErrChatRoomAccess
	}
	return nil
}

// UpdateChatRoomImagePath replaces a room's cover image path. Authorization is
// intentionally checked in the same query so only the room owner can alter it.
func UpdateChatRoomImagePath(roomID int64, ownerUserID int, imagePath string) error {
	result, err := DB.Exec(`
		UPDATE chat_rooms
		SET image_path = $1
		WHERE id = $2 AND owner_user_id = $3
	`, imagePath, roomID, ownerUserID)
	if err != nil {
		return fmt.Errorf("update chat room image: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read chat room image result: %w", err)
	}
	if updated == 0 {
		return ErrChatRoomAccess
	}
	return nil
}

// UpdateChatRoomTheme changes the presentation palette for a room, and is
// intentionally owner-only just like its cover image and visibility.
func UpdateChatRoomTheme(roomID int64, ownerUserID int, theme, customThemeColor string) error {
	result, err := DB.Exec(`
		UPDATE chat_rooms
		SET theme = $1, custom_theme_color = $2
		WHERE id = $3 AND owner_user_id = $4
	`, theme, customThemeColor, roomID, ownerUserID)
	if err != nil {
		return fmt.Errorf("update chat room theme: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read chat room theme result: %w", err)
	}
	if updated == 0 {
		return ErrChatRoomAccess
	}
	return nil
}

// UpdateChatRoomThemeImagePath replaces a room's presentation background. The
// authorization scope matches all other owner-only room presentation settings.
func UpdateChatRoomThemeImagePath(roomID int64, ownerUserID int, imagePath string) error {
	result, err := DB.Exec(`
		UPDATE chat_rooms
		SET theme_image_path = $1
		WHERE id = $2 AND owner_user_id = $3
	`, imagePath, roomID, ownerUserID)
	if err != nil {
		return fmt.Errorf("update chat room theme image: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read chat room theme image result: %w", err)
	}
	if updated == 0 {
		return ErrChatRoomAccess
	}
	return nil
}

// CreateChatRoomInvite creates a reusable invite record for a room owner.
// Only the SHA-256 hash of the browser-facing secret is stored.
func CreateChatRoomInvite(roomID int64, ownerUserID int, tokenHash string) error {
	result, err := DB.Exec(`
		INSERT INTO chat_room_invites (room_id, created_by_user_id, token_hash)
		SELECT $1, $2, $3
		WHERE EXISTS (SELECT 1 FROM chat_rooms WHERE id = $1 AND owner_user_id = $2)
	`, roomID, ownerUserID, tokenHash)
	if err != nil {
		return fmt.Errorf("create chat room invite: %w", err)
	}
	created, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read chat invite result: %w", err)
	}
	if created == 0 {
		return ErrChatRoomAccess
	}
	return nil
}

// JoinChatRoomWithInvite adds a user to the room represented by an invite hash.
func JoinChatRoomWithInvite(tokenHash string, userID int) (int64, bool, error) {
	var roomID int64
	if err := DB.QueryRow(`SELECT room_id FROM chat_room_invites WHERE token_hash = $1`, tokenHash).Scan(&roomID); err != nil {
		if err == sql.ErrNoRows {
			return 0, false, ErrChatInviteInvalid
		}
		return 0, false, fmt.Errorf("load chat invitation: %w", err)
	}
	result, err := DB.Exec(`INSERT INTO chat_room_members (room_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, roomID, userID)
	if err != nil {
		return 0, false, fmt.Errorf("join invited chat room: %w", err)
	}
	added, err := result.RowsAffected()
	if err != nil {
		return 0, false, fmt.Errorf("read invite join result: %w", err)
	}
	return roomID, added > 0, nil
}

// GetChatMessages returns recent chronological messages for a room member.
func GetChatMessages(roomID int64, userID int, limit int) ([]ChatMessage, error) {
	if _, err := GetChatRoomForMember(roomID, userID); err != nil {
		return nil, err
	}
	rows, err := DB.Query(`
		SELECT m.id, m.room_id, m.user_id,
			COALESCE(NULLIF(u.display_name, ''), split_part(u.email, '@', 1)),
			m.body, COALESCE(m.image_path, ''),
			TO_CHAR(m.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM chat_messages m
		INNER JOIN users u ON u.id = m.user_id
		WHERE m.room_id = $1
		ORDER BY m.id DESC
		LIMIT $2
	`, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("list chat messages: %w", err)
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var message ChatMessage
		if err := rows.Scan(&message.ID, &message.RoomID, &message.UserID, &message.AuthorDisplayName, &message.Body, &message.ImagePath, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat messages: %w", err)
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
	return messages, nil
}

// CreateChatMessage persists a message only for an existing room member.
func CreateChatMessage(roomID int64, userID int, body string, imagePath string) (ChatMessage, error) {
	var message ChatMessage
	err := DB.QueryRow(`
		INSERT INTO chat_messages (room_id, user_id, body, image_path)
		SELECT $1, $2, $3, $4
		WHERE EXISTS (SELECT 1 FROM chat_room_members WHERE room_id = $1 AND user_id = $2)
		RETURNING id, room_id, user_id, body, image_path,
			TO_CHAR(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	`, roomID, userID, body, imagePath).Scan(&message.ID, &message.RoomID, &message.UserID, &message.Body, &message.ImagePath, &message.CreatedAt)
	if err == sql.ErrNoRows {
		return ChatMessage{}, ErrChatRoomAccess
	}
	if err != nil {
		return ChatMessage{}, fmt.Errorf("create chat message: %w", err)
	}
	return message, nil
}

// GetDownloadedMangaWithCovers returns all manga that have downloaded chapters and cover images
func GetDownloadedMangaWithCovers() ([]DownloadedMangaInfo, error) {
	query := `
		SELECT DISTINCT m.mangadex_id, m.title, m.cover_image_path
		FROM manga m
		INNER JOIN chapters c ON m.id = c.manga_id
		WHERE c.downloaded = true
		AND m.cover_image_path IS NOT NULL
		AND m.cover_image_path != ''
		ORDER BY m.title ASC
	`

	rows, err := DB.Query(query)
	if err != nil {
		utils.Logger.Error("Failed to get downloaded manga with covers", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var mangaList []DownloadedMangaInfo
	for rows.Next() {
		var manga DownloadedMangaInfo
		err := rows.Scan(&manga.MangaDexID, &manga.Title, &manga.CoverImagePath)
		if err != nil {
			utils.Logger.Error("Failed to scan downloaded manga", zap.Error(err))
			continue
		}
		mangaList = append(mangaList, manga)
	}

	utils.Logger.Info("Retrieved downloaded manga with covers", zap.Int("count", len(mangaList)))
	return mangaList, nil
}
