# Example Project Walkthrough

Learn by doing: implement a complete feature from start to finish in a Tracks-generated application.

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

**Why:** Handlers convert HTTP requests/responses, render templ templates, and orchestrate services.

Tracks is a **hypermedia framework** - handlers render HTML templates, not JSON. HTMX enables partial page updates without full page reloads.

Create `internal/http/handlers/user.go`:

```go
package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/youruser/myapp/internal/domain/users"
	"github.com/youruser/myapp/internal/http/helpers"
	"github.com/youruser/myapp/internal/http/views/pages"
	"github.com/youruser/myapp/internal/interfaces"
)

type UserHandler struct {
	logger  interfaces.Logger
	service interfaces.UserService
}

func NewUserHandler(logger interfaces.Logger, service interfaces.UserService) *UserHandler {
	return &UserHandler{
		logger:  logger,
		service: service,
	}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	userList, err := h.service.List(r.Context())
	if err != nil {
		helpers.RenderError(w, r, http.StatusInternalServerError, "Failed to load users", h.logger)
		return
	}

	helpers.RenderPage(w, r, pages.UsersPage(userList), pages.UsersPagePartial(userList), h.logger)
}

func (h *UserHandler) New(w http.ResponseWriter, r *http.Request) {
	helpers.RenderPage(w, r, pages.UserNewPage(), pages.UserNewPagePartial(), h.logger)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		helpers.RenderError(w, r, http.StatusBadRequest, "Invalid form data", h.logger)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	user, err := h.service.Create(r.Context(), name, email)
	if err != nil {
		if errors.Is(err, users.ErrInvalidInput) {
			helpers.RenderError(w, r, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		helpers.RenderError(w, r, http.StatusInternalServerError, "Failed to create user", h.logger)
		return
	}

	if helpers.IsHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/users/"+user.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/users/"+user.ID, http.StatusSeeOther)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			helpers.RenderError(w, r, http.StatusNotFound, "User not found", h.logger)
			return
		}
		helpers.RenderError(w, r, http.StatusInternalServerError, "Failed to load user", h.logger)
		return
	}

	helpers.RenderPage(w, r, pages.UserDetailPage(user), pages.UserDetailPagePartial(user), h.logger)
}
```

**Key Points:**

- **Dependency injection** - Handler receives logger and service via constructor
- **Form parsing** - Use `r.ParseForm()` and `r.FormValue()` for HTML forms (not JSON)
- **Templ rendering** - Use `helpers.RenderPage()` with full page and partial variants
- **HTMX detection** - Check `helpers.IsHTMXRequest(r)` for partial vs full page
- **HTMX redirects** - Use `HX-Redirect` header instead of HTTP redirect for HTMX requests
- **Domain errors** - Map `ErrInvalidInput` and `ErrUserNotFound` to appropriate HTTP status

---

## Step 7: Register Routes

**Why:** Make the handlers accessible via HTTP endpoints.

**Update `internal/http/routes/routes.go`:**

```go
package routes

const (
	// Health
	HealthCheck = "/health"

	// Users
	UsersList   = "/users"
	UsersNew    = "/users/new"
	UsersCreate = "/users"
	UserGet     = "/users/{id}"
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
	userHandler := handlers.NewUserHandler(s.logger, s.userService)
	r.Get(routes.UsersList, userHandler.List)
	r.Get(routes.UsersNew, userHandler.New)
	r.Post(routes.UsersCreate, userHandler.Create)
	r.Get(routes.UserGet, userHandler.Get)
}
```

**Key Points:**

- **Web routes** - Use `/users` not `/api/users` (this is a hypermedia app, not a JSON API)
- **RESTful pattern** - `GET /users/new` for form, `POST /users` for creation
- **Route constants** - Type-safe URL references throughout the codebase

---

## Step 8: Create Templ Views

**Why:** Templ provides type-safe HTML templates that compile to Go code.

Tracks uses [templUI](https://templui.io) components for consistent styling. These components are automatically installed when you run `tracks new`.

Create `internal/http/views/pages/users.templ`:

```go
package pages

import (
	"github.com/youruser/myapp/internal/http/views/components"
	"github.com/youruser/myapp/internal/http/views/components/ui"
	"github.com/youruser/myapp/internal/http/views/layouts"
	"github.com/youruser/myapp/internal/interfaces"
)

templ usersContent(users []*interfaces.User) {
	<main class="container-app py-8">
		<div class="flex justify-between items-center mb-6">
			<h1 class="text-2xl font-bold">Users</h1>
			<a
				href="/users/new"
				hx-get="/users/new"
				hx-target="#content"
				hx-push-url="true"
			>
				@ui.Button(ui.ButtonProps{Variant: "primary"}) {
					Add User
				}
			</a>
		</div>

		<div id="users-list" class="space-y-4">
			if len(users) == 0 {
				@ui.Card(ui.CardProps{Class: "text-center py-8"}) {
					<p class="text-muted-foreground">No users yet. Create your first user!</p>
				}
			} else {
				for _, user := range users {
					@components.UserCard(user)
				}
			}
		</div>
	</main>
}

templ UsersPage(users []*interfaces.User) {
	@layouts.Base("Users", "Manage users") {
		@usersContent(users)
	}
}

templ UsersPagePartial(users []*interfaces.User) {
	@usersContent(users)
}
```

Create `internal/http/views/pages/user_new.templ`:

```go
package pages

import (
	"github.com/youruser/myapp/internal/http/views/components/ui"
	"github.com/youruser/myapp/internal/http/views/layouts"
)

templ userNewContent() {
	<main class="container-app py-8">
		<h1 class="text-2xl font-bold mb-6">Create User</h1>

		@ui.Card(ui.CardProps{Class: "max-w-lg"}) {
			@ui.CardContent() {
				<form hx-post="/users" hx-target="#content" hx-swap="innerHTML">
					<div class="space-y-4">
						<div>
							@ui.Label(ui.LabelProps{For: "name"}) {
								Name
							}
							@ui.Input(ui.InputProps{
								Type:        "text",
								ID:          "name",
								Name:        "name",
								Placeholder: "Enter full name",
								Required:    true,
							})
						</div>

						<div>
							@ui.Label(ui.LabelProps{For: "email"}) {
								Email
							}
							@ui.Input(ui.InputProps{
								Type:        "email",
								ID:          "email",
								Name:        "email",
								Placeholder: "user@example.com",
								Required:    true,
							})
						</div>

						<div class="flex gap-2 pt-4">
							@ui.Button(ui.ButtonProps{Variant: "primary", Type: "submit"}) {
								Create User
							}
							<a href="/users" hx-get="/users" hx-target="#content" hx-push-url="true">
								@ui.Button(ui.ButtonProps{Variant: "outline", Type: "button"}) {
									Cancel
								}
							</a>
						</div>
					</div>
				</form>
			}
		}
	</main>
}

templ UserNewPage() {
	@layouts.Base("Create User", "Create a new user") {
		@userNewContent()
	}
}

templ UserNewPagePartial() {
	@userNewContent()
}
```

Create `internal/http/views/pages/user_detail.templ`:

```go
package pages

import (
	"github.com/youruser/myapp/internal/http/views/components/ui"
	"github.com/youruser/myapp/internal/http/views/layouts"
	"github.com/youruser/myapp/internal/interfaces"
)

templ userDetailContent(user *interfaces.User) {
	<main class="container-app py-8">
		<div class="flex items-center gap-4 mb-6">
			<a href="/users" hx-get="/users" hx-target="#content" hx-push-url="true">
				@ui.Button(ui.ButtonProps{Variant: "outline", Size: "sm"}) {
					← Back
				}
			</a>
			<h1 class="text-2xl font-bold">User Details</h1>
		</div>

		@ui.Card(ui.CardProps{Class: "max-w-lg"}) {
			@ui.CardHeader() {
				<h2 class="text-xl font-semibold">{ user.Name }</h2>
			}
			@ui.CardContent() {
				<dl class="space-y-2">
					<div>
						<dt class="text-sm text-muted-foreground">Email</dt>
						<dd>{ user.Email }</dd>
					</div>
					<div>
						<dt class="text-sm text-muted-foreground">Created</dt>
						<dd>{ user.CreatedAt.Format("January 2, 2006") }</dd>
					</div>
				</dl>
			}
		}
	</main>
}

templ UserDetailPage(user *interfaces.User) {
	@layouts.Base(user.Name, "User details") {
		@userDetailContent(user)
	}
}

templ UserDetailPagePartial(user *interfaces.User) {
	@userDetailContent(user)
}
```

Create `internal/http/views/components/user_card.templ`:

```go
package components

import (
	"github.com/youruser/myapp/internal/http/views/components/ui"
	"github.com/youruser/myapp/internal/interfaces"
)

templ UserCard(user *interfaces.User) {
	@ui.Card(ui.CardProps{Class: "hover:shadow-md transition-shadow"}) {
		<a
			href={ templ.SafeURL("/users/" + user.ID) }
			hx-get={ "/users/" + user.ID }
			hx-target="#content"
			hx-push-url="true"
			class="block p-4"
		>
			<div class="flex justify-between items-center">
				<div>
					<h3 class="font-semibold">{ user.Name }</h3>
					<p class="text-sm text-muted-foreground">{ user.Email }</p>
				</div>
				<span class="text-sm text-muted-foreground">
					{ user.CreatedAt.Format("Jan 2, 2006") }
				</span>
			</div>
		</a>
	}
}
```

**Key Points:**

- **Full + Partial pattern** - Every page has `*Page()` (with layout) and `*PagePartial()` (content only)
- **HTMX attributes** - `hx-get`, `hx-post`, `hx-target="#content"`, `hx-push-url="true"`
- **templUI components** - `@ui.Button`, `@ui.Card`, `@ui.Input`, `@ui.Label` for consistent styling
- **Type safety** - Templates receive typed Go parameters (`*interfaces.User`, `[]*interfaces.User`)
- **Navigation** - Links use both `href` (for non-JS) and `hx-get` (for HTMX enhancement)

**Generate the Go code:**

```bash
make generate
```

This compiles `.templ` files into `.go` files that can be rendered by handlers.

---

## Step 9: Wire Dependencies

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

## Step 10: Generate Mocks

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

## Step 11: Write Tests

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

## Step 12: Verify It Works

**Start the server:**

```bash
make dev
```

**Open your browser:**

Navigate to `http://localhost:8080/users`

You should see:

- The Users list page with an "Add User" button
- An empty state message if no users exist yet

**Create a user:**

1. Click the **Add User** button
2. Fill in the form with a name and email
3. Click **Create User**
4. You'll be redirected to the user detail page

**Verify HTMX:**

Notice that navigation happens without full page reloads:

- The URL updates in the browser (`/users` → `/users/new` → `/users/{id}`)
- Only the `#content` area updates, not the entire page
- Browser back/forward buttons work correctly

**Test without JavaScript:**

Disable JavaScript in your browser and repeat the steps. The app should still work - HTMX is progressive enhancement, not a requirement.

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
- ✅ **Type safety** - SQLC for SQL, templ for HTML, route constants for URLs

### Hypermedia Patterns

- ✅ **Templ templates** - Type-safe HTML that compiles to Go code
- ✅ **HTMX enhancement** - Partial page updates without full reloads
- ✅ **Progressive enhancement** - Works without JavaScript enabled
- ✅ **templUI components** - Consistent, accessible UI components

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
