package domain

import (
	"time"

	"github.com/google/uuid"
)

type SchemaParameter struct {
	ParameterKey string `json:"parameter_key"`
	IsRequired   bool   `json:"is_required"`
}

type Schema struct {
	ID            int64             `json:"id"`
	TenantID      uuid.UUID         `json:"tenant_id"`
	SchemaKey     string            `json:"schema_key"`
	SchemaName    string            `json:"schema_name"`
	SchemaVersion string            `json:"schema_version"`
	Description   *string           `json:"description"`
	IsLatest      *bool             `json:"is_latest"`
	Lifecycle     *string           `json:"lifecycle"`
	Parameters    []SchemaParameter `json:"parameters"`
	CreatedAt     *time.Time        `json:"created_at"`
	UpdatedAt     *time.Time        `json:"updated_at"`
}
