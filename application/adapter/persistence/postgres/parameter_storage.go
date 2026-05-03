package postgres

import (
	"context"
	"database/sql"
	"fmt"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jeremyseow/unravel-be/application/ctxkey"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

type ParameterStorage struct {
	db *sql.DB
}

func NewParameterStorage(db *sql.DB) *ParameterStorage {
	return &ParameterStorage{db: db}
}

func (s *ParameterStorage) GetParameters(ctx context.Context) ([]domain.Parameter, error) {
	tenantID := ctxkey.TenantID(ctx)
	stmt := SELECT(EntityParameters.AllColumns).
		FROM(EntityParameters).
		WHERE(EntityParameters.TenantID.EQ(uuidStr(tenantID)))

	var rows []model.EntityParameters
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, err
	}

	params := make([]domain.Parameter, len(rows))
	for i, r := range rows {
		params[i] = toDomainParameter(r)
	}
	return params, nil
}

func (s *ParameterStorage) CreateParameter(ctx context.Context, param domain.Parameter) (domain.Parameter, error) {
	tenantID := ctxkey.TenantID(ctx)
	stmt := EntityParameters.INSERT(
		EntityParameters.TenantID,
		EntityParameters.ParameterKey,
		EntityParameters.ParameterName,
		EntityParameters.DataType,
		EntityParameters.Description,
		EntityParameters.SampleValues,
	).VALUES(
		tenantID,
		param.ParameterKey,
		param.ParameterName,
		param.DataType,
		param.Description,
		param.SampleValues,
	).RETURNING(EntityParameters.AllColumns)

	var row model.EntityParameters
	if err := stmt.QueryContext(ctx, s.db, &row); err != nil {
		return domain.Parameter{}, err
	}
	return toDomainParameter(row), nil
}

func (s *ParameterStorage) UpdateParameter(ctx context.Context, key string, param domain.Parameter) (domain.Parameter, error) {
	tenantID := ctxkey.TenantID(ctx)
	stmt := EntityParameters.UPDATE(
		EntityParameters.ParameterName,
		EntityParameters.DataType,
		EntityParameters.Description,
		EntityParameters.SampleValues,
	).SET(
		param.ParameterName,
		param.DataType,
		param.Description,
		param.SampleValues,
	).WHERE(
		EntityParameters.TenantID.EQ(uuidStr(tenantID)).
			AND(EntityParameters.ParameterKey.EQ(String(key))),
	).RETURNING(EntityParameters.AllColumns)

	var row model.EntityParameters
	if err := stmt.QueryContext(ctx, s.db, &row); err != nil {
		return domain.Parameter{}, err
	}
	if row.ParameterKey == "" {
		return domain.Parameter{}, fmt.Errorf("parameter not found: %s", key)
	}
	return toDomainParameter(row), nil
}

func (s *ParameterStorage) DeleteParameter(ctx context.Context, key string) error {
	tenantID := ctxkey.TenantID(ctx)
	stmt := EntityParameters.DELETE().WHERE(
		EntityParameters.TenantID.EQ(uuidStr(tenantID)).
			AND(EntityParameters.ParameterKey.EQ(String(key))),
	)
	res, err := stmt.ExecContext(ctx, s.db)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("parameter not found: %s", key)
	}
	return nil
}

func toDomainParameter(r model.EntityParameters) domain.Parameter {
	return domain.Parameter{
		ParameterKey:  r.ParameterKey,
		ParameterName: r.ParameterName,
		DataType:      r.DataType,
		Description:   r.Description,
		SampleValues:  r.SampleValues,
	}
}
