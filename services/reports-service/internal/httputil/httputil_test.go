package httputil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBytesToDataURL(t *testing.T) {
	got := BytesToDataURL([]byte{0x89, 0x50, 0x4e, 0x47}, "image/png")
	if !strings.HasPrefix(got, "data:image/png;base64,") {
		t.Fatalf("unexpected data url prefix: %q", got)
	}
}

func TestFetchMapAsDataURLs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("fake-jpeg-bytes"))
	}))
	defer srv.Close()

	client := PublicClient(time.Second, true)
	embedded := FetchMapAsDataURLs(context.Background(), client, []string{srv.URL}, 1<<20, 4, nil)
	if len(embedded) != 1 {
		t.Fatalf("expected 1 embedded image, got %d", len(embedded))
	}
	if !strings.HasPrefix(embedded[srv.URL], "data:image/") {
		t.Fatalf("expected data url, got %q", embedded[srv.URL])
	}
}

func TestFetchAsDataURLBlocksPrivateIP(t *testing.T) {
	_, err := FetchAsDataURL(
		context.Background(),
		PublicClient(time.Second, false),
		"http://127.0.0.1/secret",
		1<<20,
	)
	if err == nil {
		t.Fatal("expected block for loopback URL")
	}
}
