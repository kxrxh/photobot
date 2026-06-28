package classification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapMainGroup(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"known MONOCOT", "MONOCOT", "Однодольные"},
		{"known DICOT", "DICOT", "Двудольные"},
		{"known SPORE", "SPORE", "Споровые"},
		{"unknown code", "UNKNOWN", "UNKNOWN"},
		{"empty code", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapMainGroup(tt.code)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMapMainSubgroup(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"known ANNUAL", "ANNUAL", "Малолетние"},
		{"known PERENNIAL", "PERENNIAL", "Многолетние"},
		{"unknown code", "UNKNOWN", "UNKNOWN"},
		{"empty code", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapMainSubgroup(tt.code)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMapSubgroup(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"known ANNUAL_SPRING_EARLY", "ANNUAL_SPRING_EARLY", "Однолетние яровые ранние"},
		{"known EPHEMERAL", "EPHEMERAL", "Эфемеры"},
		{"known RHIZOME", "RHIZOME", "Корневищные"},
		{"unknown code", "UNKNOWN", "UNKNOWN"},
		{"empty code", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapSubgroup(tt.code)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMainGroupMap(t *testing.T) {
	m := MainGroupMap()
	require.NotNil(t, m)
	assert.Len(t, m, 3)
	assert.Equal(t, "Однодольные", m["MONOCOT"])
}

func TestMainSubgroupMap(t *testing.T) {
	m := MainSubgroupMap()
	require.NotNil(t, m)
	assert.Len(t, m, 2)
	assert.Equal(t, "Малолетние", m["ANNUAL"])
}

func TestSubgroupMap(t *testing.T) {
	m := SubgroupMap()
	require.NotNil(t, m)
	assert.GreaterOrEqual(t, len(m), 15)
	assert.Equal(t, "Корневищные", m["RHIZOME"])
}

func TestHierarchyMap(t *testing.T) {
	h := HierarchyMap()
	require.NotNil(t, h)
	assert.Contains(t, h, "MONOCOT")
	assert.Contains(t, h["MONOCOT"], "ANNUAL")
	assert.Contains(t, h["MONOCOT"]["ANNUAL"], "ANNUAL_SPRING_EARLY")
	assert.Contains(t, h, "DICOT")
	assert.Contains(t, h, "SPORE")
	assert.Contains(t, h["SPORE"]["PERENNIAL"], "SPORE_PERENNIAL_RHIZOME")
}
