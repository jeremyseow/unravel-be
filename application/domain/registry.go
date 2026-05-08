package domain

// EnrichedSchemaParameter merges a schema parameter mapping (key + is_required)
// with its full catalog entry (data_type, description, etc.) for proto generation.
type EnrichedSchemaParameter struct {
	ParameterKey  string
	ParameterName string
	IsRequired    bool
	DataType      string
	Description   *string
	SampleValues  *string
}
