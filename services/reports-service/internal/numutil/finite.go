package numutil

import "math"

func IsFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}
