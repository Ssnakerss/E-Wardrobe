package changejournal

import (
	"context"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/google/uuid"
)

// Model defines the interface for change journal operations.
type Model interface {
	// WriteRecord writes a change within an existing transaction.
	WriteRecord(ctx context.Context, tx database.Tx, change Change) error

	// PullChanges returns changes since the given sequence number.
	PullChanges(ctx context.Context, userID uuid.UUID, since int64, limit int32) ([]Change, error)
}
