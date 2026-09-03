package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alnovi/scheduler/v2"
	"github.com/alnovi/scheduler/v2/tasks"
)

func TestAddTask(t *testing.T) {
	handler := func(_ context.Context) error { return nil }

	testCases := []struct {
		test   string
		name   string
		task   scheduler.Task
		expErr string
	}{
		{
			test:   "success",
			name:   "task",
			task:   tasks.MustCronTask(scheduler.CronEveryMinute, handler),
			expErr: ``,
		},
		{
			test:   "fail name",
			name:   " ",
			task:   tasks.MustDurationTask(time.Minute, handler),
			expErr: `task name is empty`,
		},
		{
			test:   "fail task",
			name:   "task",
			task:   nil,
			expErr: `task is nil`,
		},
	}

	schedule := scheduler.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := schedule.AddTask(tc.name, tc.task)
			if tc.expErr == "" {
				require.NoError(t, err, "not equal error")
			} else {
				require.Error(t, err, "not equal error")
				require.ErrorContains(t, err, tc.expErr, "not contains error message")
			}
		})
	}
}
