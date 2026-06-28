package reports

import (
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	reportcontext "csort.ru/reports-service/internal/context"
)

func TestClassCountDistinct(t *testing.T) {
	a, b := "A", "B"
	objs := []analysis.Object{
		{Class: &a},
		{Class: &a},
		{Class: &b},
	}
	if n := classCountDistinct(objs); n != 2 {
		t.Fatalf("distinct classes: want 2 got %d", n)
	}
	objImg := "objects_img"
	if n := classCountDistinct([]analysis.Object{{Class: &objImg}}); n != 0 {
		t.Fatalf("objects_img should be excluded, got count %d", n)
	}
	uncls := "unclassified"
	if n := classCountDistinct([]analysis.Object{{Class: nil}, {Class: &uncls}}); n != 1 {
		t.Fatalf("blank and unclassified should collapse to one bucket, got %d", n)
	}
}

func TestThicknessSampleStd(t *testing.T) {
	if ThicknessSampleStd(nil) != nil {
		t.Fatal("nil input should yield nil std")
	}
	t1 := 1.0
	if ThicknessSampleStd([]analysis.Object{{T: &t1}}) != nil {
		t.Fatal("single thickness should yield nil (need >=2)")
	}
	t2, t3 := 1.0, 3.0
	std := ThicknessSampleStd([]analysis.Object{{T: &t2}, {T: &t3}})
	if std == nil || *std <= 0 {
		t.Fatalf("expected positive sample std, got %v", std)
	}
}

func TestEnrichReportContextEmptyObjects(t *testing.T) {
	rc := reportcontext.ReportContext{}
	cs, reps, dist, n := EnrichReportContext(&rc, nil, nil)
	if n != 0 || cs != nil || reps != nil || len(dist) != 0 {
		t.Fatalf("empty objects: n=%d cs=%v reps=%v dist=%v", n, cs, reps, dist)
	}
	if rc.ObjectCount != "0" {
		t.Fatalf("ObjectCount: %q", rc.ObjectCount)
	}
}

func TestEnrichReportContextSingleObject(t *testing.T) {
	l, w := 5.0, 2.0
	rc := reportcontext.ReportContext{}
	objs := []analysis.Object{{L: &l, W: &w}}
	_, _, _, n := EnrichReportContext(
		&rc,
		objs,
		&EnrichOptions{RepLimitWhenSingle: 5, RepLimitWhenMulti: 3},
	)
	if n != 1 {
		t.Fatalf("object count: want 1 got %d", n)
	}
	if rc.ObjectCount != "1" {
		t.Fatalf("ObjectCount: %q", rc.ObjectCount)
	}
}
