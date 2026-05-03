package usecase

import (
	"context"
	"fmt"

	"github.com/jeremyseow/unravel-be/application/domain"
)

type SchemaService interface {
	CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error)
	ListSchemas(ctx context.Context, filter domain.ListSchemasFilter) ([]domain.Schema, error)
	GetSchemas(ctx context.Context, key string) ([]domain.Schema, error)
	GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error)
	PublishSchema(ctx context.Context, key string) (domain.Schema, error)
	DeprecateSchema(ctx context.Context, key, version string) (domain.Schema, error)
}

type SchemaRepository interface {
	CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error)
	ListSchemas(ctx context.Context, filter domain.ListSchemasFilter) ([]domain.Schema, error)
	GetSchemas(ctx context.Context, key string) ([]domain.Schema, error)
	GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error)
	GetParametersByKeys(ctx context.Context, keys []string) ([]domain.Parameter, error)
	GetLatestActiveSchema(ctx context.Context, key string) (*domain.Schema, error)
	PublishSchema(ctx context.Context, key, newVersion string) (domain.Schema, error)
	DeprecateSchema(ctx context.Context, key, version string) (domain.Schema, error)
}

type schemaService struct {
	repo SchemaRepository
}

func NewSchemaService(repo SchemaRepository) SchemaService {
	return &schemaService{repo: repo}
}

func (s *schemaService) CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error) {
	if len(schema.Parameters) > 0 {
		keys := make([]string, len(schema.Parameters))
		for i, p := range schema.Parameters {
			keys[i] = p.ParameterKey
		}
		found, err := s.repo.GetParametersByKeys(ctx, keys)
		if err != nil {
			return domain.Schema{}, fmt.Errorf("validating parameters: %w", err)
		}
		if len(found) != len(keys) {
			return domain.Schema{}, fmt.Errorf("one or more parameter keys do not exist in the catalog: %w", domain.ErrParameterKeysNotFound)
		}
	}

	schema.SchemaVersion = domain.DraftVersion

	return s.repo.CreateSchema(ctx, schema)
}

func (s *schemaService) ListSchemas(ctx context.Context, filter domain.ListSchemasFilter) ([]domain.Schema, error) {
	return s.repo.ListSchemas(ctx, filter)
}

func (s *schemaService) GetSchemas(ctx context.Context, key string) ([]domain.Schema, error) {
	schemas, err := s.repo.GetSchemas(ctx, key)
	if err != nil {
		return nil, err
	}
	return annotateChanges(schemas), nil
}

func (s *schemaService) GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error) {
	return s.repo.GetSchemaVersion(ctx, key, version)
}

func (s *schemaService) PublishSchema(ctx context.Context, key string) (domain.Schema, error) {
	draft, err := s.repo.GetSchemaVersion(ctx, key, domain.DraftVersion)
	if err != nil {
		return domain.Schema{}, err
	}

	latest, err := s.repo.GetLatestActiveSchema(ctx, key)
	if err != nil {
		return domain.Schema{}, err
	}

	newVersion, err := computeNextVersion(latest, draft)
	if err != nil {
		return domain.Schema{}, err
	}

	return s.repo.PublishSchema(ctx, key, newVersion)
}

func (s *schemaService) DeprecateSchema(ctx context.Context, key, version string) (domain.Schema, error) {
	return s.repo.DeprecateSchema(ctx, key, version)
}
