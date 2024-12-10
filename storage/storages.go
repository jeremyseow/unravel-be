package storage

import (
	"database/sql"

	"github.com/jeremyseow/unravel-be/config"
	"github.com/jeremyseow/unravel-be/storage/parameter"
	_ "github.com/lib/pq"
)

type AllStorages struct {
	ParameterStorage parameter.Storage
	config           *config.Config
}

func NewAllStorages(cfg *config.Config) (*AllStorages, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &AllStorages{
		ParameterStorage: parameter.NewStorage(db),
		config:           cfg,
	}, nil
}
