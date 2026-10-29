package database

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

// StringArray makes a validated []string suitable for the PostgreSQL TEXT[]
// site-theme image column without exposing driver details to callers.
func StringArray(values []string) pq.StringArray {
	return pq.StringArray(values)
}

const (
	ChatRoomPublic            = "public"
	ChatRoomInviteOnly        = "invite_only"
	DisplayThemeMidnight      = "midnight"
	DisplayThemeForest        = "forest"
	DisplayThemePlum          = "plum"
	DisplayThemeEmber         = "ember"
	DisplayThemeParchment     = "parchment"
	SiteThemeRotationStatic   = "static"
	SiteThemeRotationRefresh  = "refresh"
	SiteThemeRotationSession  = "session"
	SiteThemeRotationInterval = "interval"
)

// IsDisplayTheme reports whether a client-supplied presentation theme is one
// of the deliberately finite, accessible palette choices.
func IsDisplayTheme(theme string) bool {
	switch theme {
	case DisplayThemeMidnight, DisplayThemeForest, DisplayThemePlum, DisplayThemeEmber, DisplayThemeParchment:
		return true
	default:
		return false
	}
}

// NormalizeCustomThemeColor accepts an optional CSS-safe hexadecimal color and
// returns its canonical lowercase form. Empty clears a previous custom color.
func NormalizeCustomThemeColor(color string) (string, bool) {
	color = strings.ToLower(strings.TrimSpace(color))
	if color == "" {
		return "", true
	}
	if len(color) != 7 || color[0] != '#' {
		return "", false
	}
	for _, character := range color[1:] {
		if !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'f') {
			return "", false
		}
	}
	return color, true
}

// IsSiteThemeRotationSeconds limits image rotation to predictable, modest
// intervals. Zero means that the selected site theme image stays static.
func IsSiteThemeRotationSeconds(seconds int) bool {
	switch seconds {
	case 0, 15, 30, 60:
		return true
	default:
		return false
	}
}

func IsSiteThemeRotationMode(mode string) bool {
	switch mode {
	case SiteThemeRotationStatic, SiteThemeRotationRefresh, SiteThemeRotationSession, SiteThemeRotationInterval:
		return true
	default:
		return false
	}
}

func IsSiteThemeRotationMinutes(minutes int) bool {
	return minutes >= 1 && minutes <= 1440
}

var (
	ErrChatRoomNotFound           = errors.New("chat room not found")
	ErrChatRoomAccess             = errors.New("chat room access denied")
	ErrChatInviteRequired         = errors.New("chat room requires an invitation")
	ErrChatInviteInvalid          = errors.New("chat invitation is invalid")
	ErrChatMessageReportDenied    = errors.New("chat message report denied")
	ErrChatMessageAlreadyReported = errors.New("chat message already reported")
	ErrChatMessageReplyTarget     = errors.New("chat message reply target is invalid")
	ErrAniListMangaCommentDenied  = errors.New("AniList manga comment action denied")
	ErrChatRoomBanned             = errors.New("chat room member is banned")
	ErrChatRoomMemberNotFound     = errors.New("chat room member not found")
	ErrChatRoomModerationDenied   = errors.New("chat room moderation action denied")
	ErrProfileCommentsDisabled    = errors.New("profile comments are disabled")
	ErrProfileCommentDenied       = errors.New("profile comment action denied")
)

// User represents a user in the system
type User struct {
	ID                            int            `json:"id" db:"id"`
	Email                         string         `json:"email" db:"email"`
	Role                          string         `json:"role" db:"role"`
	DisplayName                   string         `json:"display_name" db:"display_name"`
	ProfileImagePath              string         `json:"profile_image_path" db:"profile_image_path"`
	DisplayTheme                  string         `json:"display_theme" db:"display_theme"`
	CustomThemeColor              string         `json:"custom_theme_color" db:"custom_theme_color"`
	ThemeImagePath                string         `json:"theme_image_path" db:"theme_image_path"`
	SiteTheme                     string         `json:"site_theme" db:"site_theme"`
	SiteCustomColor               string         `json:"site_custom_color" db:"site_custom_color"`
	SiteThemeImagePaths           pq.StringArray `json:"site_theme_image_paths" db:"site_theme_image_paths"`
	SiteThemeImageRotationSeconds int            `json:"site_theme_image_rotation_seconds" db:"site_theme_image_rotation_seconds"`
	SiteThemeImageRotationMode    string         `json:"site_theme_image_rotation_mode" db:"site_theme_image_rotation_mode"`
	SiteThemeImageRotationMinutes int            `json:"site_theme_image_rotation_minutes" db:"site_theme_image_rotation_minutes"`
	PasswordHash                  string         `json:"-" db:"password_hash"` // "-" prevents password from being serialized to JSON
	CreatedAt                     time.Time      `json:"created_at" db:"created_at"`
}

// AdminDashboardStats contains aggregate-only metrics safe to expose to an
// authorized administrator dashboard. It intentionally contains no personal
// account or session data.
type AdminDashboardStats struct {
	Users           int `json:"users"`
	AniListLibrary  int `json:"aniListLibrary"`
	MangaDexLibrary int `json:"mangaDexLibrary"`
	Favorites       int `json:"favorites"`
	TrackedChapters int `json:"trackedChapters"`
}

// ChatRoomSummary is the safe room metadata shown to a signed-in user.
type ChatRoomSummary struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Visibility       string `json:"visibility"`
	ImagePath        string `json:"imagePath"`
	Theme            string `json:"theme"`
	CustomThemeColor string `json:"customThemeColor"`
	ThemeImagePath   string `json:"themeImagePath"`
	OwnerUserID      int    `json:"ownerUserId"`
	OwnerDisplayName string `json:"ownerDisplayName"`
	MemberCount      int    `json:"memberCount"`
	IsMember         bool   `json:"isMember"`
	IsOwner          bool   `json:"isOwner"`
	IsModerator      bool   `json:"isModerator"`
	IsMuted          bool   `json:"isMuted"`
	CreatedAt        string `json:"createdAt"`
}

// ChatRoomMember is the safe identity data an owner needs to manage room
// moderators. It intentionally excludes emails, account roles, and sessions.
type ChatRoomMember struct {
	UserID           int    `json:"userId"`
	DisplayName      string `json:"displayName"`
	ProfileImagePath string `json:"profileImagePath"`
	IsOwner          bool   `json:"isOwner"`
	IsModerator      bool   `json:"isModerator"`
}

// ChatRoomMemberDirectoryEntry is the minimal identity shown to other room
// members. It omits avatar paths, moderation state, email, and account data.
type ChatRoomMemberDirectoryEntry struct {
	UserID      int    `json:"userId"`
	DisplayName string `json:"displayName"`
}

// ChatMessage is a member-visible message. Email addresses and session data
// are intentionally excluded.
type ChatMessage struct {
	ID                int64  `json:"id"`
	RoomID            int64  `json:"roomId"`
	UserID            int    `json:"userId"`
	AuthorDisplayName string `json:"authorDisplayName"`
	ProfileImagePath  string `json:"profileImagePath"`
	Body              string `json:"body"`
	ImagePath         string `json:"imagePath"`
	ReplyToMessageID  int64  `json:"replyToMessageId"`
	CreatedAt         string `json:"createdAt"`
}

// AniListMangaComment is a public discussion entry. It deliberately contains
// only the author's safe profile identity, never account or session fields.
type AniListMangaComment struct {
	ID                int64  `json:"id"`
	AniListID         int    `json:"anilistId"`
	UserID            int    `json:"userId"`
	AuthorDisplayName string `json:"authorDisplayName"`
	ProfileImagePath  string `json:"profileImagePath"`
	Body              string `json:"body"`
	CreatedAt         string `json:"createdAt"`
}

// ProfileComment is a public-profile discussion entry. It includes only the
// author's safe display identity, never account or session information.
type ProfileComment struct {
	ID                int64  `json:"id"`
	ProfileUserID     int    `json:"profileUserId"`
	UserID            int    `json:"userId"`
	AuthorDisplayName string `json:"authorDisplayName"`
	ProfileImagePath  string `json:"profileImagePath"`
	Body              string `json:"body"`
	CreatedAt         string `json:"createdAt"`
}

type Notification struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	TargetPath string `json:"targetPath"`
	CreatedAt  string `json:"createdAt"`
	Read       bool   `json:"read"`
}

// Manga represents manga metadata from MangaDex
type Manga struct {
	ID                 int            `json:"id" db:"id"`
	MangaDexID         string         `json:"mangadex_id" db:"mangadex_id"` // UUID from MangaDex
	Title              string         `json:"title" db:"title"`
	AlternativeTitles  pq.StringArray `json:"alternative_titles" db:"alternative_titles"`
	Description        sql.NullString `json:"description" db:"description"`
	Author             sql.NullString `json:"author" db:"author"`
	Artist             sql.NullString `json:"artist" db:"artist"`
	Tags               pq.StringArray `json:"tags" db:"tags"`                     // Genre/theme tags
	Status             string         `json:"status" db:"status"`                 // ongoing, completed, hiatus, cancelled
	ContentRating      sql.NullString `json:"content_rating" db:"content_rating"` // safe, suggestive, erotica, pornographic
	Year               sql.NullInt32  `json:"year" db:"year"`
	LastChapter        sql.NullString `json:"last_chapter" db:"last_chapter"`
	LastVolume         sql.NullString `json:"last_volume" db:"last_volume"`
	OriginalLanguage   sql.NullString `json:"original_language" db:"original_language"`
	AvailableLanguages pq.StringArray `json:"available_languages" db:"available_languages"`
	CoverImagePath     sql.NullString `json:"cover_image_path" db:"cover_image_path"`
	CoverFileName      sql.NullString `json:"cover_filename" db:"cover_filename"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
}

// Chapter represents a manga chapter from MangaDex
type Chapter struct {
	ID                 int            `json:"id" db:"id"`
	MangaID            int            `json:"manga_id" db:"manga_id"`
	ChapterMangaDexID  string         `json:"chapter_mangadex_id" db:"chapter_mangadex_id"` // UUID from MangaDex
	ChapterNumber      sql.NullString `json:"chapter_number" db:"chapter_number"`           // Can be string like "1", "1.5", etc.
	Title              sql.NullString `json:"title" db:"title"`
	Volume             sql.NullString `json:"volume" db:"volume"`
	TranslatedLanguage string         `json:"translated_language" db:"translated_language"`
	PageCount          int            `json:"page_count" db:"page_count"`
	FolderPath         string         `json:"folder_path" db:"folder_path"` // Local path where images are stored
	ScanlationGroup    sql.NullString `json:"scanlation_group" db:"scanlation_group"`
	PublishAt          sql.NullTime   `json:"publish_at" db:"publish_at"`
	Hash               sql.NullString `json:"hash" db:"hash"`             // MangaDex chapter hash for image URLs
	Downloaded         bool           `json:"downloaded" db:"downloaded"` // Whether images have been downloaded
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
}

// ChapterPage represents individual pages of a chapter
type ChapterPage struct {
	ID         int       `json:"id" db:"id"`
	ChapterID  int       `json:"chapter_id" db:"chapter_id"`
	PageNumber int       `json:"page_number" db:"page_number"`
	FileName   string    `json:"file_name" db:"file_name"`   // Original filename from MangaDex
	LocalPath  string    `json:"local_path" db:"local_path"` // Local storage path
	Downloaded bool      `json:"downloaded" db:"downloaded"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
