package parameter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/apierror"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/application/usecase"
)

type ParameterHandler struct {
	ParameterService usecase.ParameterService
}

func NewParameterHandler(parameterService usecase.ParameterService) *ParameterHandler {
	return &ParameterHandler{ParameterService: parameterService}
}

func (h *ParameterHandler) GetParameters(c *gin.Context) error {
	parameters, err := h.ParameterService.GetParameters(c.Request.Context())
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{"data": parameters})
	return nil
}

func (h *ParameterHandler) CreateParameter(c *gin.Context) error {
	var req ParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return apierror.ValidationError(err)
	}

	param := domain.Parameter{
		ParameterKey:  req.ParameterKey,
		ParameterName: req.ParameterName,
		DataType:      req.DataType,
		Description:   req.Description,
		SampleValues:  req.SampleValues,
	}

	created, err := h.ParameterService.CreateParameter(c.Request.Context(), param)
	if err != nil {
		return err
	}
	c.JSON(http.StatusCreated, created)
	return nil
}

func (h *ParameterHandler) UpdateParameter(c *gin.Context) error {
	key := c.Param("key")

	var req ParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return apierror.ValidationError(err)
	}

	param := domain.Parameter{
		ParameterKey:  key,
		ParameterName: req.ParameterName,
		DataType:      req.DataType,
		Description:   req.Description,
		SampleValues:  req.SampleValues,
	}

	updated, err := h.ParameterService.UpdateParameter(c.Request.Context(), key, param)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, updated)
	return nil
}

func (h *ParameterHandler) DeleteParameter(c *gin.Context) error {
	key := c.Param("key")

	if err := h.ParameterService.DeleteParameter(c.Request.Context(), key); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
