package observability

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"csort.ru/reports-service/internal/logger"

	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	logglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InitOTEL(ctx context.Context, defaultServiceName string) (func(context.Context) error, error) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_SDK_DISABLED")), "true") {
		return func(context.Context) error { return nil }, nil
	}
	raw := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if raw == "" {
		return func(context.Context) error { return nil }, nil
	}

	host, insecure, err := parseOTELHTTPEndpoint(raw)
	if err != nil {
		return nil, fmt.Errorf("observability: OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
	}

	res, err := buildOTELResource(ctx, defaultServiceName)
	if err != nil {
		return nil, err
	}

	traceOpts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(host)}
	if insecure {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
	}
	traceExporter, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, fmt.Errorf("observability: otlp trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	logOpts := []otlploghttp.Option{otlploghttp.WithEndpoint(host)}
	if insecure {
		logOpts = append(logOpts, otlploghttp.WithInsecure())
	}
	logExporter, err := otlploghttp.New(ctx, logOpts...)
	if err != nil {
		_ = tp.Shutdown(ctx)
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		return nil, fmt.Errorf("observability: otlp log exporter: %w", err)
	}

	logProcessor := sdklog.NewBatchProcessor(logExporter)
	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(logProcessor),
	)
	logglobal.SetLoggerProvider(lp)
	logger.WrapGlobalWriter(wrapWriterWithOTLPJSONTap)

	return func(shutdownCtx context.Context) error {
		return errors.Join(
			tp.Shutdown(shutdownCtx),
			lp.Shutdown(shutdownCtx),
		)
	}, nil
}

func buildOTELResource(ctx context.Context, defaultServiceName string) (*resource.Resource, error) {
	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	attrs := []attribute.KeyValue{semconv.ServiceNameKey.String(serviceName)}
	attrs = append(attrs, parseOTELResourceAttributes()...)

	res, err := resource.New(ctx, resource.WithAttributes(attrs...))
	if err != nil {
		return nil, fmt.Errorf("observability: resource: %w", err)
	}
	return res, nil
}

func parseOTELHTTPEndpoint(raw string) (host string, insecure bool, err error) {
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false, err
	}
	if u.Host == "" {
		return "", false, errors.New("missing host in endpoint")
	}
	return u.Host, u.Scheme == "http", nil
}

func parseOTELResourceAttributes() []attribute.KeyValue {
	raw := strings.TrimSpace(os.Getenv("OTEL_RESOURCE_ATTRIBUTES"))
	if raw == "" {
		return nil
	}
	var out []attribute.KeyValue
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if k != "" {
			out = append(out, attribute.String(k, v))
		}
	}
	return out
}
