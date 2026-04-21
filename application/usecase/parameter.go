package usecase

import (
	"context"
	"github.com/jeremyseow/unravel-be/application/domain"
)

type ParameterService interface {
	GetParameters(ctx context.Context) ([]domain.Parameter, error)
}

type ParameterRepository interface {
	GetParameters(ctx context.Context) ([]domain.Parameter, error)
}

type parameterService struct {
	repo ParameterRepository
}

func NewParameterService(repo ParameterRepository) ParameterService {
	return &parameterService{
		repo: repo,
	}
}

func (s *parameterService) GetParameters(ctx context.Context) ([]domain.Parameter, error) {
	return s.repo.GetParameters(ctx)
}
