package observability

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	reportsTracerName = "csort.ru/reports-service"
	maxErrDescLen     = 2048
)

func Tracer() trace.Tracer {
	return otel.Tracer(reportsTracerName)
}

func StartSpan(
	ctx context.Context,
	name string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, opts...)
}

func StartClientSpan(
	ctx context.Context,
	name string,
	attrs ...attribute.KeyValue,
) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	return Tracer().Start(ctx, name, opts...)
}

func truncateErrDesc(s string) string {
	if len(s) <= maxErrDescLen {
		return s
	}
	return s[:maxErrDescLen] + "...(truncated)"
}

// annotateSpanError records err on sp (exception.type, exception event with stack trace, error status).
func annotateSpanError(sp trace.Span, err error) {
	if err == nil || !sp.IsRecording() {
		return
	}
	sp.SetAttributes(attribute.String("exception.type", fmt.Sprintf("%T", err)))
	sp.RecordError(err, trace.WithStackTrace(true))
	sp.SetStatus(codes.Error, truncateErrDesc(err.Error()))
}

func EndSpan(span trace.Span, err error) {
	if err != nil {
		annotateSpanError(span, err)
	}
	span.End()
}

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

// RecordError records err on the span from ctx: caller attrs, exception metadata, and stack capture on the exception event.
func RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	if err == nil {
		return
	}
	sp := trace.SpanFromContext(ctx)
	if !sp.IsRecording() {
		return
	}
	if len(attrs) > 0 {
		sp.SetAttributes(attrs...)
	}
	annotateSpanError(sp, err)
}

// RecordStatus marks the current span as failed without a Go error (domain / HTTP-mapped failure).
func RecordStatus(ctx context.Context, description string, attrs ...attribute.KeyValue) {
	sp := trace.SpanFromContext(ctx)
	if !sp.IsRecording() {
		return
	}
	if len(attrs) > 0 {
		sp.SetAttributes(attrs...)
	}
	sp.SetStatus(codes.Error, truncateErrDesc(description))
}

// RunPhase runs fn in a child span named name. The span ends with fn's error (recorded on the child span).
func RunPhase(
	ctx context.Context,
	name string,
	fn func(context.Context) error,
	attrs ...attribute.KeyValue,
) error {
	ctx, sp := StartSpan(ctx, name, trace.WithAttributes(attrs...))
	var err error
	defer func() { EndSpan(sp, err) }()
	err = fn(ctx)
	return err
}
