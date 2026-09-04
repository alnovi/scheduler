package tests

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alnovi/scheduler/v2"
	"github.com/alnovi/scheduler/v2/tasks"
)

func TestSchedulerStartWithLocker(t *testing.T) {
	const podCount = 3

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
			taskSleep:      time.Second,
			expStartCalls:  5 * podCount,
			expCancelCalls: 1 * podCount,
			expFinishCalls: 4 * podCount,
		},
		{
			name:     "Success cron with locker minute",
			sleep:    5 * time.Minute,
			taskExpr: scheduler.CronEveryMinute,
			taskOpts: []tasks.Option{
				tasks.WithLock(time.Minute),
			},
			taskSleep:      time.Second,
			expStartCalls:  5,
			expCancelCalls: 1,
			expFinishCalls: 4,
		},
		{
			name:     "Success cron with locker hour",
			sleep:    5 * time.Minute,
			taskExpr: scheduler.CronEveryMinute,
			taskOpts: []tasks.Option{
				tasks.WithLock(time.Hour),
			},
			taskSleep:      time.Second,
			expStartCalls:  1,
			expCancelCalls: 0,
			expFinishCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				actStartCalls := 0
				actCancelCalls := 0
				actFinishCalls := 0

				logger := slog.New(slog.DiscardHandler)
				location := time.UTC
				locker := scheduler.NewLockerMemory()

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

				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()

				for range podCount {
					schedule := scheduler.New(
						scheduler.WithLogger(logger),
						scheduler.WithLocation(location),
						scheduler.WithLocker(locker),
						scheduler.WithMetrics(scheduler.WithEnabled(false)),
					)

					err := schedule.AddTask("test-task", tasks.MustCronTask(tc.taskExpr, handler, tc.taskOpts...))
					require.NoError(t, err, "failed to add task default")

					err = schedule.Start(ctx)
					require.NoError(t, err, "failed to start scheduler")
				}

				time.Sleep(tc.sleep)
				synctest.Wait()

				cancel()
				time.Sleep(time.Second)
				synctest.Wait()

				assert.Equal(t, tc.expStartCalls, actStartCalls, "unexpected number of task start calls scheduled")
				assert.Equal(t, tc.expCancelCalls, actCancelCalls, "unexpected number of task cancel calls scheduled")
				assert.Equal(t, tc.expFinishCalls, actFinishCalls, "unexpected number of task finish calls scheduled")
			})
		})
	}
}

func TestSchedulerStartWithMockLocker(t *testing.T) {
	const podCount = 3

	synctest.Test(t, func(t *testing.T) {
		require.NotPanics(t, func() {
			var schedule *scheduler.Scheduler

			locker := NewMockLocker()
			handler := func(ctx context.Context) error {
				return nil
			}

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			for range podCount {
				schedule = scheduler.New(scheduler.WithLocker(locker))

				err := schedule.AddTask("test-task", tasks.MustCronTask(scheduler.CronEveryMinute, handler, tasks.WithLock(time.Hour)))
				require.NoError(t, err, "failed to add task default")

				err = schedule.Start(ctx)
				require.NoError(t, err, "failed to start scheduler")
			}

			time.Sleep(time.Minute)
			synctest.Wait()

			err := schedule.Stop(t.Context())
			require.NoError(t, err, "failed to stop scheduler")

			time.Sleep(time.Minute)
			synctest.Wait()
		})
	})
}

type MockLocker struct{}

func NewMockLocker() *MockLocker {
	return &MockLocker{}
}

func (l *MockLocker) LockResource(_ context.Context, resource string, ttl time.Duration) (bool, string, error) {
	return false, "", errors.New("internal error locking")
}
