# External Services

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks implements resilient integration patterns for external services using circuit breakers and the adapter pattern. This ensures that external service failures don't cascade through your application and allows easy switching between service providers.

## Goals

- Resilient integration with external services via circuit breakers
- Prevent cascading failures when services degrade
- Adapter pattern for swappable implementations
- Automatic retries with exponential backoff
- Service-specific failure thresholds

## User Stories

- As a developer, I want external service failures to not crash my app
- As a developer, I want to easily switch between email providers
- As a user, I want the app to stay responsive even when external services are slow
- As a DevOps engineer, I want circuit breaker metrics for monitoring
- As a developer, I want automatic retries for transient failures

## Circuit Breaker Pattern

Circuit breakers prevent cascading failures by monitoring service health and opening the circuit when failure thresholds are exceeded. This stops sending requests to failing services until they recover.

### Circuit Breaker Configuration

```go
// internal/pkg/circuitbreaker/breaker.go
package circuitbreaker

import (
    "time"
    "github.com/sony/gobreaker"
    "github.com/rs/zerolog/log"
)

type ServiceBreakers struct {
    Email   *gobreaker.CircuitBreaker[bool]
    SMS     *gobreaker.CircuitBreaker[bool]
    Storage *gobreaker.CircuitBreaker[bool]
}

func NewServiceBreakers() *ServiceBreakers {
    return &ServiceBreakers{
        Email: gobreaker.NewCircuitBreaker[bool](gobreaker.Settings{
            Name:        "EmailService",
            MaxRequests: 5,                // Try 5 requests when half-open
            Interval:    30 * time.Second,  // Reset failure count after 30s
            Timeout:     120 * time.Second, // Move from open to half-open after 2m
            ReadyToTrip: func(counts gobreaker.Counts) bool {
                // Open circuit if 30% of requests fail (minimum 5 requests)
                failureRatio := float64(counts.TotalFailures) /
                               float64(counts.Requests)
                return counts.Requests >= 5 && failureRatio >= 0.3
            },
            OnStateChange: func(name string, from, to gobreaker.State) {
                if to == gobreaker.StateOpen {
                    log.Error().Str("service", name).
                        Msg("Circuit breaker opened")
                } else if from == gobreaker.StateOpen && to == gobreaker.StateHalfOpen {
                    log.Info().Str("service", name).
                        Msg("Circuit breaker entering half-open state")
                } else if from == gobreaker.StateHalfOpen && to == gobreaker.StateClosed {
                    log.Info().Str("service", name).
                        Msg("Circuit breaker closed - service recovered")
                }
            },
        }),
        SMS: gobreaker.NewCircuitBreaker[bool](gobreaker.Settings{
            Name:        "SMSService",
            MaxRequests: 3,
            Interval:    30 * time.Second,
            Timeout:     60 * time.Second,
            ReadyToTrip: func(counts gobreaker.Counts) bool {
                // Open after 3 consecutive failures
                return counts.ConsecutiveFailures >= 3
            },
        }),
        Storage: gobreaker.NewCircuitBreaker[bool](gobreaker.Settings{
            Name:        "S3Storage",
            MaxRequests: 10,
            Interval:    30 * time.Second,
            Timeout:     180 * time.Second,
            ReadyToTrip: func(counts gobreaker.Counts) bool {
                // Open if 50% of requests fail (minimum 10 requests)
                failureRatio := float64(counts.TotalFailures) /
                               float64(counts.Requests)
                return counts.Requests >= 10 && failureRatio >= 0.5
            },
        }),
    }
}
```

### Circuit Breaker States

1. **Closed**: Normal operation, requests pass through
2. **Open**: Service is failing, requests immediately return error
3. **Half-Open**: Testing if service recovered, limited requests allowed

## Email Service

### Email Adapter Interface

```go
// internal/interfaces/email.go
package interfaces

import (
    "context"
    "time"
)

type EmailAdapter interface {
    Send(ctx context.Context, msg *EmailMessage) error
    SendWithRetry(ctx context.Context, msg *EmailMessage, maxRetries int) error
    GetProvider() string
}

type EmailMessage struct {
    To          []string
    From        string
    Subject     string
    HTML        string
    Text        string
    Headers     map[string]string
    Attachments []Attachment
}

type Attachment struct {
    Filename    string
    ContentType string
    Data        []byte
}

// Base adapter with common retry logic
type BaseAdapter struct {
    provider string
}

func (b *BaseAdapter) GetProvider() string {
    return b.provider
}

func (b *BaseAdapter) SendWithRetry(ctx context.Context, msg *EmailMessage, maxRetries int) error {
    var err error
    for i := 0; i <= maxRetries; i++ {
        if err = b.Send(ctx, msg); err == nil {
            return nil
        }

        if i < maxRetries {
            // Exponential backoff: 1s, 2s, 4s, 8s...
            backoff := time.Second * time.Duration(1<<uint(i))
            select {
            case <-time.After(backoff):
                continue
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }
    return err
}
```

### AWS SES Adapter

```go
// internal/adapters/email/ses.go
package email

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ses"
    "github.com/aws/aws-sdk-go-v2/service/ses/types"
    "github.com/sony/gobreaker"
)

type SESAdapter struct {
    BaseAdapter
    client *ses.Client
    cb     *gobreaker.CircuitBreaker[bool]
}

func NewSESAdapter(client *ses.Client, cb *gobreaker.CircuitBreaker[bool]) *SESAdapter {
    return &SESAdapter{
        BaseAdapter: BaseAdapter{provider: "AWS SES"},
        client:      client,
        cb:          cb,
    }
}

func (a *SESAdapter) Send(ctx context.Context, msg *EmailMessage) error {
    _, err := a.cb.Execute(func() (bool, error) {
        input := &ses.SendEmailInput{
            Destination: &types.Destination{
                ToAddresses: msg.To,
            },
            Message: &types.Message{
                Body: &types.Body{
                    Html: &types.Content{
                        Data: aws.String(msg.HTML),
                    },
                    Text: &types.Content{
                        Data: aws.String(msg.Text),
                    },
                },
                Subject: &types.Content{
                    Data: aws.String(msg.Subject),
                },
            },
            Source: aws.String(msg.From),
        }

        // Add custom headers if present
        if len(msg.Headers) > 0 {
            var headers []types.MessageHeader
            for k, v := range msg.Headers {
                headers = append(headers, types.MessageHeader{
                    Name:  aws.String(k),
                    Value: aws.String(v),
                })
            }
            input.Headers = headers
        }

        _, err := a.client.SendEmail(ctx, input)
        return true, err
    })

    return err
}
```

### Mailpit Adapter (Development)

```go
// internal/adapters/email/mailpit.go
package email

import (
    "context"
    "gopkg.in/gomail.v2"
)

type MailpitAdapter struct {
    BaseAdapter
    host string
    port int
}

func NewMailpitAdapter(host string, port int) *MailpitAdapter {
    return &MailpitAdapter{
        BaseAdapter: BaseAdapter{provider: "Mailpit"},
        host:        host,
        port:        port,
    }
}

func (a *MailpitAdapter) Send(ctx context.Context, msg *EmailMessage) error {
    m := gomail.NewMessage()
    m.SetHeader("From", msg.From)
    m.SetHeader("To", msg.To...)
    m.SetHeader("Subject", msg.Subject)
    m.SetBody("text/html", msg.HTML)
    m.AddAlternative("text/plain", msg.Text)

    // Add custom headers
    for k, v := range msg.Headers {
        m.SetHeader(k, v)
    }

    // Add attachments
    for _, att := range msg.Attachments {
        m.Attach(att.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
            _, err := w.Write(att.Data)
            return err
        }))
    }

    d := gomail.NewDialer(a.host, a.port, "", "")
    return d.DialAndSend(m)
}
```

## SMS Service

### SMS Adapter Interface

```go
// internal/interfaces/sms.go
package interfaces

import (
    "context"
)

type SMSAdapter interface {
    SendOTP(ctx context.Context, phone, code string) error
    SendMessage(ctx context.Context, phone, message string) error
    VerifyOTP(ctx context.Context, phone, code string) error
    GetProvider() string
}
```

### AWS SNS Adapter

```go
// internal/adapters/sms/sns.go
package sms

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sns"
    "github.com/sony/gobreaker"
)

type SNSAdapter struct {
    client   *sns.Client
    cb       *gobreaker.CircuitBreaker[bool]
    provider string
}

func NewSNSAdapter(client *sns.Client, cb *gobreaker.CircuitBreaker[bool]) *SNSAdapter {
    return &SNSAdapter{
        client:   client,
        cb:       cb,
        provider: "AWS SNS",
    }
}

func (a *SNSAdapter) SendOTP(ctx context.Context, phone, code string) error {
    message := fmt.Sprintf("Your verification code is: %s", code)
    return a.SendMessage(ctx, phone, message)
}

func (a *SNSAdapter) SendMessage(ctx context.Context, phone, message string) error {
    _, err := a.cb.Execute(func() (bool, error) {
        input := &sns.PublishInput{
            Message:     aws.String(message),
            PhoneNumber: aws.String(phone),
        }

        _, err := a.client.Publish(ctx, input)
        return true, err
    })

    return err
}

func (a *SNSAdapter) VerifyOTP(ctx context.Context, phone, code string) error {
    // SNS doesn't have built-in OTP verification
    // This would be handled by your application logic
    return nil
}

func (a *SNSAdapter) GetProvider() string {
    return a.provider
}
```

### Twilio Adapter

```go
// internal/adapters/sms/twilio.go
package sms

import (
    "context"
    "github.com/twilio/twilio-go"
    verify "github.com/twilio/twilio-go/rest/verify/v2"
    "github.com/sony/gobreaker"
)

type TwilioAdapter struct {
    client    *twilio.RestClient
    serviceID string
    cb        *gobreaker.CircuitBreaker[bool]
    provider  string
}

func NewTwilioAdapter(client *twilio.RestClient, serviceID string, cb *gobreaker.CircuitBreaker[bool]) *TwilioAdapter {
    return &TwilioAdapter{
        client:    client,
        serviceID: serviceID,
        cb:        cb,
        provider:  "Twilio",
    }
}

func (a *TwilioAdapter) SendOTP(ctx context.Context, phone, code string) error {
    _, err := a.cb.Execute(func() (bool, error) {
        params := &verify.CreateVerificationParams{}
        params.SetTo(phone)
        params.SetChannel("sms")

        _, err := a.client.VerifyV2.CreateVerification(
            a.serviceID, params)
        return true, err
    })

    return err
}

func (a *TwilioAdapter) SendMessage(ctx context.Context, phone, message string) error {
    _, err := a.cb.Execute(func() (bool, error) {
        params := &api.CreateMessageParams{}
        params.SetTo(phone)
        params.SetBody(message)

        _, err := a.client.Api.CreateMessage(params)
        return true, err
    })

    return err
}

func (a *TwilioAdapter) VerifyOTP(ctx context.Context, phone, code string) error {
    _, err := a.cb.Execute(func() (bool, error) {
        params := &verify.CreateVerificationCheckParams{}
        params.SetTo(phone)
        params.SetCode(code)

        check, err := a.client.VerifyV2.CreateVerificationCheck(
            a.serviceID, params)
        if err != nil {
            return false, err
        }

        if check.Status != nil && *check.Status != "approved" {
            return false, ErrInvalidOTP
        }

        return true, nil
    })

    return err
}

func (a *TwilioAdapter) GetProvider() string {
    return a.provider
}
```

## Storage Service

See [Storage →](./11_storage.md) for detailed storage service implementation with circuit breakers.

## Service Factory

```go
// internal/adapters/factory.go
package adapters

import (
    "fmt"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ses"
    "github.com/aws/aws-sdk-go-v2/service/sns"
    "github.com/twilio/twilio-go"
)

type ServiceFactory struct {
    config   *Config
    breakers *ServiceBreakers
}

func NewServiceFactory(cfg *Config) (*ServiceFactory, error) {
    return &ServiceFactory{
        config:   cfg,
        breakers: NewServiceBreakers(),
    }, nil
}

func (f *ServiceFactory) CreateEmailAdapter() (EmailAdapter, error) {
    switch f.config.EmailProvider {
    case "ses":
        cfg, err := config.LoadDefaultConfig(context.Background())
        if err != nil {
            return nil, err
        }
        client := ses.NewFromConfig(cfg)
        return NewSESAdapter(client, f.breakers.Email), nil

    case "mailpit":
        return NewMailpitAdapter(
            f.config.MailpitHost,
            f.config.MailpitPort,
        ), nil

    case "sendgrid":
        return NewSendGridAdapter(
            f.config.SendGridAPIKey,
            f.breakers.Email,
        ), nil

    default:
        return nil, fmt.Errorf("unknown email provider: %s", f.config.EmailProvider)
    }
}

func (f *ServiceFactory) CreateSMSAdapter() (SMSAdapter, error) {
    switch f.config.SMSProvider {
    case "sns":
        cfg, err := config.LoadDefaultConfig(context.Background())
        if err != nil {
            return nil, err
        }
        client := sns.NewFromConfig(cfg)
        return NewSNSAdapter(client, f.breakers.SMS), nil

    case "twilio":
        client := twilio.NewRestClient()
        return NewTwilioAdapter(
            client,
            f.config.TwilioServiceID,
            f.breakers.SMS,
        ), nil

    case "log":
        return NewLogAdapter(), nil

    default:
        return nil, fmt.Errorf("unknown SMS provider: %s", f.config.SMSProvider)
    }
}
```

## Monitoring Circuit Breakers

```go
// internal/pkg/monitoring/monitoring.go
package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    circuitBreakerState = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "circuit_breaker_state",
            Help: "Current state of circuit breaker (0=closed, 1=half-open, 2=open)",
        },
        []string{"service"},
    )

    circuitBreakerRequests = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "circuit_breaker_requests_total",
            Help: "Total number of requests through circuit breaker",
        },
        []string{"service", "result"},
    )
)

func RecordCircuitBreakerState(service string, state gobreaker.State) {
    var value float64
    switch state {
    case gobreaker.StateClosed:
        value = 0
    case gobreaker.StateHalfOpen:
        value = 1
    case gobreaker.StateOpen:
        value = 2
    }
    circuitBreakerState.WithLabelValues(service).Set(value)
}

func RecordCircuitBreakerRequest(service, result string) {
    circuitBreakerRequests.WithLabelValues(service, result).Inc()
}
```

## Best Practices

1. **Configure appropriate thresholds** - Each service has different reliability characteristics
2. **Use exponential backoff** - Don't overwhelm recovering services
3. **Monitor circuit breaker states** - Set up alerts for open circuits
4. **Test failure scenarios** - Ensure graceful degradation
5. **Provide fallback mechanisms** - Queue messages when services are down
6. **Log state changes** - Track when and why circuits open
7. **Use timeouts** - Prevent slow services from blocking

## Testing

```go
// internal/adapters/email/ses_test.go
func TestSESAdapter_CircuitBreaker(t *testing.T) {
    // Create mock client that fails
    mockClient := &mockSESClient{
        shouldFail: true,
    }

    cb := gobreaker.NewCircuitBreaker[bool](gobreaker.Settings{
        Name:        "TestEmail",
        MaxRequests: 1,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures >= 2
        },
    })

    adapter := &SESAdapter{
        client: mockClient,
        cb:     cb,
    }

    // First two requests fail and open the circuit
    for i := 0; i < 2; i++ {
        err := adapter.Send(context.Background(), &EmailMessage{})
        assert.Error(t, err)
    }

    // Circuit should be open, request fails immediately
    err := adapter.Send(context.Background(), &EmailMessage{})
    assert.Error(t, err)
    assert.Equal(t, gobreaker.ErrOpenState, err)
}
```

## Next Steps

- Continue to [Configuration →](./9_configuration.md)
- Back to [← Templates & Assets](./7_templates_assets.md)
- Return to [Summary](./0_summary.md)
