package mangadex_api

import (
	"sort"
	"strconv"
	"strings"
)

const noScanlationGroupID = "official"

// ScanlationGroupOption is a stable group filter option for one language.
type ScanlationGroupOption struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ChapterCount int    `json:"chapterCount"`
}

// SelectChapters returns a reader-facing list with one release per chapter
// key. In all-group mode it favors the group with the widest coverage for each
// language, then fills missing chapters from other groups.
func SelectChapters(chapters []ChapterData, language, groupID string) ([]ChapterData, []string, []ScanlationGroupOption) {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		language = "all"
	}
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		groupID = "all"
	}

	availableLanguages := chapterLanguages(chapters)
	filtered := make([]ChapterData, 0, len(chapters))
	for _, chapter := range chapters {
		if language == "all" || chapter.Attributes.TranslatedLanguage == language {
			filtered = append(filtered, chapter)
		}
	}

	options := []ScanlationGroupOption{}
	if language != "all" {
		options = scanlationGroupOptions(filtered)
	}

	if groupID != "all" {
		onlyGroup := make([]ChapterData, 0, len(filtered))
		for _, chapter := range filtered {
			if chapterScanlationGroupID(chapter) == groupID {
				onlyGroup = append(onlyGroup, chapter)
			}
		}
		return selectOnePerChapter(onlyGroup, nil), availableLanguages, options
	}

	return selectAllGroups(filtered), availableLanguages, options
}

func chapterLanguages(chapters []ChapterData) []string {
	seen := make(map[string]struct{})
	for _, chapter := range chapters {
		if language := strings.TrimSpace(chapter.Attributes.TranslatedLanguage); language != "" {
			seen[language] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for language := range seen {
		result = append(result, language)
	}
	sort.Strings(result)
	return result
}

func scanlationGroupOptions(chapters []ChapterData) []ScanlationGroupOption {
	keysByGroup := make(map[string]map[string]struct{})
	names := make(map[string]string)
	for _, chapter := range chapters {
		groupID, groupName := chapterScanlationGroupID(chapter), chapterScanlationGroupName(chapter)
		if _, exists := keysByGroup[groupID]; !exists {
			keysByGroup[groupID] = make(map[string]struct{})
			names[groupID] = groupName
		}
		keysByGroup[groupID][chapterKey(chapter)] = struct{}{}
	}
	result := make([]ScanlationGroupOption, 0, len(keysByGroup))
	for groupID, keys := range keysByGroup {
		result = append(result, ScanlationGroupOption{ID: groupID, Name: names[groupID], ChapterCount: len(keys)})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ChapterCount != result[j].ChapterCount {
			return result[i].ChapterCount > result[j].ChapterCount
		}
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func selectAllGroups(chapters []ChapterData) []ChapterData {
	coverageKeysByLanguage := make(map[string]map[string]map[string]struct{})
	for _, chapter := range chapters {
		language := chapter.Attributes.TranslatedLanguage
		if coverageKeysByLanguage[language] == nil {
			coverageKeysByLanguage[language] = make(map[string]map[string]struct{})
		}
		groupID := chapterScanlationGroupID(chapter)
		if coverageKeysByLanguage[language][groupID] == nil {
			coverageKeysByLanguage[language][groupID] = make(map[string]struct{})
		}
		coverageKeysByLanguage[language][groupID][chapterKey(chapter)] = struct{}{}
	}
	preferredGroup := make(map[string]string)
	for language, coverageKeys := range coverageKeysByLanguage {
		bestID, bestCount := "", -1
		for groupID, keys := range coverageKeys {
			count := len(keys)
			if count > bestCount || (count == bestCount && groupID < bestID) {
				bestID, bestCount = groupID, count
			}
		}
		preferredGroup[language] = bestID
	}
	return selectOnePerChapter(chapters, preferredGroup)
}

func selectOnePerChapter(chapters []ChapterData, preferredGroup map[string]string) []ChapterData {
	selected := make(map[string]ChapterData)
	for _, chapter := range chapters {
		key := chapterKey(chapter)
		current, exists := selected[key]
		if !exists || chapterIsPreferred(chapter, current, preferredGroup) {
			selected[key] = chapter
		}
	}
	result := make([]ChapterData, 0, len(selected))
	for _, chapter := range selected {
		result = append(result, chapter)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Attributes.Chapter != right.Attributes.Chapter || left.Attributes.Volume != right.Attributes.Volume {
			return compareChapterOrder(left, right)
		}
		if left.Attributes.TranslatedLanguage != right.Attributes.TranslatedLanguage {
			return left.Attributes.TranslatedLanguage < right.Attributes.TranslatedLanguage
		}
		return left.ID < right.ID
	})
	return result
}

func chapterIsPreferred(candidate, current ChapterData, preferredGroup map[string]string) bool {
	if preferredGroup != nil {
		preferred := preferredGroup[candidate.Attributes.TranslatedLanguage]
		candidatePreferred := chapterScanlationGroupID(candidate) == preferred
		currentPreferred := chapterScanlationGroupID(current) == preferred
		if candidatePreferred != currentPreferred {
			return candidatePreferred
		}
	}
	if !candidate.Attributes.PublishAt.Equal(current.Attributes.PublishAt) {
		return candidate.Attributes.PublishAt.Before(current.Attributes.PublishAt)
	}
	return candidate.ID < current.ID
}

func chapterKey(chapter ChapterData) string {
	chapterNumber := strings.TrimSpace(chapter.Attributes.Chapter)
	if chapterNumber == "" {
		return chapter.Attributes.TranslatedLanguage + "|id:" + chapter.ID
	}
	return chapter.Attributes.TranslatedLanguage + "|" + strings.TrimSpace(chapter.Attributes.Volume) + "|" + chapterNumber
}

func chapterScanlationGroupID(chapter ChapterData) string {
	for _, relationship := range chapter.Relationships {
		if relationship.Type == "scanlation_group" && relationship.ID != "" {
			return relationship.ID
		}
	}
	return noScanlationGroupID
}

func chapterScanlationGroupName(chapter ChapterData) string {
	for _, relationship := range chapter.Relationships {
		if relationship.Type != "scanlation_group" {
			continue
		}
		if name, ok := relationship.Attributes["name"].(string); ok && strings.TrimSpace(name) != "" {
			return name
		}
	}
	return "Official / no group"
}

func compareChapterOrder(left, right ChapterData) bool {
	leftNumber, leftError := strconv.ParseFloat(left.Attributes.Chapter, 64)
	rightNumber, rightError := strconv.ParseFloat(right.Attributes.Chapter, 64)
	if leftError == nil && rightError == nil && leftNumber != rightNumber {
		return leftNumber < rightNumber
	}
	if left.Attributes.Chapter != right.Attributes.Chapter {
		return left.Attributes.Chapter < right.Attributes.Chapter
	}
	if left.Attributes.Volume != right.Attributes.Volume {
		return left.Attributes.Volume < right.Attributes.Volume
	}
	return left.ID < right.ID
}
