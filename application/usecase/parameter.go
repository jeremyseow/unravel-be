package usecase

import (
	"context"

	"github.com/jeremyseow/unravel-be/application/domain"
)

type ParameterService interface {
	GetParameters(ctx context.Context) ([]domain.Parameter, error)
	CreateParameter(ctx context.Context, param domain.Parameter) (domain.Parameter, error)
	UpdateParameter(ctx context.Context, key string, param domain.Parameter) (domain.Parameter, error)
	DeleteParameter(ctx context.Context, key string) error
}

type ParameterRepository interface {
	GetParameters(ctx context.Context) ([]domain.Parameter, error)
	CreateParameter(ctx context.Context, param domain.Parameter) (domain.Parameter, error)
	UpdateParameter(ctx context.Context, key string, param domain.Parameter) (domain.Parameter, error)
	DeleteParameter(ctx context.Context, key string) error
}

type parameterService struct {
	repo ParameterRepository
}

func NewParameterService(repo ParameterRepository) ParameterService {
	return &parameterService{repo: repo}
}

func (s *parameterService) GetParameters(ctx context.Context) ([]domain.Parameter, error) {
	return s.repo.GetParameters(ctx)
}

func (s *parameterService) CreateParameter(ctx context.Context, param domain.Parameter) (domain.Parameter, error) {
	return s.repo.CreateParameter(ctx, param)
}

func (s *parameterService) UpdateParameter(ctx context.Context, key string, param domain.Parameter) (domain.Parameter, error) {
	return s.repo.UpdateParameter(ctx, key, param)
}

func (s *parameterService) DeleteParameter(ctx context.Context, key string) error {
	return s.repo.DeleteParameter(ctx, key)
}
