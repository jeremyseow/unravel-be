package parameter

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
