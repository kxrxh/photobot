package cache

import (
	"sync"
	"time"
)

type TTLCache[K comparable, V any] struct {
	mu      sync.RWMutex
	entries map[K]ttlEntry[V]
	ttl     time.Duration
}

type ttlEntry[V any] struct {
	value     V
	expiresAt time.Time
}

func NewTTLCache[K comparable, V any](ttl time.Duration) *TTLCache[K, V] {
	c := &TTLCache[K, V]{
		entries: make(map[K]ttlEntry[V]),
		ttl:     ttl,
	}
	return c
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		var zero V
		return zero, false
	}
	if now.After(entry.expiresAt) {
		delete(c.entries, key)
		var zero V
		return zero, false
	}
	return entry.value, true
}

func (c *TTLCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = ttlEntry[V]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}
