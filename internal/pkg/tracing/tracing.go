package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer holds the OpenTelemetry tracer provider
type Tracer struct {
	provider *sdktrace.TracerProvider
	enabled  bool
}

// New creates a new tracer with Jaeger exporter
func New(serviceName, jaegerEndpoint string, samplingRate float64, enabled bool) (*Tracer, error) {
	if !enabled {
		return &Tracer{enabled: false}, nil
	}

	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(samplingRate)),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)

	return &Tracer{
		provider: tp,
		enabled:  true,
	}, nil
}

// Start initializes the tracer
func (t *Tracer) Start(ctx context.Context) error {
	if !t.enabled {
		return nil
	}
	return nil
}

// Shutdown shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	if !t.enabled || t.provider == nil {
		return nil
	}
	return t.provider.Shutdown(ctx)
}

// GetTracer returns a named tracer
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a new span
func StartSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := GetTracer(tracerName)
	return tracer.Start(ctx, spanName, opts...)
}

// AddSpanAttributes adds attributes to the current span
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error in the current span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
