package http

import (
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/parameter"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/schema"
	"github.com/jeremyseow/unravel-be/application/usecase"
	"github.com/jeremyseow/unravel-be/config"
)

type AllHandlers struct {
	SchemaHandler    *schema.SchemaHandler
	ParameterHandler *parameter.ParameterHandler
}

func NewAllHandlers(cfg *config.Config, services *usecase.AllServices) *AllHandlers {
	return &AllHandlers{
		SchemaHandler: schema.NewSchemaHandler(
			services.SchemaService,
		),
		ParameterHandler: parameter.NewParameterHandler(
			services.ParameterService,
		),
	}
}
