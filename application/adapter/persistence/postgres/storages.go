package postgres

import (
	"database/sql"

	"github.com/jeremyseow/unravel-be/config"
	_ "github.com/lib/pq"
)

type AllStorages struct {
	ParameterStorage *ParameterStorage
	SchemaStorage    *SchemaStorage
	config           *config.Config
}

func NewAllStorages(cfg *config.Config) (*AllStorages, error) {
	allStorages := &AllStorages{
		config: cfg,
	}

	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return allStorages, err
	}

	if err := db.Ping(); err != nil {
		return allStorages, err
	}

	allStorages.ParameterStorage = NewParameterStorage(db)
	allStorages.SchemaStorage = NewSchemaStorage(db)
	return allStorages, nil
}
