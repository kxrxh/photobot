package calc

import (
	"sort"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/statutil"
)

type FieldStatistics struct {
	Min    float64
	Max    float64
	Avg    float64
	Med    float64
	Stddev float64
	Skew   float64
}

type FieldStatisticsResult map[string]*FieldStatistics

var fieldStatNames = []struct {
	Key  string
	Name string
}{
	{"l", "l"},
	{"w", "w"},
	{"t", "t"},
	{"sq", "sq"},
	{"l_w", "l_w"},
	{"r", "r"},
	{"g", "g"},
	{"b", "b"},
	{"h", "h"},
	{"s", "s"},
	{"v", "v"},
}

func FieldStatisticsFromValues(values []float64) *FieldStatistics {
	if len(values) == 0 {
		return nil
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(len(values))
	min := sorted[0]
	max := sorted[len(sorted)-1]
	return &FieldStatistics{
		Min:    min,
		Max:    max,
		Avg:    avg,
		Med:    statutil.MedianSorted(sorted),
		Stddev: statutil.SampleStdDevMean(values, avg),
		Skew:   statutil.SkewnessSample(values),
	}
}

func FieldStatisticsForObjects(objects []analysis.Object, field string) *FieldStatistics {
	var values []float64
	for _, obj := range objects {
		v := NumericFieldValue(obj, field)
		if v != nil {
			values = append(values, *v)
		}
	}
	return FieldStatisticsFromValues(values)
}

func CalculateFieldStatisticsResult(objects []analysis.Object) FieldStatisticsResult {
	res := FieldStatisticsResult{}
	if len(objects) == 0 {
		return res
	}
	valuesBy := map[string][]float64{}
	for _, pair := range fieldStatNames {
		valuesBy[pair.Name] = nil
	}
	for _, obj := range objects {
		for _, pair := range fieldStatNames {
			v := NumericFieldValue(obj, pair.Key)
			if v != nil {
				valuesBy[pair.Name] = append(valuesBy[pair.Name], *v)
			}
		}
	}
	for _, pair := range fieldStatNames {
		if vs := valuesBy[pair.Name]; len(vs) > 0 {
			if s := FieldStatisticsFromValues(vs); s != nil {
				res[pair.Name] = s
			}
		}
	}
	return res
}
