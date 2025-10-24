package saga

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SagaMetrics Saga 监控指标
type SagaMetrics struct {
	sagaTotal           *prometheus.CounterVec
	sagaDuration        *prometheus.HistogramVec
	sagaStepTotal       *prometheus.CounterVec
	sagaStepDuration    *prometheus.HistogramVec
	compensationTotal   *prometheus.CounterVec
	compensationRetries *prometheus.HistogramVec
	sagaInProgress      *prometheus.GaugeVec
	dlqSize             prometheus.Gauge
}

// NewSagaMetrics 创建 Saga 监控指标
func NewSagaMetrics(namespace string) *SagaMetrics {
	if namespace == "" {
		namespace = "payment_platform"
	}

	return &SagaMetrics{
		// Saga 执行总数（按状态和业务类型分类）
		sagaTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "saga",
				Name:      "total",
				Help:      "Total number of saga executions by status and business type",
			},
			[]string{"status", "business_type"},
		),

		// Saga 执行时长（按业务类型分类）
		sagaDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "saga",
				Name:      "duration_seconds",
				Help:      "Duration of saga execution in seconds",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
			},
			[]string{"business_type", "status"},
		),

		// Saga 步骤执行总数（按状态和步骤名称分类）
		sagaStepTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "saga",
				Name:      "step_total",
				Help:      "Total number of saga step executions by status and step name",
			},
			[]string{"step_name", "status"},
		),

		// Saga 步骤执行时长
		sagaStepDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "saga",
				Name:      "step_duration_seconds",
				Help:      "Duration of saga step execution in seconds",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"step_name", "status"},
		),

		// 补偿执行总数（按成功/失败分类）
		compensationTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "saga",
				Name:      "compensation_total",
				Help:      "Total number of compensation executions",
			},
			[]string{"step_name", "status"},
		),

		// 补偿重试次数分布
		compensationRetries: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "saga",
				Name:      "compensation_retries",
				Help:      "Number of compensation retry attempts",
				Buckets:   []float64{0, 1, 2, 3, 4, 5},
			},
			[]string{"step_name"},
		),

		// 当前正在执行的 Saga 数量
		sagaInProgress: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "saga",
				Name:      "in_progress",
				Help:      "Number of sagas currently in progress",
			},
			[]string{"business_type"},
		),

		// 死信队列大小
		dlqSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "saga",
				Name:      "dlq_size",
				Help:      "Number of sagas in dead letter queue",
			},
		),
	}
}

// RecordSagaStart 记录 Saga 开始
func (m *SagaMetrics) RecordSagaStart(businessType string) {
	m.sagaInProgress.WithLabelValues(businessType).Inc()
}

// RecordSagaComplete 记录 Saga 完成
func (m *SagaMetrics) RecordSagaComplete(businessType, status string, duration time.Duration) {
	m.sagaTotal.WithLabelValues(status, businessType).Inc()
	m.sagaDuration.WithLabelValues(businessType, status).Observe(duration.Seconds())
	m.sagaInProgress.WithLabelValues(businessType).Dec()
}

// RecordStepExecution 记录步骤执行
func (m *SagaMetrics) RecordStepExecution(stepName, status string, duration time.Duration) {
	m.sagaStepTotal.WithLabelValues(stepName, status).Inc()
	m.sagaStepDuration.WithLabelValues(stepName, status).Observe(duration.Seconds())
}

// RecordCompensation 记录补偿执行
func (m *SagaMetrics) RecordCompensation(stepName, status string, retries int) {
	m.compensationTotal.WithLabelValues(stepName, status).Inc()
	m.compensationRetries.WithLabelValues(stepName).Observe(float64(retries))
}

// SetDLQSize 设置死信队列大小
func (m *SagaMetrics) SetDLQSize(size int) {
	m.dlqSize.Set(float64(size))
}
