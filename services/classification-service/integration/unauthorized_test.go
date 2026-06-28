//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestUnauthorized_NoBearer(t *testing.T) {
	env := testEnv(t)
	b := env.BaseURL

	cases := []struct {
		name string
		run  func() *http.Response
	}{
		{
			"GET /classifications",
			func() *http.Response {
				return doGET(t, apiV1(b, "classifications"), nil)
			},
		},
		{
			"GET /params",
			func() *http.Response {
				return doGET(t, apiV1(b, "params"), nil)
			},
		},
		{
			"GET /products",
			func() *http.Response {
				return doGET(t, apiV1(b, "products"), nil)
			},
		},
		{
			"GET /markup",
			func() *http.Response {
				return doGET(t, apiV1(b, "markup"), nil)
			},
		},
		{
			"GET /user-active-classifications",
			func() *http.Response {
				return doGET(t, apiV1(b, "user-active-classifications"), nil)
			},
		},
		{
			"POST /correlation",
			func() *http.Response {
				body := []byte(
					`{"fractions":[{"name":"x","object_ids":[1]}],"parameter_groups":["all"]}`,
				)
				return doPOST(t, apiV1(b, "correlation"), "application/json", body, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireHTTPStatus(t, tc.run(), http.StatusUnauthorized)
		})
	}
}
