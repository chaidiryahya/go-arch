package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go-arch/internal/entity"
	infralog "go-arch/internal/infra/log"

	"github.com/rs/zerolog"
)

type ExampleRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Example, error)
}

type exampleRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewExampleRepository(db *sql.DB, logger zerolog.Logger) ExampleRepository {
	return &exampleRepository{db: db, logger: logger}
}

func (r *exampleRepository) GetByID(ctx context.Context, id int64) (*entity.Example, error) {
	var e entity.Example
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, created_at, updated_at FROM examples WHERE id = ?", id,
	).Scan(&e.ID, &e.Name, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		infralog.ContextLogger(ctx, r.logger).Error().
			Str("layer", "repository").
			Err(err).
			Int64("id", id).
			Msg("failed to query example")
		return nil, fmt.Errorf("querying example by id: %w", err)
	}
	return &e, nil
}
