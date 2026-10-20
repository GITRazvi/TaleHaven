package mangadex_api

import (
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

const (
	ImagesBaseDir = "images"
	MangaDir      = "manga"
	CoversDir     = "covers"
)

// DownloadChapterImages downloads all images for a chapter and saves them to disk
func DownloadChapterImages(chapterID string, quality string) error {
	utils.Logger.Info("Starting chapter download", zap.String("chapterId", chapterID))

	// Get chapter pages from MangaDex API
	pagesResponse, err := GetChapterPages(chapterID)
	if err != nil {
		return fmt.Errorf("failed to get chapter pages: %w", err)
	}

	// Get chapter details
	chapterResponse, err := GetChapterByID(chapterID)
	if err != nil {
		return fmt.Errorf("failed to get chapter details: %w", err)
	}

	// Get manga ID from chapter relationships
	var mangaID string
	for _, rel := range chapterResponse.Data.Relationships {
		if rel.Type == "manga" {
			mangaID = rel.ID
			break
		}
	}

	if mangaID == "" {
		return fmt.Errorf("manga ID not found in chapter relationships")
	}

	// Get or create manga in database
	dbManga, err := database.GetMangaByMangaDexID(mangaID)
	if err != nil {
		return fmt.Errorf("failed to get manga from database: %w", err)
	}

	// If manga doesn't exist, fetch and save it
	if dbManga == nil {
		mangaResponse, err := GetMangaByID(mangaID)
		if err != nil {
			return fmt.Errorf("failed to fetch manga from API: %w", err)
		}

		dbManga, err = saveMangaToDatabase(mangaResponse)
		if err != nil {
			return fmt.Errorf("failed to save manga: %w", err)
		}
	}

	// Check if chapter already exists in database
	existingChapter, err := database.GetChapterByMangaDexID(chapterID)
	if err != nil {
		return fmt.Errorf("failed to check existing chapter: %w", err)
	}

	// If chapter is already downloaded, skip it
	if existingChapter != nil && existingChapter.Downloaded {
		utils.Logger.Info("Chapter already downloaded, skipping",
			zap.String("chapterId", chapterID),
			zap.String("chapterNumber", chapterResponse.Data.Attributes.Chapter))
		return nil
	}

	// Save or update chapter in database
	dbChapter, err := saveChapterToDatabase(chapterResponse, dbManga.ID)
	if err != nil {
		return fmt.Errorf("failed to save chapter: %w", err)
	}

	// Create folder for chapter images using unique MangaDex IDs
	// This prevents duplicates when multiple versions of same chapter exist
	chapterFolder := filepath.Join(ImagesBaseDir, MangaDir, mangaID, "chapters", chapterID)

	if err := os.MkdirAll(chapterFolder, 0755); err != nil {
		return fmt.Errorf("failed to create chapter folder: %w", err)
	}

	// Update chapter folder path
	dbChapter.FolderPath = chapterFolder
	_, err = database.SaveChapter(dbChapter)
	if err != nil {
		utils.Logger.Error("Failed to update chapter folder path", zap.Error(err))
	}

	// Download each page
	pageURLs := GetPageURLs(pagesResponse, quality)
	pageFileNames := GetPageFileNames(pagesResponse, quality)
	successCount := 0

	for i, pageURL := range pageURLs {
		pageNumber := i + 1
		// Extract file extension from original filename
		originalFileName := pageFileNames[i]
		fileExt := filepath.Ext(originalFileName)
		if fileExt == "" {
			fileExt = ".jpg" // Fallback to jpg if no extension
		}
		fileName := fmt.Sprintf("%d%s", pageNumber, fileExt)
		localPath := filepath.Join(chapterFolder, fileName)

		// Save page info to database
		page := &database.ChapterPage{
			ChapterID:  dbChapter.ID,
			PageNumber: pageNumber,
			FileName:   fileName,
			LocalPath:  localPath,
			Downloaded: false,
		}

		if err := database.SaveChapterPage(page); err != nil {
			utils.Logger.Error("Failed to save page record", zap.Error(err), zap.Int("page", pageNumber))
		}

		// Download the image with retry logic
		if err := downloadImageWithRetry(pageURL, localPath, 3); err != nil {
			utils.Logger.Error("Failed to download page after retries",
				zap.Error(err),
				zap.Int("page", pageNumber),
				zap.String("url", pageURL))
			continue
		}

		// Update page as downloaded
		page.Downloaded = true
		if err := database.SaveChapterPage(page); err != nil {
			utils.Logger.Error("Failed to update page status", zap.Error(err))
		}

		successCount++
		utils.Logger.Info("Downloaded page",
			zap.Int("page", pageNumber),
			zap.Int("total", len(pageURLs)),
			zap.String("path", localPath))

		// Rate limiting - be nice to MangaDex servers
		time.Sleep(200 * time.Millisecond)
	}

	// Mark chapter as downloaded if all pages succeeded
	if successCount == len(pageURLs) {
		if err := database.UpdateChapterDownloadStatus(dbChapter.ID, true); err != nil {
			utils.Logger.Error("Failed to mark chapter as downloaded", zap.Error(err))
		} else {
			utils.Logger.Info("Chapter fully downloaded",
				zap.String("chapterId", chapterID),
				zap.Int("pages", successCount))
		}
	} else {
		utils.Logger.Warn("Chapter partially downloaded",
			zap.String("chapterId", chapterID),
			zap.Int("success", successCount),
			zap.Int("total", len(pageURLs)))
	}

	return nil
}

// DownloadCoverImage downloads a manga cover image
func DownloadCoverImage(mangaID string, quality string) error {
	utils.Logger.Info("Downloading cover image", zap.String("mangaId", mangaID))

	// Get manga details
	mangaResponse, err := GetMangaByID(mangaID)
	if err != nil {
		return fmt.Errorf("failed to get manga: %w", err)
	}

	// Get cover filename
	coverFileName := GetCoverFileName(mangaResponse.Data.Relationships)
	if coverFileName == "" {
		return fmt.Errorf("no cover art found for manga")
	}

	// Build cover URL - fileName should include extension
	// MangaDex format: abc123.jpg becomes abc123.jpg.512.jpg for 512px
	coverURL := BuildCoverURL(mangaID, coverFileName, quality)

	// Create covers directory
	coversDir := filepath.Join(ImagesBaseDir, CoversDir)
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		return fmt.Errorf("failed to create covers directory: %w", err)
	}

	// Local file path
	localFileName := fmt.Sprintf("%s-%s.jpg", sanitizeFileName(mangaResponse.Data.Attributes.Title["en"]), mangaID[:8])
	localPath := filepath.Join(coversDir, localFileName)

	// Download the image
	if err := downloadImage(coverURL, localPath); err != nil {
		return fmt.Errorf("failed to download cover: %w", err)
	}

	utils.Logger.Info("Cover downloaded successfully",
		zap.String("mangaId", mangaID),
		zap.String("path", localPath))

	// Update database with cover path
	dbManga, err := database.GetMangaByMangaDexID(mangaID)
	if err != nil {
		return fmt.Errorf("failed to get manga from database: %w", err)
	}

	if dbManga != nil {
		dbManga.CoverImagePath = sql.NullString{String: localPath, Valid: true}
		dbManga.CoverFileName = sql.NullString{String: coverFileName, Valid: true}
		_, err = database.SaveManga(dbManga)
		if err != nil {
			utils.Logger.Error("Failed to update manga cover path", zap.Error(err))
		}
	}

	return nil
}

// downloadImage downloads an image from URL and saves it to localPath
func downloadImage(url string, localPath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Make request
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

	// Create local file
	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy image data
	bytesWritten, err := io.Copy(out, resp.Body)
	if err != nil {
		// Clean up partial file on error
		os.Remove(localPath)
		return err
	}

	// Validate that we actually downloaded something
	if bytesWritten == 0 {
		os.Remove(localPath)
		return fmt.Errorf("downloaded file is empty")
	}

	return nil
}

// downloadImageWithRetry attempts to download an image with retry logic
func downloadImageWithRetry(url string, localPath string, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Check if file already exists and is valid
		if fileInfo, err := os.Stat(localPath); err == nil && fileInfo.Size() > 0 {
			utils.Logger.Info("File already exists, skipping download",
				zap.String("path", localPath),
				zap.Int64("size", fileInfo.Size()))
			return nil
		}

		err := downloadImage(url, localPath)
		if err == nil {
			// Success! Verify the file was created
			if fileInfo, err := os.Stat(localPath); err == nil && fileInfo.Size() > 0 {
				return nil
			}
			lastErr = fmt.Errorf("file verification failed after download")
		} else {
			lastErr = err
		}

		if attempt < maxRetries {
			// Exponential backoff: 1s, 2s, 4s
			backoffTime := time.Duration(1<<uint(attempt-1)) * time.Second
			utils.Logger.Warn("Download attempt failed, retrying",
				zap.Int("attempt", attempt),
				zap.Int("maxRetries", maxRetries),
				zap.Duration("backoff", backoffTime),
				zap.Error(lastErr))
			time.Sleep(backoffTime)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// saveMangaToDatabase converts MangaDex manga response to database model and saves it
func saveMangaToDatabase(mangaResponse *MangaResponse) (*database.Manga, error) {
	existingManga, err := database.GetMangaByMangaDexID(mangaResponse.Data.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing manga before saving: %w", err)
	}

	manga := &database.Manga{
		MangaDexID:  mangaResponse.Data.ID,
		Title:       getPreferredTitle(mangaResponse.Data.Attributes.Title),
		Description: sql.NullString{String: getPreferredDescription(mangaResponse.Data.Attributes.Description), Valid: true},
		Status:      mangaResponse.Data.Attributes.Status,
	}
	if coverFileName := GetCoverFileName(mangaResponse.Data.Relationships); coverFileName != "" {
		manga.CoverFileName = sql.NullString{String: coverFileName, Valid: true}
	}
	if existingManga != nil {
		// A title may already have a local cover download. Metadata refreshes must
		// retain that path rather than clearing the cached image.
		manga.CoverImagePath = existingManga.CoverImagePath
		if !manga.CoverFileName.Valid {
			manga.CoverFileName = existingManga.CoverFileName
		}
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

	// Save to database
	id, err := database.SaveManga(manga)
	if err != nil {
		return nil, err
	}

	manga.ID = id
	return manga, nil
}

// saveChapterToDatabase converts MangaDex chapter response to database model and saves it
func saveChapterToDatabase(chapterResponse *ChapterResponse, mangaID int) (*database.Chapter, error) {
	chapter := &database.Chapter{
		MangaID:            mangaID,
		ChapterMangaDexID:  chapterResponse.Data.ID,
		ChapterNumber:      sql.NullString{String: chapterResponse.Data.Attributes.Chapter, Valid: true},
		Title:              sql.NullString{String: chapterResponse.Data.Attributes.Title, Valid: true},
		Volume:             sql.NullString{String: chapterResponse.Data.Attributes.Volume, Valid: true},
		TranslatedLanguage: chapterResponse.Data.Attributes.TranslatedLanguage,
		PageCount:          chapterResponse.Data.Attributes.Pages,
		FolderPath:         "", // Will be set later
		Downloaded:         false,
	}

	// Extract scanlation group
	for _, rel := range chapterResponse.Data.Relationships {
		if rel.Type == "scanlation_group" {
			if name, ok := rel.Attributes["name"].(string); ok {
				chapter.ScanlationGroup = sql.NullString{String: name, Valid: true}
				break
			}
		}
	}

	if !chapterResponse.Data.Attributes.PublishAt.IsZero() {
		chapter.PublishAt = sql.NullTime{Time: chapterResponse.Data.Attributes.PublishAt, Valid: true}
	}

	// Save to database
	id, err := database.SaveChapter(chapter)
	if err != nil {
		return nil, err
	}

	chapter.ID = id
	return chapter, nil
}

// SaveMangaToDBIfNotExists saves manga to database if it doesn't already exist
func SaveMangaToDBIfNotExists(mangaID string, mangaResponse *MangaResponse) error {
	// Check if manga already exists
	existingManga, err := database.GetMangaByMangaDexID(mangaID)
	if err != nil {
		return fmt.Errorf("failed to check manga existence: %w", err)
	}

	// If already exists, skip
	if existingManga != nil {
		utils.Logger.Info("Manga already exists in database", zap.String("mangaId", mangaID))
		return nil
	}

	// Save new manga
	_, err = saveMangaToDatabase(mangaResponse)
	if err != nil {
		return fmt.Errorf("failed to save manga: %w", err)
	}

	utils.Logger.Info("New manga saved to database", zap.String("mangaId", mangaID))
	return nil
}

// SaveMangaToDB persists the supplied MangaDex response and returns its local
// database record. It is used when a user deliberately saves a title.
func SaveMangaToDB(mangaResponse *MangaResponse) (*database.Manga, error) {
	return saveMangaToDatabase(mangaResponse)
}

// ChapterPagesResult holds the result of downloading chapter pages
type ChapterPagesResult struct {
	MangaID    string
	ChapterID  string
	LocalPaths []string
}

// PageDownloadEvent is emitted for each page as it becomes available.
type PageDownloadEvent struct {
	PageNumber int
	TotalPages int
	LocalPath  string
	PublicURL  string
	Downloaded bool
	Error      string
	MangaID    string
	ChapterID  string
}

// ProgressiveDownloadResult holds aggregate info for a progressive download.
type ProgressiveDownloadResult struct {
	MangaID      string
	ChapterID    string
	TotalPages   int
	SuccessCount int
	StartPage    int
}

// DownloadAndGetChapterPages downloads chapter images and returns local paths
func DownloadAndGetChapterPages(chapterID string, quality string) (*ChapterPagesResult, error) {
	utils.Logger.Info("Downloading chapter pages on-demand", zap.String("chapterId", chapterID))

	// Get chapter pages from MangaDex API
	pagesResponse, err := GetChapterPages(chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter pages: %w", err)
	}

	// Get chapter details to find manga ID
	chapterResponse, err := GetChapterByID(chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter details: %w", err)
	}

	// Extract manga ID from chapter relationships
	var mangaID string
	for _, rel := range chapterResponse.Data.Relationships {
		if rel.Type == "manga" {
			mangaID = rel.ID
			break
		}
	}

	if mangaID == "" {
		return nil, fmt.Errorf("manga ID not found in chapter relationships")
	}

	// Create folder structure: /images/manga/{mangaId}/chapters/{chapterId}/
	chapterFolder := filepath.Join(ImagesBaseDir, MangaDir, mangaID, "chapters", chapterID)
	if err := os.MkdirAll(chapterFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create chapter folder: %w", err)
	}

	// Get page URLs
	pageURLs := GetPageURLs(pagesResponse, quality)
	localPaths := make([]string, 0, len(pageURLs))

	// Download each page
	for i, pageURL := range pageURLs {
		pageNumber := i + 1
		fileName := fmt.Sprintf("%d.jpg", pageNumber)
		localPath := filepath.Join(chapterFolder, fileName)

		// Check if file already exists
		if _, err := os.Stat(localPath); err == nil {
			// File exists, use it
			utils.Logger.Info("Page already downloaded", zap.Int("page", pageNumber))
			localPaths = append(localPaths, localPath)
			continue
		}

		// Download the image
		utils.Logger.Info("Downloading page", zap.Int("page", pageNumber), zap.Int("total", len(pageURLs)))
		if err := downloadImage(pageURL, localPath); err != nil {
			utils.Logger.Error("Failed to download page",
				zap.Error(err),
				zap.Int("page", pageNumber))
			// Continue with next page even if one fails
			continue
		}

		localPaths = append(localPaths, localPath)

		// Rate limiting - be nice to MangaDex servers
		if i < len(pageURLs)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	utils.Logger.Info("Chapter pages ready",
		zap.String("chapterId", chapterID),
		zap.Int("pages", len(localPaths)))

	return &ChapterPagesResult{
		MangaID:    mangaID,
		ChapterID:  chapterID,
		LocalPaths: localPaths,
	}, nil
}

// DownloadChapterPagesProgressively downloads chapter pages and emits each page as soon as it is ready.
func DownloadChapterPagesProgressively(chapterID string, quality string, startPage int, onStart func(totalPages, startPage int) error, onPageReady func(event PageDownloadEvent) error) (*ProgressiveDownloadResult, error) {
	utils.Logger.Info("Progressive chapter download started", zap.String("chapterId", chapterID))

	pagesResponse, err := GetChapterPages(chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter pages: %w", err)
	}

	chapterResponse, err := GetChapterByID(chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter details: %w", err)
	}

	var mangaID string
	for _, rel := range chapterResponse.Data.Relationships {
		if rel.Type == "manga" {
			mangaID = rel.ID
			break
		}
	}

	if mangaID == "" {
		return nil, fmt.Errorf("manga ID not found in chapter relationships")
	}

	chapterFolder := filepath.Join(ImagesBaseDir, MangaDir, mangaID, "chapters", chapterID)
	if err := os.MkdirAll(chapterFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create chapter folder: %w", err)
	}

	pageURLs := GetPageURLs(pagesResponse, quality)
	pageFileNames := GetPageFileNames(pagesResponse, quality)
	if len(pageURLs) == 0 {
		return nil, fmt.Errorf("chapter has no pages")
	}

	if startPage < 1 {
		startPage = 1
	}
	if startPage > len(pageURLs) {
		startPage = len(pageURLs)
	}

	if onStart != nil {
		if err := onStart(len(pageURLs), startPage); err != nil {
			return nil, fmt.Errorf("stream start callback failed: %w", err)
		}
	}

	successCount := 0
	pageIndexes := make([]int, 0, len(pageURLs))
	for i := startPage - 1; i < len(pageURLs); i++ {
		pageIndexes = append(pageIndexes, i)
	}
	forwardPageCount := len(pageIndexes)
	for i := startPage - 2; i >= 0; i-- {
		pageIndexes = append(pageIndexes, i)
	}

	for streamIndex, i := range pageIndexes {
		pageURL := pageURLs[i]
		pageNumber := i + 1
		// Extract file extension from original filename
		originalFileName := pageFileNames[i]
		fileExt := filepath.Ext(originalFileName)
		if fileExt == "" {
			fileExt = ".jpg" // Fallback to jpg if no extension
		}
		fileName := fmt.Sprintf("%d%s", pageNumber, fileExt)
		localPath := filepath.Join(chapterFolder, fileName)
		publicURL := path.Join("/images", MangaDir, mangaID, "chapters", chapterID, fileName)

		downloadedNow := false
		eventErr := ""

		if fileInfo, err := os.Stat(localPath); err != nil || fileInfo.Size() == 0 {
			utils.Logger.Info("Downloading page", zap.Int("page", pageNumber), zap.Int("total", len(pageURLs)))
			if err := downloadImageWithRetry(pageURL, localPath, 3); err != nil {
				utils.Logger.Error("Failed to download page after retries", zap.Error(err), zap.Int("page", pageNumber))
				eventErr = err.Error()
			} else {
				downloadedNow = true
				successCount++
			}
		} else {
			successCount++
		}

		if onPageReady != nil {
			err := onPageReady(PageDownloadEvent{
				PageNumber: pageNumber,
				TotalPages: len(pageURLs),
				LocalPath:  localPath,
				PublicURL:  publicURL,
				Downloaded: downloadedNow,
				Error:      eventErr,
				MangaID:    mangaID,
				ChapterID:  chapterID,
			})
			if err != nil {
				return nil, fmt.Errorf("page callback failed: %w", err)
			}
		}

		if streamIndex < len(pageIndexes)-1 {
			delay := 200 * time.Millisecond
			if streamIndex >= forwardPageCount {
				delay = 400 * time.Millisecond
			}
			time.Sleep(delay)
		}
	}

	return &ProgressiveDownloadResult{
		MangaID:      mangaID,
		ChapterID:    chapterID,
		TotalPages:   len(pageURLs),
		SuccessCount: successCount,
		StartPage:    startPage,
	}, nil
}

// sanitizeFileName removes invalid characters from filenames
func sanitizeFileName(name string) string {
	// Replace invalid characters
	replacer := []string{"/", "_", "\\", "_", ":", "-", "*", "", "?", "", "\"", "", "<", "", ">", "", "|", ""}
	for i := 0; i < len(replacer); i += 2 {
		name = replaceAll(name, replacer[i], replacer[i+1])
	}

	// Limit length
	if len(name) > 100 {
		name = name[:100]
	}

	return name
}

func replaceAll(s, old, new string) string {
	result := ""
	for _, char := range s {
		if string(char) == old {
			result += new
		} else {
			result += string(char)
		}
	}
	return result
}
