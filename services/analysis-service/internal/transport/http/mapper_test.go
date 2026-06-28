package http

import (
	"testing"

	"csort.ru/analysis-service/internal/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectToResponse_MapsMass1000(t *testing.T) {
	mass1000 := 52.86
	mass := 0.05286
	idImage := "42"
	objClass := "wheat"
	colorRhs := "Y12"
	lw := 1.97
	pr := 17.6
	hm := 0.8
	minH := 10.0
	maxH := 20.0
	obj := objects.Object{
		ObjectMetadata: objects.ObjectMetadata{
			ID:       7,
			IDImage:  &idImage,
			Class:    &objClass,
			ColorRhs: &colorRhs,
			LW:       &lw,
			Pr:       &pr,
			HM:       &hm,
			MinH:     &minH,
			MaxH:     &maxH,
			Mass1000: &mass1000,
			Mass:     &mass,
		},
	}

	resp := ObjectToResponse(obj)

	assert.Equal(t, int32(7), resp.ID)
	require.NotNil(t, resp.IDImage)
	assert.Equal(t, "42", *resp.IDImage)
	assert.Equal(t, "wheat", *resp.Class)
	require.NotNil(t, resp.ColorRhs)
	assert.Equal(t, "Y12", *resp.ColorRhs)
	require.NotNil(t, resp.LW)
	assert.Equal(t, 1.97, *resp.LW)
	require.NotNil(t, resp.Pr)
	assert.Equal(t, 17.6, *resp.Pr)
	require.NotNil(t, resp.HM)
	assert.Equal(t, 0.8, *resp.HM)
	require.NotNil(t, resp.MinH)
	assert.Equal(t, 10.0, *resp.MinH)
	require.NotNil(t, resp.MaxH)
	assert.Equal(t, 20.0, *resp.MaxH)
	require.NotNil(t, resp.Mass1000)
	assert.Equal(t, 52.86, *resp.Mass1000)
	require.NotNil(t, resp.Mass)
	assert.InDelta(t, 0.05286, *resp.Mass, 1e-9)
}
