# Testing

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks provides a comprehensive testing framework using **testify** for assertions and mocking. The system supports unit tests, integration tests with real databases, and end-to-end tests using Hurl. Test isolation is achieved through transaction rollback and temporary databases.

## Goals

- Fast unit tests with minimal setup
- Integration tests with real databases
- E2E tests with Hurl for API testing
- Test data factories for consistent fixtures
- Automatic transaction rollback for test isolation
- Parallel test execution where possible

## User Stories

- As a developer, I want fast unit tests that run in milliseconds
- As a developer, I want integration tests with real database
- As a developer, I want test data factories for easy setup
- As a QA engineer, I want E2E tests for critical paths
- As a developer, I want tests to clean up after themselves

## Test Structure

```text
tests/
├── unit/              # Unit tests (mocked dependencies)
├── integration/       # Integration tests (real database)
├── e2e/              # End-to-end tests
│   ├── api/          # Hurl test files
│   └── browser/      # Optional Playwright tests
├── fixtures/         # Test data files
├── factories/        # Test data factories
└── helpers/          # Test utilities
```

## Test Database Setup

```go
// test/testutil/db.go
package testutil

import (
    "context"
    "database/sql"
    "testing"
    "path/filepath"
    "os"

    _ "github.com/tursodatabase/libsql-client-go/libsql"
    _ "github.com/lib/pq"
    "github.com/pressly/goose/v3"
)

// TestDB wraps a test database with transaction support
type TestDB struct {
    *sql.DB
    tx *sql.Tx
}

// SetupTestDB creates a test database based on driver
func SetupTestDB(t *testing.T, driver string) *TestDB {
    var db *sql.DB
    var err error

    switch driver {
    case "go-libsql", "sqlite3":
        db, err = setupSQLiteTestDB(t)
    case "postgres":
        db, err = setupPostgresTestDB(t)
    default:
        t.Fatalf("unsupported driver: %s", driver)
    }

    require.NoError(t, err)

    t.Cleanup(func() {
        db.Close()
    })

    // Run migrations
    runMigrations(t, db, driver)

    return &TestDB{DB: db}
}

func setupSQLiteTestDB(t *testing.T) (*sql.DB, error) {
    // Create temp SQLite database for testing
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    return sql.Open("libsql", "file:"+dbPath)
}

func setupPostgresTestDB(t *testing.T) (*sql.DB, error) {
    // Use test database (requires POSTGRES_TEST_URL env var)
    dsn := os.Getenv("POSTGRES_TEST_URL")
    if dsn == "" {
        t.Skip("POSTGRES_TEST_URL not set")
    }

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    // Create unique schema for this test
    schemaName := fmt.Sprintf("test_%d", time.Now().UnixNano())
    _, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
    if err != nil {
        return nil, err
    }

    // Set search path
    _, err = db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName))
    if err != nil {
        return nil, err
    }

    t.Cleanup(func() {
        db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", schemaName))
    })

    return db, nil
}

func runMigrations(t *testing.T, db *sql.DB, driver string) {
    // Use goose for migrations
    if err := goose.SetDialect(driver); err != nil {
        t.Fatal(err)
    }

    migrationsDir := "../../db/migrations"
    if err := goose.Up(db, migrationsDir); err != nil {
        t.Fatal(err)
    }
}

// BeginTx starts a transaction for test isolation
func (tdb *TestDB) BeginTx(t *testing.T) *sql.Tx {
    tx, err := tdb.DB.Begin()
    require.NoError(t, err)

    tdb.tx = tx

    t.Cleanup(func() {
        tx.Rollback()
    })

    return tx
}
```

## Test Factories

```go
// test/factories/user_factory.go
package factories

import (
    "fmt"
    "time"

    "github.com/gofrs/uuid/v5"
    "github.com/jaswdr/faker"
)

var fake = faker.New()

type UserFactory struct {
    faker faker.Faker
}

func NewUserFactory() *UserFactory {
    return &UserFactory{
        faker: faker.New(),
    }
}

type UserParams struct {
    Email    string
    Username string
    Name     string
    Verified bool
}

func (f *UserFactory) Build(params ...UserParams) *User {
    var p UserParams
    if len(params) > 0 {
        p = params[0]
    }

    // Use provided values or generate fake ones
    if p.Email == "" {
        p.Email = f.faker.Internet().Email()
    }
    if p.Username == "" {
        p.Username = f.faker.Internet().User()
    }
    if p.Name == "" {
        p.Name = f.faker.Person().Name()
    }

    return &User{
        ID:         uuid.Must(uuid.NewV7()).String(),
        Email:      p.Email,
        Username:   p.Username,
        Name:       p.Name,
        Verified:   p.Verified,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }
}

func (f *UserFactory) BuildMany(count int, params ...UserParams) []*User {
    users := make([]*User, count)
    for i := 0; i < count; i++ {
        users[i] = f.Build(params...)
    }
    return users
}

func (f *UserFactory) Create(db *sql.DB, params ...UserParams) (*User, error) {
    user := f.Build(params...)

    _, err := db.Exec(`
        INSERT INTO users (id, email, username, name, verified, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, user.ID, user.Email, user.Username, user.Name, user.Verified, user.CreatedAt, user.UpdatedAt)

    if err != nil {
        return nil, err
    }

    return user, nil
}
```

## Unit Tests

```go
// internal/services/user_service_test.go
package services_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "myapp/internal/services"
    "myapp/test/mocks"
)

func TestUserService_Create(t *testing.T) {
    // Arrange
    mockRepo := new(mocks.UserRepository)
    mockEmail := new(mocks.EmailService)

    svc := services.NewUserService(mockRepo, mockEmail)

    dto := services.CreateUserDTO{
        Email: "test@example.com",
        Name:  "Test User",
    }

    expectedUser := &User{
        ID:    "123",
        Email: dto.Email,
        Name:  dto.Name,
    }

    // Setup expectations
    mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *User) bool {
        return u.Email == dto.Email && u.Name == dto.Name
    })).Return(nil).Once()

    mockEmail.On("SendWelcome", mock.Anything, dto.Email, dto.Name).Return(nil).Once()

    // Act
    user, err := svc.Create(context.Background(), dto)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, dto.Email, user.Email)
    assert.NotEmpty(t, user.ID)

    // Verify mocks
    mockRepo.AssertExpectations(t)
    mockEmail.AssertExpectations(t)
}

func TestUserService_Create_EmailFailure(t *testing.T) {
    // Test that user is still created even if email fails
    mockRepo := new(mocks.UserRepository)
    mockEmail := new(mocks.EmailService)

    svc := services.NewUserService(mockRepo, mockEmail)

    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    mockEmail.On("SendWelcome", mock.Anything, mock.Anything, mock.Anything).
        Return(errors.New("email service down"))

    user, err := svc.Create(context.Background(), services.CreateUserDTO{
        Email: "test@example.com",
        Name:  "Test User",
    })

    // User creation should succeed despite email failure
    require.NoError(t, err)
    assert.NotNil(t, user)
}
```

## Integration Tests

```go
// internal/services/user_service_integration_test.go
//go:build integration
// +build integration

package services_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "myapp/internal/repositories"
    "myapp/internal/services"
    "myapp/test/testutil"
    "myapp/test/factories"
)

func TestUserService_Integration(t *testing.T) {
    // Setup test database
    db := testutil.SetupTestDB(t, "go-libsql")
    tx := db.BeginTx(t)

    // Create real repositories
    userRepo := repositories.NewUserRepository(tx)

    // Use mock for external services
    mockEmail := new(mocks.EmailService)
    mockEmail.On("SendWelcome", mock.Anything, mock.Anything, mock.Anything).Return(nil)

    // Create service with real repo
    svc := services.NewUserService(userRepo, mockEmail)

    t.Run("Create and Retrieve User", func(t *testing.T) {
        dto := services.CreateUserDTO{
            Email:    "integration@test.com",
            Username: "integrationtest",
            Name:     "Integration Test",
        }

        // Create user
        user, err := svc.Create(context.Background(), dto)
        require.NoError(t, err)
        assert.NotEmpty(t, user.ID)

        // Retrieve user
        retrieved, err := svc.GetByID(context.Background(), user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Email, retrieved.Email)
    })

    t.Run("Duplicate Email", func(t *testing.T) {
        factory := factories.NewUserFactory()

        // Create first user
        user1, err := factory.Create(tx, factories.UserParams{
            Email: "duplicate@test.com",
        })
        require.NoError(t, err)

        // Try to create second user with same email
        dto := services.CreateUserDTO{
            Email:    user1.Email,
            Username: "different",
            Name:     "Different Name",
        }

        _, err = svc.Create(context.Background(), dto)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "email already exists")
    })
}
```

## Handler Tests

```go
// internal/handlers/user_handler_test.go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "myapp/internal/handlers"
    "myapp/test/mocks"
)

func TestUserHandler_Create(t *testing.T) {
    // Setup
    mockService := new(mocks.UserService)
    handler := handlers.NewUserHandler(mockService)

    router := chi.NewRouter()
    router.Post("/users", handler.Create)

    t.Run("Valid Request", func(t *testing.T) {
        // Prepare request
        reqBody := map[string]string{
            "email": "test@example.com",
            "name":  "Test User",
        }
        body, _ := json.Marshal(reqBody)

        req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        rr := httptest.NewRecorder()

        // Setup mock
        mockService.On("Create", mock.Anything, mock.Anything).
            Return(&User{ID: "123", Email: reqBody["email"]}, nil)

        // Execute
        router.ServeHTTP(rr, req)

        // Assert
        assert.Equal(t, http.StatusCreated, rr.Code)

        var response map[string]interface{}
        err := json.Unmarshal(rr.Body.Bytes(), &response)
        require.NoError(t, err)
        assert.Equal(t, "123", response["id"])
    })

    t.Run("Invalid Request", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/users", bytes.NewBufferString("invalid json"))
        req.Header.Set("Content-Type", "application/json")
        rr := httptest.NewRecorder()

        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusBadRequest, rr.Code)
    })
}
```

## E2E Tests with Hurl

```hurl
# tests/e2e/api/auth.hurl
# Authentication Flow Test

# Register new user
POST http://localhost:8080/register
Content-Type: application/x-www-form-urlencoded
```text

email=test@example.com&name=Test User

```text
HTTP 303
[Captures]
location: header "Location"

# Request OTP
POST http://localhost:8080/login
Content-Type: application/x-www-form-urlencoded
```text

email=test@example.com

```text
HTTP 303
[Asserts]
header "Location" == "/verify-otp"

# Verify OTP (would need to mock or extract from logs)
POST http://localhost:8080/verify-otp
Content-Type: application/x-www-form-urlencoded
```text

email=test@example.com&code=123456

```text
HTTP 303
[Asserts]
header "Location" == "/dashboard"
[Captures]
session: cookie "session"

# Access protected route
GET http://localhost:8080/dashboard
Cookie: session={{session}}
HTTP 200
[Asserts]
body contains "Welcome"
```

## Test Commands

```makefile
# Makefile
.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit:
	go test -v -short ./...

.PHONY: test-integration
test-integration:
	go test -v -tags=integration ./...

.PHONY: test-e2e
test-e2e:
	docker-compose up -d test-db
	go run cmd/server/main.go &
	sleep 5
	hurl --test tests/e2e/api/*.hurl
	docker-compose down

.PHONY: test-all
test-all: test-unit test-integration test-e2e

.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

## Benchmark Tests

```go
// internal/services/user_service_bench_test.go
package services_test

import (
    "context"
    "testing"

    "myapp/test/testutil"
)

func BenchmarkUserService_Create(b *testing.B) {
    db := testutil.SetupTestDB(b, "go-libsql")
    svc := setupUserService(db)
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dto := services.CreateUserDTO{
            Email:    fmt.Sprintf("bench%d@test.com", i),
            Username: fmt.Sprintf("bench%d", i),
            Name:     "Benchmark User",
        }
        svc.Create(ctx, dto)
    }
}
```

## Mock Generation

```go
// mocks/generate.go
//go:generate mockery --all --output . --outpkg mocks

package mocks

// Generate mocks with:
// go generate ./test/mocks
```

## Testing Best Practices

1. **Use table-driven tests** - Reduce boilerplate with test tables
2. **Test behavior, not implementation** - Focus on inputs/outputs
3. **Use factories for test data** - Consistent, maintainable fixtures
4. **Isolate tests with transactions** - Rollback after each test
5. **Mock external services** - Unit tests shouldn't require external services
6. **Use build tags for integration tests** - Keep unit tests fast
7. **Test edge cases** - Empty strings, nil values, large inputs
8. **Benchmark critical paths** - Identify performance regressions

## CI Configuration

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3

    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Unit Tests
      run: make test-unit

    - name: Integration Tests
      env:
        POSTGRES_TEST_URL: postgres://postgres:test@localhost/test?sslmode=disable
      run: make test-integration

    - name: Coverage
      run: make test-coverage

    - uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

## Next Steps

- Continue to [Code Generation →](./14_code_generation.md)
- Back to [← Observability](./12_observability.md)
- Return to [Summary](./0_summary.md)
