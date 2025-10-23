package tracing

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware 返回 Gin 中间件，用于自动追踪 HTTP 请求
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		// 从请求头中提取 trace context
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// 构建 span 名称（使用路由模式而非实际路径）
		spanName := c.Request.Method + " " + c.FullPath()
		if c.FullPath() == "" {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		// 开始 span
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPRoute(c.FullPath()),
				semconv.HTTPTarget(c.Request.URL.Path),
				semconv.HTTPScheme(c.Request.URL.Scheme),
				attribute.String("http.host", c.Request.Host),
				semconv.HTTPUserAgent(c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		// 将 context 存入 Gin context
		c.Request = c.Request.WithContext(ctx)

		// 将 trace ID 添加到响应头（便于调试）
		c.Header("X-Trace-ID", span.SpanContext().TraceID().String())

		// 处理请求
		c.Next()

		// 记录响应状态
		status := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCode(status),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// 如果是错误响应，标记 span 为错误
		if status >= 400 {
			span.SetStatus(codes.Error, c.Errors.String())
			if len(c.Errors) > 0 {
				span.RecordError(c.Errors[0])
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}
