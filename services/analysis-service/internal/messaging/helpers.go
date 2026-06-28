package messaging

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func (c *Client) setupPublishChannels(conn *amqp091.Connection) error {
	for i := range c.pubSlots {
		ch, err := conn.Channel()
		if err != nil {
			c.closePublishSlotsUpTo(i)
			return fmt.Errorf("open publish channel %d: %w", i, err)
		}
		if err := ch.Confirm(false); err != nil {
			_ = ch.Close()
			c.closePublishSlotsUpTo(i)
			return fmt.Errorf("enable confirms on channel %d: %w", i, err)
		}
		c.pubSlots[i] = publishSlot{
			ch:       ch,
			confirms: ch.NotifyPublish(make(chan amqp091.Confirmation, 100)),
		}
	}
	return nil
}

func (c *Client) closePublishSlotsUpTo(n int) {
	for i := range n {
		if c.pubSlots[i].ch != nil {
			_ = c.pubSlots[i].ch.Close()
			c.pubSlots[i] = publishSlot{}
		}
	}
}

func (c *Client) nextPublishSlot() *publishSlot {
	idx := c.pubSlotSeq.Add(1) % publishPoolSize
	return &c.pubSlots[idx]
}
