package schema

import "time"

// CreateSchemaRequest is the body for POST /schemas.
// Parameters lists parameter_key references from the parameter catalog.
type CreateSchemaRequest struct {
	SchemaKey   string               `json:"schema_key" binding:"required,max=32"`
	SchemaName  string               `json:"schema_name" binding:"required,max=32"`
	Description *string              `json:"description"`
	Parameters  []SchemaParameterRef `json:"parameters" binding:"required,min=1"`
}

// SchemaParameterRef references one entry in the parameter catalog.
type SchemaParameterRef struct {
	ParameterKey string `json:"parameter_key" binding:"required"`
	IsRequired   bool   `json:"is_required"`
}

// SchemaResponse is the API representation of a schema version.
type SchemaResponse struct {
	ID          int64               `json:"id"`
	SchemaKey   string              `json:"schema_key"`
	SchemaName  string              `json:"schema_name"`
	Version     string              `json:"version"`
	Description *string             `json:"description"`
	IsLatest    *bool               `json:"is_latest"`
	Lifecycle   *string             `json:"lifecycle"`
	Parameters  []SchemaParamDetail `json:"parameters"`
	CreatedAt   *time.Time          `json:"created_at"`
	UpdatedAt   *time.Time          `json:"updated_at"`
}

// SchemaParamDetail is one parameter entry inside a SchemaResponse.
type SchemaParamDetail struct {
	ParameterKey string `json:"parameter_key"`
	IsRequired   bool   `json:"is_required"`
}

// PublishResponse wraps SchemaResponse with version bump metadata.
type PublishResponse struct {
	SchemaResponse
	// VersionBump is "major", "minor", or "patch" — describes the nature of
	// the change relative to the previous active version.
	VersionBump string `json:"version_bump"`
	// PreviousVersion is the version that was superseded. Empty on first publish.
	PreviousVersion string `json:"previous_version,omitempty"`
}
