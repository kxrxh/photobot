//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"csort.ru/classification-service/integration/harness"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifications_CreateGetUpdateDelete(t *testing.T) {
	env := testEnv(t)
	v1 := func(parts ...string) string { return apiV1(env.BaseURL, parts...) }

	productID := harness.SeedProduct(t, t.Context(), env.DBPool, "Class Test Product")
	userHdr := bearer(env.UserToken)
	editorHdr := bearer(env.ClassificationEditorToken)

	createPayload := map[string]any{
		"name":      "Test Classification",
		"is_public": false,
		"product": map[string]any{
			"id":   productID.String(),
			"name": "Class Test Product",
		},
		"fractions": []map[string]any{
			{
				"name":        "Fraction 1",
				"order_index": 0,
				"conditions": []map[string]any{
					{
						"name":        "Condition 1",
						"operator":    "OR",
						"connection":  "AND",
						"order_index": 0,
						"params": []map[string]any{
							{"name": "param1", "operator": ">=", "value": 0},
						},
					},
				},
			},
		},
	}
	body, err := json.Marshal(createPayload)
	require.NoError(t, err)

	resp := doPOST(t, v1("classifications"), "application/json", body, userHdr)

	var createRes struct {
		Success bool `json:"success"`
		Result  struct {
			Classification struct {
				ID string `json:"id"`
			} `json:"classification"`
		} `json:"result"`
	}
	readResponseJSON(t, resp, http.StatusCreated, &createRes)
	require.True(t, createRes.Success)
	classID, err := uuid.Parse(createRes.Result.Classification.ID)
	require.NoError(t, err)

	t.Run("get by id", func(t *testing.T) {
		resp := doGET(t, v1("classifications", classID.String()), userHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				Classification struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"classification"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, "Test Classification", wrapper.Result.Classification.Name)
	})

	updatePayload := map[string]any{
		"name":      "Updated Classification",
		"is_public": false,
		"product": map[string]any{
			"id":   productID.String(),
			"name": "Class Test Product",
		},
		"fractions": createPayload["fractions"],
	}
	updateBody, _ := json.Marshal(updatePayload)

	t.Run("update", func(t *testing.T) {
		resp := doPUT(t, v1("classifications", classID.String()), "application/json", updateBody, userHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				Name string `json:"name"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, "Updated Classification", wrapper.Result.Name)
	})

	t.Run("make public requires admin or editor", func(t *testing.T) {
		resp := doPUT(t, v1("classifications", classID.String(), "public"), "", nil, userHdr)
		requireHTTPStatus(t, resp, http.StatusForbidden)
	})

	t.Run("make public with classification_editor", func(t *testing.T) {
		resp := doPUT(t, v1("classifications", classID.String(), "public"), "", nil, editorHdr)
		requireHTTPStatus(t, resp, http.StatusOK)
	})

	t.Run("make private with classification_editor", func(t *testing.T) {
		resp := doPUT(t, v1("classifications", classID.String(), "private"), "", nil, editorHdr)
		requireHTTPStatus(t, resp, http.StatusOK)
	})

	t.Run("delete", func(t *testing.T) {
		resp := doDELETE(t, v1("classifications", classID.String()), userHdr)

		var out successMsgBody
		readResponseJSON(t, resp, http.StatusOK, &out)
		require.True(t, out.Success)
	})

	t.Run("get deleted returns 404", func(t *testing.T) {
		resp := doGET(t, v1("classifications", classID.String()), userHdr)
		requireHTTPStatus(t, resp, http.StatusNotFound)
	})
}

func TestClassifications_GetByID_NotFound(t *testing.T) {
	env := testEnv(t)
	fakeID := uuid.New().String()
	resp := doGET(t, apiV1(env.BaseURL, "classifications", fakeID), bearer(env.UserToken))
	requireHTTPStatus(t, resp, http.StatusNotFound)
}
