package parameter

import (
	"context"
	"database/sql"
	"fmt"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

// TODO: Replace with tenant ID from auth context (Phase 6)
var defaultTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type Storage interface {
	GetParameters(ctx context.Context) ([]model.EntityParameters, error)
	CreateParameter(ctx context.Context, param model.EntityParameters) (*model.EntityParameters, error)
	UpdateParameter(ctx context.Context, key string, param model.EntityParameters) (*model.EntityParameters, error)
	DeleteParameter(ctx context.Context, key string) error
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
		EntityParameters.AllColumns,
	).FROM(
		EntityParameters,
	).WHERE(
		EntityParameters.TenantID.EQ(String(defaultTenantID.String())),
	)

	var parameters []model.EntityParameters
	err := stmt.Query(s.db, &parameters)
	return parameters, err
}

func (s *StorageImpl) CreateParameter(_ context.Context, param model.EntityParameters) (*model.EntityParameters, error) {
	param.TenantID = defaultTenantID

	stmt := EntityParameters.INSERT(
		EntityParameters.TenantID,
		EntityParameters.ParameterKey,
		EntityParameters.ParameterName,
		EntityParameters.DataType,
		EntityParameters.Description,
		EntityParameters.SampleValues,
	).VALUES(
		param.TenantID,
		param.ParameterKey,
		param.ParameterName,
		param.DataType,
		param.Description,
		param.SampleValues,
	).RETURNING(EntityParameters.AllColumns)

	var created model.EntityParameters
	err := stmt.Query(s.db, &created)
	return &created, err
}

func (s *StorageImpl) UpdateParameter(_ context.Context, key string, param model.EntityParameters) (*model.EntityParameters, error) {
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
		EntityParameters.TenantID.EQ(String(defaultTenantID.String())).
			AND(EntityParameters.ParameterKey.EQ(String(key))),
	).RETURNING(EntityParameters.AllColumns)

	var updated model.EntityParameters
	err := stmt.Query(s.db, &updated)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("parameter not found: %s", key)
	}
	return &updated, err
}

func (s *StorageImpl) DeleteParameter(_ context.Context, key string) error {
	// Block deletion if the parameter is referenced by any schema — removing it
	// would leave dangling references in both draft and active schemas.
	refStmt := SELECT(EntitySchemasParametersMappings.ParameterKey).
		FROM(EntitySchemasParametersMappings).
		WHERE(
			EntitySchemasParametersMappings.TenantID.EQ(String(defaultTenantID.String())).
				AND(EntitySchemasParametersMappings.ParameterKey.EQ(String(key))),
		).
		LIMIT(1)

	var refs []model.EntitySchemasParametersMappings
	if err := refStmt.Query(s.db, &refs); err != nil {
		return err
	}
	if len(refs) > 0 {
		return fmt.Errorf("parameter_referenced: parameter %q is used by one or more schemas and cannot be deleted", key)
	}

	stmt := EntityParameters.DELETE().WHERE(
		EntityParameters.TenantID.EQ(String(defaultTenantID.String())).
			AND(EntityParameters.ParameterKey.EQ(String(key))),
	)

	result, err := stmt.Exec(s.db)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("parameter not found: %s", key)
	}
	return nil
}
