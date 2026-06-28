//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestUnauthorized_NoBearer(t *testing.T) {
	env := testEnv(t)

	cases := []struct {
		name string
		run  func() *http.Response
	}{
		{
			"GET /users/",
			func() *http.Response {
				return doGET(t, apiV1(env.BaseURL, "users")+"/", nil)
			},
		},
		{
			"GET /users/me",
			func() *http.Response {
				return doGET(t, apiV1(env.BaseURL, "users", "me"), nil)
			},
		},
		{
			"GET /users/me/roles",
			func() *http.Response {
				return doGET(t, apiV1(env.BaseURL, "users", "me", "roles"), nil)
			},
		},
		{
			"GET /roles/",
			func() *http.Response {
				return doGET(t, apiV1(env.BaseURL, "roles")+"/", nil)
			},
		},
		{
			"GET /services/",
			func() *http.Response {
				return doGET(t, apiV1(env.BaseURL, "services")+"/", nil)
			},
		},
		{
			"GET /bots/",
			func() *http.Response {
				return doGET(t, apiV1(env.BaseURL, "bots")+"/", nil)
			},
		},
		{
			"POST /auth/link-code",
			func() *http.Response {
				return doPOST(t, apiV1(env.BaseURL, "auth", "link-code"), "application/json", nil, nil)
			},
		},
		{
			"PUT /users/me",
			func() *http.Response {
				body := mustJSON(t, map[string]string{"full_name": "x"})
				return doPUT(t, apiV1(env.BaseURL, "users", "me"), "application/json", body, nil)
			},
		},
		{
			"PUT /users/:id",
			func() *http.Response {
				body := mustJSON(t, map[string]string{"full_name": "x"})
				return doPUT(t, apiV1(env.BaseURL, "users", "1"), "application/json", body, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireHTTPStatus(t, tc.run(), http.StatusUnauthorized)
		})
	}
}
