//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"csort.ru/coffeebot/integration/harness"
	"csort.ru/coffeebot/internal/proposal"
	testutil "csort.ru/coffeebot/internal/testing"
	"csort.ru/coffeebot/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createProposalAs(t *testing.T, env *harness.TestEnv, name, rawToken string) int32 {
	t.Helper()
	body, err := json.Marshal(map[string]string{"name": name})
	require.NoError(t, err)
	resp := doPOST(
		t,
		env.BaseURL+"/api/v1/proposals/",
		"application/json",
		body,
		env.AuthHeader(rawToken),
	)
	requireHTTPStatus(t, resp, http.StatusCreated)
	return parseAPIResponse[struct {
		ID int32 `json:"id"`
	}](t, resp).ID
}

func TestProposals_FullModerationFlow(t *testing.T) {
	ctx := context.Background()
	env := harness.NewTestEnv(ctx, t)

	pid := createProposalAs(t, env, "Full flow weed", testutil.IntegrationTokenCatalogA)
	path := env.BaseURL + "/api/v1/proposals/" + strconv.Itoa(int(pid))

	rcResp := doPOST(
		t, path+"/request-changes", "application/json",
		[]byte(`{"message":"Expand description"}`),
		env.AuthHeader(testutil.IntegrationTokenModerator),
	)
	requireHTTPStatus(t, rcResp, http.StatusOK)
	drainResponseBody(t, rcResp)

	patchResp := doPATCH(
		t, path, "application/json",
		[]byte(`{"name":"Full flow weed v2"}`),
		env.AuthHeader(testutil.IntegrationTokenCatalogA),
	)
	requireHTTPStatus(t, patchResp, http.StatusOK)
	drainResponseBody(t, patchResp)

	submitResp := doPOST(t, path+"/submit", "", nil, env.AuthHeader(testutil.IntegrationTokenCatalogA))
	requireHTTPStatus(t, submitResp, http.StatusOK)
	drainResponseBody(t, submitResp)

	applyResp := doPOST(
		t, path+"/apply", "application/json",
		[]byte(`{"note":"LGTM"}`),
		env.AuthHeader(testutil.IntegrationTokenAdmin),
	)
	requireHTTPStatus(t, applyResp, http.StatusOK)
	applied := parseAPIResponse[struct {
		Status        string `json:"status"`
		AppliedWeedID *int32 `json:"applied_weed_id"`
	}](t, applyResp)
	assert.Equal(t, "applied", applied.Status)
	require.NotNil(t, applied.AppliedWeedID)
}

func TestProposals_AccessControl(t *testing.T) {
	ctx := context.Background()
	env := harness.NewTestEnv(ctx, t)

	pid := createProposalAs(t, env, "Secret draft", testutil.IntegrationTokenCatalogA)
	createProposalAs(t, env, "Belongs to B", testutil.IntegrationTokenCatalogB)

	getResp := doGET(
		t,
		env.BaseURL+"/api/v1/proposals/"+strconv.Itoa(int(pid)),
		env.AuthHeader(testutil.IntegrationTokenCatalogB),
	)
	assert.Equal(t, "Access denied", requireJSONError(t, getResp, http.StatusForbidden))

	listResp := doGET(
		t,
		env.BaseURL+"/api/v1/proposals/?limit=20&offset=0",
		env.AuthHeader(testutil.IntegrationTokenCatalogA),
	)
	requireHTTPStatus(t, listResp, http.StatusOK)
	page := parseAPIResponse[dto.PaginatedResponse[proposal.ProposalListItem]](t, listResp)
	require.Len(t, page.Data, 1)
	assert.Equal(t, "Secret draft", page.Data[0].PendingName)

	applyResp := doPOST(
		t,
		env.BaseURL+"/api/v1/proposals/"+strconv.Itoa(int(pid))+"/apply",
		"application/json",
		[]byte(`{"note":"nope"}`),
		env.AuthHeader(testutil.IntegrationTokenCatalogA),
	)
	assert.Contains(t, requireJSONError(t, applyResp, http.StatusForbidden), "permission")

	rcResp := doPOST(
		t,
		env.BaseURL+"/api/v1/proposals/"+strconv.Itoa(int(pid))+"/request-changes",
		"application/json",
		[]byte(`{"message":"fix"}`),
		env.AuthHeader(testutil.IntegrationTokenModerator),
	)
	requireHTTPStatus(t, rcResp, http.StatusOK)
	drainResponseBody(t, rcResp)

	patchResp := doPATCH(
		t,
		env.BaseURL+"/api/v1/proposals/"+strconv.Itoa(int(pid)),
		"application/json",
		[]byte(`{"name":"Hijacked"}`),
		env.AuthHeader(testutil.IntegrationTokenCatalogB),
	)
	assert.Equal(t, "Only the author can edit proposals", requireJSONError(t, patchResp, http.StatusForbidden))
}
