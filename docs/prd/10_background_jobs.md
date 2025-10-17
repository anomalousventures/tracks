# Background Jobs

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks implements background job processing using external queue services (not database polling). The system uses an adapter pattern to support different queue providers, making it easy to switch between AWS SQS, Google Pub/Sub, or in-memory queues for development.

## Goals

- External queue services for reliable job processing (no database polling)
- Adapter pattern supporting AWS SQS, Google Pub/Sub, Azure Service Bus
- Local in-memory queue for development
- Automatic retries with exponential backoff
- Observable job status through queue service dashboards

## User Stories

- As a developer, I want reliable job processing without polling the database
- As a developer, I want jobs to automatically retry on failure with backoff
- As a developer, I want to use different queue services based on deployment
- As a DevOps engineer, I want to monitor jobs through AWS/GCP consoles
- As a developer, I want simple local development without external services

## Queue Adapter Interface

```go
// internal/jobs/queue.go
package jobs

import (
    "context"
    "encoding/json"
    "time"
)

type Message struct {
    ID          string          `json:"id"`
    Type        string          `json:"type"`
    Payload     json.RawMessage `json:"payload"`
    Attempts    int             `json:"attempts"`
    MaxAttempts int             `json:"max_attempts"`
    CreatedAt   time.Time       `json:"created_at"`
    ScheduledAt *time.Time      `json:"scheduled_at,omitempty"`
}

type Queue interface {
    // Send a message to the queue
    Send(ctx context.Context, queueName string, msg Message) error

    // Receive and process messages
    Receive(ctx context.Context, queueName string, handler Handler) error

    // Schedule a message for future delivery
    SendDelayed(ctx context.Context, queueName string, msg Message, delay time.Duration) error

    // Get queue statistics (optional, for monitoring)
    Stats(ctx context.Context, queueName string) (*QueueStats, error)
}

type Handler func(context.Context, Message) error

type QueueStats struct {
    Messages    int64
    InFlight    int64
    Failed      int64
    Delayed     int64
}
```

## AWS SQS Adapter

```go
// internal/jobs/adapters/sqs.go
package adapters

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/aws/aws-sdk-go-v2/service/sqs/types"
    "github.com/aws/aws-sdk-go-v2/aws"
)

type SQSAdapter struct {
    client   *sqs.Client
    queueURL string
}

func NewSQSAdapter(client *sqs.Client, queueURL string) *SQSAdapter {
    return &SQSAdapter{
        client:   client,
        queueURL: queueURL,
    }
}

func (a *SQSAdapter) Send(ctx context.Context, queueName string, msg Message) error {
    body, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    _, err = a.client.SendMessage(ctx, &sqs.SendMessageInput{
        QueueUrl:    aws.String(a.queueURL),
        MessageBody: aws.String(string(body)),
        MessageAttributes: map[string]types.MessageAttributeValue{
            "Type": {
                DataType:    aws.String("String"),
                StringValue: aws.String(msg.Type),
            },
            "Queue": {
                DataType:    aws.String("String"),
                StringValue: aws.String(queueName),
            },
        },
    })

    return err
}

func (a *SQSAdapter) Receive(ctx context.Context, queueName string, handler Handler) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Long polling for messages
            result, err := a.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
                QueueUrl:            aws.String(a.queueURL),
                MaxNumberOfMessages: 10,
                WaitTimeSeconds:     20, // Long polling
                VisibilityTimeout:   30, // 30 seconds to process
            })

            if err != nil {
                return fmt.Errorf("receive message: %w", err)
            }

            for _, sqsMsg := range result.Messages {
                var msg Message
                if err := json.Unmarshal([]byte(*sqsMsg.Body), &msg); err != nil {
                    // Log and delete malformed message
                    a.deleteMessage(ctx, sqsMsg.ReceiptHandle)
                    continue
                }

                // Process message
                if err := handler(ctx, msg); err != nil {
                    // Message will be retried by SQS
                    continue
                }

                // Delete successful message
                a.deleteMessage(ctx, sqsMsg.ReceiptHandle)
            }
        }
    }
}

func (a *SQSAdapter) SendDelayed(ctx context.Context, queueName string, msg Message, delay time.Duration) error {
    body, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    delaySeconds := int32(delay.Seconds())
    if delaySeconds > 900 { // SQS max delay is 15 minutes
        delaySeconds = 900
    }

    _, err = a.client.SendMessage(ctx, &sqs.SendMessageInput{
        QueueUrl:     aws.String(a.queueURL),
        MessageBody:  aws.String(string(body)),
        DelaySeconds: delaySeconds,
    })

    return err
}

func (a *SQSAdapter) deleteMessage(ctx context.Context, receiptHandle *string) {
    a.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
        QueueUrl:      aws.String(a.queueURL),
        ReceiptHandle: receiptHandle,
    })
}

func (a *SQSAdapter) Stats(ctx context.Context, queueName string) (*QueueStats, error) {
    attrs, err := a.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
        QueueUrl: aws.String(a.queueURL),
        AttributeNames: []types.QueueAttributeName{
            "ApproximateNumberOfMessages",
            "ApproximateNumberOfMessagesNotVisible",
            "ApproximateNumberOfMessagesDelayed",
        },
    })
    if err != nil {
        return nil, err
    }

    // Parse attributes and return stats
    return parseQueueStats(attrs.Attributes), nil
}
```

## Google Pub/Sub Adapter

```go
// internal/jobs/adapters/pubsub.go
package adapters

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "cloud.google.com/go/pubsub"
)

type PubSubAdapter struct {
    client       *pubsub.Client
    topic        *pubsub.Topic
    subscription *pubsub.Subscription
}

func NewPubSubAdapter(client *pubsub.Client, topicName, subscriptionName string) *PubSubAdapter {
    topic := client.Topic(topicName)
    subscription := client.Subscription(subscriptionName)

    return &PubSubAdapter{
        client:       client,
        topic:        topic,
        subscription: subscription,
    }
}

func (a *PubSubAdapter) Send(ctx context.Context, queueName string, msg Message) error {
    body, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    result := a.topic.Publish(ctx, &pubsub.Message{
        Data: body,
        Attributes: map[string]string{
            "type":  msg.Type,
            "queue": queueName,
        },
    })

    // Wait for confirmation
    _, err = result.Get(ctx)
    return err
}

func (a *PubSubAdapter) Receive(ctx context.Context, queueName string, handler Handler) error {
    return a.subscription.Receive(ctx, func(ctx context.Context, pubsubMsg *pubsub.Message) {
        var msg Message
        if err := json.Unmarshal(pubsubMsg.Data, &msg); err != nil {
            pubsubMsg.Ack() // Acknowledge bad message
            return
        }

        // Check queue name
        if pubsubMsg.Attributes["queue"] != queueName {
            pubsubMsg.Nack() // Return to queue
            return
        }

        // Process message
        if err := handler(ctx, msg); err != nil {
            pubsubMsg.Nack() // Retry
            return
        }

        pubsubMsg.Ack() // Success
    })
}

func (a *PubSubAdapter) SendDelayed(ctx context.Context, queueName string, msg Message, delay time.Duration) error {
    // Pub/Sub doesn't support native delay, use Cloud Tasks or scheduler
    // For now, add scheduled_at timestamp
    msg.ScheduledAt = &time.Time{}
    *msg.ScheduledAt = time.Now().Add(delay)
    return a.Send(ctx, queueName, msg)
}
```

## In-Memory Adapter (Development)

```go
// internal/jobs/adapters/memory.go
package adapters

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type MemoryAdapter struct {
    queues map[string]chan Message
    mu     sync.RWMutex
    stats  map[string]*QueueStats
}

func NewMemoryAdapter() *MemoryAdapter {
    return &MemoryAdapter{
        queues: make(map[string]chan Message),
        stats:  make(map[string]*QueueStats),
    }
}

func (a *MemoryAdapter) Send(ctx context.Context, queueName string, msg Message) error {
    a.mu.Lock()
    if _, ok := a.queues[queueName]; !ok {
        a.queues[queueName] = make(chan Message, 100)
        a.stats[queueName] = &QueueStats{}
    }
    stats := a.stats[queueName]
    a.mu.Unlock()

    select {
    case a.queues[queueName] <- msg:
        stats.Messages++
        return nil
    case <-ctx.Done():
        return ctx.Err()
    default:
        return fmt.Errorf("queue %s is full", queueName)
    }
}

func (a *MemoryAdapter) Receive(ctx context.Context, queueName string, handler Handler) error {
    a.mu.RLock()
    queue, ok := a.queues[queueName]
    stats := a.stats[queueName]
    a.mu.RUnlock()

    if !ok {
        return fmt.Errorf("queue %s not found", queueName)
    }

    for {
        select {
        case msg := <-queue:
            stats.InFlight++

            if err := handler(ctx, msg); err != nil {
                stats.Failed++

                // Retry with exponential backoff
                if msg.Attempts < msg.MaxAttempts {
                    msg.Attempts++
                    delay := time.Duration(msg.Attempts*msg.Attempts) * time.Second

                    go func() {
                        time.Sleep(delay)
                        a.Send(context.Background(), queueName, msg)
                    }()
                }
            } else {
                stats.InFlight--
            }

        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (a *MemoryAdapter) SendDelayed(ctx context.Context, queueName string, msg Message, delay time.Duration) error {
    go func() {
        time.Sleep(delay)
        a.Send(context.Background(), queueName, msg)
    }()
    return nil
}

func (a *MemoryAdapter) Stats(ctx context.Context, queueName string) (*QueueStats, error) {
    a.mu.RLock()
    stats, ok := a.stats[queueName]
    a.mu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("queue %s not found", queueName)
    }

    return stats, nil
}
```

## Worker Implementation

```go
// internal/jobs/worker.go
package jobs

import (
    "context"
    "encoding/json"
    "log/slog"
    "github.com/gofrs/uuid/v5"  // Fixed: Using correct UUID package for UUIDv7
)

type Worker struct {
    queue    Queue
    handlers map[string]Handler
    logger   *slog.Logger
}

func NewWorker(queue Queue, logger *slog.Logger) *Worker {
    return &Worker{
        queue:    queue,
        handlers: make(map[string]Handler),
        logger:   logger,
    }
}

func (w *Worker) Register(jobType string, handler Handler) {
    w.handlers[jobType] = handler
}

func (w *Worker) Start(ctx context.Context, queueName string) error {
    w.logger.Info("starting worker", "queue", queueName)

    return w.queue.Receive(ctx, queueName, func(ctx context.Context, msg Message) error {
        handler, ok := w.handlers[msg.Type]
        if !ok {
            w.logger.Error("unknown job type",
                "type", msg.Type,
                "id", msg.ID,
            )
            return nil // Don't retry unknown types
        }

        w.logger.Info("processing job",
            "id", msg.ID,
            "type", msg.Type,
            "attempt", msg.Attempts,
        )

        // Add timeout for job execution
        jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
        defer cancel()

        if err := handler(jobCtx, msg); err != nil {
            w.logger.Error("job failed",
                "id", msg.ID,
                "type", msg.Type,
                "error", err,
                "attempt", msg.Attempts,
            )
            return err // Let queue handle retry
        }

        w.logger.Info("job completed",
            "id", msg.ID,
            "type", msg.Type,
        )
        return nil
    })
}
```

## Job Handlers

```go
// internal/jobs/handlers/email.go
package handlers

import (
    "context"
    "encoding/json"
    "fmt"

    "myapp/internal/jobs"
    "myapp/internal/services/email"
)

type EmailHandler struct {
    emailService *email.Service
}

func NewEmailHandler(emailService *email.Service) *EmailHandler {
    return &EmailHandler{
        emailService: emailService,
    }
}

func (h *EmailHandler) SendWelcomeEmail(ctx context.Context, msg jobs.Message) error {
    var payload struct {
        UserID string `json:"user_id"`
        Email  string `json:"email"`
        Name   string `json:"name"`
    }

    if err := json.Unmarshal(msg.Payload, &payload); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }

    return h.emailService.SendWelcome(ctx, payload.Email, payload.Name)
}

func (h *EmailHandler) SendPasswordReset(ctx context.Context, msg jobs.Message) error {
    var payload struct {
        Email string `json:"email"`
        Token string `json:"token"`
    }

    if err := json.Unmarshal(msg.Payload, &payload); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }

    return h.emailService.SendPasswordReset(ctx, payload.Email, payload.Token)
}
```

## Usage in Services

```go
// internal/services/user_service.go
package services

import (
    "context"
    "encoding/json"
    "time"

    "github.com/gofrs/uuid/v5"  // Fixed: Using correct UUID package
    "myapp/internal/jobs"
)

func (s *UserService) Create(ctx context.Context, dto CreateUserDTO) (*User, error) {
    user, err := s.createUser(ctx, dto)
    if err != nil {
        return nil, err
    }

    // Queue welcome email
    payload, _ := json.Marshal(map[string]string{
        "user_id": user.ID,
        "email":   user.Email,
        "name":    user.Name,
    })

    msg := jobs.Message{
        ID:          uuid.Must(uuid.NewV7()).String(),  // Fixed: Using UUIDv7
        Type:        "send_welcome_email",
        Payload:     json.RawMessage(payload),
        MaxAttempts: 3,
        CreatedAt:   time.Now(),
    }

    if err := s.queue.Send(ctx, "emails", msg); err != nil {
        // Log but don't fail user creation
        slog.Error("failed to queue welcome email",
            "error", err,
            "user_id", user.ID,
        )
    }

    // Schedule follow-up email in 3 days
    followupPayload, _ := json.Marshal(map[string]string{
        "user_id": user.ID,
        "email":   user.Email,
    })

    followupMsg := jobs.Message{
        ID:          uuid.Must(uuid.NewV7()).String(),  // Fixed: Using UUIDv7
        Type:        "send_onboarding_tips",
        Payload:     json.RawMessage(followupPayload),
        MaxAttempts: 3,
        CreatedAt:   time.Now(),
    }

    s.queue.SendDelayed(ctx, "emails", followupMsg, 72*time.Hour)

    return user, nil
}
```

## Configuration

```go
// internal/config/queue.go
package config

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "cloud.google.com/go/pubsub"

    "myapp/internal/jobs"
    "myapp/internal/jobs/adapters"
)

func NewQueue(cfg QueueConfig) (jobs.Queue, error) {
    switch cfg.Provider {
    case "sqs":
        awsCfg, err := config.LoadDefaultConfig(context.Background())
        if err != nil {
            return nil, err
        }
        client := sqs.NewFromConfig(awsCfg)
        return adapters.NewSQSAdapter(client, cfg.URL), nil

    case "pubsub":
        client, err := pubsub.NewClient(context.Background(), cfg.ProjectID)
        if err != nil {
            return nil, err
        }
        return adapters.NewPubSubAdapter(
            client,
            cfg.TopicName,
            cfg.SubscriptionName,
        ), nil

    case "memory":
        return adapters.NewMemoryAdapter(), nil

    default:
        return nil, fmt.Errorf("unknown queue provider: %s", cfg.Provider)
    }
}
```

## Starting Workers

```go
// cmd/worker/main.go
package main

import (
    "context"
    "os"
    "os/signal"
    "sync"
    "syscall"

    "myapp/internal/config"
    "myapp/internal/jobs"
    "myapp/internal/jobs/handlers"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("failed to load config:", err)
    }

    // Create queue
    queue, err := config.NewQueue(cfg.Queue)
    if err != nil {
        log.Fatal("failed to create queue:", err)
    }

    // Create services
    services := setupServices(cfg)

    // Create handlers
    emailHandler := handlers.NewEmailHandler(services.Email)

    // Create and configure workers
    emailWorker := jobs.NewWorker(queue, logger)
    emailWorker.Register("send_welcome_email", emailHandler.SendWelcomeEmail)
    emailWorker.Register("send_password_reset", emailHandler.SendPasswordReset)
    emailWorker.Register("send_onboarding_tips", emailHandler.SendOnboardingTips)

    // Start workers
    ctx, cancel := context.WithCancel(context.Background())
    var wg sync.WaitGroup

    // Email worker
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := emailWorker.Start(ctx, "emails"); err != nil {
            log.Error("email worker failed:", err)
        }
    }()

    // Wait for interrupt
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    // Graceful shutdown
    log.Info("shutting down workers...")
    cancel()
    wg.Wait()
    log.Info("workers stopped")
}
```

## Monitoring

```go
// internal/jobs/metrics.go
package jobs

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    jobsQueued = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "jobs_queued_total",
            Help: "Total number of jobs queued",
        },
        []string{"type", "queue"},
    )

    jobsProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "jobs_processed_total",
            Help: "Total number of jobs processed",
        },
        []string{"type", "status"},
    )

    jobDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "job_duration_seconds",
            Help: "Time taken to process jobs",
        },
        []string{"type"},
    )
)

func RecordJobQueued(jobType, queue string) {
    jobsQueued.WithLabelValues(jobType, queue).Inc()
}

func RecordJobProcessed(jobType, status string) {
    jobsProcessed.WithLabelValues(jobType, status).Inc()
}

func RecordJobDuration(jobType string, seconds float64) {
    jobDuration.WithLabelValues(jobType).Observe(seconds)
}
```

## Best Practices

1. **Use external queues, not database polling** - More scalable and reliable
2. **Set appropriate retry limits** - Don't retry forever
3. **Use exponential backoff** - Avoid overwhelming failed services
4. **Add job timeouts** - Prevent stuck jobs from blocking workers
5. **Monitor queue depth** - Alert on queue buildup
6. **Use dead letter queues** - Handle permanently failed messages
7. **Log job lifecycle** - Track queued, processing, completed, failed

## Testing

```go
// internal/jobs/worker_test.go
func TestWorker_ProcessMessage(t *testing.T) {
    queue := adapters.NewMemoryAdapter()
    worker := NewWorker(queue, slog.Default())

    // Register handler
    processed := false
    worker.Register("test_job", func(ctx context.Context, msg Message) error {
        processed = true
        return nil
    })

    // Queue message
    msg := Message{
        ID:          "test-123",
        Type:        "test_job",
        Payload:     json.RawMessage(`{}`),
        MaxAttempts: 1,
    }

    err := queue.Send(context.Background(), "test", msg)
    assert.NoError(t, err)

    // Process in background
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    go worker.Start(ctx, "test")

    // Wait for processing
    time.Sleep(100 * time.Millisecond)

    assert.True(t, processed)
}
```

## Next Steps

- Continue to [Storage →](./11_storage.md)
- Back to [← Configuration](./9_configuration.md)
- Return to [Summary](./0_summary.md)
