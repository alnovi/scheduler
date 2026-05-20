package scheduler

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	Pending = "pending"
	Running = "running"
	Stopped = "stopped"
)

type Task interface {
	Name() string
	Handle(ctx context.Context) error
}

type TaskContext interface {
	Context(ctx context.Context) context.Context
}

type TaskTimeout interface {
	Timeout() time.Duration
}

type TaskLocker interface {
	OnceIn() time.Duration
}

type TaskInfo struct {
	Id     string
	Name   string
	Status string
	RunAt  time.Time
}

func newTaskInfo(t *taskWrap) *TaskInfo {
	return &TaskInfo{
		Id:     t.id,
		Name:   t.name,
		Status: t.status,
		RunAt:  t.nextRun,
	}
}

type taskWrap struct {
	id        string
	name      string
	status    string
	duration  time.Duration
	onceIn    time.Duration
	nextRun   time.Time
	handleFn  func(context.Context) error
	contextFn func(context.Context) context.Context
	timeoutFn func() time.Duration
	mu        sync.RWMutex
}

func newTaskWrap(d time.Duration, t Task) (*taskWrap, error) {
	if t == nil {
		return nil, ErrTaskIsNil
	}

	if strings.TrimSpace(t.Name()) == "" {
		return nil, ErrTaskNameIsEmpty
	}

	if d.Minutes() < 1 {
		return nil, ErrTaskIncorrectDuration
	}

	task := &taskWrap{
		id:       uuid.NewString(),
		name:     t.Name(),
		status:   Pending,
		duration: d,
		handleFn: t.Handle,
		contextFn: func(ctx context.Context) context.Context {
			return ctx
		},
		timeoutFn: func() time.Duration {
			return 0
		},
	}

	if opt, ok := t.(TaskContext); ok {
		task.contextFn = opt.Context
	}

	if opt, ok := t.(TaskTimeout); ok {
		if opt.Timeout() != 0 {
			task.timeoutFn = opt.Timeout
		}
	}

	if opt, ok := t.(TaskLocker); ok {
		if opt.OnceIn() > 0 {
			task.onceIn = opt.OnceIn()
		}
	}

	return task, nil
}

func (t *taskWrap) Name() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.name
}

func (t *taskWrap) OnceIn() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.onceIn
}

func (t *taskWrap) IsPending() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status == Pending
}

func (t *taskWrap) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status == Running
}

func (t *taskWrap) IsStopped() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status == Stopped
}

func (t *taskWrap) SetStatus(status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = status
}

func (t *taskWrap) NextRun() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nextRun
}

func (t *taskWrap) CompareNextRun(now time.Time) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for now.After(t.nextRun) {
		t.nextRun = t.nextRun.Add(t.duration)
	}

	return now.Equal(t.nextRun)
}

func (t *taskWrap) SetNextRun(val time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextRun = val
}

func (t *taskWrap) ContextTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	ctx = t.contextFn(ctx)
	if t.timeoutFn() > 0 {
		ctx, cancel = context.WithTimeout(ctx, t.timeoutFn())
	} else {
		ctx, cancel = context.WithTimeout(ctx, t.duration)
	}
	return ctx, cancel
}
