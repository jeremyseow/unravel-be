package domain

import (
	"time"
	"github.com/google/uuid"
)

type Schema struct {
	ID            int64          `json:"id"`
	TenantID      uuid.UUID      `json:"tenant_id"`
	SchemaKey     string         `json:"schema_key"`
	SchemaVersion string         `json:"schema_version"`
	Description   *string        `json:"description"`
	Parameters    []string       `json:"parameters"`
	CreatedAt     *time.Time     `json:"created_at"`
	UpdatedAt     *time.Time     `json:"updated_at"`
}
