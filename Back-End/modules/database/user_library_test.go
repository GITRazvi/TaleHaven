package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"TaleHaven-Backend/modules/utils"
)

func TestMain(m *testing.M) {
	utils.InitLogger()
	if err := os.Chdir("../.."); err != nil {
		panic(fmt.Sprintf("switch to backend root for test configuration: %v", err))
	}
	utils.LoadEnv()
	InitPostgres()

	code := m.Run()
	CloseDB()
	os.Exit(code)
}

func TestUserLibraryDatabaseOperations(t *testing.T) {
	suffix := fmt.Sprintf("library-test-%d", time.Now().UnixNano())
	firstUserID := createLibraryTestUser(t, "first-"+suffix+"@example.test")
	secondUserID := createLibraryTestUser(t, "second-"+suffix+"@example.test")
	mangaDexID, mangaID := createLibraryTestManga(t, suffix)
	secondMangaDexID, secondMangaID := createLibraryTestManga(t, suffix+"-second")

	t.Cleanup(func() {
		if _, err := DB.Exec("DELETE FROM manga WHERE id IN ($1, $2)", mangaID, secondMangaID); err != nil {
			t.Errorf("clean up test manga: %v", err)
		}
		if _, err := DB.Exec("DELETE FROM users WHERE id IN ($1, $2)", firstUserID, secondUserID); err != nil {
			t.Errorf("clean up test users: %v", err)
		}
	})

	added, err := AddMangaToUserLibrary(firstUserID, mangaID)
	if err != nil {
		t.Fatalf("add manga to first user library: %v", err)
	}
	if !added {
		t.Fatal("first add should create a library entry")
	}

	added, err = AddMangaToUserLibrary(firstUserID, mangaID)
	if err != nil {
		t.Fatalf("repeat add should be idempotent: %v", err)
	}
	if added {
		t.Fatal("repeat add should not create a duplicate library entry")
	}

	firstLibrary, total, err := GetUserLibraryPage(firstUserID, 24, 0)
	if err != nil {
		t.Fatalf("list first user library: %v", err)
	}
	if total != 1 || len(firstLibrary) != 1 || firstLibrary[0].MangaDexID != mangaDexID {
		t.Fatalf("first user library = %#v, want only %q", firstLibrary, mangaDexID)
	}

	secondLibrary, total, err := GetUserLibraryPage(secondUserID, 24, 0)
	if err != nil {
		t.Fatalf("list second user library: %v", err)
	}
	if total != 0 || len(secondLibrary) != 0 {
		t.Fatalf("second user should not see first user's manga: %#v", secondLibrary)
	}

	added, err = AddMangaToUserLibrary(firstUserID, secondMangaID)
	if err != nil || !added {
		t.Fatalf("add second manga to first user library = (%t, %v), want (true, nil)", added, err)
	}
	firstPage, total, err := GetUserLibraryPage(firstUserID, 1, 0)
	if err != nil {
		t.Fatalf("load first paginated library result: %v", err)
	}
	secondPage, secondPageTotal, err := GetUserLibraryPage(firstUserID, 1, 1)
	if err != nil {
		t.Fatalf("load second paginated library result: %v", err)
	}
	if total != 2 || secondPageTotal != 2 || len(firstPage) != 1 || len(secondPage) != 1 || firstPage[0].MangaDexID == secondPage[0].MangaDexID {
		t.Fatalf("paginated library pages = %#v and %#v with totals %d/%d, want two distinct results", firstPage, secondPage, total, secondPageTotal)
	}

	added, err = AddMangaToUserLibrary(secondUserID, mangaID)
	if err != nil || !added {
		t.Fatalf("add manga to second user library = (%t, %v), want (true, nil)", added, err)
	}

	removed, err := RemoveMangaFromUserLibrary(firstUserID, mangaDexID)
	if err != nil {
		t.Fatalf("remove manga from first user library: %v", err)
	}
	if !removed {
		t.Fatal("first remove should delete the library entry")
	}

	firstLibrary, total, err = GetUserLibraryPage(firstUserID, 24, 0)
	if err != nil {
		t.Fatalf("list first user library after removal: %v", err)
	}
	if total != 1 || len(firstLibrary) != 1 || firstLibrary[0].MangaDexID != secondMangaDexID {
		t.Fatalf("first user library after removal = %#v, want only %q", firstLibrary, secondMangaDexID)
	}

	secondLibrary, total, err = GetUserLibraryPage(secondUserID, 24, 0)
	if err != nil {
		t.Fatalf("list second user library after first user removal: %v", err)
	}
	if total != 1 || len(secondLibrary) != 1 || secondLibrary[0].MangaDexID != mangaDexID {
		t.Fatalf("first user removal should not affect second user library: %#v", secondLibrary)
	}

	removed, err = RemoveMangaFromUserLibrary(firstUserID, mangaDexID)
	if err != nil {
		t.Fatalf("repeat remove should be idempotent: %v", err)
	}
	if removed {
		t.Fatal("repeat remove should not delete an entry")
	}

	removed, err = RemoveMangaFromUserLibrary(firstUserID, secondMangaDexID)
	if err != nil || !removed {
		t.Fatalf("remove second manga from first user library = (%t, %v), want (true, nil)", removed, err)
	}
	firstLibrary, total, err = GetUserLibraryPage(firstUserID, 24, 0)
	if err != nil {
		t.Fatalf("list first user library after removing all manga: %v", err)
	}
	if total != 0 || len(firstLibrary) != 0 {
		t.Fatalf("first user library after removing all manga = %#v, want empty", firstLibrary)
	}
}

func TestAniListLibraryReadingStatusIsUserScoped(t *testing.T) {
	suffix := fmt.Sprintf("anilist-status-test-%d", time.Now().UnixNano())
	firstUserID := createLibraryTestUser(t, "first-"+suffix+"@example.test")
	secondUserID := createLibraryTestUser(t, "second-"+suffix+"@example.test")
	aniListID := 1000001

	t.Cleanup(func() {
		if _, err := DB.Exec("DELETE FROM users WHERE id IN ($1, $2)", firstUserID, secondUserID); err != nil {
			t.Errorf("clean up status test users: %v", err)
		}
	})

	added, err := AddAniListMangaToUserLibrary(firstUserID, aniListID)
	if err != nil || !added {
		t.Fatalf("add AniList manga = (%t, %v), want (true, nil)", added, err)
	}

	items, total, err := GetAniListUserLibraryPage(firstUserID, 10, 0)
	if err != nil {
		t.Fatalf("list AniList library: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].ReadingStatus != AniListPlanToRead {
		t.Fatalf("new AniList library item = %#v, want default %q", items, AniListPlanToRead)
	}
	if items[0].PersonalRating != nil {
		t.Fatalf("new AniList library item personal rating = %v, want nil", *items[0].PersonalRating)
	}

	rating := 8
	updated, err := SetAniListMangaPersonalRating(firstUserID, aniListID, &rating)
	if err != nil || !updated {
		t.Fatalf("set first user personal rating = (%t, %v), want (true, nil)", updated, err)
	}
	items, _, err = GetAniListUserLibraryPage(firstUserID, 10, 0)
	if err != nil || items[0].PersonalRating == nil || *items[0].PersonalRating != rating {
		t.Fatalf("updated first user personal rating = %#v, %v", items, err)
	}
	updated, err = SetAniListMangaPersonalRating(firstUserID, aniListID, nil)
	if err != nil || !updated {
		t.Fatalf("clear first user personal rating = (%t, %v), want (true, nil)", updated, err)
	}
	items, _, err = GetAniListUserLibraryPage(firstUserID, 10, 0)
	if err != nil || items[0].PersonalRating != nil {
		t.Fatalf("cleared first user personal rating = %#v, %v", items, err)
	}
	invalidRating := 11
	if _, err := SetAniListMangaPersonalRating(firstUserID, aniListID, &invalidRating); err == nil {
		t.Fatal("invalid personal rating should be rejected")
	}

	updated, err = SetAniListMangaReadingStatus(firstUserID, aniListID, AniListCurrentlyReading)
	if err != nil || !updated {
		t.Fatalf("set first user status = (%t, %v), want (true, nil)", updated, err)
	}
	items, _, err = GetAniListUserLibraryPage(firstUserID, 10, 0)
	if err != nil || items[0].ReadingStatus != AniListCurrentlyReading {
		t.Fatalf("updated first user status = %#v, %v", items, err)
	}
	counts, err := GetAniListUserLibraryStatusCounts(firstUserID)
	if err != nil || counts[AniListCurrentlyReading] != 1 || counts[AniListPlanToRead] != 0 {
		t.Fatalf("first user status counts = %#v, %v", counts, err)
	}

	updated, err = SetAniListMangaReadingStatus(secondUserID, aniListID, AniListRead)
	if err != nil {
		t.Fatalf("set second user status: %v", err)
	}
	if updated {
		t.Fatal("a user must not update another user's saved title")
	}
	updated, err = SetAniListMangaPersonalRating(secondUserID, aniListID, &rating)
	if err != nil {
		t.Fatalf("set second user personal rating: %v", err)
	}
	if updated {
		t.Fatal("a user must not rate another user's saved title")
	}
	if _, err := SetAniListMangaReadingStatus(firstUserID, aniListID, AniListReadingStatus("INVALID")); err == nil {
		t.Fatal("invalid reading status should be rejected")
	}
}

func TestLastCompletedAniListMangaUsesLatestCompletion(t *testing.T) {
	suffix := fmt.Sprintf("anilist-last-completed-test-%d", time.Now().UnixNano())
	userID := createLibraryTestUser(t, "last-completed-"+suffix+"@example.test")
	firstID := 2000001
	secondID := 2000002

	for _, aniListID := range []int{firstID, secondID} {
		added, err := AddAniListMangaToUserLibrary(userID, aniListID)
		if err != nil || !added {
			t.Fatalf("add AniList manga %d = (%t, %v), want (true, nil)", aniListID, added, err)
		}
		updated, err := SetAniListMangaReadingStatus(userID, aniListID, AniListRead)
		if err != nil || !updated {
			t.Fatalf("set AniList manga %d to completed = (%t, %v), want (true, nil)", aniListID, updated, err)
		}
	}

	firstCompletedAt := time.Now().UTC().Add(-time.Hour)
	secondCompletedAt := time.Now().UTC()
	for _, update := range []struct {
		aniListID   int
		completedAt time.Time
	}{
		{firstID, firstCompletedAt},
		{secondID, secondCompletedAt},
	} {
		if _, err := DB.Exec(`UPDATE user_library_anilist_manga SET completed_at = $3 WHERE user_id = $1 AND anilist_id = $2`, userID, update.aniListID, update.completedAt); err != nil {
			t.Fatalf("set completion time for AniList manga %d: %v", update.aniListID, err)
		}
	}

	latestID, found, err := GetLastCompletedAniListMangaID(userID)
	if err != nil || !found || latestID != secondID {
		t.Fatalf("latest completed AniList manga = (%d, %t, %v), want (%d, true, nil)", latestID, found, err, secondID)
	}

	updated, err := SetAniListMangaReadingStatus(userID, secondID, AniListPlanToRead)
	if err != nil || !updated {
		t.Fatalf("move latest completed manga back to plan to read = (%t, %v), want (true, nil)", updated, err)
	}
	latestID, found, err = GetLastCompletedAniListMangaID(userID)
	if err != nil || !found || latestID != firstID {
		t.Fatalf("latest completed AniList manga after status change = (%d, %t, %v), want (%d, true, nil)", latestID, found, err, firstID)
	}
}

func TestFavoritesAreIndependentAndUserScoped(t *testing.T) {
	suffix := fmt.Sprintf("favorites-test-%d", time.Now().UnixNano())
	firstUserID := createLibraryTestUser(t, "favorite-first-"+suffix+"@example.test")
	secondUserID := createLibraryTestUser(t, "favorite-second-"+suffix+"@example.test")
	mangaDexID, mangaID := createLibraryTestManga(t, suffix)
	aniListID := 3000001
	if added, err := AddAniListMangaToUserLibrary(firstUserID, aniListID); err != nil || !added {
		t.Fatalf("add AniList manga to library = (%t, %v), want (true, nil)", added, err)
	}
	if added, err := AddMangaToUserLibrary(firstUserID, mangaID); err != nil || !added {
		t.Fatalf("add MangaDex manga to library = (%t, %v), want (true, nil)", added, err)
	}

	added, err := AddAniListMangaToFavorites(firstUserID, aniListID)
	if err != nil || !added {
		t.Fatalf("add AniList favorite = (%t, %v), want (true, nil)", added, err)
	}
	aniListIDs, err := GetAniListFavoriteIDs(firstUserID)
	if err != nil || len(aniListIDs) != 1 || aniListIDs[0] != aniListID {
		t.Fatalf("first user's AniList favorites = %#v, %v", aniListIDs, err)
	}
	if removed, err := RemoveAniListMangaFromUserLibrary(firstUserID, aniListID); err != nil || !removed {
		t.Fatalf("remove AniList manga from library = (%t, %v), want (true, nil)", removed, err)
	}
	aniListIDs, err = GetAniListFavoriteIDs(firstUserID)
	if err != nil || len(aniListIDs) != 0 {
		t.Fatalf("AniList favorites outside the library = %#v, %v, want empty", aniListIDs, err)
	}
	otherAniListIDs, err := GetAniListFavoriteIDs(secondUserID)
	if err != nil || len(otherAniListIDs) != 0 {
		t.Fatalf("second user's AniList favorites = %#v, %v, want empty", otherAniListIDs, err)
	}

	added, err = AddMangaToFavorites(firstUserID, mangaID)
	if err != nil || !added {
		t.Fatalf("add MangaDex favorite = (%t, %v), want (true, nil)", added, err)
	}
	mangaDexIDs, err := GetMangaDexFavoriteIDs(firstUserID)
	if err != nil || len(mangaDexIDs) != 1 || mangaDexIDs[0] != mangaDexID {
		t.Fatalf("first user's MangaDex favorites = %#v, %v", mangaDexIDs, err)
	}
	favoriteManga, err := GetMangaDexFavoriteManga(firstUserID)
	if err != nil || len(favoriteManga) != 1 || favoriteManga[0].MangaDexID != mangaDexID {
		t.Fatalf("first user's MangaDex favorite display data = %#v, %v", favoriteManga, err)
	}
	otherFavoriteManga, err := GetMangaDexFavoriteManga(secondUserID)
	if err != nil || len(otherFavoriteManga) != 0 {
		t.Fatalf("second user's MangaDex favorite display data = %#v, %v, want empty", otherFavoriteManga, err)
	}
	if removed, err := RemoveMangaFromUserLibrary(firstUserID, mangaDexID); err != nil || !removed {
		t.Fatalf("remove MangaDex manga from library = (%t, %v), want (true, nil)", removed, err)
	}
	mangaDexIDs, err = GetMangaDexFavoriteIDs(firstUserID)
	if err != nil || len(mangaDexIDs) != 0 {
		t.Fatalf("MangaDex favorites outside the library = %#v, %v, want empty", mangaDexIDs, err)
	}
	removed, err := RemoveMangaFromFavorites(firstUserID, mangaDexID)
	if err != nil || !removed {
		t.Fatalf("remove MangaDex favorite = (%t, %v), want (true, nil)", removed, err)
	}
	mangaDexIDs, err = GetMangaDexFavoriteIDs(firstUserID)
	if err != nil || len(mangaDexIDs) != 0 {
		t.Fatalf("MangaDex favorites after removal = %#v, %v, want empty", mangaDexIDs, err)
	}
}

func TestAniListCharacterFavoritesAreUserScoped(t *testing.T) {
	suffix := fmt.Sprintf("character-favorites-test-%d", time.Now().UnixNano())
	firstUserID := createLibraryTestUser(t, "character-favorite-first-"+suffix+"@example.test")
	secondUserID := createLibraryTestUser(t, "character-favorite-second-"+suffix+"@example.test")
	characterID := 4000001
	mangaID := 3000001

	added, err := AddAniListCharacterToFavorites(firstUserID, characterID, mangaID, "MAIN")
	if err != nil || !added {
		t.Fatalf("add character favorite = (%t, %v), want (true, nil)", added, err)
	}
	favorites, err := GetAniListFavoriteCharacters(firstUserID)
	if err != nil || len(favorites) != 1 || favorites[0].CharacterID != characterID || favorites[0].SourceAniListMangaID != mangaID || favorites[0].Role != "MAIN" {
		t.Fatalf("first user's character favorites = %#v, %v", favorites, err)
	}
	otherFavorites, err := GetAniListFavoriteCharacters(secondUserID)
	if err != nil || len(otherFavorites) != 0 {
		t.Fatalf("second user's character favorites = %#v, %v, want empty", otherFavorites, err)
	}
	removed, err := RemoveAniListCharacterFromFavorites(firstUserID, characterID)
	if err != nil || !removed {
		t.Fatalf("remove character favorite = (%t, %v), want (true, nil)", removed, err)
	}
	favorites, err = GetAniListFavoriteCharacters(firstUserID)
	if err != nil || len(favorites) != 0 {
		t.Fatalf("character favorites after removal = %#v, %v, want empty", favorites, err)
	}
}

func TestPreferredScanlationGroupIsScopedByUserMangaAndLanguage(t *testing.T) {
	suffix := fmt.Sprintf("scanlation-preference-test-%d", time.Now().UnixNano())
	firstUserID := createLibraryTestUser(t, "scanlation-first-"+suffix+"@example.test")
	secondUserID := createLibraryTestUser(t, "scanlation-second-"+suffix+"@example.test")
	mangaID := "manga-" + suffix
	groupID := "group-" + suffix

	if err := SetPreferredScanlationGroup(firstUserID, mangaID, "en", groupID); err != nil {
		t.Fatalf("save first user's group preference: %v", err)
	}
	group, err := GetPreferredScanlationGroup(firstUserID, mangaID, "en")
	if err != nil || group != groupID {
		t.Fatalf("first user's preference = (%q, %v), want (%q, nil)", group, err, groupID)
	}
	otherUserGroup, err := GetPreferredScanlationGroup(secondUserID, mangaID, "en")
	if err != nil || otherUserGroup != "all" {
		t.Fatalf("second user's preference = (%q, %v), want (all, nil)", otherUserGroup, err)
	}
	otherLanguageGroup, err := GetPreferredScanlationGroup(firstUserID, mangaID, "es")
	if err != nil || otherLanguageGroup != "all" {
		t.Fatalf("other language preference = (%q, %v), want (all, nil)", otherLanguageGroup, err)
	}
	if err := SetPreferredScanlationGroup(firstUserID, mangaID, "en", "all"); err != nil {
		t.Fatalf("clear group preference: %v", err)
	}
	group, err = GetPreferredScanlationGroup(firstUserID, mangaID, "en")
	if err != nil || group != "all" {
		t.Fatalf("cleared preference = (%q, %v), want (all, nil)", group, err)
	}
	if _, err := DB.Exec("DELETE FROM users WHERE id IN ($1, $2)", firstUserID, secondUserID); err != nil {
		t.Fatalf("clean up preference test users: %v", err)
	}
}

func TestPreferredReadingModeDefaultsAndIsUserScoped(t *testing.T) {
	suffix := fmt.Sprintf("reading-mode-test-%d", time.Now().UnixNano())
	firstUserID := createLibraryTestUser(t, "reading-mode-first-"+suffix+"@example.test")
	secondUserID := createLibraryTestUser(t, "reading-mode-second-"+suffix+"@example.test")
	t.Cleanup(func() {
		if _, err := DB.Exec("DELETE FROM users WHERE id IN ($1, $2)", firstUserID, secondUserID); err != nil {
			t.Errorf("clean up reading-mode test users: %v", err)
		}
	})

	defaultMode, err := GetPreferredReadingMode(firstUserID)
	if err != nil || defaultMode != "vertical" {
		t.Fatalf("default reading mode = (%q, %v), want (vertical, nil)", defaultMode, err)
	}
	if err := SetPreferredReadingMode(firstUserID, "rtl"); err != nil {
		t.Fatalf("save first user's reading mode: %v", err)
	}
	mode, err := GetPreferredReadingMode(firstUserID)
	if err != nil || mode != "rtl" {
		t.Fatalf("first user's reading mode = (%q, %v), want (rtl, nil)", mode, err)
	}
	otherMode, err := GetPreferredReadingMode(secondUserID)
	if err != nil || otherMode != "vertical" {
		t.Fatalf("second user's reading mode = (%q, %v), want (vertical, nil)", otherMode, err)
	}
}

func createLibraryTestUser(t *testing.T, email string) int {
	t.Helper()

	var userID int
	if err := DB.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email,
		"test-password-hash",
	).Scan(&userID); err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return userID
}

func createLibraryTestManga(t *testing.T, suffix string) (string, int) {
	t.Helper()

	mangaDexID := "test-" + suffix
	var mangaID int
	if err := DB.QueryRow(
		"INSERT INTO manga (mangadex_id, title) VALUES ($1, $2) RETURNING id",
		mangaDexID,
		"Library Test Manga "+suffix,
	).Scan(&mangaID); err != nil {
		t.Fatalf("create test manga: %v", err)
	}
	return mangaDexID, mangaID
}
