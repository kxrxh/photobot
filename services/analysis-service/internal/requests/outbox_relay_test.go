package requests

import (
	"context"
	"errors"
	"sync"
	"testing"

	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/messaging"
	"csort.ru/analysis-service/internal/repository/requests"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockOutboxRepo struct {
	mu sync.Mutex

	claimCalls           []int32
	resetStuckCalls      int
	markProcessingCalls  []string
	markPublishedCalls   []pgtype.UUID
	deletePublishedCalls int
	claimReturn          []requests.OutboxMessage
	claimErr             error
	resetStuckErr        error
	markProcessingErr    error
	markPublishedErr     error
	markPublishedAttempt int
	deletePublishedErr   error
}

func (m *mockOutboxRepo) ClaimOutboxMessages(
	ctx context.Context,
	limit int32,
) ([]requests.OutboxMessage, error) {
	m.mu.Lock()
	m.claimCalls = append(m.claimCalls, limit)
	calls := len(m.claimCalls)
	ret := m.claimReturn
	err := m.claimErr
	m.mu.Unlock()

	if err != nil {
		return nil, err
	}
	if calls == 1 && len(ret) > 0 {
		return ret, nil
	}
	return ret, nil
}

func (m *mockOutboxRepo) ResetStuckClaimedOutboxMessages(ctx context.Context) error {
	m.mu.Lock()
	m.resetStuckCalls++
	err := m.resetStuckErr
	m.mu.Unlock()
	return err
}

func (m *mockOutboxRepo) MarkRequestAsProcessing(ctx context.Context, id string) error {
	m.mu.Lock()
	m.markProcessingCalls = append(m.markProcessingCalls, id)
	err := m.markProcessingErr
	m.mu.Unlock()
	return err
}

func (m *mockOutboxRepo) MarkOutboxMessageAsPublished(ctx context.Context, id pgtype.UUID) error {
	m.mu.Lock()
	m.markPublishedCalls = append(m.markPublishedCalls, id)
	calls := len(m.markPublishedCalls)
	err := m.markPublishedErr
	failUntil := m.markPublishedAttempt
	m.mu.Unlock()

	if failUntil > 0 && calls < failUntil {
		return errors.New("simulated mark published error")
	}
	return err
}

func (m *mockOutboxRepo) DeletePublishedOutboxMessages(ctx context.Context) error {
	m.mu.Lock()
	m.deletePublishedCalls++
	err := m.deletePublishedErr
	m.mu.Unlock()
	return err
}

type mockPublisher struct {
	mu sync.Mutex

	publishCalls []messaging.Message
	publishErr   error
}

func (m *mockPublisher) Publish(ctx context.Context, msg messaging.Message) error {
	m.mu.Lock()
	m.publishCalls = append(m.publishCalls, msg)
	err := m.publishErr
	m.mu.Unlock()
	return err
}

func mustUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	_ = u.Scan(s)
	return u
}

func TestOutboxRelay_ProcessMessages_HappyPath(t *testing.T) {
	msgID := mustUUID("550e8400-e29b-41d4-a716-446655440000")
	payload := []byte(`{"request_id":"req-123","product":"wheat"}`)

	repo := &mockOutboxRepo{
		claimReturn: []requests.OutboxMessage{{
			ID:      msgID,
			Topic:   "detection_queue",
			Payload: payload,
			Status:  "claimed",
		}},
	}
	pub := &mockPublisher{}

	relay := NewOutboxRelay(repo, pub, &config.RabbitMQConfig{
		RequestExchange:   "ex",
		RequestRoutingKey: "rk",
	}, &config.OutboxRelayConfig{
		BatchSize:            10,
		MarkPublishedRetries: 3,
	})

	ctx := context.Background()
	err := relay.processMessages(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, len(repo.claimCalls))
	assert.Equal(t, int32(10), repo.claimCalls[0])
	assert.Equal(t, 1, len(pub.publishCalls))
	assert.Equal(t, payload, pub.publishCalls[0].Body)
	assert.Equal(t, "req-123", repo.markProcessingCalls[0])
	assert.Equal(t, 1, len(repo.markPublishedCalls))
	assert.Equal(t, msgID, repo.markPublishedCalls[0])
}

func TestOutboxRelay_ProcessMessages_PublishFails_ReturnsErrorAndLeavesClaimed(t *testing.T) {
	msgID := mustUUID("550e8400-e29b-41d4-a716-446655440001")
	payload := []byte(`{"request_id":"req-456"}`)

	repo := &mockOutboxRepo{
		claimReturn: []requests.OutboxMessage{{
			ID:      msgID,
			Topic:   "detection_queue",
			Payload: payload,
			Status:  "claimed",
		}},
	}
	pub := &mockPublisher{
		publishErr: errors.New("rabbitmq connection refused"),
	}

	relay := NewOutboxRelay(
		repo,
		pub,
		&config.RabbitMQConfig{},
		&config.OutboxRelayConfig{BatchSize: 5},
	)
	ctx := context.Background()

	err := relay.processMessages(ctx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "rabbitmq connection refused")

	assert.Equal(t, 1, len(pub.publishCalls))
	assert.Equal(t, 0, len(repo.markPublishedCalls))
	assert.Equal(t, 0, len(repo.markProcessingCalls))
}

func TestOutboxRelay_ProcessMessages_MarkRequestAsProcessingFails_StillMarksPublished(
	t *testing.T,
) {
	msgID := mustUUID("550e8400-e29b-41d4-a716-446655440002")
	payload := []byte(`{"request_id":"req-789"}`)

	repo := &mockOutboxRepo{
		claimReturn: []requests.OutboxMessage{{
			ID:      msgID,
			Topic:   "detection_queue",
			Payload: payload,
			Status:  "claimed",
		}},
		markProcessingErr: errors.New("request not found"),
	}
	pub := &mockPublisher{}

	relay := NewOutboxRelay(repo, pub, &config.RabbitMQConfig{}, &config.OutboxRelayConfig{
		BatchSize:            5,
		MarkPublishedRetries: 3,
	})
	ctx := context.Background()

	err := relay.processMessages(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, len(pub.publishCalls))
	assert.Equal(t, 1, len(repo.markProcessingCalls))
	assert.Equal(t, 1, len(repo.markPublishedCalls))
	assert.Equal(t, msgID, repo.markPublishedCalls[0])
}

func TestOutboxRelay_ProcessMessages_NoRequestID_SkipsMarkProcessing(t *testing.T) {
	msgID := mustUUID("550e8400-e29b-41d4-a716-446655440003")
	payload := []byte(`{"product":"wheat"}`)

	repo := &mockOutboxRepo{
		claimReturn: []requests.OutboxMessage{{
			ID:      msgID,
			Topic:   "detection_queue",
			Payload: payload,
			Status:  "claimed",
		}},
	}
	pub := &mockPublisher{}

	relay := NewOutboxRelay(repo, pub, &config.RabbitMQConfig{}, &config.OutboxRelayConfig{
		BatchSize:            5,
		MarkPublishedRetries: 3,
	})
	ctx := context.Background()

	err := relay.processMessages(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, len(pub.publishCalls))
	assert.Equal(t, 0, len(repo.markProcessingCalls))
	assert.Equal(t, 1, len(repo.markPublishedCalls))
}

func TestOutboxRelay_ProcessMessages_ClaimErr_ReturnsError(t *testing.T) {
	repo := &mockOutboxRepo{
		claimErr: errors.New("db connection failed"),
	}
	pub := &mockPublisher{}

	relay := NewOutboxRelay(repo, pub, &config.RabbitMQConfig{}, &config.OutboxRelayConfig{})
	ctx := context.Background()

	err := relay.processMessages(ctx)
	require.Error(t, err)
	assert.Equal(t, "db connection failed", err.Error())
	assert.Equal(t, 0, len(pub.publishCalls))
}

func TestOutboxRelay_ProcessMessages_EmptyBatch(t *testing.T) {
	repo := &mockOutboxRepo{
		claimReturn: []requests.OutboxMessage{},
	}
	pub := &mockPublisher{}

	relay := NewOutboxRelay(repo, pub, &config.RabbitMQConfig{}, &config.OutboxRelayConfig{})
	ctx := context.Background()

	err := relay.processMessages(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, len(repo.claimCalls))
	assert.Equal(t, 0, len(pub.publishCalls))
}

func TestOutboxRelay_ProcessMessages_CallsResetStuck(t *testing.T) {
	repo := &mockOutboxRepo{
		claimReturn: []requests.OutboxMessage{},
	}
	pub := &mockPublisher{}

	relay := NewOutboxRelay(repo, pub, &config.RabbitMQConfig{}, &config.OutboxRelayConfig{})
	ctx := context.Background()

	_ = relay.processMessages(ctx)

	assert.Equal(t, 1, repo.resetStuckCalls)
}

func TestOutboxRelay_ProcessMessages_MarkPublishedRetriesUntilSuccess(t *testing.T) {
	msgID := mustUUID("550e8400-e29b-41d4-a716-446655440004")
	payload := []byte(`{"request_id":"req-retry"}`)

	repo := &mockOutboxRepo{
		claimReturn: []requests.OutboxMessage{{
			ID:      msgID,
			Topic:   "detection_queue",
			Payload: payload,
			Status:  "claimed",
		}},
		markPublishedAttempt: 3,
	}
	pub := &mockPublisher{}

	relay := NewOutboxRelay(repo, pub, &config.RabbitMQConfig{}, &config.OutboxRelayConfig{
		BatchSize:            5,
		MarkPublishedRetries: 5,
	})
	ctx := context.Background()

	err := relay.processMessages(ctx)
	require.NoError(t, err)

	assert.Equal(t, 3, len(repo.markPublishedCalls))
	assert.Equal(t, msgID, repo.markPublishedCalls[0])
}

func TestExtractRequestID(t *testing.T) {
	log := zerolog.Nop()

	t.Run("valid json with request_id", func(t *testing.T) {
		payload := []byte(`{"request_id":"abc-123","product":"wheat"}`)
		assert.Equal(t, "abc-123", extractRequestID(payload, log))
	})

	t.Run("valid json without request_id", func(t *testing.T) {
		payload := []byte(`{"product":"wheat"}`)
		assert.Equal(t, "", extractRequestID(payload, log))
	})

	t.Run("invalid json", func(t *testing.T) {
		payload := []byte(`{invalid`)
		assert.Equal(t, "", extractRequestID(payload, log))
	})

	t.Run("empty payload", func(t *testing.T) {
		payload := []byte(`{}`)
		assert.Equal(t, "", extractRequestID(payload, log))
	})
}
