package tests

import (
	"bytes"
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

func TestSchedulerTaskError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelError}))

		handler := func(ctx context.Context) error {
			return errors.New("test error")
		}

		schedule := scheduler.New(scheduler.WithLogger(logger))
		err := schedule.AddTask("test-task", tasks.MustDurationTask(time.Minute, handler))
		require.NoError(t, err, "failed to add task default")

		err = schedule.Start(ctx)
		require.NoError(t, err, "failed to start scheduler")

		time.Sleep(2 * time.Minute)
		synctest.Wait()

		err = schedule.Stop(t.Context())
		require.NoError(t, err, "failed to stop scheduler")

		time.Sleep(time.Minute)
		synctest.Wait()

		assert.Contains(t, logBuffer.String(), "test-task")
		assert.Contains(t, logBuffer.String(), "test error")
	})
}
