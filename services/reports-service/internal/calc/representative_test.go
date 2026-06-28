package calc

import (
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/testutil"
)

func TestLinearPercentileInSorted(t *testing.T) {
	x := []float64{0.0, 1.0, 2.0, 3.0, 4.0}
	if got := LinearPercentileInSorted(x, 0); got != 0.0 {
		t.Fatalf("p=0: %v", got)
	}
	if got := LinearPercentileInSorted(x, 1); got != 4.0 {
		t.Fatalf("p=1: %v", got)
	}
	if got := LinearPercentileInSorted(x, 0.5); got != 2.0 {
		t.Fatalf("p=0.5: %v", got)
	}
}

func TestRepDistanceToCenter(t *testing.T) {
	center := map[string]float64{
		"l":         10.0,
		"w":         2.0,
		"sq":        20.0,
		"l_w":       5.0,
		"mass_1000": 1.0,
	}
	obj := analysis.Object{
		ID:       1,
		L:        testutil.Float64(10),
		W:        testutil.Float64(2),
		Sq:       testutil.Float64(20),
		Mass1000: testutil.Float64(1),
	}
	d := repDistanceToCenter(obj, center)
	if d < 0 || d > 1e-9 {
		t.Fatalf("expected ~0, got %v", d)
	}
}

func TestTypicalRepsNumericIDLikeAnalysisAPI(t *testing.T) {
	objs := []analysis.Object{
		{
			ID:       0,
			Class:    testutil.String("A"),
			L:        testutil.Float64(1),
			W:        testutil.Float64(1),
			Sq:       testutil.Float64(1),
			Mass1000: testutil.Float64(1),
		},
		{
			ID:       1,
			Class:    testutil.String("A"),
			L:        testutil.Float64(1),
			W:        testutil.Float64(1),
			Sq:       testutil.Float64(1),
			Mass1000: testutil.Float64(1),
		},
		{
			ID:       2,
			Class:    testutil.String("A"),
			L:        testutil.Float64(1),
			W:        testutil.Float64(1),
			Sq:       testutil.Float64(1),
			Mass1000: testutil.Float64(1),
		},
	}
	g := CalculateTypicalRepresentativesByClass(
		objs,
		RepresentativeOptions{PerClassLimit: 2, LowerPercentile: 0, UpperPercentile: 1},
	)
	if len(g) != 1 {
		t.Fatalf("groups: %d", len(g))
	}
	if len(g[0].Representatives) < 1 {
		t.Fatal("expected at least one rep")
	}
}

func TestTypicalRepsOneClass(t *testing.T) {
	objs := []analysis.Object{
		{
			ID:       1,
			Class:    testutil.String("A"),
			L:        testutil.Float64(1),
			W:        testutil.Float64(1),
			Sq:       testutil.Float64(1),
			Mass1000: testutil.Float64(1),
		},
		{
			ID:       2,
			Class:    testutil.String("A"),
			L:        testutil.Float64(1),
			W:        testutil.Float64(1),
			Sq:       testutil.Float64(1),
			Mass1000: testutil.Float64(1),
		},
		{
			ID:       3,
			Class:    testutil.String("A"),
			L:        testutil.Float64(1),
			W:        testutil.Float64(1),
			Sq:       testutil.Float64(1),
			Mass1000: testutil.Float64(1),
		},
	}
	g := CalculateTypicalRepresentativesByClass(
		objs,
		RepresentativeOptions{PerClassLimit: 2, LowerPercentile: 0, UpperPercentile: 1},
	)
	if len(g) != 1 {
		t.Fatalf("groups: %d", len(g))
	}
	if len(g[0].Representatives) < 1 {
		t.Fatal("expected at least one rep")
	}
}
