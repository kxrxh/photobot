package reports

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"csort.ru/reports-service/internal/fileutil"
	"csort.ru/reports-service/internal/logger"
	"csort.ru/reports-service/internal/observability"
	"csort.ru/reports-service/internal/pdfutil"

	"github.com/chromedp/cdproto/page"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const pdfChromedpTimeout = 55 * time.Minute

type PDFConverter struct {
	allocCtx    context.Context
	cancelAlloc context.CancelFunc
	sem         chan struct{}
	initOnce    sync.Once
}

func NewPDFConverter(maxConcurrentPages int) *PDFConverter {
	if maxConcurrentPages <= 0 {
		maxConcurrentPages = 5
	}
	return &PDFConverter{
		sem: make(chan struct{}, maxConcurrentPages),
	}
}

func (c *PDFConverter) initBrowser() {
	c.initOnce.Do(func() {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			// Do not set disable-background-networking: embedded images may still use remote URLs as fallback.
			chromedp.Flag("disable-extensions", true),
		)
		if p := pdfutil.ChromeExecPath(); p != "" {
			opts = append(opts, chromedp.ExecPath(p))
		}
		c.allocCtx, c.cancelAlloc = chromedp.NewExecAllocator(context.Background(), opts...)
	})
}

func (c *PDFConverter) ConvertHTMLToPDF(
	ctx context.Context,
	templateDir string,
	html string,
) (pdfOut []byte, err error) {
	ctx, span := observability.StartSpan(ctx, "reports.pdf.chrome",
		trace.WithAttributes(attribute.Int("reports.html_input_bytes", len(html))))
	defer func() { observability.EndSpan(span, err) }()

	log := logger.WithTrace(ctx, logger.Logger)
	start := time.Now()
	htmlLen := len(html)

	c.initBrowser()

	tmp, err := os.CreateTemp(templateDir, ".report-render-*.html")
	if err != nil {
		return nil, fmt.Errorf("pdf temp html: %w", err)
	}
	tmpPath := tmp.Name()
	removed := false
	defer func() {
		if !removed {
			_ = fileutil.RemoveWithinDir(templateDir, tmpPath)
		}
	}()

	if _, err := tmp.WriteString(html); err != nil {
		return nil, fmt.Errorf("pdf temp html write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("pdf temp html close: %w", err)
	}

	navURL, err := fileutil.BrowserFileURL(tmpPath)
	if err != nil {
		return nil, err
	}

	c.sem <- struct{}{}
	defer func() { <-c.sem }()

	pageCtx, pageCancel := chromedp.NewContext(c.allocCtx)
	defer pageCancel()

	runCtx, runCancel := context.WithTimeout(pageCtx, pdfChromedpTimeout)
	defer runCancel()
	stopOnRequest := context.AfterFunc(ctx, runCancel)
	defer stopOnRequest()

	var pdf []byte
	err = chromedp.Run(
		runCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			t0 := time.Now()
			err := chromedp.Navigate(navURL).Do(ctx)
			log.Debug().
				Str("step", "navigate_file").
				Str("url", navURL).
				Dur("duration", time.Since(t0)).
				Err(err).
				Msg("pdf chrome step")
			return err
		}),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible(".report-shell", chromedp.ByQuery),
		chromedp.Evaluate(
			printReadyJS,
			nil,
			func(p *cdpruntime.EvaluateParams) *cdpruntime.EvaluateParams {
				return p.WithAwaitPromise(true)
			},
		),
		chromedp.ActionFunc(func(ctx context.Context) error {
			t0 := time.Now()
			res, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				WithTransferMode(page.PrintToPDFTransferModeReturnAsBase64).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithMarginTop(0).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				Do(ctx)
			log.Debug().
				Str("step", "print_to_pdf").
				Dur("duration", time.Since(t0)).
				Err(err).
				Msg("pdf chrome step")
			if err != nil {
				return err
			}
			pdf = res
			return nil
		}),
	)
	if rmErr := fileutil.RemoveWithinDir(templateDir, tmpPath); rmErr == nil {
		removed = true
	}
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "target closed") ||
			strings.Contains(strings.ToLower(err.Error()), "context canceled") {
			err = fmt.Errorf("pdf render failed (browser target closed): %w", err)
		}
		log.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Int("html_bytes", htmlLen).
			Msg("pdf generation failed")
		return nil, err
	}
	if err := pdfutil.ValidateHeader(pdf); err != nil {
		log.Error().
			Err(err).
			Int("pdf_bytes", len(pdf)).
			Int("html_bytes", htmlLen).
			Msg("pdf generation produced invalid output")
		return nil, fmt.Errorf("pdf validation: %w", err)
	}
	log.Info().
		Dur("duration", time.Since(start)).
		Int("html_bytes", htmlLen).
		Int("pdf_bytes", len(pdf)).
		Msg("pdf chrome: conversion completed")
	return pdf, nil
}

func (c *PDFConverter) WarmUp(ctx context.Context, templateDir, html string) error {
	_, err := c.ConvertHTMLToPDF(ctx, templateDir, html)
	return err
}

func (c *PDFConverter) Close() {
	if c.cancelAlloc != nil {
		c.cancelAlloc()
	}
}
