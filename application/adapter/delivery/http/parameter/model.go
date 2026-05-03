package parameter

type ParameterRequest struct {
	ParameterKey  string  `json:"parameter_key" binding:"required"`
	ParameterName string  `json:"parameter_name" binding:"required"`
	DataType      string  `json:"data_type" binding:"required"`
	Description   *string `json:"description"`
	SampleValues  *string `json:"sample_values"`
}
