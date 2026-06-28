package reports

import (
	"context"
	"strings"
	"time"

	"csort.ru/reports-service/internal/calc"
	"csort.ru/reports-service/internal/httputil"
	"csort.ru/reports-service/internal/logger"
	"csort.ru/reports-service/internal/strutil"

	reportcontext "csort.ru/reports-service/internal/context"
)

const (
	imageFetchTimeout     = 30 * time.Second
	imageFetchMaxBytes    = 15 << 20 // 15 MiB per image
	imageFetchMaxParallel = 8
)

func prefetchReportImages(
	ctx context.Context,
	rc *reportcontext.ReportContext,
	reps []calc.RepresentativeGroup,
) {
	urls := collectReportImageURLs(rc, reps)
	if len(urls) == 0 {
		return
	}

	log := logger.WithTrace(ctx, logger.Logger)
	client := httputil.PublicClient(imageFetchTimeout, false)
	embedded := httputil.FetchMapAsDataURLs(
		ctx,
		client,
		urls,
		imageFetchMaxBytes,
		imageFetchMaxParallel,
		func(u string, err error) {
			log.Warn().
				Err(err).
				Str("image_url", u).
				Msg("report image prefetch failed; keeping remote URL")
		},
	)
	applyEmbeddedImageURLs(rc, reps, embedded)

	if n := len(embedded); n > 0 {
		log.Info().Int("embedded_images", n).Msg("report images embedded for PDF")
	}
}

func collectReportImageURLs(
	rc *reportcontext.ReportContext,
	reps []calc.RepresentativeGroup,
) []string {
	seen := map[string]struct{}{}
	var urls []string
	add := func(u string) {
		u = strings.TrimSpace(u)
		if !strutil.IsHTTPURL(u) {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		urls = append(urls, u)
	}
	if rc != nil {
		for _, u := range rc.Img2 {
			add(u)
		}
	}
	for _, g := range reps {
		for _, c := range g.Representatives {
			add(c.ImageDataURL)
		}
	}
	return urls
}

func applyEmbeddedImageURLs(
	rc *reportcontext.ReportContext,
	reps []calc.RepresentativeGroup,
	embedded map[string]string,
) {
	if rc != nil {
		for i, u := range rc.Img2 {
			if data, ok := embedded[u]; ok {
				rc.Img2[i] = data
			}
		}
	}
	for gi := range reps {
		for ci := range reps[gi].Representatives {
			u := reps[gi].Representatives[ci].ImageDataURL
			if data, ok := embedded[u]; ok {
				reps[gi].Representatives[ci].ImageDataURL = data
			}
		}
	}
}
