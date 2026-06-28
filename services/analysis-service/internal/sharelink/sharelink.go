package sharelink

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
)

const ScopeAll = "all"

const canonicalPrefix = "v1|"

var ErrInvalid = errors.New("invalid share link")

func CanonicalV1(analysisID string, expUnix int64) string {
	return fmt.Sprintf("%s%s|%s|%d", canonicalPrefix, analysisID, ScopeAll, expUnix)
}

func Sign(secret []byte, analysisID string, expUnix int64) string {
	if len(secret) == 0 {
		return ""
	}
	msg := CanonicalV1(analysisID, expUnix)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func Verify(
	secret []byte,
	analysisID, expQ, sigQ string,
	now time.Time,
	maxSkew time.Duration,
) error {
	if len(secret) == 0 {
		return ErrInvalid
	}
	expQ = strings.TrimSpace(expQ)
	sigQ = strings.TrimSpace(sigQ)
	if expQ == "" || sigQ == "" {
		return ErrInvalid
	}
	expUnix, err := strconv.ParseInt(expQ, 10, 64)
	if err != nil || expUnix < 0 {
		return ErrInvalid
	}
	deadline := time.Unix(expUnix, 0)
	if now.After(deadline.Add(maxSkew)) {
		return ErrInvalid
	}
	expected := Sign(secret, analysisID, expUnix)
	sigBytes, err := base64.RawURLEncoding.DecodeString(sigQ)
	if err != nil {
		return ErrInvalid
	}
	expBytes, err := base64.RawURLEncoding.DecodeString(expected)
	if err != nil || len(sigBytes) != len(expBytes) {
		return ErrInvalid
	}
	if !hmac.Equal(sigBytes, expBytes) {
		return ErrInvalid
	}
	return nil
}

func HasShareQuery(expQ, sigQ string) bool {
	return strings.TrimSpace(expQ) != "" && strings.TrimSpace(sigQ) != ""
}

func IdentityHasServiceRole(id *auth.Identity) bool {
	if id == nil {
		return false
	}
	return slices.Contains(id.Roles, "service")
}

func IdentityOwnsAnalysis(id *auth.Identity, analysisUserID int64) bool {
	if id == nil {
		return false
	}
	if id.TelegramID != nil && *id.TelegramID == analysisUserID {
		return true
	}
	if id.MaxID != nil && *id.MaxID == analysisUserID {
		return true
	}
	return false
}
