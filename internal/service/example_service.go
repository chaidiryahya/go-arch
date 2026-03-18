package service

import (
	"context"
	"fmt"

	"go-arch/internal/entity"
	"go-arch/internal/repository"

	"github.com/rs/zerolog"
)

type ExampleService interface {
	GetByID(ctx context.Context, id int64) (*entity.Example, error)
}

type exampleService struct {
	repo   repository.ExampleRepository
	logger zerolog.Logger
}

func NewExampleService(repo repository.ExampleRepository, logger zerolog.Logger) ExampleService {
	return &exampleService{
		repo:   repo,
		logger: logger,
	}
}

func (s *exampleService) GetByID(ctx context.Context, id int64) (*entity.Example, error) {
	example, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().Str("layer", "service").Err(err).Int64("id", id).Msg("failed to get example")
		return nil, fmt.Errorf("getting example: %w", err)
	}
	return example, nil
}
