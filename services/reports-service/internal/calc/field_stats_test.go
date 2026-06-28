package calc

import (
	"math"
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/testutil"
)

func TestFieldStatisticsFromValues(t *testing.T) {
	s := FieldStatisticsFromValues([]float64{1, 2, 3, 4})
	if s == nil {
		t.Fatal("nil stats")
	}
	if s.Min != 1 || s.Max != 4 {
		t.Fatalf("min max: %+v", s)
	}
	if math.Abs(s.Avg-2.5) > 1e-9 {
		t.Fatalf("avg: %v", s.Avg)
	}
}

func TestCalculateFieldStatisticsResultEmpty(t *testing.T) {
	if len(CalculateFieldStatisticsResult(nil)) != 0 {
		t.Fatal("expected empty map")
	}
}

func TestFieldStatisticsFromValuesEmptyReturnsNil(t *testing.T) {
	if FieldStatisticsFromValues(nil) != nil || FieldStatisticsFromValues([]float64{}) != nil {
		t.Fatal("empty values should yield nil statistics")
	}
}

func TestFieldStatisticsSingleValue(t *testing.T) {
	s := FieldStatisticsFromValues([]float64{42})
	if s == nil {
		t.Fatal("nil stats")
	}
	if s.Min != 42 || s.Max != 42 || s.Avg != 42 || s.Med != 42 {
		t.Fatalf("single value mismatch: %+v", s)
	}
}

func TestFieldStatisticsForObjects(t *testing.T) {
	l1, l2 := 10.0, 20.0
	objs := []analysis.Object{{L: &l1}, {L: &l2}}
	st := FieldStatisticsForObjects(objs, "l")
	if st == nil || st.Min != 10 || st.Max != 20 {
		t.Fatalf("unexpected field stats for l: %+v", st)
	}
	if FieldStatisticsForObjects(objs, "unknown_field") != nil {
		t.Fatal("unknown field key should yield nil")
	}
}

func TestCalculateFieldStatisticsResultPopulated(t *testing.T) {
	l, w, sq := 1.0, 2.0, 4.0
	ar, ag := 0.1, 0.2
	objs := []analysis.Object{{
		L: &l, W: &w, Sq: &sq,
		R: &ar, G: &ag,
		H: testutil.Float64(0.5), S: testutil.Float64(0.5), V: testutil.Float64(0.5),
		LW: &l,
	}}
	res := CalculateFieldStatisticsResult(objs)
	if res["l"] == nil || res["w"] == nil || res["sq"] == nil || res["l_w"] == nil {
		t.Fatalf("expected l,w,sq,l_w entries: %+v", res)
	}
	if res["t"] != nil {
		t.Fatal("unset thickness should not appear in aggregate")
	}
}
