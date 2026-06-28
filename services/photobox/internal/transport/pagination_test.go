package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQueryInt32(t *testing.T) {
	assert.Equal(t, int32(10), ParseQueryInt32("", 10))
	assert.Equal(t, int32(10), ParseQueryInt32("bad", 10))
	assert.Equal(t, int32(25), ParseQueryInt32("25", 10))
}

func TestClampPagination(t *testing.T) {
	limit, offset := ClampPagination(0, -5)
	assert.Equal(t, int32(10), limit)
	assert.Equal(t, int32(0), offset)

	limit, offset = ClampPagination(200, 3)
	assert.Equal(t, MaxPageLimit, limit)
	assert.Equal(t, int32(3), offset)
}

func TestClampPaginationFromQuery(t *testing.T) {
	limit, offset := ClampPaginationFromQuery("", "5")
	assert.Equal(t, int32(10), limit)
	assert.Equal(t, int32(5), offset)
}
