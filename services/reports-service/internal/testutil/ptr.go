package testutil

import "math"

// Float64 returns a pointer to a copy of v.
func Float64(v float64) *float64 {
	return &v
}

// FiniteFloat64 returns nil if v is NaN or Inf; otherwise Float64(v).
func FiniteFloat64(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

// String returns a pointer to a copy of s.
func String(s string) *string {
	return &s
}

// Int32 returns a pointer to a copy of v.
func Int32(v int32) *int32 {
	return &v
}

// Int64 returns a pointer to a copy of v.
func Int64(v int64) *int64 {
	return &v
}

// Bool returns a pointer to a copy of v.
func Bool(v bool) *bool {
	return &v
}
