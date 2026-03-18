package usecase

import (
	"context"
	"fmt"

	"go-arch/internal/entity"
	"go-arch/internal/service"

	"github.com/rs/zerolog"
)

type ExampleUsecase interface {
	GetByID(ctx context.Context, id int64) (*entity.Example, error)
}

type exampleUsecase struct {
	exampleService service.ExampleService
	logger         zerolog.Logger
}

func NewExampleUsecase(exampleService service.ExampleService, logger zerolog.Logger) ExampleUsecase {
	return &exampleUsecase{
		exampleService: exampleService,
		logger:         logger,
	}
}

func (u *exampleUsecase) GetByID(ctx context.Context, id int64) (*entity.Example, error) {
	example, err := u.exampleService.GetByID(ctx, id)
	if err != nil {
		u.logger.Error().Str("layer", "usecase").Err(err).Int64("id", id).Msg("failed to get example")
		return nil, fmt.Errorf("getting example by id: %w", err)
	}
	return example, nil
}
