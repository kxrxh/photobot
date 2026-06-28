package messaging

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/logger"

	"github.com/rabbitmq/amqp091-go"
)

func NewClient(
	cfg *config.RabbitMQConfig,
	enableAutoRecovery bool,
	monitoringInterval time.Duration,
) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	client := &Client{
		config: cfg,
	}

	if err := client.Reconnect(); err != nil {
		return nil, err
	}

	if enableAutoRecovery {
		client.StartAutoRecovery(monitoringInterval)
	}

	return client, nil
}

func (c *Client) Publish(ctx context.Context, msg Message) error {
	if err := c.EnsureConnected(); err != nil {
		return err
	}

	slot := c.nextPublishSlot()
	slot.mu.Lock()
	defer slot.mu.Unlock()

	pub := amqp091.Publishing{
		ContentType:  "application/json",
		Body:         msg.Body,
		DeliveryMode: amqp091.Persistent,
		Timestamp:    time.Now(),
		Headers:      msg.Headers,
	}
	if msg.TTL != nil {
		pub.Expiration = strconv.FormatInt(msg.TTL.Milliseconds(), 10)
	}

	if err := slot.ch.PublishWithContext(
		ctx,
		msg.Exchange,
		msg.RoutingKey,
		false,
		false,
		pub,
	); err != nil {
		return err
	}

	select {
	case conf, ok := <-slot.confirms:
		if !ok {
			return errors.New("confirm channel closed")
		}
		if !conf.Ack {
			return fmt.Errorf("broker nacked publish (tag=%d)", conf.DeliveryTag)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) EnsureConnected() error {
	if c.IsConnected() {
		return nil
	}

	log := logger.GetLogger("rabbitmq.client")
	log.Warn().Msg("RabbitMQ connection lost, attempting to reconnect")

	return c.Reconnect()
}

func (c *Client) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var dialURL string
	if c.config.URL != "" {
		dialURL = c.config.URL
	} else {
		vhost := c.config.VHost
		if vhost == "" {
			vhost = "/"
		} else if !strings.HasPrefix(vhost, "/") {
			vhost = "/" + vhost
		}
		u := url.URL{
			Scheme: "amqp",
			User:   url.UserPassword(c.config.User, c.config.Password),
			Host:   fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
			Path:   vhost,
		}
		dialURL = u.String()
	}

	oldConn := c.connection
	oldChannels := make([]*amqp091.Channel, 0, publishPoolSize)
	for i := range c.pubSlots {
		if c.pubSlots[i].ch != nil {
			oldChannels = append(oldChannels, c.pubSlots[i].ch)
		}
	}

	conn, err := amqp091.Dial(dialURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	if err := c.setupExchangesWithConn(conn); err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to setup exchanges: %w", err)
	}

	if err := c.setupPublishChannels(conn); err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to setup publish channels: %w", err)
	}

	c.connection = conn

	for _, ch := range oldChannels {
		_ = ch.Close()
	}

	if oldConn != nil {
		_ = oldConn.Close()
	}

	return nil
}

func (c *Client) setupExchangesWithConn(conn *amqp091.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create ephemeral channel: %w", err)
	}
	defer func() { _ = ch.Close() }()

	if err := ch.ExchangeDeclare(
		c.config.RequestExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare request exchange: %w", err)
	}

	if c.config.RequestQueue != "" {
		if _, err := ch.QueueDeclare(
			c.config.RequestQueue,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to declare detection queue: %w", err)
		}

		if err := ch.QueueBind(
			c.config.RequestQueue,
			c.config.RequestRoutingKey,
			c.config.RequestExchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind detection queue: %w", err)
		}
	}

	return nil
}

func (c *Client) StartAutoRecovery(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.autoRecoveryEnabled {
		return
	}

	if interval <= 0 {
		interval = 30 * time.Second
	}

	c.monitoringInterval = interval
	c.stopChan = make(chan struct{})
	c.autoRecoveryEnabled = true

	go c.monitorConnection()
}

func (c *Client) monitorConnection() {
	ticker := time.NewTicker(c.monitoringInterval)
	defer ticker.Stop()

	log := logger.GetLogger("rabbitmq.client")

	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			enabled := c.autoRecoveryEnabled
			c.mu.RUnlock()

			if !enabled {
				return
			}

			if !c.IsConnected() {
				log.Warn().Msg("Connection lost, attempting automatic recovery")
				if err := c.Reconnect(); err != nil {
					log.Error().Err(err).Msg("Failed to automatically recover connection")
				} else {
					log.Info().Msg("Successfully recovered RabbitMQ connection")
				}
			}
		case <-c.stopChan:
			return
		}
	}
}

func (c *Client) StopAutoRecovery() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.autoRecoveryEnabled {
		return
	}

	close(c.stopChan)
	c.autoRecoveryEnabled = false
	c.stopChan = nil
}

func (c *Client) Close() error {
	c.StopAutoRecovery()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connection != nil {
		err := c.connection.Close()
		c.connection = nil
		return err
	}
	return nil
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connection != nil && !c.connection.IsClosed()
}
