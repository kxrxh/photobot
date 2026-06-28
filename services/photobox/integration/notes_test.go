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
	"csort.ru/coffeebot/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotes_CRUDAndAuthorization(t *testing.T) {
	ctx := context.Background()
	env := harness.NewTestEnv(ctx, t)

	queries := database.New(env.DBPool)
	weedRow, err := queries.CreateWeed(ctx, database.CreateWeedParams{
		Name:         "Notes weed",
		IsQuarantine: false,
	})
	require.NoError(t, err)

	weedNotes := env.BaseURL + "/api/v1/weeds/" + strconv.Itoa(int(weedRow.ID)) + "/notes/"
	tokenA := testutil.IntegrationTokenCatalogA
	tokenB := testutil.IntegrationTokenCatalogB

	createBody, err := json.Marshal(map[string]string{"note": "By A"})
	require.NoError(t, err)
	createResp := doPOST(t, weedNotes, "application/json", createBody, env.AuthHeader(tokenA))
	requireHTTPStatus(t, createResp, http.StatusCreated)
	noteID := parseAPIResponse[dto.CreateNoteResponse](t, createResp).ID

	hijackBody, err := json.Marshal(map[string]string{"note": "Hijack"})
	require.NoError(t, err)
	hijackResp := doPUT(
		t,
		env.BaseURL+"/api/v1/notes/"+strconv.Itoa(int(noteID)),
		"application/json",
		hijackBody,
		env.AuthHeader(tokenB),
	)
	msg := requireJSONError(t, hijackResp, http.StatusForbidden)
	assert.Equal(t, "You can only edit your own notes", msg)

	updateBody, err := json.Marshal(map[string]string{"note": "By admin"})
	require.NoError(t, err)
	updateResp := doPUT(
		t,
		env.BaseURL+"/api/v1/notes/"+strconv.Itoa(int(noteID)),
		"application/json",
		updateBody,
		env.AuthHeader(testutil.IntegrationTokenAdmin),
	)
	requireHTTPStatus(t, updateResp, http.StatusOK)
	updated := parseAPIResponse[dto.NoteListItem](t, updateResp)
	assert.Equal(t, "By admin", updated.Note)

	postNote(t, env, weedRow.ID, "By B", tokenB)

	listResp := doGET(t, weedNotes, env.AuthHeader(tokenA))
	requireHTTPStatus(t, listResp, http.StatusOK)
	list := parseAPIResponse[[]dto.NoteListItem](t, listResp)
	require.Len(t, list, 2)

	delResp := doDELETE(
		t,
		env.BaseURL+"/api/v1/notes/"+strconv.Itoa(int(noteID)),
		env.AuthHeader(testutil.IntegrationTokenAdmin),
	)
	requireNoContent(t, delResp)
}

func postNote(t *testing.T, env *harness.TestEnv, weedID int32, text, rawToken string) {
	t.Helper()
	body, err := json.Marshal(map[string]string{"note": text})
	require.NoError(t, err)
	resp := doPOST(
		t,
		env.BaseURL+"/api/v1/weeds/"+strconv.Itoa(int(weedID))+"/notes/",
		"application/json",
		body,
		env.AuthHeader(rawToken),
	)
	requireHTTPStatus(t, resp, http.StatusCreated)
}
