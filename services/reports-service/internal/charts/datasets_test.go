package charts

import (
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/testutil"
)

func TestBinsForFieldEmptyInputs(t *testing.T) {
	obj := analysis.Object{L: testutil.Float64(5)}
	if len(BinsForField(nil, "l", 0, 10)) != 0 {
		t.Fatal("nil objects should yield empty bins")
	}
	if len(BinsForField([]analysis.Object{obj}, "l", 10, 5)) != 0 {
		t.Fatal("invalid range min>=max should yield empty bins")
	}
	if len(BinsForField([]analysis.Object{obj}, "l", -1, 5)) != 0 {
		t.Fatal("negative min should fail validPositiveRange")
	}
}

func TestBinsForFieldCountsInRange(t *testing.T) {
	objs := []analysis.Object{
		{L: testutil.Float64(1)},
		{L: testutil.Float64(2)},
		{L: testutil.Float64(2)},
	}
	bins := BinsForField(objs, "l", 0, 10)
	var total int
	for _, n := range bins {
		total += n
	}
	if total != 3 {
		t.Fatalf("expected 3 values binned, got total=%d bins=%+v", total, bins)
	}
}

func TestBinsForMass1000Buckets(t *testing.T) {
	if len(BinsForMass1000(nil)) != 0 {
		t.Fatal("nil objects should yield empty map")
	}
	mLow := 25.0
	mMid := 29.0
	mHigh := 35.0
	out := BinsForMass1000([]analysis.Object{
		{Mass1000: &mLow},
		{Mass1000: &mMid},
		{Mass1000: &mHigh},
	})
	if out["<28"] != 1 {
		t.Fatalf("expected one <28 bin, got %+v", out)
	}
	if out[binLabel(28, 30)] != 1 {
		t.Fatalf("expected 28-30 bin, got %+v", out)
	}
}
