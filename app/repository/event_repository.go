package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiket-in-backend/app/model"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	query := `
		INSERT INTO events (
			organizer_id, category_id, title, description, location, start_at, end_at, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		RETURNING id, created_at, updated_at`

	if event.Status == "" {
		event.Status = "published"
	}

	return r.pool.QueryRow(ctx, query,
		event.OrganizerID,
		event.CategoryID,
		event.Title,
		event.Description,
		event.Location,
		event.StartAt,
		event.EndAt,
		event.Status,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

func (r *EventRepository) FindByID(ctx context.Context, id int64) (*model.Event, error) {
	query := `
		SELECT 
			e.id, e.organizer_id, e.category_id, e.title, e.description, 
			e.location, e.start_at, e.end_at, e.status, e.created_at, e.updated_at,
			COALESCE(c.name, '') AS category_name,
			COALESCE(u.name, '') AS organizer_name
		FROM events e
		LEFT JOIN categories c ON c.id = e.category_id
		LEFT JOIN users u ON u.id = e.organizer_id
		WHERE e.id = $1`

	var event model.Event
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.OrganizerID,
		&event.CategoryID,
		&event.Title,
		&event.Description,
		&event.Location,
		&event.StartAt,
		&event.EndAt,
		&event.Status,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.CategoryName,
		&event.OrganizerName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByID event: %w", err)
	}

	return &event, nil
}

func (r *EventRepository) FindAll(ctx context.Context, q model.EventListQuery) ([]model.Event, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 10
	}
	offset := (q.Page - 1) * q.Limit

	countQuery := `
		SELECT COUNT(*)
		FROM events e
		WHERE ($1 = '' OR e.title ILIKE '%' || $1 || '%' OR e.description ILIKE '%' || $1 || '%')
		  AND ($2 = 0 OR e.category_id = $2)
		  AND ($3 = '' OR e.status = $3)`

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, q.Search, q.CategoryID, q.Status).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	dataQuery := `
		SELECT 
			e.id, e.organizer_id, e.category_id, e.title, e.description, 
			e.location, e.start_at, e.end_at, e.status, e.created_at, e.updated_at,
			COALESCE(c.name, '') AS category_name,
			COALESCE(u.name, '') AS organizer_name
		FROM events e
		LEFT JOIN categories c ON c.id = e.category_id
		LEFT JOIN users u ON u.id = e.organizer_id
		WHERE ($1 = '' OR e.title ILIKE '%' || $1 || '%' OR e.description ILIKE '%' || $1 || '%')
		  AND ($2 = 0 OR e.category_id = $2)
		  AND ($3 = '' OR e.status = $3)
		ORDER BY e.created_at DESC
		LIMIT $4 OFFSET $5`

	rows, err := r.pool.Query(ctx, dataQuery, q.Search, q.CategoryID, q.Status, q.Limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var event model.Event
		if err := rows.Scan(
			&event.ID,
			&event.OrganizerID,
			&event.CategoryID,
			&event.Title,
			&event.Description,
			&event.Location,
			&event.StartAt,
			&event.EndAt,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.CategoryName,
			&event.OrganizerName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan event row: %w", err)
		}
		events = append(events, event)
	}

	return events, total, nil
}

func (r *EventRepository) Update(ctx context.Context, event *model.Event) error {
	query := `
		UPDATE events
		SET category_id = $1,
		    title = $2,
		    description = $3,
		    location = $4,
		    start_at = $5,
		    end_at = $6,
		    status = $7,
		    updated_at = now()
		WHERE id = $8
		RETURNING updated_at`

	event.UpdatedAt = time.Now()
	return r.pool.QueryRow(ctx, query,
		event.CategoryID,
		event.Title,
		event.Description,
		event.Location,
		event.StartAt,
		event.EndAt,
		event.Status,
		event.ID,
	).Scan(&event.UpdatedAt)
}
