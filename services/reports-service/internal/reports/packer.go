package reports

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"
)

const reportPackScope = "v1|"

type ReportPacker struct {
	hmacKey    string
	TTLSeconds int
}

func NewReportPacker(secret string, ttlSeconds int) ReportPacker {
	return ReportPacker{hmacKey: secret, TTLSeconds: ttlSeconds}
}

func (p ReportPacker) Query(analysisID string) string {
	secret := strings.TrimSpace(p.hmacKey)
	analysisID = strings.TrimSpace(analysisID)
	if secret == "" || analysisID == "" {
		return ""
	}
	ttl := p.TTLSeconds
	if ttl <= 0 {
		ttl = 604800
	}
	exp := time.Now().Unix() + int64(ttl)
	return reportPackQueryWithExp(secret, analysisID, exp)
}

func reportPackQueryWithExp(secret, analysisID string, expUnix int64) string {
	msg := reportPackScope + analysisID + "|all|" + strconv.FormatInt(expUnix, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(msg))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return "?exp=" + strconv.FormatInt(expUnix, 10) + "&sig=" + sig
}

var ErrInvalidPackQuery = errors.New("invalid report pack query")

func VerifyPackQuery(
	secret, analysisID, expQ, sigQ string,
	now time.Time,
	maxSkew time.Duration,
) error {
	secret = strings.TrimSpace(secret)
	analysisID = strings.TrimSpace(analysisID)
	expQ = strings.TrimSpace(expQ)
	sigQ = strings.TrimSpace(sigQ)
	if secret == "" || analysisID == "" || expQ == "" || sigQ == "" {
		return ErrInvalidPackQuery
	}
	expUnix, err := strconv.ParseInt(expQ, 10, 64)
	if err != nil || expUnix < 0 {
		return ErrInvalidPackQuery
	}
	if now.After(time.Unix(expUnix, 0).Add(maxSkew)) {
		return ErrInvalidPackQuery
	}
	msg := reportPackScope + analysisID + "|all|" + strconv.FormatInt(expUnix, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(msg))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	sigBytes, err := base64.RawURLEncoding.DecodeString(sigQ)
	if err != nil {
		return ErrInvalidPackQuery
	}
	expBytes, err := base64.RawURLEncoding.DecodeString(expected)
	if err != nil || len(sigBytes) != len(expBytes) {
		return ErrInvalidPackQuery
	}
	if !hmac.Equal(sigBytes, expBytes) {
		return ErrInvalidPackQuery
	}
	return nil
}
