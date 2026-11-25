package sync

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/categories"
	"github.com/Ssnakerss/E-Wardrobe/internal/items"
	"github.com/Ssnakerss/E-Wardrobe/internal/photos"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Service handles sync orchestration.
type Service struct {
	pool     *database.Pool
	repo     *Repository
	journal  *changejournal.Writer
	catSvc   *categories.Service
	itemSvc  *items.Service
	photoSvc *photos.Service
}

// NewService creates a new sync service.
func NewService(
	pool *database.Pool,
	repo *Repository,
	journal *changejournal.Writer,
	catSvc *categories.Service,
	itemSvc *items.Service,
	photoSvc *photos.Service,
) *Service {
	return &Service{
		pool:     pool,
		repo:     repo,
		journal:  journal,
		catSvc:   catSvc,
		itemSvc:  itemSvc,
		photoSvc: photoSvc,
	}
}

// Push processes a push request from a client.
func (s *Service) Push(ctx context.Context, userID uuid.UUID, req PushRequest) (*PushResponse, error) {
	return s.processPush(ctx, userID, req)
}

// Pull returns changes since the given sequence number.
func (s *Service) Pull(ctx context.Context, userID uuid.UUID, since int64, limit int32) (*PullResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	changes, err := s.repo.ListSyncChanges(ctx, s.pool, userID, since, limit)
	if err != nil {
		return nil, fmt.Errorf("list sync changes: %w", err)
	}

	hasMore := false
	nextSince := since

	if len(changes) > int(limit) {
		hasMore = true
		changes = changes[:limit]
	}

	if len(changes) > 0 {
		nextSince = changes[len(changes)-1].Seq
	}

	if len(changes) == 0 {
		nextSince = since
	}

	return &PullResponse{
		Changes:   changes,
		NextSince: nextSince,
		HasMore:   hasMore,
	}, nil
}

// txAdapter wraps pgx.Tx to implement database.Tx.
type txAdapter struct {
	pgx.Tx
}

func (a *txAdapter) ExecContext(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return a.Tx.Exec(ctx, query, args...)
}

func (a *txAdapter) QueryContext(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return a.Tx.Query(ctx, query, args...)
}

func (a *txAdapter) QueryRowContext(ctx context.Context, query string, args ...any) pgx.Row {
	return a.Tx.QueryRow(ctx, query, args...)
}

func (a *txAdapter) Begin(ctx context.Context) (pgx.Tx, error) {
	return a.Tx.Begin(ctx)
}

// execAdapter wraps pgx.Tx to implement database.Executor.
type execAdapter struct {
	pgx.Tx
}

func (a *execAdapter) ExecContext(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return a.Tx.Exec(ctx, query, args...)
}

func (a *execAdapter) QueryContext(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return a.Tx.Query(ctx, query, args...)
}

func (a *execAdapter) QueryRowContext(ctx context.Context, query string, args ...any) pgx.Row {
	return a.Tx.QueryRow(ctx, query, args...)
}

// dispatch routes a mutation to the appropriate domain applier.
func (s *Service) dispatch(ctx context.Context, tx pgx.Tx, userID uuid.UUID, change changejournal.Change, operation string) error {
	// Wrap pgx.Tx as database.Tx for domain methods
	wrapped := &txAdapter{tx}

	switch change.Entity {
	case "category":
		return s.dispatchCategory(ctx, wrapped, userID, change, operation)
	case "item":
		return s.dispatchItem(ctx, wrapped, userID, change, operation)
	case "photo":
		return s.dispatchPhoto(ctx, wrapped, userID, change, operation)
	default:
		return fmt.Errorf("unknown entity type: %s", change.Entity)
	}
}

// dispatchCategory routes a category mutation to the categories service.
func (s *Service) dispatchCategory(ctx context.Context, tx database.Tx, userID uuid.UUID, change changejournal.Change, operation string) error {
	switch operation {
	case "create":
		return s.catSvc.ApplyCreate(ctx, tx, userID, change)
	case "update":
		return s.catSvc.ApplyUpdate(ctx, tx, userID, change)
	case "delete":
		return s.catSvc.ApplyDelete(ctx, tx, userID, change)
	default:
		return fmt.Errorf("unknown category operation: %s", operation)
	}
}

// dispatchItem routes an item mutation to the items service.
func (s *Service) dispatchItem(ctx context.Context, tx database.Tx, userID uuid.UUID, change changejournal.Change, operation string) error {
	switch operation {
	case "create":
		return items.ApplyCreate(ctx, tx, userID, change, s.itemSvc)
	case "update":
		return items.ApplyUpdate(ctx, tx, userID, change, s.itemSvc)
	case "delete":
		return items.ApplyDelete(ctx, tx, userID, change, s.itemSvc)
	case "archive":
		return items.ApplyArchive(ctx, tx, userID, change, s.itemSvc)
	case "unarchive":
		return items.ApplyUnarchive(ctx, tx, userID, change, s.itemSvc)
	default:
		return fmt.Errorf("unknown item operation: %s", operation)
	}
}

// dispatchPhoto routes a photo mutation to the photos service.
func (s *Service) dispatchPhoto(ctx context.Context, tx database.Tx, userID uuid.UUID, change changejournal.Change, operation string) error {
	switch operation {
	case "create":
		return photos.ApplyCreate(ctx, tx, userID, change, s.photoSvc)
	case "update":
		return photos.ApplyUpdate(ctx, tx, userID, change, s.photoSvc)
	case "delete":
		return photos.ApplyDelete(ctx, tx, userID, change, s.photoSvc)
	default:
		return fmt.Errorf("unknown photo operation: %s", operation)
	}
}
