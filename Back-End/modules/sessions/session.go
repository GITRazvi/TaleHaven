package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Session represents a user session
type Session struct {
	ID        string
	UserID    *int       // Optional - for authenticated users
	CreatedAt time.Time
	ExpiresAt time.Time
	LastSeen  time.Time
	IPAddress string
	UserAgent string
}

// GenerateSessionID creates a new secure session ID
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Refresh extends the session expiration time
func (s *Session) Refresh(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
	s.LastSeen = time.Now()
}

