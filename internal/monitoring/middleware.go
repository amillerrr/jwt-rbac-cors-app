package monitoring

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (m *Monitor) HTTPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		if m.Metrics != nil {
			m.Metrics.HTTPRequestsInFlight.Inc()
			defer m.Metrics.HTTPRequestsInFlight.Dec()
		}

		var span trace.Span
		ctx := r.Context()
		
		if m.Tracer != nil {
			ctx = trace.ContextWithRemoteSpanContext(ctx, trace.SpanContextFromContext(r.Context()))
			
			ctx, span = m.Tracer.Start(ctx, 
				fmt.Sprintf("%s %s", r.Method, r.URL.Path),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.target", r.URL.Path),
					attribute.String("http.host", r.Host),
					attribute.String("http.scheme", r.URL.Scheme),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
				),
			)
			defer span.End()
		}

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200 if WriteHeader isn't called
		}

		next(rw, r.WithContext(ctx))

		duration := time.Since(start)

		if m.Metrics != nil {
			m.Metrics.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(rw.statusCode),
			).Inc()

			m.Metrics.HTTPRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration.Seconds())
		}

		if span != nil {
			span.SetAttributes(
				attribute.Int("http.status_code", rw.statusCode),
				attribute.Int("http.response_size", rw.bytesWritten),
			)

			if rw.statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}
		}

		if m.Logger != nil {
			level := slog.LevelInfo
			if rw.statusCode >= 500 {
				level = slog.LevelError
			} else if rw.statusCode >= 400 {
				level = slog.LevelWarn
			}

			logAttrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Duration("duration", duration),
				slog.Int("bytes", rw.bytesWritten),
				slog.String("user_agent", r.UserAgent()),
				slog.String("remote_addr", r.RemoteAddr),
			}

			if span != nil && span.SpanContext().HasTraceID() {
				logAttrs = append(logAttrs, 
					slog.String("trace_id", span.SpanContext().TraceID().String()),
					slog.String("span_id", span.SpanContext().SpanID().String()),
				)
			}

			m.Logger.LogAttrs(r.Context(), level, "HTTP request", logAttrs...)
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func (m *Monitor) TraceSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) func() {
	if m.Tracer == nil {
		return func() {}
	}

	_, span := m.Tracer.Start(ctx, name, trace.WithAttributes(attrs...))
	return func() {
		span.End()
	}
}

func (m *Monitor) LogError(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	if m.Logger == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		attrs = append(attrs,
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	attrs = append(attrs, slog.String("error", err.Error()))

	m.Logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

func (m *Monitor) LogInfo(ctx context.Context, msg string, attrs ...slog.Attr) {
	if m.Logger == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		attrs = append(attrs,
			slog.String("trace_id", span.SpanContext().TraceID().String()),
		)
	}

	m.Logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}
