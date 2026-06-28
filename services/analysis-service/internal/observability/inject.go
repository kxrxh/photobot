package observability

import (
	"context"
	"net/http"

	"github.com/valyala/fasthttp"
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

func InjectTraceFasthttp(ctx context.Context, req *fasthttp.Request) {
	InjectTraceHeaders(ctx, func(k, v string) {
		req.Header.Add(k, v)
	})
}
