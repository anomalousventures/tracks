# Observability

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks provides comprehensive observability using OpenTelemetry (OTEL) for tracing, metrics, and logging. The system automatically instruments HTTP requests, database queries, and external service calls, providing vendor-agnostic observability that works with any OTEL-compatible backend.

## Goals

- Zero-config observability with OpenTelemetry
- Vendor-agnostic metrics and tracing
- Automatic instrumentation of HTTP and database calls
- Structured logging with correlation IDs
- Prometheus-compatible metrics endpoint
- Performance monitoring without overhead

## User Stories

- As a DevOps engineer, I want automatic tracing of all requests
- As a developer, I want to see slow database queries in traces
- As a developer, I want structured logs with request IDs
- As a DevOps engineer, I want metrics exported to Prometheus
- As a developer, I want to switch observability backends without code changes

## OpenTelemetry Setup

```go
// internal/otel/setup.go
package otel

import (
    "context"
    "errors"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/exporters/prometheus"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/metric"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type OTELConfig struct {
    ServiceName    string
    ServiceVersion string
    Environment    string
    ExporterType   string // "otlp-http", "otlp-grpc", "stdout", "none"
    Endpoint       string
    Headers        map[string]string
    SampleRate     float64
}

func SetupOTEL(ctx context.Context, cfg OTELConfig) (func(context.Context) error, error) {
    var shutdownFuncs []func(context.Context) error

    // Configure propagator for distributed tracing
    prop := propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    )
    otel.SetTextMapPropagator(prop)

    // Create resource with service information
    res, err := resource.Merge(
        resource.Default(),
        resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(cfg.ServiceName),
            semconv.ServiceVersion(cfg.ServiceVersion),
            semconv.DeploymentEnvironment(cfg.Environment),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Setup trace exporter
    if cfg.ExporterType != "none" {
        tracerProvider, shutdown, err := setupTracing(ctx, cfg, res)
        if err != nil {
            return nil, err
        }
        shutdownFuncs = append(shutdownFuncs, shutdown)
        otel.SetTracerProvider(tracerProvider)
    }

    // Setup metrics
    meterProvider, shutdown, err := setupMetrics(ctx, cfg, res)
    if err != nil {
        return nil, err
    }
    shutdownFuncs = append(shutdownFuncs, shutdown)
    otel.SetMeterProvider(meterProvider)

    // Return combined shutdown function
    return func(ctx context.Context) error {
        var err error
        for _, fn := range shutdownFuncs {
            err = errors.Join(err, fn(ctx))
        }
        return err
    }, nil
}

func setupTracing(ctx context.Context, cfg OTELConfig, res *resource.Resource) (*trace.TracerProvider, func(context.Context) error, error) {
    var exporter trace.SpanExporter
    var err error

    switch cfg.ExporterType {
    case "otlp-http":
        exporter, err = otlptracehttp.New(ctx,
            otlptracehttp.WithEndpoint(cfg.Endpoint),
            otlptracehttp.WithHeaders(cfg.Headers),
        )
    case "otlp-grpc":
        exporter, err = otlptracegrpc.New(ctx,
            otlptracegrpc.WithEndpoint(cfg.Endpoint),
            otlptracegrpc.WithHeaders(cfg.Headers),
            otlptracegrpc.WithInsecure(),
        )
    case "stdout":
        exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
    default:
        exporter, err = stdouttrace.New(stdouttrace.WithWriter(io.Discard))
    }

    if err != nil {
        return nil, nil, err
    }

    // Create sampler
    sampler := trace.TraceIDRatioBased(cfg.SampleRate)
    if cfg.SampleRate >= 1.0 {
        sampler = trace.AlwaysSample()
    }

    // Create tracer provider
    tracerProvider := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
        trace.WithSampler(sampler),
    )

    return tracerProvider, tracerProvider.Shutdown, nil
}

func setupMetrics(ctx context.Context, cfg OTELConfig, res *resource.Resource) (*metric.MeterProvider, func(context.Context) error, error) {
    // Prometheus exporter for metrics
    prometheusExporter, err := prometheus.New()
    if err != nil {
        return nil, nil, err
    }

    meterProvider := metric.NewMeterProvider(
        metric.WithReader(prometheusExporter),
        metric.WithResource(res),
    )

    return meterProvider, meterProvider.Shutdown, nil
}
```

## Database Instrumentation

```go
// internal/db/instrumented.go
package db

import (
    "context"
    "database/sql"

    "github.com/XSAM/otelsql"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// OpenInstrumentedDB creates an OTEL-instrumented database connection
func OpenInstrumentedDB(ctx context.Context, driverName, dsn string) (*sql.DB, error) {
    var opts []otelsql.Option

    // Set appropriate DB system based on driver
    switch driverName {
    case "go-libsql", "sqlite3":
        opts = append(opts, otelsql.WithAttributes(
            semconv.DBSystemSqlite,
        ))
    case "pgx", "postgres":
        opts = append(opts, otelsql.WithAttributes(
            semconv.DBSystemPostgreSQL,
        ))
    default:
        opts = append(opts, otelsql.WithAttributes(
            semconv.DBSystemKey.String(driverName),
        ))
    }

    // Add common instrumentation options
    opts = append(opts,
        otelsql.WithSpanOptions(otelsql.SpanOptions{
            Ping:                 true,
            RowsNext:            true,
            DisableQuery:        false,  // Include queries in traces
            DisableErrSkip:      false,
            RecordError:         func(err error) bool { return true },
            AllowRoot:           false,
            DisableSQLStatementInAttributes: false,
        }),
        otelsql.WithSQLCommenter(true),  // Add trace context to SQL comments
    )

    // Open instrumented connection
    db, err := otelsql.Open(driverName, dsn, opts...)
    if err != nil {
        return nil, err
    }

    // Register database connection metrics
    if err := otelsql.RegisterDBStatsMetrics(db, opts...); err != nil {
        db.Close()
        return nil, err
    }

    return db, nil
}
```

## HTTP Instrumentation

```go
// internal/http/middleware/tracing.go
package middleware

import (
    "net/http"

    "github.com/riandyrn/otelchi"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

// Tracing middleware for Chi router
func Tracing(serviceName string) func(http.Handler) http.Handler {
    return otelchi.Middleware(serviceName,
        otelchi.WithChiRoutes(true),
        otelchi.WithRequestMethodInSpanName(true),
    )
}

// Custom span attributes
func AddSpanAttributes(r *http.Request, attrs ...attribute.KeyValue) {
    span := trace.SpanFromContext(r.Context())
    span.SetAttributes(attrs...)
}

// Create child span for service operations
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
    return otel.Tracer("myapp").Start(ctx, name, opts...)
}
```

## Structured Logging

```go
// internal/logging/logger.go
package logging

import (
    "context"
    "os"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "go.opentelemetry.io/otel/trace"
)

// SetupLogger configures zerolog with OTEL integration
func SetupLogger(environment string) {
    // Configure output format
    if environment == "development" {
        log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
    } else {
        log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
    }

    // Set global level
    switch environment {
    case "development":
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    case "production":
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    default:
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    }
}

// LoggerWithContext adds trace information to logger
func LoggerWithContext(ctx context.Context) *zerolog.Logger {
    logger := log.With().Logger()

    // Add trace ID if available
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        logger = logger.With().
            Str("trace_id", span.SpanContext().TraceID().String()).
            Str("span_id", span.SpanContext().SpanID().String()).
            Logger()
    }

    // Add request ID if available
    if requestID, ok := ctx.Value("request_id").(string); ok {
        logger = logger.With().Str("request_id", requestID).Logger()
    }

    return &logger
}

// Log levels with context
func Debug(ctx context.Context) *zerolog.Event {
    return LoggerWithContext(ctx).Debug()
}

func Info(ctx context.Context) *zerolog.Event {
    return LoggerWithContext(ctx).Info()
}

func Warn(ctx context.Context) *zerolog.Event {
    return LoggerWithContext(ctx).Warn()
}

func Error(ctx context.Context) *zerolog.Event {
    return LoggerWithContext(ctx).Error()
}
```

## Custom Metrics

```go
// internal/metrics/metrics.go
package metrics

import (
    "context"
    "time"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/metric"
)

type Metrics struct {
    requestCounter  metric.Int64Counter
    requestDuration metric.Float64Histogram
    dbQueryDuration metric.Float64Histogram
    activeUsers     metric.Int64UpDownCounter
}

func NewMetrics() (*Metrics, error) {
    meter := otel.Meter("myapp")

    requestCounter, err := meter.Int64Counter("http_requests_total",
        metric.WithDescription("Total number of HTTP requests"),
        metric.WithUnit("1"),
    )
    if err != nil {
        return nil, err
    }

    requestDuration, err := meter.Float64Histogram("http_request_duration_seconds",
        metric.WithDescription("HTTP request duration in seconds"),
        metric.WithUnit("s"),
    )
    if err != nil {
        return nil, err
    }

    dbQueryDuration, err := meter.Float64Histogram("db_query_duration_seconds",
        metric.WithDescription("Database query duration in seconds"),
        metric.WithUnit("s"),
    )
    if err != nil {
        return nil, err
    }

    activeUsers, err := meter.Int64UpDownCounter("active_users",
        metric.WithDescription("Number of active users"),
        metric.WithUnit("1"),
    )
    if err != nil {
        return nil, err
    }

    return &Metrics{
        requestCounter:  requestCounter,
        requestDuration: requestDuration,
        dbQueryDuration: dbQueryDuration,
        activeUsers:     activeUsers,
    }, nil
}

// Record HTTP request
func (m *Metrics) RecordRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
    attrs := []attribute.KeyValue{
        attribute.String("method", method),
        attribute.String("path", path),
        attribute.Int("status", status),
    }

    m.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
    m.requestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// Record database query
func (m *Metrics) RecordQuery(ctx context.Context, query string, duration time.Duration) {
    m.dbQueryDuration.Record(ctx, duration.Seconds(),
        metric.WithAttributes(attribute.String("query", query)))
}

// Track active users
func (m *Metrics) UserLogin(ctx context.Context) {
    m.activeUsers.Add(ctx, 1)
}

func (m *Metrics) UserLogout(ctx context.Context) {
    m.activeUsers.Add(ctx, -1)
}
```

## Health & Metrics Endpoints

```go
// internal/http/handlers/health.go
package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type HealthChecker struct {
    db      *sql.DB
    storage Storage
}

func NewHealthChecker(db *sql.DB, storage Storage) *HealthChecker {
    return &HealthChecker{
        db:      db,
        storage: storage,
    }
}

// Health check endpoint
func (h *HealthChecker) Check(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().UTC(),
        "checks": map[string]string{},
    }

    // Check database
    if err := h.db.PingContext(ctx); err != nil {
        health["status"] = "unhealthy"
        health["checks"].(map[string]string)["database"] = "failed"
    } else {
        health["checks"].(map[string]string)["database"] = "ok"
    }

    // Check storage
    if _, err := h.storage.List(ctx, "health-check"); err != nil {
        health["checks"].(map[string]string)["storage"] = "failed"
    } else {
        health["checks"].(map[string]string)["storage"] = "ok"
    }

    // Set appropriate status code
    statusCode := http.StatusOK
    if health["status"] == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(health)
}

// Readiness check
func (h *HealthChecker) Ready(w http.ResponseWriter, r *http.Request) {
    // Similar to health but checks if service is ready to accept traffic
    h.Check(w, r)
}

// Liveness check
func (h *HealthChecker) Live(w http.ResponseWriter, r *http.Request) {
    // Simple check that the process is alive
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "alive",
    })
}
```

## Router Integration

```go
// internal/http/server.go
func NewServer(cfg config.Config, services *app.Services) *Server {
    r := chi.NewRouter()

    // ... other middleware ...

    // Add OTEL tracing
    r.Use(middleware.Tracing(cfg.ServiceName))

    // ... routes ...

    // Health and metrics endpoints
    r.Route("/api", func(api chi.Router) {
        api.Use(middleware.ContentTypeJSON)

        // Health checks
        api.Get("/health", healthChecker.Check)
        api.Get("/health/ready", healthChecker.Ready)
        api.Get("/health/live", healthChecker.Live)

        // Prometheus metrics
        api.Handle("/metrics", promhttp.Handler())
    })

    return &Server{Router: r}
}
```

## Configuration

```go
// internal/config/observability.go
package config

type ObservabilityConfig struct {
    // OpenTelemetry
    OTELExporter   string  `mapstructure:"OTEL_EXPORTER"`      // otlp-http, otlp-grpc, stdout, none
    OTELEndpoint   string  `mapstructure:"OTEL_ENDPOINT"`
    OTELHeaders    string  `mapstructure:"OTEL_HEADERS"`
    OTELSampleRate float64 `mapstructure:"OTEL_SAMPLE_RATE"`

    // Service info
    ServiceName    string `mapstructure:"SERVICE_NAME"`
    ServiceVersion string `mapstructure:"SERVICE_VERSION"`

    // Logging
    LogLevel  string `mapstructure:"LOG_LEVEL"`
    LogFormat string `mapstructure:"LOG_FORMAT"` // json, console
}
```

## Example Usage

```go
// cmd/server/main.go
func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatal().Err(err).Msg("failed to load config")
    }

    // Setup observability
    shutdown, err := otel.SetupOTEL(context.Background(), otel.OTELConfig{
        ServiceName:    cfg.ServiceName,
        ServiceVersion: cfg.ServiceVersion,
        Environment:    cfg.Environment,
        ExporterType:   cfg.OTELExporter,
        Endpoint:       cfg.OTELEndpoint,
        SampleRate:     cfg.OTELSampleRate,
    })
    if err != nil {
        log.Fatal().Err(err).Msg("failed to setup OTEL")
    }
    defer shutdown(context.Background())

    // Setup logging
    logging.SetupLogger(cfg.Environment)

    // Open instrumented database
    db, err := db.OpenInstrumentedDB(context.Background(), cfg.DatabaseDriver, cfg.DatabaseURL)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to open database")
    }
    defer db.Close()

    // Start server
    // ...
}
```

## Best Practices

1. **Use context propagation** - Always pass context for distributed tracing
2. **Add custom attributes** - Enrich spans with business context
3. **Sample appropriately** - 100% sampling in dev, lower in production
4. **Use structured logging** - Correlate logs with traces
5. **Monitor key metrics** - Request rate, error rate, duration
6. **Set up alerts** - Based on SLIs and SLOs
7. **Test observability** - Ensure traces and metrics work in staging

## Next Steps

- Continue to [Testing →](./13_testing.md)
- Back to [← Storage](./11_storage.md)
- Return to [Summary](./0_summary.md)
