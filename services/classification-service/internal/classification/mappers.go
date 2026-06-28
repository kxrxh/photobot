package classification

import (
	"time"

	database "csort.ru/classification-service/internal/database"
	prod "csort.ru/classification-service/internal/product"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func classificationFromProductFields(
	id uuid.UUID,
	name string,
	createdBy int32,
	isPublic bool,
	createdAt, updatedAt time.Time,
	productID uuid.UUID,
	productName string,
	productCreatedAt, productUpdatedAt time.Time,
) Classification {
	return Classification{
		ID:        id,
		Name:      name,
		CreatedBy: createdBy,
		IsPublic:  isPublic,
		Product: prod.Product{
			ID:        productID,
			Name:      productName,
			CreatedAt: productCreatedAt,
			UpdatedAt: productUpdatedAt,
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func classificationFromFiltersActiveRow(
	row database.GetClassificationsWithFiltersAndActiveRow,
) Classification {
	return classificationFromProductFields(
		row.ID,
		row.Name,
		row.CreatedBy,
		row.IsPublic,
		row.CreatedAt,
		row.UpdatedAt,
		row.ProductIDFull,
		row.ProductName,
		row.ProductCreatedAt,
		row.ProductUpdatedAt,
	)
}

func filtersToActiveParams(
	userID int32,
	filters ClassificationFilters,
) database.GetClassificationsWithFiltersAndActiveParams {
	params := database.GetClassificationsWithFiltersAndActiveParams{
		UserID: userID,
	}
	if filters.CreatedBy != nil {
		params.CreatedBy = filters.CreatedBy
	}
	if filters.ProductID != nil {
		params.ProductID = pgtype.UUID{Bytes: *filters.ProductID, Valid: true}
	}
	if filters.Name != nil {
		params.Name = filters.Name
	}
	return params
}

type conditionKey struct {
	fractionID uuid.UUID
	orderIndex int32
}

type paramKey struct {
	conditionID uuid.UUID
	name        string
}

type fractionInsertPlan struct {
	sourceIndex int
	orderIndex  int32
}

type conditionInsertPlan struct {
	fractionIndex  int
	conditionIndex int
	orderIndex     int32
}

type paramInsertPlan struct {
	fractionIndex  int
	conditionIndex int
	paramIndex     int
	sourceValue    float32
}
