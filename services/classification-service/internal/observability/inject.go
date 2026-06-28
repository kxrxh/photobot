package observability

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func InjectTraceHeaders(ctx context.Context, add func(key, val string)) {
	carrier := make(http.Header)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(carrier))
	for k, vals := range carrier {
		for _, v := range vals {
			add(k, v)
		}
	}
}

func InjectTraceHTTPRequest(ctx context.Context, req *http.Request) {
	InjectTraceHeaders(ctx, req.Header.Add)
}
