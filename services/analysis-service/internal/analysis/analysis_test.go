package analysis

import (
	"testing"

	"csort.ru/analysis-service/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestGetAnalysesPaginatedRequest_SetDefaults(t *testing.T) {
	t.Run("sets default limit when zero", func(t *testing.T) {
		p := GetAnalysesPaginatedRequest{}
		p.SetDefaults()
		assert.Equal(t, int32(DefaultLimit), p.Limit)
	})

	t.Run("caps limit at MaxLimit when exceeded", func(t *testing.T) {
		p := GetAnalysesPaginatedRequest{PaginatedRequest: core.PaginatedRequest{Limit: 500}}
		p.SetDefaults()
		assert.Equal(t, int32(MaxLimit), p.Limit)
	})

	t.Run("preserves limit when within range", func(t *testing.T) {
		p := GetAnalysesPaginatedRequest{PaginatedRequest: core.PaginatedRequest{Limit: 25}}
		p.SetDefaults()
		assert.Equal(t, int32(25), p.Limit)
	})

	t.Run("sets default sort_by when empty", func(t *testing.T) {
		p := GetAnalysesPaginatedRequest{}
		p.SetDefaults()
		assert.Equal(t, DefaultSortBy, p.SortBy)
	})

	t.Run("sets default sort_order when empty", func(t *testing.T) {
		p := GetAnalysesPaginatedRequest{}
		p.SetDefaults()
		assert.Equal(t, DefaultSortOrder, p.SortOrder)
	})

	t.Run("preserves sort_by and sort_order when set", func(t *testing.T) {
		p := GetAnalysesPaginatedRequest{
			SortBy:    "product",
			SortOrder: "asc",
		}
		p.SetDefaults()
		assert.Equal(t, "product", p.SortBy)
		assert.Equal(t, "asc", p.SortOrder)
	})
}
