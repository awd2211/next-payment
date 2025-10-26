package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// StructuredLogger 结构化日志记录器（适配ELK/Loki）
type StructuredLogger struct {
	*zap.Logger
	serviceName string
	environment string
}

// LogEntry 日志条目（JSON格式，适配ELK）
type LogEntry struct {
	Timestamp   string                 `json:"@timestamp"`
	Level       string                 `json:"level"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	TraceID     string                 `json:"trace_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	IP          string                 `json:"ip,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	Message     string                 `json:"message"`
	Error       string                 `json:"error,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// NewStructuredLogger 创建结构化日志器
func NewStructuredLogger(serviceName, environment string) (*StructuredLogger, error) {
	// ELK/Loki友好的配置
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "@timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stack"

	// JSON编码
	config.Encoding = "json"

	// 输出到stdout（由Filebeat/Promtail收集）
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", serviceName),
			zap.String("environment", environment),
		),
	)
	if err != nil {
		return nil, err
	}

	return &StructuredLogger{
		Logger:      logger,
		serviceName: serviceName,
		environment: environment,
	}, nil
}

// LoggingMiddleware Gin中间件 - 记录所有HTTP请求
func (l *StructuredLogger) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start)

		// 提取元数据
		entry := &LogEntry{
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Level:       getLogLevel(c.Writer.Status()),
			Service:     l.serviceName,
			Environment: l.environment,
			TraceID:     c.GetString("trace_id"),
			UserID:      c.GetString("user_id"),
			IP:          c.ClientIP(),
			Method:      c.Request.Method,
			Path:        path,
			StatusCode:  c.Writer.Status(),
			Duration:    duration.Milliseconds(),
			Message:     fmt.Sprintf("%s %s", c.Request.Method, path),
			Fields: map[string]interface{}{
				"query":       query,
				"user_agent":  c.Request.UserAgent(),
				"request_id":  c.GetString("request_id"),
				"bytes_sent":  c.Writer.Size(),
				"protocol":    c.Request.Proto,
				"remote_addr": c.Request.RemoteAddr,
			},
		}

		// 记录错误
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// 输出日志
		l.logEntry(entry)
	}
}

// logEntry 输出日志条目
func (l *StructuredLogger) logEntry(entry *LogEntry) {
	// 转换为zap字段
	fields := []zap.Field{
		zap.String("trace_id", entry.TraceID),
		zap.String("user_id", entry.UserID),
		zap.String("ip", entry.IP),
		zap.String("method", entry.Method),
		zap.String("path", entry.Path),
		zap.Int("status_code", entry.StatusCode),
		zap.Int64("duration_ms", entry.Duration),
	}

	// 添加自定义字段
	for k, v := range entry.Fields {
		fields = append(fields, zap.Any(k, v))
	}

	// 根据级别输出
	switch entry.Level {
	case "error":
		if entry.Error != "" {
			fields = append(fields, zap.String("error", entry.Error))
		}
		l.Error(entry.Message, fields...)
	case "warn":
		l.Warn(entry.Message, fields...)
	case "debug":
		l.Debug(entry.Message, fields...)
	default:
		l.Info(entry.Message, fields...)
	}
}

// getLogLevel 根据HTTP状态码确定日志级别
func getLogLevel(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "error"
	case statusCode >= 400:
		return "warn"
	case statusCode >= 300:
		return "info"
	default:
		return "info"
	}
}

// LogSecurityEvent 记录安全事件（高优先级）
func (l *StructuredLogger) LogSecurityEvent(ctx context.Context, event, description string, fields map[string]interface{}) {
	entry := &LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Level:       "warn",
		Service:     l.serviceName,
		Environment: l.environment,
		Message:     fmt.Sprintf("SECURITY_EVENT: %s", event),
		Fields: map[string]interface{}{
			"event_type":  "security",
			"event_name":  event,
			"description": description,
		},
	}

	// 合并额外字段
	for k, v := range fields {
		entry.Fields[k] = v
	}

	l.logEntry(entry)
}

// LogAuditEvent 记录审计事件
func (l *StructuredLogger) LogAuditEvent(ctx context.Context, adminID, action, resource, resourceID string, fields map[string]interface{}) {
	entry := &LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Level:       "info",
		Service:     l.serviceName,
		Environment: l.environment,
		UserID:      adminID,
		Message:     fmt.Sprintf("AUDIT: %s %s", action, resource),
		Fields: map[string]interface{}{
			"event_type":  "audit",
			"action":      action,
			"resource":    resource,
			"resource_id": resourceID,
		},
	}

	for k, v := range fields {
		entry.Fields[k] = v
	}

	l.logEntry(entry)
}

// FlushToLoki 发送日志到Loki（可选）
func (l *StructuredLogger) FlushToLoki(lokiURL string, entries []*LogEntry) error {
	// Loki Push API格式
	type lokiStream struct {
		Stream map[string]string `json:"stream"`
		Values [][]string        `json:"values"`
	}

	type lokiRequest struct {
		Streams []lokiStream `json:"streams"`
	}

	// 按标签分组
	streamsByLabels := make(map[string][][]string)
	for _, entry := range entries {
		labels := fmt.Sprintf(`{service="%s",level="%s",environment="%s"}`,
			entry.Service, entry.Level, entry.Environment)

		// 转换为Loki格式 [timestamp_ns, line]
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
		line, _ := json.Marshal(entry)

		streamsByLabels[labels] = append(streamsByLabels[labels], []string{
			timestamp,
			string(line),
		})
	}

	// 构建请求
	var streams []lokiStream
	for labels, values := range streamsByLabels {
		var labelMap map[string]string
		_ = json.Unmarshal([]byte(labels), &labelMap)

		streams = append(streams, lokiStream{
			Stream: labelMap,
			Values: values,
		})
	}

	req := &lokiRequest{Streams: streams}
	_, _ = json.Marshal(req)

	// TODO: 实际发送HTTP请求到Loki
	// resp, err := http.Post(lokiURL+"/loki/api/v1/push", "application/json", bytes.NewBuffer(data))

	return nil
}

// ElasticsearchIndex Elasticsearch索引名称生成
func (l *StructuredLogger) ElasticsearchIndex() string {
	// 按日期分割索引（便于管理）
	date := time.Now().Format("2006.01.02")
	return fmt.Sprintf("admin-bff-%s-%s", l.environment, date)
}

// ShouldSample 采样决策（减少日志量）
func (l *StructuredLogger) ShouldSample(path string, statusCode int) bool {
	// 错误请求始终记录
	if statusCode >= 400 {
		return true
	}

	// 健康检查仅1%采样
	if path == "/health" || path == "/ping" {
		return time.Now().UnixNano()%100 == 0
	}

	// 正常请求100%记录
	return true
}
