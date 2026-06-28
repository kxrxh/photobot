package charts

import (
	"strings"
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/testutil"
)

func TestBuildDistributionSVGMapEmptyObjects(t *testing.T) {
	aggs := map[string]*FieldAgg{"l": {Min: 1, Max: 5, Stddev: 0.1, Skew: 0}}
	out := BuildDistributionSVGMap(ChartRanges{LenMin: "1", LenMax: "10"}, nil, aggs)
	if len(out) != 0 {
		t.Fatalf("expected empty map for no objects, got %d keys", len(out))
	}
}

func TestBuildDistributionSVGMapProducesLengthChart(t *testing.T) {
	objs := []analysis.Object{{L: testutil.Float64(5)}}
	aggs := map[string]*FieldAgg{
		"l": {Min: 1, Max: 9, Stddev: 0.2, Skew: 0.1},
	}
	ranges := ChartRanges{LenMin: "1", LenMax: "10"}
	out := BuildDistributionSVGMap(ranges, objs, aggs)
	svg, ok := out[KeyLength]
	if !ok || !strings.Contains(svg, "<svg") {
		t.Fatalf("expected SVG for length distribution, got ok=%v len=%d", ok, len(svg))
	}
}

func TestBuildDistributionSVGMapMass1000(t *testing.T) {
	objs := []analysis.Object{{Mass1000: testutil.Float64(32)}}
	out := BuildDistributionSVGMap(ChartRanges{}, objs, nil)
	svg, ok := out[KeyMass1000]
	if !ok || !strings.Contains(svg, "<svg") {
		t.Fatalf("expected mass chart, got ok=%v", ok)
	}
}
