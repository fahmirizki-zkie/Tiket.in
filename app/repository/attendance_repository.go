package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiket-in-backend/app/model"
)

var ErrDuplicateAttendance = errors.New("tiket sudah digunakan untuk check-in")

type AttendanceRepository struct {
	pool *pgxpool.Pool
}

func NewAttendanceRepository(pool *pgxpool.Pool) *AttendanceRepository {
	return  &AttendanceRepository{pool: pool}
}

func (r *AttendanceRepository) Create(ctx context.Context, tx DBTX, att *model.Attendance) error {
	query := `
		INSERT INTO attendances (registration_id, checked_in_by, checked_in_at)
		VALUES ($1, $2, now())
		RETURNING id, checked_in_at`
	
	err := tx.QueryRow(ctx, query, att.RegistrationID, att.CheckedInBy).
		Scan(&att.ID, &att.CheckedInAt)
	if err !=nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateAttendance
		}
		return fmt.Errorf("Create attendance: %w", err)
	}
	return nil
}
