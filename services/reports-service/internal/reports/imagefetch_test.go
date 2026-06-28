package reports

import (
	"testing"

	"csort.ru/reports-service/internal/calc"
	reportcontext "csort.ru/reports-service/internal/context"
)

func TestCollectReportImageURLsDedupes(t *testing.T) {
	rc := reportcontext.ReportContext{Img2: []string{"https://a/x", "https://a/x"}}
	reps := []calc.RepresentativeGroup{
		{Representatives: []calc.RepresentativeCard{{ImageDataURL: "https://a/x"}}},
	}
	urls := collectReportImageURLs(&rc, reps)
	if len(urls) != 1 {
		t.Fatalf("expected 1 unique url, got %d: %#v", len(urls), urls)
	}
}
