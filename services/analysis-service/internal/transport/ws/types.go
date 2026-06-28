package ws

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	LocalsUserIDKey  = "user_id"
	LocalsUserIDsKey = "user_ids" // merged-account registrations: all ws user IDs
)

type MessageType string

const (
	MessageTypeRequestUpdate MessageType = "request_update"
)

type Message struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Data      any         `json:"data,omitempty"`
}

type wsConn interface {
	Locals(key string, value ...interface{}) interface{}
	WriteJSON(v interface{}) error
	Close() error
}

type Hub struct {
	redisClient  *redis.Client
	clients      map[string]map[*clientWriter]bool
	connWriters  map[wsConn]*clientWriter
	totalClients int
	clientsMux   sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}
