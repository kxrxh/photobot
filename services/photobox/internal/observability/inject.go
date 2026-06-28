package observability

import (
	"context"
	"net/http"

	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectTraceHeaders writes W3C trace headers into outbound requests.
func InjectTraceHeaders(ctx context.Context, add func(key, val string)) {
	carrier := make(http.Header)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(carrier))
	for k, vals := range carrier {
		for _, v := range vals {
			add(k, v)
		}
	}
}

// InjectTraceFasthttp adds propagation headers to a fasthttp request.
func InjectTraceFasthttp(ctx context.Context, req *fasthttp.Request) {
	InjectTraceHeaders(ctx, req.Header.Add)
}

// InjectTraceHTTPRequest adds propagation headers to a net/http request.
func InjectTraceHTTPRequest(ctx context.Context, req *http.Request) {
	InjectTraceHeaders(ctx, req.Header.Add)
}
