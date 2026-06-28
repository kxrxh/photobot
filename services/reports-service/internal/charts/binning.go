package charts

import (
	"fmt"
	"math"
)

func adaptiveBinSize(min, max float64, preferredBins int) float64 {
	if preferredBins <= 0 {
		preferredBins = 15
	}
	if min >= max || math.IsNaN(min) || math.IsNaN(max) || math.IsInf(min, 0) ||
		math.IsInf(max, 0) {
		return 1
	}
	rng := max - min
	binSize := rng / float64(preferredBins)
	if binSize <= 0 {
		return 1
	}
	magnitude := math.Pow(10, math.Floor(math.Log10(binSize)))
	normalized := binSize / magnitude
	var nice float64
	switch {
	case normalized <= 1:
		nice = 1
	case normalized <= 2:
		nice = 2
	case normalized <= 5:
		nice = 5
	default:
		nice = 10
	}
	return nice * magnitude
}

func binLabel(start, end float64) string {
	return fmt.Sprintf("%.1f-%.1f", start, end)
}

func validPositiveRange(min, max float64) bool {
	return !math.IsNaN(min) && !math.IsNaN(max) && !math.IsInf(min, 0) && !math.IsInf(max, 0) &&
		min < max && min >= 0
}
