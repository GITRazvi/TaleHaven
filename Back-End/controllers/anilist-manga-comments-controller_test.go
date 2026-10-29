package controllers

import (
	"strings"
	"testing"
)

func TestNormalizeAniListMangaComment(t *testing.T) {
	if body, valid := normalizeAniListMangaComment("  A thoughtful comment.  "); !valid || body != "A thoughtful comment." {
		t.Fatalf("normalized comment = %q, valid = %t", body, valid)
	}
	for _, value := range []string{"", "  ", strings.Repeat("a", maxAniListMangaCommentRunes+1)} {
		if _, valid := normalizeAniListMangaComment(value); valid {
			t.Fatalf("expected invalid comment for %q", value)
		}
	}
}
