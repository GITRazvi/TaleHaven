package mangadex_api

import (
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var preferredTitleLocales = []string{"en", "ja-ro", "ja", "ko", "zh"}

// MatchMetadata carries optional AniList fields that can break an otherwise
// unresolved tie between safe MangaDex title matches.
type MatchMetadata struct {
	Titles       []string
	Year         int
	CreatorNames []string
}

// GetPreferredMangaTitle returns the best display title across MangaDex's
// primary and alternative titles, preferring an English name when available.
func GetPreferredMangaTitle(manga MangaData) string {
	candidates := mangaTitleCandidates(manga)
	if len(candidates) == 0 {
		return "Unknown Title"
	}
	return candidates[0]
}

// MangaAlternativeTitles returns every distinct title other than the displayed
// title. It includes MangaDex's localized primary title and every altTitles value.
func MangaAlternativeTitles(manga MangaData, displayedTitle string) []string {
	result := make([]string, 0)
	displayedKey := normalizeTitle(displayedTitle)
	for _, title := range mangaTitleCandidates(manga) {
		if normalizeTitle(title) == displayedKey {
			continue
		}
		result = append(result, title)
	}
	return result
}

// MatchMangaTitle compares a query with primary and alternative MangaDex titles.
// A lower rank is a better match: exact (0), prefix (1), partial (2), no match (3).
func MatchMangaTitle(manga MangaData, query string) (matchedTitle, matchType string, rank int) {
	queryKey := normalizeTitle(query)
	if queryKey == "" {
		return "", "", 3
	}

	bestRank := 3
	bestTitle := ""
	for _, title := range mangaTitleCandidates(manga) {
		titleKey := normalizeTitle(title)
		candidateRank := 3
		candidateType := ""
		switch {
		case titleKey == queryKey:
			candidateRank, candidateType = 0, "exact"
		case strings.HasPrefix(titleKey, queryKey):
			candidateRank, candidateType = 1, "prefix"
		case strings.Contains(titleKey, queryKey):
			candidateRank, candidateType = 2, "partial"
		}

		if candidateRank < bestRank {
			bestRank, bestTitle, matchType = candidateRank, title, candidateType
		}
	}
	return bestTitle, matchType, bestRank
}

// FindConfidentMangaMatch returns a single MangaDex result only when an
// AniList title variant maps safely to it. It intentionally favors no match
// over returning chapters for a similarly named sequel or spin-off.
func FindConfidentMangaMatch(manga []MangaData, titles []string) (MangaData, bool) {
	return FindConfidentMangaMatchWithMetadata(manga, MatchMetadata{Titles: titles})
}

// FindConfidentMangaMatchWithMetadata returns a single MangaDex result only
// when an AniList title variant maps safely to it. Publication year and
// author/artist metadata are used only to choose between otherwise tied safe
// title matches; they never turn a weak title match into a valid result.
func FindConfidentMangaMatchWithMetadata(manga []MangaData, metadata MatchMetadata) (MangaData, bool) {
	type match struct {
		manga         MangaData
		rank          int
		tieBreakScore int
	}
	best := make([]match, 0, 1)
	bestRank := 3
	for _, candidate := range manga {
		candidateRank := 3
		for _, title := range metadata.Titles {
			for _, candidateTitle := range mangaTitleCandidates(candidate) {
				rank := confidentTitleRank(title, candidateTitle)
				if rank < candidateRank {
					candidateRank = rank
				}
			}
		}
		if candidateRank > 1 {
			continue
		}
		if candidateRank < bestRank {
			bestRank = candidateRank
			best = []match{{manga: candidate, rank: candidateRank}}
		} else if candidateRank == bestRank {
			best = append(best, match{manga: candidate, rank: candidateRank})
		}
	}
	if len(best) == 1 {
		return best[0].manga, true
	}
	if len(best) == 0 {
		return MangaData{}, false
	}

	bestScore := 0
	bestIndex := -1
	for index := range best {
		best[index].tieBreakScore = mangaMatchTieBreakScore(best[index].manga, metadata)
		if best[index].tieBreakScore > bestScore {
			bestScore = best[index].tieBreakScore
			bestIndex = index
		} else if best[index].tieBreakScore == bestScore {
			bestIndex = -1
		}
	}
	if bestIndex == -1 || bestScore == 0 {
		return MangaData{}, false
	}
	return best[bestIndex].manga, true
}

func mangaMatchTieBreakScore(candidate MangaData, metadata MatchMetadata) int {
	score := 0
	if metadata.Year > 0 && candidate.Attributes.Year > 0 {
		yearDifference := metadata.Year - candidate.Attributes.Year
		if yearDifference < 0 {
			yearDifference = -yearDifference
		}
		switch yearDifference {
		case 0:
			score += 2
		case 1:
			score++
		}
	}
	if hasMatchingCreator(candidate, metadata.CreatorNames) {
		score += 3
	}
	return score
}

func hasMatchingCreator(candidate MangaData, creators []string) bool {
	creatorKeys := make(map[string]struct{}, len(creators))
	for _, creator := range creators {
		if key := normalizeTitle(creator); key != "" {
			creatorKeys[key] = struct{}{}
		}
	}
	if len(creatorKeys) == 0 {
		return false
	}
	for _, relationship := range candidate.Relationships {
		if relationship.Type != "author" && relationship.Type != "artist" {
			continue
		}
		name, _ := relationship.Attributes["name"].(string)
		if _, found := creatorKeys[normalizeTitle(name)]; found && normalizeTitle(name) != "" {
			return true
		}
	}
	return false
}

// confidentTitleRank returns 0 for an exact normalized title and 1 when the
// query and candidate contain the same words in any order, with at most one
// descriptive word added by the candidate. Word counts are preserved so a
// repeated word cannot accidentally make a shorter title look equivalent.
func confidentTitleRank(query, candidate string) int {
	queryKey, candidateKey := normalizeTitle(query), normalizeTitle(candidate)
	if queryKey == "" || candidateKey == "" {
		return 3
	}
	if queryKey == candidateKey {
		return 0
	}
	queryWords, candidateWords := titleWords(query), titleWords(candidate)
	if len(queryWords) < 2 ||
		len(candidateWords) < len(queryWords) ||
		len(candidateWords) > len(queryWords)+1 {
		return 3
	}
	candidateCounts := make(map[string]int, len(candidateWords))
	for _, word := range candidateWords {
		candidateCounts[word]++
	}
	for _, word := range queryWords {
		if candidateCounts[word] == 0 {
			return 3
		}
		candidateCounts[word]--
	}
	return 1
}

func titleWords(value string) []string {
	words := strings.FieldsFunc(strings.ToLower(norm.NFD.String(value)), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	result := make([]string, 0, len(words))
	for _, word := range words {
		word = normalizeTitle(word)
		if word != "" {
			result = append(result, word)
		}
	}
	return result
}

func mangaTitleCandidates(manga MangaData) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	add := func(title string) {
		title = strings.TrimSpace(title)
		key := normalizeTitle(title)
		if key == "" {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		result = append(result, title)
	}

	// Inspect each preferred locale in all title maps first. This makes the
	// result stable and lets an English alternative take precedence over a
	// romanized primary title.
	for _, locale := range preferredTitleLocales {
		add(manga.Attributes.Title[locale])
		for _, altTitles := range manga.Attributes.AltTitles {
			add(altTitles[locale])
		}
	}

	addRemaining := func(titles map[string]string) {
		locales := make([]string, 0, len(titles))
		for locale := range titles {
			if !isPreferredTitleLocale(locale) {
				locales = append(locales, locale)
			}
		}
		sort.Strings(locales)
		for _, locale := range locales {
			add(titles[locale])
		}
	}
	addRemaining(manga.Attributes.Title)
	for _, altTitles := range manga.Attributes.AltTitles {
		addRemaining(altTitles)
	}

	return result
}

func isPreferredTitleLocale(locale string) bool {
	for _, preferred := range preferredTitleLocales {
		if locale == preferred {
			return true
		}
	}
	return false
}

func normalizeTitle(value string) string {
	var builder strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(value)) {
		if unicode.IsMark(r) {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
