package observability

import (
	"context"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const authTracerName = "csort.ru/auth-service/auth"

// AuthTracer returns the tracer used for auth-domain spans.
func AuthTracer() trace.Tracer {
	return otel.Tracer(authTracerName)
}

// StartSpan starts a child span on AuthTracer and returns the derived context.
func StartSpan(
	ctx context.Context,
	name string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return AuthTracer().Start(ctx, name, opts...)
}

// StartClientSpan starts a CLIENT span for outbound HTTP to downstream services.
func StartClientSpan(
	ctx context.Context,
	name string,
	attrs ...attribute.KeyValue,
) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	return AuthTracer().Start(ctx, name, opts...)
}

// EndHTTPClientSpan completes an outbound HTTP CLIENT span (set statusCode when known).
func EndHTTPClientSpan(sp trace.Span, statusCode int, err error) {
	if statusCode > 0 {
		sp.SetAttributes(attribute.Int("http.response.status_code", statusCode))
	}
	if err != nil {
		EndSpan(sp, err)
		return
	}
	if statusCode >= 500 {
		sp.SetStatus(codes.Error, http.StatusText(statusCode))
	} else if statusCode >= 400 {
		sp.SetStatus(codes.Error, "HTTP "+strconv.Itoa(statusCode))
	}
	sp.End()
}

// EndSpan ends the span; if err is non-nil it records the error and sets span status.
func EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
