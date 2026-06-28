package storage

import (
	"strings"
	"testing"
)

func TestRewritePresignedToPublic_preservesQueryStripRawPath(t *testing.T) {
	pub, err := ParsePresignedPublicBase("https://example.com/storage/reports")
	if err != nil {
		t.Fatal(err)
	}
	internal := "http://minio:9000/reports/u1/p%20file.pdf?X-Amz-Credential=test"
	out, err := RewritePresignedToPublic(internal, pub)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "/storage/reports/reports/u1/p%20file.pdf?") {
		t.Fatalf("unexpected url: %s", out)
	}
}

func TestRewritePresignedToPublic(t *testing.T) {
	pub, err := ParsePresignedPublicBase("https://example.com/storage/reports")
	if err != nil {
		t.Fatal(err)
	}
	internal := "http://minio-reports:9000/reports/u1/out.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=x"
	out, err := RewritePresignedToPublic(internal, pub)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "https://example.com/storage/reports/reports/u1/out.pdf?") {
		t.Fatalf("unexpected url: %s", out)
	}
	if !strings.Contains(out, "X-Amz-Algorithm=AWS4-HMAC-SHA256") {
		t.Fatalf("query dropped: %s", out)
	}
	got, err := RewritePresignedToPublic(internal, nil)
	if err != nil || got != internal {
		t.Fatalf("nil public: got %q err %v", got, err)
	}
}
