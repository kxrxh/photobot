package observability

import "testing"

func TestBodyFromZerologMap_prefersMessageNoDuplicateHTTPLine(t *testing.T) {
	m := map[string]interface{}{
		"message":          "HTTP GET /api/v1/analyses?x=1 -> 200 in 26.436ms",
		"method":           "GET",
		"url.full":         "/api/v1/analyses?x=1",
		"http.status_code": float64(200),
		"http.latency_ms":  float64(26),
		"latency":          "26.436ms",
	}
	got := bodyFromZerologMap(m)
	want := m["message"].(string)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBodyFromZerologMap_syntheticWhenNoMessage(t *testing.T) {
	m := map[string]interface{}{
		"method":           "POST",
		"url.full":         "/api/v1/x",
		"http.status_code": float64(201),
		"http.latency_ms":  float64(5),
	}
	got := bodyFromZerologMap(m)
	want := "HTTP POST /api/v1/x -> 201 in 5ms"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
