package numutil

import (
	"math"
	"testing"
)

func TestFormatFloat(t *testing.T) {
	if g, w := FormatFloat(1.2, 2), "1.20"; g != w {
		t.Fatalf("got %q want %q", g, w)
	}
}

func TestRoundPlaces(t *testing.T) {
	if v := RoundPlaces(1.234, 2); v != 1.23 {
		t.Fatalf("RoundPlaces 2: got %v", v)
	}
	if v := Round3(1.2345); v != 1.235 {
		t.Fatalf("Round3: got %v", v)
	}
}

func TestIsFinite(t *testing.T) {
	if !IsFinite(0) || !IsFinite(-1.5) {
		t.Fatal("expected finite")
	}
	if IsFinite(math.NaN()) || IsFinite(math.Inf(1)) || IsFinite(math.Inf(-1)) {
		t.Fatal("expected non-finite")
	}
}

func TestNonZero(t *testing.T) {
	if NonZero(0) || NonZero(1e-13) || NonZero(-1e-13) {
		t.Fatal("expected zero or below epsilon as not non-zero")
	}
	if !NonZero(1e-11) || !NonZero(-2.5) {
		t.Fatal("expected above-epsilon as non-zero")
	}
}
