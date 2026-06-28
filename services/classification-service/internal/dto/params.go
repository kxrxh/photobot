package dto

import (
	"time"

	"github.com/google/uuid"
)

type SaveParamRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type ParamResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
