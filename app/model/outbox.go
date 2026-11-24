package model

import (
	"encoding/json"
	"time"
)

type Outbox struct {
	ID         int64           `json:"id"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
	Status     string          `json:"status"` // pending, processing, completed, failed
	RetryCount int             `json:"retry_count"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
