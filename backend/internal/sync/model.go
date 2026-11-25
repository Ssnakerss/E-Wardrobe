package sync

import (
	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/google/uuid"
)

// MutationStatus represents the result of processing a single mutation.
type MutationStatus string

const (
	StatusApplied        MutationStatus = "applied"
	StatusDuplicate      MutationStatus = "duplicate"
	StatusConflict       MutationStatus = "conflict"
	StatusValidationError MutationStatus = "validation_error"
	StatusDependencyMissing MutationStatus = "dependency_missing"
	StatusMutationIDReused MutationStatus = "mutation_id_reused"
)

// ValidEntityType checks if the entity type is supported.
func ValidEntityType(entity string) bool {
	switch entity {
	case "category", "item", "photo":
		return true
	}
	return false
}

// ValidOperation checks if the operation is supported.
func ValidOperation(op string) bool {
	switch op {
	case "create", "update", "delete", "archive", "unarchive":
		return true
	}
	return false
}

// Mutation represents a client mutation from the outbox.
type Mutation struct {
	MutationID  uuid.UUID `json:"mutation_id"`
	Entity      string    `json:"entity"`
	EntityID    uuid.UUID `json:"entity_id"`
	Operation   string    `json:"operation"`
	BaseVersion int64     `json:"base_version"`
	Payload     map[string]any `json:"payload"`
}

// MutationResult holds the outcome of processing one mutation.
type MutationResult struct {
	MutationID    uuid.UUID      `json:"mutation_id"`
	Entity        string         `json:"entity"`
	EntityID      uuid.UUID      `json:"entity_id"`
	Status        MutationStatus `json:"status"`
	EntityVersion int64          `json:"entity_version,omitempty"`
	Error         string         `json:"error,omitempty"`
	Current       map[string]any `json:"current,omitempty"`
}

// PushRequest is the client push payload.
type PushRequest struct {
	InstallationID uuid.UUID    `json:"installation_id"`
	Mutations      []Mutation   `json:"mutations"`
}

// PushResponse is the server response to a push request.
type PushResponse struct {
	Results []MutationResult `json:"results"`
}

// SyncChange represents a single change in the sync journal for pull responses.
type SyncChange struct {
	Seq           int64            `json:"seq"`
	Entity        string           `json:"entity"`
	EntityID      uuid.UUID        `json:"entity_id"`
	Operation     string           `json:"operation"`
	EntityVersion int64            `json:"entity_version"`
	Payload       map[string]any   `json:"payload"`
}

// PullResponse is the server response to a pull request.
type PullResponse struct {
	Changes  []SyncChange `json:"changes"`
	NextSince int64       `json:"next_since"`
	HasMore  bool         `json:"has_more"`
}

// Applier applies a mutation to a domain entity.
type Applier interface {
	ApplyCreate(ctx interface{ ExecContext(interface{}, string, ...interface{}) (interface{}, error) }, userID uuid.UUID, mutation changejournal.Change) error
	ApplyUpdate(ctx interface{ ExecContext(interface{}, string, ...interface{}) (interface{}, error) }, userID uuid.UUID, mutation changejournal.Change) error
	ApplyDelete(ctx interface{ ExecContext(interface{}, string, ...interface{}) (interface{}, error) }, userID uuid.UUID, mutation changejournal.Change) error
}

// Snapshotter produces a full snapshot map for a given entity.
type Snapshotter interface {
	SnapshotCategory(id uuid.UUID, userID uuid.UUID) (map[string]any, error)
	SnapshotItem(id uuid.UUID, userID uuid.UUID) (map[string]any, error)
	SnapshotPhoto(id uuid.UUID, userID uuid.UUID) (map[string]any, error)
}
