package view

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"csort.ru/reports-service/internal/calc"
	reportcontext "csort.ru/reports-service/internal/context"
	"csort.ru/reports-service/internal/maputil"
)

type PageParams struct {
	Context          reportcontext.ReportContext
	ClassStats       calc.ClassStatisticsResult
	Reps             []calc.RepresentativeGroup
	Dist             map[string]string
	Objects          int
	LogoSrc          string
	CsvURL           string
	ObjectArchiveURL string
	ImageBaseURL     string
	ReportPackQuery  string
	Img2             []string
	Img2DownloadIdx  []int
}

func sortedClassNames(cs calc.ClassStatisticsResult) []string {
	var n []string
	for k := range cs {
		if k != "objects_img" {
			n = append(n, k)
		}
	}
	sort.Strings(n)
	return n
}

func timePartFromDateTime(dt string) string {
	if dt == "" {
		return ""
	}
	var t string
	if i := strings.IndexByte(dt, 'T'); i >= 0 {
		t = dt[i+1:]
	} else if i := strings.IndexByte(dt, ' '); i >= 0 {
		t = dt[i+1:]
	} else {
		return ""
	}
	if len(t) >= 5 {
		return t[:5]
	}
	return ""
}

func rgbFromStats(rc *reportcontext.ReportContext) (int, int, int, bool) {
	if rc == nil {
		return 0, 0, 0, false
	}
	ra, ok1 := maputil.ParseFloatString(rc.Stats.R.Avg)
	ga, ok2 := maputil.ParseFloatString(rc.Stats.G.Avg)
	ba, ok3 := maputil.ParseFloatString(rc.Stats.B.Avg)
	if !ok1 || !ok2 || !ok3 {
		return 0, 0, 0, false
	}
	r := int(math.Round(ra))
	g := int(math.Round(ga))
	bl := int(math.Round(ba))
	if r < 0 {
		r = 0
	}
	if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	}
	if g > 255 {
		g = 255
	}
	if bl < 0 {
		bl = 0
	}
	if bl > 255 {
		bl = 255
	}
	return r, g, bl, true
}

func hasMassPosStr(s string) bool {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	if s == "" || s == "-" {
		return false
	}
	f, err := strconv.ParseFloat(s, 64)
	return err == nil && f > 0
}

func hasMassCalcStr(s string) bool {
	return s != "" && s != "-"
}

func stringOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
