package sessions

import (
	"TaleHaven-Backend/modules/auth"
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

var (
	ErrInvalidAuthenticatedSession = errors.New("invalid session")
	ErrExpiredAuthenticatedSession = errors.New("expired session")
)

const (
	AuthenticatedSessionIdleDuration     = 30 * time.Minute
	AuthenticatedSessionAbsoluteDuration = 14 * 24 * time.Hour
)

// AuthenticatedSession is the server-side state for a browser session.
type AuthenticatedSession struct {
	ID                string
	UserID            int
	CreatedAt         time.Time
	LastSeen          time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	RevokedAt         *time.Time
	IPAddress         string
	UserAgent         string
}

// SessionCredentials separates a public lookup identifier from the secret
// verifier. Only Token may leave the server; the database retains only the
// verifier hash.
type SessionCredentials struct {
	ID           string
	Token        string
	VerifierHash string
}

func NewAuthenticatedSessionCredentials() (*SessionCredentials, error) {
	idBytes := make([]byte, 16)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, err
	}
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}

	id := hex.EncodeToString(idBytes)
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)
	return &SessionCredentials{
		ID:           id,
		Token:        id + "." + verifier,
		VerifierHash: authenticatedSessionVerifierHash(verifier),
	}, nil
}

// ParseAuthenticatedSessionToken refuses an ID-only cookie and validates the
// expected identifier and verifier encodings before a database lookup.
func ParseAuthenticatedSessionToken(token string) (id, verifier string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || len(parts[0]) != 32 || parts[1] == "" {
		return "", "", ErrInvalidAuthenticatedSession
	}
	if _, decodeErr := hex.DecodeString(parts[0]); decodeErr != nil {
		return "", "", ErrInvalidAuthenticatedSession
	}
	if _, decodeErr := base64.RawURLEncoding.DecodeString(parts[1]); decodeErr != nil {
		return "", "", ErrInvalidAuthenticatedSession
	}
	return parts[0], parts[1], nil
}

func authenticatedSessionVerifierHash(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return hex.EncodeToString(digest[:])
}

func authenticatedSessionVerifierMatches(verifier, expectedHash string) bool {
	actualHash := authenticatedSessionVerifierHash(verifier)
	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}

// CreateAuthenticatedSession stores a revocable session and returns the opaque
// cookie token. The raw verifier is intentionally never persisted.
func CreateAuthenticatedSession(userID int, ipAddress, userAgent string) (string, error) {
	credentials, err := NewAuthenticatedSessionCredentials()
	if err != nil {
		return "", fmt.Errorf("generate session credentials: %w", err)
	}

	now := time.Now().UTC()
	idleExpiry := now.Add(AuthenticatedSessionIdleDuration)
	absoluteExpiry := now.Add(AuthenticatedSessionAbsoluteDuration)
	query := `
		INSERT INTO sessions (
			id, verifier_hash, user_id, created_at, expires_at, last_seen,
			idle_expires_at, absolute_expires_at, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = database.DB.Exec(query, credentials.ID, credentials.VerifierHash, userID,
		now, absoluteExpiry, now, idleExpiry, absoluteExpiry, ipAddress, userAgent)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	utils.Logger.Info("Authenticated session created", zap.String("session_id", credentials.ID), zap.Int("user_id", userID))
	return credentials.Token, nil
}

// AuthenticateSession validates the cookie token against a server-stored hash,
// checks both expiry windows and returns the associated user.
func AuthenticateSession(token string) (*AuthenticatedSession, *database.User, error) {
	sessionID, verifier, err := ParseAuthenticatedSessionToken(token)
	if err != nil {
		return nil, nil, err
	}

	session := &AuthenticatedSession{}
	var verifierHash string
	var revokedAt sql.NullTime
	query := `
		SELECT id, user_id, verifier_hash, created_at, last_seen, idle_expires_at,
			absolute_expires_at, revoked_at, ip_address, user_agent
		FROM sessions WHERE id = $1
	`
	err = database.DB.QueryRow(query, sessionID).Scan(
		&session.ID, &session.UserID, &verifierHash, &session.CreatedAt,
		&session.LastSeen, &session.IdleExpiresAt, &session.AbsoluteExpiresAt,
		&revokedAt, &session.IPAddress, &session.UserAgent,
	)
	if err == sql.ErrNoRows {
		return nil, nil, ErrInvalidAuthenticatedSession
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get session: %w", err)
	}
	if verifierHash == "" || !authenticatedSessionVerifierMatches(verifier, verifierHash) {
		return nil, nil, ErrInvalidAuthenticatedSession
	}
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}
	now := time.Now().UTC()
	if session.RevokedAt != nil || !now.Before(session.IdleExpiresAt) || !now.Before(session.AbsoluteExpiresAt) {
		return nil, nil, ErrExpiredAuthenticatedSession
	}

	user, err := auth.GetUserByID(session.UserID)
	if err != nil {
		return nil, nil, ErrInvalidAuthenticatedSession
	}

	newIdleExpiry := now.Add(AuthenticatedSessionIdleDuration)
	if _, err := database.DB.Exec(`UPDATE sessions SET last_seen = $1, idle_expires_at = $2 WHERE id = $3`, now, newIdleExpiry, session.ID); err != nil {
		return nil, nil, fmt.Errorf("refresh session: %w", err)
	}
	session.LastSeen = now
	session.IdleExpiresAt = newIdleExpiry
	return session, user, nil
}

// RevokeAuthenticatedSession invalidates a session at the server. An invalid
// browser cookie is an idempotent logout rather than an error.
func RevokeAuthenticatedSession(token string) error {
	sessionID, _, err := ParseAuthenticatedSessionToken(token)
	if err != nil {
		return nil
	}
	_, err = database.DB.Exec(`UPDATE sessions SET revoked_at = COALESCE(revoked_at, $1) WHERE id = $2`, time.Now().UTC(), sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	utils.Logger.Info("Session revoked", zap.String("session_id", sessionID))
	return nil
}

// RevokeAllAuthenticatedSessions invalidates every session after a password
// change or future account-security event.
func RevokeAllAuthenticatedSessions(userID int) error {
	_, err := database.DB.Exec(`UPDATE sessions SET revoked_at = COALESCE(revoked_at, $1) WHERE user_id = $2`, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

// CleanupExpiredAuthenticatedSessions removes session records that are already
// revoked or beyond either server-enforced expiry window.
func CleanupExpiredAuthenticatedSessions() error {
	now := time.Now().UTC()
	_, err := database.DB.Exec(`
		DELETE FROM sessions
		WHERE revoked_at IS NOT NULL
			OR idle_expires_at < $1
			OR absolute_expires_at < $1
	`, now)
	if err != nil {
		return fmt.Errorf("cleanup authenticated sessions: %w", err)
	}
	return nil
}
