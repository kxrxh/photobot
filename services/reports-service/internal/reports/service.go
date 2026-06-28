package reports

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/authz"
	"csort.ru/reports-service/internal/calc"
	"csort.ru/reports-service/internal/config"
	reportcontext "csort.ru/reports-service/internal/context"
	"csort.ru/reports-service/internal/domain"
	"csort.ru/reports-service/internal/logger"
	"csort.ru/reports-service/internal/observability"
	"csort.ru/reports-service/internal/render"
	"csort.ru/reports-service/internal/storage"
	"csort.ru/reports-service/internal/view"

	"go.opentelemetry.io/otel/attribute"
)

type Service struct {
	cfg          config.Config
	analysis     *analysis.Client
	minio        *storage.MinIOService
	pdf          *PDFConverter
	reportPack   ReportPacker
	csvFormatVer string
	pdfFormatVer string
}

func NewService(
	cfg config.Config,
	analysisClient *analysis.Client,
	minio *storage.MinIOService,
	pdf *PDFConverter,
) *Service {
	return &Service{
		cfg:          cfg,
		analysis:     analysisClient,
		minio:        minio,
		pdf:          pdf,
		csvFormatVer: cfg.ReportFormat.CSV,
		pdfFormatVer: cfg.ReportFormat.PDF,
		reportPack: NewReportPacker(
			cfg.ReportPack.HMACSecret,
			cfg.ReportPack.TTLSeconds,
		),
	}
}

func (s *Service) EnsureCurrent(
	ctx context.Context,
	analysisID, token string,
	fileTypes []string,
) (domain.ReportResult, error) {
	if len(fileTypes) == 0 {
		fileTypes = []string{"csv", "pdf"}
	}
	log := logger.WithTrace(ctx, logger.Logger)
	var needsGenerate bool
	regenErr := observability.RunPhase(
		ctx,
		"reports.ensure.needs_regeneration",
		func(cctx context.Context) error {
			var e error
			needsGenerate, e = s.needsRegeneration(cctx, analysisID, fileTypes)
			return e
		},
		attribute.String("reports.analysis_id", analysisID),
		attribute.String("reports.file_types", strings.Join(fileTypes, ",")),
	)
	if regenErr != nil {
		observability.RecordError(
			ctx,
			regenErr,
			attribute.String("reports.phase", "needs_regeneration"),
		)
		log.Error().
			Err(regenErr).
			Str("analysis_id", analysisID).
			Msg("report ensure: storage check failed")
		return domain.ReportResult{}, regenErr
	}
	if !needsGenerate {
		log.Info().
			Str("analysis_id", analysisID).
			Strs("file_types", fileTypes).
			Str("outcome", "cache_hit").
			Msg("report ensure")
		return domain.ReportResult{
			Success:    true,
			AnalysisID: analysisID,
			Files: domain.ReportFiles{
				CSV: analysisID + ".csv",
				PDF: analysisID + ".pdf",
			},
		}, nil
	}
	log.Info().
		Str("analysis_id", analysisID).
		Strs("file_types", fileTypes).
		Str("outcome", "regenerate").
		Msg("report ensure")
	return s.Generate(ctx, analysisID, token)
}

func (s *Service) Generate(
	ctx context.Context,
	analysisID, token string,
) (domain.ReportResult, error) {
	log := logger.WithTrace(ctx, logger.Logger)
	genStart := time.Now()

	log.Info().Str("analysis_id", analysisID).Msg("report generate started")

	var (
		analysisResult *analysis.AnalysisResult
		status         int
		analysisErr    error
		objects        []analysis.Object
		objectsStatus  int
		objectsErr     error
	)
	var upstreamWG sync.WaitGroup
	upstreamWG.Add(2)
	go func() {
		defer upstreamWG.Done()
		analysisErr = observability.RunPhase(
			ctx,
			"reports.get_analysis",
			func(cctx context.Context) error {
				var e error
				analysisResult, status, e = s.analysis.GetAnalysis(cctx, analysisID, token)
				return e
			},
			attribute.String("reports.analysis_id", analysisID),
		)
	}()
	go func() {
		defer upstreamWG.Done()
		objectsErr = observability.RunPhase(
			ctx,
			"reports.get_objects",
			func(cctx context.Context) error {
				var e error
				objects, objectsStatus, e = s.analysis.GetObjects(cctx, analysisID, token)
				return e
			},
			attribute.String("reports.analysis_id", analysisID),
		)
	}()
	upstreamWG.Wait()

	if analysisErr != nil {
		observability.RecordError(ctx, analysisErr,
			attribute.String("reports.phase", "get_analysis"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.Int("upstream.http.status", status),
		)
		log.Error().
			Err(analysisErr).
			Str("analysis_id", analysisID).
			Str("phase", "get_analysis").
			Int("upstream_status", status).
			Dur("elapsed", time.Since(genStart)).
			Msg("report generate failed")
		return domain.ReportResult{
			Success:    false,
			AnalysisID: analysisID,
			Files:      domain.ReportFiles{},
			Error:      analysisErr.Error(),
			StatusCode: analysis.MapStatus(status),
		}, nil
	}

	log.Info().
		Str("analysis_id", analysisID).
		Str("phase", "get_analysis").
		Int64("user_id", reportResultUserID(analysisResult)).
		Int("upstream_status", status).
		Dur("elapsed", time.Since(genStart)).
		Msg("report generate: analysis fetched")

	if objectsErr != nil {
		observability.RecordError(ctx, objectsErr,
			attribute.String("reports.phase", "get_objects"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.Int("upstream.http.status", objectsStatus),
		)
		log.Error().
			Err(objectsErr).
			Str("analysis_id", analysisID).
			Str("phase", "get_objects").
			Int("upstream_status", objectsStatus).
			Dur("elapsed", time.Since(genStart)).
			Msg("report generate failed")
		return domain.ReportResult{
			Success:    false,
			AnalysisID: analysisID,
			Files:      domain.ReportFiles{},
			Error:      objectsErr.Error(),
			StatusCode: analysis.MapStatus(objectsStatus),
		}, nil
	}

	log.Info().
		Str("analysis_id", analysisID).
		Str("phase", "get_objects").
		Int("object_count", len(objects)).
		Dur("elapsed", time.Since(genStart)).
		Msg("report generate: objects loaded")
	nObj := len(objects)

	reportCtx := ReportContextFromAnalysis(analysisResult, analysisID)
	attachAnalysisImageURLs(&reportCtx, analysisResult)

	cs, reps, dist, _ := EnrichReportContext(
		&reportCtx,
		objects,
		DefaultEnrichOptions,
	)
	attachRepresentativeImageURLs(reps, objects)
	prefetchReportImages(ctx, &reportCtx, reps)

	page := s.pageParams(analysisID, &reportCtx, cs, reps, dist, nObj)

	var (
		csvBytes  []byte
		pdfBytes  []byte
		htmlBytes int
		csvErr    error
		pdfErr    error
	)
	var artifactsWG sync.WaitGroup
	artifactsWG.Add(2)
	go func() {
		defer artifactsWG.Done()
		csvErr = observability.RunPhase(ctx, "reports.build_csv", func(cctx context.Context) error {
			var e error
			csvBytes, e = BuildCSV(&reportCtx)
			return e
		}, attribute.String("reports.analysis_id", analysisID))
	}()
	go func() {
		defer artifactsWG.Done()
		var html string
		html, pdfBytes, pdfErr = s.buildReportPDF(ctx, analysisID, page)
		htmlBytes = len(html)
	}()
	artifactsWG.Wait()

	if csvErr != nil {
		observability.RecordError(ctx, csvErr,
			attribute.String("reports.phase", "build_csv"),
			attribute.String("reports.analysis_id", analysisID),
		)
		log.Error().
			Err(csvErr).
			Str("analysis_id", analysisID).
			Str("phase", "build_csv").
			Dur("elapsed", time.Since(genStart)).
			Msg("report generate failed")
		return domain.ReportResult{}, csvErr
	}
	if pdfErr != nil {
		observability.RecordError(ctx, pdfErr,
			attribute.String("reports.phase", "convert_pdf"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.Int("reports.html_bytes", htmlBytes),
		)
		log.Error().
			Err(pdfErr).
			Str("analysis_id", analysisID).
			Str("phase", "convert_pdf").
			Dur("elapsed", time.Since(genStart)).
			Int("html_bytes", htmlBytes).
			Msg("report generate failed")
		return domain.ReportResult{}, pdfErr
	}

	log.Info().
		Str("analysis_id", analysisID).
		Str("phase", "build_artifacts").
		Int("html_bytes", htmlBytes).
		Int("csv_bytes", len(csvBytes)).
		Int("pdf_bytes", len(pdfBytes)).
		Dur("elapsed", time.Since(genStart)).
		Msg("report generate: CSV and PDF ready")

	csvName, err := storage.ReportObjectKey(analysisID, "csv")
	if err != nil {
		log.Error().Err(err).Str("analysis_id", analysisID).Msg("report generate failed")
		return domain.ReportResult{}, err
	}
	pdfName, err := storage.ReportObjectKey(analysisID, "pdf")
	if err != nil {
		log.Error().Err(err).Str("analysis_id", analysisID).Msg("report generate failed")
		return domain.ReportResult{}, err
	}

	ownerMeta := map[string]string{
		"report-version":              s.csvFormatVer,
		storage.MetaReportOwnerUserID: strconv.FormatInt(reportResultUserID(analysisResult), 10),
	}
	pdfOwnerMeta := map[string]string{
		"report-version":              s.pdfFormatVer,
		storage.MetaReportOwnerUserID: strconv.FormatInt(reportResultUserID(analysisResult), 10),
	}

	var csvUploadErr, pdfUploadErr error
	var uploadWG sync.WaitGroup
	uploadWG.Add(2)
	go func() {
		defer uploadWG.Done()
		csvUploadErr = observability.RunPhase(
			ctx,
			"reports.upload_csv",
			func(cctx context.Context) error {
				return s.minio.UploadBuffer(cctx, csvName, csvBytes, "text/csv", ownerMeta)
			},
			attribute.String("reports.analysis_id", analysisID),
			attribute.String("reports.object_key", csvName),
		)
	}()
	go func() {
		defer uploadWG.Done()
		pdfUploadErr = observability.RunPhase(
			ctx,
			"reports.upload_pdf",
			func(cctx context.Context) error {
				return s.minio.UploadBuffer(
					cctx,
					pdfName,
					pdfBytes,
					"application/pdf",
					pdfOwnerMeta,
				)
			},
			attribute.String("reports.analysis_id", analysisID),
			attribute.String("reports.object_key", pdfName),
		)
	}()
	uploadWG.Wait()

	if csvUploadErr != nil {
		observability.RecordError(ctx, csvUploadErr,
			attribute.String("reports.phase", "upload_csv"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.String("reports.object_key", csvName),
		)
		log.Error().
			Err(csvUploadErr).
			Str("analysis_id", analysisID).
			Str("phase", "upload_csv").
			Str("object_key", csvName).
			Dur("elapsed", time.Since(genStart)).
			Msg("report generate failed")
		return domain.ReportResult{}, csvUploadErr
	}
	if pdfUploadErr != nil {
		observability.RecordError(ctx, pdfUploadErr,
			attribute.String("reports.phase", "upload_pdf"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.String("reports.object_key", pdfName),
		)
		log.Error().
			Err(pdfUploadErr).
			Str("analysis_id", analysisID).
			Str("phase", "upload_pdf").
			Str("object_key", pdfName).
			Dur("elapsed", time.Since(genStart)).
			Msg("report generate failed")
		return domain.ReportResult{}, pdfUploadErr
	}

	log.Info().
		Str("analysis_id", analysisID).
		Int64("user_id", reportResultUserID(analysisResult)).
		Dur("duration", time.Since(genStart)).
		Int("objects", nObj).
		Int("html_bytes", htmlBytes).
		Int("csv_bytes", len(csvBytes)).
		Int("pdf_bytes", len(pdfBytes)).
		Str("csv_key", csvName).
		Str("pdf_key", pdfName).
		Msg("report generate completed")

	return domain.ReportResult{
		Success:    true,
		AnalysisID: analysisID,
		UserID:     reportResultUserID(analysisResult),
		Files: domain.ReportFiles{
			CSV: csvName,
			PDF: pdfName,
		},
	}, nil
}

func (s *Service) buildReportPDF(
	ctx context.Context,
	analysisID string,
	page view.PageParams,
) (html string, pdfBytes []byte, err error) {
	log := logger.WithTrace(ctx, logger.Logger)

	bodyHTML, bodyErr := render.RenderBody(view.BuildBody(page))
	if bodyErr != nil {
		log.Warn().
			Err(bodyErr).
			Str("analysis_id", analysisID).
			Str("phase", "render_body").
			Msg("report body template failed; continuing with empty body")
		bodyHTML = ""
	}
	data := render.ReportHTMLData{BodyHTML: bodyHTML}
	html, htmlErr := render.RenderReportHTML(data)
	if htmlErr != nil {
		log.Warn().
			Err(htmlErr).
			Str("analysis_id", analysisID).
			Str("phase", "render_html").
			Msg("report HTML template failed; using fallback")
		html = render.LoadTemplateFallback(render.MainHTMLPath(s.cfg.Templates.Dir))
	} else {
		html = render.InjectQRSVG(html, s.cfg.Templates.Dir)
	}

	log.Info().
		Str("analysis_id", analysisID).
		Str("phase", "render_html").
		Int("html_bytes", len(html)).
		Bool("body_template_ok", bodyErr == nil).
		Bool("shell_template_ok", htmlErr == nil).
		Msg("report generate: HTML ready")

	pdfBytes, err = s.pdf.ConvertHTMLToPDF(ctx, s.cfg.Templates.Dir, html)
	return html, pdfBytes, err
}

func (s *Service) pageParams(
	analysisID string,
	rc *reportcontext.ReportContext,
	cs calc.ClassStatisticsResult,
	reps []calc.RepresentativeGroup,
	dist map[string]string,
	nObj int,
) view.PageParams {
	if rc == nil {
		rc = &reportcontext.ReportContext{}
	}
	h := strings.TrimRight(strings.TrimSpace(s.cfg.AnalysisService.Host), "/")
	csvBase := strings.TrimRight(strings.TrimSpace(s.cfg.Server.PublicBaseURL), "/")
	csvURL := ""
	if csvBase != "" {
		csvURL = csvBase + "/api/reports/" + analysisID + "/csv"
	}
	return view.PageParams{
		Context:          *rc,
		ClassStats:       cs,
		Reps:             reps,
		Dist:             dist,
		Objects:          nObj,
		LogoSrc:          render.LogoRelPath,
		CsvURL:           csvURL,
		ObjectArchiveURL: h + "/analyses/" + analysisID + "/images/objects/archive",
		ImageBaseURL:     h + "/analyses/" + analysisID + "/images",
		ReportPackQuery:  s.reportPack.Query(analysisID),
		Img2:             rc.Img2,
		Img2DownloadIdx:  rc.Img2DownloadIndices,
	}
}

func (s *Service) PresignedDownloadURL(
	ctx context.Context,
	analysisID, token string,
	format string,
	subj authz.Subject,
) (domain.PresignedDownloadResponse, domain.ReportResult, error) {
	ft, err := storage.NormalizeReportFormat(format)
	if err != nil {
		observability.RecordStatus(ctx, err.Error(),
			attribute.String("reports.phase", "normalize_format"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.String("reports.format.input", format),
		)
		return domain.PresignedDownloadResponse{}, domain.ReportResult{
			Success:    false,
			AnalysisID: analysisID,
			Error:      err.Error(),
			StatusCode: 400,
		}, nil
	}

	needsGen, err := s.needsRegeneration(ctx, analysisID, []string{ft})
	if err != nil {
		observability.RecordError(ctx, err,
			attribute.String("reports.phase", "needs_regeneration_presign"),
			attribute.String("reports.analysis_id", analysisID),
		)
		z := logger.WithTrace(ctx, logger.Logger)
		z.Error().
			Err(err).
			Str("analysis_id", analysisID).
			Str("op", "presign_download").Msg("report presign: storage check failed")
		return domain.PresignedDownloadResponse{}, domain.ReportResult{}, err
	}

	if !needsGen {
		key, keyErr := storage.ReportObjectKey(analysisID, ft)
		if keyErr != nil {
			return domain.PresignedDownloadResponse{}, domain.ReportResult{}, keyErr
		}
		st, statErr := s.minio.GetFileStatus(ctx, key)
		if statErr != nil {
			return domain.PresignedDownloadResponse{}, domain.ReportResult{}, statErr
		}
		ownerStr := st.Metadata[storage.MetaReportOwnerUserID]
		if ownerStr == "" {
			observability.RecordStatus(ctx, "missing owner metadata",
				attribute.String("reports.phase", "presign_hot_path"),
				attribute.String("reports.analysis_id", analysisID),
			)
			return domain.PresignedDownloadResponse{}, domain.ReportResult{
				Success:    false,
				AnalysisID: analysisID,
				Error:      "report not found",
				StatusCode: 404,
			}, nil
		}
		ownerID, parseErr := strconv.ParseInt(ownerStr, 10, 64)
		if parseErr != nil {
			return domain.PresignedDownloadResponse{}, domain.ReportResult{
				Success:    false,
				AnalysisID: analysisID,
				Error:      "report not found",
				StatusCode: 404,
			}, nil
		}
		if !subj.CanPresignCachedReport(ownerID) {
			observability.RecordStatus(ctx, "forbidden cached report",
				attribute.String("reports.phase", "presign_hot_path_authz"),
				attribute.String("reports.analysis_id", analysisID),
			)
			return domain.PresignedDownloadResponse{}, domain.ReportResult{
				Success:    false,
				AnalysisID: analysisID,
				Error:      "you do not have access to this report",
				StatusCode: 403,
			}, nil
		}
		zHot := logger.WithTrace(ctx, logger.Logger)
		zHot.Info().
			Str("analysis_id", analysisID).
			Str("format", ft).
			Str("outcome", "cache_hit_presign").
			Msg("report presign")
		return s.issuePresignedURL(ctx, analysisID, ft)
	}

	result, ensureErr := s.EnsureCurrent(ctx, analysisID, token, []string{ft})
	if ensureErr != nil {
		observability.RecordError(ctx, ensureErr,
			attribute.String("reports.phase", "ensure_current"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.String("reports.op", "presign_download"),
		)
		z := logger.WithTrace(ctx, logger.Logger)
		z.Error().
			Err(ensureErr).
			Str("analysis_id", analysisID).
			Str("op", "presign_download").Msg("report presign failed")
		return domain.PresignedDownloadResponse{}, domain.ReportResult{}, ensureErr
	}
	if !result.Success {
		observability.RecordStatus(ctx, result.Error,
			attribute.String("reports.phase", "ensure_before_presign"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.Int("reports.response.status_code", result.StatusCode),
			attribute.String("reports.format", ft),
		)
		return domain.PresignedDownloadResponse{}, result, nil
	}
	zCold := logger.WithTrace(ctx, logger.Logger)
	zCold.Info().
		Str("analysis_id", analysisID).
		Str("format", ft).
		Str("outcome", "generate_then_presign").
		Msg("report presign")
	return s.issuePresignedURL(ctx, analysisID, ft)
}

func (s *Service) issuePresignedURL(
	ctx context.Context,
	analysisID, ft string,
) (domain.PresignedDownloadResponse, domain.ReportResult, error) {
	var raw string
	var ttl time.Duration
	if err := observability.RunPhase(
		ctx,
		"reports.minio.presign",
		func(cctx context.Context) error {
			var e error
			raw, ttl, e = s.minio.PresignedReportDownloadURL(cctx, analysisID, ft)
			return e
		},
		attribute.String("reports.analysis_id", analysisID),
		attribute.String("reports.format", ft),
	); err != nil {
		observability.RecordError(ctx, err,
			attribute.String("reports.phase", "presign_url"),
			attribute.String("reports.analysis_id", analysisID),
			attribute.String("reports.format", ft),
		)
		z := logger.WithTrace(ctx, logger.Logger)
		z.Error().
			Err(err).
			Str("analysis_id", analysisID).
			Str("format", ft).Str("op", "presign_download").Msg("minio presign failed")
		return domain.PresignedDownloadResponse{}, domain.ReportResult{}, err
	}
	resp := domain.PresignedDownloadResponse{
		URL:              raw,
		ExpiresInSeconds: int64(ttl.Round(time.Second) / time.Second),
	}
	result := domain.ReportResult{
		Success:    true,
		AnalysisID: analysisID,
		Files: domain.ReportFiles{
			CSV: analysisID + ".csv",
			PDF: analysisID + ".pdf",
		},
	}
	zPresign := logger.WithTrace(ctx, logger.Logger)
	zPresign.Info().
		Str("analysis_id", analysisID).
		Str("format", ft).
		Int64("expires_sec", resp.ExpiresInSeconds).
		Str("op", "presign_download").
		Msg("report presign: URL issued")
	return resp, result, nil
}

func (s *Service) needsRegeneration(
	ctx context.Context,
	analysisID string,
	fileTypes []string,
) (bool, error) {
	for _, t := range fileTypes {
		key, err := storage.ReportObjectKey(analysisID, t)
		if err != nil {
			return false, err
		}
		status, err := s.minio.GetFileStatus(ctx, key)
		if err != nil {
			return false, err
		}
		if !status.Exists {
			return true, nil
		}
		expected := s.csvFormatVer
		if t == "pdf" {
			expected = s.pdfFormatVer
		}
		version := status.Metadata["report-version"]
		if version == "" || version != expected {
			return true, nil
		}
	}
	return false, nil
}

func reportResultUserID(ar *analysis.AnalysisResult) int64 {
	if ar == nil {
		return 0
	}
	return ar.UserID
}
