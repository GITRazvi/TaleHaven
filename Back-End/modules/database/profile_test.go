package database

import (
	"fmt"
	"testing"
	"time"
)

func TestUserProfileUpdatesAreScopedToTheRequestedUser(t *testing.T) {
	suffix := time.Now().UnixNano()
	var firstUserID int
	err := DB.QueryRow(`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`, fmt.Sprintf("profile-first-%d@example.test", suffix), "profile-test-password").Scan(&firstUserID)
	if err != nil {
		t.Fatalf("create first user: %v", err)
	}
	var secondUserID int
	err = DB.QueryRow(`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`, fmt.Sprintf("profile-second-%d@example.test", suffix), "profile-test-password").Scan(&secondUserID)
	if err != nil {
		t.Fatalf("create second user: %v", err)
	}
	t.Cleanup(func() {
		if _, err := DB.Exec(`DELETE FROM users WHERE id IN ($1, $2)`, firstUserID, secondUserID); err != nil {
			t.Errorf("clean up profile test users: %v", err)
		}
	})

	var firstName, firstImage, firstTheme, firstCustomThemeColor string
	err = DB.QueryRow(`UPDATE users SET display_name = $1, display_theme = $2, custom_theme_color = $3 WHERE id = $4 RETURNING display_name, profile_image_path, display_theme, custom_theme_color`, "First Reader", DisplayThemeForest, "#387c5d", firstUserID).Scan(&firstName, &firstImage, &firstTheme, &firstCustomThemeColor)
	if err != nil {
		t.Fatalf("update first display name: %v", err)
	}
	if firstName != "First Reader" || firstImage != "" || firstTheme != DisplayThemeForest || firstCustomThemeColor != "#387c5d" {
		t.Fatalf("updated first profile = (%q, %q, %q, %q)", firstName, firstImage, firstTheme, firstCustomThemeColor)
	}

	var secondName, secondImage string
	err = DB.QueryRow(`UPDATE users SET profile_image_path = $1 WHERE id = $2 RETURNING display_name, profile_image_path`, "/images/profiles/user-2.jpg", secondUserID).Scan(&secondName, &secondImage)
	if err != nil {
		t.Fatalf("update second profile photo: %v", err)
	}
	if secondImage != "/images/profiles/user-2.jpg" {
		t.Fatalf("updated second profile = (%q, %q)", secondName, secondImage)
	}

	if err := DB.QueryRow(`SELECT display_name, profile_image_path, display_theme, custom_theme_color FROM users WHERE id = $1`, firstUserID).Scan(&firstName, &firstImage, &firstTheme, &firstCustomThemeColor); err != nil {
		t.Fatalf("reload first user: %v", err)
	}
	if firstName != "First Reader" || firstImage != "" || firstTheme != DisplayThemeForest || firstCustomThemeColor != "#387c5d" {
		t.Fatalf("first user should not receive second user's image update: (%q, %q, %q, %q)", firstName, firstImage, firstTheme, firstCustomThemeColor)
	}
}

func TestNormalizeCustomThemeColor(t *testing.T) {
	color, valid := NormalizeCustomThemeColor(" #A1b2C3 ")
	if !valid || color != "#a1b2c3" {
		t.Fatalf("normalize custom theme color = (%q, %t)", color, valid)
	}
	for _, invalid := range []string{"#123", "blue", "#12345g", "#1234567"} {
		if _, valid := NormalizeCustomThemeColor(invalid); valid {
			t.Fatalf("invalid custom theme color %q succeeded", invalid)
		}
	}
}
