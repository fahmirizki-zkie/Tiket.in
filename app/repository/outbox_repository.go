package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiket-in-backend/app/model"
)

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

// CreateWithTx menyimpan event outbox dalam transaksi yang sama
func (r *OutboxRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, eventType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}

	query := `
		INSERT INTO outboxes (event_type, payload, status, retry_count, created_at, updated_at)
		VALUES ($1, $2, 'pending', 0, now(), now())`

	_, err = tx.Exec(ctx, query, eventType, raw)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return nil
}

// FetchPending mengambil daftar outbox pending untuk diproses worker
func (r *OutboxRepository) FetchPending(ctx context.Context, limit int) ([]model.Outbox, error) {
	if r.pool == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, event_type, payload, status, retry_count, created_at, updated_at
		FROM outboxes
		WHERE status = 'pending' AND retry_count < 5
		ORDER BY created_at ASC
		LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch pending outbox: %w", err)
	}
	defer rows.Close()

	var list []model.Outbox
	for rows.Next() {
		var o model.Outbox
		if err := rows.Scan(&o.ID, &o.EventType, &o.Payload, &o.Status, &o.RetryCount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan outbox: %w", err)
		}
		list = append(list, o)
	}
	return list, nil
}

// MarkCompleted mengubah status outbox menjadi completed
func (r *OutboxRepository) MarkCompleted(ctx context.Context, id int64) error {
	if r.pool == nil {
		return nil
	}
	query := `
		UPDATE outboxes
		SET status = 'completed', updated_at = now()
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// IncrementRetry menambah jumlah percobaan gagal
func (r *OutboxRepository) IncrementRetry(ctx context.Context, id int64) error {
	if r.pool == nil {
		return nil
	}
	query := `
		UPDATE outboxes
		SET retry_count = retry_count + 1,
		    status = CASE WHEN retry_count + 1 >= 5 THEN 'failed' ELSE 'pending' END,
		    updated_at = now()
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
