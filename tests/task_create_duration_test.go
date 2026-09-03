package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alnovi/scheduler/v2/tasks"
)

func TestTaskCreateDuration(t *testing.T) {
	stubHand := func(_ context.Context) error {
		return nil
	}

	testCases := []struct {
		name         string
		taskDuration time.Duration
		taskHandler  func(ctx context.Context) error
		taskOptions  []tasks.Option
		expErr       string
	}{
		{
			name:         "Success",
			taskDuration: time.Minute,
			taskHandler:  stubHand,
			taskOptions:  []tasks.Option{nil},
			expErr:       ``,
		},
		{
			name:         "Invalid handler",
			taskDuration: time.Minute,
			taskHandler:  nil,
			taskOptions:  nil,
			expErr:       `task handler is empty`,
		},
		{
			name:         "Invalid expression",
			taskDuration: 0,
			taskHandler:  stubHand,
			taskOptions:  nil,
			expErr:       `task is incorrect duration`,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("New %s", tc.name), func(t *testing.T) {
			require.NotPanics(t, func() {
				_, err := tasks.NewDurationTask(tc.taskDuration, tc.taskHandler, tc.taskOptions...)
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

			_ = tasks.MustDurationTask(tc.taskDuration, tc.taskHandler, tc.taskOptions...)
		})
	}
}
