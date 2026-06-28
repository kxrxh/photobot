package calc

import (
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
)

func TestIsDefectClassName(t *testing.T) {
	if !IsDefectClassName("broken") || !IsDefectClassName("  ломанные ") {
		t.Fatal("expected defect class names")
	}
	if IsDefectClassName("good") {
		t.Fatal("good class should not be defect")
	}
}

func TestHasAnyClassMetrics(t *testing.T) {
	if HasAnyClassMetrics(nil) {
		t.Fatal("nil")
	}
	if HasAnyClassMetrics(&ClassStatistics{Count: 1}) {
		t.Fatal("count only")
	}
	if !HasAnyClassMetrics(&ClassStatistics{L: &ClassFieldStats{}}) {
		t.Fatal("expected metrics")
	}
}

func TestCalculateClassStatisticsResultGroups(t *testing.T) {
	c1, c2 := "A", "B"
	l1, l2 := 1.0, 2.0
	res := CalculateClassStatisticsResult([]analysis.Object{
		{Class: &c1, L: &l1},
		{Class: &c2, L: &l2},
	})
	if len(res) != 2 {
		t.Fatalf("expected 2 groups, got %d: %+v", len(res), res)
	}
	if res["A"] == nil || res["A"].L == nil || res["A"].L.Min != l1 {
		t.Fatalf("class A: %+v", res["A"])
	}
}

func TestCalculateClassStatisticsResultBrokenSkipsMeasurements(t *testing.T) {
	c := "broken"
	l := 5.0
	res := CalculateClassStatisticsResult([]analysis.Object{{Class: &c, L: &l}})
	bs := res["broken"]
	if bs == nil || bs.Count != 1 {
		t.Fatalf("broken group missing: %+v", bs)
	}
	if bs.L != nil {
		t.Fatal("defect class should omit dimension stats")
	}
}
