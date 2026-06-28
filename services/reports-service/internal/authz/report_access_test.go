package authz

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestSubject_CanPresignCachedReport_ownerByTelegram(t *testing.T) {
	tid := int64(42)
	s := Subject{TelegramID: &tid}
	if !s.CanPresignCachedReport(42) {
		t.Fatal("expected owner match")
	}
	if s.CanPresignCachedReport(99) {
		t.Fatal("expected no access")
	}
}

func TestSubject_CanPresignCachedReport_service(t *testing.T) {
	s := Subject{Roles: []string{"service"}}
	if !s.CanPresignCachedReport(999) {
		t.Fatal("service role should allow")
	}
}

func TestSubjectFromMapClaims(t *testing.T) {
	tid := float64(7)
	claims := jwt.MapClaims{
		"telegram_id": tid,
		"roles":       []any{"user", "service"},
	}
	s := SubjectFromMapClaims(claims)
	if s.TelegramID == nil || *s.TelegramID != 7 {
		t.Fatalf("telegram: %+v", s.TelegramID)
	}
	if !s.HasServiceRole() {
		t.Fatal("roles")
	}
}

func TestSubject_CanPresignCachedReport_serviceIDClaim(t *testing.T) {
	svc := "analysis-service"
	s := Subject{ServiceID: &svc}
	if !s.CanPresignCachedReport(12345) {
		t.Fatal("service_id claim should allow presign")
	}
}
