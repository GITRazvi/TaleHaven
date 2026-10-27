package anilist_api

import (
	"fmt"
	"strings"
)

// Pagination describes AniList's PageInfo object for a paged manga search.
type Pagination struct {
	CurrentPage int  `json:"currentPage"`
	HasNextPage bool `json:"hasNextPage"`
	PerPage     int  `json:"perPage"`
	Total       int  `json:"total"`
}

// MangaSearchResult is a public, paged AniList manga search result.
type MangaSearchResult struct {
	PageInfo Pagination      `json:"pageInfo"`
	Manga    []MangaMetadata `json:"manga"`
}

// MangaSearchFilters are the public AniList catalog constraints accepted by
// TaleHaven. Empty optional fields are omitted from the upstream GraphQL query.
type MangaSearchFilters struct {
	Title  string
	Genre  string
	Tag    string
	Year   int
	Format string
	Status string
	Sort   string
}

var supportedMangaFormats = map[string]struct{}{
	"MANGA": {}, "NOVEL": {}, "ONE_SHOT": {},
}

var supportedMangaStatuses = map[string]struct{}{
	"FINISHED": {}, "RELEASING": {}, "NOT_YET_RELEASED": {}, "CANCELLED": {}, "HIATUS": {},
}

var supportedMangaSorts = map[string]struct{}{
	"SEARCH_MATCH": {}, "POPULARITY_DESC": {}, "SCORE_DESC": {}, "START_DATE_DESC": {},
}

// Normalize validates API-facing search filters and supplies a sensible sort
// when the caller leaves it empty.
func (filters MangaSearchFilters) Normalize() (MangaSearchFilters, error) {
	filters.Title = strings.TrimSpace(filters.Title)
	filters.Genre = strings.TrimSpace(filters.Genre)
	filters.Tag = strings.TrimSpace(filters.Tag)
	filters.Format = strings.ToUpper(strings.TrimSpace(filters.Format))
	filters.Status = strings.ToUpper(strings.TrimSpace(filters.Status))
	filters.Sort = strings.ToUpper(strings.TrimSpace(filters.Sort))

	if len([]rune(filters.Title)) > 200 {
		return MangaSearchFilters{}, fmt.Errorf("AniList title must be at most 200 characters")
	}
	if len([]rune(filters.Genre)) > 100 {
		return MangaSearchFilters{}, fmt.Errorf("AniList genre must be at most 100 characters")
	}
	if len([]rune(filters.Tag)) > 100 {
		return MangaSearchFilters{}, fmt.Errorf("AniList tag must be at most 100 characters")
	}
	if filters.Year != 0 && (filters.Year < 1900 || filters.Year > 2100) {
		return MangaSearchFilters{}, fmt.Errorf("AniList release year must be between 1900 and 2100")
	}
	if filters.Format != "" {
		if _, supported := supportedMangaFormats[filters.Format]; !supported {
			return MangaSearchFilters{}, fmt.Errorf("unsupported AniList manga format %q", filters.Format)
		}
	}
	if filters.Status != "" {
		if _, supported := supportedMangaStatuses[filters.Status]; !supported {
			return MangaSearchFilters{}, fmt.Errorf("unsupported AniList manga status %q", filters.Status)
		}
	}
	if filters.Sort != "" {
		if _, supported := supportedMangaSorts[filters.Sort]; !supported {
			return MangaSearchFilters{}, fmt.Errorf("unsupported AniList manga sort %q", filters.Sort)
		}
	}
	if filters.Title == "" && filters.Genre == "" && filters.Tag == "" && filters.Year == 0 && filters.Format == "" && filters.Status == "" && filters.Sort == "" {
		return MangaSearchFilters{}, fmt.Errorf("at least one AniList search filter is required")
	}
	if filters.Sort == "" {
		if filters.Title != "" {
			filters.Sort = "SEARCH_MATCH"
		} else {
			filters.Sort = "POPULARITY_DESC"
		}
	}
	return filters, nil
}

// DiscoverySection names a curated public AniList manga list shown on the
// TaleHaven home page.
type DiscoverySection string

const (
	DiscoveryTrending DiscoverySection = "trending"
	DiscoveryPopular  DiscoverySection = "popular"
	DiscoveryTopRated DiscoverySection = "top-rated"
	DiscoveryRecent   DiscoverySection = "recent"
)

// Sort returns the fixed AniList ordering associated with a discovery section.
func (section DiscoverySection) Sort() ([]string, bool) {
	switch section {
	case DiscoveryTrending:
		return []string{"TRENDING_DESC"}, true
	case DiscoveryPopular:
		return []string{"POPULARITY_DESC"}, true
	case DiscoveryTopRated:
		return []string{"SCORE_DESC"}, true
	case DiscoveryRecent:
		return []string{"START_DATE_DESC"}, true
	default:
		return nil, false
	}
}

// MangaMetadata contains only public AniList fields needed for TaleHaven's
// future catalog and details views.
type MangaMetadata struct {
	ID              int                 `json:"id"`
	IDMal           int                 `json:"idMal"`
	Title           Title               `json:"title"`
	Synonyms        []string            `json:"synonyms"`
	Description     string              `json:"description"`
	CoverImage      CoverImage          `json:"coverImage"`
	BannerImage     string              `json:"bannerImage"`
	Format          string              `json:"format"`
	Status          string              `json:"status"`
	Chapters        int                 `json:"chapters"`
	Volumes         int                 `json:"volumes"`
	Genres          []string            `json:"genres"`
	AverageScore    int                 `json:"averageScore"`
	Popularity      int                 `json:"popularity"`
	SiteURL         string              `json:"siteUrl"`
	IsAdult         bool                `json:"isAdult"`
	StartDate       FuzzyDate           `json:"startDate"`
	EndDate         FuzzyDate           `json:"endDate"`
	CountryOfOrigin string              `json:"countryOfOrigin"`
	Source          string              `json:"source"`
	Favourites      int                 `json:"favourites"`
	Rankings        []Ranking           `json:"rankings"`
	Tags            []Tag               `json:"tags"`
	Staff           StaffConnection     `json:"staff"`
	Characters      CharacterConnection `json:"characters"`
	ExternalLinks   []ExternalLink      `json:"externalLinks"`
	Relations       RelationConnection  `json:"relations"`
}

type Title struct {
	Romaji        string `json:"romaji"`
	English       string `json:"english"`
	Native        string `json:"native"`
	UserPreferred string `json:"userPreferred"`
}

type CoverImage struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
	Color  string `json:"color"`
}

type FuzzyDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

// Ranking is a public AniList rank for the manga, such as all-time rating or
// popularity. A title may have more than one ranking context.
type Ranking struct {
	Rank    int    `json:"rank"`
	Type    string `json:"type"`
	AllTime bool   `json:"allTime"`
	Context string `json:"context"`
}

type Tag struct {
	Name           string `json:"name"`
	Rank           int    `json:"rank"`
	IsMediaSpoiler bool   `json:"isMediaSpoiler"`
}

type StaffConnection struct {
	Edges []StaffEdge `json:"edges"`
}

type StaffEdge struct {
	Role string      `json:"role"`
	Node StaffMember `json:"node"`
}

type StaffMember struct {
	ID   int       `json:"id"`
	Name StaffName `json:"name"`
}

type StaffName struct {
	Full          string `json:"full"`
	Native        string `json:"native"`
	UserPreferred string `json:"userPreferred"`
}

// CharacterConnection contains the first page of story characters returned
// with a manga-details query.
type CharacterConnection struct {
	PageInfo Pagination      `json:"pageInfo"`
	Edges    []CharacterEdge `json:"edges"`
}

type CharacterEdge struct {
	Role string    `json:"role"`
	Node Character `json:"node"`
}

type Character struct {
	ID          int             `json:"id"`
	Name        CharacterName   `json:"name"`
	Image       CharacterImage  `json:"image"`
	Description string          `json:"description"`
	Gender      string          `json:"gender"`
	DateOfBirth FuzzyDate       `json:"dateOfBirth"`
	Age         string          `json:"age"`
	BloodType   string          `json:"bloodType"`
	SiteURL     string          `json:"siteUrl"`
	Favourites  int             `json:"favourites"`
	Media       MediaConnection `json:"media"`
}

type CharacterName struct {
	Full          string `json:"full"`
	Native        string `json:"native"`
	UserPreferred string `json:"userPreferred"`
}

type CharacterImage struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
}

// MediaConnection represents the works a character appears in.
type MediaConnection struct {
	PageInfo Pagination  `json:"pageInfo"`
	Edges    []MediaEdge `json:"edges"`
}

type MediaEdge struct {
	CharacterRole string         `json:"characterRole"`
	Node          CharacterMedia `json:"node"`
}

type CharacterMedia struct {
	ID         int        `json:"id"`
	Type       string     `json:"type"`
	Title      Title      `json:"title"`
	CoverImage CoverImage `json:"coverImage"`
	Format     string     `json:"format"`
	Status     string     `json:"status"`
	SiteURL    string     `json:"siteUrl"`
}

type ExternalLink struct {
	ID       int    `json:"id"`
	Site     string `json:"site"`
	URL      string `json:"url"`
	Type     string `json:"type"`
	Language string `json:"language"`
}

type RelationConnection struct {
	Edges []RelationEdge `json:"edges"`
}

type RelationEdge struct {
	RelationType string       `json:"relationType"`
	Node         RelatedManga `json:"node"`
}

type RelatedManga struct {
	ID         int        `json:"id"`
	Type       string     `json:"type"`
	Title      Title      `json:"title"`
	CoverImage CoverImage `json:"coverImage"`
	Format     string     `json:"format"`
	Status     string     `json:"status"`
	SiteURL    string     `json:"siteUrl"`
}
