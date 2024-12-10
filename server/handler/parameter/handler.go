package parameter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/storage/parameter"
)

// ParameterHandler
type ParameterHandler struct {
	ParameterStorage parameter.Storage
}

// NewParameterHandler creates a new parameter handler
func NewParameterHandler(parameterStorage parameter.Storage) *ParameterHandler {
	return &ParameterHandler{
		ParameterStorage: parameterStorage,
	}
}

func (p *ParameterHandler) GetParameters(c *gin.Context) {
	parameters, err := p.ParameterStorage.GetParameters(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": parameters})
}
