package model

import "time"

type TicketType struct {
	ID             int64     `json:"id"`
	EventID        int64     `json:"event_id"`
	Name           string    `json:"name"`
	Price          float64   `json:"price"`
	Quota          int       `json:"quota"`
	Sold           int       `json:"sold"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AvailableSeats int       `json:"available_seats"`
}

type CreateTicketTypeRequest struct {
	Name  string  `json:"name"  validate:"required,min=2,max=100"`
	Price float64 `json:"price" validate:"gte=0"`
	Quota int     `json:"quota" validate:"required,gt=0"`
}
