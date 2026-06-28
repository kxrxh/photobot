package handlers

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"csort.ru/reports-service/internal/authz"
	"csort.ru/reports-service/internal/config"
	"csort.ru/reports-service/internal/domain"
	"csort.ru/reports-service/internal/http/middleware"
	"csort.ru/reports-service/internal/logger"
	"csort.ru/reports-service/internal/objectstore"
	"csort.ru/reports-service/internal/observability"
	"csort.ru/reports-service/internal/reports"
	"csort.ru/reports-service/internal/storage"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

type ReportsHandler struct {
	service *reports.Service
	cfg     config.Config
	minio   *storage.MinIOService
}

func NewReportsHandler(
	service *reports.Service,
	minio *storage.MinIOService,
	cfg config.Config,
) *ReportsHandler {
	return &ReportsHandler{service: service, minio: minio, cfg: cfg}
}

func (h *ReportsHandler) Generate(c fiber.Ctx) error {
	analysisID, err := parseAnalysisID(c)
	if err != nil {
		return err
	}

	result, svcErr := h.service.EnsureCurrent(
		c.Context(),
		analysisID,
		middleware.BearerTokenFromFiber(c),
		nil,
	)
	if err := writeReportResultError(c, analysisID, "Generate", result, svcErr); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(domain.ReportSuccessResponse{
		Success:    true,
		AnalysisID: analysisID,
		Message:    "Report generated successfully",
	})
}

func (h *ReportsHandler) DownloadURL(c fiber.Ctx) error {
	analysisID, err := parseAnalysisID(c)
	if err != nil {
		return err
	}
	format := c.Query("format")
	if format == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(domain.ErrorResponse{Error: `query "format" is required (csv or pdf)`})
	}
	subj := authz.SubjectFromMapClaims(middleware.JWTMapClaimsFromFiber(c))
	resp, result, svcErr := h.service.PresignedDownloadURL(
		c.Context(),
		analysisID,
		middleware.BearerTokenFromFiber(c),
		format,
		subj,
	)
	if err := writeReportResultError(c, analysisID, "DownloadURL", result, svcErr); err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *ReportsHandler) DownloadCSVPackSigned(c fiber.Ctx) error {
	analysisID, err := parseAnalysisID(c)
	if err != nil {
		return err
	}
	secret := strings.TrimSpace(h.cfg.ReportPack.HMACSecret)
	if secret == "" {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(domain.ErrorResponse{Error: "report pack is not configured"})
	}
	if err := reports.VerifyPackQuery(
		secret,
		analysisID,
		c.Query("exp"),
		c.Query("sig"),
		time.Now(),
		3*time.Minute,
	); err != nil {
		return c.Status(fiber.StatusUnauthorized).
			JSON(domain.ErrorResponse{Error: "invalid or expired share link"})
	}
	key, ferr := storage.ReportObjectKey(analysisID, "csv")
	if ferr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: ferr.Error()})
	}
	st, statErr := h.minio.GetFileStatus(c.Context(), key)
	if statErr != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(domain.ErrorResponse{Error: statErr.Error()})
	}
	if !st.Exists {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "report not found"})
	}
	obj, openErr := h.minio.GetObjectReader(c.Context(), key)
	if openErr != nil {
		if errors.Is(openErr, objectstore.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).
				JSON(domain.ErrorResponse{Error: "report not found"})
		}
		return c.Status(fiber.StatusInternalServerError).
			JSON(domain.ErrorResponse{Error: openErr.Error()})
	}
	defer func() {
		if cerr := obj.Close(); cerr != nil {
			z := logger.WithTrace(c.Context(), logger.Logger)
			z.Warn().
				Err(cerr).
				Str("handler", "DownloadCSVPackSigned").
				Str("analysis_id", analysisID).
				Msg("minio object close failed")
		}
	}()
	info, infoErr := obj.Stat()
	if infoErr != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(domain.ErrorResponse{Error: infoErr.Error()})
	}
	fn := strings.ReplaceAll(analysisID, `"`, "_") + "_report.csv"
	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+fn+`"`)
	c.Set(fiber.HeaderCacheControl, "no-cache")
	if info.Size > 0 {
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(info.Size, 10))
	}
	_, copyErr := io.Copy(c.Response().BodyWriter(), obj)
	return copyErr
}

func writeReportResultError(
	c fiber.Ctx,
	analysisID, handlerName string,
	result domain.ReportResult,
	svcErr error,
) error {
	if svcErr != nil {
		z := logger.WithTrace(c.Context(), logger.Logger)
		observability.RecordError(c.Context(), svcErr,
			attribute.String("reports.handler", handlerName),
			attribute.String("reports.analysis_id", analysisID),
		)
		z.Error().
			Err(svcErr).Str("handler", handlerName).Str("analysis_id", analysisID).
			Msg("report handler failed")
		return c.Status(fiber.StatusInternalServerError).
			JSON(domain.ErrorResponse{Error: svcErr.Error()})
	}
	if !result.Success {
		status := result.StatusCode
		if status == 0 {
			status = fiber.StatusInternalServerError
		}
		z := logger.WithTrace(c.Context(), logger.Logger)
		observability.RecordStatus(c.Context(), result.Error,
			attribute.String("reports.handler", handlerName),
			attribute.String("reports.analysis_id", analysisID),
			attribute.Int("reports.response.status_code", status),
		)
		z.Warn().
			Str("handler", handlerName).
			Str("analysis_id", analysisID).
			Int("status", status).
			Str("error", result.Error).
			Msg("report handler: upstream or validation error")
		return c.Status(status).JSON(domain.ErrorResponse{Error: result.Error})
	}
	return nil
}

func parseAnalysisID(c fiber.Ctx) (string, error) {
	analysisID := c.Params("analysisId")
	if _, err := uuid.Parse(analysisID); err != nil {
		return "", c.Status(fiber.StatusBadRequest).
			JSON(domain.ErrorResponse{Error: "Invalid analysisId: must be a UUID"})
	}
	return analysisID, nil
}
