package parameter

type Parameter struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

type ParameterRequest struct {
	Name     string `json:"name" binding:"required"`
	DataType string `json:"data_type" binding:"required"`
}
