package apifiber

import (
	"net/http"

	"csort.ru/analysis-service/internal/api/common"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/logger"
	"github.com/gofiber/fiber/v3"
)

var errorHandlerLog = logger.GetLogger("api.errors")

const ErrorMessageLocalsKey = "error_message"

type errorResponse struct {
	Success bool `json:"success"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details,omitempty"`
		Path    string `json:"path"`
	} `json:"error"`
}

func ErrorHandler(c fiber.Ctx, err error) error {
	if ae, ok := apierrors.From(err); ok {
		logUpstreamCause(c, ae.Code, err)
		c.Locals(ErrorMessageLocalsKey, ae.Message)
		var resp errorResponse
		resp.Success = false
		resp.Error.Code = ae.Code
		resp.Error.Message = ae.Message
		resp.Error.Details = ae.Details
		resp.Error.Path = c.Path()
		return c.Status(ae.Code).JSON(resp)
	}
	return fiber.DefaultErrorHandler(c, err)
}

func logUpstreamCause(c fiber.Ctx, statusCode int, err error) {
	if statusCode < http.StatusInternalServerError {
		return
	}
	upstream := common.Extract(err)
	if upstream == nil {
		return
	}
	evt := errorHandlerLog.Error().Ctx(c.Context()).
		Int("http.response.status_code", statusCode).
		Str("upstream.peer.service", upstream.PeerService).
		Int("upstream.http.response.status_code", upstream.StatusCode).
		Str("upstream.url.full", upstream.Endpoint).
		Str("upstream.http.request.method", upstream.Method)
	if preview := upstream.BodyPreview(common.DefaultBodyPreviewLen); preview != "" {
		evt = evt.Str("upstream.http.response.body.preview", preview)
	}
	evt.Msg("request failed due to upstream service error")
}
