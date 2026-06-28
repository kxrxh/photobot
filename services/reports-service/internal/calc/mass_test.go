package calc

import (
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/testutil"
)

func TestCalculateMassDerivedFromObjectsEmpty(t *testing.T) {
	md := CalculateMassDerivedFromObjects(nil)
	if md.AvgMass1000 != nil || md.SumMassGrams != nil {
		t.Fatalf("expected nil derived fields, got %+v", md)
	}
	md = CalculateMassDerivedFromObjects([]analysis.Object{})
	if md.AvgMass1000 != nil || md.SumMassGrams != nil {
		t.Fatalf("expected nil derived fields for no objects, got %+v", md)
	}
}

func TestCalculateMassDerivedFromObjectsNoMass(t *testing.T) {
	objs := []analysis.Object{{L: testutil.Float64(1)}}
	md := CalculateMassDerivedFromObjects(objs)
	if md.AvgMass1000 != nil || md.SumMassGrams != nil {
		t.Fatalf("expected nil when no mass_1000, got %+v", md)
	}
}

func TestCalculateMassDerivedFromObjectsAverages(t *testing.T) {
	m1 := 1000.0
	m2 := 2000.0
	objs := []analysis.Object{
		{Mass1000: &m1},
		{Mass1000: &m2},
	}
	md := CalculateMassDerivedFromObjects(objs)
	if md.AvgMass1000 == nil || md.SumMassGrams == nil {
		t.Fatal("expected non-nil derived fields")
	}
	wantAvg := 1500.0
	if *md.AvgMass1000 != wantAvg {
		t.Fatalf("AvgMass1000: want %v got %v", wantAvg, *md.AvgMass1000)
	}
	wantSumG := 3.0 // (1000+2000)/1000
	if *md.SumMassGrams != wantSumG {
		t.Fatalf("SumMassGrams: want %v got %v", wantSumG, *md.SumMassGrams)
	}
}
