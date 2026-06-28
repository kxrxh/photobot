package http

import (
	"context"
	"reflect"
	"strconv"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/dto"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/middleware"
	"csort.ru/analysis-service/internal/observability"
	"csort.ru/analysis-service/internal/requests"
	"csort.ru/analysis-service/internal/transport/response"
	validatepkg "csort.ru/analysis-service/internal/validator"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

const httpRequestsComponent = "transport.http.requests"

type validationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type RequestsHandler struct {
	service   *requests.Service
	redis     *redis.Client
	validator *validator.Validate
}

func NewRequestsHandler(
	service *requests.Service,
	redisClient *redis.Client,
	v *validator.Validate,
) *RequestsHandler {
	return &RequestsHandler{
		service:   service,
		redis:     redisClient,
		validator: v,
	}
}

func deriveUserIDAndPlatform(identity *auth.Identity) (userID string, platform string) {
	if identity == nil {
		return "", "telegram"
	}
	if identity.MaxID != nil {
		return strconv.FormatInt(*identity.MaxID, 10), "max"
	}
	if identity.TelegramID != nil {
		return strconv.FormatInt(*identity.TelegramID, 10), "telegram"
	}
	return "", "telegram"
}

func userIDForPlatform(identity *auth.Identity, platform string) string {
	if identity == nil {
		return ""
	}
	switch platform {
	case "telegram":
		if identity.TelegramID != nil {
			return strconv.FormatInt(*identity.TelegramID, 10)
		}
	case "max":
		if identity.MaxID != nil {
			return strconv.FormatInt(*identity.MaxID, 10)
		}
	}
	return ""
}

func userPlatformPairsFromIdentity(identity *auth.Identity) []requests.UserPlatformPair {
	if identity == nil {
		return nil
	}
	var pairs []requests.UserPlatformPair
	if identity.TelegramID != nil {
		pairs = append(
			pairs,
			requests.UserPlatformPair{
				UserID:   strconv.FormatInt(*identity.TelegramID, 10),
				Platform: "telegram",
			},
		)
	}
	if identity.MaxID != nil {
		pairs = append(
			pairs,
			requests.UserPlatformPair{
				UserID:   strconv.FormatInt(*identity.MaxID, 10),
				Platform: "max",
			},
		)
	}
	return pairs
}

func (h *RequestsHandler) GetRequests(c fiber.Ctx) error {
	log := logger.Component(c, httpRequestsComponent)
	identity, ok := c.Locals(middleware.UserDataKey).(*auth.Identity)
	if !ok || identity == nil {
		log.Warn().Msg("get requests rejected: auth required")
		return apierrors.New(fiber.StatusUnauthorized, "Authentication required")
	}

	userIDStr, platform := deriveUserIDAndPlatform(identity)
	if userIDStr == "" {
		log.Warn().Msg("get requests rejected: no messenger id")
		return apierrors.New(fiber.StatusBadRequest, "user has no messenger id (telegram or max)")
	}

	var q dto.GetRequestsQueryRequest
	if err := c.Bind().Query(&q); err != nil {
		log.Error().Err(err).Msg("get requests rejected: invalid query params")
		return apierrors.Wrap(err, fiber.StatusBadRequest, "invalid query params")
	}

	platformOmitted := q.Platform == nil
	var params requests.GetRequestsRequest
	if q.Platform == nil {
		params = GetRequestsQueryRequestToParams(q, userIDStr, &platform)
	} else {
		uid := userIDForPlatform(identity, *q.Platform)
		if uid == "" {
			log.Warn().
				Str("platform", *q.Platform).
				Msg("get requests rejected: no id for platform")
			return apierrors.New(fiber.StatusBadRequest, "user has no id for platform "+*q.Platform)
		}
		params = GetRequestsQueryRequestToParams(q, uid, q.Platform)
	}

	if err := h.validator.Struct(q); err != nil {
		var validationErrors []validationErrorDetail
		for _, fe := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, validationErrorDetail{
				Field:   validatepkg.GetJSONFieldName(fe.Field(), reflect.TypeOf(q)),
				Message: validatepkg.Translate(fe),
			})
		}
		log.Error().
			Err(err).
			Interface("params", q).
			Msg("get requests rejected: validation failed")
		return apierrors.WithDetails(apierrors.BadRequest("Validation failed"), validationErrors)
	}

	pairs := userPlatformPairsFromIdentity(identity)
	if platformOmitted && ok && len(pairs) > 1 {
		listParams := GetRequestsQueryRequestToParams(q, userIDStr, &platform)
		reqs, err := h.service.ListForPairs(c.Context(), listParams, pairs)
		if err != nil {
			log.Error().Err(err).Msg("get requests failed")
			return err
		}
		log.Info().
			Str("flow", "requests.list").
			Bool("multi_platform_merge", true).
			Int("platform_count", len(pairs)).
			Int("returned", len(reqs)).
			Int32("limit", listParams.Limit).
			Int32("offset", listParams.Offset).
			Msg("requests flow: list completed")
		return response.OK(c, RequestsToGetRequestsResponse(reqs))
	}

	reqs, err := h.service.List(c.Context(), params)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", params.UserID).
			Str("platform", *params.Platform).
			Msg("get requests failed")
		return err
	}
	log.Info().
		Str("flow", "requests.list").
		Str("user_id", params.UserID).
		Str("platform", *params.Platform).
		Int("returned", len(reqs)).
		Int32("limit", params.Limit).
		Int32("offset", params.Offset).
		Msg("requests flow: list completed")
	return response.OK(c, RequestsToGetRequestsResponse(reqs))
}

func (h *RequestsHandler) GetRequestByID(c fiber.Ctx) error {
	log := logger.Component(c, httpRequestsComponent)
	identity, ok := c.Locals(middleware.UserDataKey).(*auth.Identity)
	if !ok || identity == nil {
		log.Warn().Msg("get request by id rejected: auth required")
		return apierrors.New(fiber.StatusUnauthorized, "Authentication required")
	}

	requestID := c.Params("requestId")
	if requestID == "" {
		log.Warn().Msg("get request by id rejected: missing request_id")
		return apierrors.New(fiber.StatusBadRequest, "Missing requestId parameter")
	}

	pairs := userPlatformPairsFromIdentity(identity)
	req, err := h.service.GetByID(c.Context(), pairs, requestID)
	if err != nil {
		log.Error().Err(err).Str("request_id", requestID).Msg("get request by id failed")
		return err
	}

	log.Info().
		Str("request_id", requestID).
		Str("status", string(req.Status)).
		Str("product", req.Product).
		Msg("requests flow: get by id completed")
	return response.OK(c, RequestToResponse(*req))
}

func (h *RequestsHandler) NotifyProcessingCompletion(c fiber.Ctx) error {
	log := logger.Component(c, httpRequestsComponent)
	var input requests.NotifyProcessingCompletionRequest

	if err := c.Bind().Body(&input); err != nil {
		log.Error().Err(err).Msg("notify processing completion rejected: invalid payload")
		return apierrors.Wrap(err, fiber.StatusBadRequest, "unable to parse request payload")
	}

	if err := h.validator.Struct(input); err != nil {
		var validationErrors []validationErrorDetail
		for _, fe := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, validationErrorDetail{
				Field:   validatepkg.GetJSONFieldName(fe.Field(), reflect.TypeOf(input)),
				Message: validatepkg.Translate(fe),
			})
		}
		log.Error().
			Err(err).
			Interface("input", input).
			Msg("notify processing completion rejected: validation failed")
		return apierrors.WithDetails(apierrors.BadRequest("Validation failed"), validationErrors)
	}

	notifyEv := log.Info().
		Str("flow", "requests.notify_processing").
		Str("request_id", input.RequestID).
		Str("temp_id", input.TempID).
		Bool("success", input.Success).
		Int("error_maps", len(input.Errors)).
		Interface("errors", input.Errors)
	if input.Message != "" {
		notifyEv = notifyEv.Str("notify_message", input.Message)
	}
	notifyEv.Msg("requests flow: processing completion callback accepted")

	userID, platform, err := h.service.GetRequestOwnerByID(c.Context(), input.RequestID)
	if err != nil {
		log.Warn().
			Err(err).
			Str("request_id", input.RequestID).
			Msg("notify processing completion: failed to fetch owner for limiter decrement")
	}

	if err := h.service.NotifyProcessingCompletion(c.Context(), input); err != nil {
		log.Error().
			Err(err).
			Str("request_id", input.RequestID).
			Bool("success", input.Success).
			Int("error_maps", len(input.Errors)).
			Interface("errors", input.Errors).
			Msg("notify processing completion failed")
		return err
	}

	platformUserIDKey := "limiter:task:"
	switch platform {
	case "max":
		platformUserIDKey += "max:" + userID
	case "telegram":
		platformUserIDKey += "telegram:" + userID
	default:
		platformUserIDKey = ""
	}
	if h.redis != nil && platformUserIDKey != "" {
		if err := middleware.DecrementRateLimit(
			c.Context(),
			h.redis,
			platformUserIDKey,
		); err != nil {
			log.Warn().
				Err(err).
				Str("request_id", input.RequestID).
				Str("rate_limit_key", platformUserIDKey).
				Msg("notify processing completion: failed to decrement pending limiter")
		}
	}

	log.Info().Str("request_id", input.RequestID).Msg("notify processing completion completed")
	return response.OK(
		c,
		dto.MessageResponse{Message: "Processing completion notified successfully"},
	)
}

func (h *RequestsHandler) ConfirmRequest(c fiber.Ctx) (err error) {
	ctx := context.WithoutCancel(c.Context())
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	defer func() {
		if err != nil {
			observability.RecordError(ctx, err,
				attribute.String("http.handler", "requests.confirm"),
			)
		}
	}()

	log := logger.Component(c, httpRequestsComponent)
	identity, ok := c.Locals(middleware.UserDataKey).(*auth.Identity)
	if !ok || identity == nil {
		log.Warn().Msg("confirm request rejected: auth required")
		return apierrors.New(fiber.StatusUnauthorized, "Authentication required")
	}

	var input requests.ConfirmAnalysisRequest

	if err := c.Bind().Body(&input); err != nil {
		log.Error().Err(err).Msg("confirm request rejected: invalid payload")
		return apierrors.Wrap(err, fiber.StatusBadRequest, "unable to parse request payload")
	}

	if err := h.validator.Struct(input); err != nil {
		var validationErrors []validationErrorDetail
		for _, fe := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, validationErrorDetail{
				Field:   validatepkg.GetJSONFieldName(fe.Field(), reflect.TypeOf(input)),
				Message: validatepkg.Translate(fe),
			})
		}
		log.Error().
			Err(err).
			Interface("input", input).
			Msg("confirm request rejected: validation failed")
		return apierrors.WithDetails(apierrors.BadRequest("Validation failed"), validationErrors)
	}

	pairs := userPlatformPairsFromIdentity(identity)
	log.Info().
		Str("flow", "requests.confirm").
		Str("request_id", input.RequestID).
		Int("excluded_objects", len(input.ExcludeObjects)).
		Msg("requests flow: confirm request started")
	if err := h.service.ConfirmRequest(
		ctx,
		pairs,
		input.RequestID,
		input.ExcludeObjects,
	); err != nil {
		if ae, ok := apierrors.From(err); ok && ae.Code == fiber.StatusAccepted {
			log.Info().
				Str("request_id", input.RequestID).
				Msg("confirm request accepted for background finalization")
			return response.Accepted(c, dto.MessageResponse{Message: ae.Message})
		}
		log.Error().Err(err).Str("request_id", input.RequestID).Msg("confirm request failed")
		return err
	}

	log.Info().Str("request_id", input.RequestID).Msg("confirm request completed")
	return response.OK(c, dto.MessageResponse{Message: "Request success confirmed successfully"})
}
