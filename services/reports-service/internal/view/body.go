package view

import (
	"html/template"
	"strconv"
	"strings"

	"csort.ru/reports-service/internal/calc"
	reportcontext "csort.ru/reports-service/internal/context"
)

type Body struct {
	Header  Header
	Summary Summary
	Mass    *Mass

	SingleClass    *SingleClass
	ShowSingleReps bool
	SingleReps     []RepGroup

	MultiClass *MultiClass

	ChartsP1 []Chart
	ChartsP2 []Chart

	Images ImageStrip
}

type Header struct {
	LogoSrc string

	ShowDownloads    bool
	ObjectArchiveURL string
	CsvURL           string
	ReportPackQuery  string

	AnalysisNo string
	Date       string
	TimePart   string
	UserID     int64
	BotLink    string
	Company    string
}

type Summary struct {
	Product        string
	ProductMissing bool
	Sort           string
	SortMissing    bool

	Reproduction        string
	ReproductionMissing bool
	Germination         string
	GerminationMissing  bool

	Objects string

	Len          string
	LenMissing   bool
	Width        string
	WidthMissing bool
	ShowThick    bool
	Thick        string

	ShowExtra   bool
	ShowSeeds50 bool
	Seeds50     string
	ShowCount50 bool
	Count50     string
	ShowBroken  bool
	BrokenLabel string
	BrokenPers  string

	RHSMissing       bool
	ColorRHS         string
	RGBR, RGBG, RGBB int
	RGBOK            bool
}

type MassPair struct {
	Fact string
	Calc string
}

type Mass struct {
	Sample     *MassPair
	Grains1000 *MassPair
	Liter      *MassPair
}

type SingleClass struct {
	PrimaryCount int
	Details      []ClassDetail
}

type ClassDetail struct {
	ShowTitleRow bool
	Title        string
	Count        int
	Tables       []MetricTable
}

type ClassPill struct {
	Label string
	Count int
}

type MultiClass struct {
	TotalObjects int
	Pills        []ClassPill
	HasReps      bool
	Reps         []RepGroup
	Details      []ClassDetail
}

type RepGroup struct {
	Title        string
	TotalObjects int
	Cards        []RepCard
}

type RepCard struct {
	HasImage     bool
	ImageDataURL template.URL
	ObjectID     string
	Value        string
}

type MetricTable struct {
	Heading string
	Rows    []MetricRow
}

type MetricRow struct {
	Label string
	Min   string
	Max   string
	Avg   string
	Med   string
}

type Chart struct {
	HasSVG bool
	SVG    template.HTML
	Alt    string
	Miss   string
}

type ImageStrip struct {
	HasImages bool
	Cards     []ImageCard
}

type ImageCard struct {
	URL          template.URL
	Alt          string
	DisplayIndex int
	DownloadIdx  int

	HasImageBase bool
	ImageBaseURL string
	Query        string
}

func BuildBody(p PageParams) Body {
	rc := &p.Context
	v := Body{
		Header:  buildHeader(p, rc),
		Summary: buildSummary(rc),
		Mass:    buildMass(rc),
	}

	names := sortedClassNames(p.ClassStats)
	classCount := len(names)
	hasReps := len(p.Reps) > 0
	objN := p.Objects
	if n, err := strconv.Atoi(
		strings.TrimSpace(rc.ObjectCount),
	); err == nil &&
		rc.ObjectCount != "" {
		objN = n
	}
	isSingle := classCount == 1

	if isSingle && len(names) == 1 {
		cn := names[0]
		singleN := 0
		if p.ClassStats[cn] != nil {
			singleN = p.ClassStats[cn].Count
		}
		if !calc.IsDefectClassName(cn) {
			var details []ClassDetail
			for _, className := range names {
				if className == "objects_img" {
					continue
				}
				cd := p.ClassStats[className]
				if calc.IsDefectClassName(className) || !calc.HasAnyClassMetrics(cd) {
					continue
				}
				tables := classMetricTables(className, cd)
				if len(tables) == 0 {
					continue
				}
				details = append(details, ClassDetail{Tables: tables})
			}
			if len(details) > 0 {
				v.SingleClass = &SingleClass{PrimaryCount: singleN, Details: details}
			}
		}
	}

	v.ShowSingleReps = isSingle && hasReps
	if v.ShowSingleReps {
		for _, g := range p.Reps {
			if gv, ok := repGroup("Основной класс", g); ok {
				v.SingleReps = append(v.SingleReps, gv)
			}
		}
	}

	if objN > 0 {
		multi := classCount > 1
		if multi && (classCount > 0 || hasReps) {
			mc := &MultiClass{TotalObjects: objN, HasReps: hasReps}
			for _, className := range names {
				if className == "objects_img" {
					continue
				}
				row := p.ClassStats[className]
				if row == nil {
					continue
				}
				mc.Pills = append(mc.Pills, ClassPill{
					Label: TranslateClassName(className),
					Count: row.Count,
				})
			}
			for _, g := range p.Reps {
				if gv, ok := repGroup(TranslateClassName(g.ClassName), g); ok {
					mc.Reps = append(mc.Reps, gv)
				}
			}
			for _, className := range names {
				if className == "objects_img" {
					continue
				}
				cd := p.ClassStats[className]
				if calc.IsDefectClassName(className) || !calc.HasAnyClassMetrics(cd) {
					continue
				}
				tables := classMetricTables(className, cd)
				if len(tables) == 0 {
					continue
				}
				mc.Details = append(mc.Details, ClassDetail{
					ShowTitleRow: true,
					Title:        TranslateClassName(className),
					Count:        cd.Count,
					Tables:       tables,
				})
			}
			v.MultiClass = mc
		}
	}

	d := p.Dist
	v.ChartsP1 = []Chart{
		chartItem(d["distChartLength"], "График распределения длины", "длины"),
		chartItem(d["distChartWidth"], "График распределения ширины", "ширины"),
		chartItem(d["distChartLW"], "График распределения соотношения L/W", "L/W"),
	}
	v.ChartsP2 = []Chart{
		chartItem(d["distChartSQ"], "График распределения площади", "площади"),
		chartItem(
			d["distChartMass1000"],
			"График распределения массы 1000 зёрен",
			"массы 1000 зёрен",
		),
	}
	v.Images = buildImageStrip(p)
	return v
}

func buildHeader(p PageParams, rc *reportcontext.ReportContext) Header {
	h := Header{
		LogoSrc:          p.LogoSrc,
		ShowDownloads:    p.ObjectArchiveURL != "" || p.CsvURL != "",
		ObjectArchiveURL: p.ObjectArchiveURL,
		CsvURL:           p.CsvURL,
		ReportPackQuery:  p.ReportPackQuery,
		AnalysisNo:       rc.AnalysisID,
		Date:             stringOrDash(rc.Date),
		TimePart:         timePartFromDateTime(rc.DateTime),
		UserID:           rc.UserID,
		BotLink:          rc.BotLink,
	}
	comp := rc.Company
	if comp == "" {
		comp = "-"
	}
	h.Company = comp
	return h
}

func buildSummary(rc *reportcontext.ReportContext) Summary {
	if rc == nil {
		return Summary{}
	}
	product := rc.Product
	tprod := TranslateProductName(product)
	plow := strings.ToLower(product)
	s := Summary{
		Objects: rc.ObjectCount,
	}
	if tprod != "" && tprod != "-" {
		s.Product = tprod
	} else {
		s.ProductMissing = true
	}
	if v := rc.Sort; v != "" && v != "-" {
		s.Sort = v
	} else {
		s.SortMissing = true
	}
	if v := rc.Reproduction; v != "" && v != "-" {
		s.Reproduction = v
	} else {
		s.ReproductionMissing = true
	}
	if g := rc.Germination; g != "" && g != "-" {
		s.Germination = g
	} else {
		s.GerminationMissing = true
	}
	s.ShowThick = rc.ThickAvg != "" && rc.ThickAvg != "0" && rc.ThickAvg != "-"
	if s.ShowThick {
		s.Thick = rc.ThickAvg
	}
	if v := rc.LenAvg; v != "" {
		s.Len = v
	} else {
		s.LenMissing = true
	}
	if v := rc.WidthAvg; v != "" {
		s.Width = v
	} else {
		s.WidthMissing = true
	}
	s.BrokenLabel = "% негодных*"
	if strings.Contains(plow, "kernal") || strings.Contains(plow, "kernel") {
		s.BrokenLabel = "Ломанные %*"
	}
	s.ShowSeeds50 = plow == "seeds" && rc.EntitiesFor50Gramm != "" &&
		rc.EntitiesFor50Gramm != "-" &&
		rc.EntitiesFor50Gramm != "0"
	if s.ShowSeeds50 {
		s.Seeds50 = rc.EntitiesFor50Gramm
	}
	s.ShowCount50 = plow == "seeds_striped" && rc.Count50 != "" && rc.Count50 != "-" &&
		rc.Count50 != "0"
	if s.ShowCount50 {
		s.Count50 = rc.Count50
	}
	s.ShowBroken = rc.BrokenPers != ""
	if s.ShowBroken {
		s.BrokenPers = rc.BrokenPers
	}
	s.ShowExtra = s.ShowSeeds50 || s.ShowCount50 || s.ShowBroken

	colorRHS := rc.ColorRHS
	if colorRHS != "" && colorRHS != "-" {
		s.ColorRHS = colorRHS
		if r, g, b, ok := rgbFromStats(rc); ok {
			s.RGBOK = true
			s.RGBR, s.RGBG, s.RGBB = r, g, b
		}
	} else {
		s.RHSMissing = true
	}
	return s
}

func buildMass(rc *reportcontext.ReportContext) *Mass {
	if rc == nil {
		return nil
	}
	h1000 := hasMassPosStr(rc.Weight1000Grains)
	h1000c := hasMassCalcStr(rc.Weight1000GrainsCalc)
	hs := hasMassPosStr(rc.SampleMass)
	hsc := hasMassCalcStr(rc.SampleMassCalc)
	hml := hasMassPosStr(rc.MassLiter)
	hmlc := hasMassCalcStr(rc.MassLiterCalc)
	if !h1000 && !h1000c && !hs && !hsc && !hml && !hmlc {
		return nil
	}
	m := &Mass{}
	if hs || hsc {
		p := MassPair{Fact: "—", Calc: "—"}
		if hs {
			p.Fact = rc.SampleMass + " г"
		}
		if hsc {
			p.Calc = rc.SampleMassCalc
		}
		m.Sample = &p
	}
	if h1000 || h1000c {
		p := MassPair{Fact: "—", Calc: "—"}
		if h1000 {
			p.Fact = rc.Weight1000Grains + " г"
		}
		if h1000c {
			p.Calc = rc.Weight1000GrainsCalc
		}
		m.Grains1000 = &p
	}
	if hml || hmlc {
		p := MassPair{Fact: "—", Calc: "—"}
		if hml {
			p.Fact = rc.MassLiter + " г"
		}
		if hmlc {
			p.Calc = rc.MassLiterCalc
		}
		m.Liter = &p
	}
	return m
}

func chartItem(svg, alt, miss string) Chart {
	if svg != "" {
		return Chart{HasSVG: true, SVG: template.HTML(svg), Alt: alt, Miss: miss}
	}
	return Chart{HasSVG: false, Alt: alt, Miss: miss}
}

func buildImageStrip(p PageParams) ImageStrip {
	if len(p.Img2) == 0 {
		return ImageStrip{HasImages: false}
	}
	v := ImageStrip{HasImages: true}
	for i, url := range p.Img2 {
		dl := i
		if len(p.Img2DownloadIdx) > i {
			dl = p.Img2DownloadIdx[i]
		}
		v.Cards = append(v.Cards, ImageCard{
			URL:          template.URL(url),
			Alt:          "Обработанное изображение " + strconv.Itoa(i+1),
			DisplayIndex: i + 1,
			DownloadIdx:  dl,
			HasImageBase: p.ImageBaseURL != "",
			ImageBaseURL: p.ImageBaseURL,
			Query:        p.ReportPackQuery,
		})
	}
	return v
}
