package sessions

import (
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	SessionDuration = 30 * 24 * time.Hour // 30 days
)

// CreateSession creates a new session in the database
func CreateSession(ipAddress, userAgent string) (*Session, error) {
	sessionID, err := GenerateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	session := &Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(SessionDuration),
		LastSeen:  time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	query := `
		INSERT INTO sessions (id, user_id, created_at, expires_at, last_seen, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = database.DB.Exec(query, session.ID, nil, session.CreatedAt, session.ExpiresAt, 
		session.LastSeen, session.IPAddress, session.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	utils.Logger.Info("Session created", zap.String("sessionId", session.ID))
	return session, nil
}

// GetSession retrieves a session by ID
func GetSession(sessionID string) (*Session, error) {
	session := &Session{}

	query := `
		SELECT id, user_id, created_at, expires_at, last_seen, ip_address, user_agent
		FROM sessions
		WHERE id = $1
	`

	var userID sql.NullInt32
	err := database.DB.QueryRow(query, sessionID).Scan(
		&session.ID,
		&userID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastSeen,
		&session.IPAddress,
		&session.UserAgent,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if userID.Valid {
		uid := int(userID.Int32)
		session.UserID = &uid
	}

	return session, nil
}

// UpdateSessionActivity updates the last seen time and extends expiration
func UpdateSessionActivity(sessionID string) error {
	query := `
		UPDATE sessions
		SET last_seen = $1, expires_at = $2
		WHERE id = $3
	`

	_, err := database.DB.Exec(query, time.Now(), time.Now().Add(SessionDuration), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	return nil
}

// DeleteSession removes a session
func DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := database.DB.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	utils.Logger.Info("Session deleted", zap.String("sessionId", sessionID))
	return nil
}

// CleanupExpiredSessions removes expired sessions (call periodically)
func CleanupExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < $1`
	result, err := database.DB.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		utils.Logger.Info("Cleaned up expired sessions", zap.Int64("count", count))
	}

	return nil
}

// AssociateSessionWithUser associates an existing session with a user (for login)
func AssociateSessionWithUser(sessionID string, userID int) error {
	query := `UPDATE sessions SET user_id = $1 WHERE id = $2`
	_, err := database.DB.Exec(query, userID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to associate session with user: %w", err)
	}

	utils.Logger.Info("Session associated with user",
		zap.String("session_id", sessionID),
		zap.Int("user_id", userID))
	return nil
}

// DisassociateSessionFromUser removes the user association from a session (for logout)
func DisassociateSessionFromUser(sessionID string) error {
	query := `UPDATE sessions SET user_id = NULL WHERE id = $1`
	_, err := database.DB.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to disassociate session from user: %w", err)
	}

	utils.Logger.Info("Session disassociated from user", zap.String("session_id", sessionID))
	return nil
}

