//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkup_CreateGetUpdateDelete(t *testing.T) {
	env := testEnv(t)
	v1 := func(parts ...string) string { return apiV1(env.BaseURL, parts...) }
	userHdr := bearer(env.UserToken)

	createPayload := map[string]any{
		"name":         "Test Markup",
		"fractions":    []map[string]any{{"name": "F1", "object_ids": []int64{}}},
		"analyses_ids": []int64{},
	}
	body, err := json.Marshal(createPayload)
	require.NoError(t, err)

	resp := doPOST(t, v1("markup"), "application/json", body, userHdr)

	var createRes struct {
		Success bool `json:"success"`
		Result  struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	readResponseJSON(t, resp, http.StatusCreated, &createRes)
	require.True(t, createRes.Success)
	assert.Equal(t, "Test Markup", createRes.Result.Name)

	markupID, err := uuid.Parse(createRes.Result.ID)
	require.NoError(t, err)

	t.Run("get by id", func(t *testing.T) {
		resp := doGET(t, v1("markup", markupID.String()), userHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, "Test Markup", wrapper.Result.Name)
	})

	updatePayload := map[string]any{
		"name":         "Updated Markup",
		"fractions":    []map[string]any{{"name": "F1", "object_ids": []int64{}}},
		"analyses_ids": []int64{},
	}
	updateBody, _ := json.Marshal(updatePayload)

	t.Run("update", func(t *testing.T) {
		resp := doPUT(t, v1("markup", markupID.String()), "application/json", updateBody, userHdr)
		requireHTTPStatus(t, resp, http.StatusOK)

		getResp := doGET(t, v1("markup", markupID.String()), userHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				Name string `json:"name"`
			} `json:"result"`
		}
		readResponseJSON(t, getResp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, "Updated Markup", wrapper.Result.Name)
	})

	t.Run("delete", func(t *testing.T) {
		resp := doDELETE(t, v1("markup", markupID.String()), userHdr)

		var out successMsgBody
		readResponseJSON(t, resp, http.StatusOK, &out)
		require.True(t, out.Success)
	})

	t.Run("get deleted returns 404", func(t *testing.T) {
		resp := doGET(t, v1("markup", markupID.String()), userHdr)
		requireHTTPStatus(t, resp, http.StatusNotFound)
	})
}
