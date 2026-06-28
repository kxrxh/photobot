package reports

import (
	"strconv"
	"strings"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/calc"
	"csort.ru/reports-service/internal/strutil"

	reportcontext "csort.ru/reports-service/internal/context"
)

func attachAnalysisImageURLs(rc *reportcontext.ReportContext, ar *analysis.AnalysisResult) {
	if rc == nil || ar == nil {
		return
	}
	urls := strutil.TrimNonEmpty(ar.FilesOutputURLs)
	if len(urls) == 0 {
		return
	}
	rc.Img2 = urls
	rc.Img2DownloadIndices = make([]int, len(urls))
	for i := range urls {
		rc.Img2DownloadIndices[i] = i
	}
}

func attachRepresentativeImageURLs(reps []calc.RepresentativeGroup, objects []analysis.Object) {
	imageByObjectID := map[string]string{}
	for _, obj := range objects {
		id := strconv.FormatInt(int64(obj.ID), 10)
		u := strutil.TrimPtr(obj.ImageURL)
		if id == "" || u == "" {
			continue
		}
		imageByObjectID[id] = u
	}
	if len(imageByObjectID) == 0 {
		return
	}
	for gi := range reps {
		attachRepresentativeCardImageURLs(reps[gi].Representatives, imageByObjectID)
	}
}

func attachRepresentativeCardImageURLs(
	cards []calc.RepresentativeCard,
	imageByObjectID map[string]string,
) {
	for i := range cards {
		id := strings.TrimSpace(cards[i].ObjectID)
		if id == "" {
			continue
		}
		if u := imageByObjectID[id]; u != "" {
			cards[i].ImageDataURL = u
		}
	}
}
