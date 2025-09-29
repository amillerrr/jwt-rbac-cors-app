package monitoring

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Monitor struct {
	Logger        *slog.Logger
	TracerProvider *sdktrace.TracerProvider
	Tracer        trace.Tracer
	Metrics       *Metrics
	logFile       *os.File
}

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	LoginAttempts        *prometheus.CounterVec
	LoginSuccesses       prometheus.Counter
	LoginFailures        *prometheus.CounterVec
	RegistrationAttempts prometheus.Counter
	TokenGenerations     prometheus.Counter
	TokenValidations     *prometheus.CounterVec

	DBQueriesTotal    *prometheus.CounterVec
	DBQueryDuration   *prometheus.HistogramVec
	DBConnectionsOpen prometheus.Gauge

	UsersTotal        prometheus.Gauge
	UsersActive       prometheus.Gauge
	ProductsTotal     prometheus.Gauge
}

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	LogLevel       slog.Level
	LogFormat      string // "json" or "text"
	OTLPEndpoint   string // e.g., "localhost:4318" for Jaeger
	EnableMetrics  bool
	EnableTracing  bool
	EnableLogging  bool
}

func NewMonitor(cfg Config) (*Monitor, error) {
	m := &Monitor{}

	if cfg.EnableLogging {
		if err := m.initLogger(cfg); err != nil {
			return nil, fmt.Errorf("failed to initialize logger: %w", err)
		}
	}

	if cfg.EnableMetrics {
		m.initMetrics()
	}

	if cfg.EnableTracing {
		if err := m.initTracing(cfg); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	return m, nil
}

func (m *Monitor) initLogger(cfg Config) error {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	logFile, err := os.OpenFile(
		fmt.Sprintf("logs/app-%s.log", time.Now().Format("2006-01-02")),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	m.logFile = logFile

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	var handler slog.Handler
	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level: cfg.LogLevel,
			AddSource: true,
		})
	} else {
		handler = slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
			Level: cfg.LogLevel,
			AddSource: true,
		})
	}

	m.Logger = slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
		slog.String("version", cfg.ServiceVersion),
		slog.String("environment", cfg.Environment),
	)

	slog.SetDefault(m.Logger)
	return nil
}

func (m *Monitor) initMetrics() {
	m.Metrics = &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests by method, endpoint, and status",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		LoginAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_login_attempts_total",
				Help: "Total number of login attempts by result",
			},
			[]string{"result"}, // "success" or "failure"
		),
		LoginSuccesses: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "auth_login_successes_total",
				Help: "Total number of successful logins",
			},
		),
		LoginFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_login_failures_total",
				Help: "Total number of failed logins by reason",
			},
			[]string{"reason"}, // "invalid_credentials", "user_not_found", etc.
		),
		RegistrationAttempts: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "auth_registration_attempts_total",
				Help: "Total number of user registration attempts",
			},
		),
		TokenGenerations: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "auth_token_generations_total",
				Help: "Total number of JWT tokens generated",
			},
		),
		TokenValidations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_token_validations_total",
				Help: "Total number of token validations by result",
			},
			[]string{"result"}, // "valid", "invalid", "expired"
		),

		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries by operation and status",
			},
			[]string{"operation", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation"},
		),
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_open",
				Help: "Current number of open database connections",
			},
		),

		UsersTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "auth_app_users_total",
				Help: "Total number of registered users",
			},
		),
		UsersActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "auth_app_users_active_total",
				Help: "Number of active users (logged in last 24 hours)",
			},
		),
		ProductsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "auth_app_products_total",
				Help: "Total number of products in the system",
			},
		),
	}
}

func (m *Monitor) initTracing(cfg Config) error {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // Use HTTP (not HTTPS) for local development
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	m.TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(m.TracerProvider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	m.Tracer = m.TracerProvider.Tracer(cfg.ServiceName)

	return nil
}

func (m *Monitor) Shutdown(ctx context.Context) error {
	var errs []error

	if m.TracerProvider != nil {
		if err := m.TracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer: %w", err))
		}
	}

	if m.logFile != nil {
		if err := m.logFile.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close log file: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}
