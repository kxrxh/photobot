package charts

import (
	"math"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/numutil"
)

func BinsForField(objects []analysis.Object, field string, min, max float64) map[string]int {
	values := make([]float64, 0, len(objects))
	for _, obj := range objects {
		if v := numericFromObject(obj, field); v != nil {
			values = append(values, *v)
		}
	}
	return binsFromValues(values, min, max)
}

func BinsForMass1000(objects []analysis.Object) map[string]int {
	values := make([]float64, 0, len(objects))
	for _, obj := range objects {
		if m := numericFromObject(obj, "mass_1000"); m != nil {
			values = append(values, *m)
		}
	}
	return binsFromMassValues(values)
}

type distributionValues struct {
	length []float64
	width  []float64
	lw     []float64
	sq     []float64
	mass   []float64
}

func collectDistributionValues(objects []analysis.Object) distributionValues {
	var dv distributionValues
	for _, obj := range objects {
		if v := numericFromObject(obj, "l"); v != nil {
			dv.length = append(dv.length, *v)
		}
		if v := numericFromObject(obj, "w"); v != nil {
			dv.width = append(dv.width, *v)
		}
		if v := numericFromObject(obj, "l_w"); v != nil {
			dv.lw = append(dv.lw, *v)
		}
		if v := numericFromObject(obj, "sq"); v != nil {
			dv.sq = append(dv.sq, *v)
		}
		if v := numericFromObject(obj, "mass_1000"); v != nil {
			dv.mass = append(dv.mass, *v)
		}
	}
	return dv
}

func binsFromValues(values []float64, min, max float64) map[string]int {
	if len(values) == 0 || !validPositiveRange(min, max) {
		return map[string]int{}
	}
	binW := adaptiveBinSize(min, max, 15)
	var binOrder []string
	bins := map[string]int{}
	for binStart := min; binStart < max; binStart += binW {
		binEnd := math.Min(binStart+binW, max)
		key := binLabel(binStart, binEnd)
		binOrder = append(binOrder, key)
		bins[key] = 0
	}
	if len(binOrder) == 0 {
		return map[string]int{}
	}
	lastKey := binOrder[len(binOrder)-1]
	for _, rv := range values {
		rv = numutil.Round3(rv)
		binIndex := int(math.Floor((rv - min) / binW))
		binStart := min + float64(binIndex)*binW
		binEnd := math.Min(binStart+binW, max)
		key := binLabel(binStart, binEnd)
		if _, ok := bins[key]; ok {
			bins[key]++
		} else {
			bins[lastKey]++
		}
	}
	out := map[string]int{}
	for k, c := range bins {
		if c > 0 {
			out[k] = c
		}
	}
	return out
}

func binsFromMassValues(values []float64) map[string]int {
	out := map[string]int{}
	for _, mass := range values {
		if mass < 28 {
			out["<28"]++
			continue
		}
		var key string
		if mass < 30 {
			key = binLabel(28, 30)
		} else {
			lo := 30 + math.Floor((mass-30)/5)*5
			hi := 30 + (math.Floor((mass-30)/5)+1)*5
			key = binLabel(lo, hi)
		}
		out[key]++
	}
	return out
}
