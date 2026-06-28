package classification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderIndexFromRange(t *testing.T) {
	got, err := orderIndexFromRange(3)
	require.NoError(t, err)
	assert.Equal(t, int32(3), got)

	_, err = orderIndexFromRange(1 << 40)
	require.Error(t, err)
}
