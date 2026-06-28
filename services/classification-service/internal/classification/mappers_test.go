package classification

import (
	"testing"
	"time"

	database "csort.ru/classification-service/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFiltersToActiveParams(t *testing.T) {
	createdBy := int32(7)
	productID := uuid.New()
	name := "foo"

	params := filtersToActiveParams(42, ClassificationFilters{
		CreatedBy: &createdBy,
		ProductID: &productID,
		Name:      &name,
	})

	assert.Equal(t, int32(42), params.UserID)
	require.NotNil(t, params.CreatedBy)
	assert.Equal(t, int32(7), *params.CreatedBy)
	assert.True(t, params.ProductID.Valid)
	assert.Equal(t, productID, uuid.UUID(params.ProductID.Bytes))
	require.NotNil(t, params.Name)
	assert.Equal(t, "foo", *params.Name)
}

func TestClassificationFromFiltersActiveRow(t *testing.T) {
	id := uuid.New()
	productID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	row := database.GetClassificationsWithFiltersAndActiveRow{
		ID:               id,
		Name:             "cls",
		CreatedBy:        1,
		IsPublic:         true,
		ProductID:        productID,
		CreatedAt:        now,
		UpdatedAt:        now,
		ProductIDFull:    productID,
		ProductName:      "prod",
		ProductCreatedAt: now,
		ProductUpdatedAt: now,
	}

	got := classificationFromFiltersActiveRow(row)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "cls", got.Name)
	assert.Equal(t, productID, got.Product.ID)
	assert.Equal(t, "prod", got.Product.Name)
}
