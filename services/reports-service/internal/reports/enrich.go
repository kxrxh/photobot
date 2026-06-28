package reports

import (
	"strconv"
	"strings"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/calc"
	"csort.ru/reports-service/internal/charts"
	reportcontext "csort.ru/reports-service/internal/context"
	"csort.ru/reports-service/internal/numutil"
	"csort.ru/reports-service/internal/statutil"
)

type EnrichOptions struct {
	RepLimitWhenSingle int
	RepLimitWhenMulti  int
}

var DefaultEnrichOptions = &EnrichOptions{RepLimitWhenSingle: 21, RepLimitWhenMulti: 4}

func EnrichReportContext(
	rc *reportcontext.ReportContext,
	objects []analysis.Object,
	opt *EnrichOptions,
) (
	cs calc.ClassStatisticsResult,
	reps []calc.RepresentativeGroup,
	dist map[string]string,
	n int,
) {
	n = len(objects)
	if n == 0 {
		rc.ObjectCount = "0"
		return nil, nil, map[string]string{}, 0
	}
	if opt == nil {
		opt = DefaultEnrichOptions
	}
	singleLimit := opt.RepLimitWhenSingle
	multiLimit := opt.RepLimitWhenMulti
	if singleLimit <= 0 {
		singleLimit = 21
	}
	if multiLimit <= 0 {
		multiLimit = 4
	}

	st := calc.CalculateFieldStatisticsResult(objects)
	reportcontext.ApplyFieldStatisticsFallbackToContext(rc, st)
	reportcontext.SyncStatLenWidthForCharts(rc)
	if st["l"] != nil && reportcontext.IsMissingReportString(rc.LenStd) {
		rc.LenStd = reportcontext.FormatFixed(st["l"].Stddev, 2)
	}
	if st["w"] != nil && reportcontext.IsMissingReportString(rc.WidthStd) {
		rc.WidthStd = reportcontext.FormatFixed(st["w"].Stddev, 2)
	}
	if tstd := ThicknessSampleStd(
		objects,
	); tstd != nil &&
		reportcontext.IsMissingReportString(rc.ThickStd) {
		rc.ThickStd = reportcontext.FormatFixed(*tstd, 2)
	}

	cs = calc.CalculateClassStatisticsResult(objects)
	classN := classCountDistinct(objects)
	lim := multiLimit
	if classN == 1 {
		lim = singleLimit
	}
	reps = calc.CalculateTypicalRepresentativesByClass(
		objects,
		calc.RepresentativeOptions{PerClassLimit: lim},
	)
	dist = charts.BuildDistributionSVGMap(rc.Stats.ChartRanges(), objects, fieldAggsForCharts(st))

	if ShouldCalculateMassFromObjects(rc) {
		AppendMassCalculatedFields(rc, objects)
	} else {
		rc.Weight1000GrainsCalc = "-"
		rc.SampleMassCalc = "-"
		rc.MassLiterCalc = "-"
	}

	rc.ObjectCount = strconv.Itoa(n)
	return cs, reps, dist, n
}

func classCountDistinct(objects []analysis.Object) int {
	seen := map[string]struct{}{}
	for _, o := range objects {
		c := "unclassified"
		if o.Class != nil {
			if s := strings.TrimSpace(*o.Class); s != "" {
				c = s
			}
		}
		if c == "objects_img" {
			continue
		}
		seen[c] = struct{}{}
	}
	return len(seen)
}

func ThicknessSampleStd(objects []analysis.Object) *float64 {
	var vals []float64
	for _, o := range objects {
		if v := calc.NumericFieldValue(o, "t"); v != nil && numutil.NonZero(*v) {
			vals = append(vals, *v)
		}
	}
	if len(vals) < 2 {
		return nil
	}
	s := statutil.SampleStdDev(vals)
	return &s
}

func fieldAggsForCharts(st calc.FieldStatisticsResult) map[string]*charts.FieldAgg {
	out := make(map[string]*charts.FieldAgg, len(st))
	for k, v := range st {
		if v == nil {
			continue
		}
		a := charts.FieldAgg{
			Min: v.Min, Max: v.Max, Stddev: v.Stddev, Skew: v.Skew,
		}
		agg := a
		out[k] = &agg
	}
	return out
}
