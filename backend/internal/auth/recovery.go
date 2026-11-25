package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// CreateRecoveryToken creates a recovery token for the given user.
func CreateRecoveryToken(ctx context.Context, q sqlc.Querier, userID uuid.UUID, ttl time.Duration) (string, error) {
	token, err := GenerateRandomHex(32)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	tokenHash := HashToken(token)

	_, err = q.CreateRecoveryToken(ctx, sqlc.CreateRecoveryTokenParams{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(ttl),
	})
	if err != nil {
		return "", fmt.Errorf("create recovery token: %w", err)
	}

	return token, nil
}

// ValidateRecoveryToken checks if a recovery token is valid and returns the user ID.
func ValidateRecoveryToken(ctx context.Context, q sqlc.Querier, token string) (uuid.UUID, error) {
	tokenHash := HashToken(token)

	t, err := q.GetRecoveryTokenByHash(ctx, tokenHash)
	if err != nil {
		return uuid.Nil, fmt.Errorf("recovery token invalid or expired")
	}

	return t.UserID, nil
}

// CompleteRecovery updates the user's password and revokes all sessions.
// The token parameter is the raw recovery token string.
func CompleteRecovery(ctx context.Context, q sqlc.Querier, token string, newPassword string) error {
	tokenHash := HashToken(token)

	// Get the token by hash first
	t, err := q.GetRecoveryTokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("recovery token invalid or expired")
	}

	// Mark token as used
	_, err = q.MarkRecoveryTokenUsed(ctx, t.ID)
	if err != nil {
		return fmt.Errorf("mark recovery token used: %w", err)
	}

	// Get the user
	user, err := q.GetUserByID(ctx, t.UserID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	// Hash new password
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Update password
	_, err = q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Revoke all sessions
	if err := RevokeUserSessions(ctx, q, user.ID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}

	return nil
}
