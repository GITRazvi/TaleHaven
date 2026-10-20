package database

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"TaleHaven-Backend/modules/utils"
)

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
