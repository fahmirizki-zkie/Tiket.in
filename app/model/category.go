package model

import "time"

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}
