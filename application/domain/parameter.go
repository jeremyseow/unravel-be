package domain

type Parameter struct {
	ParameterKey  string  `json:"parameter_key"`
	ParameterName string  `json:"parameter_name"`
	DataType      string  `json:"data_type"`
	Description   *string `json:"description"`
	SampleValues  *string `json:"sample_values"`
}
