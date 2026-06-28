//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"csort.ru/coffeebot/integration/harness"
	"csort.ru/coffeebot/internal/database"
	testutil "csort.ru/coffeebot/internal/testing"
	"csort.ru/coffeebot/internal/weed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeeds_GetUpdateDelete(t *testing.T) {
	ctx := context.Background()
	env := harness.NewTestEnv(ctx, t)

	queries := database.New(env.DBPool)
	weedRow, err := queries.CreateWeed(ctx, database.CreateWeedParams{
		Name:         "Integration weed",
		IsQuarantine: false,
	})
	require.NoError(t, err)

	id := strconv.Itoa(int(weedRow.ID))
	authHeader := map[string]string{"Authorization": env.GetToken(1, []string{"admin"})}

	getResp := doGET(t, env.BaseURL+"/api/v1/weeds/"+id)
	requireHTTPStatus(t, getResp, http.StatusOK)
	got := parseAPIResponse[weed.Weed](t, getResp)
	assert.Equal(t, weedRow.ID, got.ID)

	reqBody, err := json.Marshal(weed.SaveWeedParams{Name: "Updated", IsQuarantine: false})
	require.NoError(t, err)
	putResp := doPUT(t, env.BaseURL+"/api/v1/weeds/"+id, "application/json", reqBody, authHeader)
	requireHTTPStatus(t, putResp, http.StatusOK)
	updated := parseAPIResponse[weed.Weed](t, putResp)
	assert.Equal(t, "Updated", updated.Name)

	delResp := doDELETE(t, env.BaseURL+"/api/v1/weeds/"+id, authHeader)
	requireNoContent(t, delResp)

	notFound := doGET(t, env.BaseURL+"/api/v1/weeds/99999")
	defer notFound.Body.Close()
	assert.Equal(t, http.StatusNotFound, notFound.StatusCode)
}

func TestWeeds_Put_CatalogUserForbidden(t *testing.T) {
	ctx := context.Background()
	env := harness.NewTestEnv(ctx, t)

	queries := database.New(env.DBPool)
	weedRow, err := queries.CreateWeed(ctx, database.CreateWeedParams{
		Name:         "Role gate weed",
		IsQuarantine: false,
	})
	require.NoError(t, err)

	body := []byte(`{"name":"Hijack","is_quarantine":false}`)
	resp := doPUT(
		t,
		env.BaseURL+"/api/v1/weeds/"+strconv.Itoa(int(weedRow.ID)),
		"application/json",
		body,
		env.AuthHeader(testutil.IntegrationTokenCatalogA),
	)
	msg := requireJSONError(t, resp, http.StatusForbidden)
	assert.Contains(t, msg, "permission")
}
