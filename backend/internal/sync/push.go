package sync

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// maxMutationsPerPush is the maximum number of mutations in a single push request.
const maxMutationsPerPush = 100

// processPush validates and processes a batch of mutations.
// Each mutation is processed in its own transaction for independence.
func (s *Service) processPush(ctx context.Context, userID uuid.UUID, req PushRequest) (*PushResponse, error) {
	if len(req.Mutations) == 0 {
		return &PushResponse{Results: []MutationResult{}}, nil
	}

	if len(req.Mutations) > maxMutationsPerPush {
		return nil, fmt.Errorf("too many mutations: max %d, got %d", maxMutationsPerPush, len(req.Mutations))
	}

	results := make([]MutationResult, 0, len(req.Mutations))

	for _, mutation := range req.Mutations {
		result := s.processOneMutation(ctx, userID, mutation)
		results = append(results, result)
	}

	return &PushResponse{Results: results}, nil
}

// processOneMutation processes a single mutation in its own transaction.
func (s *Service) processOneMutation(ctx context.Context, userID uuid.UUID, mutation Mutation) MutationResult {
	result := MutationResult{
		MutationID: mutation.MutationID,
		Entity:     mutation.Entity,
		EntityID:   mutation.EntityID,
	}

	// Validate basic fields
	if !ValidEntityType(mutation.Entity) {
		result.Status = StatusValidationError
		result.Error = "unsupported entity type: " + mutation.Entity
		return result
	}

	if !ValidOperation(mutation.Operation) {
		result.Status = StatusValidationError
		result.Error = "unsupported operation: " + mutation.Operation
		return result
	}

	// Compute canonical hash
	requestHash := CanonicalHash(mutation.Payload)

	// Each mutation in its own transaction
	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		// Wrap pgx.Tx for sqlc queries
		exec := &execAdapter{tx}
		q := sqlc.NewFromExecutor(exec)

		// 1. Lock user_sync_state FOR UPDATE
		_, err := q.GetOrCreateUserSyncState(ctx, userID)
		if err != nil {
			return fmt.Errorf("lock sync state: %w", err)
		}

		// 2. Check processed_mutations
		existing, err := q.GetProcessedMutation(ctx, sqlc.GetProcessedMutationParams{
			UserID:     userID,
			MutationID: mutation.MutationID,
		})
		if err != nil && err != sqlc.ErrNotFound {
			return fmt.Errorf("check processed: %w", err)
		}

		if existing.MutationID != uuid.Nil {
			// Already processed — compare hash
			if existing.RequestHash == requestHash {
				// Same mutation — return stored result
				result.Status = StatusDuplicate
				if status, ok := existing.Result["status"].(string); ok {
					result.Status = MutationStatus(status)
				}
				if ver, ok := existing.Result["entity_version"].(float64); ok {
					result.EntityVersion = int64(ver)
				}
				return nil
			}

			// Same mutation_id, different payload — reused ID
			result.Status = StatusMutationIDReused
			result.Error = "mutation_id reused with different payload"
			return nil
		}

		// 3. Check base_version for creates
		if mutation.Operation == "create" && mutation.BaseVersion != 0 {
			result.Status = StatusValidationError
			result.Error = "create mutation must have base_version=0"
			return nil
		}

		// 4. Build changejournal.Change for dispatch
		change := changejournal.Change{
			UserID:        userID,
			Entity:        mutation.Entity,
			EntityID:      mutation.EntityID,
			Operation:     toSyncOperation(mutation.Operation),
			EntityVersion: mutation.BaseVersion + 1,
			Payload:       mutation.Payload,
		}

		// 5. Dispatch to domain applier (wraps tx as database.Tx)
		err = s.dispatch(ctx, tx, userID, change, mutation.Operation)
		if err != nil {
			result.Status = s.mapDomainError(err)
			result.Error = err.Error()
			return nil // not a transaction error
		}

		// 6. Allocate next seq and write sync_change
		seq, err := q.NextSeq(ctx, userID)
		if err != nil {
			return fmt.Errorf("allocate seq: %w", err)
		}

		change.Seq = seq
		err = s.journal.WriteRecord(ctx, exec, change)
		if err != nil {
			return fmt.Errorf("write journal: %w", err)
		}

		// 7. Save processed mutation
		result.Status = StatusApplied
		result.EntityVersion = change.EntityVersion

		resultPayload := map[string]any{
			"status":         string(result.Status),
			"entity_version": result.EntityVersion,
		}

		err = q.CreateProcessedMutation(ctx, sqlc.CreateProcessedMutationParams{
			UserID:      userID,
			MutationID:  mutation.MutationID,
			RequestHash: requestHash,
			Entity:      mutation.Entity,
			EntityID:    mutation.EntityID,
			Result:      resultPayload,
		})
		if err != nil {
			return fmt.Errorf("save processed mutation: %w", err)
		}

		return nil
	})

	if err != nil {
		// Transaction error (not a domain error)
		if result.Status == "" {
			result.Status = StatusValidationError
		}
		if result.Error == "" {
			result.Error = err.Error()
		}
	}

	return result
}

// toSyncOperation converts a client operation to a sync journal operation.
func toSyncOperation(op string) string {
	switch op {
	case "create", "update":
		return "upsert"
	case "delete", "archive", "unarchive":
		return "delete"
	default:
		return "upsert"
	}
}

// mapDomainError maps domain errors to sync status codes.
func (s *Service) mapDomainError(err error) MutationStatus {
	errStr := lowercase(err.Error())

	// Check for version conflict
	if containsStr(errStr, "version conflict") || containsStr(errStr, "optimistic lock") {
		return StatusConflict
	}

	// Check for validation errors
	if containsStr(errStr, "validation") || containsStr(errStr, "invalid") || containsStr(errStr, "required") || containsStr(errStr, "must be") {
		return StatusValidationError
	}

	// Check for missing dependencies
	if containsStr(errStr, "not found") || containsStr(errStr, "does not exist") || containsStr(errStr, "foreign key") || containsStr(errStr, "dependency") {
		return StatusDependencyMissing
	}

	// Check for ownership errors
	if containsStr(errStr, "not owned") || containsStr(errStr, "forbidden") || containsStr(errStr, "permission denied") {
		return StatusValidationError
	}

	return StatusValidationError
}

// containsStr checks if s contains sub.
func containsStr(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	if len(s) < len(sub) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// lowercase is a simple ASCII lowercase.
func lowercase(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + 32
		}
		result[i] = c
	}
	return string(result)
}
