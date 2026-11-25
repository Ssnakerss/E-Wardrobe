package changejournal

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// Repository handles database operations for the change journal.
type Repository struct {
	pool *database.Pool
}

// NewRepository creates a new change journal repository.
func NewRepository(pool *database.Pool) *Repository {
	return &Repository{pool: pool}
}

// WriteRecord writes a change record within a transaction.
func (r *Repository) WriteRecord(ctx context.Context, tx database.Tx, change Change) error {
	q := sqlc.NewFromTx(tx)

	// Allocate next seq
	seq, err := q.NextSeq(ctx, change.UserID)
	if err != nil {
		return fmt.Errorf("allocate seq: %w", err)
	}

	// Write the change record
	err = q.CreateSyncChange(ctx, sqlc.CreateSyncChangeParams{
		UserID:        change.UserID,
		Seq:           seq,
		Entity:        change.Entity,
		EntityID:      change.EntityID,
		Operation:     change.Operation,
		EntityVersion: change.EntityVersion,
		Payload:       change.Payload,
	})
	if err != nil {
		return fmt.Errorf("write change: %w", err)
	}

	return nil
}

// PullChanges returns changes since the given sequence number.
func (r *Repository) PullChanges(ctx context.Context, userID uuid.UUID, since int64, limit int32) ([]Change, error) {
	q := sqlc.New(r.pool)

	changes, err := q.ListSyncChanges(ctx, sqlc.ListSyncChangesParams{
		UserID: userID,
		Since:  since,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list sync changes: %w", err)
	}

	result := make([]Change, len(changes))
	for i, c := range changes {
		result[i] = Change{
			UserID:        c.UserID,
			Seq:           c.Seq,
			Entity:        c.Entity,
			EntityID:      c.EntityID,
			Operation:     c.Operation,
			EntityVersion: c.EntityVersion,
			Payload:       c.Payload,
			ChangedAt:     c.ChangedAt,
		}
	}

	return result, nil
}
