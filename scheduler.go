package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrTaskIsNil             = errors.New("task is nil")
	ErrTaskNotFound          = errors.New("task not found")
	ErrTaskIsRunning         = errors.New("task is running")
	ErrTaskNameIsEmpty       = errors.New("task name is empty")
	ErrTaskIncorrectDuration = errors.New("task is incorrect duration")
)

type Locker interface {
	LockResource(ctx context.Context, resource string, ttl time.Duration) (bool, string, error)
}

type Scheduler struct {
	isRun     bool
	log       *slog.Logger
	locker    Locker
	metrics   *Metrics
	location  *time.Location
	ticker    *time.Ticker
	tasks     map[string]*taskWrap
	stopCh    chan struct{}
	wg        sync.WaitGroup
	mu        sync.RWMutex
	contextFn func() context.Context
}

func New(options ...Option) *Scheduler {
	scheduler := &Scheduler{
		isRun:     false,
		log:       slog.New(slog.DiscardHandler),
		metrics:   NewMetrics(false),
		ticker:    time.NewTicker(time.Minute),
		location:  time.UTC,
		tasks:     make(map[string]*taskWrap),
		stopCh:    make(chan struct{}),
		wg:        sync.WaitGroup{},
		contextFn: context.Background,
	}

	for _, option := range options {
		option(scheduler)
	}

	return scheduler
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRun {
		return
	}

	s.isRun = true

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.stopCh:
				return
			case now := <-s.ticker.C:
				s.runTasks(now.In(s.location).Truncate(time.Minute))
			}
		}
	}()
}

func (s *Scheduler) Shutdown(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ticker.Stop()
	close(s.stopCh)
	s.wg.Wait()
	s.isRun = false
	return nil
}

func (s *Scheduler) Tasks() []*TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*TaskInfo, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, newTaskInfo(task))
	}
	return tasks
}

func (s *Scheduler) AddDurationTask(d time.Duration, t Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, err := newTaskWrap(d, t)
	if err != nil {
		return err
	}

	task.nextRun = time.Now().In(s.location).Truncate(time.Minute).Add(task.duration)

	s.tasks[task.id] = task

	return nil
}

func (s *Scheduler) AddDayAtTask(hour, minute int, t Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, _ := time.ParseDuration("24h") // nolint:gosec

	task, err := newTaskWrap(d, t)
	if err != nil {
		return err
	}

	now := time.Now()
	task.nextRun = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, s.location)

	s.tasks[task.id] = task

	return nil
}

func (s *Scheduler) StartTask(taskId string) (*TaskInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[taskId]
	if !ok {
		return nil, ErrTaskNotFound
	}
	if task.IsStopped() {
		task.SetStatus(Pending)
	}
	return newTaskInfo(task), nil
}

func (s *Scheduler) StopTask(taskId string) (*TaskInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[taskId]
	if !ok {
		return nil, ErrTaskNotFound
	}
	if task.IsRunning() {
		return nil, ErrTaskIsRunning
	}
	task.SetStatus(Stopped)
	return newTaskInfo(task), nil
}

func (s *Scheduler) LockTask(task *taskWrap) (bool, error) {
	if s.locker == nil {
		return true, nil
	}

	if task.OnceIn() == 0 {
		return true, nil
	}

	resource := fmt.Sprintf("scheduler:lock:%s", task.Name())

	ok, _, err := s.locker.LockResource(context.Background(), resource, task.OnceIn())

	return ok, err
}

func (s *Scheduler) runTasks(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, task := range s.tasks {
		if task.CompareNextRun(now) {
			s.runTask(task)
		}
	}
}

func (s *Scheduler) runTask(task *taskWrap) {
	if !task.IsPending() {
		return
	}

	isLocked, err := s.LockTask(task)
	if err != nil {
		s.log.Error("fail lock task", slog.String("task", task.Name()), slog.String("error", err.Error()))
	}

	if !isLocked {
		return
	}

	task.SetStatus(Running)

	ctx, cancel := task.ContextTimeout(s.contextFn())

	s.contextCancelWithStopCh(ctx, cancel)

	s.wg.Add(1)
	go func() {
		now := time.Now()

		s.log.Debug("started", slog.String("task", task.Name()))

		defer func() {
			s.log.Debug("finished", slog.String("task", task.Name()))
			if err := recover(); err != nil {
				s.log.Error("PANIC", slog.String("error", err.(error).Error()), slog.String("task", task.Name()))
			}
			task.SetStatus(Pending)
			cancel()
			s.wg.Done()
		}()

		if err = task.handleFn(ctx); err == nil {
			s.log.Info("Task exec ok", slog.String("task", task.Name()))
			s.metrics.TaskProcessOkInc(task.Name())
			s.metrics.TaskProcessDurationOk(task.Name(), now)
		} else {
			s.log.Error("Task exec err", slog.String("task", task.Name()), slog.String("error", err.Error()))
			s.metrics.TaskProcessErrInc(task.Name())
			s.metrics.TaskProcessDurationErr(task.Name(), now)
		}
	}()
}

func (s *Scheduler) contextCancelWithStopCh(ctx context.Context, cancel context.CancelFunc) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				cancel()
				return
			}
		}
	}()
}
