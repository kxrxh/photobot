package classification

import (
	"time"

	"csort.ru/classification-service/internal/product"
	"github.com/google/uuid"
)

type CompleteClassification struct {
	Classification Classification `json:"classification"`
	Fractions      []Fraction     `json:"fractions"`
}

type Classification struct {
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	ID        uuid.UUID       `json:"id"`
	CreatedBy int32           `json:"created_by"`
	Product   product.Product `json:"product"`
	Name      string          `json:"name"`
	IsPublic  bool            `json:"is_public"`
}

type Fraction struct {
	ID         uuid.UUID   `json:"id"`
	Conditions []Condition `json:"conditions"`
	Name       string      `json:"name"`
	OrderIndex int32       `json:"order_index"`
}

type Condition struct {
	ID         uuid.UUID `json:"id"`
	Params     []Param   `json:"params"`
	Name       string    `json:"name"`
	Operator   string    `json:"operator"`
	Connection string    `json:"connection"`
	OrderIndex int32     `json:"order_index"`
}

type Param struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Operator string    `json:"operator"`
	Value    float32   `json:"value"`
}

type SaveCompleteClassificationRequest struct {
	Product   product.Product `json:"product"   validate:"required"`
	Fractions []Fraction      `json:"fractions" validate:"required,min=1,dive"`
	Name      string          `json:"name"      validate:"required,min=2,max=255"`
	IsPublic  bool            `json:"is_public"`
}

type ClassificationFilters struct {
	ProductID *uuid.UUID `json:"product_id" validate:"omitempty,uuid"`
	CreatedBy *int32     `json:"created_by" validate:"omitempty,min=1"`
	Name      *string    `json:"name"       validate:"omitempty,min=1,max=255"`
}
