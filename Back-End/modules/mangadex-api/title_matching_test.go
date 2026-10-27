package mangadex_api

import "testing"

func TestAlternativeTitlesAreUsedForDisplayAndSearch(t *testing.T) {
	manga := MangaData{Attributes: MangaAttributes{
		Title: map[string]string{"ja-ro": "Shingeki no Kyojin"},
		AltTitles: []map[string]string{
			{"en": "Attack on Titan"},
			{"ja": "進撃の巨人"},
		},
	}}

	simplified := TransformToSimpleManga(manga)
	if simplified.Title != "Attack on Titan" {
		t.Fatalf("display title = %q, want English alternative title", simplified.Title)
	}
	if !containsTitle(simplified.AlternativeTitles, "Shingeki no Kyojin") || !containsTitle(simplified.AlternativeTitles, "進撃の巨人") {
		t.Fatalf("alternativeTitles = %#v, want primary and Japanese aliases", simplified.AlternativeTitles)
	}

	matchedTitle, matchType, rank := MatchMangaTitle(manga, "Attack on Titan!")
	if matchedTitle != "Attack on Titan" || matchType != "exact" || rank != 0 {
		t.Fatalf("match = (%q, %q, %d), want exact alternative title match", matchedTitle, matchType, rank)
	}
}

func TestMatchMangaTitleRanksExactBeforePrefixAndPartial(t *testing.T) {
	exact := MangaData{Attributes: MangaAttributes{AltTitles: []map[string]string{{"en": "Attack on Titan"}}}}
	prefix := MangaData{Attributes: MangaAttributes{Title: map[string]string{"en": "Attack on Titan Junior High"}}}
	partial := MangaData{Attributes: MangaAttributes{Title: map[string]string{"en": "The Attack on Titan Guide"}}}

	_, _, exactRank := MatchMangaTitle(exact, "attack on titan")
	_, _, prefixRank := MatchMangaTitle(prefix, "attack on titan")
	_, _, partialRank := MatchMangaTitle(partial, "attack on titan")
	if exactRank != 0 || prefixRank != 1 || partialRank != 2 {
		t.Fatalf("ranks = exact %d, prefix %d, partial %d; want 0, 1, 2", exactRank, prefixRank, partialRank)
	}
}

func TestFindConfidentMangaMatchRequiresOneSafeCandidate(t *testing.T) {
	exact := MangaData{ID: "exact", Attributes: MangaAttributes{Title: map[string]string{"en": "Omniscient Reader's Viewpoint"}}}
	sequel := MangaData{ID: "sequel", Attributes: MangaAttributes{Title: map[string]string{"en": "Omniscient Reader's Viewpoint Side Story"}}}
	if matched, ok := FindConfidentMangaMatch([]MangaData{exact, sequel}, []string{"Omniscient Readers Viewpoint"}); !ok || matched.ID != "exact" {
		t.Fatalf("match = %#v, %v; want exact candidate", matched, ok)
	}
	if _, ok := FindConfidentMangaMatch([]MangaData{exact, MangaData{ID: "duplicate", Attributes: exact.Attributes}}, []string{"Omniscient Reader's Viewpoint"}); ok {
		t.Fatal("ambiguous exact candidates must not be selected")
	}
}

func TestFindConfidentMangaMatchAcceptsReorderedAlternativeTitle(t *testing.T) {
	candidate := MangaData{
		ID: "reordered-alias",
		Attributes: MangaAttributes{
			Title:     map[string]string{"en": "Unrelated Display Title"},
			AltTitles: []map[string]string{{"en": "Ragnarok: Solo Leveling"}},
		},
	}

	matched, ok := FindConfidentMangaMatch(
		[]MangaData{candidate},
		[]string{"Solo Leveling Ragnarok"},
	)
	if !ok || matched.ID != candidate.ID {
		t.Fatalf("match = %#v, %v; want reordered alternative title", matched, ok)
	}
}

func TestConfidentTitleRankRejectsMissingRepeatedWord(t *testing.T) {
	if rank := confidentTitleRank("My My Hero", "My Hero"); rank != 3 {
		t.Fatalf("rank = %d, want no match when a repeated word is missing", rank)
	}
}

func TestFindConfidentMangaMatchUsesYearAndCreatorToBreakATie(t *testing.T) {
	first := MangaData{
		ID: "first",
		Attributes: MangaAttributes{
			Title: map[string]string{"en": "The Same Title"},
			Year:  2020,
		},
		Relationships: []Relationship{{Type: "author", Attributes: map[string]interface{}{"name": "Alice Example"}}},
	}
	second := MangaData{
		ID: "second",
		Attributes: MangaAttributes{
			Title: map[string]string{"en": "The Same Title"},
			Year:  2024,
		},
		Relationships: []Relationship{{Type: "artist", Attributes: map[string]interface{}{"name": "Bob Example"}}},
	}

	matched, ok := FindConfidentMangaMatchWithMetadata([]MangaData{first, second}, MatchMetadata{
		Titles:       []string{"The Same Title"},
		Year:         2020,
		CreatorNames: []string{"Alice Example"},
	})
	if !ok || matched.ID != "first" {
		t.Fatalf("match = %#v, %v; want first candidate", matched, ok)
	}
}

func containsTitle(titles []string, wanted string) bool {
	for _, title := range titles {
		if title == wanted {
			return true
		}
	}
	return false
}
