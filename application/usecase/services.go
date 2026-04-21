package usecase

import (
	"github.com/jeremyseow/unravel-be/application/adapter/persistence/postgres"
)

type AllServices struct {
	SchemaService    SchemaService
	ParameterService ParameterService
}

func NewAllServices(storages *postgres.AllStorages) *AllServices {
	return &AllServices{
		SchemaService:    NewSchemaService(storages.SchemaStorage),
		ParameterService: NewParameterService(storages.ParameterStorage),
	}
}
