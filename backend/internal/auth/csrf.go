package auth

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// GetCSRFToken returns the CSRF token for the given session.
func GetCSRFToken(ctx context.Context, q sqlc.Querier, sessionID uuid.UUID) (string, error) {
	s, err := q.GetSessionByTokenHash(ctx, sessionID.String())
	if err != nil {
		return "", fmt.Errorf("get session: %w", err)
	}

	if s.CSRFTokenHash == nil {
		return "", fmt.Errorf("session does not have CSRF token")
	}

	return *s.CSRFTokenHash, nil
}
