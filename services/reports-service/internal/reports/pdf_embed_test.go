package reports

import (
	"strings"
	"testing"
)

func TestEmbeddedPDFAssets(t *testing.T) {
	if !strings.Contains(PDFWarmupHTML, "report-shell") {
		t.Fatalf("warmup html missing report-shell: %q", PDFWarmupHTML)
	}
	if !strings.Contains(printReadyJS, "document.fonts.ready") {
		t.Fatalf("print_ready.js missing font wait: %q", printReadyJS)
	}
	if !strings.Contains(printReadyJS, "(async () =>") {
		t.Fatalf("print_ready.js should contain async IIFE expression")
	}
}
