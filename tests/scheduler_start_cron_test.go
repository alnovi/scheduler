package tests

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alnovi/scheduler/v2"
	"github.com/alnovi/scheduler/v2/tasks"
)

func TestSchedulerCronStart(t *testing.T) {
	testCases := []struct {
		name           string
		sleep          time.Duration
		taskExpr       string
		taskOpts       []tasks.Option
		taskSleep      time.Duration
		expStartCalls  int
		expCancelCalls int
		expFinishCalls int
	}{
		{
			name:           "Success cron default",
			sleep:          5 * time.Minute,
			taskExpr:       scheduler.CronEveryMinute,
			taskOpts:       nil,
			taskSleep:      time.Second,
			expStartCalls:  5,
			expCancelCalls: 1,
			expFinishCalls: 4,
		},
		{
			name:     "Success cron WithEnabled",
			sleep:    5 * time.Minute,
			taskExpr: scheduler.CronEveryMinute,
			taskOpts: []tasks.Option{
				tasks.WithEnabled(false),
			},
			taskSleep:      time.Second,
			expStartCalls:  0,
			expCancelCalls: 0,
			expFinishCalls: 0,
		},
		{
			name:     "Success cron WithDelay",
			sleep:    5 * time.Minute,
			taskExpr: scheduler.CronEveryMinute,
			taskOpts: []tasks.Option{
				tasks.WithDelay(time.Minute),
			},
			taskSleep:      time.Second,
			expStartCalls:  4,
			expCancelCalls: 1,
			expFinishCalls: 3,
		},
		{
			name:     "Success cron WithTimeout 1",
			sleep:    5 * time.Minute,
			taskExpr: scheduler.CronEveryMinute,
			taskOpts: []tasks.Option{
				tasks.WithTimeout(time.Minute),
			},
			taskSleep:      time.Hour,
			expStartCalls:  5,
			expCancelCalls: 5,
			expFinishCalls: 0,
		},
		{
			name:     "Success cron WithTimeout 2",
			sleep:    5 * time.Minute,
			taskExpr: scheduler.CronEveryMinute,
			taskOpts: []tasks.Option{
				tasks.WithTimeout(2 * time.Minute),
			},
			taskSleep:      time.Hour,
			expStartCalls:  3,
			expCancelCalls: 3,
			expFinishCalls: 0,
		},
		{
			name:     "Success cron WithTimeout 3",
			sleep:    5 * time.Minute,
			taskExpr: scheduler.CronEveryMinute,
			taskOpts: []tasks.Option{
				tasks.WithTimeout(100 * time.Second),
			},
			taskSleep:      time.Hour,
			expStartCalls:  3,
			expCancelCalls: 3,
			expFinishCalls: 0,
		},
		{
			name:           "Success cron forever",
			sleep:          5 * time.Minute,
			taskExpr:       scheduler.CronEveryMinute,
			taskOpts:       nil,
			taskSleep:      time.Hour,
			expStartCalls:  1,
			expCancelCalls: 1,
			expFinishCalls: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				actStartCalls := 0
				actCancelCalls := 0
				actFinishCalls := 0

				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()

				schedule := scheduler.New()
				handler := func(ctx context.Context) error {
					later := time.Duration(0)

					actStartCalls++

					for {
						select {
						case <-ctx.Done():
							actCancelCalls++
							return ctx.Err()
						default:
							if later >= tc.taskSleep {
								actFinishCalls++
								return nil
							}
							later += time.Second
							time.Sleep(time.Second)
						}
					}
				}

				err := schedule.AddTask("test-task", tasks.MustCronTask(tc.taskExpr, handler, tc.taskOpts...))
				require.NoError(t, err, "failed to add task default")

				err = schedule.Start(ctx)
				require.NoError(t, err, "failed to start scheduler")

				time.Sleep(tc.sleep)
				synctest.Wait()

				err = schedule.Stop(t.Context())
				require.NoError(t, err, "failed to stop scheduler")

				time.Sleep(time.Minute)
				synctest.Wait()

				assert.Equal(t, tc.expStartCalls, actStartCalls, "unexpected number of task start calls scheduled")
				assert.Equal(t, tc.expCancelCalls, actCancelCalls, "unexpected number of task cancel calls scheduled")
				assert.Equal(t, tc.expFinishCalls, actFinishCalls, "unexpected number of task finish calls scheduled")
			})
		})
	}
}
