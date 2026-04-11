package parameter

import "fmt"

// ValidDataTypes lists all Protobuf-compatible data types accepted by the parameter catalog.
var ValidDataTypes = map[string]bool{
	"string":          true,
	"int32":           true,
	"int64":           true,
	"float":           true,
	"double":          true,
	"bool":            true,
	"bytes":           true,
	"repeated_string": true,
	"repeated_int32":  true,
	"repeated_int64":  true,
	"repeated_float":  true,
	"repeated_double": true,
	"repeated_bool":   true,
}

func validateDataType(dt string) error {
	if !ValidDataTypes[dt] {
		return fmt.Errorf(
			"invalid data_type %q: must be one of string, int32, int64, float, double, bool, bytes, repeated_string, repeated_int32, repeated_int64, repeated_float, repeated_double, repeated_bool",
			dt,
		)
	}
	return nil
}

type CreateParameterRequest struct {
	ParameterKey  string  `json:"parameter_key" binding:"required,max=32"`
	ParameterName string  `json:"parameter_name" binding:"required,max=32"`
	DataType      string  `json:"data_type" binding:"required"`
	Description   *string `json:"description"`
	SampleValues  *string `json:"sample_values"`
}

type UpdateParameterRequest struct {
	ParameterName string  `json:"parameter_name" binding:"required,max=32"`
	DataType      string  `json:"data_type" binding:"required"`
	Description   *string `json:"description"`
	SampleValues  *string `json:"sample_values"`
}
