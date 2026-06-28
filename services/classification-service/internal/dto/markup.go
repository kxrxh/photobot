package dto

import (
	"time"

	"github.com/google/uuid"
)

type SaveMarkupFractionRequest struct {
	ObjectIDs []int64 `json:"object_ids"`
	Name      string  `json:"name"`
}

type SaveMarkupRequest struct {
	Name        string                      `json:"name"         validate:"required,min=1,max=255"`
	Fractions   []SaveMarkupFractionRequest `json:"fractions"    validate:"required"`
	AnalysesIDs []int64                     `json:"analyses_ids" validate:"required"`
}

type MarkupFractionResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ObjectIDs []int64   `json:"object_ids"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MarkupResponse struct {
	ID          uuid.UUID                `json:"id"`
	Name        string                   `json:"name"`
	CreatedBy   int32                    `json:"created_by"`
	Fractions   []MarkupFractionResponse `json:"fractions"`
	AnalysesIDs []int64                  `json:"analyses_ids"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}
