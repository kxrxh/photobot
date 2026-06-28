package product

import (
	"time"

	"github.com/google/uuid"
)

type SaveProduct struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
}

type Product struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
