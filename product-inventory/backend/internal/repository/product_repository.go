package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"inventory/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

const selectProduct = `
SELECT p.id, p.name, p.sku, p.category_id, c.name, p.price::float8,
       p.quantity, p.status, p.created_at, p.updated_at
FROM products p JOIN categories c ON c.id = p.category_id`

func scanProduct(row pgx.Row) (*model.Product, error) {
	var p model.Product
	err := row.Scan(&p.ID, &p.Name, &p.SKU, &p.CategoryID, &p.Category, &p.Price,
		&p.Quantity, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

// mapErr translates Postgres constraint violations into domain errors.
func mapErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if strings.Contains(pgErr.ConstraintName, "categories") {
				return ErrDuplicateCategory
			}
			return ErrDuplicateSKU
		case "23503":
			return ErrCategoryNotFound
		case "23514":
			return ErrInsufficientStock
		}
	}
	return err
}

func (r *ProductRepository) List(ctx context.Context, f model.ListFilter) ([]model.Product, error) {
	q := selectProduct + " WHERE 1=1"
	args := []any{}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += fmt.Sprintf(" AND (p.name ILIKE $%d OR p.sku ILIKE $%d)", len(args), len(args))
	}
	if f.CategoryID > 0 {
		args = append(args, f.CategoryID)
		q += fmt.Sprintf(" AND p.category_id = $%d", len(args))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		q += fmt.Sprintf(" AND p.status = $%d", len(args))
	}
	q += " ORDER BY p.id DESC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []model.Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *p)
	}
	return products, rows.Err()
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	return scanProduct(r.db.QueryRow(ctx, selectProduct+" WHERE p.id = $1", id))
}

func (r *ProductRepository) Create(ctx context.Context, in model.ProductInput) (*model.Product, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO products (name, sku, category_id, price, quantity, status)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		in.Name, in.SKU, in.CategoryID, in.Price, in.Quantity, in.Status).Scan(&id)
	if err != nil {
		return nil, mapErr(err)
	}
	return r.GetByID(ctx, id)
}

func (r *ProductRepository) Update(ctx context.Context, id int64, in model.ProductInput) (*model.Product, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE products SET name=$1, sku=$2, category_id=$3, price=$4, quantity=$5,
		        status=$6, updated_at=NOW() WHERE id=$7`,
		in.Name, in.SKU, in.CategoryID, in.Price, in.Quantity, in.Status, id)
	if err != nil {
		return nil, mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetQuantity sets stock to an absolute value.
func (r *ProductRepository) SetQuantity(ctx context.Context, id int64, qty int) (*model.Product, error) {
	return r.updateStock(ctx, `UPDATE products SET quantity=$1, updated_at=NOW() WHERE id=$2`, qty, id)
}

// AdjustQuantity applies a delta atomically; the CHECK constraint rejects negative stock.
func (r *ProductRepository) AdjustQuantity(ctx context.Context, id int64, delta int) (*model.Product, error) {
	return r.updateStock(ctx, `UPDATE products SET quantity=quantity+$1, updated_at=NOW() WHERE id=$2`, delta, id)
}

func (r *ProductRepository) updateStock(ctx context.Context, q string, val int, id int64) (*model.Product, error) {
	tag, err := r.db.Exec(ctx, q, val, id)
	if err != nil {
		return nil, mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}
