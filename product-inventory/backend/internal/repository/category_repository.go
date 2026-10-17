package repository

import (
	"context"

	"inventory/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cats := []model.Category{}
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *CategoryRepository) Create(ctx context.Context, name string) (*model.Category, error) {
	c := model.Category{Name: name}
	err := r.db.QueryRow(ctx, `INSERT INTO categories (name) VALUES ($1) RETURNING id`, name).Scan(&c.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}
