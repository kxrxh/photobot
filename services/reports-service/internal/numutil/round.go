package numutil

import "math"

func RoundPlaces(x float64, places int) float64 {
	if places < 0 {
		places = 0
	}
	p := math.Pow10(places)
	return math.Round(x*p) / p
}

func Round3(x float64) float64 { return RoundPlaces(x, 3) }
