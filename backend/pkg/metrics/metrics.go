package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PaymentMetrics 支付相关的 Prometheus 指标
type PaymentMetrics struct {
	// 支付总数计数器
	PaymentTotal *prometheus.CounterVec
	// 支付金额直方图
	PaymentAmount *prometheus.HistogramVec
	// 支付处理时长直方图
	PaymentDuration *prometheus.HistogramVec
	// 退款总数计数器
	RefundTotal *prometheus.CounterVec
	// 退款金额直方图
	RefundAmount *prometheus.HistogramVec
}

// HTTPMetrics HTTP 请求相关的 Prometheus 指标
type HTTPMetrics struct {
	// 请求总数计数器
	RequestsTotal *prometheus.CounterVec
	// 请求时长直方图
	RequestDuration *prometheus.HistogramVec
	// 请求体大小直方图
	RequestSize *prometheus.HistogramVec
	// 响应体大小直方图
	ResponseSize *prometheus.HistogramVec
}

// DBMetrics 数据库操作相关的 Prometheus 指标
type DBMetrics struct {
	// 查询总数计数器
	QueryTotal *prometheus.CounterVec
	// 查询时长直方图
	QueryDuration *prometheus.HistogramVec
}

// NewPaymentMetrics 创建支付指标
func NewPaymentMetrics(namespace string) *PaymentMetrics {
	return &PaymentMetrics{
		PaymentTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "payment_total",
				Help:      "Total number of payments",
			},
			[]string{"status", "channel", "currency"},
		),
		PaymentAmount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "payment_amount",
				Help:      "Payment amount distribution",
				Buckets:   []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
			},
			[]string{"currency", "channel"},
		),
		PaymentDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "payment_duration_seconds",
				Help:      "Payment processing duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation", "status"},
		),
		RefundTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "refund_total",
				Help:      "Total number of refunds",
			},
			[]string{"status", "currency"},
		),
		RefundAmount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "refund_amount",
				Help:      "Refund amount distribution",
				Buckets:   []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
			},
			[]string{"currency"},
		),
	}
}

// NewHTTPMetrics 创建 HTTP 指标
func NewHTTPMetrics(namespace string) *HTTPMetrics {
	return &HTTPMetrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		RequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
		ResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
	}
}

// NewDBMetrics 创建数据库指标
func NewDBMetrics(namespace string) *DBMetrics {
	return &DBMetrics{
		QueryTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_query_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		QueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
	}
}

// RecordPayment 记录支付指标
func (m *PaymentMetrics) RecordPayment(status, channel, currency string, amount float64, duration time.Duration) {
	m.PaymentTotal.WithLabelValues(status, channel, currency).Inc()
	m.PaymentAmount.WithLabelValues(currency, channel).Observe(amount)
	m.PaymentDuration.WithLabelValues("create_payment", status).Observe(duration.Seconds())
}

// RecordRefund 记录退款指标
func (m *PaymentMetrics) RecordRefund(status, currency string, amount float64) {
	m.RefundTotal.WithLabelValues(status, currency).Inc()
	m.RefundAmount.WithLabelValues(currency).Observe(amount)
}

// RecordHTTPRequest 记录 HTTP 请求指标
func (m *HTTPMetrics) RecordHTTPRequest(method, path, status string, duration time.Duration, reqSize, respSize int) {
	m.RequestsTotal.WithLabelValues(method, path, status).Inc()
	m.RequestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
	if reqSize > 0 {
		m.RequestSize.WithLabelValues(method, path).Observe(float64(reqSize))
	}
	if respSize > 0 {
		m.ResponseSize.WithLabelValues(method, path).Observe(float64(respSize))
	}
}

// RecordDBQuery 记录数据库查询指标
func (m *DBMetrics) RecordDBQuery(operation, table, status string, duration time.Duration) {
	m.QueryTotal.WithLabelValues(operation, table, status).Inc()
	m.QueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}
