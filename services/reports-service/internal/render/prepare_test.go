package render

import "testing"

func TestInjectQRSVG_NoPlaceholder(t *testing.T) {
	html := "<html><body>ok</body></html>"
	got := InjectQRSVG(html, "/tmp/does-not-matter")
	if got != html {
		t.Fatalf("expected unchanged HTML when placeholder absent, got %q", got)
	}
}
