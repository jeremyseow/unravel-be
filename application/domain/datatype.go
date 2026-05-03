package domain

const (
	DataTypeString       = "string"
	DataTypeInteger      = "integer"
	DataTypeFloat        = "float"
	DataTypeBoolean      = "boolean"
	DataTypeTimestamp    = "timestamp"
	DataTypeBytes        = "bytes"
	DataTypeJSON         = "json"
	DataTypeStringArray  = "string_array"
	DataTypeIntegerArray = "integer_array"
	DataTypeFloatArray   = "float_array"
	DataTypeBooleanArray = "boolean_array"
)

var EnabledDataTypes = map[string]bool{
	DataTypeString:       true,
	DataTypeInteger:      true,
	DataTypeFloat:        true,
	DataTypeBoolean:      true,
	DataTypeTimestamp:    true,
	DataTypeBytes:        false,
	DataTypeJSON:         true,
	DataTypeStringArray:  true,
	DataTypeIntegerArray: true,
	DataTypeFloatArray:   true,
	DataTypeBooleanArray: true,
}

func IsDataTypeEnabled(dataType string) bool {
	return EnabledDataTypes[dataType]
}

// ProtoType maps a general data type to its Protobuf scalar equivalent.
// timestamp maps to int64 (Unix epoch nanoseconds).
var ProtoType = map[string]string{
	DataTypeString:       "string",
	DataTypeInteger:      "int64",
	DataTypeFloat:        "double",
	DataTypeBoolean:      "bool",
	DataTypeTimestamp:    "int64",
	DataTypeBytes:        "bytes",
	DataTypeJSON:         "google.protobuf.Struct",
	DataTypeStringArray:  "repeated string",
	DataTypeIntegerArray: "repeated int64",
	DataTypeFloatArray:   "repeated double",
	DataTypeBooleanArray: "repeated bool",
}

// ToProtoType converts a DataType to its corresponding Protobuf scalar type.
// It returns the scalar type name (e.g., "string", "int64", "google.protobuf.Struct")
// or an empty string if the data type is not supported or enabled.
func ToProtoType(dataType string) string {
	return ProtoType[dataType]
}
