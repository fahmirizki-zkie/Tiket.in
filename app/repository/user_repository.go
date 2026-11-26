package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"tiket-in-backend/app/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		user.Name, user.Email, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users WHERE email = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users WHERE id = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindAll mengambil daftar user dengan pencarian (nama/email), filter role, dan pagination.
func (r *UserRepository) FindAll(ctx context.Context, q model.ListQuery, role string) ([]model.User, int, error) {
	const filter = `
		WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
		  AND ($2::text = '' OR role = $2::text)`

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`+filter, q.Search, role).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderCol := "id"
	switch q.Sort {
	case "name", "email", "role", "created_at":
		orderCol = q.Sort
	}
	orderDir := "ASC"
	if q.Order == "desc" {
		orderDir = "DESC"
	}

	offset := (q.Page - 1) * q.Limit

	dataQuery := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users` + filter + `
		ORDER BY ` + orderCol + ` ` + orderDir + `, id ASC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, dataQuery, q.Search, role, q.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, rows.Err()
}

// UpdateRole mengubah role user. Mengembalikan pgx.ErrNoRows bila user tidak ada.
func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role string) (*model.User, error) {
	query := `
		UPDATE users SET role = $1, updated_at = now()
		WHERE id = $2
		RETURNING id, name, email, password_hash, role, created_at`

	var u model.User
	err := r.pool.QueryRow(ctx, query, role, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// CountByRole menghitung jumlah user dengan role tertentu.
func (r *UserRepository) CountByRole(ctx context.Context, role string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = $1`, role).Scan(&n)
	return n, err
}

// UpdateProfile mengubah nama dan email user.
func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, name, email string) (*model.User, error) {
	query := `
		UPDATE users SET name = $1, email = $2, updated_at = now()
		WHERE id = $3
		RETURNING id, name, email, password_hash, role, created_at`

	var u model.User
	err := r.pool.QueryRow(ctx, query, name, email, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdatePassword mengganti hash password user.
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`,
		passwordHash, id,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgxNoRows
	}
	return nil
}

// ExistsByEmail memeriksa apakah email sudah dipakai user lain (excludeID dikecualikan).
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string, excludeID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`,
		email, excludeID,
	).Scan(&exists)
	return exists, err
}
