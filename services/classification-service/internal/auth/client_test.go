package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofiber/fiber/v3"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://identity.example.com", "svc-id", "secret")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.baseURL != "https://identity.example.com" {
		t.Errorf("baseURL = %q, want https://identity.example.com", client.baseURL)
	}
	if client.serviceID != "svc-id" {
		t.Errorf("serviceID = %q, want svc-id", client.serviceID)
	}
	if client.fiberClient == nil {
		t.Error("fiberClient is nil")
	}
}

func TestGetUserByMessengerID(t *testing.T) {
	want := &User{ID: 1, TelegramID: 98765, FullName: "Messenger User"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/by-messenger-id/98765" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.RawQuery != "platform=max" {
			t.Errorf("query = %q", r.URL.RawQuery)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer token123" {
			t.Errorf("Authorization = %q", auth)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "result": want})
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	got, err := client.GetUserByMessengerID(context.Background(), 98765, "max", "token123")
	if err != nil {
		t.Fatalf("GetUserByMessengerID: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %d, want %d", got.ID, want.ID)
	}
}

func TestGetUserByMessengerID_NormalizesPlatform(t *testing.T) {
	want := &User{ID: 7, TelegramID: 33333, FullName: "Normalized User"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/by-messenger-id/33333" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.RawQuery != "platform=max" {
			t.Errorf("query = %q", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "result": want})
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	got, err := client.GetUserByMessengerID(context.Background(), 33333, " MAX ", "token123")
	if err != nil {
		t.Fatalf("GetUserByMessengerID: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %d, want %d", got.ID, want.ID)
	}
}

func TestGetUserByMessengerID_RejectsInvalidPlatform(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called for invalid platform")
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	_, err := client.GetUserByMessengerID(context.Background(), 33333, "vk", "token123")
	if err == nil {
		t.Fatal("expected error for invalid platform")
	}
}

func TestLoginAsService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Header.Get("X-Grant-Type") != "client_credentials" {
			t.Errorf("X-Grant-Type = %q", r.Header.Get("X-Grant-Type"))
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["service_id"] != "svc" || body["audience"] != "auth-service" {
			t.Errorf("body = %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result": map[string]string{
				"access_token":  "access-123",
				"refresh_token": "refresh-456",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	access, refresh, err := client.LoginAsService(context.Background(), "auth-service")
	if err != nil {
		t.Fatalf("LoginAsService: %v", err)
	}
	if access != "access-123" {
		t.Errorf("access = %q, want access-123", access)
	}
	if refresh != "refresh-456" {
		t.Errorf("refresh = %q, want refresh-456", refresh)
	}
}

func TestRefreshTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/refresh" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["refresh_token"] != "old-refresh" {
			t.Errorf("refresh_token = %q", body["refresh_token"])
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result": map[string]string{
				"access_token":  "new-access",
				"refresh_token": "new-refresh",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	access, refresh, err := client.RefreshTokens(context.Background(), "old-refresh")
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}
	if access != "new-access" {
		t.Errorf("access = %q, want new-access", access)
	}
	if refresh != "new-refresh" {
		t.Errorf("refresh = %q, want new-refresh", refresh)
	}
}

func TestUpdateJWKS(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	jwk := jose.JSONWebKey{Key: key.Public(), KeyID: "test-kid", Algorithm: string(jose.RS256)}
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/.well-known/jwks.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	err = client.UpdateJWKS(context.Background())
	if err != nil {
		t.Fatalf("UpdateJWKS: %v", err)
	}
}

func TestUpdateJWKS_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	err := client.UpdateJWKS(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	privateJWK := jose.JSONWebKey{Key: key, KeyID: "test-kid", Algorithm: string(jose.RS256)}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateJWK}, nil)
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	publicJWK := jose.JSONWebKey{
		Key:       key.Public(),
		KeyID:     "test-kid",
		Algorithm: string(jose.RS256),
	}
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{publicJWK}}

	exp := time.Now().Add(time.Hour)
	tokenStr, err := jwt.Signed(signer).
		Claims(jwt.Claims{Subject: "user", Expiry: jwt.NewNumericDate(exp)}).
		Claims(CustomClaims{UserID: 42, TelegramID: 12345, Roles: []string{"admin"}}).
		Serialize()
	if err != nil {
		t.Fatalf("Serialize token: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	err = client.UpdateJWKS(context.Background())
	if err != nil {
		t.Fatalf("UpdateJWKS: %v", err)
	}

	resp, err := client.ValidateToken(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if !resp.Valid {
		t.Error("expected Valid=true")
	}
	if resp.Identity == nil {
		t.Fatal("Identity is nil")
	}
	if resp.Identity.UserID != 42 {
		t.Errorf("UserID = %d, want 42", resp.Identity.UserID)
	}
	if resp.Identity.TelegramID != 12345 {
		t.Errorf("TelegramID = %d, want 12345", resp.Identity.TelegramID)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	client := NewClient("http://localhost:0", "svc", "secret")
	resp, err := client.ValidateToken(context.Background(), "invalid.jwt.token")
	if err == nil {
		t.Fatal("expected error")
	}
	if resp != nil && resp.Valid {
		t.Error("expected Valid=false")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	privateJWK := jose.JSONWebKey{Key: key, KeyID: "exp-kid", Algorithm: string(jose.RS256)}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateJWK}, nil)
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	publicJWK := jose.JSONWebKey{Key: key.Public(), KeyID: "exp-kid", Algorithm: string(jose.RS256)}
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{publicJWK}}

	exp := time.Now().Add(-time.Hour)
	tokenStr, err := jwt.Signed(signer).
		Claims(jwt.Claims{Subject: "user", Expiry: jwt.NewNumericDate(exp)}).
		Claims(CustomClaims{UserID: 1, Roles: []string{"user"}}).
		Serialize()
	if err != nil {
		t.Fatalf("Serialize token: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	err = client.UpdateJWKS(context.Background())
	if err != nil {
		t.Fatalf("UpdateJWKS: %v", err)
	}

	resp, err := client.ValidateToken(context.Background(), tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if resp != nil && resp.Valid {
		t.Error("expected Valid=false")
	}
}

func TestValidateToken_JWKSNotInitialized(t *testing.T) {
	client := NewClient("http://localhost:0", "svc", "secret")
	resp, err := client.ValidateToken(
		context.Background(),
		"eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.x",
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if resp != nil && resp.Valid {
		t.Error("expected Valid=false")
	}
}

func TestValidateToken_RefreshJWKSAndRetry(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	privateJWK := jose.JSONWebKey{Key: key, KeyID: "retry-kid", Algorithm: string(jose.RS256)}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateJWK}, nil)
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	publicJWK := jose.JSONWebKey{
		Key:       key.Public(),
		KeyID:     "retry-kid",
		Algorithm: string(jose.RS256),
	}
	validJWKS := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{publicJWK}}

	tokenStr, err := jwt.Signed(signer).
		Claims(jwt.Claims{Subject: "user", Expiry: jwt.NewNumericDate(time.Now().Add(time.Hour))}).
		Claims(CustomClaims{UserID: 99, Roles: []string{"user"}}).
		Serialize()
	if err != nil {
		t.Fatalf("Serialize token: %v", err)
	}

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{}})
			return
		}
		_ = json.NewEncoder(w).Encode(validJWKS)
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	_ = client.UpdateJWKS(context.Background())

	resp, err := client.ValidateToken(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("ValidateToken after refresh: %v", err)
	}
	if !resp.Valid || resp.Identity == nil || resp.Identity.UserID != 99 {
		t.Errorf("expected valid token, got Valid=%v Identity=%v", resp.Valid, resp.Identity)
	}
}

func TestDoRequest_IdentityServiceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "internal error",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	_, _, err := client.RefreshTokens(context.Background(), "refresh-token")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "auth service error: internal error" {
		t.Errorf("err = %q", err.Error())
	}
}

func TestStart(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	jwk := jose.JSONWebKey{Key: key.Public(), KeyID: "kid", Algorithm: string(jose.RS256)}
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = client.Start(ctx)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
}

func TestStart_JWKSError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(fiber.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "svc", "secret")
	err := client.Start(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
