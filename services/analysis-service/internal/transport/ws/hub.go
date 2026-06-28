package ws

import (
	"context"

	"csort.ru/analysis-service/internal/logger"
	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

var log = logger.GetLogger("transport.ws.hub")

const (
	RedisUserChannelPrefix = "ws:user:"
)

func NewHub(ctx context.Context, redisClient *redis.Client) *Hub {
	hubCtx, cancel := context.WithCancel(ctx)
	hub := &Hub{
		redisClient: redisClient,
		clients:     make(map[string]map[*clientWriter]bool),
		connWriters: make(map[wsConn]*clientWriter),
		ctx:         hubCtx,
		cancel:      cancel,
	}

	go hub.listenRedis(hubCtx)

	return hub
}

func (hub *Hub) listenRedis(ctx context.Context) {
	pubsub := hub.redisClient.PSubscribe(ctx, RedisUserChannelPrefix+"*")
	defer func() {
		err := pubsub.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close Redis pubsub")
		}
	}()

	ch := pubsub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			userId := msg.Channel[len(RedisUserChannelPrefix):]

			hub.clientsMux.RLock()
			userConns, isLocal := hub.clients[userId]
			hub.clientsMux.RUnlock()

			if isLocal && len(userConns) > 0 {
				var wsMsg Message
				if err := sonic.Unmarshal([]byte(msg.Payload), &wsMsg); err != nil {
					log.Error().Err(err).Msg("Error marshaling WebSocket message")
					continue
				}

				for cw := range userConns {
					cw.enqueue(wsMsg)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *Hub) Shutdown() {
	h.cancel()

	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	log.Info().Msg("Shutting down WebSocket hub")

	for userID := range h.clients {
		delete(h.clients, userID)
	}
	for conn, cw := range h.connWriters {
		_ = conn.Close()
		cw.close()
	}
	h.connWriters = make(map[wsConn]*clientWriter)
	h.totalClients = 0
}

func (h *Hub) Register(userID string, conn wsConn) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	cw, exists := h.connWriters[conn]
	if !exists {
		cw = newClientWriter(conn)
		h.connWriters[conn] = cw
		h.totalClients++
	}
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*clientWriter]bool)
	}
	if !h.clients[userID][cw] {
		h.clients[userID][cw] = true
	}

	log.Debug().
		Str("userID", userID).
		Int("total_clients", h.totalClients).
		Msg("WebSocket client registered")
}

func (h *Hub) Unregister(conn wsConn) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	cw, hasWriter := h.connWriters[conn]

	if ids, ok := conn.Locals(LocalsUserIDsKey).([]string); ok && len(ids) > 0 {
		for _, userID := range ids {
			h.removeConnFromUser(userID, conn, cw)
		}
		if hasWriter {
			h.removeConnWriter(conn, cw)
		}
		return
	}

	userID, ok := conn.Locals(LocalsUserIDKey).(string)
	if !ok {
		log.Error().Msg("Failed to retrieve userID from connection locals")
		return
	}

	h.removeConnFromUser(userID, conn, cw)
	if hasWriter {
		h.removeConnWriter(conn, cw)
	}
}

func (h *Hub) removeConnFromUser(userID string, conn wsConn, cw *clientWriter) {
	userConns, exists := h.clients[userID]
	if !exists {
		return
	}
	if cw != nil {
		delete(userConns, cw)
	} else {
		for writer := range userConns {
			if writer.conn == conn {
				delete(userConns, writer)
				break
			}
		}
	}
	if len(userConns) == 0 {
		delete(h.clients, userID)
	}
}

func (h *Hub) removeConnWriter(conn wsConn, cw *clientWriter) {
	if cw == nil {
		return
	}
	for _, userConns := range h.clients {
		delete(userConns, cw)
	}
	delete(h.connWriters, conn)
	cw.close()
	if h.totalClients > 0 {
		h.totalClients--
	}
}

func (h *Hub) BroadcastToUser(ctx context.Context, userID string, message Message) {
	payload, err := sonic.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal message")
		return
	}

	channel := RedisUserChannelPrefix + userID

	if err := h.redisClient.Publish(ctx, channel, payload).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to publish to Redis")
	}
}
