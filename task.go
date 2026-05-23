package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alnovi/gron"
	"github.com/google/uuid"
)

const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusStopped = "stopped"

	ClassDuration = "duration"
	ClassCron     = "cron"
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
	class     string
	name      string
	status    string
	cron      string
	duration  time.Duration
	onceIn    time.Duration
	nextRun   time.Time
	handleFn  func(context.Context) error
	contextFn func(context.Context) context.Context
	timeoutFn func() time.Duration
	mu        sync.RWMutex
}

func newTaskWrapDuration(d time.Duration, t Task) (*taskWrap, error) {
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
		class:    ClassDuration,
		name:     t.Name(),
		status:   StatusPending,
		cron:     "",
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

func newTaskWrapCron(expression string, t Task) (*taskWrap, error) {
	if t == nil {
		return nil, ErrTaskIsNil
	}

	if strings.TrimSpace(t.Name()) == "" {
		return nil, ErrTaskNameIsEmpty
	}

	if _, err := gron.NextTime(expression); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTaskCronExpression, err)
	}

	task := &taskWrap{
		id:       uuid.NewString(),
		class:    ClassCron,
		name:     t.Name(),
		status:   StatusPending,
		cron:     expression,
		duration: 0,
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
	return t.status == StatusPending
}

func (t *taskWrap) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status == StatusRunning
}

func (t *taskWrap) IsStopped() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status == StatusStopped
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

func (t *taskWrap) CompareNextRun(now time.Time) (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for now.After(t.nextRun) {
		switch t.class {
		case ClassCron:
			next, err := gron.NextTime(t.cron)
			if err != nil {
				return false, err
			}
			t.nextRun = next
		case ClassDuration:
			t.nextRun = t.nextRun.Add(t.duration)
		}
	}

	return now.Equal(t.nextRun), nil
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
