package reports

import (
	"strings"
	"testing"
	"time"
)

func TestReportPacker_Query_EmptySecret(t *testing.T) {
	p := NewReportPacker("", 3600)
	if q := p.Query("id-1"); q != "" {
		t.Fatalf("expected empty query, got %q", q)
	}
}

func TestReportPacker_Query_Signed(t *testing.T) {
	p := NewReportPacker("secret", 60)
	q := p.Query("analysis-uuid")
	if !strings.HasPrefix(q, "?exp=") || !strings.Contains(q, "&sig=") {
		t.Fatalf("unexpected query: %q", q)
	}
}

func TestVerifyPackQuery_roundTrip(t *testing.T) {
	const secret = "k"
	id := "550e8400-e29b-41d4-a716-446655440000"
	p := NewReportPacker(secret, 3600)
	q := p.Query(id)
	// q is ?exp=N&sig=S
	q = strings.TrimPrefix(q, "?")
	var exp, sig string
	for _, part := range strings.Split(q, "&") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "exp":
			exp = kv[1]
		case "sig":
			sig = kv[1]
		}
	}
	if err := VerifyPackQuery(secret, id, exp, sig, time.Now(), time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := VerifyPackQuery(secret, id, exp, sig+"x", time.Now(), time.Hour); err == nil {
		t.Fatal("expected err")
	}
}
