package handler

import (
	"github.com/jeremyseow/unravel-be/config"
	"github.com/jeremyseow/unravel-be/server/handler/parameter"
	"github.com/jeremyseow/unravel-be/server/handler/schema"
	"github.com/jeremyseow/unravel-be/storage"
)

type AllHandlers struct {
	SchemaHandler    *schema.SchemaHandler
	ParameterHandler *parameter.ParameterHandler
}

func NewAllHandlers(cfg *config.Config, allStorages *storage.AllStorages) *AllHandlers {
	return &AllHandlers{
		SchemaHandler:    schema.NewSchemaHandler(allStorages.SchemaStorage),
		ParameterHandler: parameter.NewParameterHandler(allStorages.ParameterStorage),
	}
}
