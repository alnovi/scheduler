package tasks

import (
	"context"
	"sync"
	"time"

	"github.com/alnovi/scheduler/v2"
)

type HandlerFn func(ctx context.Context) error

type Base struct {
	status  scheduler.Status
	timeout time.Duration
	delay   time.Duration
	lock    time.Duration
	handler HandlerFn
	mu      sync.RWMutex
}

func NewBase(handler HandlerFn, opts ...Option) (*Base, error) {
	if handler == nil {
		return nil, scheduler.ErrTaskHandlerIsEmpty
	}

	task := &Base{
		status:  scheduler.StatusPending,
		timeout: 0,
		delay:   0,
		lock:    0,
		handler: handler,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(task)
		}
	}

	return task, nil
}

func (b *Base) GetStatus() scheduler.Status {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

func (b *Base) SetStatus(status scheduler.Status) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
}

func (b *Base) Timeout() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.timeout
}

func (b *Base) Delay() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delay
}

func (b *Base) Lock() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lock
}

func (b *Base) Handle(ctx context.Context) error {
	if b.GetStatus() != scheduler.StatusPending {
		return nil
	}

	b.SetStatus(scheduler.StatusRunning)
	defer func() {
		if b.GetStatus() == scheduler.StatusRunning {
			b.SetStatus(scheduler.StatusPending)
		}
	}()

	return b.handler(ctx)
}

type Option func(b *Base)

func WithEnabled(enabled bool) Option {
	return func(b *Base) {
		if !enabled {
			b.status = scheduler.StatusStopped
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(b *Base) {
		if timeout > 0 {
			if timeout.Minutes() > 0 {
				timeout = timeout - (5 * time.Second)
			}
			b.timeout = timeout
		}
	}
}

func WithDelay(delay time.Duration) Option {
	return func(b *Base) {
		if delay.Minutes() > 0 {
			b.delay = delay.Round(time.Minute)
		}
	}
}

func WithLock(lock time.Duration) Option {
	return func(b *Base) {
		if lock > 0 {
			b.lock = lock
		}
	}
}
