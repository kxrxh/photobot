package calc

import (
	"strings"

	"csort.ru/reports-service/internal/api/analysis"
)

type ClassFieldStats struct {
	Min float64
	Max float64
	Avg float64
	Med float64
}

type ClassStatistics struct {
	Count            int
	L, W, T, SQ, LW  *ClassFieldStats
	R, G, B, H, S, V *ClassFieldStats
}

type ClassStatisticsResult map[string]*ClassStatistics

func IsDefectClassName(className string) bool {
	n := strings.ToLower(strings.TrimSpace(className))
	return n == "broken" || n == "ломанные"
}

func HasAnyClassMetrics(c *ClassStatistics) bool {
	if c == nil {
		return false
	}
	return c.L != nil || c.W != nil || c.T != nil || c.SQ != nil || c.LW != nil ||
		c.R != nil || c.G != nil || c.B != nil || c.H != nil || c.S != nil || c.V != nil
}

func CalculateClassStatisticsResult(objects []analysis.Object) ClassStatisticsResult {
	if len(objects) == 0 {
		return nil
	}
	groups := map[string][]analysis.Object{}
	for _, obj := range objects {
		cn := "unclassified"
		if obj.Class != nil {
			if s := strings.TrimSpace(*obj.Class); s != "" {
				cn = s
			}
		}
		groups[cn] = append(groups[cn], obj)
	}
	out := ClassStatisticsResult{}
	for className, objs := range groups {
		cs := &ClassStatistics{Count: len(objs)}
		if !IsDefectClassName(className) {
			applyClassFieldStats(cs, classFieldStatisticsForObjects(objs))
		}
		out[className] = cs
	}
	return out
}

var classStatFields = []struct {
	key string
	set func(*ClassStatistics, *FieldStatistics)
}{
	{"l", func(c *ClassStatistics, s *FieldStatistics) { c.L = classFieldStatsFrom(s) }},
	{"w", func(c *ClassStatistics, s *FieldStatistics) { c.W = classFieldStatsFrom(s) }},
	{"t", func(c *ClassStatistics, s *FieldStatistics) { c.T = classFieldStatsFrom(s) }},
	{"sq", func(c *ClassStatistics, s *FieldStatistics) { c.SQ = classFieldStatsFrom(s) }},
	{"l_w", func(c *ClassStatistics, s *FieldStatistics) { c.LW = classFieldStatsFrom(s) }},
	{"r", func(c *ClassStatistics, s *FieldStatistics) { c.R = classFieldStatsFrom(s) }},
	{"g", func(c *ClassStatistics, s *FieldStatistics) { c.G = classFieldStatsFrom(s) }},
	{"b", func(c *ClassStatistics, s *FieldStatistics) { c.B = classFieldStatsFrom(s) }},
	{"h", func(c *ClassStatistics, s *FieldStatistics) { c.H = classFieldStatsFrom(s) }},
	{"s", func(c *ClassStatistics, s *FieldStatistics) { c.S = classFieldStatsFrom(s) }},
	{"v", func(c *ClassStatistics, s *FieldStatistics) { c.V = classFieldStatsFrom(s) }},
}

func classFieldStatsFrom(s *FieldStatistics) *ClassFieldStats {
	if s == nil {
		return nil
	}
	return &ClassFieldStats{Min: s.Min, Max: s.Max, Avg: s.Avg, Med: s.Med}
}

func classFieldStatisticsForObjects(objects []analysis.Object) map[string]*FieldStatistics {
	valuesBy := make(map[string][]float64, len(classStatFields))
	for _, pair := range classStatFields {
		valuesBy[pair.key] = nil
	}
	for _, obj := range objects {
		for _, pair := range classStatFields {
			if v := NumericFieldValue(obj, pair.key); v != nil {
				valuesBy[pair.key] = append(valuesBy[pair.key], *v)
			}
		}
	}
	out := make(map[string]*FieldStatistics, len(classStatFields))
	for _, pair := range classStatFields {
		if s := FieldStatisticsFromValues(valuesBy[pair.key]); s != nil {
			out[pair.key] = s
		}
	}
	return out
}

func applyClassFieldStats(cs *ClassStatistics, stats map[string]*FieldStatistics) {
	for _, pair := range classStatFields {
		if s := stats[pair.key]; s != nil {
			pair.set(cs, s)
		}
	}
}
