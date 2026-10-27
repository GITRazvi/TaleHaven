package sessions

import (
	"TaleHaven-Backend/modules/auth"
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"errors"
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

func TestAuthenticatedSessionCredentialsAreOpaqueAndParseable(t *testing.T) {
	credentials, err := NewAuthenticatedSessionCredentials()
	if err != nil {
		t.Fatalf("create credentials: %v", err)
	}
	if credentials.ID == "" || credentials.Token == "" || credentials.VerifierHash == "" {
		t.Fatalf("credentials should contain ID, token and verifier hash: %#v", credentials)
	}
	if credentials.Token == credentials.VerifierHash {
		t.Fatal("browser token must not equal the server-side verifier hash")
	}

	id, verifier, err := ParseAuthenticatedSessionToken(credentials.Token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if id != credentials.ID {
		t.Fatalf("token ID = %q, want %q", id, credentials.ID)
	}
	if !authenticatedSessionVerifierMatches(verifier, credentials.VerifierHash) {
		t.Fatal("valid verifier should match its stored hash")
	}
	if authenticatedSessionVerifierMatches(verifier+"tampered", credentials.VerifierHash) {
		t.Fatal("tampered verifier should not match its stored hash")
	}
}

func TestParseAuthenticatedSessionTokenRejectsMalformedValues(t *testing.T) {
	for _, token := range []string{"", "identifier-only", "not-hex-not-hex.not-base64", "1234.bad"} {
		if _, _, err := ParseAuthenticatedSessionToken(token); err != ErrInvalidAuthenticatedSession {
			t.Fatalf("parse %q error = %v, want ErrInvalidAuthenticatedSession", token, err)
		}
	}
}

func TestAuthenticatedSessionCanBeRevoked(t *testing.T) {
	suffix := time.Now().UnixNano()
	user, err := auth.CreateUser(fmt.Sprintf("session-%d@example.test", suffix), "password-for-session-test")
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	t.Cleanup(func() {
		if _, err := database.DB.Exec(`DELETE FROM users WHERE id = $1`, user.ID); err != nil {
			t.Errorf("clean up session test user: %v", err)
		}
	})

	token, err := CreateAuthenticatedSession(user.ID, "127.0.0.1", "session-test")
	if err != nil {
		t.Fatalf("create authenticated session: %v", err)
	}
	if _, authenticatedUser, err := AuthenticateSession(token); err != nil || authenticatedUser.ID != user.ID {
		t.Fatalf("authenticate new session = (%v, %v), want test user", authenticatedUser, err)
	}

	sessionID, _, err := ParseAuthenticatedSessionToken(token)
	if err != nil {
		t.Fatalf("parse session token: %v", err)
	}
	var storedVerifierHash string
	if err := database.DB.QueryRow(`SELECT verifier_hash FROM sessions WHERE id = $1`, sessionID).Scan(&storedVerifierHash); err != nil {
		t.Fatalf("load stored session verifier: %v", err)
	}
	if storedVerifierHash == token {
		t.Fatal("raw browser session token must not be stored in the database")
	}

	if err := RevokeAuthenticatedSession(token); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	if _, _, err := AuthenticateSession(token); !errors.Is(err, ErrExpiredAuthenticatedSession) {
		t.Fatalf("authenticate revoked session error = %v, want ErrExpiredAuthenticatedSession", err)
	}
}
