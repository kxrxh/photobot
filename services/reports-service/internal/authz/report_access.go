package authz

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Subject struct {
	ServiceID  *string
	TelegramID *int64
	MaxID      *int64
	Roles      []string
}

func SubjectFromMapClaims(claims jwt.MapClaims) Subject {
	if claims == nil {
		return Subject{}
	}
	var sid *string
	switch v := claims["service_id"].(type) {
	case string:
		v = strings.TrimSpace(v)
		if v != "" {
			sid = &v
		}
	default:
	}
	return Subject{
		ServiceID:  sid,
		TelegramID: int64PtrFromClaim(claims["telegram_id"]),
		MaxID:      int64PtrFromClaim(claims["max_id"]),
		Roles:      rolesFromClaims(claims["roles"]),
	}
}

func (s Subject) HasServiceRole() bool {
	if s.ServiceID != nil && strings.TrimSpace(*s.ServiceID) != "" {
		return true
	}
	return slices.Contains(s.Roles, "service")
}

func (s Subject) OwnsReportObject(analysisOwnerUserID int64) bool {
	if s.TelegramID != nil && *s.TelegramID == analysisOwnerUserID {
		return true
	}
	if s.MaxID != nil && *s.MaxID == analysisOwnerUserID {
		return true
	}
	return false
}

func (s Subject) CanPresignCachedReport(analysisOwnerUserID int64) bool {
	if s.HasServiceRole() {
		return true
	}
	return s.OwnsReportObject(analysisOwnerUserID)
}

func rolesFromClaims(v any) []string {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		if strings.TrimSpace(x) == "" {
			return nil
		}
		return []string{x}
	default:
		return nil
	}
}

func int64PtrFromClaim(v any) *int64 {
	if v == nil {
		return nil
	}
	n, ok := int64FromClaim(v)
	if !ok {
		return nil
	}
	return &n
}

func int64FromClaim(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case float32:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	case int32:
		return int64(x), true
	case json.Number:
		n, err := x.Int64()
		return n, err == nil
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}
