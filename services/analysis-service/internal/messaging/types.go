package messaging

import (
	"sync"
	"sync/atomic"
	"time"

	"csort.ru/analysis-service/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

const publishPoolSize = 5

type publishSlot struct {
	ch       *amqp091.Channel
	confirms <-chan amqp091.Confirmation
	mu       sync.Mutex
}

type Client struct {
	config     *config.RabbitMQConfig
	connection *amqp091.Connection

	pubSlots   [publishPoolSize]publishSlot
	pubSlotSeq atomic.Uint32

	stopChan            chan struct{}
	monitoringInterval  time.Duration
	autoRecoveryEnabled bool
	mu                  sync.RWMutex
}

type Message struct {
	Exchange   string
	RoutingKey string
	Body       []byte
	Headers    map[string]any
	TTL        *time.Duration
}
