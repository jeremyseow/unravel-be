package storage

import (
	"database/sql"

	"github.com/jeremyseow/unravel-be/config"
	"github.com/jeremyseow/unravel-be/storage/parameter"
	"github.com/jeremyseow/unravel-be/storage/schema"
	_ "github.com/lib/pq"
)

type AllStorages struct {
	ParameterStorage parameter.Storage
	SchemaStorage    schema.Storage
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

	allStorages.ParameterStorage = parameter.NewStorage(db)
	allStorages.SchemaStorage = schema.NewStorage(db)
	return allStorages, nil
}
