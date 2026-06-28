package charts

import (
	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/maputil"
)

type FieldAgg struct {
	Min, Max float64
	Stddev   float64
	Skew     float64
}

const (
	KeyLength   = "distChartLength"
	KeyWidth    = "distChartWidth"
	KeyLW       = "distChartLW"
	KeyArea     = "distChartSQ"
	KeyMass1000 = "distChartMass1000"
)

type ChartRanges struct {
	LenMin, LenMax, WidthMin, WidthMax string
}

func BuildDistributionSVGMap(
	ranges ChartRanges,
	objects []analysis.Object,
	aggs map[string]*FieldAgg,
) map[string]string {
	out := map[string]string{}
	if len(objects) == 0 {
		return out
	}
	total := len(objects)
	dv := collectDistributionValues(objects)

	if mn, mx, ok := rangeFromStrings(ranges.LenMin, ranges.LenMax); ok {
		bins := binsFromValues(dv.length, mn, mx)
		if len(bins) > 0 {
			var std, sk *float64
			if a := aggs["l"]; a != nil {
				std = &a.Stddev
				sk = &a.Skew
			}
			out[KeyLength] = BarChartSVG(
				"График распределения: Длина (мм)",
				bins,
				total,
				std,
				sk,
				"мм",
			)
		}
	}
	if mn, mx, ok := rangeFromStrings(ranges.WidthMin, ranges.WidthMax); ok {
		bins := binsFromValues(dv.width, mn, mx)
		if len(bins) > 0 {
			var std, sk *float64
			if a := aggs["w"]; a != nil {
				std = &a.Stddev
				sk = &a.Skew
			}
			out[KeyWidth] = BarChartSVG(
				"График распределения: Ширина (мм)",
				bins,
				total,
				std,
				sk,
				"мм",
			)
		}
	}
	if a := aggs["l_w"]; a != nil {
		bins := binsFromValues(dv.lw, a.Min, a.Max)
		if len(bins) > 0 {
			std, skw := a.Stddev, a.Skew
			out[KeyLW] = BarChartSVG(
				"График распределения: Соотношение L/W",
				bins,
				total,
				&std,
				&skw,
				"",
			)
		}
	}
	if a := aggs["sq"]; a != nil {
		bins := binsFromValues(dv.sq, a.Min, a.Max)
		if len(bins) > 0 {
			std, skw := a.Stddev, a.Skew
			out[KeyArea] = BarChartSVG(
				"График распределения: Площадь (мм²)",
				bins,
				total,
				&std,
				&skw,
				"мм²",
			)
		}
	}
	massBins := binsFromMassValues(dv.mass)
	if len(massBins) > 0 {
		out[KeyMass1000] = BarChartSVG(
			"График распределения: Масса 1000 зёрен (г)",
			massBins,
			total,
			nil,
			nil,
			"",
		)
	}
	return out
}

func rangeFromStrings(minS, maxS string) (float64, float64, bool) {
	a, ok1 := maputil.ParseFloatString(minS)
	b, ok2 := maputil.ParseFloatString(maxS)
	if !ok1 || !ok2 || a >= b {
		return 0, 0, false
	}
	if !validPositiveRange(a, b) {
		return 0, 0, false
	}
	return a, b, true
}
