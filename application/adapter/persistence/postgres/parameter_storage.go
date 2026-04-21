package postgres

import (
	"context"
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

type ParameterStorage struct {
	db *sql.DB
}

func NewParameterStorage(db *sql.DB) *ParameterStorage {
	return &ParameterStorage{
		db: db,
	}
}

func (s *ParameterStorage) GetParameters(_ context.Context) ([]domain.Parameter, error) {
	stmt := SELECT(
		EntityParameters.ParameterKey,
		EntityParameters.DataType,
	).FROM(
		EntityParameters,
	)

	var dbParameters []model.EntityParameters
	err := stmt.Query(s.db, &dbParameters)
	if err != nil {
		return nil, err
	}

	var domainParameters []domain.Parameter
	for _, p := range dbParameters {
		domainParameters = append(domainParameters, domain.Parameter{
			ParameterKey: p.ParameterKey,
			DataType:     p.DataType,
		})
	}

	return domainParameters, nil
}
