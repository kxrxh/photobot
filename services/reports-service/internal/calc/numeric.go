package calc

import (
	"math"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/numutil"
)

func floatFromPtr(p *float64) (float64, bool) {
	if p == nil {
		return 0, false
	}
	v := *p
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

func floatPtr(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

func NumericFieldValue(obj analysis.Object, field string) *float64 {
	switch field {
	case "l_w":
		if v, ok := floatFromPtr(obj.LW); ok {
			return floatPtr(v)
		}
		l, lOK := floatFromPtr(obj.L)
		w, wOK := floatFromPtr(obj.W)
		if lOK && wOK && w != 0 {
			return floatPtr(l / w)
		}
		return nil
	case "t":
		if v, ok := floatFromPtr(obj.T); ok && numutil.NonZero(v) {
			return floatPtr(v)
		}
		return nil
	case "l":
		if v, ok := floatFromPtr(obj.L); ok {
			return floatPtr(v)
		}
		return nil
	case "w":
		if v, ok := floatFromPtr(obj.W); ok {
			return floatPtr(v)
		}
		return nil
	case "sq":
		if v, ok := floatFromPtr(obj.Sq); ok {
			return floatPtr(v)
		}
		return nil
	case "mass_1000":
		if v, ok := floatFromPtr(obj.Mass1000); ok {
			return floatPtr(v)
		}
		return nil
	case "r":
		if v, ok := floatFromPtr(obj.R); ok {
			return floatPtr(v)
		}
		return nil
	case "g":
		if v, ok := floatFromPtr(obj.G); ok {
			return floatPtr(v)
		}
		return nil
	case "b":
		if v, ok := floatFromPtr(obj.B); ok {
			return floatPtr(v)
		}
		return nil
	case "h":
		if v, ok := floatFromPtr(obj.H); ok {
			return floatPtr(v)
		}
		return nil
	case "s":
		if v, ok := floatFromPtr(obj.S); ok {
			return floatPtr(v)
		}
		return nil
	case "v":
		if v, ok := floatFromPtr(obj.V); ok {
			return floatPtr(v)
		}
		return nil
	default:
		return nil
	}
}
