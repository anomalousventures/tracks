# Testing Guide

Learn how to test Tracks-generated applications at every layer.

## Testing Philosophy

Tracks-generated code is designed for testing:

- **Dependency injection** makes every component mockable
- **Interface-based design** decouples layers
- **Mockery integration** auto-generates mocks from interfaces
- **No global state** means tests are isolated

Every layer has its own testing strategy.

## Test Types

### Unit Tests

Fast, isolated tests with no external dependencies:

- Mock all dependencies
- Test business logic in isolation
- Run with `-short` flag
- Use the race detector (`-race`)

```bash
go test -v -race -short ./...
```

### Integration Tests

Test component integration without mocking:

- Real database (transaction rollback for cleanup)
- Test SQL queries work correctly
- Verify layers communicate properly
- Slower than unit tests

```bash
go test -v ./tests/integration/...
```

## Mock Generation

Tracks uses [mockery](https://github.com/vektra/mockery) to generate mocks from interfaces.

### Configuration

Mocks are configured in `.mockery.yaml`:

```yaml
with-expecter: true
outpkg: mocks
output: tests/mocks
packages:
  github.com/yourmodule/internal/interfaces:
    interfaces:
      UserService:
      UserRepository:
      HealthService:
      HealthRepository:
```

### Generating Mocks

```bash
# Generate all mocks
make generate-mocks

# Or use mockery directly
go tool mockery
```

This creates mocks in `tests/mocks/`:

```text
tests/mocks/
├── mock_UserService.go
├── mock_UserRepository.go
├── mock_HealthService.go
└── mock_HealthRepository.go
```

### Using Mocks in Tests

```go
import (
    "testing"
    "github.com/stretchr/testify/mock"
    "yourmodule/tests/mocks"
)

func TestUserService_GetByID(t *testing.T) {
    // Create mock
    mockRepo := mocks.NewMockUserRepository(t)

    // Set expectations
    mockRepo.EXPECT().
        FindByID(mock.Anything, "123").
        Return(&interfaces.User{
            ID:   "123",
            Name: "John Doe",
        }, nil).
        Once()

    // Create service with mock
    svc := NewService(mockRepo)

    // Test
    user, err := svc.GetByID(context.Background(), "123")
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
}
```

## Testing Services

Services contain business logic - mock the repository.

**Example:** Testing user creation with validation

```go
func TestService_Create(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        mockRepo := mocks.NewMockUserRepository(t)

        mockRepo.EXPECT().
            Insert(mock.Anything, mock.MatchedBy(func(u *interfaces.User) bool {
                return u.Name == "Jane Doe" && u.Email == "jane@example.com"
            })).
            Return(nil).
            Once()

        svc := NewService(mockRepo)
        user, err := svc.Create(context.Background(), "Jane Doe", "jane@example.com")

        assert.NoError(t, err)
        assert.NotEmpty(t, user.ID)
        assert.Equal(t, "Jane Doe", user.Name)
    })

    t.Run("validation error - empty name", func(t *testing.T) {
        mockRepo := mocks.NewMockUserRepository(t)

        svc := NewService(mockRepo)
        _, err := svc.Create(context.Background(), "", "jane@example.com")

        assert.ErrorIs(t, err, ErrInvalidInput)
    })

    t.Run("repository error", func(t *testing.T) {
        mockRepo := mocks.NewMockUserRepository(t)

        mockRepo.EXPECT().
            Insert(mock.Anything, mock.Anything).
            Return(errors.New("db error")).
            Once()

        svc := NewService(mockRepo)
        _, err := svc.Create(context.Background(), "Jane Doe", "jane@example.com")

        assert.Error(t, err)
        assert.Contains(t, err.Error(), "inserting user")
    })
}
```

**Key Points:**

- Test happy path and error cases
- Mock repository calls
- Verify business logic (validation, transformations)
- Don't test the database

## Testing Repositories

Repositories wrap SQLC-generated code - use integration tests with real database.

**Example:** Integration test for user repository

```go
func TestRepository_Insert(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Start transaction (rollback at end for cleanup)
    tx, err := db.BeginTx(context.Background(), nil)
    require.NoError(t, err)
    defer tx.Rollback()

    repo := NewRepository(tx)

    // Test insert
    user := &interfaces.User{
        ID:    uuid.New().String(),
        Name:  "Test User",
        Email: "test@example.com",
    }

    err = repo.Insert(context.Background(), user)
    assert.NoError(t, err)

    // Verify inserted
    found, err := repo.FindByID(context.Background(), user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
    assert.Equal(t, user.Email, found.Email)
}

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)

    // Run migrations
    migrationFiles := filepath.Glob("../../internal/db/migrations/*.sql")
    for _, file := range migrationFiles {
        sql, err := os.ReadFile(file)
        require.NoError(t, err)
        _, err = db.Exec(string(sql))
        require.NoError(t, err)
    }

    return db
}
```

**Key Points:**

- Use real database (in-memory SQLite or test container)
- Transaction rollback for cleanup
- Test SQLC queries work correctly
- Skip with `testing.Short()`

## Testing Handlers

Handlers orchestrate services - mock services, test HTTP concerns.

**Example:** Testing user creation handler

```go
func TestUserHandler_Create(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        mockService := mocks.NewMockUserService(t)

        mockService.EXPECT().
            Create(mock.Anything, "John Doe", "john@example.com").
            Return(&interfaces.User{
                ID:    "123",
                Name:  "John Doe",
                Email: "john@example.com",
            }, nil).
            Once()

        handler := NewUserHandler(mockService)

        // Create request
        reqBody := `{"name":"John Doe","email":"john@example.com"}`
        req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(reqBody))
        req.Header.Set("Content-Type", "application/json")

        // Record response
        w := httptest.NewRecorder()

        // Execute
        handler.Create(w, req)

        // Assert
        assert.Equal(t, http.StatusOK, w.Code)

        var resp map[string]interface{}
        err := json.NewDecoder(w.Body).Decode(&resp)
        assert.NoError(t, err)
        assert.Equal(t, "123", resp["id"])
        assert.Equal(t, "John Doe", resp["name"])
    })

    t.Run("invalid request body", func(t *testing.T) {
        mockService := mocks.NewMockUserService(t)
        handler := NewUserHandler(mockService)

        req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader("invalid json"))
        w := httptest.NewRecorder()

        handler.Create(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })

    t.Run("service error", func(t *testing.T) {
        mockService := mocks.NewMockUserService(t)

        mockService.EXPECT().
            Create(mock.Anything, mock.Anything, mock.Anything).
            Return(nil, errors.New("service error")).
            Once()

        handler := NewUserHandler(mockService)

        reqBody := `{"name":"John Doe","email":"john@example.com"}`
        req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(reqBody))
        w := httptest.NewRecorder()

        handler.Create(w, req)

        assert.Equal(t, http.StatusInternalServerError, w.Code)
    })
}
```

**Key Points:**

- Mock all service dependencies
- Use `httptest.NewRequest` and `httptest.NewRecorder`
- Test HTTP status codes
- Test request/response marshaling
- Don't test business logic (that's in service tests)

## Testing Cross-Domain Handlers

Handlers using multiple services need multiple mocks.

**Example:** Dashboard handler

```go
func TestDashboardHandler_Get(t *testing.T) {
    mockUserService := mocks.NewMockUserService(t)
    mockPostService := mocks.NewMockPostService(t)
    mockStatsService := mocks.NewMockStatsService(t)

    // Setup user expectation
    mockUserService.EXPECT().
        GetCurrent(mock.Anything).
        Return(&interfaces.User{ID: "123", Name: "John"}, nil).
        Once()

    // Setup posts expectation
    mockPostService.EXPECT().
        ListByAuthor(mock.Anything, "123", 5).
        Return([]*interfaces.Post{
            {ID: "post1", Title: "Hello"},
        }, nil).
        Once()

    // Setup stats expectation
    mockStatsService.EXPECT().
        GetForUser(mock.Anything, "123").
        Return(&interfaces.Stats{PostCount: 42}, nil).
        Once()

    handler := NewDashboardHandler(mockUserService, mockPostService, mockStatsService)

    req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
    w := httptest.NewRecorder()

    handler.Get(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var resp DashboardResponse
    err := json.NewDecoder(w.Body).Decode(&resp)
    assert.NoError(t, err)
    assert.Equal(t, "John", resp.User.Name)
    assert.Len(t, resp.Posts, 1)
    assert.Equal(t, 42, resp.Stats.PostCount)
}
```

## Testing Middleware

Middleware tests verify the middleware chain behavior.

**Example:** Testing logging middleware

```go
func TestLoggingMiddleware(t *testing.T) {
    // Create mock logger
    mockLogger := mocks.NewMockLogger(t)

    // Expect info logs
    mockLogger.EXPECT().
        Info("request started", mock.Anything).
        Once()

    mockLogger.EXPECT().
        Info("request completed", mock.Anything).
        Once()

    // Create middleware
    mw := NewLogging(mockLogger)

    // Create test handler
    called := false
    testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        called = true
        w.WriteHeader(http.StatusOK)
    })

    // Wrap handler
    wrapped := mw(testHandler)

    // Test
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    w := httptest.NewRecorder()

    wrapped.ServeHTTP(w, req)

    assert.True(t, called, "handler should have been called")
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## Test Structure

### Directory Layout

```text
internal/
├── domain/
│   └── users/
│       ├── service.go
│       ├── service_test.go       # Unit tests
│       ├── repository.go
│       └── repository_test.go    # Integration tests
├── http/
│   └── handlers/
│       ├── user.go
│       └── user_test.go          # Unit tests
tests/
├── mocks/                        # Generated mocks
│   ├── mock_UserService.go
│   └── mock_UserRepository.go
└── integration/                  # E2E integration tests
    └── users_test.go
```

### Test File Naming

- Unit tests: `*_test.go` next to source file
- Integration tests: `tests/integration/*_test.go`
- Build tag for slow tests: `//go:build integration`

### Test Function Naming

```go
func TestServiceName_MethodName(t *testing.T) {}
func TestHandlerName_HTTPMethod(t *testing.T) {}
func TestRepositoryName_MethodName(t *testing.T) {}
```

## Running Tests

### All Unit Tests

```bash
make test
# or
go test -v -race -short ./...
```

### Integration Tests

```bash
go test -v ./tests/integration/...
```

### Specific Package

```bash
go test -v ./internal/domain/users/...
```

### With Coverage

```bash
make test-coverage
```

### Watch Mode (with Air)

```bash
# In .air.toml
cmd = "go test -v ./..."
```

## Test Coverage

Generated projects include coverage reports:

```bash
make test-coverage
```

This generates:

- `coverage-unit.out` - Unit test coverage
- `coverage-integration.out` - Integration test coverage
- `coverage.html` - HTML coverage report

View coverage:

```bash
go tool cover -html=coverage.html
```

## Best Practices

### DO

- ✅ Write tests before implementation (TDD)
- ✅ Test happy path and error cases
- ✅ Use table-driven tests for multiple scenarios
- ✅ Mock all external dependencies
- ✅ Use `t.Helper()` in test helper functions
- ✅ Run tests with race detector
- ✅ Use integration tests for repositories
- ✅ Keep tests focused and isolated
- ✅ Regenerate mocks after interface changes

### DON'T

- ❌ Test implementation details
- ❌ Use real database in unit tests
- ❌ Skip error case testing
- ❌ Write flaky tests (time-dependent, order-dependent)
- ❌ Test third-party code (SQLC, standard library)
- ❌ Commit without running tests
- ❌ Ignore test failures
- ❌ Write tests without assertions

## Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {
            name:    "valid email",
            email:   "user@example.com",
            wantErr: false,
        },
        {
            name:    "missing @",
            email:   "userexample.com",
            wantErr: true,
        },
        {
            name:    "missing domain",
            email:   "user@",
            wantErr: true,
        },
        {
            name:    "empty",
            email:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Next Steps

- [**Architecture Overview**](./architecture-overview.md) - Core principles
- [**Layer Guide**](./layer-guide.md) - Deep dive on each layer
- [**Patterns**](./patterns.md) - Common patterns for extending

## See Also

- [CLI: tracks new](../cli/new.md) - Creating projects
- [testify Documentation](https://github.com/stretchr/testify) - Assertion library
- [mockery Documentation](https://github.com/vektra/mockery) - Mock generation
