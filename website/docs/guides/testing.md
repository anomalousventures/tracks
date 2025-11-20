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
  github.com/youruser/yourproject/internal/interfaces:
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
    "github.com/youruser/yourproject/tests/mocks"
)

func TestUserService_GetByID(t *testing.T) {
    mockRepo := mocks.NewMockUserRepository(t)

    mockRepo.EXPECT().
        FindByID(mock.Anything, "123").
        Return(&interfaces.User{
            ID:   "123",
            Name: "John Doe",
        }, nil).
        Once()

    svc := NewService(mockRepo)

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

    db := setupTestDB(t)
    defer db.Close()

    tx, err := db.BeginTx(context.Background(), nil)
    require.NoError(t, err)
    defer tx.Rollback()

    repo := NewRepository(tx)

    user := &interfaces.User{
        ID:    uuid.New().String(),
        Name:  "Test User",
        Email: "test@example.com",
    }

    err = repo.Insert(context.Background(), user)
    assert.NoError(t, err)

    found, err := repo.FindByID(context.Background(), user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
    assert.Equal(t, user.Email, found.Email)
}

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)

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
            Create(mock.Anything, "johndoe", "john@example.com").
            Return(&interfaces.User{
                ID:       "123",
                Username: "johndoe",
                Email:    "john@example.com",
            }, nil).
            Once()

        handler := NewUserHandler(mockService)

        form := url.Values{}
        form.Add("username", "johndoe")
        form.Add("email", "john@example.com")
        req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

        w := httptest.NewRecorder()

        handler.Create(w, req)

        assert.Equal(t, http.StatusSeeOther, w.Code)
        assert.Equal(t, "/u/johndoe", w.Header().Get("Location"))
    })

    t.Run("invalid form data", func(t *testing.T) {
        mockService := mocks.NewMockUserService(t)
        handler := NewUserHandler(mockService)

        req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(""))
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

        form := url.Values{}
        form.Add("username", "johndoe")
        form.Add("email", "john@example.com")
        req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        w := httptest.NewRecorder()

        handler.Create(w, req)

        assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
    })
}
```

**Key Points:**

- Mock all service dependencies
- Use `httptest.NewRequest` and `httptest.NewRecorder`
- Test HTTP status codes
- Test request/response marshaling
- Don't test business logic (that's in service tests)

## Testing Templ Components

Tracks uses [templ](https://templ.guide) for type-safe HTML templating. Testing templ components focuses on **accessible queries** that find elements by semantic meaning (roles, labels, text content) rather than brittle selectors (IDs, classes).

For comprehensive templ testing guidance, see the [official templ testing guide](https://templ.guide/core-concepts/testing/).

### Testing Philosophy

- **Accessible queries** - Find elements by text, role, or semantic attributes
- **Maintainable tests** - Tests survive styling changes (class names, IDs)
- **User-centric** - Test what users see, not implementation details
- **goquery for parsing** - jQuery-like API for Go

### Testing Individual Components

Test components in isolation by rendering to a buffer and parsing with goquery.

**Example:** Testing a navigation component

```go
package components

import (
    "bytes"
    "testing"

    "github.com/PuerkitoBio/goquery"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "yourproject/internal/http/views/components"
)

func TestNav_Render(t *testing.T) {
    // Render component to buffer
    var buf bytes.Buffer
    err := components.Nav().Render(context.Background(), &buf)
    require.NoError(t, err)

    // Parse HTML with goquery
    doc, err := goquery.NewDocumentFromReader(&buf)
    require.NoError(t, err)

    // Use accessible queries - find by text content
    homeLink := doc.Find("a:contains('Home')").First()
    assert.Equal(t, 1, homeLink.Length(), "should have Home link")
    assert.Equal(t, "/", homeLink.AttrOr("href", ""), "Home link should point to /")

    aboutLink := doc.Find("a:contains('About')").First()
    assert.Equal(t, 1, aboutLink.Length(), "should have About link")
    assert.Equal(t, "/about", aboutLink.AttrOr("href", ""), "About link should point to /about")

    // Verify semantic structure
    nav := doc.Find("nav")
    assert.Equal(t, 1, nav.Length(), "should have nav element")
}
```

**Example:** Testing a meta component with table-driven tests

```go
func TestMeta_Render(t *testing.T) {
    tests := []struct {
        name        string
        title       string
        description string
        wantTitle   string
        wantDesc    string
    }{
        {
            name:        "renders all metadata",
            title:       "Test Page",
            description: "A test page description",
            wantTitle:   "Test Page",
            wantDesc:    "A test page description",
        },
        {
            name:        "handles empty description",
            title:       "Test Page",
            description: "",
            wantTitle:   "Test Page",
            wantDesc:    "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var buf bytes.Buffer
            err := components.Meta(tt.title, tt.description).Render(context.Background(), &buf)
            require.NoError(t, err)

            doc, err := goquery.NewDocumentFromReader(&buf)
            require.NoError(t, err)

            // Find title by element name
            title := doc.Find("title").Text()
            assert.Equal(t, tt.wantTitle, title)

            // Find meta description by attribute
            metaDesc := doc.Find("meta[name='description']").AttrOr("content", "")
            assert.Equal(t, tt.wantDesc, metaDesc)
        })
    }
}
```

### Testing Pages with Layout

Test full pages including base layout to verify complete rendering.

**Example:** Testing home page with layout

```go
func TestHomePage_Render(t *testing.T) {
    var buf bytes.Buffer
    err := pages.Home().Render(context.Background(), &buf)
    require.NoError(t, err)

    doc, err := goquery.NewDocumentFromReader(&buf)
    require.NoError(t, err)

    // Verify page title
    title := doc.Find("title").Text()
    assert.Contains(t, title, "Home")

    // Verify main heading by text content
    h1 := doc.Find("h1:contains('Welcome')").First()
    assert.Equal(t, 1, h1.Length(), "should have welcome heading")

    // Verify navigation is present
    homeLink := doc.Find("a:contains('Home')")
    assert.GreaterOrEqual(t, homeLink.Length(), 1, "should have Home link in nav")

    // Verify footer is present
    footer := doc.Find("footer")
    assert.Equal(t, 1, footer.Length(), "should have footer")
}
```

### Testing Error Pages

Use table-driven tests for multiple error scenarios.

**Example:** Testing 404 and 500 error pages

```go
func TestErrorPages_Render(t *testing.T) {
    tests := []struct {
        name           string
        statusCode     int
        wantHeading    string
        wantStatusText string
    }{
        {
            name:           "404 Not Found",
            statusCode:     404,
            wantHeading:    "404",
            wantStatusText: "Not Found",
        },
        {
            name:           "500 Internal Server Error",
            statusCode:     500,
            wantHeading:    "500",
            wantStatusText: "Internal Server Error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var buf bytes.Buffer
            err := pages.Error(tt.statusCode).Render(context.Background(), &buf)
            require.NoError(t, err)

            doc, err := goquery.NewDocumentFromReader(&buf)
            require.NoError(t, err)

            // Find heading by text content
            heading := doc.Find(fmt.Sprintf("h1:contains('%s')", tt.wantHeading)).First()
            assert.Equal(t, 1, heading.Length(), "should have error code heading")

            // Find status text
            assert.Contains(t, buf.String(), tt.wantStatusText, "should contain status text")
        })
    }
}
```

### Testing HTMX Partial Rendering

Test both full page and HTMX partial responses.

**Example:** Integration test with HTMX header detection

```go
func TestPages_HTMXPartials(t *testing.T) {
    tests := []struct {
        name            string
        path            string
        headers         map[string]string
        isHTMXPartial   bool
        wantContains    []string
        wantNotContains []string
    }{
        {
            name:         "full page render without HTMX header",
            path:         "/",
            wantContains: []string{"<html", "</html>", "<h1>Welcome"},
        },
        {
            name:            "partial render with HTMX header",
            path:            "/",
            headers:         map[string]string{"HX-Request": "true"},
            isHTMXPartial:   true,
            wantContains:    []string{"<h1>Welcome"},
            wantNotContains: []string{"<html", "</html>"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, tt.path, nil)

            // Add HTMX header if testing partial
            for key, value := range tt.headers {
                req.Header.Set(key, value)
            }

            rec := httptest.NewRecorder()
            handler.ServeHTTP(rec, req)

            body := rec.Body.String()

            // For partials, use string matching
            if tt.isHTMXPartial {
                for _, contains := range tt.wantContains {
                    assert.Contains(t, body, contains)
                }
                for _, notContains := range tt.wantNotContains {
                    assert.NotContains(t, body, notContains)
                }
                return
            }

            // For full pages, use goquery
            doc, err := goquery.NewDocumentFromReader(rec.Body)
            require.NoError(t, err)

            for _, contains := range tt.wantContains {
                assert.Contains(t, body, contains)
            }
        })
    }
}
```

### Accessible Query Patterns with goquery

**Good patterns (accessible, maintainable):**

```go
// Find by text content
doc.Find("a:contains('Home')")
doc.Find("h1:contains('Welcome')")
doc.Find("button:contains('Submit')")

// Find by semantic element
doc.Find("nav")
doc.Find("footer")
doc.Find("main")

// Find by attribute value
doc.Find("meta[name='description']")
doc.Find("a[href='/about']")
doc.Find("input[type='email']")

// Find by ARIA attributes (accessibility)
doc.Find("[aria-label='Main navigation']")
doc.Find("[role='alert']")

// Combining selectors
doc.Find("nav a:contains('Home')")  // Home link inside nav
```

**Bad patterns (brittle, implementation-dependent):**

```go
// ❌ Avoid - CSS class names change
doc.Find(".btn-primary")
doc.Find(".nav-link")

// ❌ Avoid - IDs are implementation details
doc.Find("#submit-button")
doc.Find("#main-nav")

// ❌ Avoid - Position-dependent selectors
doc.Find("div > div > a")  // Fragile to markup changes
```

### goquery API Reference

Common goquery methods for testing:

```go
// Selection
doc.Find("selector")              // Find by CSS selector
selection.First()                 // First matching element
selection.Last()                  // Last matching element
selection.Eq(index)               // Element at index

// Content
selection.Text()                  // Text content
selection.Html()                  // HTML content
selection.AttrOr("name", "default") // Attribute value with default

// Traversal
selection.Children()              // Direct children
selection.Parent()                // Parent element
selection.Siblings()              // Sibling elements

// Filtering
selection.Length()                // Number of elements
selection.Has("selector")         // Filter by descendant
selection.Filter("selector")      // Filter selection

// Assertions
assert.Equal(t, 1, selection.Length())
assert.Contains(t, selection.Text(), "expected")
assert.Equal(t, "value", selection.AttrOr("href", ""))
```

### Best Practices for Templ Testing

**DO:**

- ✅ Use accessible queries (text, semantic elements, ARIA)
- ✅ Test what users see, not implementation
- ✅ Use table-driven tests for multiple scenarios
- ✅ Test both full pages and HTMX partials
- ✅ Keep component tests focused and fast
- ✅ Verify semantic HTML structure

**DON'T:**

- ❌ Query by CSS class names (they change with styling)
- ❌ Query by element IDs (implementation details)
- ❌ Use position-dependent selectors (fragile)
- ❌ Test inline styles or exact HTML structure
- ❌ Test third-party component internals
- ❌ Skip error page testing

### Integration Tests for Pages

Full integration tests verify pages render correctly via HTTP.

**Example:** From `tests/integration/pages_test.go`

```go
func TestPages_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    logger := logging.NewLogger("test")
    cfg := &config.ServerConfig{Port: ":8080"}

    mockHealthService := mocks.NewMockHealthService(t)
    server := httpserver.NewServer(cfg, logger).
        WithHealthService(mockHealthService).
        RegisterRoutes()

    tests := []struct {
        name       string
        path       string
        wantStatus int
    }{
        {
            name:       "should render home page",
            path:       "/",
            wantStatus: http.StatusOK,
        },
        {
            name:       "should render about page",
            path:       "/about",
            wantStatus: http.StatusOK,
        },
        {
            name:       "should return 404 for nonexistent route",
            path:       "/nonexistent",
            wantStatus: http.StatusNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, tt.path, nil)
            rec := httptest.NewRecorder()

            server.ServeHTTP(rec, req)

            assert.Equal(t, tt.wantStatus, rec.Code)

            if rec.Code == http.StatusOK {
                doc, err := goquery.NewDocumentFromReader(rec.Body)
                require.NoError(t, err)

                // Verify common elements with accessible queries
                assert.GreaterOrEqual(t, doc.Find("a:contains('Home')").Length(), 1)
                assert.GreaterOrEqual(t, doc.Find("a:contains('About')").Length(), 1)
                assert.Equal(t, 1, doc.Find("title").Length())
            }
        })
    }
}
```

## Testing Cross-Domain Handlers

Handlers using multiple services need multiple mocks.

**Example:** Dashboard handler

```go
func TestDashboardHandler_Get(t *testing.T) {
    mockUserService := mocks.NewMockUserService(t)
    mockPostService := mocks.NewMockPostService(t)
    mockStatsService := mocks.NewMockStatsService(t)

    mockUserService.EXPECT().
        GetCurrent(mock.Anything).
        Return(&interfaces.User{ID: "123", Name: "John"}, nil).
        Once()

    mockPostService.EXPECT().
        ListByAuthor(mock.Anything, "123", 5).
        Return([]*interfaces.Post{
            {ID: "post1", Title: "Hello"},
        }, nil).
        Once()

    mockStatsService.EXPECT().
        GetForUser(mock.Anything, "123").
        Return(&interfaces.Stats{PostCount: 42}, nil).
        Once()

    handler := NewDashboardHandler(mockUserService, mockPostService, mockStatsService)

    req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
    w := httptest.NewRecorder()

    handler.Get(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
    assert.Contains(t, w.Body.String(), "John")
    assert.Contains(t, w.Body.String(), "Hello")
}
```

## Testing Middleware

Middleware tests verify the middleware chain behavior.

**Example:** Testing logging middleware

```go
func TestLoggingMiddleware(t *testing.T) {
    mockLogger := mocks.NewMockLogger(t)

    mockLogger.EXPECT().
        Info("request started", mock.Anything).
        Once()

    mockLogger.EXPECT().
        Info("request completed", mock.Anything).
        Once()

    mw := NewLogging(mockLogger)

    called := false
    testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        called = true
        w.WriteHeader(http.StatusOK)
    })

    wrapped := mw(testHandler)

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

- [CLI: tracks new](../cli/new.mdx) - Creating projects
- [testify Documentation](https://github.com/stretchr/testify) - Assertion library
- [mockery Documentation](https://github.com/vektra/mockery) - Mock generation
- [templ Testing Guide](https://templ.guide/core-concepts/testing/) - Official templ testing docs
- [goquery Documentation](https://github.com/PuerkitoBio/goquery) - jQuery-like DOM parsing
