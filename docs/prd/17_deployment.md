# Deployment

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks provides comprehensive deployment support with graceful shutdown, health checks, and container-ready configurations. The framework supports multiple deployment targets including Docker, Kubernetes, AWS ECS, and traditional VMs with zero-downtime deployments.

## Goals

- Zero dropped requests during deployments
- Clean shutdown of all resources in correct order
- Container-optimized builds with minimal image size
- Support for multiple deployment platforms
- Built-in health checks and readiness probes
- Proper cleanup of database connections and background jobs

## User Stories

- As a user, I don't want my requests dropped during deployments
- As a developer, I want predictable deployment behavior
- As a DevOps engineer, I want container-ready applications
- As a developer, I want easy rollback capabilities
- As a team lead, I want consistent deployment across environments

## Graceful Shutdown

```go
// cmd/server/main.go
package main

import (
    "context"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gofrs/uuid/v5"
    "github.com/rs/zerolog/log"
)

type Application struct {
    server       *http.Server
    dbPool       *sql.DB
    jobQueue     JobQueue
    sessionStore SessionStore
    otelStop     func(context.Context) error
}

func main() {
    // Setup signal handling
    ctx, stop := signal.NotifyContext(context.Background(),
        os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Initialize app with proper resource management
    app := &Application{
        server: &http.Server{
            Addr:         ":8080",
            Handler:      setupRouter(),
            ReadTimeout:  15 * time.Second,
            WriteTimeout: 15 * time.Second,
            IdleTimeout:  60 * time.Second,
            BaseContext: func(_ net.Listener) context.Context {
                return ctx
            },
        },
        dbPool:       setupDatabase(),
        jobQueue:     setupJobQueue(),
        sessionStore: setupSessions(),
        otelStop:     setupOpenTelemetry(),
    }

    // Start server
    go func() {
        log.Info().Str("addr", app.server.Addr).
            Str("env", os.Getenv("APP_ENV")).
            Msg("Starting HTTP server")

        if err := app.server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal().Err(err).Msg("Server failed")
        }
    }()

    // Wait for interrupt signal
    <-ctx.Done()
    log.Info().Msg("Shutdown signal received")

    // Graceful shutdown with timeout
    app.gracefulShutdown()
}

func (app *Application) gracefulShutdown() {
    shutdownCtx, cancel := context.WithTimeout(
        context.Background(), 30*time.Second)
    defer cancel()

    // Shutdown sequence (order matters!)
    shutdownSteps := []struct {
        name string
        fn   func(context.Context) error
    }{
        {"HTTP server", app.server.Shutdown},
        {"Job queue", app.jobQueue.Shutdown},
        {"Session store", app.sessionStore.Close},
        {"Database", func(ctx context.Context) error {
            return app.dbPool.Close()
        }},
        {"Telemetry", app.otelStop},
    }

    for _, step := range shutdownSteps {
        log.Info().Str("component", step.name).Msg("Shutting down")

        if err := step.fn(shutdownCtx); err != nil {
            log.Error().Err(err).
                Str("component", step.name).
                Msg("Shutdown error")
        }
    }

    log.Info().Msg("Graceful shutdown complete")
}
```

## Docker Configuration

### Multi-stage Dockerfile

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always)" \
    -a -installsuffix cgo \
    -o tracks cmd/server/main.go

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 tracks && \
    adduser -D -u 1000 -G tracks tracks

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=tracks:tracks /build/tracks .

# Copy static assets
COPY --from=builder --chown=tracks:tracks /build/static ./static
COPY --from=builder --chown=tracks:tracks /build/templates ./templates

# Create data directory
RUN mkdir -p /app/data && chown -R tracks:tracks /app/data

# Switch to non-root user
USER tracks

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/tracks", "health"]

# Expose port
EXPOSE 8080

# Run application
ENTRYPOINT ["/app/tracks"]
CMD ["server"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DATABASE_URL=postgres://postgres:password@db:5432/tracks?sslmode=disable
      - REDIS_URL=redis://cache:6379
      - SESSION_SECRET=${SESSION_SECRET}
    depends_on:
      db:
        condition: service_healthy
      cache:
        condition: service_started
    volumes:
      - uploads:/app/data/uploads
      - ./config:/app/config:ro
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=tracks
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  cache:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

volumes:
  postgres_data:
  redis_data:
  uploads:
```

## Kubernetes Deployment

### Deployment Manifest

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tracks-app
  labels:
    app: tracks
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: tracks
  template:
    metadata:
      labels:
        app: tracks
    spec:
      containers:
      - name: tracks
        image: tracks:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: APP_ENV
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: tracks-secrets
              key: database-url
        - name: SESSION_SECRET
          valueFrom:
            secretKeyRef:
              name: tracks-secrets
              key: session-secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
        volumeMounts:
        - name: uploads
          mountPath: /app/data/uploads
      volumes:
      - name: uploads
        persistentVolumeClaim:
          claimName: tracks-uploads-pvc
```

### Service and Ingress

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: tracks-service
spec:
  selector:
    app: tracks
  ports:
    - port: 80
      targetPort: 8080
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tracks-ingress
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - app.example.com
    secretName: tracks-tls
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: tracks-service
            port:
              number: 80
```

## Health Checks

```go
// internal/health/checks.go
package health

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
)

type HealthStatus struct {
    Status    string                 `json:"status"`
    Version   string                 `json:"version"`
    Timestamp time.Time              `json:"timestamp"`
    Checks    map[string]CheckResult `json:"checks,omitempty"`
}

type CheckResult struct {
    Status  string        `json:"status"`
    Latency time.Duration `json:"latency_ms"`
    Error   string        `json:"error,omitempty"`
}

func Routes(r chi.Router, deps *Dependencies) {
    r.Get("/health/live", LivenessHandler())
    r.Get("/health/ready", ReadinessHandler(deps))
    r.Get("/health/startup", StartupHandler(deps))
}

// Liveness probe - is the app running?
func LivenessHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        status := HealthStatus{
            Status:    "alive",
            Version:   Version,
            Timestamp: time.Now(),
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(status)
    }
}

// Readiness probe - can the app serve traffic?
func ReadinessHandler(deps *Dependencies) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        checks := map[string]CheckResult{}
        allHealthy := true

        // Check database
        start := time.Now()
        if err := deps.DB.PingContext(ctx); err != nil {
            checks["database"] = CheckResult{
                Status:  "unhealthy",
                Latency: time.Since(start),
                Error:   err.Error(),
            }
            allHealthy = false
        } else {
            checks["database"] = CheckResult{
                Status:  "healthy",
                Latency: time.Since(start),
            }
        }

        // Check cache
        start = time.Now()
        if err := deps.Cache.Ping(ctx); err != nil {
            checks["cache"] = CheckResult{
                Status:  "unhealthy",
                Latency: time.Since(start),
                Error:   err.Error(),
            }
            // Cache is not critical
        } else {
            checks["cache"] = CheckResult{
                Status:  "healthy",
                Latency: time.Since(start),
            }
        }

        status := HealthStatus{
            Status:    "ready",
            Version:   Version,
            Timestamp: time.Now(),
            Checks:    checks,
        }

        if !allHealthy {
            status.Status = "degraded"
            w.WriteHeader(http.StatusServiceUnavailable)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(status)
    }
}
```

## Production Configuration

### Environment Variables

```bash
# .env.production
# Server
APP_ENV=production
PORT=8080
HOST=0.0.0.0

# Database
DATABASE_DRIVER=postgres
DATABASE_URL=postgres://user:pass@db.example.com/tracks?sslmode=require
DATABASE_MAX_CONNS=25
DATABASE_MAX_IDLE_CONNS=5

# Redis
REDIS_URL=redis://cache.example.com:6379
REDIS_MAX_RETRIES=3

# Sessions
SESSION_SECRET=long-random-string-minimum-32-chars
SESSION_SECURE=true
SESSION_HTTP_ONLY=true
SESSION_SAME_SITE=strict

# Security
CORS_ORIGINS=https://app.example.com
CSP_REPORT_URI=https://app.example.com/api/csp-report

# Storage
STORAGE_DRIVER=s3
AWS_REGION=us-west-2
AWS_BUCKET=tracks-uploads
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=secret

# Email
EMAIL_DRIVER=ses
EMAIL_FROM=noreply@example.com

# Monitoring
OTEL_EXPORTER_OTLP_ENDPOINT=https://otel.example.com:4317
SENTRY_DSN=https://public@sentry.example.com/123456

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

### Deployment Scripts

```bash
#!/bin/bash
# scripts/deploy.sh

set -e

# Configuration
APP_NAME="tracks"
ENVIRONMENT=${1:-staging}
VERSION=$(git describe --tags --always)

echo "Deploying ${APP_NAME} version ${VERSION} to ${ENVIRONMENT}"

# Build and tag Docker image
docker build -t ${APP_NAME}:${VERSION} .
docker tag ${APP_NAME}:${VERSION} ${APP_NAME}:latest

# Push to registry
docker push ${APP_NAME}:${VERSION}
docker push ${APP_NAME}:latest

# Deploy based on environment
case ${ENVIRONMENT} in
    production)
        # Kubernetes deployment
        kubectl set image deployment/${APP_NAME} \
            ${APP_NAME}=${APP_NAME}:${VERSION} \
            --record

        # Wait for rollout
        kubectl rollout status deployment/${APP_NAME}
        ;;

    staging)
        # Docker Swarm deployment
        docker service update \
            --image ${APP_NAME}:${VERSION} \
            ${APP_NAME}_app
        ;;

    *)
        echo "Unknown environment: ${ENVIRONMENT}"
        exit 1
        ;;
esac

echo "Deployment complete!"

# Run smoke tests
./scripts/smoke-test.sh ${ENVIRONMENT}
```

## Monitoring and Alerts

### Prometheus Metrics

```go
// internal/metrics/prometheus.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "status"},
    )

    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    DatabaseQueries = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "database_query_duration_seconds",
            Help: "Database query duration in seconds",
        },
        []string{"query_type"},
    )

    JobQueueLength = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "job_queue_length",
            Help: "Current length of job queues",
        },
        []string{"queue"},
    )
)
```

### Alert Rules

```yaml
# prometheus/alerts.yml
groups:
  - name: tracks
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: SlowResponses
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: Slow response times
          description: "95th percentile response time is {{ $value }}s"

      - alert: DatabaseDown
        expr: up{job="postgres"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: Database is down
          description: "PostgreSQL database is not responding"

      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes / 1024 / 1024 > 450
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High memory usage
          description: "Process using {{ $value }}MB of memory"
```

## Zero-Downtime Deployment

```go
// internal/server/rolling.go
package server

import (
    "context"
    "net/http"
    "time"
)

// Middleware for rolling deployments
func RollingDeploymentMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if shutdown has been initiated
        if isShuttingDown() {
            // Return 503 to load balancer
            // This signals to remove this instance from rotation
            w.Header().Set("Connection", "close")
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }

        // Add deployment version header
        w.Header().Set("X-Deployment-Version", Version)

        next.ServeHTTP(w, r)
    })
}

// Pre-stop hook for Kubernetes
func PreStopHook() {
    // Mark instance as shutting down
    setShuttingDown(true)

    // Wait for load balancer to remove from rotation
    time.Sleep(15 * time.Second)

    // Now safe to start graceful shutdown
}
```

## Deployment Checklist

1. **Pre-deployment**
   - [ ] Run full test suite
   - [ ] Check database migrations
   - [ ] Review configuration changes
   - [ ] Update documentation
   - [ ] Create rollback plan

2. **Deployment**
   - [ ] Build and tag Docker image
   - [ ] Push to container registry
   - [ ] Update deployment manifests
   - [ ] Apply configuration changes
   - [ ] Trigger rolling update

3. **Post-deployment**
   - [ ] Verify health checks passing
   - [ ] Check application metrics
   - [ ] Run smoke tests
   - [ ] Monitor error rates
   - [ ] Verify rollback capability

## Next Steps

- Continue to [Dependencies →](./18_dependencies.md)
- Back to [← TUI Mode](./16_tui_mode.md)
- Return to [Summary](./0_summary.md)
