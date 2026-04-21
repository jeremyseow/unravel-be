package schema

type ParameterRef struct {
	ParameterKey string `json:"parameter_key" binding:"required"`
	IsRequired   bool   `json:"is_required"`
}

type SchemaRequest struct {
	SchemaKey   string         `json:"schema_key" binding:"required"`
	SchemaName  string         `json:"schema_name" binding:"required"`
	Description *string        `json:"description"`
	Parameters  []ParameterRef `json:"parameters" binding:"required"`
}
