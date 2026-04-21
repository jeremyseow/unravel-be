package parameter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/usecase"
)

// ParameterHandler
type ParameterHandler struct {
	ParameterService usecase.ParameterService
}

// NewParameterHandler creates a new parameter handler
func NewParameterHandler(parameterService usecase.ParameterService) *ParameterHandler {
	return &ParameterHandler{
		ParameterService: parameterService,
	}
}

func (p *ParameterHandler) GetParameters(c *gin.Context) {
	parameters, err := p.ParameterService.GetParameters(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": parameters})
}
