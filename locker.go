package scheduler

import (
	"context"
	"sync"
	"time"
)

type LockerMemory struct {
	state map[string]time.Time
	mu    sync.RWMutex
}

func NewLockerMemory() *LockerMemory {
	return &LockerMemory{state: make(map[string]time.Time)}
}

func (l *LockerMemory) LockResource(_ context.Context, resource string, ttl time.Duration) (bool, string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if deadline, ok := l.state[resource]; ok {
		if time.Now().Before(deadline) {
			return false, "", nil
		}
	}

	l.state[resource] = time.Now().Add(ttl)
	return true, resource, nil
}
