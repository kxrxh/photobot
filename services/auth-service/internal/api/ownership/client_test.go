package ownership

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"csort.ru/auth-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_NoDownstreamReturnsNil(t *testing.T) {
	cfg := &config.MergeConfig{}
	c := NewClient(cfg, func() (string, error) { return "t", nil })
	assert.Nil(t, c)
}

func TestNoopTransfer_NoOp(t *testing.T) {
	err := NoopTransfer{}.TransferOwnership(context.Background(), 1, 2)
	assert.NoError(t, err)
}

func TestClient_TransferOwnership_PostsToBothServices(t *testing.T) {
	var clsBodies, anaBodies []OwnershipTransferRequest
	var clsAuth, anaAuth string

	cls := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		clsAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		var req OwnershipTransferRequest
		require.NoError(t, json.Unmarshal(b, &req))
		clsBodies = append(clsBodies, req)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer cls.Close()

	ana := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		anaAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		var req OwnershipTransferRequest
		require.NoError(t, json.Unmarshal(b, &req))
		anaBodies = append(anaBodies, req)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ana.Close()

	cfg := &config.MergeConfig{
		ClassificationServiceURL: cls.URL,
		AnalysisServiceURL:       ana.URL,
	}
	c := NewClient(cfg, func() (string, error) { return "svc-token", nil })
	require.NotNil(t, c)

	err := c.TransferOwnership(context.Background(), 5, 10)
	require.NoError(t, err)

	require.Len(t, clsBodies, 1)
	require.Len(t, anaBodies, 1)
	assert.Equal(t, OwnershipTransferRequest{FromUserID: 5, ToUserID: 10}, clsBodies[0])
	assert.Equal(t, OwnershipTransferRequest{FromUserID: 5, ToUserID: 10}, anaBodies[0])
	assert.Equal(t, "Bearer svc-token", clsAuth)
	assert.Equal(t, "Bearer svc-token", anaAuth)
}

func TestClient_TransferOwnership_FailFastOnError(t *testing.T) {
	clsCalls := 0
	cls := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clsCalls++
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer cls.Close()

	anaCalls := 0
	ana := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		anaCalls++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ana.Close()

	cfg := &config.MergeConfig{
		ClassificationServiceURL: cls.URL,
		AnalysisServiceURL:       ana.URL,
	}
	c := NewClient(cfg, func() (string, error) { return "t", nil })
	require.NotNil(t, c)

	err := c.TransferOwnership(context.Background(), 1, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "classification")
	assert.Equal(t, 1, clsCalls)
	assert.Equal(t, 0, anaCalls)
}
