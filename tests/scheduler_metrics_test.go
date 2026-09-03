package tests

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/alnovi/scheduler/v2"
)

func TestSchedulerMetrics(t *testing.T) {
	metrics := scheduler.NewMetrics(true,
		scheduler.WithEnabled(true),
		scheduler.WithNamespace("testing"),
		scheduler.WithRegister(prometheus.DefaultRegisterer),
	)

	metrics.TaskProcessOkInc("task")
	metrics.TaskProcessDurationOk("task", time.Now().Add(time.Minute))

	metrics.TaskProcessErrInc("task")
	metrics.TaskProcessDurationErr("task", time.Now().Add(time.Minute))
}
