package parameter

import (
	// dot import so go code would resemble as much as native SQL
	// dot import is not mandatory
	"context"
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jeremyseow/unravel-be/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/.gen/unravel-db/public/table"
)

type Storage interface {
	GetParameters(ctx context.Context) ([]model.EntityParameters, error)
}

type StorageImpl struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) Storage {
	return &StorageImpl{
		db: db,
	}
}

func (s *StorageImpl) GetParameters(_ context.Context) ([]model.EntityParameters, error) {
	stmt := SELECT(
		EntityParameters.ParameterName,
		EntityParameters.DataType,
	).FROM(
		EntityParameters,
	)

	var parameters []model.EntityParameters
	err := stmt.Query(s.db, &parameters)

	return parameters, err
}
