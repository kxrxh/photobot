//go:build integration

package integration

import (
	"net/http"
	"testing"

	"csort.ru/classification-service/integration/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserActiveClassifications_SetGetDelete(t *testing.T) {
	env := testEnv(t)
	v1 := func(parts ...string) string { return apiV1(env.BaseURL, parts...) }

	productID := harness.SeedProduct(t, t.Context(), env.DBPool, "UAC Product")
	classID := harness.SeedClassification(
		t,
		t.Context(),
		env.DBPool,
		productID,
		harness.TestUserID,
		"UAC Classification",
	)

	userHdr := bearer(env.UserToken)
	svcHdr := bearer(env.ServiceToken)

	t.Run("set active classification", func(t *testing.T) {
		body := mustJSON(t, map[string]any{
			"user_id":           harness.TestUserID,
			"classification_id": classID.String(),
		})
		resp := doPOST(t, v1("user-active-classifications"), "application/json", body, userHdr)

		var out successMsgBody
		readResponseJSON(t, resp, http.StatusOK, &out)
		require.True(t, out.Success)
	})

	t.Run("set active classification without user_id in body", func(t *testing.T) {
		body := mustJSON(t, map[string]any{
			"classification_id": classID.String(),
		})
		resp := doPOST(t, v1("user-active-classifications"), "application/json", body, userHdr)

		var out successMsgBody
		readResponseJSON(t, resp, http.StatusOK, &out)
		require.True(t, out.Success)
	})

	t.Run("get own active classification", func(t *testing.T) {
		resp := doGET(t, v1("user-active-classifications"), userHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				Classification struct {
					ID string `json:"id"`
				} `json:"classification"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, classID.String(), wrapper.Result.Classification.ID)
	})

	t.Run("get by messenger_user_id requires admin or service", func(t *testing.T) {
		resp := doGET(t, v1("user-active-classifications", "12345"), userHdr)
		requireHTTPStatus(t, resp, http.StatusForbidden)
	})

	t.Run("get by messenger_user_id with service token", func(t *testing.T) {
		url := v1("user-active-classifications", "12345") + "?platform=telegram"
		resp := doGET(t, url, svcHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				Classification struct {
					ID string `json:"id"`
				} `json:"classification"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, classID.String(), wrapper.Result.Classification.ID)
	})

	t.Run("get by messenger_user_id with normalized MAX platform", func(t *testing.T) {
		url := v1("user-active-classifications", "12345") + "?platform=%20MAX%20"
		resp := doGET(t, url, svcHdr)
		requireHTTPStatus(t, resp, http.StatusOK)
	})

	t.Run("get by messenger_user_id without platform returns 400", func(t *testing.T) {
		resp := doGET(t, v1("user-active-classifications", "12345"), svcHdr)
		requireHTTPStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("get by messenger_user_id with invalid platform returns 400", func(t *testing.T) {
		url := v1("user-active-classifications", "12345") + "?platform=vk"
		resp := doGET(t, url, svcHdr)
		requireHTTPStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("delete", func(t *testing.T) {
		resp := doDELETE(t, v1("user-active-classifications"), userHdr)

		var out successMsgBody
		readResponseJSON(t, resp, http.StatusOK, &out)
		require.True(t, out.Success)
	})

	t.Run("get after delete returns 200 with null", func(t *testing.T) {
		resp := doGET(t, v1("user-active-classifications"), userHdr)
		var wrapper struct {
			Success bool `json:"success"`
			Result  any  `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Nil(t, wrapper.Result)
	})
}
