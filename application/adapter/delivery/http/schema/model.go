package schema

import (
	"encoding/json"
	"fmt"
)

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
