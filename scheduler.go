package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
)

var (
	ErrTaskIsNil             = errors.New("task is nil")
	ErrTaskNameIsEmpty       = errors.New("task name is empty")
	ErrTaskHandlerIsEmpty    = errors.New("task handler is empty")
	ErrTaskIncorrectDuration = errors.New("task is incorrect duration")
	ErrTaskCronExpression    = errors.New("task cron expression is invalid")
)

type Status string

type Task interface {
	Init(now time.Time) error
	GetStatus() Status
	Timeout() time.Duration
	Lock() time.Duration
	Compare(now time.Time) (bool, error)
	Handle(ctx context.Context) error
}

type Locker interface {
	LockResource(ctx context.Context, resource string, ttl time.Duration) (bool, string, error)
}

type Scheduler struct {
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *slog.Logger
	location *time.Location
	ticker   *time.Ticker
	locker   Locker
	metrics  *Metrics
	tasks    map[string]Task
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

func New(opts ...Option) *Scheduler {
	s := &Scheduler{
		ctx:      context.Background(),
		cancel:   nil,
		logger:   slog.New(slog.DiscardHandler),
		location: time.UTC,
		ticker:   nil,
		locker:   nil,
		metrics:  NewMetrics(false),
		tasks:    make(map[string]Task),
		wg:       sync.WaitGroup{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}

	return s
}

func (s *Scheduler) AddTask(name string, task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = strings.TrimSpace(name)
	if name == "" {
		return ErrTaskNameIsEmpty
	}

	if task == nil {
		return ErrTaskIsNil
	}

	s.tasks[name] = task

	return nil
}

func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRun() {
		return nil
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.ticker = time.NewTicker(time.Minute)

	for _, task := range s.tasks {
		if err := task.Init(time.Now().In(s.location).Truncate(time.Minute)); err != nil {
			return err
		}
	}

	s.wg.Go(func() {
		defer func() {
			_ = s.shutdown(context.Background()) // nolint:gosec
		}()

		for {
			select {
			case <-s.ctx.Done():
				return
			case now := <-s.ticker.C:
				s.runTasks(now.In(s.location).Truncate(time.Minute))
			}
		}
	})

	s.logger.Debug("scheduler started")

	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shutdown(ctx)
}

func (s *Scheduler) isRun() bool {
	return s.cancel != nil
}

func (s *Scheduler) runTasks(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, task := range s.tasks {
		if task.GetStatus() != StatusPending {
			continue
		}

		isLocked, err := s.lockTask(name, task)
		if err != nil {
			s.logger.Error("failed to lock task", slog.String("task", name), slog.Any("err", err))
			continue
		}

		if !isLocked {
			continue
		}

		canRun, err := task.Compare(now)
		if err != nil {
			s.logger.Error("fail to compare next run", slog.String("task", name), slog.String("error", err.Error()))
			continue
		}

		if canRun {
			s.runTask(name, task)
		}
	}
}

func (s *Scheduler) runTask(name string, task Task) {
	s.wg.Go(func() {
		s.logger.Debug("task started", slog.String("task", name))

		ctx, cancel := s.taskContext(task)
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("task panic", slog.String("task", name), slog.Any("error", err))
			}
			cancel()
		}()

		started := time.Now()

		err := task.Handle(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				s.logger.Info("task deadline", slog.String("task", name))
				s.metrics.TaskProcessOkInc(name)
				s.metrics.TaskProcessDurationOk(name, started)
			case errors.Is(err, context.Canceled):
				s.logger.Warn("task cancelled", slog.String("task", name))
				s.metrics.TaskProcessOkInc(name)
				s.metrics.TaskProcessDurationOk(name, started)
			default:
				s.logger.Error("task failed", slog.String("task", name), slog.Any("error", err))
				s.metrics.TaskProcessErrInc(name)
				s.metrics.TaskProcessDurationErr(name, started)
			}
			return
		}

		s.logger.Info("task success", slog.String("task", name))
		s.metrics.TaskProcessOkInc(name)
		s.metrics.TaskProcessDurationOk(name, started)
	})
}

func (s *Scheduler) lockTask(name string, task Task) (bool, error) {
	if s.locker == nil {
		return true, nil
	}

	if task.Lock() <= 0 {
		return true, nil
	}

	resource := fmt.Sprintf("scheduler:lock:%s", name)

	ok, _, err := s.locker.LockResource(context.Background(), resource, task.Lock())

	return ok, err
}

func (s *Scheduler) taskContext(task Task) (context.Context, context.CancelFunc) {
	if task.Timeout() > 0 {
		return context.WithTimeout(s.ctx, task.Timeout())
	}
	return context.WithCancel(s.ctx)
}

func (s *Scheduler) shutdown(ctx context.Context) error {
	if !s.isRun() {
		return nil
	}

	s.cancel()
	s.cancel = nil
	s.ticker.Stop()

	defer func() {
		s.logger.Debug("scheduler stopped")
	}()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return errors.New("shutdown timeout: some tasks did not finish")
	}
}
