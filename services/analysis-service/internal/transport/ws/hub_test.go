package ws

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConn struct {
	mu sync.Mutex

	locals     map[string]interface{}
	writeCalls []interface{}
	closeCalls int
	writeErr   error
	closeErr   error
}

func newMockConn(locals map[string]interface{}) *mockConn {
	if locals == nil {
		locals = make(map[string]interface{})
	}
	return &mockConn{locals: locals}
}

func (m *mockConn) Locals(key string, value ...interface{}) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(value) > 0 {
		m.locals[key] = value[0]
		return value[0]
	}
	return m.locals[key]
}

func (m *mockConn) WriteJSON(v interface{}) error {
	m.mu.Lock()
	m.writeCalls = append(m.writeCalls, v)
	err := m.writeErr
	m.mu.Unlock()
	return err
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	m.closeCalls++
	err := m.closeErr
	m.mu.Unlock()
	return err
}

func (m *mockConn) writeCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.writeCalls)
}

func (m *mockConn) lastWriteCall() interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.writeCalls) == 0 {
		return nil
	}
	return m.writeCalls[len(m.writeCalls)-1]
}

func TestNewHub(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	hub := NewHub(ctx, rdb)
	require.NotNil(t, hub)
	hub.Shutdown()
}

func TestHub_Register(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)
	defer hub.Shutdown()

	conn := newMockConn(nil)
	hub.Register("user1", conn)

	hub.clientsMux.RLock()
	userConns, ok := hub.clients["user1"]
	hub.clientsMux.RUnlock()

	require.True(t, ok)
	assert.Len(t, userConns, 1)
	hub.clientsMux.RLock()
	_, hasWriter := hub.connWriters[conn]
	hub.clientsMux.RUnlock()
	assert.True(t, hasWriter)

	conn2 := newMockConn(nil)
	hub.Register("user1", conn2)

	hub.clientsMux.RLock()
	userConns = hub.clients["user1"]
	hub.clientsMux.RUnlock()
	assert.Len(t, userConns, 2)

	hub.Register("user2", conn)
	hub.clientsMux.RLock()
	assert.Len(t, hub.clients["user2"], 1)
	hub.clientsMux.RUnlock()
}

func TestHub_Unregister_SingleUserID(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)
	defer hub.Shutdown()

	conn := newMockConn(map[string]interface{}{
		LocalsUserIDKey: "user1",
	})
	hub.Register("user1", conn)

	hub.Unregister(conn)

	hub.clientsMux.RLock()
	_, ok := hub.clients["user1"]
	hub.clientsMux.RUnlock()
	assert.False(t, ok)
}

func TestHub_Unregister_MultipleUserIDs(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)
	defer hub.Shutdown()

	conn := newMockConn(map[string]interface{}{
		LocalsUserIDsKey: []string{"tg:123", "max:456"},
	})
	hub.Register("tg:123", conn)
	hub.Register("max:456", conn)

	hub.Unregister(conn)

	hub.clientsMux.RLock()
	_, ok1 := hub.clients["tg:123"]
	_, ok2 := hub.clients["max:456"]
	hub.clientsMux.RUnlock()
	assert.False(t, ok1)
	assert.False(t, ok2)
}

func TestHub_Unregister_NoUserIDInLocals(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)
	defer hub.Shutdown()

	conn := newMockConn(nil)
	hub.Register("user1", conn)

	hub.Unregister(conn)

	hub.clientsMux.RLock()
	userConns, ok := hub.clients["user1"]
	hub.clientsMux.RUnlock()
	assert.True(t, ok)
	assert.Len(t, userConns, 1)
}

func TestHub_Unregister_PrefersUserIDsOverUserID(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)
	defer hub.Shutdown()

	conn := newMockConn(map[string]interface{}{
		LocalsUserIDKey:  "single",
		LocalsUserIDsKey: []string{"tg:1", "max:2"},
	})
	hub.Register("tg:1", conn)
	hub.Register("max:2", conn)

	hub.Unregister(conn)

	hub.clientsMux.RLock()
	_, ok1 := hub.clients["tg:1"]
	_, ok2 := hub.clients["max:2"]
	hub.clientsMux.RUnlock()
	assert.False(t, ok1)
	assert.False(t, ok2)
}

func TestHub_BroadcastToUser(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)
	defer hub.Shutdown()

	sub := rdb.Subscribe(ctx, RedisUserChannelPrefix+"user1")
	defer func() { _ = sub.Close() }()

	msg := Message{
		Type:      MessageTypeRequestUpdate,
		RequestID: "req-123",
		Data:      map[string]string{"status": "done"},
	}
	hub.BroadcastToUser(ctx, "user1", msg)

	received, err := sub.ReceiveMessage(ctx)
	require.NoError(t, err)
	assert.Equal(t, RedisUserChannelPrefix+"user1", received.Channel)
	assert.Contains(t, received.Payload, "request_update")
	assert.Contains(t, received.Payload, "req-123")
}

func TestHub_Shutdown(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)

	conn := newMockConn(map[string]interface{}{LocalsUserIDKey: "user1"})
	hub.Register("user1", conn)

	hub.Shutdown()

	assert.Equal(t, 1, conn.closeCalls)
	hub.clientsMux.RLock()
	assert.Empty(t, hub.clients)
	hub.clientsMux.RUnlock()
}

func TestHub_Shutdown_CloseError(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)

	conn := newMockConn(map[string]interface{}{LocalsUserIDKey: "user1"})
	conn.closeErr = errors.New("close failed")
	hub.Register("user1", conn)

	hub.Shutdown()

	hub.clientsMux.RLock()
	assert.Empty(t, hub.clients)
	hub.clientsMux.RUnlock()
}

func TestHub_ListenRedis_ForwardsToLocalClients(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)
	defer hub.Shutdown()

	conn := newMockConn(map[string]interface{}{LocalsUserIDKey: "user1"})
	hub.Register("user1", conn)

	payload := `{"type":"request_update","request_id":"req-xyz","data":{"status":"completed"}}`
	require.Eventually(t, func() bool {
		_ = rdb.Publish(ctx, RedisUserChannelPrefix+"user1", payload).Err()
		return conn.writeCallCount() >= 1
	}, 5*time.Second, 25*time.Millisecond)

	assert.GreaterOrEqual(t, conn.writeCallCount(), 1)
	last := conn.lastWriteCall()
	require.NotNil(t, last)
	msg, ok := last.(Message)
	require.True(t, ok)
	assert.Equal(t, MessageTypeRequestUpdate, msg.Type)
	assert.Equal(t, "req-xyz", msg.RequestID)
}

func TestHub_ListenRedis_IgnoresNonLocalUser(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	hub := NewHub(ctx, rdb)
	defer hub.Shutdown()

	conn := newMockConn(map[string]interface{}{LocalsUserIDKey: "user1"})
	hub.Register("user1", conn)

	payload := `{"type":"request_update","request_id":"req-other"}`
	err := rdb.Publish(ctx, RedisUserChannelPrefix+"user2", payload).Err()
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, conn.writeCallCount())
}
