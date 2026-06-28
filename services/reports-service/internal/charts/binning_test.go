package charts

import (
	"math"
	"testing"
)

func TestValidPositiveRange(t *testing.T) {
	t.Run("valid_finite_positive_range", func(t *testing.T) {
		if !validPositiveRange(7, 12) {
			t.Fatal("expected finite positive range to be valid")
		}
		if !validPositiveRange(0, 5) {
			t.Fatal("expected min=0 with max>min to be valid")
		}
	})
	t.Run("min_not_strictly_less_than_max", func(t *testing.T) {
		if validPositiveRange(12, 7) {
			t.Fatal("expected min>max to be invalid")
		}
		if validPositiveRange(3, 3) {
			t.Fatal("expected min==max to be invalid")
		}
	})
	t.Run("negative_min", func(t *testing.T) {
		if validPositiveRange(-1, 5) {
			t.Fatal("expected negative min to be invalid")
		}
		if validPositiveRange(-0.001, 10) {
			t.Fatal("expected slightly negative min to be invalid")
		}
	})
	t.Run("non_finite", func(t *testing.T) {
		if validPositiveRange(1, math.Inf(1)) {
			t.Fatal("expected infinite max to be invalid")
		}
		if validPositiveRange(math.Inf(-1), 5) {
			t.Fatal("expected infinite min to be invalid")
		}
		if validPositiveRange(1, math.NaN()) {
			t.Fatal("expected NaN max to be invalid")
		}
		if validPositiveRange(math.NaN(), 5) {
			t.Fatal("expected NaN min to be invalid")
		}
	})
}

func TestAdaptiveBinSize_invalidOrDegenerate_returnsOne(t *testing.T) {
	cases := []struct {
		name string
		min  float64
		max  float64
		bins int
	}{
		{"min_equals_max", 5, 5, 10},
		{"min_greater_than_max", 10, 3, 10},
		{"nan_min", math.NaN(), 10, 10},
		{"nan_max", 0, math.NaN(), 10},
		{"inf_min", math.Inf(1), 10, 10},
		{"inf_max", 0, math.Inf(-1), 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := adaptiveBinSize(tc.min, tc.max, tc.bins)
			if got != 1 {
				t.Fatalf("expected 1, got %g", got)
			}
		})
	}
}

func TestAdaptiveBinSize_preferredBins_nonPositive_usesDefault(t *testing.T) {
	a := adaptiveBinSize(0, 100, 0)
	b := adaptiveBinSize(0, 100, -3)
	if a != b {
		t.Fatalf("zero and negative preferredBins should match, got %g vs %g", a, b)
	}
	// rng=100, default 15 bins -> raw step ~6.67 -> rounds to nice step 10
	if a != 10 {
		t.Fatalf("expected nice bin size 10 for [0,100] with default bins, got %g", a)
	}
}

func TestAdaptiveBinSize_knownNiceSteps(t *testing.T) {
	cases := []struct {
		name     string
		min, max float64
		bins     int
		want     float64
	}{
		{"decade_step_ten", 0, 100, 10, 10},
		{"small_range_tenth", 0, 1, 10, 0.1},
		{"normalized_to_two", 0, 20, 10, 2},
		{"normalized_to_five", 0, 24, 10, 5},
		{"shifted_range_same_binsize", 50, 150, 10, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := adaptiveBinSize(tc.min, tc.max, tc.bins)
			if got != tc.want {
				t.Fatalf(
					"adaptiveBinSize(%g,%g,%d): got %g, want %g",
					tc.min,
					tc.max,
					tc.bins,
					got,
					tc.want,
				)
			}
		})
	}
}

func TestBinLabel(t *testing.T) {
	if got := binLabel(28, 30); got != "28.0-30.0" {
		t.Fatalf("binLabel(28,30): got %q, want %q", got, "28.0-30.0")
	}
	if got := binLabel(0, 0.5); got != "0.0-0.5" {
		t.Fatalf("binLabel(0,0.5): got %q, want %q", got, "0.0-0.5")
	}
}
