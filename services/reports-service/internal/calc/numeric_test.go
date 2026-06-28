package calc

import (
	"math"
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/testutil"
)

func TestNumericFieldValueLW(t *testing.T) {
	obj := analysis.Object{L: testutil.Float64(6), W: testutil.Float64(2)}
	v := NumericFieldValue(obj, "l_w")
	if v == nil || *v != 3.0 {
		t.Fatalf("l_w: got %v", v)
	}
}

func TestNumericFieldValueLWFromLWField(t *testing.T) {
	obj := analysis.Object{LW: testutil.Float64(2.5)}
	v := NumericFieldValue(obj, "l_w")
	if v == nil || *v != 2.5 {
		t.Fatalf("l_w from field: got %v", v)
	}
}

func TestNumericFieldValueThicknessZeroOmitted(t *testing.T) {
	obj := analysis.Object{T: testutil.Float64(0)}
	if v := NumericFieldValue(obj, "t"); v != nil {
		t.Fatalf("t=0 should be treated as not calculated, got %v", v)
	}
	obj2 := analysis.Object{T: testutil.Float64(1.2)}
	if v := NumericFieldValue(obj2, "t"); v == nil || *v != 1.2 {
		t.Fatalf("t>0: got %v", v)
	}
}

func TestNumericFieldValueUnknownField(t *testing.T) {
	if NumericFieldValue(analysis.Object{L: testutil.Float64(1)}, "zzz") != nil {
		t.Fatal("unknown field should return nil")
	}
}

func TestNumericFieldValueLWWZeroWidth(t *testing.T) {
	obj := analysis.Object{L: testutil.Float64(6), W: testutil.Float64(0)}
	if NumericFieldValue(obj, "l_w") != nil {
		t.Fatal("division by zero width should yield nil when LW fallback missing")
	}
}

func TestNumericFieldValueRejectsNanInPtr(t *testing.T) {
	nan := math.NaN()
	obj := analysis.Object{L: &nan}
	if NumericFieldValue(obj, "l") != nil {
		t.Fatal("NaN leaf value should be rejected")
	}
}

func TestNumericFieldValueRGBAndHSVParts(t *testing.T) {
	r := 7.0
	if x := NumericFieldValue(analysis.Object{R: &r}, "r"); x == nil || *x != 7 {
		t.Fatalf("r: got %v", x)
	}
	h, s, v := 0.1, 0.2, 0.3
	obj := analysis.Object{H: &h, S: &s, V: &v}
	if x := NumericFieldValue(obj, "s"); x == nil || *x != 0.2 {
		t.Fatalf("s: got %v", x)
	}
	if x := NumericFieldValue(obj, "v"); x == nil || *x != 0.3 {
		t.Fatalf("v: got %v", x)
	}
}

func TestNumericFieldValueMass1000(t *testing.T) {
	m := 42.5
	obj := analysis.Object{Mass1000: &m}
	v := NumericFieldValue(obj, "mass_1000")
	if v == nil || *v != 42.5 {
		t.Fatalf("mass_1000: %v", v)
	}
}
