package scheduler

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	subsystem = "scheduler"
	statusOk  = "ok"
	statusErr = "error"
)

type Metrics struct {
	enabled             bool
	namespace           string
	register            prometheus.Registerer
	taskProcessCount    *prometheus.CounterVec
	taskProcessDuration *prometheus.HistogramVec
}

func NewMetrics(enabled bool, opts ...MetricsOption) *Metrics {
	m := &Metrics{
		enabled:   enabled,
		namespace: "",
		register:  prometheus.DefaultRegisterer,
	}

	for _, opt := range opts {
		opt(m)
	}

	if !m.enabled {
		return m
	}

	m.taskProcessCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: m.namespace,
		Subsystem: subsystem,
		Name:      "task_exec_count",
		Help:      "Number of task processed",
	}, []string{"status", "task"})
	m.register.MustRegister(m.taskProcessCount)

	m.taskProcessDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: m.namespace,
		Subsystem: subsystem,
		Name:      "task_exec_duration_minutes",
		Help:      "Duration of time to task executed",
		Buckets:   []float64{1, 2, 3, 5, 10, 15, 30},
	}, []string{"status", "task"})
	m.register.MustRegister(m.taskProcessDuration)

	return m
}

func (m *Metrics) TaskProcessOkInc(task string) {
	if m.enabled {
		m.taskProcessCount.With(prometheus.Labels{"status": statusOk, "task": task}).Inc()
	}
}

func (m *Metrics) TaskProcessErrInc(task string) {
	if m.enabled {
		m.taskProcessCount.With(prometheus.Labels{"status": statusErr, "task": task}).Inc()
	}
}

func (m *Metrics) TaskProcessDurationOk(task string, start time.Time) {
	if m.enabled {
		labels := prometheus.Labels{"status": statusOk, "task": task}
		m.taskProcessDuration.With(labels).Observe(time.Since(start).Minutes())
	}
}

func (m *Metrics) TaskProcessDurationErr(task string, start time.Time) {
	if m.enabled {
		labels := prometheus.Labels{"status": statusErr, "task": task}
		m.taskProcessDuration.With(labels).Observe(time.Since(start).Minutes())
	}
}

type MetricsOption func(*Metrics)

func WithEnabled(enabled bool) MetricsOption {
	return func(m *Metrics) {
		m.enabled = enabled
	}
}

func WithNamespace(namespace string) MetricsOption {
	return func(m *Metrics) {
		m.namespace = namespace
	}
}

func WithRegister(register prometheus.Registerer) MetricsOption {
	return func(m *Metrics) {
		if register != nil {
			m.register = register
		}
	}
}
