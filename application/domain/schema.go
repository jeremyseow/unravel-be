package domain

import (
	"time"

	"github.com/google/uuid"
)

type SchemaParameter struct {
	ParameterKey string `json:"parameter_key"`
	IsRequired   bool   `json:"is_required"`
}

type SchemaChanges struct {
	BumpType       string   `json:"bump_type"`
	AddedParams    []string `json:"added_params,omitempty"`
	RemovedParams  []string `json:"removed_params,omitempty"`
	PromotedParams []string `json:"promoted_params,omitempty"` // optional → required
	DemotedParams  []string `json:"demoted_params,omitempty"`  // required → optional
}

type ListSchemasFilter struct {
	Lifecycle *string
	Name      *string
}

type Schema struct {
	ID            int64           `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	SchemaKey     string          `json:"schema_key"`
	SchemaName    string          `json:"schema_name"`
	SchemaVersion string          `json:"schema_version"`
	Description   *string         `json:"description"`
	IsLatest      *bool           `json:"is_latest"`
	Lifecycle     *string         `json:"lifecycle"`
	Parameters    []SchemaParameter `json:"parameters"`
	Changes       *SchemaChanges  `json:"changes,omitempty"`
	CreatedAt     *time.Time      `json:"created_at"`
	UpdatedAt     *time.Time      `json:"updated_at"`
}
