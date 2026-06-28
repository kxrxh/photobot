package requests

import (
	"context"
	"time"

	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/messaging"
	"csort.ru/analysis-service/internal/repository/requests"
	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type outboxRepo interface {
	ClaimOutboxMessages(ctx context.Context, limit int32) ([]requests.OutboxMessage, error)
	ResetStuckClaimedOutboxMessages(ctx context.Context) error
	MarkRequestAsProcessing(ctx context.Context, id string) error
	MarkOutboxMessageAsPublished(ctx context.Context, id pgtype.UUID) error
	DeletePublishedOutboxMessages(ctx context.Context) error
}

type messagePublisher interface {
	Publish(ctx context.Context, msg messaging.Message) error
}

type OutboxRelay struct {
	repo      outboxRepo
	publisher messagePublisher
	rabbitCfg *config.RabbitMQConfig
	relayCfg  *config.OutboxRelayConfig
	logger    zerolog.Logger
}

func NewOutboxRelay(
	repo outboxRepo,
	publisher messagePublisher,
	rabbitCfg *config.RabbitMQConfig,
	relayCfg *config.OutboxRelayConfig,
) *OutboxRelay {
	cfg := relayCfg
	if cfg == nil {
		cfg = &config.OutboxRelayConfig{
			BatchSize:            10,
			PollIntervalSec:      1,
			CleanupIntervalHr:    1,
			MessageTTLHours:      24,
			MarkPublishedRetries: 3,
		}
	}
	return &OutboxRelay{
		repo:      repo,
		publisher: publisher,
		rabbitCfg: rabbitCfg,
		relayCfg:  cfg,
		logger:    logger.GetLogger("outbox.relay"),
	}
}

func (r *OutboxRelay) Start(ctx context.Context) {
	r.logger.Info().Msg("outbox relay started")
	interval := time.Duration(r.relayCfg.PollIntervalSec) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info().Msg("outbox relay stopped")
			return
		case <-ticker.C:
			if err := r.processMessages(ctx); err != nil {
				r.logger.Error().Err(err).Msg("process outbox messages failed")
			}
		}
	}
}

const maxOutboxConcurrency = 5

func (r *OutboxRelay) processMessages(ctx context.Context) error {
	if err := r.repo.ResetStuckClaimedOutboxMessages(ctx); err != nil {
		r.logger.Warn().Err(err).Msg("reset stuck outbox messages failed")
	}
	messages, err := r.repo.ClaimOutboxMessages(ctx, int32(r.relayCfg.BatchSize))
	if err != nil {
		return err
	}

	sem := make(chan struct{}, maxOutboxConcurrency)
	g, gctx := errgroup.WithContext(ctx)
	for _, msg := range messages {
		msg := msg
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()
			return r.processOneMessage(gctx, msg)
		})
	}
	return g.Wait()
}

func (r *OutboxRelay) processOneMessage(ctx context.Context, msg requests.OutboxMessage) error {
	if err := r.publishMessage(ctx, msg); err != nil {
		r.logger.Error().
			Err(err).
			Str("message_id", msg.ID.String()).
			Msg("publish outbox message failed; message will be retried after claim timeout")
		return err
	}

	if requestID := extractRequestID(msg.Payload, r.logger); requestID != "" {
		if err := r.repo.MarkRequestAsProcessing(ctx, requestID); err != nil {
			r.logger.Warn().
				Err(err).
				Str("message_id", msg.ID.String()).
				Str("request_id", requestID).
				Msg("mark request as processing failed")
		}
	}

	if err := r.markAsPublishedWithRetry(ctx, msg.ID, r.relayCfg.MarkPublishedRetries); err != nil {
		r.logger.Error().
			Err(err).
			Str("message_id", msg.ID.String()).
			Msg("mark outbox message published failed")
	}
	return nil
}

func (r *OutboxRelay) markAsPublishedWithRetry(
	ctx context.Context,
	msgID pgtype.UUID,
	maxAttempts int,
) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := r.repo.MarkOutboxMessageAsPublished(ctx, msgID); err == nil {
			return nil
		} else if attempt < maxAttempts {
			r.logger.Warn().
				Err(err).
				Str("message_id", msgID.String()).
				Int("attempt", attempt).
				Msg("mark outbox message published failed, retrying")
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		} else {
			return err
		}
	}
	return nil
}

func (r *OutboxRelay) publishMessage(ctx context.Context, msg requests.OutboxMessage) error {
	ttl := time.Duration(r.relayCfg.MessageTTLHours) * time.Hour
	return r.publisher.Publish(ctx, messaging.Message{
		Exchange:   r.rabbitCfg.RequestExchange,
		RoutingKey: r.rabbitCfg.RequestRoutingKey,
		Body:       msg.Payload,
		TTL:        &ttl,
	})
}

func (r *OutboxRelay) StartCleanup(ctx context.Context) {
	interval := time.Duration(r.relayCfg.CleanupIntervalHr) * time.Hour
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := r.repo.DeletePublishedOutboxMessages(ctx); err != nil {
			r.logger.Error().Err(err).Msg("cleanup outbox messages failed")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func extractRequestID(payload []byte, log zerolog.Logger) string {
	var v struct {
		RequestID string `json:"request_id"`
	}
	if err := sonic.Unmarshal(payload, &v); err != nil {
		log.Debug().Err(err).Msg("parse request_id from outbox payload failed")
		return ""
	}
	return v.RequestID
}
