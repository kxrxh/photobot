//go:build integration

package integration

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBotsCRUD(t *testing.T) {
	env := testEnv(t)
	base := apiV1(env.BaseURL)
	adminToken := getAdminToken(t, env)
	authHeader := bearer(adminToken)

	botName := "integration-test-bot"
	botToken := "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"

	var botID int32
	var svcToken string
	t.Run("create bot", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"name":     botName,
			"token":    botToken,
			"platform": "telegram",
		})
		resp := doPOST(t, base+"/bots/", "application/json", body, authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		var r CreateIDResponse
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		botID = r.Result.ID
		require.Greater(t, botID, int32(0))
	})

	t.Run("list bots", func(t *testing.T) {
		resp := doGET(t, base+"/bots/", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result BotListResponse
		parseJSON(t, resp, &result)
		require.True(t, result.Success)
		var found bool
		for _, b := range result.Result {
			if b.Name == botName && b.Platform == "telegram" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("get bot by name requires service token", func(t *testing.T) {
		svcBody := mustJSON(t, map[string]string{
			"service_id":     "bot-reader-service",
			"service_secret": "secret67890",
		})
		createResp := doPOST(t, base+"/services/", "application/json", svcBody, authHeader)
		createResp.Body.Close()
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		loginBody := mustJSON(t, map[string]string{
			"service_id":     "bot-reader-service",
			"service_secret": "secret67890",
		})
		loginResp := doPOST(t, base+"/auth/login", "application/json", loginBody, map[string]string{
			"X-Grant-Type": "client_credentials",
		})
		defer loginResp.Body.Close()
		require.Equal(t, http.StatusOK, loginResp.StatusCode)

		var loginResult ServiceLoginResponse
		parseJSON(t, loginResp, &loginResult)
		svcToken = loginResult.Result.AccessToken
		require.NotEmpty(t, svcToken)

		h := bearer(svcToken)
		h["X-Bot-Name"] = botName
		h["X-Grant-Type"] = "client_credentials"
		resp := doGET(t, base+"/bots/"+botName, h)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var botResult BotResponse
		parseJSON(t, resp, &botResult)
		require.True(t, botResult.Success)
		assert.Equal(t, botName, botResult.Result.Name)
	})

	t.Run("get bot token by name and platform", func(t *testing.T) {
		require.NotEmpty(t, svcToken)

		h := bearer(svcToken)
		h["X-Bot-Name"] = botName
		h["X-Messenger-Platform"] = "telegram"
		h["X-Grant-Type"] = "client_credentials"
		resp := doGET(t, base+"/bots/token", h)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var tokenResult BotTokenResponse
		parseJSON(t, resp, &tokenResult)
		require.True(t, tokenResult.Success)
		assert.NotEmpty(t, tokenResult.Result.Token)
	})

	t.Run("update bot", func(t *testing.T) {
		newToken := "1234567890:XYZupdatedTokenABCDEFGH"
		body := mustJSON(t, map[string]string{"token": newToken})
		resp := doPUT(
			t,
			base+"/bots/"+strconv.Itoa(int(botID)),
			"application/json",
			body,
			authHeader,
		)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("delete bot", func(t *testing.T) {
		resp := doDELETE(t, base+"/bots/"+strconv.Itoa(int(botID)), authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
