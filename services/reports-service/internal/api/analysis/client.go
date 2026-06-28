package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"csort.ru/reports-service/internal/observability"

	"github.com/gofiber/fiber/v3/client"
	"go.opentelemetry.io/otel/attribute"
)

type Client struct {
	hc *client.Client
}

func New(baseURL string) *Client {
	raw := strings.TrimSpace(baseURL)
	trimmed := strings.TrimSuffix(raw, "/")

	hc := client.New()
	hc.SetBaseURL(trimmed)
	hc.SetTimeout(30 * time.Second)

	return &Client{hc: hc}
}

func (c *Client) headerForRequest(ctx context.Context, bearerToken string) map[string]string {
	h := map[string]string{}
	if t := strings.TrimSpace(bearerToken); t != "" {
		h["Authorization"] = "Bearer " + t
	}
	observability.InjectTraceHeaders(ctx, func(k, v string) {
		h[k] = v
	})
	return h
}

// get performs a GET, returns status and body. For HTTP status >= 400, err is non-nil.
func (c *Client) get(
	ctx context.Context,
	path, bearerToken string,
) (code int, raw []byte, err error) {
	ctx, sp := observability.StartClientSpan(ctx, "analysis.api",
		attribute.String("http.request.method", "GET"),
		attribute.String("url.path", path),
	)
	defer func() { observability.EndHTTPClientSpan(sp, code, err) }()

	resp, err := c.hc.Get(path, client.Config{
		Ctx:    ctx,
		Header: c.headerForRequest(ctx, bearerToken),
	})
	if err != nil {
		return 0, nil, err
	}
	defer resp.Close()

	code = resp.StatusCode()
	raw = resp.Body()
	if code >= 400 {
		return code, raw, fmt.Errorf("GET %s: %d: %s", path, code, string(raw))
	}
	return code, raw, nil
}

func (c *Client) GetAnalysis(
	ctx context.Context,
	analysisID, bearerToken string,
) (result *AnalysisResult, code int, err error) {
	path := "/analyses/" + analysisID
	code, raw, err := c.get(ctx, path, bearerToken)
	if err != nil {
		return nil, code, err
	}

	var env AnalysisAPIResponse
	if err = json.Unmarshal(raw, &env); err != nil {
		return nil, code, err
	}
	if !env.Success {
		return nil, code, errors.New("analysis response success=false")
	}
	return &env.Result, code, nil
}

func (c *Client) GetObjects(
	ctx context.Context,
	analysisID, bearerToken string,
) (out []Object, code int, err error) {
	path := "/analyses/" + analysisID + "/objects"
	code, raw, err := c.get(ctx, path, bearerToken)
	if err != nil {
		return nil, code, err
	}

	var wrapped struct {
		Result []Object `json:"result"`
	}
	if err = json.Unmarshal(raw, &wrapped); err != nil {
		return nil, code, err
	}
	return wrapped.Result, code, nil
}

func MapStatus(upstream int) int {
	if upstream == 404 {
		return 404
	}
	if upstream == 401 {
		return 401
	}
	if upstream == 403 {
		return 403
	}
	if upstream >= 500 && upstream < 600 {
		return 502
	}
	return 500
}
