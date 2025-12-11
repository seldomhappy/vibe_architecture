package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	pkgcontext "github.com/seldomhappy/vibe_architecture/internal/pkg/context"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/metrics"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/tracing"
	"github.com/seldomhappy/vibe_architecture/logger"
)

// RecoveryMiddleware handles panics and returns a 500 error
func RecoveryMiddleware(log logger.ILogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("Panic recovered: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error":"internal server error"}`)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDMiddleware generates or extracts request ID
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			ctx := pkgcontext.WithRequestID(r.Context(), requestID)
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TracingMiddleware creates a root span for the request
func TracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracing.StartSpan(r.Context(), "http-server", r.URL.Path)
			defer span.End()

			traceID := pkgcontext.GetTraceID(ctx)
			if traceID != "" {
				w.Header().Set("X-Trace-ID", traceID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(log logger.ILogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := pkgcontext.GetRequestID(r.Context())
			traceID := pkgcontext.GetTraceID(r.Context())

			log.Info("[%s][trace:%s] %s %s", requestID, traceID, r.Method, r.URL.Path)

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			log.Info("[%s][trace:%s] %s %s - %d (%v)",
				requestID, traceID, r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// MetricsMiddleware records HTTP metrics
func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			m.IncHTTPRequestsInFlight()
			defer m.DecHTTPRequestsInFlight()

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			m.RecordHTTPRequest(
				r.Method,
				r.URL.Path,
				fmt.Sprintf("%d", wrapped.statusCode),
				duration,
			)
		})
	}
}

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
