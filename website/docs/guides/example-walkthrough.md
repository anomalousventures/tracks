# Example Project Walkthrough

Learn by doing: implement a complete feature from start to finish in a Tracks-generated application.

:::warning Incomplete - Awaiting Templ Template Support

**This walkthrough is currently incomplete and shows JSON API patterns.**

Tracks is a **hypermedia server framework** using templ templates and HTMX, not a JSON API framework. This guide currently demonstrates:

- ❌ Handlers returning JSON responses (incorrect for Tracks)
- ❌ Missing templ template rendering
- ❌ Missing HTMX patterns for partial page updates

**This guide will be completely rewritten** once templ template generation is implemented. For now, it demonstrates the lower layers (interfaces, repositories, services, testing) which remain valid, but the HTTP handler layer is incorrect.

Use this guide to understand:

- ✅ Database migrations and SQLC
- ✅ Repository and service patterns
- ✅ Testing with mocks
- ⚠️ **Ignore the handler examples** - they will be replaced with templ/HTMX patterns

:::

## What You'll Build

In this tutorial, you'll add a complete **user management** feature to a Tracks application, implementing all layers from database schema to HTTP handlers. By the end, you'll have:

- ✅ User create, read, and list operations
- ✅ Database migrations and type-safe queries
- ✅ Complete test coverage at every layer
- ✅ Production-ready code following Tracks patterns

**Time:** ~45 minutes

## Prerequisites

Before starting, you should have:

1. Completed the [quickstart tutorial](../getting-started/quickstart.mdx)
2. A generated Tracks project (we'll use `myapp` with `go-libsql` driver)
3. Basic familiarity with Go and HTTP APIs

## Learning Objectives

By following this tutorial, you'll learn:

- How to add a new domain to a Tracks application
- The complete flow from database to HTTP handler
- How to write testable code with dependency injection
- Testing strategies for each layer (service, repository, handler)
- How to use generated mocks effectively

## Architecture Refresher

Tracks applications use a clean layered architecture:

```text
HTTP Request
    ↓
Handler (orchestration, HTTP concerns)
    ↓
Service (business logic)
    ↓
Repository (data access)
    ↓
Database (SQLC-generated queries)
```

Each layer has a single responsibility and communicates via interfaces. Let's build it step by step.

---

## Step 1: Define Interfaces

**Why:** Interfaces enable testing and decouple layers. We define them first in `internal/interfaces/`.

Create `internal/interfaces/user.go`:

```go
package interfaces

import (
	"context"
	"time"
)

//go:generate mockery --name=UserService --outpkg=mocks --output=../../tests/mocks
//go:generate mockery --name=UserRepository --outpkg=mocks --output=../../tests/mocks

// UserService defines business logic for user operations
type UserService interface {
	Create(ctx context.Context, name, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]*User, error)
}

// UserRepository defines data access for users
type UserRepository interface {
	Insert(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
}

// User represents a user in the system
type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}
```

**Key Points:**

- `//go:generate` directives tell mockery to generate test mocks
- Interfaces live in `internal/interfaces/` (not in implementation packages)
- Services accept context as the first parameter
- Return concrete types (`*User`), accept interfaces (`UserService`)

---

## Step 2: Create Database Migration

**Why:** Schema changes are versioned migrations, not manual SQL scripts.

Create `internal/db/migrations/sqlite/20250112000000_create_users.sql`:

```sql
-- +goose Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP TABLE IF EXISTS users;
```

**Run the migration:**

```bash
make db-migrate
```

**Expected output:**

```text
2025/01/12 10:00:00 OK   20250112000000_create_users.sql (15.2ms)
goose: successfully migrated database
```

**Verify:**

```bash
sqlite3 data/myapp.db ".schema users"
```

You should see the table schema.

---

## Step 3: Add SQLC Queries

**Why:** SQLC generates type-safe Go code from SQL queries. No raw SQL in application code.

Create `internal/db/queries/users.sql`:

```sql
-- name: GetUser :one
SELECT id, name, email, created_at
FROM users
WHERE id = ?;

-- name: ListUsers :many
SELECT id, name, email, created_at
FROM users
ORDER BY created_at DESC;

-- name: CreateUser :exec
INSERT INTO users (id, name, email, created_at)
VALUES (?, ?, ?, ?);
```

**Generate the code:**

```bash
make generate
```

**What happened:**

SQLC read your queries and generated:

- `internal/db/generated/users.sql.go` - Generated query functions
- Type-safe parameters and return types

**Check the generated code:**

```bash
head -n 30 internal/db/generated/users.sql.go
```

You'll see functions like `CreateUser(ctx, CreateUserParams)` and `GetUser(ctx, string)`.

---

## Step 4: Implement Repository

**Why:** Repository wraps SQLC-generated code, maps to domain types, and handles errors.

Create `internal/domain/users/repository.go`:

```go
package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/youruser/myapp/internal/db"
	"github.com/youruser/myapp/internal/interfaces"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

var _ interfaces.UserRepository = (*Repository)(nil)

type Repository struct {
	database *sql.DB
	queries  *db.Queries
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{
		database: database,
		queries:  db.New(database),
	}
}

func (r *Repository) Insert(ctx context.Context, user *interfaces.User) error {
	params := db.CreateUserParams{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	if err := r.queries.CreateUser(ctx, params); err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*interfaces.User, error) {
	row, err := r.queries.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return &interfaces.User{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *Repository) FindAll(ctx context.Context) ([]*interfaces.User, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	users := make([]*interfaces.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, &interfaces.User{
			ID:        row.ID,
			Name:      row.Name,
			Email:     row.Email,
			CreatedAt: row.CreatedAt,
		})
	}

	return users, nil
}
```

**Key Points:**

- `var _ interfaces.UserRepository = (*Repository)(nil)` - Compile-time interface check
- Wrap all errors with context using `fmt.Errorf(...: %w, err)`
- Convert `sql.ErrNoRows` to domain error `ErrUserNotFound`
- Map SQLC types to domain types (`interfaces.User`)

---

## Step 5: Implement Service

**Why:** Services contain business logic and validation. Handlers should be thin.

Create `internal/domain/users/service.go`:

```go
package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/youruser/myapp/internal/interfaces"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

var _ interfaces.UserService = (*Service)(nil)

type Service struct {
	repo interfaces.UserRepository
}

func NewService(repo interfaces.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, email string) (*interfaces.User, error) {
	if err := validateCreateInput(name, email); err != nil {
		return nil, err
	}

	user := &interfaces.User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     strings.ToLower(email),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Insert(ctx, user); err != nil {
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	return user, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*interfaces.User, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	return user, nil
}

func (s *Service) List(ctx context.Context) ([]*interfaces.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	return users, nil
}

func validateCreateInput(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	if !strings.Contains(email, "@") {
		return fmt.Errorf("%w: invalid email format", ErrInvalidInput)
	}

	return nil
}
```

**Key Points:**

- Service accepts `interfaces.UserRepository`, not concrete type
- All validation happens here (email format, required fields)
- Business rules live in services (e.g., normalize email to lowercase)
- UUIDs generated at service layer, not database layer

---

## Step 6: Add HTTP Handler

**Why:** Handlers convert HTTP requests/responses and orchestrate services.

:::caution This Section Will Be Replaced

**The code below shows JSON handlers, which is INCORRECT for Tracks.**

Tracks handlers should:

- Render **templ templates** returning HTML
- Accept **HTMX requests** for partial page updates
- Return **HTML fragments**, not JSON

This section will be completely rewritten once templ template generation is implemented. The patterns below (dependency injection, error handling) remain valid, but the response format is wrong.

:::

Create `internal/http/handlers/user.go`:

```go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/youruser/myapp/internal/domain/users"
	"github.com/youruser/myapp/internal/interfaces"
)

type UserHandler struct {
	userService interfaces.UserService
}

func NewUserHandler(userService interfaces.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Create(r.Context(), req.Name, req.Email)
	if err != nil {
		if errors.Is(err, users.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.List(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]UserResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
```

**Key Points:**

- DTOs (Data Transfer Objects) for request/response (`CreateUserRequest`, `UserResponse`)
- Domain errors mapped to HTTP status codes
- Generic errors return 500, specific errors return appropriate codes
- Handler is thin - just HTTP concerns, no business logic

---

## Step 7: Register Routes

**Why:** Make the handlers accessible via HTTP endpoints.

**Update `internal/http/routes/routes.go`:**

```go
package routes

const (
	// Health
	HealthCheck = "/api/health"

	// Users
	UsersList   = "/api/users"
	UsersCreate = "/api/users"
	UserGet     = "/api/users/{id}"
)
```

**Update `internal/http/routes.go`:**

```go
package http

import (
	"github.com/youruser/myapp/internal/http/handlers"
	"github.com/youruser/myapp/internal/http/routes"
)

func registerRoutes(s *Server) {
	r := s.router

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.NewLogging(s.logger))

	// Health
	healthHandler := handlers.NewHealthHandler(s.healthService)
	r.Get(routes.HealthCheck, healthHandler.Handle)

	// Users (NEW)
	userHandler := handlers.NewUserHandler(s.userService)
	r.Get(routes.UsersList, userHandler.List)
	r.Post(routes.UsersCreate, userHandler.Create)
	r.Get(routes.UserGet, userHandler.Get)
}
```

---

## Step 8: Wire Dependencies

**Why:** Dependency injection connects all the layers.

**Update `internal/http/server.go`:**

Add userService field:

```go
type Server struct {
	cfg    *config.ServerConfig
	logger interfaces.Logger
	router chi.Router

	healthService interfaces.HealthService
	userService   interfaces.UserService  // NEW
}
```

Add builder method:

```go
func (s *Server) WithUserService(svc interfaces.UserService) *Server {
	s.userService = svc
	return s
}
```

**Update `cmd/server/main.go`:**

Add repository and service initialization:

```go
// Repositories
healthRepo := health.NewRepository(database)
userRepo := users.NewRepository(database)  // NEW

// Services
healthService := health.NewService(healthRepo)
userService := users.NewService(userRepo)  // NEW

// Server
srv := http.NewServer(&cfg.Server, logger).
	WithHealthService(healthService).
	WithUserService(userService).  // NEW
	RegisterRoutes()
```

---

## Step 9: Generate Mocks

**Why:** Mocks enable testing services and handlers without real dependencies.

```bash
make generate-mocks
```

**Expected output:**

```text
2025-01-12T10:05:00.000 INF adding interface to collection collection=tests/mocks/mock_UserService.go
2025-01-12T10:05:00.000 INF adding interface to collection collection=tests/mocks/mock_UserRepository.go
```

**Verify:**

```bash
ls tests/mocks/ | grep -i user
```

You should see:

- `mock_UserRepository.go`
- `mock_UserService.go`

---

## Step 10: Write Tests

### Repository Tests

Create `internal/domain/users/repository_test.go`:

```go
package users_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/youruser/myapp/internal/domain/users"
	"github.com/youruser/myapp/internal/interfaces"

	_ "github.com/tursodatabase/go-libsql"
)

func TestRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := users.NewRepository(db)
	ctx := context.Background()

	user := &interfaces.User{
		ID:        uuid.New().String(),
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Insert(ctx, user)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Name, found.Name)
	assert.Equal(t, user.Email, found.Email)
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := users.NewRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, users.ErrUserNotFound)
}

func TestRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := users.NewRepository(db)
	ctx := context.Background()

	user1 := &interfaces.User{
		ID:        uuid.New().String(),
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC(),
	}
	user2 := &interfaces.User{
		ID:        uuid.New().String(),
		Name:      "Bob",
		Email:     "bob@example.com",
		CreatedAt: time.Now().UTC().Add(1 * time.Second),
	}

	require.NoError(t, repo.Insert(ctx, user1))
	require.NoError(t, repo.Insert(ctx, user2))

	users, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "Bob", users[0].Name)
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("libsql", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	return db
}
```

### Service Tests

Create `internal/domain/users/service_test.go`:

```go
package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/youruser/myapp/internal/domain/users"
	"github.com/youruser/myapp/internal/interfaces"
	"github.com/youruser/myapp/tests/mocks"
)

func TestService_Create(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	service := users.NewService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Insert", ctx, mock.AnythingOfType("*interfaces.User")).
		Return(nil)

	user, err := service.Create(ctx, "Alice", "alice@EXAMPLE.com")

	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, "alice@example.com", user.Email)
	mockRepo.AssertExpectations(t)
}

func TestService_Create_ValidationError(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	service := users.NewService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name  string
		email string
		want  error
	}{
		{"", "alice@example.com", users.ErrInvalidInput},
		{"Alice", "", users.ErrInvalidInput},
		{"Alice", "invalid", users.ErrInvalidInput},
	}

	for _, tt := range tests {
		_, err := service.Create(ctx, tt.name, tt.email)
		assert.ErrorIs(t, err, tt.want)
	}
}

func TestService_GetByID(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	service := users.NewService(mockRepo)
	ctx := context.Background()

	expected := &interfaces.User{
		ID:    "123",
		Name:  "Alice",
		Email: "alice@example.com",
	}

	mockRepo.On("FindByID", ctx, "123").Return(expected, nil)

	user, err := service.GetByID(ctx, "123")

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
	mockRepo.AssertExpectations(t)
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	service := users.NewService(mockRepo)
	ctx := context.Background()

	mockRepo.On("FindByID", ctx, "999").Return(nil, users.ErrUserNotFound)

	_, err := service.GetByID(ctx, "999")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, users.ErrUserNotFound))
	mockRepo.AssertExpectations(t)
}
```

### Handler Tests

Create `internal/http/handlers/user_test.go`:

```go
package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/youruser/myapp/internal/domain/users"
	"github.com/youruser/myapp/internal/http/handlers"
	"github.com/youruser/myapp/internal/interfaces"
	"github.com/youruser/myapp/tests/mocks"
)

func TestUserHandler_Create(t *testing.T) {
	mockService := mocks.NewUserService(t)
	handler := handlers.NewUserHandler(mockService)

	reqBody := `{"name":"Alice","email":"alice@example.com"}`
	mockUser := &interfaces.User{
		ID:    "123",
		Name:  "Alice",
		Email: "alice@example.com",
	}

	mockService.On("Create", mock.Anything, "Alice", "alice@example.com").
		Return(mockUser, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var resp handlers.UserResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "123", resp.ID)
	assert.Equal(t, "Alice", resp.Name)

	mockService.AssertExpectations(t)
}

func TestUserHandler_Create_ValidationError(t *testing.T) {
	mockService := mocks.NewUserService(t)
	handler := handlers.NewUserHandler(mockService)

	mockService.On("Create", mock.Anything, "", "alice@example.com").
		Return(nil, users.ErrInvalidInput)

	reqBody := `{"name":"","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Get(t *testing.T) {
	mockService := mocks.NewUserService(t)
	handler := handlers.NewUserHandler(mockService)

	mockUser := &interfaces.User{
		ID:    "123",
		Name:  "Alice",
		Email: "alice@example.com",
	}

	mockService.On("GetByID", mock.Anything, "123").Return(mockUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.UserResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "123", resp.ID)

	mockService.AssertExpectations(t)
}

func TestUserHandler_Get_NotFound(t *testing.T) {
	mockService := mocks.NewUserService(t)
	handler := handlers.NewUserHandler(mockService)

	mockService.On("GetByID", mock.Anything, "999").
		Return(nil, users.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Get(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}
```

**Run the tests:**

```bash
make test
```

**Expected output:**

```text
=== RUN   TestRepository_Insert
--- PASS: TestRepository_Insert (0.01s)
=== RUN   TestRepository_FindAll
--- PASS: TestRepository_FindAll (0.01s)
=== RUN   TestService_Create
--- PASS: TestService_Create (0.00s)
=== RUN   TestService_GetByID
--- PASS: TestService_GetByID (0.00s)
=== RUN   TestUserHandler_Create
--- PASS: TestUserHandler_Create (0.00s)
=== RUN   TestUserHandler_Get
--- PASS: TestUserHandler_Get (0.00s)
PASS
```

---

## Step 11: Verify It Works

**Start the server:**

```bash
make dev
```

**Create a user:**

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

**Response:**

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "Alice",
  "email": "alice@example.com",
  "created_at": "2025-01-12T15:30:45Z"
}
```

**List users:**

```bash
curl http://localhost:8080/api/users
```

**Get user by ID:**

```bash
curl http://localhost:8080/api/users/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

---

## What You Learned

Congratulations! You just implemented a complete feature across all layers. Here's what you learned:

### Architecture

- ✅ **Layered design** - HTTP → Service → Repository → Database
- ✅ **Dependency injection** - Services receive dependencies via constructors
- ✅ **Interface-based** - Layers communicate via interfaces, not concrete types

### Development Workflow

- ✅ **Define interfaces first** - Enables parallel development and testing
- ✅ **Database migrations** - Schema changes are versioned and reversible
- ✅ **SQLC code generation** - Type-safe SQL queries without ORM magic
- ✅ **Test-driven development** - Mocks enable testing each layer independently

### Testing Strategies

- ✅ **Repository tests** - Use in-memory database for integration tests
- ✅ **Service tests** - Mock repository to test business logic in isolation
- ✅ **Handler tests** - Mock service to test HTTP concerns separately

### Best Practices

- ✅ **Error wrapping** - Preserve error chain with `fmt.Errorf(...: %w, err)`
- ✅ **Domain errors** - Convert technical errors to domain-specific errors
- ✅ **Validation** - Business rules live in services, not handlers
- ✅ **Type safety** - SQLC for SQL, route constants for URLs, DTOs for HTTP

---

## Next Steps

Now that you understand the complete workflow, you can:

1. **Add more operations** - Update, delete, search users
2. **Add another domain** - Posts, comments, profiles
3. **Add authentication** - Middleware for auth checks
4. **Add relationships** - Users have many posts

## See Also

- [**Common Patterns**](./patterns.md) - Reference guide for common tasks
- [**Layer Guide**](./layer-guide.md) - Deep dive on each layer
- [**Testing Guide**](./testing.md) - Advanced testing strategies
- [**Architecture Overview**](./architecture-overview.md) - Core principles
