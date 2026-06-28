package charts

import (
	"html"
	"math"
	"sort"
	"strconv"
	"strings"

	"csort.ru/reports-service/internal/numutil"
)

const (
	defaultWidth  = 960.0
	defaultHeight = 380.0
	pdfFontFamily = "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif"
)

func niceYAxisMaxPercent(maxBar float64) float64 {
	if maxBar >= 100 {
		return 100
	}
	if maxBar <= 0 {
		return 10
	}
	headroom := math.Max(5, math.Ceil(maxBar*0.08))
	candidate := math.Min(100, maxBar+headroom)
	candidate = math.Ceil(candidate/5) * 5
	return math.Min(100, math.Max(10, candidate))
}

func BarChartSVG(
	title string,
	bins map[string]int,
	total int,
	stddev, skew *float64,
	stdUnit string,
) string {
	if total <= 0 || len(bins) == 0 {
		return emptyBarChartSVG(title)
	}
	keys := make([]string, 0, len(bins))
	for k := range bins {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	counts := make([]int, len(keys))
	pcts := make([]int, len(keys))
	maxPct := 0.0
	for i, k := range keys {
		c := bins[k]
		counts[i] = c
		p := math.Round((float64(c) / float64(total)) * 100)
		pi := int(p)
		pcts[i] = pi
		if float64(pi) > maxPct {
			maxPct = float64(pi)
		}
	}
	yMax := niceYAxisMaxPercent(maxPct)
	headerH := 44.0
	if stddev != nil && skew != nil {
		headerH = 56
	} else if stddev != nil || skew != nil {
		headerH = 44
	}
	plotTop := headerH + 6
	bottomPad := 44.0
	plotH := defaultHeight - bottomPad - plotTop
	x0, x1 := 44.0, defaultWidth-22
	grad := `<defs><linearGradient id="g1" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stop-color="#475569"/><stop offset="100%" stop-color="#334155"/></linearGradient></defs>`

	b := &strings.Builder{}
	b.WriteString(`<svg width="`)
	b.WriteString(strconv.FormatInt(int64(defaultWidth), 10))
	b.WriteString(`" height="`)
	b.WriteString(strconv.FormatInt(int64(defaultHeight), 10))
	b.WriteString(`" xmlns="http://www.w3.org/2000/svg">`)
	b.WriteString(grad)
	b.WriteString(
		`<rect width="100%" height="100%" fill="#ffffff" rx="10" stroke="#e2e8f0" stroke-width="1"/>`,
	)
	b.WriteString(`<rect x="0" y="0" width="100%" height="`)
	b.WriteString(numutil.FormatFloat(headerH, 2))
	b.WriteString(`" fill="#f1f5f9"/><line x1="0" y1="`)
	b.WriteString(numutil.FormatFloat(headerH, 2))
	b.WriteString(`" x2="`)
	b.WriteString(numutil.FormatFloat(defaultWidth, 2))
	b.WriteString(`" y2="`)
	b.WriteString(numutil.FormatFloat(headerH, 2))
	b.WriteString(`" stroke="#e2e8f0"/>`)
	b.WriteString(
		`<text x="16" y="30" font-size="17" font-weight="700" fill="#0f172a" font-family="`,
	)
	b.WriteString(html.EscapeString(pdfFontFamily))
	b.WriteString(`">`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</text>`)
	rx := defaultWidth - 16
	switch {
	case stddev != nil && skew != nil:
		b.WriteString(`<text x="`)
		b.WriteString(numutil.FormatFloat(rx, 2))
		b.WriteString(`" y="24" text-anchor="end" font-size="12" fill="#64748b" font-family="`)
		b.WriteString(html.EscapeString(pdfFontFamily))
		b.WriteString(`">СКО `)
		b.WriteString(numutil.FormatFloat(*stddev, 2))
		if stdUnit != "" {
			b.WriteString(" " + html.EscapeString(stdUnit))
		}
		b.WriteString(`</text>`)
		b.WriteString(`<text x="`)
		b.WriteString(numutil.FormatFloat(rx, 2))
		b.WriteString(`" y="39" text-anchor="end" font-size="12" fill="#64748b" font-family="`)
		b.WriteString(html.EscapeString(pdfFontFamily))
		b.WriteString(`">Асимметрия `)
		b.WriteString(numutil.FormatFloat(*skew, 2))
		b.WriteString(`</text>`)
	case stddev != nil:
		b.WriteString(`<text x="`)
		b.WriteString(numutil.FormatFloat(rx, 2))
		b.WriteString(`" y="30" text-anchor="end" font-size="12" fill="#64748b" font-family="`)
		b.WriteString(html.EscapeString(pdfFontFamily))
		b.WriteString(`">СКО `)
		b.WriteString(numutil.FormatFloat(*stddev, 2))
		if stdUnit != "" {
			b.WriteString(" " + html.EscapeString(stdUnit))
		}
		b.WriteString(`</text>`)
	case skew != nil:
		b.WriteString(`<text x="`)
		b.WriteString(numutil.FormatFloat(rx, 2))
		b.WriteString(`" y="30" text-anchor="end" font-size="12" fill="#64748b" font-family="`)
		b.WriteString(html.EscapeString(pdfFontFamily))
		b.WriteString(`">Асимметрия `)
		b.WriteString(numutil.FormatFloat(*skew, 2))
		b.WriteString(`</text>`)
	}

	ticksN := 5
	for i := 0; i <= ticksN; i++ {
		pct := yMax * float64(i) / float64(ticksN)
		yLine := plotTop + plotH - (pct/yMax)*plotH
		b.WriteString(`<line x1="`)
		b.WriteString(numutil.FormatFloat(x0, 2))
		b.WriteString(`" y1="`)
		b.WriteString(numutil.FormatFloat(yLine, 2))
		b.WriteString(`" x2="`)
		b.WriteString(numutil.FormatFloat(x1, 2))
		b.WriteString(`" y2="`)
		b.WriteString(numutil.FormatFloat(yLine, 2))
		b.WriteString(`" stroke="#e2e8f0" stroke-width="1"/>`)
		b.WriteString(`<text x="`)
		b.WriteString(numutil.FormatFloat(x0-8, 2))
		b.WriteString(`" y="`)
		b.WriteString(numutil.FormatFloat(yLine+4, 2))
		b.WriteString(`" text-anchor="end" font-size="11" fill="#64748b" font-family="`)
		b.WriteString(html.EscapeString(pdfFontFamily))
		b.WriteString(`">`)
		b.WriteString(strconv.Itoa(int(math.Round(pct))))
		b.WriteString(`%</text>`)
	}
	b.WriteString(`<line x1="`)
	b.WriteString(numutil.FormatFloat(x0, 2))
	b.WriteString(`" y1="`)
	b.WriteString(numutil.FormatFloat(plotTop+plotH, 2))
	b.WriteString(`" x2="`)
	b.WriteString(numutil.FormatFloat(x1, 2))
	b.WriteString(`" y2="`)
	b.WriteString(numutil.FormatFloat(plotTop+plotH, 2))
	b.WriteString(`" stroke="#94a3b8" stroke-width="2"/>`)

	chartAreaW := x1 - x0
	barW := math.Min(360, math.Max(36, 700/float64(len(keys))))
	if len(keys) == 0 {
		barW = 36
	}
	barGap := 12.0
	totalW := float64(len(keys))*barW + float64(len(keys)-1)*barGap
	startX := x0 + math.Max(0, (chartAreaW-totalW)/2)
	chartAreaBottom := plotTop + plotH
	catY := defaultHeight - 11
	for i, k := range keys {
		pct := pcts[i]
		c := counts[i]
		h := (float64(pct) / yMax) * plotH
		x := startX + float64(i)*(barW+barGap)
		y := chartAreaBottom - h
		fill := "url(#g1)"
		if k == "<28" {
			fill = "#e01818"
		}
		b.WriteString(`<rect x="`)
		b.WriteString(numutil.FormatFloat(x, 2))
		b.WriteString(`" y="`)
		b.WriteString(numutil.FormatFloat(y, 2))
		b.WriteString(`" width="`)
		b.WriteString(numutil.FormatFloat(barW, 2))
		b.WriteString(`" height="`)
		b.WriteString(numutil.FormatFloat(h, 2))
		b.WriteString(`" fill="`)
		b.WriteString(fill)
		b.WriteString(`" stroke="#1e293b" stroke-width="1" rx="5"/>`)
		lb := html.EscapeString(strconv.Itoa(pct) + "% (" + strconv.Itoa(c) + ")")
		b.WriteString(`<text x="`)
		b.WriteString(numutil.FormatFloat(x+barW/2, 2))
		b.WriteString(`" y="`)
		b.WriteString(numutil.FormatFloat(y-6, 2))
		b.WriteString(
			`" text-anchor="middle" font-size="12" font-weight="600" fill="#0f172a" font-family="`,
		)
		b.WriteString(html.EscapeString(pdfFontFamily))
		b.WriteString(`">`)
		b.WriteString(lb)
		b.WriteString(`</text>`)
		b.WriteString(`<text x="`)
		b.WriteString(numutil.FormatFloat(x+barW/2, 2))
		b.WriteString(`" y="`)
		b.WriteString(numutil.FormatFloat(catY, 2))
		b.WriteString(`" text-anchor="middle" font-size="12" fill="#64748b" font-family="`)
		b.WriteString(html.EscapeString(pdfFontFamily))
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(k))
		b.WriteString(`</text>`)
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func emptyBarChartSVG(title string) string {
	return `<svg width="` + numutil.FormatFloat(
		defaultWidth,
		2,
	) + `" height="` + numutil.FormatFloat(
		defaultHeight,
		2,
	) + `" xmlns="http://www.w3.org/2000/svg"><rect width="100%" height="100%" fill="#fff" rx="10" stroke="#e2e8f0"/>` +
		`<text x="` + numutil.FormatFloat(
		defaultWidth/2,
		2,
	) + `" y="` + numutil.FormatFloat(
		defaultHeight/2,
		2,
	) + `" text-anchor="middle" fill="#64748b" font-size="20">` + html.EscapeString(
		title,
	) + `</text></svg>`
}
