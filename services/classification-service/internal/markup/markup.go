package markup

import (
	"time"

	"github.com/google/uuid"
)

type MarkupFraction struct {
	ObjectIDs []int64   `json:"object_ids"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
}

type Markup struct {
	Fractions   []MarkupFraction `json:"fractions"`
	AnalysesIDs []int64          `json:"analyses_ids"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	CreatedBy   int32            `json:"created_by"`
}

type SaveMarkupFraction struct {
	ObjectIDs []int64 `json:"object_ids"`
	Name      string  `json:"name"`
}

type SaveMarkupRequest struct {
	Name        string               `json:"name"         validate:"required,min=1,max=255"`
	CreatedBy   int32                `json:"-"`
	Fractions   []SaveMarkupFraction `json:"fractions"    validate:"required"`
	AnalysesIDs []int64              `json:"analyses_ids" validate:"required"`
}

type MarkupFilters struct {
	CreatedBy *int32  `json:"created_by" validate:"omitempty,min=1"`
	Name      *string `json:"name"       validate:"omitempty,min=1,max=255"`
}
