package http

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"csort.ru/classification-service/internal/observability"
	"csort.ru/classification-service/internal/transport/response"

	"github.com/gofiber/fiber/v3"
	fiberclient "github.com/gofiber/fiber/v3/client"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
)

const healthProbeTimeout = 2 * time.Second

var healthProbeClient = func() *fiberclient.Client {
	c := fiberclient.New()
	c.SetTimeout(healthProbeTimeout / 2)
	return c
}()

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db                    *pgxpool.Pool
	identityServiceURL    string
	correlationServiceURL string
	pinger                Pinger
}

func NewHealthHandler(
	db *pgxpool.Pool,
	identityServiceURL string,
	correlationServiceURL string,
) *HealthHandler {
	return &HealthHandler{
		db:                    db,
		identityServiceURL:    identityServiceURL,
		correlationServiceURL: correlationServiceURL,
	}
}

func (h *HealthHandler) GetHealth(c fiber.Ctx) error {
	if h.pinger != nil {
		return h.getHealthFromPinger(c)
	}

	ctx, cancel := context.WithTimeout(c.Context(), healthProbeTimeout)
	defer cancel()

	checks := make(map[string]string)
	allOk := true

	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = err.Error()
			allOk = false
		} else {
			checks["database"] = "ok"
		}
	}

	if h.identityServiceURL != "" {
		if err := pingHTTP(ctx, h.identityServiceURL); err != nil {
			checks["identity_service"] = err.Error()
			allOk = false
		} else {
			checks["identity_service"] = "ok"
		}
	}

	if h.correlationServiceURL != "" {
		if err := pingHTTP(ctx, h.correlationServiceURL); err != nil {
			checks["correlation_service"] = err.Error()
			allOk = false
		} else {
			checks["correlation_service"] = "ok"
		}
	}

	resp := fiber.Map{
		"service": "classification-service",
		"time":    time.Now().Format(time.RFC3339),
		"checks":  checks,
	}
	if allOk {
		resp["status"] = "healthy"
		return response.OK(c, resp)
	}
	resp["status"] = "unhealthy"
	return response.JSON(c, fiber.StatusServiceUnavailable, resp)
}

func pingHTTP(ctx context.Context, baseURL string) error {
	u, err := url.JoinPath(baseURL, "/health")
	if err != nil {
		return err
	}
	ctx, sp := observability.StartClientSpan(ctx, "health.http.probe",
		attribute.String("http.request.method", fiber.MethodGet),
		attribute.String("url.full", u),
	)
	if pu, perr := url.Parse(u); perr == nil && pu.Host != "" {
		sp.SetAttributes(attribute.String("server.address", pu.Host))
	}

	headers := map[string]string{}
	observability.InjectTraceHeaders(ctx, func(k, v string) { headers[k] = v })

	resp, err := healthProbeClient.Get(
		u,
		fiberclient.Config{ //nolint:gosec // G704: URL from config
			Ctx:    ctx,
			Header: headers,
		},
	)
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode()
	}
	if err != nil {
		observability.EndHTTPClientSpan(sp, statusCode, err)
		return err
	}
	defer resp.Close()
	if resp.StatusCode() != fiber.StatusOK {
		apiErr := &httpStatusError{status: resp.StatusCode()}
		observability.EndHTTPClientSpan(sp, resp.StatusCode(), apiErr)
		return apiErr
	}
	observability.EndHTTPClientSpan(sp, resp.StatusCode(), nil)
	return nil
}

type httpStatusError struct{ status int }

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.status)
}

func (h *HealthHandler) getHealthFromPinger(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), healthProbeTimeout)
	defer cancel()

	checks := make(map[string]string)
	if err := h.pinger.Ping(ctx); err != nil {
		checks["database"] = err.Error()
		resp := fiber.Map{
			"status":  "unhealthy",
			"service": "classification-service",
			"time":    time.Now().Format(time.RFC3339),
			"checks":  checks,
		}
		return response.JSON(c, fiber.StatusServiceUnavailable, resp)
	}
	checks["database"] = "ok"
	return response.OK(c, fiber.Map{
		"status":  "healthy",
		"service": "classification-service",
		"time":    time.Now().Format(time.RFC3339),
		"checks":  checks,
	})
}
