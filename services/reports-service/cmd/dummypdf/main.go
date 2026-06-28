package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/calc"
	reportcontext "csort.ru/reports-service/internal/context"
	"csort.ru/reports-service/internal/fileutil"
	"csort.ru/reports-service/internal/render"
	"csort.ru/reports-service/internal/reports"
	"csort.ru/reports-service/internal/view"
)

//go:embed testdata/local-dummy-analysis.json
var embeddedLocalDummyAnalysisJSON []byte

const localDummyDefaultID = "local-dummy-preview"

var localPreviewEnrichOptions = &reports.EnrichOptions{RepLimitWhenSingle: 16, RepLimitWhenMulti: 4}

func main() {
	os.Exit(runMain())
}

func runMain() int {
	outPDF := flag.String("out", "local-preview/dummy-report.pdf", "output .pdf")
	outHTML := flag.String("html", "local-preview/dummy-preview.html", "output .html")
	templatesDir := flag.String(
		"templates",
		"./templates",
		"templates root (html/main.html, assets/css, assets/qr, …)",
	)
	assetsDir := flag.String("assets", "./preview-assets", "optional images for preview")
	host := flag.String(
		"host",
		"http://localhost:9999",
		"analysis service base URL for header links",
	)
	analysisID := flag.String("analysis-id", localDummyDefaultID, "analysis id shown in report")
	analysisJSON := flag.String(
		"analysis-json",
		"",
		`path to Analysis API JSON {"success":true,"result":{...}}; empty = embedded fixture`,
	)
	objectDisplay := flag.String(
		"object-display",
		envOrDefault("DUMMY_OBJECT_DISPLAY", ""),
		"override object count label (e.g. 999)",
	)
	htmlOnly := flag.Bool("html-only", false, "skip PDF")
	flag.Parse()

	tpl := render.MainHTMLPath(*templatesDir)
	if err := render.LoadReportTemplate(tpl); err != nil {
		fmt.Fprintf(os.Stderr, "template: %v\n", err)
		return 1
	}

	payload, err := loadLocalDummyAnalysis(*analysisJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "analysis-json: %v\n", err)
		return 1
	}

	page, _, err := buildLocalPreviewPage(
		*host,
		*assetsDir,
		*analysisID,
		strings.TrimSpace(*objectDisplay),
		payload,
		reports.NewReportPacker("", 0),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "preview: %v\n", err)
		return 1
	}

	bodyHTML, err := render.RenderBody(view.BuildBody(page))
	if err != nil {
		fmt.Fprintf(os.Stderr, "render body: %v\n", err)
		return 1
	}

	html, err := render.RenderReportHTML(render.ReportHTMLData{BodyHTML: bodyHTML})
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		return 1
	}
	html = render.InjectQRSVG(html, *templatesDir)

	_ = os.MkdirAll(filepath.Dir(*outHTML), 0o750)
	if err := os.WriteFile(*outHTML, []byte(html), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "html: %v\n", err)
		return 1
	}
	fmt.Println(*outHTML)

	if *htmlOnly {
		return 0
	}

	conv := reports.NewPDFConverter(1)
	defer conv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pdf, err := conv.ConvertHTMLToPDF(ctx, *templatesDir, html)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pdf: %v\n", err)
		return 1
	}

	_ = os.MkdirAll(filepath.Dir(*outPDF), 0o750)
	if err := os.WriteFile(*outPDF, pdf, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		return 1
	}
	fmt.Println(*outPDF, len(pdf), "bytes")
	return 0
}

func loadLocalDummyAnalysis(path string) (*analysis.AnalysisResult, error) {
	var raw []byte
	if strings.TrimSpace(path) == "" {
		raw = embeddedLocalDummyAnalysisJSON
	} else {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	var env analysis.AnalysisAPIResponse
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, errors.New("analysis-json: success is false")
	}
	return &env.Result, nil
}

func assignPreviewImagesToRepresentatives(groups []calc.RepresentativeGroup, images []string) {
	if len(images) == 0 {
		return
	}
	idx := 0
	for g := range groups {
		for r := range groups[g].Representatives {
			url := images[idx%len(images)]
			c := groups[g].Representatives[r]
			c.ImageDataURL = url
			groups[g].Representatives[r] = c
			idx++
		}
	}
}

func buildLocalPreviewPage(
	serviceHost, assetsDir, analysisID, objectCountDisplay string,
	ar *analysis.AnalysisResult,
	reportPack reports.ReportPacker,
) (view.PageParams, reportcontext.ReportContext, error) {
	if ar == nil {
		return view.PageParams{}, reportcontext.ReportContext{}, errors.New(
			"dummypdf: analysis result is nil",
		)
	}
	objects := ar.Objects
	if analysisID == "" {
		analysisID = localDummyDefaultID
	}
	ar.ID = analysisID
	rc := reports.ReportContextFromAnalysis(ar, analysisID)
	cs, reps, dist, n := reports.EnrichReportContext(
		&rc,
		objects,
		localPreviewEnrichOptions,
	)
	img2 := fileURLsInDir(assetsDir)
	rc.Img2 = img2
	repImages := fileURLsInDirOrFallback(filepath.Join(assetsDir, "objects"), img2)
	assignPreviewImagesToRepresentatives(reps, repImages)
	if objectCountDisplay != "" {
		rc.ObjectCount = strings.TrimSpace(objectCountDisplay)
	}
	host := strings.TrimRight(strings.TrimSpace(serviceHost), "/")
	return view.PageParams{
		Context:          rc,
		ClassStats:       cs,
		Reps:             reps,
		Dist:             dist,
		Objects:          n,
		LogoSrc:          render.LogoRelPath,
		CsvURL:           host + "/analyses/" + analysisID + "/report/csv",
		ObjectArchiveURL: host + "/analyses/" + analysisID + "/images/objects/archive",
		ImageBaseURL:     host + "/analyses/" + analysisID + "/images",
		ReportPackQuery:  reportPack.Query(analysisID),
		Img2:             rc.Img2,
		Img2DownloadIdx:  rc.Img2DownloadIndices,
	}, rc, nil
}

func isImageName(n string) bool {
	lower := strings.ToLower(n)
	return strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".webp")
}

func listSortedImageNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if isImageName(n) {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}

func fileURLsInDir(dir string) []string {
	return imagePathsToFileURLs(listSortedImagePaths(dir))
}

func fileURLsInDirOrFallback(dir string, fallback []string) []string {
	u := fileURLsInDir(dir)
	if len(u) == 0 {
		return append([]string(nil), fallback...)
	}
	return u
}

func listSortedImagePaths(dir string) []string {
	names := listSortedImageNames(dir)
	if len(names) == 0 {
		return nil
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		out = append(out, filepath.Join(dir, n))
	}
	return out
}

func imagePathsToFileURLs(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	var urls []string
	for _, p := range paths {
		u, err := fileutil.BrowserFileURL(p)
		if err != nil {
			continue
		}
		urls = append(urls, u)
	}
	if len(urls) == 0 {
		return nil
	}
	return urls
}

func envOrDefault(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
