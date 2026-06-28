package testutil

import (
	"math"
	"testing"
)

func TestFiniteFloat64NilOnNonFinite(t *testing.T) {
	if FiniteFloat64(math.NaN()) != nil || FiniteFloat64(math.Inf(1)) != nil {
		t.Fatal("NaN and Inf should yield nil")
	}
}

func TestFiniteFloat64Ok(t *testing.T) {
	p := FiniteFloat64(3.5)
	if p == nil || *p != 3.5 {
		t.Fatalf("want 3.5, got %v", p)
	}
}
