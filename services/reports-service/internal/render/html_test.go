package render

import (
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	reportcontext "csort.ru/reports-service/internal/context"
	"csort.ru/reports-service/internal/view"
)

func TestLoadReportTemplateRender(t *testing.T) {
	root := filepath.Join("..", "..", "templates")
	if err := LoadReportTemplate(MainHTMLPath(root)); err != nil {
		t.Fatalf("LoadReportTemplate: %v", err)
	}
	out, err := RenderReportHTML(ReportHTMLData{BodyHTML: template.HTML(`<p id="x">abc-1</p>`)})
	if err != nil {
		t.Fatalf("RenderReportHTML: %v", err)
	}
	if !strings.Contains(out, "abc-1") {
		t.Fatalf("expected output to contain analysis id, got: %q", out[:min(200, len(out))])
	}
}

func TestRenderBody(t *testing.T) {
	if err := LoadReportTemplate(MainHTMLPath(filepath.Join("..", "..", "templates"))); err != nil {
		t.Fatalf("LoadReportTemplate: %v", err)
	}
	b := view.BuildBody(view.PageParams{
		Context: reportcontext.ReportContext{
			AnalysisID: "A1",
			Date:       "2025-01-01",
			UserID:     9001,
			BotLink:    "https://example.com/b",
		},
		LogoSrc: LogoRelPath,
	})
	h, err := RenderBody(b)
	if err != nil {
		t.Fatalf("RenderBody: %v", err)
	}
	s := string(h)
	if !strings.Contains(s, `class="report-header"`) || !strings.Contains(s, "A1") {
		t.Fatalf("unexpected body output: %q", s[:min(400, len(s))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
