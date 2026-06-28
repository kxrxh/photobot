package storage

import (
	"strings"
	"testing"
)

func TestRewritePresignedURLForPublicAccess_preservesQueryStripRawPath(t *testing.T) {
	internal := "http://minio:9000/analysis/u1/p%20file.jpg?X-Amz-Credential=test"
	pub := "https://example.com/storage/analysis"
	out, err := rewritePresignedURLForPublicAccess(internal, pub)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "/storage/analysis/analysis/u1/p%20file.jpg?") {
		t.Fatalf("unexpected url: %s", out)
	}
}

func TestRewritePresignedURLForPublicAccess(t *testing.T) {
	internal := "http://minio-analysis:9000/analysis/u1/source/image_0.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=x"
	pub := "https://example.com/storage/analysis"
	out, err := rewritePresignedURLForPublicAccess(internal, pub)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(
		out,
		"https://example.com/storage/analysis/analysis/u1/source/image_0.jpg?",
	) {
		t.Fatalf("unexpected url: %s", out)
	}
	if !strings.Contains(out, "X-Amz-Algorithm=AWS4-HMAC-SHA256") {
		t.Fatalf("query dropped: %s", out)
	}
	got, err := rewritePresignedURLForPublicAccess(internal, " \t")
	if err != nil || got != internal {
		t.Fatalf("empty external: got %q err %v", got, err)
	}
}
