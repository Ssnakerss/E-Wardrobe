package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

var (
	ErrInvitationExpired  = errors.New("invitation expired")
	ErrInvitationRevoked  = errors.New("invitation revoked")
	ErrInvitationUsed     = errors.New("invitation already used")
	ErrInvitationNotFound = errors.New("invitation not found")
)

// Invitation represents an invitation token with its metadata.
type Invitation struct {
	ID        uuid.UUID
	Email     *string
	ExpiresAt time.Time
	UsedAt    *time.Time
	RevokedAt *time.Time
}

// CreateInvitation creates a new invitation with a random token.
func CreateInvitation(ctx context.Context, q sqlc.Querier, email *string, ttl time.Duration) (*Invitation, error) {
	token, err := GenerateRandomHex(32)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	tokenHash := HashToken(token)

	var emailSQL sql.NullString
	if email != nil {
		emailSQL = sql.NullString{String: *email, Valid: true}
	}

	invite, err := q.CreateInvitation(ctx, sqlc.CreateInvitationParams{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		Email:     emailSQL,
		ExpiresAt: time.Now().Add(ttl),
	})
	if err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	return toInvitation(invite), nil
}

// ValidateInvitation checks if an invitation token is valid and returns its details.
func ValidateInvitation(ctx context.Context, q sqlc.Querier, token string) (*Invitation, error) {
	tokenHash := HashToken(token)

	invite, err := q.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == sqlc.ErrNotFound {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("get invitation: %w", err)
	}

	if invite.RevokedAt.Valid {
		return nil, ErrInvitationRevoked
	}
	if invite.UsedAt.Valid {
		return nil, ErrInvitationUsed
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, ErrInvitationExpired
	}

	return toInvitation(invite), nil
}

// MarkInvitationUsed marks an invitation as used by the given user.
func MarkInvitationUsed(ctx context.Context, q sqlc.Querier, inviteID uuid.UUID, userID uuid.UUID) error {
	_, err := q.MarkInvitationUsed(ctx, sqlc.MarkInvitationUsedParams{
		ID:           inviteID,
		UsedByUserID: userID,
	})
	if err != nil {
		if err == sqlc.ErrNotFound {
			return ErrInvitationUsed
		}
		return fmt.Errorf("mark invitation used: %w", err)
	}
	return nil
}

// RevokeInvitation revokes an invitation.
func RevokeInvitation(ctx context.Context, q sqlc.Querier, inviteID uuid.UUID) error {
	_, err := q.RevokeInvitation(ctx, inviteID)
	if err != nil {
		if err == sqlc.ErrNotFound {
			return ErrInvitationNotFound
		}
		return fmt.Errorf("revoke invitation: %w", err)
	}
	return nil
}

func toInvitation(i sqlc.Invitation) *Invitation {
	return &Invitation{
		ID:        i.ID,
		Email:     nullableStringFromSQL(i.Email),
		ExpiresAt: i.ExpiresAt,
		UsedAt:    nullableTimeFromSQL(i.UsedAt),
		RevokedAt: nullableTimeFromSQL(i.RevokedAt),
	}
}
