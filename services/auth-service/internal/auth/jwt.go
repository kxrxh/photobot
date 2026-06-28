package auth

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// KeyIDMismatchError is returned when the JWT header kid does not match the active verification key.
type KeyIDMismatchError struct {
	Expected string
	Got      string
}

func (e *KeyIDMismatchError) Error() string {
	return fmt.Sprintf("invalid key ID: expected %s, got %s", e.Expected, e.Got)
}

var keyIDMismatchRegexp = regexp.MustCompile(
	`invalid key ID: expected ([^,]+), got ([^\s]+)`,
)

const (
	issuer = "auth-service"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID     *int32    `json:"user_id,omitempty"`
	ServiceID  *string   `json:"service_id,omitempty"`
	TelegramID *int64    `json:"telegram_id,omitempty"`
	MaxID      *int64    `json:"max_id,omitempty"`
	Roles      []string  `json:"roles,omitempty"`
	GTY        GrantType `json:"gty,omitempty"`
	Type       TokenType `json:"type,omitempty"`
	jwt.RegisteredClaims
}

// GenerationParams contains parameters for token generation
type GenerationParams struct {
	UserID     *int32    `json:"user_id"`
	ServiceID  *string   `json:"service_id"`
	TelegramID *int64    `json:"telegram_id"`
	MaxID      *int64    `json:"max_id"`
	Roles      []string  `json:"roles"`
	GTY        GrantType `json:"gty,omitempty"`
	JTI        string    `json:"jti,omitempty"`
	Audience   string    `json:"aud,omitempty"`
}

// GenerateJWT generates a JWT token using RS256 with the RSA private key
func GenerateJWT(params *GenerationParams, tokenType TokenType, ttl time.Duration) (string, error) {
	keyManager := GetKeyManager()
	privateKey := keyManager.GetSigningKey()
	if privateKey == nil {
		return "", errors.New("private key not initialized")
	}

	now := time.Now()
	expirationTime := now.Add(ttl)

	roleNames := make([]string, len(params.Roles))
	copy(roleNames, params.Roles)

	claims := &Claims{
		Roles: roleNames,
		GTY:   params.GTY,
		Type:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}

	if params.Audience != "" {
		claims.Audience = jwt.ClaimStrings{params.Audience}
	}

	if params.UserID != nil {
		claims.UserID = params.UserID
	}
	if params.TelegramID != nil {
		claims.TelegramID = params.TelegramID
	}
	if params.MaxID != nil {
		claims.MaxID = params.MaxID
	}
	if params.ServiceID != nil {
		claims.ServiceID = params.ServiceID
	}

	if params.JTI != "" {
		claims.ID = params.JTI
	}

	switch {
	case params.UserID != nil:
		claims.Subject = fmt.Sprintf("user:%d", *params.UserID)
	case params.ServiceID != nil:
		claims.Subject = "service:" + *params.ServiceID
	case params.GTY == GrantTypePassword:
		claims.Subject = "admin"
	}

	if claims.Type == "" {
		claims.Type = AccessToken
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyManager.GetKeyID()

	return token.SignedString(privateKey)
}

// IssueMergeToken returns a short-lived service JWT for internal merge calls (this service's signing key).
func IssueMergeToken() (string, error) {
	serviceId := "auth-service"
	return GenerateJWT(&GenerationParams{
		ServiceID: &serviceId,
		Roles:     []string{ServiceRole},
		GTY:       GrantTypeService,
	}, AccessToken, 5*time.Minute)
}

// ParseJWT parses and validates a JWT token using RS256 with the RSA public key
func ParseJWT(tokenString string) (*Claims, error) {
	keyManager := GetKeyManager()
	publicKey := keyManager.GetPublicKey()
	expectedKeyID := keyManager.GetKeyID()
	if publicKey == nil {
		return nil, errors.New("public key not initialized")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf(
				"invalid signing algorithm: expected RS256, got %s",
				token.Method.Alg(),
			)
		}

		if kid, ok := token.Header["kid"].(string); ok && kid != expectedKeyID {
			return nil, &KeyIDMismatchError{Expected: expectedKeyID, Got: kid}
		}

		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", issuer, claims.Issuer)
	}

	return claims, nil
}

// UnverifiedClaims parses JWT claims without signature verification.
// For diagnostic logging only; never use for authorization decisions.
func UnverifiedClaims(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// KeyIDMismatchDetails returns expected and token kid when err wraps KeyIDMismatchError
// or matches its legacy string form (e.g. after logging-only copies).
func KeyIDMismatchDetails(err error) (expected, tokenKID string, ok bool) {
	if err == nil {
		return "", "", false
	}
	var km *KeyIDMismatchError
	if errors.As(err, &km) {
		return km.Expected, km.Got, true
	}
	m := keyIDMismatchRegexp.FindStringSubmatch(err.Error())
	if len(m) == 3 {
		return m[1], m[2], true
	}
	return "", "", false
}

// JWTRejectReason returns a short machine-readable reason for a failed ParseJWT.
func JWTRejectReason(err error) string {
	if err == nil {
		return ""
	}
	if _, _, ok := KeyIDMismatchDetails(err); ok {
		return "invalid_key_id"
	}
	switch {
	case errors.Is(err, jwt.ErrTokenMalformed):
		return "malformed"
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return "signature_invalid"
	case errors.Is(err, jwt.ErrTokenExpired):
		return "expired"
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return "not_valid_yet"
	case errors.Is(err, jwt.ErrTokenInvalidAudience):
		return "invalid_audience"
	case errors.Is(err, jwt.ErrTokenInvalidIssuer):
		return "invalid_issuer"
	case errors.Is(err, jwt.ErrTokenInvalidClaims):
		return "invalid_claims"
	case errors.Is(err, jwt.ErrTokenUnverifiable):
		return "unverifiable"
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "invalid issuer:"):
		return "invalid_issuer"
	case strings.Contains(msg, "unexpected signing method:"),
		strings.Contains(msg, "invalid signing algorithm:"):
		return "invalid_signing_method"
	case strings.Contains(msg, "invalid key ID:"):
		return "invalid_key_id"
	case strings.Contains(msg, "token is unverifiable"),
		strings.Contains(msg, "error while executing keyfunc"):
		return "unverifiable"
	case strings.Contains(msg, "public key not initialized"):
		return "keys_uninitialized"
	case strings.Contains(msg, "invalid token"):
		return "token_invalid"
	default:
		return "parse_failed"
	}
}
