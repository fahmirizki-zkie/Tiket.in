package model

import "time"

type Attendance struct {
	ID             int64     `json:"id"`
	RegistrationID int64     `json:"registration_id"`
	CheckedInBy    int64     `json:"checked_in_by"`
	CheckedInAt    time.Time `json:"checked_in_at"`

	// field tambahan untuk response, tidak disimpan di tabel
	TicketCode string `json:"ticket_code,omitempty"`
	EventTitle string `json:"event_title,omitempty"`
}
