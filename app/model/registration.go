package model

import "time"

type Registration struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	EventID      int64     `json:"event_id"`
	TicketTypeID int64     `json:"ticket_type_id"`
	TicketCode   string    `json:"ticket_code"`
	Status       string    `json:"status"` // registered, cancelled, checked_in
	RegisteredAt time.Time `json:"registered_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// field hasil JOIN
	EventTitle     string    `json:"event_title,omitempty"`
	EventLocation  string    `json:"event_location,omitempty"`
	EventStartAt   time.Time `json:"event_start_at,omitempty"`
	TicketTypeName string    `json:"ticket_type_name,omitempty"`
	UserName       string    `json:"user_name,omitempty"`
}

type CreateRegistrationRequest struct {
	TicketTypeID int64 `json:"ticket_type_id" validate:"required,gt=0"`
}

type CheckInRequest struct {
	TicketCode string `json:"ticket_code" validate:"required,min=5,max=20"`
}

type RegistrationListQuery struct {
	Status string `query:"status"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

