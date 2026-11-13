package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiket-in-backend/app/model"
)

var pgxNoRows = pgx.ErrNoRows

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) Create(ctx context.Context, cat *model.Category) error {
	query := `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query, cat.Name).Scan(&cat.ID, &cat.CreatedAt)
}

func (r *CategoryRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Category, int, error) {
	countQuery := `SELECT COUNT(*) FROM categories WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, q.Search).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit

	orderCol := "id"
	if q.Sort == "name" {
		orderCol = "name"
	}
	orderDir := "ASC"
	if q.Order == "desc" {
		orderDir = "DESC"
	}

	dataQuery := `
		SELECT id, name, created_at
		FROM categories
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
		ORDER BY ` + orderCol + ` ` + orderDir + `
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, dataQuery, q.Search, q.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		categories = append(categories, c)
	}

	return categories, total, rows.Err()
}

func (r *CategoryRepository) FindByID(ctx context.Context, id int) (*model.Category, error) {
	query := `SELECT id, name, created_at FROM categories WHERE id = $1`
	var c model.Category
	err := r.pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) Update(ctx context.Context, id int, name string) (*model.Category, error) {
	query := `
		UPDATE categories SET name = $1
		WHERE id = $2
		RETURNING id, name, created_at`

	var c model.Category
	err := r.pool.QueryRow(ctx, query, name, id).Scan(&c.ID, &c.Name, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgxNoRows
	}
	return nil
}

func (r *CategoryRepository) ExistsByName(ctx context.Context, name string, excludeID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND id != $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, name, excludeID).Scan(&exists)
	return exists, err
}
