# ADR-005: HTTP Layer Architecture and Cross-Domain Orchestration

**Status:** Accepted
**Date:** 2025-11-01
**Context:** Epic 3 - Project Generation

## Context

During Epic 3 planning for project generation, we encountered a critical architectural question about handlers that need to orchestrate services from multiple domains (e.g., a dashboard handler using user, post, and notification services). This prompted a comprehensive review of our HTTP layer architecture to establish clear patterns for the code generator.

When designing the HTTP layer for generated Tracks applications, we faced a critical architectural question: **Where should HTTP handlers live, and how should they interact with domain services?**

### The Problem

Generated applications need to support both simple single-domain handlers and complex cross-domain orchestration handlers. Consider these examples:

**Simple Handler (Single Domain):**

```go
// User profile page - only needs user service
func (h *ProfileHandler) Show(w http.ResponseWriter, r *http.Request) {
    user, _ := h.userService.GetByID(ctx, userID)
    pages.Profile(user).Render(ctx, w)
}
```

**Complex Handler (Multiple Domains):**

```go
// Dashboard page - needs user, post, and notification services
func (h *DashboardHandler) Show(w http.ResponseWriter, r *http.Request) {
    user, _ := h.userService.GetByID(ctx, userID)
    posts, _ := h.postService.GetRecentByUser(ctx, userID, 10)
    notifs, _ := h.notifService.GetUnread(ctx, userID)
    pages.Dashboard(user, posts, notifs).Render(ctx, w)
}
```

The architectural question: **Should handlers be colocated with domains, or should they live in a separate HTTP layer?**

### Additional Concerns

1. **Interface Placement:** Where should service interfaces live to avoid import cycles?
2. **Server Organization:** Where should server.go, routes.go, and middleware live?
3. **Template Location:** Should templates be colocated with handlers or separate?
4. **Testing:** How do we make handlers easily testable with mocked services?

## Decision

We adopt an **HTTP-layer grouped architecture** where all web-facing code lives under `internal/http/`, and handlers orchestrate domain services via interfaces from `internal/interfaces/`.

### Key Decisions

1. **HTTP Layer Structure:** All HTTP concerns live under `internal/http/`

   ```text
   internal/http/
   ├── server.go          # HTTP server setup and dependency injection
   ├── routes.go          # Route registration and middleware chain
   ├── handlers/          # HTTP request handlers
   ├── middleware/        # HTTP middleware (auth, security, logging, etc.)
   └── views/             # templ templates (layouts, pages, components)
   ```

2. **Handler Location:** `internal/http/handlers/` (HTTP-layer grouped, not domain-colocated)

3. **Cross-Domain Orchestration:** Handlers can import and use services from multiple domains via interfaces

   ```go
   // internal/http/handlers/dashboard_handler.go
   import "myapp/internal/interfaces"

   type DashboardHandler struct {
       userService   interfaces.UserService
       postService   interfaces.PostService
       notifService  interfaces.NotificationService
   }
   ```

4. **Interface Principle:** "Accept interfaces, return structs"
   - Handlers accept service interfaces (for testability)
   - Services return concrete types (no interface pollution)

5. **Interface Placement:** All interfaces in `internal/interfaces/` (never colocated with implementations)

   ```go
   // ✅ CORRECT
   internal/interfaces/
   ├── user_service.go           # UserService interface
   ├── post_service.go           # PostService interface
   └── notification_service.go   # NotificationService interface

   internal/domain/user/
   └── service.go                # implements interfaces.UserService

   // ❌ WRONG - interfaces colocated with implementations
   internal/domain/user/
   ├── service_interface.go      # DON'T DO THIS
   └── service.go
   ```

### Complete Architecture Example

```go
// ✅ CORRECT: HTTP Layer Architecture

// internal/interfaces/user_service.go
package interfaces

type UserService interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

// internal/domain/user/service.go
package user

import "myapp/internal/interfaces"

type Service struct {
    repo Repository
}

// Ensure Service implements interfaces.UserService
var _ interfaces.UserService = (*Service)(nil)

func (s *Service) GetByID(ctx context.Context, id string) (*domain.User, error) {
    return s.repo.GetByID(ctx, id)
}

// internal/http/server.go
package http

import "myapp/internal/interfaces"

type Server struct {
    userService   interfaces.UserService
    postService   interfaces.PostService
    notifService  interfaces.NotificationService
}

func NewServer(
    userService interfaces.UserService,
    postService interfaces.PostService,
    notifService interfaces.NotificationService,
) *Server {
    return &Server{
        userService:  userService,
        postService:  postService,
        notifService: notifService,
    }
}

// internal/http/handlers/dashboard_handler.go
package handlers

import "myapp/internal/interfaces"

type DashboardHandler struct {
    userService   interfaces.UserService
    postService   interfaces.PostService
    notifService  interfaces.NotificationService
}

func NewDashboardHandler(
    userService interfaces.UserService,
    postService interfaces.PostService,
    notifService interfaces.NotificationService,
) *DashboardHandler {
    return &DashboardHandler{
        userService:  userService,
        postService:  postService,
        notifService: notifService,
    }
}

func (h *DashboardHandler) Show(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := middleware.GetUserID(ctx)

    // Orchestrate multiple domain services
    user, err := h.userService.GetByID(ctx, userID)
    if err != nil {
        // handle error
        return
    }

    posts, err := h.postService.GetRecentByUser(ctx, userID, 10)
    if err != nil {
        // handle error
        return
    }

    notifs, err := h.notifService.GetUnread(ctx, userID)
    if err != nil {
        // handle error
        return
    }

    // Render with templ
    pages.Dashboard(user, posts, notifs).Render(ctx, w)
}
```

### Testing with Interfaces

```go
// ✅ CORRECT: Easy to test with mocks

func TestDashboardHandler_Show(t *testing.T) {
    // Create mocks using mockery
    userService := mocks.NewUserService(t)
    postService := mocks.NewPostService(t)
    notifService := mocks.NewNotificationService(t)

    // Setup expectations
    userService.EXPECT().GetByID(mock.Anything, "user-123").
        Return(&domain.User{ID: "user-123"}, nil)

    postService.EXPECT().GetRecentByUser(mock.Anything, "user-123", 10).
        Return([]*domain.Post{}, nil)

    notifService.EXPECT().GetUnread(mock.Anything, "user-123").
        Return([]*domain.Notification{}, nil)

    // Create handler with mocks
    handler := NewDashboardHandler(userService, postService, notifService)

    // Test handler
    req := httptest.NewRequest("GET", "/dashboard", nil)
    req = req.WithContext(middleware.WithUserID(req.Context(), "user-123"))
    rr := httptest.NewRecorder()

    handler.Show(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
}
```

## Consequences

### Positive

- **Clean Separation:** HTTP concerns are isolated from domain logic
- **Cross-Domain Support:** Handlers can easily orchestrate multiple domains
- **Testability:** Handlers are trivial to test with mocked interfaces
- **No Import Cycles:** Interfaces in `internal/interfaces/` prevent circular dependencies
- **Explicit Dependencies:** Constructor injection makes dependencies visible
- **Single Responsibility:** Each layer has a clear, focused purpose
- **Easy Navigation:** All HTTP code is in one place (`internal/http/`)
- **Type Safety:** Compile-time verification that services implement interfaces

### Negative

- **More Files:** Separate files for interfaces vs implementations
- **Indirection:** Extra interface layer between handlers and services
- **Potential for Fat Handlers:** Handlers that orchestrate many services can become large
- **Learning Curve:** Developers must understand the interface placement pattern

### Neutral

- **Convention Required:** Need clear naming conventions (e.g., `UserService` interface, `user.Service` implementation)
- **DI Boilerplate:** Constructor injection creates boilerplate (but improves testability)
- **Mock Generation:** Requires mockery to generate test mocks (automated via `make generate-mocks`)

## Implementation Rules

### Rule 1: Accept Interfaces, Return Structs

Handlers and services should accept interfaces as dependencies but return concrete types.

```go
// ✅ CORRECT
func NewDashboardHandler(
    userService interfaces.UserService,  // Accept interface
) *DashboardHandler {                     // Return struct
    return &DashboardHandler{userService: userService}
}

func (s *Service) GetUser(ctx context.Context, id string) (*domain.User, error) {
    return s.repo.GetByID(ctx, id)  // Return concrete type
}

// ❌ WRONG - Don't return interfaces
func NewDashboardHandler(...) interfaces.DashboardHandler { // DON'T DO THIS
    return &DashboardHandler{...}
}
```

### Rule 2: Interfaces in Consumer Package

Interfaces should be defined where they are consumed, not where they are implemented.

```go
// ✅ CORRECT - Interface where it's used (HTTP layer)
// internal/interfaces/user_service.go
package interfaces

type UserService interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
}

// internal/domain/user/service.go - Implementation
package user

func (s *Service) GetByID(ctx context.Context, id string) (*domain.User, error) {
    // implementation
}

// ❌ WRONG - Interface colocated with implementation
// internal/domain/user/service_interface.go - DON'T DO THIS
package user

type ServiceInterface interface { // DON'T DO THIS
    GetByID(ctx context.Context, id string) (*domain.User, error)
}
```

### Rule 3: Explicit Interface Compliance

Services should explicitly declare interface compliance.

```go
// ✅ CORRECT - Compile-time check
package user

import "myapp/internal/interfaces"

type Service struct {
    repo Repository
}

// Verify Service implements UserService at compile time
var _ interfaces.UserService = (*Service)(nil)

// ❌ WRONG - No compliance check
package user

type Service struct {
    repo Repository
}
// Might not implement all interface methods!
```

## Alternatives Considered

### 1. Domain-Colocated Handlers

**Structure:**

```text
internal/domain/user/
├── service.go
├── repository.go
└── handler.go           # Handler lives with domain
```

**Rejected:** This creates tight coupling between HTTP concerns and domain logic. When a handler needs services from multiple domains (dashboard example), it becomes unclear which domain should "own" the handler. This also makes it harder to change HTTP frameworks or add other interfaces (GraphQL, gRPC) later.

### 2. Feature-Sliced Handlers

**Structure:**

```text
internal/features/dashboard/
├── handler.go
├── service.go
├── dependencies.go      # Aggregates services from multiple domains
└── template.templ
```

**Rejected:** This creates feature-level coupling and duplicates orchestration logic. Each feature would need its own service aggregation layer, leading to code duplication. Better to have a thin HTTP layer that orchestrates existing domain services.

### 3. Flat Handler Structure (No Hierarchy)

**Structure:**

```text
internal/handlers/
├── dashboard_handler.go
├── profile_handler.go
└── auth_handler.go
```

**Rejected:** This doesn't group related HTTP concerns together. Server, routes, and middleware would live in separate top-level packages (`internal/server/`, `internal/routes/`, `internal/middleware/`), making it harder to understand the HTTP layer as a cohesive unit.

## Implementation Notes

### Documentation Updates

All PRD documentation has been updated to reflect this architecture:

- PR #206: Updated 13 PRD files in `docs/prd/`
- Updated `CLAUDE.md` with layer responsibilities
- See commit: `docs: align HTTP layer architecture and improve validation tests`

### Epic 3 Integration

Epic 3 (Project Generation) will generate projects following this architecture:

- Server scaffolding with DI setup
- Handler templates with interface injection
- Middleware organization under `internal/http/middleware/`
- View templates under `internal/http/views/`

### Migration Path

For existing code following old patterns:

1. Move handlers from `internal/handlers/` → `internal/http/handlers/`
2. Move middleware from `internal/middleware/` → `internal/http/middleware/`
3. Move templates from `internal/templates/` → `internal/http/views/`
4. Create interfaces in `internal/interfaces/` for cross-domain dependencies
5. Update imports throughout codebase

## References

- [Clean Architecture by Robert Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Go Proverb: "Accept interfaces, return structs"](https://bryanftan.medium.com/accept-interfaces-return-structs-in-go-d4cab29a301b)
- [ADR-002: Interface Placement in Consumer Packages](./002-interface-placement-in-consumer-packages.md)
- [Epic 3: Project Generation](../roadmap/phases/phase-0/epic-3-project-generation.md)
- [PRD: Core Architecture](../prd/1_core_architecture.md)
- [PRD: Web Layer](../prd/5_web_layer.md)
- PR #206: HTTP Layer Architecture Alignment
