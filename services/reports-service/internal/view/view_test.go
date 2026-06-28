package view

import (
	"bytes"
	"html/template"
	"testing"

	"csort.ru/reports-service/internal/calc"
	reportcontext "csort.ru/reports-service/internal/context"
	"csort.ru/reports-service/internal/numutil"
)

func TestTranslateClassName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"whole", "Целые"},
		{"  LARGE ", "Крупные"},
		{"Unknown", "Unknown"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := TranslateClassName(tt.in); got != tt.want {
			t.Errorf("TranslateClassName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTranslateProductName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"wheat", "Пшеница"},
		{"  SOYBEANS ", "Соя"},
		{"seeds_striped", "Подсолнечник полосатый"},
		{"not-a-product", "not-a-product"},
	}
	for _, tt := range tests {
		if got := TranslateProductName(tt.in); got != tt.want {
			t.Errorf("TranslateProductName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSortedClassNames_skipsObjectsImgAndSorts(t *testing.T) {
	t.Parallel()
	cs := calc.ClassStatisticsResult{
		"zebra":       {Count: 1},
		"objects_img": {Count: 99},
		"alpha":       {Count: 2},
	}
	got := sortedClassNames(cs)
	want := []string{"alpha", "zebra"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestTimePartFromDateTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"2026-01-15", ""},
		{"2026-01-15T14:30:00Z", "14:30"},
		{"2026-01-15 09:05:00", "09:05"},
		{"2026-01-15T12:03", "12:03"},
		{"2026-01-15T12", ""},
	}
	for _, tt := range tests {
		if got := timePartFromDateTime(tt.in); got != tt.want {
			t.Errorf("timePartFromDateTime(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRgbFromStats(t *testing.T) {
	t.Parallel()
	if _, _, _, ok := rgbFromStats(nil); ok {
		t.Fatal("nil context should be !ok")
	}
	rc := &reportcontext.ReportContext{
		Stats: reportcontext.ReportObjectStats{
			R: reportcontext.ObjectStatSeries4{Avg: "not-a-number"},
			G: reportcontext.ObjectStatSeries4{Avg: "1"},
			B: reportcontext.ObjectStatSeries4{Avg: "2"},
		},
	}
	if _, _, _, ok := rgbFromStats(rc); ok {
		t.Fatal("invalid R should be !ok")
	}
	rc2 := &reportcontext.ReportContext{
		Stats: reportcontext.ReportObjectStats{
			R: reportcontext.ObjectStatSeries4{Avg: "300"},
			G: reportcontext.ObjectStatSeries4{Avg: "-5"},
			B: reportcontext.ObjectStatSeries4{Avg: "128,5"},
		},
	}
	r, g, b, ok := rgbFromStats(rc2)
	if !ok {
		t.Fatal("expected ok")
	}
	if r != 255 || g != 0 || b != 129 {
		t.Fatalf("clamped/clipped RGB = (%d,%d,%d), want (255,0,129)", r, g, b)
	}
}

func TestHasMassPosStrAndHasMassCalcStr(t *testing.T) {
	t.Parallel()
	if hasMassPosStr("") || hasMassPosStr("-") || hasMassPosStr("0") || hasMassPosStr("-1") {
		t.Fatal("empty/dash/zero/nonpositive should be false")
	}
	if !hasMassPosStr("1,5") || !hasMassPosStr(" 2.25 ") {
		t.Fatal("positive numbers should be true")
	}
	if hasMassCalcStr("") || hasMassCalcStr("-") {
		t.Fatal("empty and dash calc should be false")
	}
	if !hasMassCalcStr("0") || !hasMassCalcStr("computed") {
		t.Fatal("nonempty non-dash should be true")
	}
}

func TestStringOrDash(t *testing.T) {
	t.Parallel()
	if stringOrDash("") != "-" || stringOrDash("x") != "x" {
		t.Fatal("stringOrDash mismatch")
	}
}

func TestClassMetricTables(t *testing.T) {
	t.Parallel()
	if classMetricTables("large", nil) != nil {
		t.Fatal("nil class data")
	}
	if classMetricTables("broken", &calc.ClassStatistics{L: &calc.ClassFieldStats{}}) != nil {
		t.Fatal("defect class should skip tables")
	}
	cd := &calc.ClassStatistics{
		Count: 3,
		L:     &calc.ClassFieldStats{Min: 1, Max: 2, Avg: 1.5, Med: 1.4},
	}
	tables := classMetricTables("large", cd)
	if len(tables) != 1 || tables[0].Heading == "" || len(tables[0].Rows) == 0 {
		t.Fatalf("unexpected tables: %+v", tables)
	}
	cdOnlyColor := &calc.ClassStatistics{
		Count: 1,
		R:     &calc.ClassFieldStats{Min: 0, Max: 1, Avg: 0.5, Med: 0.5},
	}
	tab2 := classMetricTables("small", cdOnlyColor)
	if len(tab2) != 1 || tab2[0].Heading != "Цвет (R, G, B, H, S, V)" {
		t.Fatalf("want color-only table, got %+v", tab2)
	}
}

func TestAppendMetricRow_nilField(t *testing.T) {
	t.Parallel()
	rows := []MetricRow{{Label: "keep"}}
	out := appendMetricRow(rows, nil, "ignored")
	if len(out) != 1 || out[0].Label != "keep" {
		t.Fatalf("expected unchanged slice, got %+v", out)
	}
}

func TestRepGroup(t *testing.T) {
	t.Parallel()
	g := calc.RepresentativeGroup{ClassName: "large", Representatives: []calc.RepresentativeCard{}}
	if _, ok := repGroup("x", g); ok {
		t.Fatal("empty representatives")
	}
	g2 := calc.RepresentativeGroup{
		TotalObjects: 10,
		Representatives: []calc.RepresentativeCard{
			{ObjectID: "42", Value: 1.2345, ImageDataURL: "data:image/png;base64,x"},
			{ObjectID: "43", Value: 2, ImageDataURL: "   "},
		},
	}
	rg, ok := repGroup("Title", g2)
	if !ok || rg.Title != "Title" || rg.TotalObjects != 10 || len(rg.Cards) != 2 {
		t.Fatalf("unexpected RepGroup %+v ok=%v", rg, ok)
	}
	if !rg.Cards[0].HasImage || rg.Cards[1].HasImage {
		t.Fatal("HasImage should follow trimmed ImageDataURL")
	}
	if rg.Cards[0].ObjectID != "42" || rg.Cards[0].Value != numutil.FormatFloat(1.2345, 3) {
		t.Fatalf("card0: ObjectID=%q Value=%q", rg.Cards[0].ObjectID, rg.Cards[0].Value)
	}
}

func TestChartItem(t *testing.T) {
	t.Parallel()
	a := chartItem("<svg/>", "alt", "miss")
	if !a.HasSVG || string(a.SVG) != "<svg/>" || a.Alt != "alt" {
		t.Fatalf("nonempty svg: %+v", a)
	}
	b := chartItem("", "alt", "miss")
	if b.HasSVG || b.Alt != "alt" {
		t.Fatalf("empty svg: %+v", b)
	}
}

func TestBuildHeader(t *testing.T) {
	t.Parallel()
	p := PageParams{
		LogoSrc:          "/logo.svg",
		CsvURL:           "http://csv",
		ObjectArchiveURL: "",
		ReportPackQuery:  "pack=1",
		Context: reportcontext.ReportContext{
			AnalysisID: "A-1",
			Date:       "",
			DateTime:   "2026-04-29T09:41:02Z",
			UserID:     9001,
			BotLink:    "tg://bot",
			Company:    "",
		},
	}
	h := buildHeader(p, &p.Context)
	if !h.ShowDownloads || h.LogoSrc != "/logo.svg" || h.Company != "-" {
		t.Fatalf("header fields: %+v", h)
	}
	if h.TimePart != timePartFromDateTime(p.Context.DateTime) || h.AnalysisNo != "A-1" {
		t.Fatalf("time/analysis: %+v", h)
	}
}

func TestBuildSummary_nilAndRGB(t *testing.T) {
	t.Parallel()
	if (buildSummary(nil) != Summary{}) {
		t.Fatal("nil summary should be zero")
	}
	rc := &reportcontext.ReportContext{
		Product:      "kernels",
		Sort:         "-",
		Reproduction: "-",
		Germination:  "-",
		ObjectCount:  "5",
		ColorRHS:     "RHS 1",
		Stats: reportcontext.ReportObjectStats{
			R: reportcontext.ObjectStatSeries4{Avg: "10"},
			G: reportcontext.ObjectStatSeries4{Avg: "20"},
			B: reportcontext.ObjectStatSeries4{Avg: "30"},
		},
	}
	s := buildSummary(rc)
	if s.BrokenLabel != "Ломанные %*" || !s.SortMissing || !s.RGBOK || s.RGBR != 10 {
		t.Fatalf("summary: %+v", s)
	}
}

func TestBuildMass(t *testing.T) {
	t.Parallel()
	if buildMass(nil) != nil {
		t.Fatal("nil mass")
	}
	empty := &reportcontext.ReportContext{}
	if buildMass(empty) != nil {
		t.Fatal("all empty mass fields")
	}
	rc := &reportcontext.ReportContext{
		SampleMass:           "100",
		SampleMassCalc:       "calc",
		Weight1000GrainsCalc: "-",
	}
	m := buildMass(rc)
	if m == nil || m.Sample == nil || m.Grains1000 != nil || m.Sample.Fact == "—" {
		t.Fatalf("mass: %+v", m)
	}
}

func TestBuildMass_grains1000AndLiter(t *testing.T) {
	t.Parallel()
	rc := &reportcontext.ReportContext{
		Weight1000Grains: "50",
		MassLiter:        "800",
		MassLiterCalc:    "calcL",
	}
	m := buildMass(rc)
	if m == nil || m.Grains1000 == nil || m.Liter == nil {
		t.Fatalf("expected Grains1000 and Liter: %+v", m)
	}
	if m.Grains1000.Fact != "50 г" || m.Liter.Calc != "calcL" {
		t.Fatalf("pairs: grains=%+v liter=%+v", m.Grains1000, m.Liter)
	}
}

func TestBuildSummary_thickSeedsCountBrokenAndRHSMissing(t *testing.T) {
	t.Parallel()
	rc := &reportcontext.ReportContext{
		Product:            "seeds",
		Sort:               "s",
		Reproduction:       "r",
		Germination:        "g",
		ObjectCount:        "1",
		LenAvg:             "1",
		WidthAvg:           "2",
		ThickAvg:           "3",
		EntitiesFor50Gramm: "10",
		BrokenPers:         "5",
		ColorRHS:           "-",
	}
	s := buildSummary(rc)
	if !s.ShowThick || s.Thick != "3" || !s.ShowSeeds50 || s.Seeds50 != "10" {
		t.Fatalf("thick/seeds: %+v", s)
	}
	if !s.ShowBroken || s.BrokenPers != "5" || !s.RHSMissing || s.RGBOK {
		t.Fatalf("broken/RHS: %+v", s)
	}
	rc2 := &reportcontext.ReportContext{
		Product:      "seeds_striped",
		Sort:         "s",
		Reproduction: "r",
		Germination:  "g",
		ObjectCount:  "1",
		LenAvg:       "1",
		WidthAvg:     "1",
		Count50:      "7",
		ColorRHS:     "RHS",
		Stats: reportcontext.ReportObjectStats{
			R: reportcontext.ObjectStatSeries4{Avg: "x"},
			G: reportcontext.ObjectStatSeries4{Avg: "1"},
			B: reportcontext.ObjectStatSeries4{Avg: "2"},
		},
	}
	s2 := buildSummary(rc2)
	if !s2.ShowCount50 || s2.Count50 != "7" || s2.RHSMissing {
		t.Fatalf("striped/RGB fail: %+v", s2)
	}
}

func TestBuildImageStrip(t *testing.T) {
	t.Parallel()
	if s := buildImageStrip(PageParams{}); s.HasImages || len(s.Cards) != 0 {
		t.Fatalf("empty images: %+v", s)
	}
	p := PageParams{
		Img2:            []string{"http://a", "http://b"},
		Img2DownloadIdx: []int{9, 8},
		ImageBaseURL:    "https://cdn/",
		ReportPackQuery: "x=1",
	}
	st := buildImageStrip(p)
	if !st.HasImages || len(st.Cards) != 2 || st.Cards[0].DownloadIdx != 9 ||
		st.Cards[1].DisplayIndex != 2 {
		t.Fatalf("strip: %+v", st.Cards)
	}
	if !st.Cards[0].HasImageBase || st.Cards[0].ImageBaseURL != "https://cdn/" ||
		st.Cards[0].Query != "x=1" {
		t.Fatalf("card0 download fields: %+v", st.Cards[0])
	}
}

func TestBuildBody_singleClassWithMetricsAndReps(t *testing.T) {
	t.Parallel()
	p := PageParams{
		Context: reportcontext.ReportContext{
			ObjectCount: "12",
			LenAvg:      "1",
			WidthAvg:    "2",
			ColorRHS:    "-",
		},
		ClassStats: calc.ClassStatisticsResult{
			"large": {
				Count: 12,
				L:     &calc.ClassFieldStats{Min: 1, Max: 2, Avg: 1.5, Med: 1},
			},
		},
		Reps: []calc.RepresentativeGroup{
			{
				ClassName:       "large",
				Representatives: []calc.RepresentativeCard{{ObjectID: "1", Value: 1}},
			},
		},
		Dist: map[string]string{},
	}
	b := BuildBody(p)
	if b.SingleClass == nil || b.ShowSingleReps == false || len(b.SingleReps) != 1 {
		t.Fatalf("single class body: SingleClass=%v ShowSingleReps=%v reps=%d",
			b.SingleClass != nil, b.ShowSingleReps, len(b.SingleReps))
	}
	if b.MultiClass != nil {
		t.Fatal("expected no multi-class")
	}
}

func TestBuildBody_multiClassChartAndImg(t *testing.T) {
	t.Parallel()
	p := PageParams{
		Context: reportcontext.ReportContext{
			ObjectCount: "50",
			LenAvg:      "1",
			WidthAvg:    "1",
			ColorRHS:    "-",
		},
		Objects: 50,
		ClassStats: calc.ClassStatisticsResult{
			"large": {Count: 20, L: &calc.ClassFieldStats{Min: 1, Max: 1, Avg: 1, Med: 1}},
			"small": {Count: 30, L: &calc.ClassFieldStats{Min: 1, Max: 1, Avg: 1, Med: 1}},
		},
		Reps: []calc.RepresentativeGroup{},
		Dist: map[string]string{
			"distChartLength": "<svg>n</svg>",
		},
		Img2: []string{"u1"},
	}
	b := BuildBody(p)
	if b.MultiClass == nil || b.MultiClass.TotalObjects != 50 || len(b.ChartsP1) != 3 ||
		!b.ChartsP1[0].HasSVG || !b.Images.HasImages {
		tot := 0
		if b.MultiClass != nil {
			tot = b.MultiClass.TotalObjects
		}
		t.Fatalf(
			"multi/charts/images: multi_nil=%v total=%d charts_len=%d firstHasSVG=%v has_images=%v",
			b.MultiClass == nil,
			tot,
			len(b.ChartsP1),
			b.ChartsP1[0].HasSVG,
			b.Images.HasImages,
		)
	}
}

func TestBuildBody_objectCountOverridesObjectsField(t *testing.T) {
	t.Parallel()
	p := PageParams{
		Context: reportcontext.ReportContext{
			ObjectCount: "99",
			LenAvg:      "1",
			WidthAvg:    "1",
			ColorRHS:    "-",
		},
		Objects: 1,
		ClassStats: calc.ClassStatisticsResult{
			"a": {Count: 1, L: &calc.ClassFieldStats{Min: 1, Max: 1, Avg: 1, Med: 1}},
			"b": {Count: 1, L: &calc.ClassFieldStats{Min: 1, Max: 1, Avg: 1, Med: 1}},
		},
	}
	b := BuildBody(p)
	if b.MultiClass == nil || b.MultiClass.TotalObjects != 99 {
		t.Fatalf("TotalObjects = %d", b.MultiClass.TotalObjects)
	}
}

func TestBuildBody_objNZeroSkipsMultiClass(t *testing.T) {
	t.Parallel()
	p := PageParams{
		Context: reportcontext.ReportContext{
			ObjectCount: "0",
			LenAvg:      "1",
			WidthAvg:    "1",
			ColorRHS:    "-",
		},
		Objects: 0,
		ClassStats: calc.ClassStatisticsResult{
			"a": {Count: 1, L: &calc.ClassFieldStats{Min: 1, Max: 1, Avg: 1, Med: 1}},
			"b": {Count: 1, L: &calc.ClassFieldStats{Min: 1, Max: 1, Avg: 1, Med: 1}},
		},
	}
	b := BuildBody(p)
	if b.MultiClass != nil {
		t.Fatalf("objN=0 should not build MultiClass, got %+v", b.MultiClass)
	}
}

func TestBuildBody_singleClassBrokenSkipsSingleClassDetails(t *testing.T) {
	t.Parallel()
	p := PageParams{
		Context: reportcontext.ReportContext{
			ObjectCount: "5",
			LenAvg:      "1",
			WidthAvg:    "1",
			ColorRHS:    "-",
		},
		ClassStats: calc.ClassStatisticsResult{
			"broken": {
				Count: 5,
				L:     &calc.ClassFieldStats{Min: 1, Max: 1, Avg: 1, Med: 1},
			},
		},
	}
	b := BuildBody(p)
	if b.SingleClass != nil {
		t.Fatal("single defect class should not populate SingleClass details")
	}
}

// Exercise template field types referenced from HTML (compile-time sanity for view types).
func TestBodyTypesUsedInTemplates(t *testing.T) {
	t.Parallel()
	b := Body{
		Header:  Header{},
		Summary: Summary{RGBOK: true, RGBR: 1, RGBG: 2, RGBB: 3},
		ChartsP1: []Chart{
			{HasSVG: true, SVG: template.HTML("<svg></svg>"), Alt: "a"},
		},
		Images: ImageStrip{
			HasImages: true,
			Cards:     []ImageCard{{URL: template.URL("http://x"), DisplayIndex: 1}},
		},
	}
	const tpl = `{{.Summary.RGBOK}}-{{.ChartsP1}}</body>`
	n := template.Must(template.New("t").Parse(tpl))
	var buf bytes.Buffer
	if err := n.Execute(&buf, b); err != nil {
		t.Fatal(err)
	}
}
