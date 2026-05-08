package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/jeremyseow/unravel-be/application/domain"
)

type RegistryService interface {
	GetLatestSchema(ctx context.Context, key string) (domain.Schema, []domain.EnrichedSchemaParameter, error)
	GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, []domain.EnrichedSchemaParameter, error)
}

// RegistryRepository is satisfied by SchemaStorage with no extra methods needed.
type RegistryRepository interface {
	GetLatestActiveSchema(ctx context.Context, key string) (*domain.Schema, error)
	GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error)
	GetParametersByKeys(ctx context.Context, keys []string) ([]domain.Parameter, error)
}

type registryService struct {
	repo RegistryRepository
}

func NewRegistryService(repo RegistryRepository) RegistryService {
	return &registryService{repo: repo}
}

func (s *registryService) GetLatestSchema(ctx context.Context, key string) (domain.Schema, []domain.EnrichedSchemaParameter, error) {
	schema, err := s.repo.GetLatestActiveSchema(ctx, key)
	if err != nil {
		return domain.Schema{}, nil, err
	}
	if schema == nil {
		return domain.Schema{}, nil, fmt.Errorf("schema %q: %w", key, domain.ErrNotFound)
	}
	enriched, err := s.enrich(ctx, schema.Parameters)
	if err != nil {
		return domain.Schema{}, nil, err
	}
	return *schema, enriched, nil
}

func (s *registryService) GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, []domain.EnrichedSchemaParameter, error) {
	schema, err := s.repo.GetSchemaVersion(ctx, key, version)
	if err != nil {
		return domain.Schema{}, nil, err
	}
	enriched, err := s.enrich(ctx, schema.Parameters)
	if err != nil {
		return domain.Schema{}, nil, err
	}
	return schema, enriched, nil
}

func (s *registryService) enrich(ctx context.Context, schemaParams []domain.SchemaParameter) ([]domain.EnrichedSchemaParameter, error) {
	keys := make([]string, len(schemaParams))
	for i, p := range schemaParams {
		keys[i] = p.ParameterKey
	}

	catalog, err := s.repo.GetParametersByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}

	catalogMap := make(map[string]domain.Parameter, len(catalog))
	for _, p := range catalog {
		catalogMap[p.ParameterKey] = p
	}

	// Sort by key for stable proto field numbers within a version.
	sorted := make([]domain.SchemaParameter, len(schemaParams))
	copy(sorted, schemaParams)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ParameterKey < sorted[j].ParameterKey
	})

	enriched := make([]domain.EnrichedSchemaParameter, len(sorted))
	for i, sp := range sorted {
		cat := catalogMap[sp.ParameterKey]
		enriched[i] = domain.EnrichedSchemaParameter{
			ParameterKey:  sp.ParameterKey,
			ParameterName: cat.ParameterName,
			IsRequired:    sp.IsRequired,
			DataType:      cat.DataType,
			Description:   cat.Description,
			SampleValues:  cat.SampleValues,
		}
	}
	return enriched, nil
}
