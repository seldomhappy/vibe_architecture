package context

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	requestIDKey     contextKey = "request_id"
	userIDKey        contextKey = "user_id"
	correlationIDKey contextKey = "correlation_id"
)

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) int64 {
	if userID, ok := ctx.Value(userIDKey).(int64); ok {
		return userID
	}
	return 0
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// GetTraceID retrieves the trace ID from the OpenTelemetry span context
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID retrieves the span ID from the OpenTelemetry span context
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// NewRequestContext creates a new context with request ID and optional user ID
func NewRequestContext(ctx context.Context, requestID string, userID int64) context.Context {
	ctx = WithRequestID(ctx, requestID)
	if userID > 0 {
		ctx = WithUserID(ctx, userID)
	}
	return ctx
}
