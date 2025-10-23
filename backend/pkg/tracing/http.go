package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// HTTPClient 包装了标准 http.Client，自动添加追踪
type HTTPClient struct {
	client     *http.Client
	tracerName string
}

// NewHTTPClient 创建带追踪的 HTTP 客户端
func NewHTTPClient(client *http.Client, tracerName string) *HTTPClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPClient{
		client:     client,
		tracerName: tracerName,
	}
}

// Do 执行 HTTP 请求并自动添加追踪
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// 开始 span
	tracer := otel.Tracer(c.tracerName)
	spanName := req.Method + " " + req.URL.Host + req.URL.Path

	ctx, span := tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.HTTPMethod(req.Method),
			semconv.HTTPTarget(req.URL.Path),
			semconv.HTTPScheme(req.URL.Scheme),
			attribute.String("http.host", req.URL.Host),
			semconv.HTTPURL(req.URL.String()),
		),
	)
	defer span.End()

	// 将 trace context 注入到请求头
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// 更新请求的 context
	req = req.WithContext(ctx)

	// 执行请求
	resp, err := c.client.Do(req)

	// 记录响应
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return resp, err
	}

	// 记录响应状态
	span.SetAttributes(semconv.HTTPStatusCode(resp.StatusCode))

	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, "HTTP error")
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return resp, nil
}

// InjectTraceContext 将 trace context 注入到 HTTP 请求头
func InjectTraceContext(ctx context.Context, req *http.Request) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// ExtractTraceContext 从 HTTP 请求头提取 trace context
func ExtractTraceContext(req *http.Request) context.Context {
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))
}

// AddHTTPClientSpan 为 HTTP 客户端调用添加 span
func AddHTTPClientSpan(ctx context.Context, tracerName string, method string, url string) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	spanName := method + " " + url

	return tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.HTTPMethod(method),
			semconv.HTTPURL(url),
		),
	)
}

// SetHTTPClientSpanStatus 设置 HTTP 客户端 span 状态
func SetHTTPClientSpanStatus(span trace.Span, statusCode int, err error) {
	span.SetAttributes(attribute.Int("http.status_code", statusCode))

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else if statusCode >= 400 {
		span.SetStatus(codes.Error, "HTTP error")
	} else {
		span.SetStatus(codes.Ok, "")
	}
}
