package parameter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/application/usecase"
)

type ParameterHandler struct {
	ParameterService usecase.ParameterService
}

func NewParameterHandler(parameterService usecase.ParameterService) *ParameterHandler {
	return &ParameterHandler{ParameterService: parameterService}
}

func (h *ParameterHandler) GetParameters(c *gin.Context) {
	parameters, err := h.ParameterService.GetParameters(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": parameters})
}

func (h *ParameterHandler) CreateParameter(c *gin.Context) {
	var req ParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	param := domain.Parameter{
		ParameterKey:  req.ParameterKey,
		ParameterName: req.ParameterName,
		DataType:      req.DataType,
		Description:   req.Description,
		SampleValues:  req.SampleValues,
	}

	created, err := h.ParameterService.CreateParameter(c, param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *ParameterHandler) UpdateParameter(c *gin.Context) {
	key := c.Param("key")

	var req ParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	param := domain.Parameter{
		ParameterKey:  key,
		ParameterName: req.ParameterName,
		DataType:      req.DataType,
		Description:   req.Description,
		SampleValues:  req.SampleValues,
	}

	updated, err := h.ParameterService.UpdateParameter(c, key, param)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *ParameterHandler) DeleteParameter(c *gin.Context) {
	key := c.Param("key")

	if err := h.ParameterService.DeleteParameter(c, key); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
