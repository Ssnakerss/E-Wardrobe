package sync

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// Repository handles database operations for sync.
type Repository struct {
	pool *database.Pool
}

// NewRepository creates a new sync repository.
func NewRepository(pool *database.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetProcessedMutation checks if a mutation has already been processed.
func (r *Repository) GetProcessedMutation(ctx context.Context, exec database.Executor, userID, mutationID uuid.UUID) (sqlc.ProcessedMutation, error) {
	q := sqlc.NewFromExecutor(exec)

	pm, err := q.GetProcessedMutation(ctx, sqlc.GetProcessedMutationParams{
		UserID:     userID,
		MutationID: mutationID,
	})
	if err != nil {
		if err == sqlc.ErrNotFound {
			return sqlc.ProcessedMutation{}, nil
		}
		return sqlc.ProcessedMutation{}, fmt.Errorf("get processed mutation: %w", err)
	}

	return pm, nil
}

// CreateProcessedMutation records a mutation as processed.
func (r *Repository) CreateProcessedMutation(ctx context.Context, exec database.Executor, userID, mutationID uuid.UUID, requestHash string, entity string, entityID uuid.UUID, result map[string]any) error {
	q := sqlc.NewFromExecutor(exec)

	err := q.CreateProcessedMutation(ctx, sqlc.CreateProcessedMutationParams{
		UserID:      userID,
		MutationID:  mutationID,
		RequestHash: requestHash,
		Entity:      entity,
		EntityID:    entityID,
		Result:      result,
	})
	if err != nil {
		return fmt.Errorf("create processed mutation: %w", err)
	}

	return nil
}

// ListSyncChanges returns changes since the given sequence number.
func (r *Repository) ListSyncChanges(ctx context.Context, exec database.Executor, userID uuid.UUID, since int64, limit int32) ([]SyncChange, error) {
	q := sqlc.NewFromExecutor(exec)

	changes, err := q.ListSyncChanges(ctx, sqlc.ListSyncChangesParams{
		UserID: userID,
		Since:  since,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list sync changes: %w", err)
	}

	result := make([]SyncChange, len(changes))
	for i, c := range changes {
		result[i] = SyncChange{
			Seq:           c.Seq,
			Entity:        c.Entity,
			EntityID:      c.EntityID,
			Operation:     c.Operation,
			EntityVersion: c.EntityVersion,
			Payload:       c.Payload,
		}
	}

	return result, nil
}

// GetOrCreateUserSyncState gets or creates the sync state for a user.
func (r *Repository) GetOrCreateUserSyncState(ctx context.Context, exec database.Executor, userID uuid.UUID) (sqlc.UserSyncState, error) {
	q := sqlc.NewFromExecutor(exec)

	state, err := q.GetOrCreateUserSyncState(ctx, userID)
	if err != nil {
		if err == sqlc.ErrNotFound {
			return q.UpsertUserSyncState(ctx, userID)
		}
		return sqlc.UserSyncState{}, fmt.Errorf("get or create user sync state: %w", err)
	}

	return state, nil
}
