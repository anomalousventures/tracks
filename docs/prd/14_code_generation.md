# Code Generation

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks provides sophisticated code generation through an interactive TUI, generating idiomatic Go code that developers would write themselves. The system generates complete resources including models, repositories, services, handlers, and templates with zero magic.

## Goals

- Generate idiomatic Go code that developers would write themselves
- Interactive DTO generation with field selection and validation
- Service layer with dependency injection for easy testing
- Zero magic - all generated code is readable and debuggable
- Automatic route path helper generation
- Type-safe SQL query generation via SQLC

## User Stories

- As a developer, I want to generate a complete CRUD resource with one command
- As a developer, I want to select which fields to expose in DTOs
- As a developer, I want to customize validation rules for each field
- As a developer, I want generated services with injected dependencies for testing
- As a developer, I want type-safe route helper functions

## Interactive TUI Generators

The Tracks TUI provides interactive generators that guide you through the code generation process with live previews.

### Launching the TUI

```bash
# Launch interactive TUI
tracks

# Or specific generator
tracks generate
```

### Resource Generator UI

```text
┌─ Generate Resource ─────────────────────────────┐
│ Resource Name: Post                             │
│                                                  │
│ Fields:                                          │
│ ┌────────────────────────────────────────────┐ │
│ │ 1. title     [string    ▼] ☑required ☐null│ │
│ │ 2. slug      [string    ▼] ☑unique  ☐null │ │
│ │ 3. content   [text      ▼] ☑required ☐null│ │
│ │ 4. author    [relation  ▼] ☑required ☐null│ │
│ │ 5. published [bool      ▼] ☐required ☐null│ │
│ │                                             │ │
│ │ [+ Add Field]                               │ │
│ └────────────────────────────────────────────┘ │
│                                                  │
│ Options:                                         │
│ ☑ Generate repository                           │
│ ☑ Generate service                              │
│ ☑ Generate handler                              │
│ ☑ Generate views (templ)                        │
│ ☑ Generate tests                                │
│ ☑ Add to routes                                 │
│                                                  │
│ [Preview Code]  [Generate]  [Cancel]            │
└──────────────────────────────────────────────────┘
```

## Interactive DTO Generation

DTOs are generated interactively, allowing developers to select which fields to expose and configure validation rules.

```text
┌─ Generate DTO from Post model ─────────────────┐
│ Model: Post                                     │
│                                                  │
│ Select fields to expose:                        │
│                                                  │
│ ☐ id         (internal - never expose)         │
│ ☑ title      [required,min=1,max=200   ▼]     │
│ ☑ slug       [required,slug            ▼]     │
│ ☑ content    [required,min=1           ▼]     │
│ ☐ authorID   (internal reference)              │
│ ☑ published  [                         ▼]     │
│ ☐ createdAt  (auto-managed)                    │
│ ☐ updatedAt  (auto-managed)                    │
│                                                  │
│ DTO Type: [CreatePost ▼]                       │
│                                                  │
│ [Generate]  [Cancel]                           │
└──────────────────────────────────────────────────┘
```

### Generated DTO

```go
// internal/dto/post_dto.go
package dto

import (
    "time"
)

type CreatePostDTO struct {
    Title     string `json:"title" validate:"required,min=1,max=200"`
    Slug      string `json:"slug" validate:"required,slug"`
    Content   string `json:"content" validate:"required,min=1"`
    Published bool   `json:"published"`
}

type UpdatePostDTO struct {
    Title     *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
    Slug      *string `json:"slug,omitempty" validate:"omitempty,slug"`
    Content   *string `json:"content,omitempty" validate:"omitempty,min=1"`
    Published *bool   `json:"published,omitempty"`
}

type PostResponseDTO struct {
    ID         string    `json:"id"`
    Title      string    `json:"title"`
    Slug       string    `json:"slug"`
    Content    string    `json:"content"`
    Published  bool      `json:"published"`
    Author     AuthorDTO `json:"author"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type AuthorDTO struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Name     string `json:"name"`
}
```

## Service Layer Generation

Services follow a strict pattern with interfaces in `internal/interfaces/` to avoid import cycles.

### Interface Definition (internal/interfaces/post.go)

```go
// Generated by: tracks generate resource post
package interfaces

import "context"

type PostService interface {
    Create(ctx context.Context, userID string, dto CreatePostDTO) (*Post, error)
    GetBySlug(ctx context.Context, slug string) (*Post, error)
    Update(ctx context.Context, userID, slug string, dto UpdatePostDTO) (*Post, error)
}

type PostRepository interface {
    Create(ctx context.Context, post *Post) error
    GetBySlug(ctx context.Context, slug string) (*Post, error)
    Update(ctx context.Context, post *Post) error
}

// Domain types (also in interfaces package)
type Post struct {
    ID        string
    Title     string
    Slug      string
    Content   string
    AuthorID  string
    Published bool
}

type CreatePostDTO struct {
    Title     string
    Slug      string
    Content   string
    Published bool
}

type UpdatePostDTO struct {
    Title     *string
    Content   *string
    Published *bool
}
```

### Service Implementation (internal/domain/posts/service.go)

```go
package posts

import (
    "context"
    "fmt"

    "github.com/gofrs/uuid/v5"
    "github.com/gosimple/slug"
    "myapp/internal/interfaces"
)

type service struct {
    repo interfaces.PostRepository
}

func NewService(repo interfaces.PostRepository) interfaces.PostService {
    return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID string, dto interfaces.CreatePostDTO) (*interfaces.Post, error) {
    if dto.Slug == "" {
        dto.Slug = slug.Make(dto.Title)
    }

    if existing, _ := s.repo.GetBySlug(ctx, dto.Slug); existing != nil {
        dto.Slug = fmt.Sprintf("%s-%s", dto.Slug, uuid.Must(uuid.NewV4()).String()[:8])
    }

    post := &interfaces.Post{
        ID:        uuid.Must(uuid.NewV7()).String(),
        Title:     dto.Title,
        Slug:      dto.Slug,
        Content:   dto.Content,
        AuthorID:  userID,
        Published: dto.Published,
    }

    if err := s.repo.Create(ctx, post); err != nil {
        return nil, fmt.Errorf("create post: %w", err)
    }

    return post, nil
}

func (s *service) GetBySlug(ctx context.Context, slug string) (*interfaces.Post, error) {
    post, err := s.repo.GetBySlug(ctx, slug)
    if err != nil {
        return nil, fmt.Errorf("get post by slug: %w", err)
    }
    return post, nil
}
```

### Repository Implementation (internal/domain/posts/repository.go)

```go
package posts

import (
    "context"
    "database/sql"

    "myapp/db/generated"
    "myapp/internal/interfaces"
)

type repository struct {
    queries *generated.Queries
}

func NewRepository(db *sql.DB) interfaces.PostRepository {
    return &repository{
        queries: generated.New(db),
    }
}

func (r *repository) Create(ctx context.Context, post *interfaces.Post) error {
    return r.queries.CreatePost(ctx, generated.CreatePostParams{
        ID:        post.ID,
        Title:     post.Title,
        Slug:      post.Slug,
        Content:   post.Content,
        AuthorID:  post.AuthorID,
        Published: post.Published,
    })
}

func (r *repository) GetBySlug(ctx context.Context, slug string) (*interfaces.Post, error) {
    row, err := r.queries.GetPostBySlug(ctx, slug)
    if err != nil {
        return nil, err
    }

    return &interfaces.Post{
        ID:        row.ID,
        Title:     row.Title,
        Slug:      row.Slug,
        Content:   row.Content,
        AuthorID:  row.AuthorID,
        Published: row.Published,
    }, nil
}
```

### Incremental Updates

When generating a new resource, tracks:

1. ✅ Creates `internal/interfaces/post.go`
2. ✅ Creates `internal/domain/posts/` with `service.go`, `repository.go`
3. ✅ Creates `db/queries/posts.sql`
4. ✅ Creates `db/migrations/002_create_posts.sql`
5. ✅ Updates `cmd/server/main.go` (inserts at markers)
6. ✅ Runs `sqlc generate` (updates `db/generated/`)
7. ✅ Runs `mockery` (generates `test/mocks/post_service.go`)

## Repository Generation

Repositories are generated with SQLC for type-safe database access.

```sql
-- db/queries/posts.sql
-- name: CreatePost :one
INSERT INTO posts (
    id, title, slug, content, author_id, published, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetPostByID :one
SELECT p.*, u.username as author_username, u.name as author_name
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.id = ? LIMIT 1;

-- name: GetPostBySlug :one
SELECT p.*, u.username as author_username, u.name as author_name
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.slug = ? LIMIT 1;

-- name: UpdatePost :one
UPDATE posts SET
    title = ?,
    content = ?,
    published = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = ?;

-- name: ListPosts :many
SELECT p.*, u.username as author_username, u.name as author_name
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE published = true
ORDER BY created_at DESC
LIMIT ? OFFSET ?;
```

## Handler Generation

Handlers are generated with proper validation and error handling.

```go
// internal/handlers/post_handler.go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-playground/validator/v10"
    "myapp/internal/dto"
    "myapp/internal/services"
)

type PostHandler struct {
    postService *services.PostService
    validator   *validator.Validate
}

func NewPostHandler(postService *services.PostService) *PostHandler {
    return &PostHandler{
        postService: postService,
        validator:   validator.New(),
    }
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)

    var createDTO dto.CreatePostDTO
    if err := json.NewDecoder(r.Body).Decode(&createDTO); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.validator.Struct(createDTO); err != nil {
        renderValidationErrors(w, err)
        return
    }

    post, err := h.postService.Create(r.Context(), userID, createDTO)
    if err != nil {
        handleServiceError(w, err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(toResponseDTO(post))
}

func (h *PostHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, "slug")

    post, err := h.postService.GetBySlug(r.Context(), slug)
    if err != nil {
        if err == services.ErrNotFound {
            http.NotFound(w, r)
            return
        }
        handleServiceError(w, err)
        return
    }

    // For HTML response (default)
    if r.Header.Get("Accept") != "application/json" {
        views.PostPage(post).Render(r.Context(), w)
        return
    }

    // For JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(toResponseDTO(post))
}
```

## Route Helper Generation

Route helpers are automatically generated for type-safe URL building.

```go
// internal/routes/generated.go (auto-generated)
package routes

import "github.com/a-h/templ"

// Route constants
const (
    PostList   = "/posts"
    PostView   = "/posts/{slug}"
    PostNew    = "/posts/new"
    PostEdit   = "/posts/{slug}/edit"
    PostCreate = "/posts"
    PostUpdate = "/posts/{slug}"
    PostDelete = "/posts/{slug}"
)

// Path builders for parameterized routes
func PostViewPath(slug string) string {
    return "/posts/" + slug
}

func PostEditPath(slug string) string {
    return "/posts/" + slug + "/edit"
}

// Type-safe URL builders for templates
func PostViewURL(slug string) templ.SafeURL {
    return templ.URL("/posts/" + slug)
}

func PostEditURL(slug string) templ.SafeURL {
    return templ.URL("/posts/" + slug + "/edit")
}
```

## Template Generation

Templates are generated using templ with proper component composition.

```go
// internal/views/posts/list.templ
package posts

import (
    "myapp/internal/models"
    "myapp/internal/routes"
    "myapp/internal/views/components"
)

templ PostList(posts []*models.Post) {
    @components.Layout("Posts") {
        <div class="container mx-auto px-4 py-8">
            <div class="flex justify-between items-center mb-6">
                <h1 class="text-3xl font-bold">Posts</h1>
                <a href={ routes.PostNew } class="btn btn-primary">
                    New Post
                </a>
            </div>

            <div class="grid gap-4">
                for _, post := range posts {
                    @PostCard(post)
                }
            </div>
        </div>
    }
}

templ PostCard(post *models.Post) {
    <article class="border rounded-lg p-6 hover:shadow-lg transition">
        <h2 class="text-xl font-semibold mb-2">
            <a href={ routes.PostViewURL(post.Slug) } class="hover:text-blue-600">
                { post.Title }
            </a>
        </h2>
        <p class="text-gray-600 mb-4">{ truncate(post.Content, 150) }</p>
        <div class="flex justify-between text-sm text-gray-500">
            <span>By { post.Author.Name }</span>
            <time>{ post.CreatedAt.Format("Jan 2, 2006") }</time>
        </div>
    </article>
}
```

## Test Generation

Tests are automatically generated for services with mocks.

```go
// internal/services/post_service_test.go
package services_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "myapp/internal/dto"
    "myapp/internal/services"
    "myapp/test/mocks"
)

func TestPostService_Create(t *testing.T) {
    // Setup mocks
    mockRepo := new(mocks.PostRepository)
    mockCache := new(mocks.CacheService)
    mockEvents := new(mocks.EventPublisher)

    svc := services.NewPostService(mockRepo, mockCache, mockEvents)

    // Test data
    createDTO := dto.CreatePostDTO{
        Title:     "Test Post",
        Content:   "Test content",
        Published: true,
    }

    // Expectations
    mockRepo.On("GetBySlug", mock.Anything, mock.Anything).
        Return(nil, services.ErrNotFound)
    mockRepo.On("Create", mock.Anything, mock.Anything).
        Return(nil)
    mockCache.On("Delete", mock.Anything, "posts:list").
        Return(nil)
    mockEvents.On("Publish", mock.Anything, mock.Anything).
        Return(nil)

    // Execute
    post, err := svc.Create(context.Background(), "user-123", createDTO)

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, post.ID)
    assert.Equal(t, "test-post", post.Slug)

    // Verify mocks
    mockRepo.AssertExpectations(t)
    mockCache.AssertExpectations(t)
    mockEvents.AssertExpectations(t)
}
```

## CLI Commands

```bash
# Generate a complete resource
tracks generate resource post title:string content:text author:relation published:bool

# Generate individual components
tracks generate handler webhook
tracks generate service notification
tracks generate repository comment
tracks generate migration add_published_to_posts

# Generate from existing database table
tracks generate from-table posts

# Interactive mode (launches TUI)
tracks generate
```

## Best Practices

1. **Never expose internal IDs** - Use slugs or usernames for public access
2. **Generate DTOs for each operation** - Create, Update, Response
3. **Use dependency injection** - Makes testing much easier
4. **Generate comprehensive tests** - Unit tests with mocks by default
5. **Keep generated code readable** - Should look like hand-written code
6. **Use SQLC for queries** - Type-safe database access
7. **Generate route helpers** - Avoid hardcoded URLs

## Next Steps

- Continue to [MCP Server →](./15_mcp_server.md)
- Back to [← Testing](./13_testing.md)
- Return to [Summary](./0_summary.md)
