//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func readResponseBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	return body
}

func requireHTTPStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode == want {
		return
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.FailNowf(t, "HTTP status", "want %d, got %d, body=%q", want, resp.StatusCode, string(body))
}

func drainResponseBody(t *testing.T, resp *http.Response) {
	t.Helper()
	_, err := io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func requireNoContent(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		require.FailNowf(t, "HTTP status", "want %d, got %d, body=%q", http.StatusNoContent, resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Empty(t, body, "expected empty body for 204")
}

func doGET(t *testing.T, url string, headers ...map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	for _, h := range headers {
		for k, v := range h {
			req.Header.Set(k, v)
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func doPOST(
	t *testing.T,
	url, contentType string,
	body []byte,
	headers map[string]string,
) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func doPUT(
	t *testing.T,
	url, contentType string,
	body []byte,
	headers map[string]string,
) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPut, url, bodyReader)
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func doDELETE(t *testing.T, url string, headers map[string]string, body ...[]byte) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if len(body) > 0 && body[0] != nil {
		bodyReader = bytes.NewReader(body[0])
	}
	req, err := http.NewRequest(http.MethodDelete, url, bodyReader)
	require.NoError(t, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func doPATCH(
	t *testing.T,
	url, contentType string,
	body []byte,
	headers map[string]string,
) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPatch, url, bodyReader)
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

type APIResponse[T any] struct {
	Success bool `json:"success"`
	Result  T    `json:"result"`
}

type apiErrorEnvelope struct {
	Success bool `json:"success"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func requireJSONError(t *testing.T, resp *http.Response, wantHTTP int) string {
	t.Helper()
	require.Equal(t, wantHTTP, resp.StatusCode, "body should be JSON error")
	body := readResponseBody(t, resp)
	var w apiErrorEnvelope
	require.NoError(t, json.Unmarshal(body, &w), "body=%q", body)
	require.False(t, w.Success, "body=%q", body)
	return w.Error.Message
}

func parseAPIResponse[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	body := readResponseBody(t, resp)
	var w APIResponse[T]
	if err := json.Unmarshal(body, &w); err != nil {
		t.Fatalf("decode API wrapper: %v body=%q", err, body)
	}
	require.True(t, w.Success, "success=false body=%q", body)
	return w.Result
}
