package sessions

import (
	"testing"
	"time"
)

func TestCountCompletedChaptersCountsOnlyFinalPages(t *testing.T) {
	progress := &UserReadingData{
		Mangas: map[string]*MangaProgress{
			"manga-one": {
				Chapters: map[string]*ReadingProgress{
					"chapter-one": {CurrentPage: 12, TotalPages: 12},
					"chapter-two": {CurrentPage: 5, TotalPages: 12},
				},
			},
			"manga-two": {
				Chapters: map[string]*ReadingProgress{
					"chapter-three":   {CurrentPage: 8, TotalPages: 8},
					"chapter-unknown": {CurrentPage: 10, TotalPages: 0},
				},
			},
		},
	}

	completed := countCompletedChapters(progress)
	if completed != 2 {
		t.Fatalf("completed chapters = %d, want 2", completed)
	}
}

func TestFindLastReadMangaUsesLatestChapterActivity(t *testing.T) {
	earlier := time.Date(2026, time.October, 20, 10, 0, 0, 0, time.UTC)
	later := earlier.Add(time.Hour)
	progress := &UserReadingData{
		Mangas: map[string]*MangaProgress{
			"manga-one": {Chapters: map[string]*ReadingProgress{"chapter-one": {LastReadAt: earlier}}},
			"manga-two": {Chapters: map[string]*ReadingProgress{"chapter-two": {LastReadAt: later}}},
		},
	}

	lastRead := findLastReadManga(progress)
	if lastRead == nil || lastRead.MangaID != "manga-two" || !lastRead.LastReadAt.Equal(later) {
		t.Fatalf("last read manga = %#v, want manga-two at %v", lastRead, later)
	}
}
