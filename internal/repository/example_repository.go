package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go-arch/internal/entity"
)

type ExampleRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Example, error)
}

type exampleRepository struct {
	db *sql.DB
}

func NewExampleRepository(db *sql.DB) ExampleRepository {
	return &exampleRepository{db: db}
}

func (r *exampleRepository) GetByID(ctx context.Context, id int64) (*entity.Example, error) {
	var e entity.Example
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, created_at, updated_at FROM examples WHERE id = ?", id,
	).Scan(&e.ID, &e.Name, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("querying example by id: %w", err)
	}
	return &e, nil
}
