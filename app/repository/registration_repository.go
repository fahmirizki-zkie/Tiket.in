package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiket-in-backend/app/model"
)

// Error ini dikembalikan repository agar service bisa memetakannya ke
// error bisnis (repository tidak boleh meng-import package service).
var (
	ErrDuplicateRegistration = errors.New("peserta sudah terdaftar pada event ini")
	ErrDuplicateTicketCode   = errors.New("kode tiket bentrok")
)

type RegistrationRepository struct {
	pool *pgxpool.Pool
}

func NewRegistrationRepository(pool *pgxpool.Pool) *RegistrationRepository {
	return &RegistrationRepository{pool: pool}
}

// Create menyimpan pendaftaran baru di dalam transaksi.
func (r *RegistrationRepository) Create(ctx context.Context, tx DBTX, reg *model.Registration) error {
	query := `
		INSERT INTO registrations (user_id, event_id, ticket_type_id, ticket_code, status, registered_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())
		RETURNING id, registered_at, updated_at`

	err := tx.QueryRow(ctx, query,
		reg.UserID,
		reg.EventID,
		reg.TicketTypeID,
		reg.TicketCode,
		reg.Status,
	).Scan(&reg.ID, &reg.RegisteredAt, &reg.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "uq_registrations_user_event":
				return ErrDuplicateRegistration
			case "registrations_ticket_code_key":
				return ErrDuplicateTicketCode
			}
		}
		return fmt.Errorf("Create registration: %w", err)
	}
	return nil
}

// ExistsByUserAndEvent mengecek apakah peserta sudah terdaftar pada event.
func (r *RegistrationRepository) ExistsByUserAndEvent(ctx context.Context, tx DBTX, userID, eventID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM registrations
			WHERE user_id = $1 AND event_id = $2
		)`

	var exists bool
	if err := tx.QueryRow(ctx, query, userID, eventID).Scan(&exists); err != nil {
		return false, fmt.Errorf("ExistsByUserAndEvent: %w", err)
	}
	return exists, nil
}

// FindByUserID mengambil tiket milik peserta beserta info event dan jenis tiket.
// status boleh kosong untuk menampilkan semua status.
func (r *RegistrationRepository) FindByUserID(ctx context.Context, userID int64, status string, limit, offset int) ([]model.Registration, error) {
	query := `
		SELECT r.id, r.user_id, r.event_id, r.ticket_type_id, r.ticket_code, r.status,
		       r.registered_at, r.updated_at,
		       e.title, e.location, e.start_at, tt.name
		FROM registrations r
		JOIN events e ON e.id = r.event_id
		JOIN ticket_types tt ON tt.id = r.ticket_type_id
		WHERE r.user_id = $1
		  AND ($2::text = '' OR r.status = $2)
		ORDER BY r.registered_at DESC, r.id DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, userID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("FindByUserID registrations: %w", err)
	}
	defer rows.Close()

	registrations := []model.Registration{} // slice kosong supaya JSON-nya [] bukan null
	for rows.Next() {
		var reg model.Registration
		if err := rows.Scan(
			&reg.ID,
			&reg.UserID,
			&reg.EventID,
			&reg.TicketTypeID,
			&reg.TicketCode,
			&reg.Status,
			&reg.RegisteredAt,
			&reg.UpdatedAt,
			&reg.EventTitle,
			&reg.EventLocation,
			&reg.EventStartAt,
			&reg.TicketTypeName,
		); err != nil {
			return nil, fmt.Errorf("scan registration row: %w", err)
		}
		registrations = append(registrations, reg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate registration rows: %w", err)
	}

	return registrations, nil
}

// FindByTicketCodeForUpdate dipakai saat check-in: baris dikunci supaya
// dua panitia tidak bisa check-in tiket yang sama bersamaan.
func (r *RegistrationRepository) FindByTicketCodeForUpdate(ctx context.Context, tx DBTX, code string) (*model.Registration, error) {
	query := `
		SELECT id, user_id, event_id, ticket_type_id, ticket_code, status, registered_at, updated_at
		FROM registrations
		WHERE ticket_code = $1
		FOR UPDATE`

	var reg model.Registration
	err := tx.QueryRow(ctx, query, code).Scan(
		&reg.ID,
		&reg.UserID,
		&reg.EventID,
		&reg.TicketTypeID,
		&reg.TicketCode,
		&reg.Status,
		&reg.RegisteredAt,
		&reg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByTicketCodeForUpdate: %w", err)
	}
	return &reg, nil
}

// UpdateStatus mengubah status pendaftaran (registered, cancelled, checked_in).
func (r *RegistrationRepository) UpdateStatus(ctx context.Context, tx DBTX, id int64, status string) error {
	query := `
		UPDATE registrations
		SET status = $1,
		    updated_at = now()
		WHERE id = $2`

	cmd, err := tx.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("UpdateStatus registration: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("registrasi tidak ditemukan untuk diupdate")
	}
	return nil
}