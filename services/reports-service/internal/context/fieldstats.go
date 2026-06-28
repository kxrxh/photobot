package reportcontext

import (
	"strings"

	"csort.ru/reports-service/internal/calc"
	"csort.ru/reports-service/internal/numutil"
)

func isMissingDisplay(s string) bool {
	t := strings.ToLower(strings.TrimSpace(s))
	return t == "" || t == "-" || t == "nan"
}

func setFormattedFloatIfMissing(target *string, value *float64, digits int) {
	if value == nil || !numutil.IsFinite(*value) {
		return
	}
	if *target != "" && !isMissingDisplay(*target) {
		return
	}
	*target = numutil.FormatFloat(*value, digits)
}

func setStatFloatIfMissing(target *string, val float64, digits int) {
	v := val
	setFormattedFloatIfMissing(target, &v, digits)
}

func FormatFixed(v float64, digits int) string {
	return numutil.FormatFloat(v, digits)
}

func IsMissingReportString(s string) bool {
	return isMissingDisplay(s)
}

type dimStatDesc struct {
	field   string
	hasSkew bool
}

var dimensionStats = []dimStatDesc{
	{"l", true},
	{"w", true},
	{"t", true},
	{"sq", true},
	{"l_w", false},
}

var colorStats = []string{"r", "g", "b", "h", "s", "v"}

func (st *ReportObjectStats) series5(field string) *ObjectStatSeries5 {
	switch field {
	case "l":
		return &st.L
	case "w":
		return &st.W
	case "t":
		return &st.T
	case "sq":
		return &st.Sq
	default:
		return nil
	}
}

func (st *ReportObjectStats) series4Lw() *ObjectStatSeries4 {
	return &st.Lw
}

func (st *ReportObjectStats) series4Color(k string) *ObjectStatSeries4 {
	switch k {
	case "r":
		return &st.R
	case "g":
		return &st.G
	case "b":
		return &st.B
	case "h":
		return &st.H
	case "s":
		return &st.S
	case "v":
		return &st.V
	default:
		return nil
	}
}

func applyOneDimension(rc *ReportContext, s *calc.FieldStatistics, d dimStatDesc) {
	if d.field == "l_w" {
		applyLwSeries(rc.Stats.series4Lw(), s)
		return
	}
	ser := rc.Stats.series5(d.field)
	if ser == nil {
		return
	}
	set := func(target *string, val float64) {
		setStatFloatIfMissing(target, val, 2)
	}
	dim := d.field
	set(&ser.Min, s.Min)
	set(&ser.Max, s.Max)
	set(&ser.Avg, s.Avg)
	set(&ser.Med, s.Med)
	if d.hasSkew {
		set(&ser.Asym, s.Skew)
	}
	if dim == "t" {
		set(&rc.StatThickMin, s.Min)
		set(&rc.StatThickMax, s.Max)
		set(&rc.StatThickAvg, s.Avg)
		set(&rc.StatThickMed, s.Med)
	}
}

func applyLwSeries(lw *ObjectStatSeries4, s *calc.FieldStatistics) {
	set := func(target *string, val float64) {
		setStatFloatIfMissing(target, val, 2)
	}
	set(&lw.Min, s.Min)
	set(&lw.Max, s.Max)
	set(&lw.Avg, s.Avg)
	set(&lw.Med, s.Med)
}

func applyColorDimension(rc *ReportContext, s *calc.FieldStatistics, k string) {
	col := rc.Stats.series4Color(k)
	if col == nil {
		return
	}
	set := func(target *string, val float64) {
		setStatFloatIfMissing(target, val, 2)
	}
	set(&col.Min, s.Min)
	set(&col.Max, s.Max)
	set(&col.Avg, s.Avg)
	set(&col.Med, s.Med)
}

func applyFieldStatisticsToContext(rc *ReportContext, st calc.FieldStatisticsResult) {
	for _, d := range dimensionStats {
		if s := st[d.field]; s != nil {
			applyOneDimension(rc, s, d)
		}
	}
	for _, k := range colorStats {
		if s := st[k]; s != nil {
			applyColorDimension(rc, s, k)
		}
	}
}

func ApplyFieldStatisticsFallbackToContext(rc *ReportContext, st calc.FieldStatisticsResult) {
	applyFieldStatisticsToContext(rc, st)
}

func SyncStatLenWidthForCharts(rc *ReportContext) {
	st := &rc.Stats
	pairs := []struct{ from, to *string }{
		{&st.L.Min, &st.ChartLenMin},
		{&st.L.Max, &st.ChartLenMax},
		{&st.W.Min, &st.ChartWidthMin},
		{&st.W.Max, &st.ChartWidthMax},
	}
	for _, p := range pairs {
		if *p.to != "" {
			continue
		}
		if *p.from != "" {
			*p.to = *p.from
		}
	}
}
