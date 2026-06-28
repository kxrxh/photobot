package statutil

import (
	"math"
	"sort"
)

func Mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	var s float64
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}

func SampleVarianceMean(x []float64, mean float64) float64 {
	n := len(x)
	if n < 2 {
		return 0
	}
	var sum float64
	for _, v := range x {
		d := v - mean
		sum += d * d
	}
	return sum / float64(n-1)
}

func SampleStdDevMean(x []float64, mean float64) float64 {
	return math.Sqrt(SampleVarianceMean(x, mean))
}

func SampleStdDev(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	return SampleStdDevMean(x, Mean(x))
}

func Median(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := append([]float64(nil), x...)
	sort.Float64s(s)
	return MedianSorted(s)
}

func MedianSorted(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	mid := len(x) / 2
	if len(x)%2 == 0 {
		return (x[mid-1] + x[mid]) / 2
	}
	return x[mid]
}

func SkewnessSample(x []float64) float64 {
	n := len(x)
	if n < 3 {
		return 0
	}
	mean := Mean(x)
	variance := SampleVarianceMean(x, mean)
	if variance == 0 {
		return 0
	}
	std := math.Sqrt(variance)
	nf := float64(n)
	var third float64
	for _, v := range x {
		z := (v - mean) / std
		third += z * z * z
	}
	bias := nf / ((nf - 1) * (nf - 2))
	return bias * third
}
