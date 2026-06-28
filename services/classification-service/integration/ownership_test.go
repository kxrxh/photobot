//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassificationOwnershipTransfer(t *testing.T) {
	env := testEnv(t)
	url := apiV1(env.BaseURL, "classifications", "ownership-transfers")

	svc := bearer(env.ServiceToken)
	user := bearer(env.UserToken)
	validBody := mustJSON(t, map[string]int32{"from_user_id": 1, "to_user_id": 2})

	authCases := []struct {
		name   string
		hdr    map[string]string
		body   []byte
		status int
	}{
		{"no auth returns 401", nil, validBody, http.StatusUnauthorized},
		{"non-service role returns 403", user, validBody, http.StatusForbidden},
	}
	for _, tc := range authCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doPOST(t, url, "application/json", tc.body, tc.hdr)
			requireHTTPStatus(t, resp, tc.status)
		})
	}

	badBodyCases := []struct {
		name   string
		body   []byte
		status int
	}{
		{"invalid JSON returns 400", []byte(`not-json`), http.StatusBadRequest},
		{
			"nonPositive from_user_id returns 400",
			mustJSON(t, map[string]int32{"from_user_id": 0, "to_user_id": 1}),
			http.StatusBadRequest,
		},
		{
			"same from and to returns 400",
			mustJSON(t, map[string]int32{"from_user_id": 1, "to_user_id": 1}),
			http.StatusBadRequest,
		},
	}
	for _, tc := range badBodyCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doPOST(t, url, "application/json", tc.body, svc)
			requireHTTPStatus(t, resp, tc.status)
		})
	}

	t.Run("service token succeeds", func(t *testing.T) {
		body := mustJSON(t, map[string]int32{"from_user_id": 1, "to_user_id": 2})
		resp := doPOST(t, url, "application/json", body, svc)

		var out struct {
			Success bool `json:"success"`
			Result  struct {
				Message string `json:"message"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &out)
		require.True(t, out.Success)
		require.Equal(t, "Ownership transfer completed", out.Result.Message)
	})
}
