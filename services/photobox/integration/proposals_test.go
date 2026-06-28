//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"csort.ru/coffeebot/integration/harness"
	"csort.ru/coffeebot/internal/weed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProposals_Apply_CreatesWeed(t *testing.T) {
	ctx := context.Background()
	env := harness.NewTestEnv(ctx, t)

	proposalID := createProposal(t, env, "Apply creates weed")
	authHeader := map[string]string{"Authorization": env.GetToken(1, []string{"admin"})}

	resp := doPOST(
		t,
		env.BaseURL+"/api/v1/proposals/"+strconv.Itoa(int(proposalID))+"/apply",
		"application/json",
		[]byte(`{"note":"Approved"}`),
		authHeader,
	)
	requireHTTPStatus(t, resp, http.StatusOK)

	result := parseAPIResponse[struct {
		Status        string `json:"status"`
		AppliedWeedID *int32 `json:"applied_weed_id"`
	}](t, resp)
	assert.Equal(t, "applied", result.Status)
	require.NotNil(t, result.AppliedWeedID)

	weedResp := doGET(t, env.BaseURL+"/api/v1/weeds/"+strconv.Itoa(int(*result.AppliedWeedID)))
	requireHTTPStatus(t, weedResp, http.StatusOK)
	weedOut := parseAPIResponse[weed.Weed](t, weedResp)
	assert.Equal(t, "Apply creates weed", weedOut.Name)
}

func createProposal(t *testing.T, env *harness.TestEnv, name string) int32 {
	t.Helper()
	authHeader := map[string]string{"Authorization": env.GetToken(1, []string{"admin"})}
	body, err := json.Marshal(map[string]string{"name": name})
	require.NoError(t, err)
	resp := doPOST(t, env.BaseURL+"/api/v1/proposals/", "application/json", body, authHeader)
	requireHTTPStatus(t, resp, http.StatusCreated)
	return parseAPIResponse[struct {
		ID int32 `json:"id"`
	}](t, resp).ID
}
