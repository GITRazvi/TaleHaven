package mangadex_api

import "testing"

func TestSelectChaptersMergesAllGroupsAndFillsGaps(t *testing.T) {
	chapters := []ChapterData{
		chapterFor("a1", "en", "1", "group-a", "Group A"),
		chapterFor("a2", "en", "2", "group-a", "Group A"),
		chapterFor("b1", "en", "1", "group-b", "Group B"),
		chapterFor("b3", "en", "3", "group-b", "Group B"),
		chapterFor("ro1", "ro", "1", "group-ro", "Romanian Group"),
	}

	selected, languages, groups := SelectChapters(chapters, "en", "all")
	if len(selected) != 3 || selected[0].ID != "a1" || selected[1].ID != "a2" || selected[2].ID != "b3" {
		t.Fatalf("selected = %#v, want Group A with Group B fallback for chapter 3", selected)
	}
	if len(languages) != 2 || languages[0] != "en" || languages[1] != "ro" {
		t.Fatalf("languages = %#v, want en and ro", languages)
	}
	if len(groups) != 2 || groups[0].ID != "group-a" || groups[0].ChapterCount != 2 {
		t.Fatalf("groups = %#v, want ordered English coverage", groups)
	}
}

func TestSelectChaptersFiltersOneGroupAndKeepsUnnumberedChapters(t *testing.T) {
	chapters := []ChapterData{
		chapterFor("a1", "en", "1", "group-a", "Group A"),
		chapterFor("b1", "en", "1", "group-b", "Group B"),
		chapterFor("extra", "en", "", "group-a", "Group A"),
	}

	selected, _, _ := SelectChapters(chapters, "en", "group-a")
	selectedIDs := map[string]bool{}
	for _, chapter := range selected {
		selectedIDs[chapter.ID] = true
	}
	if len(selected) != 2 || !selectedIDs["a1"] || !selectedIDs["extra"] {
		t.Fatalf("selected = %#v, want only Group A including unnumbered chapter", selected)
	}
}

func TestSelectChaptersOrdersAllLanguagesByChapterBeforeLanguage(t *testing.T) {
	chapters := []ChapterData{
		chapterFor("romanian-50", "ro", "50", "group-ro", "Romanian Group"),
		chapterFor("english-1", "en", "1", "group-en", "English Group"),
		chapterFor("romanian-1", "ro", "1", "group-ro", "Romanian Group"),
		chapterFor("english-2", "en", "2", "group-en", "English Group"),
	}

	selected, _, _ := SelectChapters(chapters, "all", "all")
	if len(selected) != 4 {
		t.Fatalf("selected = %#v, want four language-specific releases", selected)
	}
	got := []string{selected[0].ID, selected[1].ID, selected[2].ID, selected[3].ID}
	want := []string{"english-1", "romanian-1", "english-2", "romanian-50"}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("all-language chapter order = %#v, want %#v", got, want)
		}
	}
}

func chapterFor(id, language, number, groupID, groupName string) ChapterData {
	return ChapterData{
		ID:            id,
		Attributes:    ChapterAttributes{TranslatedLanguage: language, Chapter: number},
		Relationships: []Relationship{{Type: "scanlation_group", ID: groupID, Attributes: map[string]interface{}{"name": groupName}}},
	}
}
