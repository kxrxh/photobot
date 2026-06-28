package services

import (
	"context"
	"sync"
)

type backgroundCancelLoop struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

func (b *backgroundCancelLoop) bind(cancel context.CancelFunc) {
	b.mu.Lock()
	b.cancel = cancel
	b.mu.Unlock()
}

func (b *backgroundCancelLoop) state() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cancel == nil {
		return "stopped", nil
	}
	return "running", nil
}

func (b *backgroundCancelLoop) stop() {
	b.mu.Lock()
	c := b.cancel
	b.cancel = nil
	b.mu.Unlock()
	if c != nil {
		c()
	}
}
