package middleware

import (
	"context"
	"net/http"

	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/observability"
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const requestTraceTracer = "csort.ru/analysis-service/http"

func RequestTrace() fiber.Handler {
	tr := otel.Tracer(requestTraceTracer)
	return func(c fiber.Ctx) error {
		parent := c.Context()
		if parent == nil {
			parent = context.Background()
		}

		hdr := make(http.Header)
		for k, vv := range c.GetReqHeaders() {
			for _, v := range vv {
				hdr.Add(k, v)
			}
		}
		ctx := otel.GetTextMapPropagator().Extract(parent, propagation.HeaderCarrier(hdr))

		routeTpl := c.Path()
		if r := c.Route(); r.Path != "" {
			routeTpl = r.Path
		}
		spanName := c.Method() + " " + routeTpl

		ctx, sp := tr.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Method()),
				semconv.HTTPRoute(routeTpl),
				attribute.String("url.path", c.Path()),
				attribute.String("url.full", c.OriginalURL()),
			),
		)
		defer sp.End()

		c.SetContext(ctx)

		err := c.Next()

		status := c.Response().StatusCode()
		sp.SetAttributes(semconv.HTTPResponseStatusCode(status))

		if err != nil {
			recAttrs := []attribute.KeyValue{
				semconv.HTTPRoute(routeTpl),
				semconv.HTTPResponseStatusCode(status),
				attribute.String("url.full", c.OriginalURL()),
			}
			if ae, ok := apierrors.From(err); ok {
				recAttrs = append(recAttrs, attribute.Int("app.api_error_code", ae.Code))
			}
			observability.RecordError(ctx, err, recAttrs...)
		} else if status >= 500 {
			observability.RecordError(ctx, observability.HTTPStatusError(status),
				semconv.HTTPRoute(routeTpl),
				semconv.HTTPResponseStatusCode(status),
			)
		}

		return err
	}
}
