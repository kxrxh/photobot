package calc

import "csort.ru/reports-service/internal/api/analysis"

type MassDerived struct {
	AvgMass1000  *float64
	SumMassGrams *float64
}

func CalculateMassDerivedFromObjects(objects []analysis.Object) MassDerived {
	if len(objects) == 0 {
		return MassDerived{}
	}
	var sum float64
	var n int
	for _, o := range objects {
		if v := NumericFieldValue(o, "mass_1000"); v != nil {
			sum += *v
			n++
		}
	}
	if n == 0 {
		return MassDerived{}
	}
	avg := sum / float64(n)
	sumGrams := sum / 1000
	return MassDerived{AvgMass1000: &avg, SumMassGrams: &sumGrams}
}
