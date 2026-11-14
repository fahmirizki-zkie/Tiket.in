package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiket-in-backend/app/model"
)

type TicketTypeRepository struct {
	pool *pgxpool.Pool
}

func NewTicketTypeRepository(pool *pgxpool.Pool) *TicketTypeRepository {
	return &TicketTypeRepository{pool: pool}
}

func (r *TicketTypeRepository) Create(ctx context.Context, ticket *model.TicketType) error {
	query := `
		INSERT INTO ticket_types (event_id, name, price, quota, sold, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		ticket.EventID,
		ticket.Name,
		ticket.Price,
		ticket.Quota,
		ticket.Sold,
	).Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)
}

func (r *TicketTypeRepository) FindByEventID(ctx context.Context, eventID int64) ([]model.TicketType, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold, created_at, updated_at
		FROM ticket_types
		WHERE event_id = $1
		ORDER BY price ASC, id ASC`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("FindByEventID ticket_types: %w", err)
	}
	defer rows.Close()

	var tickets []model.TicketType
	for rows.Next() {
		var t model.TicketType
		if err := rows.Scan(
			&t.ID,
			&t.EventID,
			&t.Name,
			&t.Price,
			&t.Quota,
			&t.Sold,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ticket_type row: %w", err)
		}
		t.AvailableSeats = t.Quota - t.Sold
		tickets = append(tickets, t)
	}

	return tickets, nil
}

func (r *TicketTypeRepository) FindByID(ctx context.Context, id int64) (*model.TicketType, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold, created_at, updated_at
		FROM ticket_types
		WHERE id = $1`

	var t model.TicketType
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.EventID,
		&t.Name,
		&t.Price,
		&t.Quota,
		&t.Sold,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByID ticket_type: %w", err)
	}

	t.AvailableSeats = t.Quota - t.Sold
	return &t, nil
}

func (r *TicketTypeRepository) FindByNameAndEventID(ctx context.Context, eventID int64, name string) (*model.TicketType, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold, created_at, updated_at
		FROM ticket_types
		WHERE event_id = $1 AND LOWER(name) = LOWER($2)`

	var t model.TicketType
	err := r.pool.QueryRow(ctx, query, eventID, name).Scan(
		&t.ID,
		&t.EventID,
		&t.Name,
		&t.Price,
		&t.Quota,
		&t.Sold,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByNameAndEventID ticket_type: %w", err)
	}

	t.AvailableSeats = t.Quota - t.Sold
	return &t, nil
}

// FindByIDForUpdate mengunci baris jenis tiket (pessimistic lock) dalam transaksi
func (r *TicketTypeRepository) FindByIDForUpdate(ctx context.Context, tx DBTX, id int64) (*model.TicketType, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold, created_at, updated_at
		FROM ticket_types
		WHERE id = $1
		FOR UPDATE`

	var t model.TicketType
	err := tx.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.EventID,
		&t.Name,
		&t.Price,
		&t.Quota,
		&t.Sold,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByIDForUpdate ticket_type: %w", err)
	}

	t.AvailableSeats = t.Quota - t.Sold
	return &t, nil
}

// IncrementSold menambahkan counter tiket terjual secara atomik dalam transaksi
func (r *TicketTypeRepository) IncrementSold(ctx context.Context, tx DBTX, id int64, count int) error {
	query := `
		UPDATE ticket_types
		SET sold = sold + $1,
		    updated_at = now()
		WHERE id = $2`

	cmd, err := tx.Exec(ctx, query, count, id)
	if err != nil {
		return fmt.Errorf("IncrementSold: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("tiket tidak ditemukan untuk diupdate")
	}
	return nil
}
