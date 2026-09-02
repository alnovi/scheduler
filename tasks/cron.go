package tasks

import (
	"fmt"
	"time"

	"github.com/alnovi/gron"

	"github.com/alnovi/scheduler/v2"
)

type CronTask struct {
	*Base
	next       time.Time
	expression string
}

func NewCronTask(expression string, handler HandlerFn, opts ...Option) (*CronTask, error) {
	base, err := NewBase(handler, opts...)
	if err != nil {
		return nil, err
	}

	_, err = gron.NextAfter(time.Now(), expression)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", scheduler.ErrTaskCronExpression, err)
	}

	return &CronTask{
		Base:       base,
		next:       time.Time{},
		expression: expression,
	}, nil
}

func MustCronTask(expression string, handler HandlerFn, opts ...Option) *CronTask {
	task, err := NewCronTask(expression, handler, opts...)
	if err != nil {
		panic(err)
	}
	return task
}

func (t *CronTask) Init(now time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var err error
	if t.next, err = gron.NextAfter(now.Add(t.delay), t.expression); err != nil {
		return err
	}

	return nil
}

func (t *CronTask) Compare(now time.Time) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var err error

	for now.After(t.next) {
		t.next, err = gron.NextAfter(t.next, t.expression)
		if err != nil {
			return false, err
		}
	}

	return now.Equal(t.next), nil
}
