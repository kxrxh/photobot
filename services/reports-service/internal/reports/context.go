package reports

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"csort.ru/reports-service/internal/api/analysis"
	reportcontext "csort.ru/reports-service/internal/context"
	"csort.ru/reports-service/internal/numutil"
	"csort.ru/reports-service/internal/strutil"
)

const reportBotLink = "https://max.ru/id2222810514_bot"

func ReportContextFromAnalysis(
	ar *analysis.AnalysisResult,
	analyzeID string,
) reportcontext.ReportContext {
	if ar == nil {
		return emptyReportContext(analyzeID)
	}

	product := strutil.TrimPtr(ar.Product)
	company := strutil.TrimPtr(ar.Company)
	sort := strutil.TrimPtr(ar.Sort)
	reproduction := strutil.TrimPtr(ar.Reproduction)

	germination := "-"
	if ar.Germination != nil {
		germination = fmt.Sprintf("%.0f%%", *ar.Germination)
	}

	dateTime := ar.DateTime
	date := "-"
	if len(dateTime) >= 10 {
		if t, err := time.Parse("2006-01-02", dateTime[:10]); err == nil {
			date = fmt.Sprintf("%02d.%02d.%d", t.Day(), int(t.Month()), t.Year())
		}
	}

	ap := ar.AnalysisParams
	var lStats, wStats, tStats *analysis.ChannelStats
	if ap != nil {
		lStats = ap.L
		wStats = ap.W
		tStats = ap.T
	}

	tAbsent := aggregateThicknessNotCalculated(tStats)
	var thickAvgStr, thickStdStr, stThickMin, stThickMax, stThickAvg, stThickMed string
	if tAbsent {
		dash := "-"
		thickAvgStr, thickStdStr = dash, dash
		stThickMin, stThickMax, stThickAvg, stThickMed = dash, dash, dash, dash
	} else {
		thickAvgStr, thickStdStr = statAvg(tStats), statStd(tStats)
		stThickMin, stThickMax, stThickAvg, stThickMed = statMin(tStats), statMax(
			tStats,
		), statAvg(
			tStats,
		), statMed(
			tStats,
		)
	}

	rc := reportcontext.ReportContext{
		Product:      strutil.EmptyAsDash(product),
		LenAvg:       statAvg(lStats),
		LenStd:       statStd(lStats),
		WidthAvg:     statAvg(wStats),
		WidthStd:     statStd(wStats),
		ThickAvg:     thickAvgStr,
		ThickStd:     thickStdStr,
		StatThickMin: stThickMin,
		StatThickMax: stThickMax,
		StatThickAvg: stThickAvg,
		StatThickMed: stThickMed,
		Weight1000Grains: floatFieldFromParams(ap, func(p *analysis.AnalysisParams) *float64 {
			return p.Mass1000
		}),
		SampleMass: floatFieldFromParams(ap, func(p *analysis.AnalysisParams) *float64 {
			return p.Mass
		}),
		MassLiter:          massLiterStr(ap),
		EntitiesFor50Gramm: "-",
		Count50:            "-",
		ColorRHS:           colorRhsFromParams(ap),
		UserID:             ar.UserID,
		Date:               date,
		DateTime:           strutil.EmptyAsDash(dateTime),
		AnalysisID:         analyzeID,
		BotLink:            reportBotLink,
		Company:            strutil.EmptyAsDash(company),
		Sort:               strutil.EmptyAsDash(sort),
		Reproduction:       strutil.EmptyAsDash(reproduction),
		Germination:        germination,
		Stats: reportcontext.ReportObjectStats{
			ChartLenMin:   statMin(lStats),
			ChartLenMax:   statMax(lStats),
			ChartLenAvg:   statAvg(lStats),
			ChartLenMed:   statMed(lStats),
			ChartWidthMin: statMin(wStats),
			ChartWidthMax: statMax(wStats),
			ChartWidthAvg: statAvg(wStats),
			ChartWidthMed: statMed(wStats),
		},
	}

	if ap != nil && ap.Mass != nil && *ap.Mass > 0 {
		rc.EntitiesFor50Gramm = fmt.Sprintf("%.2f", 50/(*ap.Mass))
	}

	productLower := strings.ToLower(product)
	if ap != nil && ap.Count50 != nil && *ap.Count50 > 0 && productLower == "seeds_striped" {
		rc.Count50 = trimIntish(*ap.Count50)
	}

	var broken *float64
	if ap != nil {
		if ap.BrokenPercent != nil {
			broken = ap.BrokenPercent
		} else if ap.MassPercent != nil {
			broken = ap.MassPercent
		}
	}
	if broken != nil {
		rc.BrokenPers = numutil.FormatFloat(*broken, 2)
	}

	return rc
}

func colorRhsFromParams(ap *analysis.AnalysisParams) string {
	if ap == nil {
		return "-"
	}
	return strutil.PtrOrDash(ap.ColorRhs)
}

func floatFieldFromParams(
	ap *analysis.AnalysisParams,
	get func(*analysis.AnalysisParams) *float64,
) string {
	if ap == nil {
		return "-"
	}
	v := get(ap)
	if v == nil {
		return "-"
	}
	t := *v
	if t != t {
		return "-"
	}
	return numutil.FormatFloat(t, 2)
}

func emptyReportContext(analyzeID string) reportcontext.ReportContext {
	return reportcontext.ReportContext{
		Product:            "-",
		UserID:             reportcontext.UserIDUnavailable,
		Company:            "-",
		Sort:               "-",
		Reproduction:       "-",
		Germination:        "-",
		Date:               "-",
		DateTime:           "-",
		AnalysisID:         analyzeID,
		LenAvg:             "-",
		LenStd:             "-",
		WidthAvg:           "-",
		WidthStd:           "-",
		ThickAvg:           "-",
		ThickStd:           "-",
		Weight1000Grains:   "-",
		SampleMass:         "-",
		MassLiter:          "-",
		Count50:            "-",
		ColorRHS:           "-",
		BotLink:            reportBotLink,
		EntitiesFor50Gramm: "-",
	}
}

func massLiterStr(ap *analysis.AnalysisParams) string {
	if ap == nil || ap.MassLiter == nil {
		return "-"
	}
	f := *ap.MassLiter
	if f != f {
		return "-"
	}
	return numutil.FormatFloat(f, 2)
}

func statMin(m *analysis.ChannelStats) string {
	if m == nil || m.Min == nil {
		return "-"
	}
	return numutil.FormatFloat(*m.Min, 2)
}

func statMax(m *analysis.ChannelStats) string {
	if m == nil || m.Max == nil {
		return "-"
	}
	return numutil.FormatFloat(*m.Max, 2)
}

func statAvg(m *analysis.ChannelStats) string {
	if m == nil || m.Avg == nil {
		return "-"
	}
	return numutil.FormatFloat(*m.Avg, 2)
}

func statMed(m *analysis.ChannelStats) string {
	md := analysis.CoalesceMedian(m)
	if md == nil {
		return "-"
	}
	return numutil.FormatFloat(*md, 2)
}

func statStd(m *analysis.ChannelStats) string {
	if m == nil || m.Stddev == nil {
		return "-"
	}
	return numutil.FormatFloat(*m.Stddev, 2)
}

func trimIntish(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	return numutil.FormatFloat(f, 2)
}

func aggregateThicknessNotCalculated(m *analysis.ChannelStats) bool {
	if m == nil {
		return true
	}
	avg, min, max := f64OrZero(m.Avg), f64OrZero(m.Min), f64OrZero(m.Max)
	med := f64OrZero(analysis.CoalesceMedian(m))
	return avg == 0 && min == 0 && max == 0 && med == 0
}

func f64OrZero(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
