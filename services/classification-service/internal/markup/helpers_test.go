package markup

import (
	"testing"

	database "csort.ru/classification-service/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupObjectsByFractionID(t *testing.T) {
	f1, f2 := uuid.New(), uuid.New()
	got := groupObjectsByFractionID([]database.MarkupFractionObject{
		{MarkupFractionID: f1, ObjectID: 10},
		{MarkupFractionID: f1, ObjectID: 20},
		{MarkupFractionID: f2, ObjectID: 30},
	})
	assert.Equal(t, []int64{10, 20}, got[f1])
	assert.Equal(t, []int64{30}, got[f2])
}

func TestGroupAnalysesByMarkupID(t *testing.T) {
	m1, m2 := uuid.New(), uuid.New()
	got := groupAnalysesByMarkupID([]database.MarkupAnalysis{
		{MarkupID: m1, AnalysisID: 1},
		{MarkupID: m1, AnalysisID: 2},
		{MarkupID: m2, AnalysisID: 3},
	})
	assert.Equal(t, []int64{1, 2}, got[m1])
	assert.Equal(t, []int64{3}, got[m2])
}

func TestBuildMarkupFractions_emptyObjectIDs(t *testing.T) {
	fracID := uuid.New()
	now := database.MarkupFraction{}.CreatedAt
	fractions := []database.MarkupFraction{{
		ID:        fracID,
		Name:      "f1",
		CreatedAt: now,
		UpdatedAt: now,
	}}

	got := buildMarkupFractions(fractions, map[uuid.UUID][]int64{})
	require.Len(t, got, 1)
	require.NotNil(t, got[0].ObjectIDs)
	assert.Empty(t, got[0].ObjectIDs)
}
