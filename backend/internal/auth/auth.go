package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrSessionExpired  = errors.New("session expired")
)

// Session represents an active user session.
type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Client     string
	DeviceName *string
	ExpiresAt  time.Time
	LastUsed   *time.Time
}

// CreateUser creates a new user with the given email and hashed password.
func CreateUser(ctx context.Context, q sqlc.Querier, email, passwordHash string) (*sqlc.User, error) {
	user, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

// AuthenticateUser looks up a user by email and verifies the password.
func AuthenticateUser(ctx context.Context, q sqlc.Querier, email string, password string) (*sqlc.User, error) {
	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sqlc.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if !VerifyPassword(password, user.PasswordHash) {
		return nil, ErrInvalidPassword
	}

	return &user, nil
}

// CreateSession creates a new session for the user.
func CreateSession(ctx context.Context, q sqlc.Querier, userID uuid.UUID, client string, deviceName *string, ttl time.Duration) (*Session, error) {
	token, err := GenerateRandomHex(32)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	tokenHash := HashToken(token)

	var csrfToken *string
	if client == "web" {
		cs, err := GenerateRandomHex(32)
		if err != nil {
			return nil, fmt.Errorf("generate csrf: %w", err)
		}
		csrfToken = &cs
	}

	var deviceNameSQL sql.NullString
	if deviceName != nil {
		deviceNameSQL = sql.NullString{String: *deviceName, Valid: true}
	}

	s, err := q.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:            uuid.New(),
		UserID:        userID,
		TokenHash:     tokenHash,
		CSRFTokenHash: csrfToken,
		Client:        client,
		DeviceName:    deviceNameSQL,
		ExpiresAt:     time.Now().Add(ttl),
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &Session{
		ID:         s.ID,
		UserID:     s.UserID,
		Client:     s.Client,
		DeviceName: nullableStringFromSQL(s.DeviceName),
		ExpiresAt:  s.ExpiresAt,
		LastUsed:   nullableTimeFromSQL(s.LastUsed),
	}, nil
}

// GetSession looks up a session by its opaque token.
func GetSession(ctx context.Context, q sqlc.Querier, token string) (*Session, error) {
	tokenHash := HashToken(token)

	s, err := q.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == sqlc.ErrNotFound {
			return nil, ErrSessionExpired
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	return &Session{
		ID:         s.ID,
		UserID:     s.UserID,
		Client:     s.Client,
		DeviceName: nullableStringFromSQL(s.DeviceName),
		ExpiresAt:  s.ExpiresAt,
		LastUsed:   nullableTimeFromSQL(s.LastUsed),
	}, nil
}

// RevokeSession revokes a session.
func RevokeSession(ctx context.Context, q sqlc.Querier, sessionID uuid.UUID) error {
	_, err := q.RevokeSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// RevokeUserSessions revokes all sessions for a user.
func RevokeUserSessions(ctx context.Context, q sqlc.Querier, userID uuid.UUID) error {
	_, err := q.RevokeUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

// GetUserIDFromSession extracts the user ID from a session token.
func GetUserIDFromSession(ctx context.Context, q sqlc.Querier, token string) (uuid.UUID, error) {
	s, err := GetSession(ctx, q, token)
	if err != nil {
		return uuid.Nil, err
	}
	return s.UserID, nil
}

// GetUserIDFromSessionPool extracts the user ID from a session token using a database pool.
func GetUserIDFromSessionPool(ctx context.Context, pool *database.Pool, token string) (uuid.UUID, error) {
	q := sqlc.New(pool)
	return GetUserIDFromSession(ctx, q, token)
}

func nullableStringFromSQL(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func nullableTimeFromSQL(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
