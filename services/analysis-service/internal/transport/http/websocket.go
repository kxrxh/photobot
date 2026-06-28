package http

import (
	"context"
	"fmt"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/transport/ws"
	websocket "github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

const (
	wsAuthTimeoutSeconds = 15
	wsAuthFailureCode    = 4001
)

type WebSocketHandler struct {
	log        zerolog.Logger
	hub        *ws.Hub
	authClient *auth.Client
}

func NewWebSocketHandler(
	log zerolog.Logger,
	hub *ws.Hub,
	authClient *auth.Client,
) *WebSocketHandler {
	return &WebSocketHandler{
		log:        log,
		hub:        hub,
		authClient: authClient,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c fiber.Ctx) error {
	return websocket.New(func(conn *websocket.Conn) {
		_ = conn.SetReadDeadline(time.Now().Add(wsAuthTimeoutSeconds * time.Second))

		var firstMsg struct {
			Type  string `json:"type"`
			Token string `json:"token"`
		}
		closeWithAuthFailure := func(msg string) {
			_ = conn.WriteJSON(map[string]string{"type": "auth_error", "message": msg})
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(wsAuthFailureCode, msg),
			)
		}

		if err := conn.ReadJSON(&firstMsg); err != nil {
			closeWithAuthFailure("expected auth message")
			return
		}

		if firstMsg.Type != "auth" || firstMsg.Token == "" {
			closeWithAuthFailure("expected auth message with token")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		validationResp, err := h.authClient.ValidateToken(ctx, firstMsg.Token)
		if err != nil || !validationResp.Valid || validationResp.Identity == nil {
			h.log.Warn().
				Str("remote_addr", conn.RemoteAddr().String()).
				Msg("websocket auth rejected")
			closeWithAuthFailure("invalid or expired token")
			return
		}

		pairs := userPlatformPairsFromIdentity(validationResp.Identity)
		if len(pairs) == 0 {
			closeWithAuthFailure("user has no messenger id (telegram or max)")
			return
		}

		wsUserIDs := make([]string, len(pairs))
		for i, p := range pairs {
			wsUserIDs[i] = fmt.Sprintf("%s:%s", p.Platform, p.UserID)
			h.hub.Register(wsUserIDs[i], conn)
		}
		conn.Locals(ws.LocalsUserIDKey, wsUserIDs[0])
		conn.Locals(ws.LocalsUserIDsKey, wsUserIDs)
		defer h.hub.Unregister(conn)

		_ = conn.WriteJSON(map[string]string{"type": "auth_ok"})

		h.log.Debug().
			Str("remote_addr", conn.RemoteAddr().String()).
			Strs("user_ids", wsUserIDs).
			Msg("websocket connection established")

		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		go h.keepAlive(conn)

		for {
			var msg map[string]any
			err := conn.ReadJSON(&msg)
			if err != nil {
				break
			}
		}
	}, websocket.Config{EnableCompression: false})(c)
}

func (h *WebSocketHandler) UpgradeMiddleware(c fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}
	return c.Next()
}

func (h *WebSocketHandler) keepAlive(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			return
		}
	}
}
