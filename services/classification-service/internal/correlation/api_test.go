package correlation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://correlation:8000")
	require.NotNil(t, client)
}

func TestClient_CalculateCorrelation_Success(t *testing.T) {
	want := []CorrelationWithTest{
		{CorrelationBase: CorrelationBase{Name: "param1", Conditions: []CorrelationCondition{}}},
		{CorrelationBase: CorrelationBase{Name: "param2", Conditions: []CorrelationCondition{}}},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/correlation-api/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))

		_ = json.NewEncoder(w).Encode(want)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &CorrelationRequest{
		Fractions:       []ObjectGroup{{Name: "test", ObjectIDs: []int{1, 2}}},
		ParameterGroups: []string{ParameterGroupAll},
	}
	got, err := client.CalculateCorrelation(context.Background(), req, "token123")
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "param1", got[0].Name)
	assert.Equal(t, "param2", got[1].Name)
}

func TestClient_CalculateCorrelation_NoToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"))
		_ = json.NewEncoder(w).Encode([]CorrelationWithTest{})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &CorrelationRequest{}
	got, err := client.CalculateCorrelation(context.Background(), req, "")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestClient_CalculateCorrelation_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.CalculateCorrelation(context.Background(), &CorrelationRequest{}, "bad-token")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestClient_CalculateCorrelation_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.CalculateCorrelation(context.Background(), &CorrelationRequest{}, "token")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrUnauthorized)
}
