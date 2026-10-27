package mangadex_api

import "time"

// Response wrapper types
type MangaListResponse struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     []MangaData `json:"data"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Total    int         `json:"total"`
}

type MangaResponse struct {
	Result   string    `json:"result"`
	Response string    `json:"response"`
	Data     MangaData `json:"data"`
}

type ChapterListResponse struct {
	Result   string        `json:"result"`
	Response string        `json:"response"`
	Data     []ChapterData `json:"data"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
	Total    int           `json:"total"`
}

type ChapterResponse struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     ChapterData `json:"data"`
}

type CoverListResponse struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     []CoverData `json:"data"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Total    int         `json:"total"`
}

// Core entity types
type MangaData struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    MangaAttributes `json:"attributes"`
	Relationships []Relationship  `json:"relationships"`
}

type MangaAttributes struct {
	Title                          map[string]string   `json:"title"`
	AltTitles                      []map[string]string `json:"altTitles"`
	Description                    map[string]string   `json:"description"`
	IsLocked                       bool                `json:"isLocked"`
	Links                          map[string]string   `json:"links"`
	OriginalLanguage               string              `json:"originalLanguage"`
	LastVolume                     string              `json:"lastVolume"`
	LastChapter                    string              `json:"lastChapter"`
	PublicationDemographic         string              `json:"publicationDemographic"`
	Status                         string              `json:"status"`
	Year                           int                 `json:"year"`
	ContentRating                  string              `json:"contentRating"`
	Tags                           []Tag               `json:"tags"`
	State                          string              `json:"state"`
	ChapterNumbersResetOnNewVolume bool                `json:"chapterNumbersResetOnNewVolume"`
	CreatedAt                      time.Time           `json:"createdAt"`
	UpdatedAt                      time.Time           `json:"updatedAt"`
	Version                        int                 `json:"version"`
	AvailableTranslatedLanguages   []string            `json:"availableTranslatedLanguages"`
	LatestUploadedChapter          string              `json:"latestUploadedChapter"`
}

type ChapterData struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    ChapterAttributes `json:"attributes"`
	Relationships []Relationship    `json:"relationships"`
}

type ChapterAttributes struct {
	Volume             string    `json:"volume"`
	Chapter            string    `json:"chapter"`
	Title              string    `json:"title"`
	TranslatedLanguage string    `json:"translatedLanguage"`
	ExternalUrl        string    `json:"externalUrl"`
	PublishAt          time.Time `json:"publishAt"`
	ReadableAt         time.Time `json:"readableAt"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	Pages              int       `json:"pages"`
	Version            int       `json:"version"`
}

type CoverData struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    CoverAttributes `json:"attributes"`
	Relationships []Relationship  `json:"relationships"`
}

type CoverAttributes struct {
	Description string    `json:"description"`
	Volume      string    `json:"volume"`
	FileName    string    `json:"fileName"`
	Locale      string    `json:"locale"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Version     int       `json:"version"`
}

type Relationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Related    string                 `json:"related,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type Tag struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    TagAttributes  `json:"attributes"`
	Relationships []Relationship `json:"relationships"`
}

type TagAttributes struct {
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	Group       string            `json:"group"`
	Version     int               `json:"version"`
}

type AuthorData struct {
	ID            string           `json:"id"`
	Type          string           `json:"type"`
	Attributes    AuthorAttributes `json:"attributes"`
	Relationships []Relationship   `json:"relationships"`
}

type AuthorAttributes struct {
	Name      string            `json:"name"`
	ImageUrl  string            `json:"imageUrl"`
	Biography map[string]string `json:"biography"`
	Twitter   string            `json:"twitter"`
	Pixiv     string            `json:"pixiv"`
	MelonBook string            `json:"melonBook"`
	FanBox    string            `json:"fanBox"`
	Booth     string            `json:"booth"`
	NicoVideo string            `json:"nicoVideo"`
	Skeb      string            `json:"skeb"`
	Fantia    string            `json:"fantia"`
	Tumblr    string            `json:"tumblr"`
	Youtube   string            `json:"youtube"`
	Weibo     string            `json:"weibo"`
	Naver     string            `json:"naver"`
	Website   string            `json:"website"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Version   int               `json:"version"`
}

// Error response
type ErrorResponse struct {
	Result string  `json:"result"`
	Errors []Error `json:"errors"`
}

type Error struct {
	ID     string `json:"id"`
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// Aggregated response for chapters with page info
type ChapterPagesResponse struct {
	Result  string `json:"result"`
	BaseUrl string `json:"baseUrl"`
	Chapter struct {
		Hash      string   `json:"hash"`
		Data      []string `json:"data"`
		DataSaver []string `json:"dataSaver"`
	} `json:"chapter"`
}

// Simplified response types for frontend
type SimpleManga struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	AlternativeTitles []string `json:"alternativeTitles"`
	MatchedTitle      string   `json:"matchedTitle,omitempty"`
	MatchType         string   `json:"matchType,omitempty"`
	Description       string   `json:"description"`
	CoverUrl          string   `json:"coverUrl"`
	Author            string   `json:"author"`
	Artist            string   `json:"artist"`
	Status            string   `json:"status"`
	Year              int      `json:"year"`
	Tags              []string `json:"tags"`
	Rating            string   `json:"contentRating"`
	LastChapter       string   `json:"lastChapter"`
	Languages         []string `json:"availableLanguages"`
}

type SimpleChapter struct {
	ID                 string    `json:"id"`
	MangaID            string    `json:"mangaId"`
	Chapter            string    `json:"chapter"`
	Title              string    `json:"title"`
	Volume             string    `json:"volume"`
	TranslatedLanguage string    `json:"translatedLanguage"`
	Pages              int       `json:"pages"`
	PublishAt          time.Time `json:"publishAt"`
	ScanlationGroup    string    `json:"scanlationGroup,omitempty"`
	ScanlationGroupID  string    `json:"scanlationGroupId,omitempty"`
}
