package observability

import (
	"context"
	"strings"

	"csort.ru/analysis-service/internal/api/common"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
)

// UpstreamHTTPError records an upstream HTTP failure on the active span and emits a structured log.
func UpstreamHTTPError(
	ctx context.Context,
	log zerolog.Logger,
	peerService, operation, method, endpoint string,
	statusCode int,
	body []byte,
	err error,
) {
	if err == nil {
		return
	}

	bodyPreview := common.NewHTTPError(peerService, method, endpoint, statusCode, body, "", nil).
		BodyPreview(common.DefaultBodyPreviewLen)

	attrs := []attribute.KeyValue{
		attribute.String("peer.service", peerService),
		attribute.String("upstream.operation", operation),
		attribute.String("http.request.method", method),
		attribute.String("url.full", endpoint),
	}
	if statusCode > 0 {
		attrs = append(attrs, attribute.Int("http.response.status_code", statusCode))
	}
	if bodyPreview != "" {
		attrs = append(attrs, attribute.String("http.response.body.preview", bodyPreview))
	}

	RecordError(ctx, err, attrs...)

	evt := log.Error().Ctx(ctx).Err(err).
		Str("peer.service", peerService).
		Str("upstream.operation", operation).
		Str("http.request.method", method).
		Str("url.full", endpoint)
	if statusCode > 0 {
		evt = evt.Int("http.response.status_code", statusCode)
	}
	if bodyPreview != "" {
		evt = evt.Str("http.response.body.preview", bodyPreview)
	}
	evt.Msg("upstream HTTP request failed")
}

func TruncateBody(body []byte, maxLen int) string {
	if len(body) == 0 {
		return ""
	}
	if maxLen <= 0 {
		maxLen = common.DefaultBodyPreviewLen
	}
	s := strings.TrimSpace(string(body))
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
