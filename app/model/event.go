package model

import "time"

type Event struct {
	ID            int64     `json:"id"`
	OrganizerID   int64     `json:"organizer_id"`
	CategoryID    int64     `json:"category_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Location      string    `json:"location"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	Status        string    `json:"status"` // draft, published, finished, cancelled
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CategoryName  string    `json:"category_name,omitempty"`
	OrganizerName string    `json:"organizer_name,omitempty"`
}

type CreateEventRequest struct {
	CategoryID  int64     `json:"category_id" validate:"required,gt=0"`
	Title       string    `json:"title"       validate:"required,min=3,max=255"`
	Description string    `json:"description" validate:"max=5000"`
	Location    string    `json:"location"    validate:"required,min=3,max=255"`
	StartAt     time.Time `json:"start_at"    validate:"required"`
	EndAt       time.Time `json:"end_at"      validate:"required"`
}

type UpdateEventRequest struct {
	CategoryID  int64      `json:"category_id,omitempty" validate:"omitempty,gt=0"`
	Title       string     `json:"title,omitempty"       validate:"omitempty,min=3,max=255"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
	Location    string     `json:"location,omitempty"    validate:"omitempty,min=3,max=255"`
	StartAt     *time.Time `json:"start_at,omitempty"`
	EndAt       *time.Time `json:"end_at,omitempty"`
	Status      string     `json:"status,omitempty"      validate:"omitempty,oneof=draft published finished cancelled"`
}

type EventListQuery struct {
	Search     string `query:"search"`
	CategoryID int64  `query:"category_id"`
	Status     string `query:"status"`
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
}

type EventDetailResponse struct {
	Event
	TicketTypes []TicketType `json:"ticket_types,omitempty"`
}
