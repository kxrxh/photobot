package statutil

import (
	"math"
	"testing"
)

func TestMean(t *testing.T) {
	if g, w := Mean([]float64{2, 4, 6}), 4.0; math.Abs(g-w) > 1e-9 {
		t.Fatalf("Mean: got %v want %v", g, w)
	}
	if Mean(nil) != 0 {
		t.Fatalf("Mean nil: got %v", Mean(nil))
	}
}

func TestSampleStdDev(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	want := math.Sqrt(2.5)
	if g := SampleStdDev(x); math.Abs(g-want) > 1e-9 {
		t.Fatalf("SampleStdDev: got %v want %v", g, want)
	}
	if SampleStdDev([]float64{1}) != 0 {
		t.Fatal("SampleStdDev n=1")
	}
}

func TestMedian(t *testing.T) {
	if g, w := Median([]float64{3, 1, 2}), 2.0; g != w {
		t.Fatalf("Median odd: got %v want %v", g, w)
	}
	if g, w := Median([]float64{1, 2, 3, 4}), 2.5; g != w {
		t.Fatalf("Median even: got %v want %v", g, w)
	}
}

func TestMedianSorted(t *testing.T) {
	x := []float64{1, 2, 3, 4}
	if g := MedianSorted(x); g != 2.5 {
		t.Fatalf("MedianSorted: got %v", g)
	}
}

func TestSkewnessSample(t *testing.T) {
	if SkewnessSample([]float64{1, 2}) != 0 {
		t.Fatal("Skewness short")
	}
	x := []float64{-1, 0, 1}
	if g := SkewnessSample(x); math.Abs(g) > 1e-9 {
		t.Fatalf("Skewness symmetric: got %v", g)
	}
}
