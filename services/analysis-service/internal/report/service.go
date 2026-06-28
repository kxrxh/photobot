package report

import (
	"context"
	"errors"
	"io"
	"net/http"

	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/observability"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
)

type HTTPStatusError interface {
	error
	HTTPStatus() int
}

type ReportsAPIClient interface {
	GenerateReport(ctx context.Context, analysisID string) (*ReportResponse, error)
	DownloadReportToWriter(
		ctx context.Context,
		analysisID, fileType string,
		dst io.Writer,
	) (int64, error)
}

type Service struct {
	reportsClient ReportsAPIClient
	logger        zerolog.Logger
}

func NewReportService(reportsClient ReportsAPIClient) *Service {
	return &Service{
		reportsClient: reportsClient,
		logger:        logger.GetLogger("report.service"),
	}
}

func (s *Service) Generate(ctx context.Context, analysisID string) (*ReportResponse, error) {
	s.logger.Info().Str("analysisID", analysisID).Msg("Generating report")

	reportResp, err := s.reportsClient.GenerateReport(ctx, analysisID)
	if err != nil {
		observability.RecordError(ctx, err,
			attribute.String("report.operation", "generate"),
			attribute.String("analysis.id", analysisID),
		)
		s.logger.Error().Err(err).Str("analysisID", analysisID).Msg("failed to generate report")
		var statusErr HTTPStatusError
		if errors.As(err, &statusErr) {
			if st := statusErr.HTTPStatus(); st >= 400 && st < 600 {
				return nil, apierrors.Wrap(err, st, "failed to generate report")
			}
		}
		return nil, apierrors.InternalWrap(err, "failed to generate report")
	}

	s.logger.Info().Str("analysisID", analysisID).Msg("Report generated successfully")
	return reportResp, nil
}

func (s *Service) DownloadFile(
	ctx context.Context,
	analysisID, fileType string,
	dst io.Writer,
) (int64, error) {
	if fileType != "csv" && fileType != "pdf" {
		s.logger.Warn().Str("fileType", fileType).Msg("unsupported file type")
		return 0, apierrors.BadRequestf("unsupported file type: %s", fileType)
	}

	n, err := s.reportsClient.DownloadReportToWriter(ctx, analysisID, fileType, dst)
	if err != nil {
		var statusErr HTTPStatusError
		if errors.As(err, &statusErr) {
			st := statusErr.HTTPStatus()
			attrs := []attribute.KeyValue{
				attribute.String("report.operation", "download"),
				attribute.String("analysis.id", analysisID),
				attribute.String("report.format", fileType),
				attribute.Int("http.status_code", st),
			}
			observability.RecordError(ctx, err, attrs...)
		} else {
			observability.RecordError(ctx, err,
				attribute.String("report.operation", "download"),
				attribute.String("analysis.id", analysisID),
				attribute.String("report.format", fileType),
			)
		}
	}
	if err != nil {
		var statusErr HTTPStatusError
		if errors.As(err, &statusErr) {
			st := statusErr.HTTPStatus()
			if st == http.StatusNotFound {
				return 0, apierrors.NotFound("report not found")
			}
			if st >= 400 && st < 600 {
				msg := "failed to download report file"
				switch st {
				case http.StatusBadGateway,
					http.StatusServiceUnavailable,
					http.StatusGatewayTimeout:
					msg = "report download failed: reports service is unavailable"
				}
				return n, apierrors.Wrap(err, st, msg)
			}
		}
		s.logger.Error().
			Err(err).
			Str("analysisID", analysisID).
			Str("fileType", fileType).
			Msg("failed to download report file")
		return n, apierrors.InternalWrap(err, "failed to download report file")
	}

	s.logger.Info().
		Str("analysisID", analysisID).
		Str("fileType", fileType).
		Int64("bytes", n).
		Msg("Report file downloaded successfully")

	return n, nil
}
