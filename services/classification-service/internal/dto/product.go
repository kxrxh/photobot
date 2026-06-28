package dto

import (
	"time"

	"github.com/google/uuid"
)

type SaveProductRequest struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
}

type ProductResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
