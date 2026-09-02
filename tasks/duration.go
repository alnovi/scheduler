package tasks

import (
	"time"

	"github.com/alnovi/scheduler/v2"
)

const durationMin = time.Minute

type DurationTask struct {
	*Base
	next     time.Time
	duration time.Duration
}

func NewDurationTask(duration time.Duration, handler HandlerFn, opts ...Option) (*DurationTask, error) {
	base, err := NewBase(handler, opts...)
	if err != nil {
		return nil, err
	}

	if duration < durationMin {
		return nil, scheduler.ErrTaskIncorrectDuration
	}

	return &DurationTask{
		Base:     base,
		next:     time.Time{},
		duration: duration,
	}, nil
}

func MustDurationTask(duration time.Duration, handler HandlerFn, opts ...Option) *DurationTask {
	task, err := NewDurationTask(duration, handler, opts...)
	if err != nil {
		panic(err)
	}
	return task
}

func (t *DurationTask) Init(now time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.next = now.Add(t.delay).Add(t.duration)
	return nil
}

func (t *DurationTask) Compare(now time.Time) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for now.After(t.next) {
		t.next = t.next.Add(t.duration)
	}

	return now.Equal(t.next), nil
}
