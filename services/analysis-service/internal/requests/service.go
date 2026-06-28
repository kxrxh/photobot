package requests

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"csort.ru/analysis-service/internal/api/classification"
	"csort.ru/analysis-service/internal/api/common"
	"csort.ru/analysis-service/internal/api/worker"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/identity"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/observability"
	"csort.ru/analysis-service/internal/repository/requests"
	"csort.ru/analysis-service/internal/storage"
	"csort.ru/analysis-service/internal/transport/ws"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kxrxh/gopt"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
)

const requestsCleanupBatchSize int32 = 500

type Service struct {
	repo           *requests.Queries
	pool           *pgxpool.Pool
	analysisWorker *worker.Client
	classification *classification.Client
	wsHub          *ws.Hub
	tempStorage    *storage.Client
	logger         zerolog.Logger
	finalizeCh     chan finalizeTask
	finalizeOnce   sync.Once
}

func generateLockID(key string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	sum := h.Sum64()
	if sum > 1<<63-1 {
		sum %= (1 << 63)
	}
	return int64(sum)
}

func (s *Service) tryAcquireConfirmLock(
	ctx context.Context,
	requestID string,
) (release func(), acquired bool) {
	if s.pool == nil {
		return func() {}, true
	}
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("requestID", requestID).
			Msg("failed to acquire db connection for confirm lock")
		return func() {}, true
	}

	lockID := generateLockID("requests:confirm:" + requestID)
	var ok bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&ok); err != nil {
		s.logger.Warn().Err(err).Str("requestID", requestID).Msg("failed to acquire confirm lock")
		conn.Release()
		return func() {}, true
	}
	if !ok {
		conn.Release()
		return func() {}, false
	}

	return func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		var released bool
		if err := conn.QueryRow(releaseCtx, "SELECT pg_advisory_unlock($1)", lockID).
			Scan(&released); err != nil {
			s.logger.Warn().
				Err(err).
				Str("requestID", requestID).
				Msg("failed to release confirm lock")
		}
		conn.Release()
	}, true
}

type RequestsServiceParams struct {
	Repo           *requests.Queries
	Pool           *pgxpool.Pool
	AnalysisClient *worker.Client
	Classification *classification.Client
	WebSocketHub   *ws.Hub
	MinIOClient    *storage.Client
}

func NewRequestsService(params RequestsServiceParams) *Service {
	return &Service{
		repo:           params.Repo,
		pool:           params.Pool,
		analysisWorker: params.AnalysisClient,
		classification: params.Classification,
		wsHub:          params.WebSocketHub,
		tempStorage:    params.MinIOClient,
		logger:         logger.GetLogger("requests.service"),
	}
}

func (s *Service) broadcastToUser(ctx context.Context, wsUserID string, msg ws.Message) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.BroadcastToUser(ctx, wsUserID, msg)
}

func (s *Service) deleteTempFiles(ctx context.Context, fileIDs []string) {
	if s.tempStorage == nil || len(fileIDs) == 0 {
		return
	}
	s.tempStorage.DeleteFiles(ctx, fileIDs)
}

func (s *Service) List(ctx context.Context, params GetRequestsRequest) ([]Request, error) {
	return s.ListForPairs(ctx, params, nil)
}

func (s *Service) ListForPairs(
	ctx context.Context,
	params GetRequestsRequest,
	pairs []identity.UserPlatformPair,
) ([]Request, error) {
	limit := params.Limit
	offset := params.Offset

	if len(pairs) > 1 {
		return s.listMultiPlatform(ctx, params, pairs)
	}

	platform := "telegram"
	userID := params.UserID
	if len(pairs) == 1 {
		userID = pairs[0].UserID
		platform = pairs[0].Platform
	} else if params.Platform != nil && *params.Platform != "" {
		platform = *params.Platform
	}

	if params.Status != nil && *params.Status != "" {
		s.logger.Debug().
			Str("userID", userID).
			Str("platform", platform).
			Str("status", string(*params.Status)).
			Msg("getting requests from database by userID and platform and status")
		rows, err := s.repo.ListRequestsByUserIDAndPlatformAndStatus(
			ctx,
			requests.ListRequestsByUserIDAndPlatformAndStatusParams{
				UserID:   userID,
				Platform: platform,
				Status:   requests.RequestStatus(*params.Status),
				Limit:    limit,
				Offset:   offset,
			},
		)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("userID", userID).
				Msg("failed to get requests from database")
			return nil, apierrors.InternalWrap(err, "failed to get requests")
		}
		return listRowsToRequestsFromStatus(rows), nil
	}

	s.logger.Debug().
		Str("userID", userID).
		Str("platform", platform).
		Msg("getting requests from database by userID and platform")
	rows, err := s.repo.ListRequestsByUserIDAndPlatform(
		ctx,
		requests.ListRequestsByUserIDAndPlatformParams{
			UserID:   userID,
			Platform: platform,
			Limit:    limit,
			Offset:   offset,
		},
	)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("failed to get requests from database")
		return nil, apierrors.InternalWrap(err, "failed to get requests")
	}
	return listRowsToRequests(rows), nil
}

func (s *Service) listMultiPlatform(
	ctx context.Context,
	params GetRequestsRequest,
	pairs []identity.UserPlatformPair,
) ([]Request, error) {
	pairParams := platformPairListParams(pairs)
	if params.Status != nil && *params.Status != "" {
		rows, err := s.repo.ListRequestsByUserPlatformPairsAndStatus(
			ctx,
			requests.ListRequestsByUserPlatformPairsAndStatusParams{
				UserID1:   pairParams.UserID1,
				Platform1: pairParams.Platform1,
				UserID2:   pairParams.UserID2,
				Platform2: pairParams.Platform2,
				Status:    requests.RequestStatus(*params.Status),
				Limit:     params.Limit,
				Offset:    params.Offset,
			},
		)
		if err != nil {
			return nil, apierrors.InternalWrap(err, "failed to get requests")
		}
		return multiPlatformListRowsToRequestsFromStatus(rows), nil
	}

	rows, err := s.repo.ListRequestsByUserPlatformPairs(
		ctx,
		requests.ListRequestsByUserPlatformPairsParams{
			UserID1:   pairParams.UserID1,
			Platform1: pairParams.Platform1,
			UserID2:   pairParams.UserID2,
			Platform2: pairParams.Platform2,
			Limit:     params.Limit,
			Offset:    params.Offset,
		},
	)
	if err != nil {
		return nil, apierrors.InternalWrap(err, "failed to get requests")
	}
	return multiPlatformListRowsToRequests(rows), nil
}

type platformPairQueryParams struct {
	UserID1   string
	Platform1 string
	UserID2   string
	Platform2 string
}

func platformPairListParams(pairs []identity.UserPlatformPair) platformPairQueryParams {
	out := platformPairQueryParams{
		UserID1:   pairs[0].UserID,
		Platform1: pairs[0].Platform,
	}
	if len(pairs) > 1 {
		out.UserID2 = pairs[1].UserID
		out.Platform2 = pairs[1].Platform
	}
	return out
}

func (s *Service) TransferRequestOwnership(ctx context.Context, fromUserID, toUserID int32) error {
	if s.repo == nil {
		s.logger.Error().Msg("requests repository is not available")
		return apierrors.Internal("requests repository is not available")
	}
	fromStr := strconv.FormatInt(int64(fromUserID), 10)
	toStr := strconv.FormatInt(int64(toUserID), 10)
	if err := s.repo.MergeReassignRequests(ctx, requests.MergeReassignRequestsParams{
		UserID:   fromStr,
		UserID_2: toStr,
	}); err != nil {
		s.logger.Error().
			Err(err).
			Int32("from_user_id", fromUserID).
			Int32("to_user_id", toUserID).
			Msg("failed to transfer request ownership")
		return apierrors.InternalWrap(err, "failed to transfer ownership")
	}
	return nil
}

func dbRequestToRequest(req requests.Request) Request {
	return Request{
		ID:             req.ID,
		UserID:         req.UserID,
		Platform:       req.Platform,
		Product:        req.Product,
		Status:         RequestStatus(req.Status),
		Year:           gopt.FromPtr(req.Year).UnwrapOr(""),
		MassLiter:      req.MassLiter,
		Location:       gopt.FromPtr(req.Location).UnwrapOr(""),
		Mass1000:       req.Mass1000,
		Mass:           req.Mass,
		Images:         req.Images,
		Classification: req.Classification,
		TempID:         gopt.FromPtr(req.TempID).UnwrapOr(""),
		ErrorMessage:   gopt.FromPtr(req.ErrorMessage).UnwrapOr(""),
		CreatedAt:      req.CreatedAt.Time,
		UpdatedAt:      req.UpdatedAt.Time,
	}
}

func (s *Service) GetByID(
	ctx context.Context,
	pairs []identity.UserPlatformPair,
	requestID string,
) (*Request, error) {
	s.logger.Debug().Str("requestID", requestID).Msg("getting request by ID")
	if len(pairs) == 0 {
		return nil, apierrors.Unauthorized("at least one user platform pair is required")
	}
	p1, p2 := pairs[0], pairs[0]
	if len(pairs) > 1 {
		p2 = pairs[1]
	}

	req, err := s.repo.GetRequestByIDAndUserPairs(ctx, requests.GetRequestByIDAndUserPairsParams{
		ID:         requestID,
		UserID:     p1.UserID,
		Platform:   p1.Platform,
		UserID_2:   p2.UserID,
		Platform_2: p2.Platform,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Str("requestID", requestID).Msg("request not found")
			return nil, apierrors.NotFound("request not found")
		}
		s.logger.Error().Err(err).Str("requestID", requestID).Msg("failed to get request by ID")
		return nil, apierrors.InternalWrap(err, "failed to get request")
	}

	r := dbRequestToRequest(req)
	return &r, nil
}

func (s *Service) Create(ctx context.Context, p *CreateParams) (requestID string, err error) {
	logEvent := s.logger.Debug().
		Str("product", p.Product).
		Str("user_id", p.UserID).
		Str("platform", p.Platform).
		Str("bot", p.Bot).
		Str("year", p.Year).
		Str("location", p.Location)

	if p.MassLiter != nil {
		logEvent = logEvent.Float64("massLiter", *p.MassLiter)
	}
	logEvent.Msg("creating request")
	requestID = uuid.New().String()

	var classificationJSON []byte
	if s.classification != nil {
		result, cerr := s.classification.GetUserActiveClassification(ctx, p.UserID, p.Platform)
		if cerr != nil {
			if errors.Is(cerr, context.Canceled) || errors.Is(cerr, context.DeadlineExceeded) {
				return "", apierrors.BadGatewayWrap(cerr, "classification service request failed")
			}
			if apiErr := common.Extract(cerr); apiErr != nil {
				s.logger.Error().Ctx(ctx).Err(cerr).
					Str("upstream.peer.service", apiErr.PeerService).
					Int("upstream.http.response.status_code", apiErr.StatusCode).
					Str("upstream.url.full", apiErr.Endpoint).
					Str("upstream.http.response.body.preview", apiErr.BodyPreview(common.DefaultBodyPreviewLen)).
					Msg("classification service returned an error")
				return "", apierrors.BadGatewayWrap(
					cerr,
					"classification service returned an error",
				)
			}
			return "", apierrors.BadGatewayWrap(cerr, "classification service request failed")
		}
		classificationJSON = result
	}

	if s.repo == nil {
		s.logger.Error().Msg("requests repository is not available")
		return "", errors.New("requests repository is not available")
	}
	if s.pool == nil {
		s.logger.Error().Msg("requests pool is not available")
		return "", errors.New("requests pool is not available")
	}

	defer func() {
		if err != nil {
			observability.RecordError(ctx, err,
				attribute.String("requests.operation", "create"),
				attribute.String("requests.request_id", requestID),
			)
		}
	}()

	var tx pgx.Tx
	tx, err = s.pool.Begin(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to begin transaction")
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				s.logger.Error().
					Err(rollbackErr).
					Str("requestID", requestID).
					Msg("failed to rollback transaction")
			}
		}
	}()

	txQueries := s.repo.WithTx(tx)

	imageIDs := make([]string, len(p.Images))
	for i, f := range p.Images {
		imageIDs[i] = f.ID
	}
	var imagesJSON []byte
	imagesJSON, err = sonic.Marshal(imageIDs)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to marshal images")
		return "", fmt.Errorf("failed to marshal images: %w", err)
	}

	year := p.Year
	_, err = txQueries.CreateRequest(ctx, requests.CreateRequestParams{
		ID:             requestID,
		UserID:         p.UserID,
		Platform:       p.Platform,
		Product:        p.Product,
		Images:         imagesJSON,
		Classification: classificationJSON,
		Year:           &year,
		MassLiter:      p.MassLiter,
		Location:       &p.Location,
		Mass1000:       p.Mass1000,
		Mass:           p.Mass,
		TempID:         nil,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create request")
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	message := CreateAnalysisMessage{
		Product:        p.Product,
		UserID:         p.UserID,
		Platform:       p.Platform,
		Bot:            p.Bot,
		Year:           p.Year,
		MassLiter:      p.MassLiter,
		Location:       p.Location,
		Mass1000:       p.Mass1000,
		Mass:           p.Mass,
		Images:         p.Images,
		Classification: classificationJSON,
		RequestID:      requestID,
	}

	var messageBytes []byte
	messageBytes, err = sonic.Marshal(message)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to marshal message")
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = txQueries.CreateOutboxMessage(ctx, requests.CreateOutboxMessageParams{
		Topic:   "detection_queue",
		Payload: messageBytes,
		Status:  "pending",
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create outbox message")
		return "", fmt.Errorf("failed to create outbox message: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to commit transaction")
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	createLog := s.logger.Info().
		Str("request_id", requestID).
		Str("product", p.Product).
		Str("user_id", p.UserID).
		Str("platform", p.Platform).
		Int("image_count", len(p.Images))
	if len(classificationJSON) == 0 {
		createLog.Bool("classification_present", false).
			Msg("requests flow: analysis request persisted; worker payload has no classification (omit or null)")
	} else {
		createLog.Bool("classification_present", true).
			Int("classification_bytes", len(classificationJSON)).
			RawJSON("classification", classificationJSON).
			Msg("requests flow: analysis request persisted; worker classification payload snapshot")
	}

	return requestID, nil
}

func buildErrorMessageFromNotify(input NotifyProcessingCompletionRequest) string {
	if len(input.Errors) > 0 {
		var parts []string
		for _, m := range input.Errors {
			for filename, msg := range m {
				if filename == "" {
					parts = append(parts, msg)
					continue
				}
				parts = append(parts, filename+": "+msg)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n")
		}
	}
	if input.Message != "" {
		return input.Message
	}
	return "Произошла ошибка при обработке файла."
}

func (s *Service) GetRequestOwnerByID(
	ctx context.Context,
	requestID string,
) (userID string, platform string, err error) {
	if s.repo == nil {
		s.logger.Error().Msg("requests repository is not available")
		return "", "", errors.New("requests repository is not available")
	}
	request, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		return "", "", err
	}
	return request.UserID, request.Platform, nil
}

func (s *Service) NotifyProcessingCompletion(
	ctx context.Context,
	input NotifyProcessingCompletionRequest,
) error {
	if s.repo == nil {
		s.logger.Error().Msg("requests repository is not available")
		return errors.New("requests repository is not available")
	}

	request, err := s.repo.GetRequest(ctx, input.RequestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Str("requestID", input.RequestID).Msg("request not found")
			return apierrors.NotFound("request not found")
		}
		s.logger.Error().Err(err).Str("requestID", input.RequestID).Msg("failed to get request")
		return apierrors.InternalWrap(err, "failed to get request")
	}
	if request.Status != requests.RequestStatusProcessing {
		s.logger.Warn().
			Str("requestID", input.RequestID).
			Str("status", string(request.Status)).
			Bool("success", input.Success).
			Int("error_maps", len(input.Errors)).
			Interface("errors", input.Errors).
			Msg("request is not in processing status")
		return apierrors.BadRequest("request is not in processing status")
	}

	if !input.Success {
		info := buildErrorMessageFromNotify(input)
		failEv := s.logger.Info().
			Str("requestID", input.RequestID).
			Interface("errors", input.Errors).
			Str("resolved_error_message", info)
		if input.Message != "" {
			failEv = failEv.Str("notify_message", input.Message)
		}
		failEv.Msg("notify processing completion: marking request as failed")
		if err := s.repo.MarkRequestAsFailed(ctx, requests.MarkRequestAsFailedParams{
			ID:           input.RequestID,
			ErrorMessage: &info,
		}); err != nil {
			s.logger.Error().
				Err(err).
				Str("requestID", input.RequestID).
				Msg("failed to mark request failed")
			return fmt.Errorf("failed to mark request failed: %w", err)
		}

		wsUserID := request.Platform + ":" + request.UserID
		s.broadcastToUser(ctx, wsUserID, ws.Message{
			Type:      ws.MessageTypeRequestUpdate,
			RequestID: input.RequestID,
			Data: map[string]any{
				"status":        "failed",
				"error_message": info,
			},
		})

		return nil
	}

	err = s.repo.MarkRequestAsWaitingForConfirmation(
		ctx,
		requests.MarkRequestAsWaitingForConfirmationParams{
			ID:     input.RequestID,
			TempID: &input.TempID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows in result set") {
			s.logger.Warn().Str("requestID", input.RequestID).Msg("request not found")
			return apierrors.NotFound("request not found")
		}
		s.logger.Error().
			Err(err).
			Str("requestID", input.RequestID).
			Msg("failed to update request waiting_for_confirmation")
		return fmt.Errorf("failed to update request waiting_for_confirmation: %w", err)
	}

	wsUserID := request.Platform + ":" + request.UserID
	s.broadcastToUser(ctx, wsUserID, ws.Message{
		Type:      ws.MessageTypeRequestUpdate,
		RequestID: input.RequestID,
		Data: map[string]any{
			"status":  "waiting_for_confirmation",
			"temp_id": input.TempID,
		},
	})

	return nil
}

func (s *Service) CleanupRequestFiles(ctx context.Context, requestID string) {
	if s.repo == nil {
		return
	}

	request, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		s.logger.Error().Err(err).Str("requestID", requestID).Msg("failed to get request")
		return
	}

	if len(request.Images) == 0 {
		s.logger.Warn().Str("requestID", requestID).Msg("request has no images")
		return
	}

	var imageIDs []string
	if err := sonic.Unmarshal(request.Images, &imageIDs); err != nil {
		s.logger.Error().Err(err).Str("requestID", requestID).Msg("failed to unmarshal images")
		return
	}

	s.deleteTempFiles(ctx, imageIDs)
}

func (s *Service) ConfirmRequest(
	ctx context.Context,
	pairs []identity.UserPlatformPair,
	requestID string,
	excludeObjects []string,
) error {
	log := logger.WithTrace(ctx, s.logger)
	if s.repo == nil {
		log.Error().Msg("requests repository is not available")
		return apierrors.Internal("requests repository is not available")
	}
	if len(pairs) == 0 {
		return apierrors.Unauthorized("at least one user platform pair is required")
	}

	release, acquired := s.tryAcquireConfirmLock(ctx, requestID)
	if !acquired {
		return apierrors.TooManyRequests(
			"request confirmation is already in progress, please retry",
		)
	}

	p1, p2 := pairs[0], pairs[0]
	if len(pairs) > 1 {
		p2 = pairs[1]
	}

	request, err := s.repo.GetRequestByIDAndUserPairs(
		ctx,
		requests.GetRequestByIDAndUserPairsParams{
			ID:         requestID,
			UserID:     p1.UserID,
			Platform:   p1.Platform,
			UserID_2:   p2.UserID,
			Platform_2: p2.Platform,
		},
	)
	if err != nil {
		release()
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn().Str("requestID", requestID).Msg("request not found")
			return apierrors.NotFoundf("request with id %s not found", requestID)
		}
		log.Error().Err(err).Str("requestID", requestID).Msg("failed to get request")
		return apierrors.InternalWrap(err, "failed to get request")
	}

	if request.Status != requests.RequestStatusWaitingForConfirmation {
		release()
		if request.Status == requests.RequestStatusCompleted {
			go s.CleanupRequestFiles(context.WithoutCancel(ctx), requestID)
			return nil
		}

		log.Warn().
			Str("requestID", requestID).
			Msg("request is not in waiting_for_confirmation status")
		return apierrors.BadRequestf(
			"request is not in waiting_for_confirmation status (current status: %s)",
			request.Status,
		)
	}

	release()

	if s.analysisWorker == nil {
		return apierrors.Internal("analysis worker client is not configured")
	}
	_, recalcErr := s.analysisWorker.RecalculateAnalysis(
		ctx,
		gopt.FromPtr(request.TempID).UnwrapOr(""),
		excludeObjects,
	)
	if recalcErr != nil {
		observability.RecordError(ctx, recalcErr,
			attribute.String("requests.operation", "confirm.recalculate_analysis"),
			attribute.String("requests.request_id", requestID),
		)
		log.Error().
			Err(recalcErr).
			Str("requestID", requestID).
			Msg("failed to recalculate analysis with external service")
		return apierrors.UpstreamFailed(recalcErr)
	}

	markParams := requests.MarkRequestAsCompletedIfWaitingParams{
		ID:     requestID,
		TempID: gopt.FromPtr(request.TempID).ToPointer(),
	}
	var markErr error
	var marked int64
	for attempt := 1; attempt <= 3; attempt++ {
		marked, markErr = s.repo.MarkRequestAsCompletedIfWaiting(ctx, markParams)
		if markErr == nil && marked > 0 {
			break
		}
		if markErr == nil && marked == 0 {
			markErr = fmt.Errorf("request %s is no longer waiting for confirmation", requestID)
			break
		}
		time.Sleep(time.Duration(attempt*100) * time.Millisecond)
	}
	if markErr != nil {
		observability.RecordError(ctx, markErr,
			attribute.String("requests.operation", "confirm.mark_completed"),
			attribute.String("requests.request_id", requestID),
		)
		log.Error().
			Err(markErr).
			Str("requestID", requestID).
			Msg("failed to mark request completed after retries; scheduling background retry")
		s.enqueueFinalize(ctx, finalizeTask{ctx: ctx, params: markParams})
		return apierrors.Wrap(
			markErr,
			http.StatusAccepted,
			"Confirmation accepted; finalizing in background. Please wait for status update.",
		)
	}

	go s.CleanupRequestFiles(context.WithoutCancel(ctx), requestID)

	wsUserID := request.Platform + ":" + request.UserID
	s.broadcastToUser(ctx, wsUserID, ws.Message{
		Type:      ws.MessageTypeRequestUpdate,
		RequestID: requestID,
		Data: map[string]any{
			"status":           "completed",
			"excluded_objects": excludeObjects,
		},
	})

	return nil
}

func (s *Service) GetRequestData(
	ctx context.Context,
	requestID string,
) (*AnalysisRequestData, error) {
	if s.repo == nil {
		s.logger.Error().Msg("requests repository is not available")
		return nil, apierrors.Internal("requests repository is not available")
	}

	row, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Str("requestID", requestID).Msg("request not found")
			return nil, apierrors.NotFound("request not found")
		}
		s.logger.Error().Err(err).Str("requestID", requestID).Msg("failed to get request")
		return nil, apierrors.InternalWrap(err, "failed to get request")
	}

	return &AnalysisRequestData{
		Product:   row.Product,
		UserID:    row.UserID,
		Platform:  row.Platform,
		TempID:    gopt.FromPtr(row.TempID).UnwrapOr(""),
		Status:    row.Status,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (s *Service) GetUserRequests(
	ctx context.Context,
	userID string,
	platform string,
) ([]UserAnalysisRequestInfo, error) {
	if s.repo == nil {
		s.logger.Error().Msg("requests repository is not available")
		return nil, apierrors.Internal("requests repository is not available")
	}

	rows, err := s.repo.ListRequestsByUserIDAndPlatform(
		ctx,
		requests.ListRequestsByUserIDAndPlatformParams{
			UserID:   userID,
			Platform: platform,
			Limit:    1000,
			Offset:   0,
		},
	)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID).Msg("failed to list user analysis requests")
		return nil, apierrors.InternalWrap(err, "failed to list user analysis requests")
	}

	results := make([]UserAnalysisRequestInfo, 0, len(rows))
	for _, row := range rows {
		results = append(results, UserAnalysisRequestInfo{
			RequestID: row.ID,
			Product:   row.Product,
			Status:    string(row.Status),
			CreatedAt: row.CreatedAt.Time,
		})
	}

	return results, nil
}

func (s *Service) DeleteOldRequests(ctx context.Context) error {
	if s.repo == nil {
		s.logger.Error().Msg("requests repository is not available")
		return apierrors.Internal("requests repository is not available")
	}

	for {
		batch, err := s.repo.GetOldRequestsBatch(ctx, requestsCleanupBatchSize)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to get old requests")
			return apierrors.InternalWrap(err, "failed to get old requests")
		}
		if len(batch) == 0 {
			s.logger.Debug().Msg("no old requests found")
			return nil
		}

		imageCount := 0
		for _, req := range batch {
			if len(req.Images) > 0 {
				imageCount++
			}
		}
		allImageIDs := make([]string, 0, imageCount*4)
		ids := make([]string, len(batch))
		for i, req := range batch {
			ids[i] = req.ID
			if len(req.Images) == 0 {
				continue
			}
			var imageIDs []string
			if err := sonic.Unmarshal(req.Images, &imageIDs); err == nil {
				allImageIDs = append(allImageIDs, imageIDs...)
			}
		}
		if len(allImageIDs) > 0 {
			s.deleteTempFiles(ctx, allImageIDs)
		}

		if err := s.repo.DeleteRequestsByIDs(ctx, ids); err != nil {
			s.logger.Error().Err(err).Msg("failed to delete old requests batch")
			return apierrors.InternalWrap(err, "failed to delete old requests")
		}

		if len(batch) < int(requestsCleanupBatchSize) {
			return nil
		}
	}
}

func (s *Service) CleanupStuckProcessingRequests(ctx context.Context) error {
	if s.repo == nil {
		s.logger.Error().Msg("requests repository is not available")
		return apierrors.Internal("requests repository is not available")
	}

	msg := "processing timed out after 10 minutes"

	for {
		stuckRequests, err := s.repo.GetStuckProcessingJobsBatch(ctx, requestsCleanupBatchSize)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to get stuck processing requests")
			return apierrors.InternalWrap(err, "failed to get stuck processing requests")
		}
		if len(stuckRequests) == 0 {
			s.logger.Debug().Msg("no stuck processing requests found")
			return nil
		}

		ids := make([]string, len(stuckRequests))
		for i, req := range stuckRequests {
			ids[i] = req.ID
		}
		if err := s.repo.MarkRequestsAsFailedByIDs(ctx, requests.MarkRequestsAsFailedByIDsParams{
			ErrorMessage: &msg,
			Column2:      ids,
		}); err != nil {
			s.logger.Error().Err(err).Msg("failed to mark stuck requests as failed")
			return apierrors.InternalWrap(err, "failed to mark stuck requests as failed")
		}

		for _, req := range stuckRequests {
			if len(req.Images) == 0 {
				continue
			}
			var imageIDs []string
			if err := sonic.Unmarshal(req.Images, &imageIDs); err != nil {
				s.logger.Error().
					Err(err).
					Str("requestID", req.ID).
					Msg("failed to unmarshal images for stuck request")
				continue
			}
			s.deleteTempFiles(ctx, imageIDs)
		}

		if len(stuckRequests) < int(requestsCleanupBatchSize) {
			return nil
		}
	}
}

func (s *Service) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		if err := s.DeleteOldRequests(ctx); err != nil {
			s.logger.Error().Err(err).Msg("failed to cleanup old requests")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
