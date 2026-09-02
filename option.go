package scheduler

import (
	"log/slog"
	"time"
)

type Option func(s *Scheduler)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Scheduler) {
		if logger != nil {
			s.logger = logger
		}
	}
}

func WithLocation(location *time.Location) Option {
	return func(s *Scheduler) {
		if location != nil {
			s.location = location
		}
	}
}

func WithLocker(locker Locker) Option {
	return func(s *Scheduler) {
		if locker != nil {
			s.locker = locker
		}
	}
}

func WithMetrics(opts ...MetricsOption) Option {
	return func(s *Scheduler) {
		s.metrics = NewMetrics(true, opts...)
	}
}
