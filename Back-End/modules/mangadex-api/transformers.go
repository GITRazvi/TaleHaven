package mangadex_api

// TransformToSimpleManga converts MangaData to a simplified format
func TransformToSimpleManga(manga MangaData) SimpleManga {
	// Extract title (prefer an English primary or alternative title).
	title := GetPreferredMangaTitle(manga)

	// Extract description (prefer English)
	description := getPreferredDescription(manga.Attributes.Description)

	// Extract author and artist from relationships
	author := ""
	artist := ""
	coverID := ""

	for _, rel := range manga.Relationships {
		switch rel.Type {
		case "author":
			if name, ok := rel.Attributes["name"].(string); ok {
				author = name
			}
		case "artist":
			if name, ok := rel.Attributes["name"].(string); ok {
				artist = name
			}
		case "cover_art":
			coverID = rel.ID
		}
	}

	// Extract tags
	tags := make([]string, 0)
	for _, tag := range manga.Attributes.Tags {
		if tagName, ok := tag.Attributes.Name["en"]; ok {
			tags = append(tags, tagName)
		}
	}

	// Build cover URL if available
	coverUrl := ""
	if coverID != "" {
		// Cover URL will need to be constructed with the cover filename
		// This requires a separate API call or should be done at the controller level
		coverUrl = coverID // Store cover ID, actual URL built elsewhere
	}

	return SimpleManga{
		ID:                manga.ID,
		Title:             title,
		AlternativeTitles: MangaAlternativeTitles(manga, title),
		Description:       description,
		CoverUrl:          coverUrl,
		Author:            author,
		Artist:            artist,
		Status:            manga.Attributes.Status,
		Year:              manga.Attributes.Year,
		Tags:              tags,
		Rating:            manga.Attributes.ContentRating,
		LastChapter:       manga.Attributes.LastChapter,
		Languages:         manga.Attributes.AvailableTranslatedLanguages,
	}
}

// TransformToSimpleChapter converts ChapterData to a simplified format
func TransformToSimpleChapter(chapter ChapterData) SimpleChapter {
	mangaID := ""
	scanlationGroup := ""
	scanlationGroupID := ""

	for _, rel := range chapter.Relationships {
		switch rel.Type {
		case "manga":
			mangaID = rel.ID
		case "scanlation_group":
			scanlationGroupID = rel.ID
			if name, ok := rel.Attributes["name"].(string); ok {
				scanlationGroup = name
			}
		}
	}

	return SimpleChapter{
		ID:                 chapter.ID,
		MangaID:            mangaID,
		Chapter:            chapter.Attributes.Chapter,
		Title:              chapter.Attributes.Title,
		Volume:             chapter.Attributes.Volume,
		TranslatedLanguage: chapter.Attributes.TranslatedLanguage,
		Pages:              chapter.Attributes.Pages,
		PublishAt:          chapter.Attributes.PublishAt,
		ScanlationGroup:    scanlationGroup,
		ScanlationGroupID:  scanlationGroupID,
	}
}

// GetPreferredTitle returns the English title or falls back to first available
func getPreferredTitle(titles map[string]string) string {
	if title, ok := titles["en"]; ok {
		return title
	}
	// Try common alternatives
	for _, lang := range []string{"ja-ro", "ja", "ko", "zh"} {
		if title, ok := titles[lang]; ok {
			return title
		}
	}
	// Return first available
	for _, title := range titles {
		return title
	}
	return "Unknown Title"
}

// GetPreferredDescription returns the English description or falls back to first available
func getPreferredDescription(descriptions map[string]string) string {
	if desc, ok := descriptions["en"]; ok {
		return desc
	}
	// Return first available
	for _, desc := range descriptions {
		return desc
	}
	return ""
}

// GetCoverFileName extracts the cover filename from relationships
func GetCoverFileName(relationships []Relationship) string {
	for _, rel := range relationships {
		if rel.Type == "cover_art" {
			if fileName, ok := rel.Attributes["fileName"].(string); ok {
				return fileName
			}
		}
	}
	return ""
}

// BuildFullCoverURL builds the complete cover URL with manga ID and filename
// MangaDex URL format: covers/{mangaId}/{fileName}.{quality}.jpg
// where fileName already includes extension (e.g., abc123.jpg becomes abc123.jpg.512.jpg)
func BuildFullCoverURL(mangaID string, relationships []Relationship, quality string) string {
	fileName := GetCoverFileName(relationships)
	if fileName == "" {
		return ""
	}

	// Do NOT strip extension - MangaDex expects fileName with extension
	// Example: abc123.jpg -> abc123.jpg.512.jpg
	return BuildCoverURL(mangaID, fileName, quality)
}

// GetPageURLs constructs full URLs for chapter pages
func GetPageURLs(pagesResponse *ChapterPagesResponse, quality string) []string {
	baseURL := pagesResponse.BaseUrl
	hash := pagesResponse.Chapter.Hash

	var pageFiles []string
	if quality == "saver" {
		pageFiles = pagesResponse.Chapter.DataSaver
	} else {
		pageFiles = pagesResponse.Chapter.Data
	}

	urls := make([]string, len(pageFiles))
	for i, filename := range pageFiles {
		if quality == "saver" {
			urls[i] = baseURL + "/data-saver/" + hash + "/" + filename
		} else {
			urls[i] = baseURL + "/data/" + hash + "/" + filename
		}
	}

	return urls
}

// GetPageFileNames returns the original filenames (with extensions) for chapter pages
func GetPageFileNames(pagesResponse *ChapterPagesResponse, quality string) []string {
	var pageFiles []string
	if quality == "saver" {
		pageFiles = pagesResponse.Chapter.DataSaver
	} else {
		pageFiles = pagesResponse.Chapter.Data
	}

	return pageFiles
}

// FilterChaptersByLanguage filters chapters by language
func FilterChaptersByLanguage(chapters []ChapterData, language string) []ChapterData {
	filtered := make([]ChapterData, 0)
	for _, chapter := range chapters {
		if chapter.Attributes.TranslatedLanguage == language {
			filtered = append(filtered, chapter)
		}
	}
	return filtered
}

// GroupChaptersByVolume groups chapters by volume number
func GroupChaptersByVolume(chapters []ChapterData) map[string][]ChapterData {
	grouped := make(map[string][]ChapterData)
	for _, chapter := range chapters {
		volume := chapter.Attributes.Volume
		if volume == "" {
			volume = "none"
		}
		grouped[volume] = append(grouped[volume], chapter)
	}
	return grouped
}
