package auth

import (
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	utils.InitLogger()
	if err := os.Chdir("../.."); err != nil {
		panic(fmt.Sprintf("switch to backend root for test configuration: %v", err))
	}
	utils.LoadEnv()
	database.InitPostgres()
	code := m.Run()
	database.CloseDB()
	os.Exit(code)
}

func TestProfileUpdateOperationsReturnSafeUpdatedUser(t *testing.T) {
	user, err := CreateUser(fmt.Sprintf("auth-profile-%d@example.test", time.Now().UnixNano()), "profile-test-password")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		if _, err := database.DB.Exec(`DELETE FROM users WHERE id = $1`, user.ID); err != nil {
			t.Errorf("clean up user: %v", err)
		}
	})

	updated, err := UpdateUserDisplayName(user.ID, "New Display Name")
	if err != nil {
		t.Fatalf("update display name: %v", err)
	}
	if updated.DisplayName != "New Display Name" || updated.PasswordHash != "" {
		t.Fatalf("display name update returned %#v", updated)
	}

	updated, err = UpdateUserProfileImagePath(user.ID, "/images/profiles/user-test.png")
	if err != nil {
		t.Fatalf("update profile image: %v", err)
	}
	if updated.ProfileImagePath != "/images/profiles/user-test.png" || updated.PasswordHash != "" {
		t.Fatalf("profile image update returned %#v", updated)
	}

	updated, err = UpdateUserDisplayTheme(user.ID, database.DisplayThemePlum, "#8a4fa3")
	if err != nil {
		t.Fatalf("update display theme: %v", err)
	}
	if updated.DisplayTheme != database.DisplayThemePlum || updated.CustomThemeColor != "#8a4fa3" || updated.PasswordHash != "" {
		t.Fatalf("display theme update returned %#v", updated)
	}
}
