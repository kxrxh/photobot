package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProductRef struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ParamRef struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Operator string    `json:"operator"`
	Value    float32   `json:"value"`
}

type ConditionRef struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Operator   string     `json:"operator"`
	Connection string     `json:"connection"`
	OrderIndex int32      `json:"order_index"`
	Params     []ParamRef `json:"params"`
}

type FractionRef struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	OrderIndex int32          `json:"order_index"`
	Conditions []ConditionRef `json:"conditions"`
}

type SaveClassificationRequest struct {
	Product   ProductRef    `json:"product"   validate:"required"`
	Fractions []FractionRef `json:"fractions" validate:"required,min=1,dive"`
	Name      string        `json:"name"      validate:"required,min=2,max=255"`
	IsPublic  bool          `json:"is_public"`
}

type ClassificationResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	CreatedBy int32      `json:"created_by"`
	Product   ProductRef `json:"product"`
	IsPublic  bool       `json:"is_public"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CompleteClassificationResponse struct {
	Classification ClassificationResponse `json:"classification"`
	Fractions      []FractionRef          `json:"fractions"`
}

type ListClassificationsResponse struct {
	Classifications      []ClassificationResponse `json:"classifications"`
	ActiveClassification *ClassificationResponse  `json:"active_classification,omitempty"`
}
