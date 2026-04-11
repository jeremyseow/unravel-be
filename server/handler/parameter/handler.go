package parameter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	"github.com/jeremyseow/unravel-be/storage/parameter"
)

type ParameterHandler struct {
	ParameterStorage parameter.Storage
}

func NewParameterHandler(parameterStorage parameter.Storage) *ParameterHandler {
	return &ParameterHandler{
		ParameterStorage: parameterStorage,
	}
}

func (h *ParameterHandler) GetParameters(c *gin.Context) {
	parameters, err := h.ParameterStorage.GetParameters(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": parameters})
}

func (h *ParameterHandler) CreateParameter(c *gin.Context) {
	var req CreateParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateDataType(req.DataType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	param := model.EntityParameters{
		ParameterKey:  req.ParameterKey,
		ParameterName: req.ParameterName,
		DataType:      req.DataType,
		Description:   req.Description,
		SampleValues:  req.SampleValues,
	}

	created, err := h.ParameterStorage.CreateParameter(c, param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *ParameterHandler) UpdateParameter(c *gin.Context) {
	key := c.Param("key")

	var req UpdateParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateDataType(req.DataType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	param := model.EntityParameters{
		ParameterName: req.ParameterName,
		DataType:      req.DataType,
		Description:   req.Description,
		SampleValues:  req.SampleValues,
	}

	updated, err := h.ParameterStorage.UpdateParameter(c, key, param)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *ParameterHandler) DeleteParameter(c *gin.Context) {
	key := c.Param("key")

	if err := h.ParameterStorage.DeleteParameter(c, key); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
