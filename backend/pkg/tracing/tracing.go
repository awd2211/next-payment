package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider 全局 tracer provider
var tracerProvider *sdktrace.TracerProvider

// Config Jaeger 追踪配置
type Config struct {
	ServiceName    string  // 服务名称
	ServiceVersion string  // 服务版本
	Environment    string  // 环境（development/production）
	JaegerEndpoint string  // Jaeger collector 端点
	SamplingRate   float64 // 采样率 (0.0-1.0)
}

// InitTracer 初始化 Jaeger tracer
func InitTracer(config Config) (func(context.Context) error, error) {
	// 创建 Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
	if err != nil {
		return nil, fmt.Errorf("创建 Jaeger exporter 失败: %w", err)
	}

	// 创建资源（服务标识）
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	// 设置采样率
	samplingRate := config.SamplingRate
	if samplingRate <= 0 {
		samplingRate = 1.0 // 默认全采样
	}

	// 创建 TracerProvider
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(samplingRate)),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tracerProvider)

	// 设置全局 Propagator（用于跨服务传播 trace context）
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 返回 shutdown 函数
	return tracerProvider.Shutdown, nil
}

// GetTracer 获取 tracer
func GetTracer(instrumentationName string) trace.Tracer {
	return otel.Tracer(instrumentationName)
}

// StartSpan 开始一个新的 span
func StartSpan(ctx context.Context, tracerName string, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := GetTracer(tracerName)
	return tracer.Start(ctx, spanName, opts...)
}

// AddSpanTags 向当前 span 添加标签
func AddSpanTags(ctx context.Context, tags map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	for key, value := range tags {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case int:
			span.SetAttributes(attribute.Int(key, v))
		case int64:
			span.SetAttributes(attribute.Int64(key, v))
		case float64:
			span.SetAttributes(attribute.Float64(key, v))
		case bool:
			span.SetAttributes(attribute.Bool(key, v))
		default:
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
}

// AddSpanEvent 向当前 span 添加事件
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// RecordError 记录错误到当前 span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span != nil && err != nil {
		span.RecordError(err)
	}
}

// GetTraceID 获取当前 trace ID
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID 获取当前 span ID
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		return span.SpanContext().SpanID().String()
	}
	return ""
}
