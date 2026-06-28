package sharelink

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"testing"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"github.com/stretchr/testify/require"
)

func TestSignVerify_roundTrip(t *testing.T) {
	secret := []byte("test-secret-key-32bytes-long!!")
	aid := "550e8400-e29b-41d4-a716-446655440000"
	exp := int64(1893456000) // ~2030
	sig := Sign(secret, aid, exp)
	require.NotEmpty(t, sig)

	now := time.Unix(exp-3600, 0)
	err := Verify(secret, aid, "1893456000", sig, now, 60*time.Second)
	require.NoError(t, err)
}

func TestVerify_expired(t *testing.T) {
	secret := []byte("test-secret-key-32bytes-long!!")
	aid := "550e8400-e29b-41d4-a716-446655440000"
	exp := int64(1000000000)
	sig := Sign(secret, aid, exp)
	err := Verify(secret, aid, "1000000000", sig, time.Now(), 60*time.Second)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestVerify_reportsServicePackQueryFormat(t *testing.T) {
	secret := []byte("hmac-secret")
	aid := "analysis-uuid"
	const reportPackScope = "v1|"
	exp := time.Now().Unix() + 3600
	expStr := strconv.FormatInt(exp, 10)
	msg := reportPackScope + aid + "|all|" + expStr
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(msg))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	err := Verify(secret, aid, expStr, sig, time.Now(), time.Minute)
	require.NoError(t, err)
}

func TestIdentityOwnsAnalysis(t *testing.T) {
	tg := int64(42)
	id := &auth.Identity{TelegramID: &tg}
	require.True(t, IdentityOwnsAnalysis(id, 42))
	require.False(t, IdentityOwnsAnalysis(id, 99))
}

func TestIdentityHasServiceRole(t *testing.T) {
	require.True(t, IdentityHasServiceRole(&auth.Identity{Roles: []string{"service"}}))
	require.False(t, IdentityHasServiceRole(&auth.Identity{Roles: []string{"user"}}))
}
