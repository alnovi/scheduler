package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alnovi/scheduler/v2"
	"github.com/alnovi/scheduler/v2/tasks"
)

func TestTaskCreateCron(t *testing.T) {
	stubHand := func(_ context.Context) error {
		return nil
	}

	testCases := []struct {
		name     string
		taskExpr string
		taskHand func(ctx context.Context) error
		taskOpts []tasks.Option
		expErr   string
	}{
		{
			name:     "Success",
			taskExpr: scheduler.CronEveryMinute,
			taskHand: stubHand,
			taskOpts: nil,
			expErr:   ``,
		},
		{
			name:     "Invalid handler",
			taskExpr: scheduler.CronEveryMinute,
			taskHand: nil,
			taskOpts: nil,
			expErr:   `task handler is empty`,
		},
		{
			name:     "Invalid expression",
			taskExpr: "* * * *",
			taskHand: stubHand,
			taskOpts: nil,
			expErr:   `task cron expression is invalid`,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("New %s", tc.name), func(t *testing.T) {
			require.NotPanics(t, func() {
				_, err := tasks.NewCronTask(tc.taskExpr, tc.taskHand, tc.taskOpts...)
				if tc.expErr == "" {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					require.ErrorContains(t, err, tc.expErr)
				}
			})
		})

		t.Run(fmt.Sprintf("Must %s", tc.name), func(t *testing.T) {
			defer func() {
				err, _ := recover().(error)
				if tc.expErr == "" {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					require.ErrorContains(t, err, tc.expErr)
				}
			}()

			_ = tasks.MustCronTask(tc.taskExpr, tc.taskHand, tc.taskOpts...)
		})
	}
}
