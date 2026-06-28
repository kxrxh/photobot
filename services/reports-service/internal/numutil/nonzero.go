package numutil

import "math"

const FloatEpsilon = 1e-12

func NonZero(f float64) bool {
	return math.Abs(f) >= FloatEpsilon
}
