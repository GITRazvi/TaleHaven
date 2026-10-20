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
