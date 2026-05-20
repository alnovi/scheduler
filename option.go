package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type Option func(s *Scheduler)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Scheduler) {
		if logger != nil {
			s.log = logger
		}
	}
}

func WithLocker(locker Locker) Option {
	return func(s *Scheduler) {
		s.locker = locker
	}
}

func WithMetrics(opts ...MetricsOption) Option {
	return func(s *Scheduler) {
		s.metrics = NewMetrics(true, opts...)
	}
}

func WithLocation(location *time.Location) Option {
	return func(s *Scheduler) {
		if location != nil {
			s.location = location
		}
	}
}

func WithContextFn(fn func() context.Context) Option {
	return func(s *Scheduler) {
		if fn != nil && fn() != nil {
			s.contextFn = fn
		}
	}
}
