package schema

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Schema represents a versioned schema definition
type Schema struct {
	ID        uuid.UUID          `json:"id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Parameters map[string]any    `json:"parameters"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// SchemaRequest represents the incoming request to create/update a schema
type SchemaRequest struct {
	Name       string         `json:"name" binding:"required"`
	Parameters map[string]any `json:"parameters" binding:"required"`
}

// ValidateParameters checks if the parameters are valid JSON
func (s *SchemaRequest) ValidateParameters() error {
	if _, err := json.Marshal(s.Parameters); err != nil {
		return fmt.Errorf("invalid parameters format: %v", err)
	}
	return nil
}
