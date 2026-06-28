package reports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
)

func TestReportContextFromAnalysisFixtureRoundTrip(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	fixture := filepath.Join(root, "cmd", "dummypdf", "testdata", "local-dummy-analysis.json")
	raw, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var env analysis.AnalysisAPIResponse
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !env.Success {
		t.Fatal("fixture success=false")
	}
	rc := ReportContextFromAnalysis(&env.Result, "test-id")
	if rc.AnalysisID != "test-id" {
		t.Fatalf("AnalysisID = %q", rc.AnalysisID)
	}
	if !strings.Contains(rc.DateTime, "2026") {
		t.Fatalf("DateTime = %q", rc.DateTime)
	}
	if rc.LenAvg == "-" || rc.WidthAvg == "-" {
		t.Fatalf("expected len/width from params, got len=%q width=%q", rc.LenAvg, rc.WidthAvg)
	}
	if rc.UserID != 9001 {
		t.Fatalf("UserID = %d", rc.UserID)
	}
}

func TestReportContextFromAnalysisNil(t *testing.T) {
	rc := ReportContextFromAnalysis(nil, "x")
	if rc.AnalysisID != "x" || rc.Product != "-" {
		t.Fatalf("unexpected context: %+v", rc)
	}
}
