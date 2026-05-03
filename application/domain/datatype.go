package domain

const (
	DataTypeString       = "string"
	DataTypeInteger      = "integer"
	DataTypeFloat        = "float"
	DataTypeBoolean      = "boolean"
	DataTypeTimestamp    = "timestamp"
	DataTypeBytes        = "bytes"
	DataTypeStringArray  = "string_array"
	DataTypeIntegerArray = "integer_array"
	DataTypeFloatArray   = "float_array"
	DataTypeBooleanArray = "boolean_array"
)

var validDataTypes = map[string]struct{}{
	DataTypeString:       {},
	DataTypeInteger:      {},
	DataTypeFloat:        {},
	DataTypeBoolean:      {},
	DataTypeTimestamp:    {},
	DataTypeBytes:        {},
	DataTypeStringArray:  {},
	DataTypeIntegerArray: {},
	DataTypeFloatArray:   {},
	DataTypeBooleanArray: {},
}

func IsValidDataType(dt string) bool {
	_, ok := validDataTypes[dt]
	return ok
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
	DataTypeStringArray:  "repeated string",
	DataTypeIntegerArray: "repeated int64",
	DataTypeFloatArray:   "repeated double",
	DataTypeBooleanArray: "repeated bool",
}
