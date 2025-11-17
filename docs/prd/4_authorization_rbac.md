# Authorization & RBAC

**[← Back to Summary](./0_summary.md)**

## Overview

Authorization in Tracks is built on **Casbin**, a powerful and flexible authorization library that supports multiple access control models. Casbin is a **framework choice** - it's built into Tracks and not configurable. The system implements Role-Based Access Control (RBAC) with database persistence, audit trails, and seamless integration with templates and middleware.

## Goals

- Permission-based access control using Casbin (Framework Choice)
- Roles aggregate permissions for easier management
- Middleware enforces permissions at the route level
- Template helpers show/hide UI based on permissions
- Audit trail for permission changes
- Default roles and permissions for common patterns

## User Stories

- As a developer, I want permissions checked automatically by middleware
- As an admin, I want to assign roles to users without coding
- As a user, I only want to see UI elements I have permission to use
- As a developer, I want to add new permissions easily
- As a security engineer, I want an audit trail of permission changes
- As a developer, I want RBAC built in from day one, not retrofitted

## Design Principles

### Using Casbin for Authorization

- Casbin is a mature, battle-tested authorization library (Framework Choice)
- Supports RBAC, ABAC, ACL and custom models
- Efficient permission checking with built-in caching
- Database persistence via adapters
- Policies: `user, resource, action` (e.g., `alice, posts, create`)
- Roles: `user, role` (e.g., `alice, editor`)

### Permissions Drive Everything

- Check permissions (`user, resource, action`), not roles, in code
- Roles are collections of permissions for convenience
- Users can have multiple roles
- Casbin handles the enforcement logic

## Database Schema

Casbin stores policies in a simple table structure. We extend it with audit trails.

```sql
-- migrations/2024_01_15_14_40_create_rbac.sql

-- Casbin policy table (compatible with all drivers)
CREATE TABLE casbin_rule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ptype TEXT NOT NULL,           -- p (policy) or g (grouping/role)
    v0 TEXT NOT NULL,              -- subject (user_id or role)
    v1 TEXT,                       -- object (resource)
    v2 TEXT,                       -- action
    v3 TEXT,
    v4 TEXT,
    v5 TEXT
);

CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype);
CREATE INDEX idx_casbin_rule_v0 ON casbin_rule(v0);
CREATE INDEX idx_casbin_rule_v1 ON casbin_rule(v1);

-- Audit trail for permission changes
CREATE TABLE permission_audit (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    action TEXT NOT NULL,                -- role_granted, role_revoked, permission_granted, permission_revoked
    subject TEXT NOT NULL,               -- user_id or role being modified
    object TEXT,                         -- resource (e.g., "posts")
    action_type TEXT,                    -- action (e.g., "create")
    role TEXT,                           -- role name if role operation
    performed_by TEXT NOT NULL,          -- user_id who made change
    performed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reason TEXT,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (performed_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_permission_audit_user ON permission_audit(user_id);
CREATE INDEX idx_permission_audit_performed_at ON permission_audit(performed_at);
CREATE INDEX idx_permission_audit_subject ON permission_audit(subject);

-- Optional: roles table for metadata (not required by Casbin)
CREATE TABLE roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### PostgreSQL Variant

```sql
-- For PostgreSQL deployments
CREATE TABLE casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype TEXT NOT NULL,
    v0 TEXT NOT NULL,
    v1 TEXT,
    v2 TEXT,
    v3 TEXT,
    v4 TEXT,
    v5 TEXT
);

-- Rest of schema similar with TIMESTAMPTZ instead of TIMESTAMP
```

## Casbin Model Configuration

```ini
# internal/rbac/model.conf
# RBAC model for Casbin

[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

## Casbin Setup

```go
// internal/rbac/enforcer.go
package rbac

import (
    "database/sql"
    "embed"

    "github.com/casbin/casbin/v2"
    sqladapter "github.com/casbin/casbin/v2/persist/sql-adapter"
)

//go:embed model.conf
var modelFS embed.FS

// NewEnforcer creates a Casbin enforcer with database adapter
func NewEnforcer(db *sql.DB, driverName string) (*casbin.Enforcer, error) {
    // Create adapter for database persistence
    adapter, err := sqladapter.NewAdapter(driverName, db)
    if err != nil {
        return nil, err
    }

    // Read model from embedded file
    modelData, err := modelFS.ReadFile("model.conf")
    if err != nil {
        return nil, err
    }

    // Create enforcer
    enforcer, err := casbin.NewEnforcer()
    if err != nil {
        return nil, err
    }

    // Load model
    if err := enforcer.InitWithModelAndAdapter(string(modelData), adapter); err != nil {
        return nil, err
    }

    // Load policies from database
    if err := enforcer.LoadPolicy(); err != nil {
        return nil, err
    }

    // Enable auto-save (persists changes immediately)
    enforcer.EnableAutoSave(true)

    return enforcer, nil
}
```

## Permission Constants

```go
// internal/rbac/permissions.go
package rbac

// Resources
const (
    ResourcePosts    = "posts"
    ResourceUsers    = "users"
    ResourceComments = "comments"
    ResourceAdmin    = "admin"
)

// Actions
const (
    ActionCreate    = "create"
    ActionRead      = "read"
    ActionUpdate    = "update"
    ActionDelete    = "delete"
    ActionPublish   = "publish"
    ActionModerate  = "moderate"
    ActionAccess    = "access"
    ActionManage    = "manage"
)

// Default roles
const (
    RoleUser      = "user"
    RoleEditor    = "editor"
    RoleModerator = "moderator"
    RoleAdmin     = "admin"
)
```

## Permission Service

```go
// internal/domain/permissions/service.go
package permissions

import (
    "context"
    "time"

    "github.com/casbin/casbin/v2"
    "myapp/internal/interfaces"
    "github.com/gofrs/uuid/v5"  // Fixed: Using correct UUID package for UUIDv7
    "myapp/internal/rbac"
)

type PermissionService struct {
    enforcer *casbin.Enforcer
    auditRepo AuditRepository
}

func NewPermissionService(enforcer *casbin.Enforcer, auditRepo AuditRepository) *PermissionService {
    return &PermissionService{
        enforcer:  enforcer,
        auditRepo: auditRepo,
    }
}

// Can checks if user has permission to perform action on resource
// This is the primary method used throughout the application
func (s *PermissionService) Can(ctx context.Context, userID, resource, action string) (bool, error) {
    ok, err := s.enforcer.Enforce(userID, resource, action)
    return ok, err
}

// CanAny checks if user has permission for any of the given resource/action pairs
func (s *PermissionService) CanAny(ctx context.Context, userID string, checks []struct{ Resource, Action string }) (bool, error) {
    for _, check := range checks {
        ok, err := s.Can(ctx, userID, check.Resource, check.Action)
        if err != nil {
            return false, err
        }
        if ok {
            return true, nil
        }
    }
    return false, nil
}

// GrantRole assigns a role to a user with audit trail
func (s *PermissionService) GrantRole(ctx context.Context, userID, role, grantedBy, reason string) error {
    // Add role mapping in Casbin
    if _, err := s.enforcer.AddGroupingPolicy(userID, role); err != nil {
        return err
    }

    // Audit trail
    return s.auditRepo.Log(ctx, PermissionAudit{
        ID:          uuid.Must(uuid.NewV7()).String(),  // Fixed: Using UUIDv7
        UserID:      userID,
        Action:      "role_granted",
        Subject:     userID,
        Role:        role,
        PerformedBy: grantedBy,
        PerformedAt: time.Now(),
        Reason:      reason,
    })
}

// RevokeRole removes a role from a user with audit trail
func (s *PermissionService) RevokeRole(ctx context.Context, userID, role, revokedBy, reason string) error {
    // Remove role mapping in Casbin
    if _, err := s.enforcer.RemoveGroupingPolicy(userID, role); err != nil {
        return err
    }

    // Audit trail
    return s.auditRepo.Log(ctx, PermissionAudit{
        ID:          uuid.Must(uuid.NewV7()).String(),  // Fixed: Using UUIDv7
        UserID:      userID,
        Action:      "role_revoked",
        Subject:     userID,
        Role:        role,
        PerformedBy: revokedBy,
        PerformedAt: time.Now(),
        Reason:      reason,
    })
}

// GrantPermission assigns a specific permission to a user/role
func (s *PermissionService) GrantPermission(ctx context.Context, subject, resource, action, grantedBy, reason string) error {
    // Add policy in Casbin
    if _, err := s.enforcer.AddPolicy(subject, resource, action); err != nil {
        return err
    }

    // Audit trail
    return s.auditRepo.Log(ctx, PermissionAudit{
        ID:          uuid.Must(uuid.NewV7()).String(),  // Fixed: Using UUIDv7
        UserID:      subject,
        Action:      "permission_granted",
        Subject:     subject,
        Object:      resource,
        ActionType:  action,
        PerformedBy: grantedBy,
        PerformedAt: time.Now(),
        Reason:      reason,
    })
}

// RevokePermission removes a specific permission from a user/role
func (s *PermissionService) RevokePermission(ctx context.Context, subject, resource, action, revokedBy, reason string) error {
    // Remove policy in Casbin
    if _, err := s.enforcer.RemovePolicy(subject, resource, action); err != nil {
        return err
    }

    // Audit trail
    return s.auditRepo.Log(ctx, PermissionAudit{
        ID:          uuid.Must(uuid.NewV7()).String(),  // Fixed: Using UUIDv7
        UserID:      subject,
        Action:      "permission_revoked",
        Subject:     subject,
        Object:      resource,
        ActionType:  action,
        PerformedBy: revokedBy,
        PerformedAt: time.Now(),
        Reason:      reason,
    })
}

// GetUserRoles returns all roles assigned to a user
func (s *PermissionService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
    roles, err := s.enforcer.GetRolesForUser(userID)
    return roles, err
}

// GetUsersForRole returns all users with a specific role
func (s *PermissionService) GetUsersForRole(ctx context.Context, role string) ([]string, error) {
    users, err := s.enforcer.GetUsersForRole(role)
    return users, err
}

// GetRolePermissions returns all permissions for a role
func (s *PermissionService) GetRolePermissions(ctx context.Context, role string) ([][]string, error) {
    perms := s.enforcer.GetPermissionsForUser(role)
    return perms, nil
}
```

## Permission Middleware

```go
// internal/http/middleware/permission.go
package middleware

import (
    "net/http"
    "myapp/internal/rbac"
)

// RequirePermission middleware checks if user has specific permission
func RequirePermission(permSvc *services.PermissionService, resource, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := r.Context().Value("user_id").(string)

            can, err := permSvc.Can(r.Context(), userID, resource, action)
            if err != nil {
                http.Error(w, "Permission check failed", http.StatusInternalServerError)
                return
            }

            if !can {
                http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// RequireAnyPermission checks if user has any of the given resource/action pairs
func RequireAnyPermission(permSvc *services.PermissionService, checks []struct{ Resource, Action string }) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := r.Context().Value("user_id").(string)

            can, err := permSvc.CanAny(r.Context(), userID, checks)
            if err != nil {
                http.Error(w, "Permission check failed", http.StatusInternalServerError)
                return
            }

            if !can {
                http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// RequireRole checks if user has a specific role
func RequireRole(permSvc *services.PermissionService, role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := r.Context().Value("user_id").(string)

            roles, err := permSvc.GetUserRoles(r.Context(), userID)
            if err != nil {
                http.Error(w, "Permission check failed", http.StatusInternalServerError)
                return
            }

            hasRole := false
            for _, r := range roles {
                if r == role {
                    hasRole = true
                    break
                }
            }

            if !hasRole {
                http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Route Protection

```go
// internal/http/server.go
func NewServer(cfg config.Config, services *app.Services) *Server {
    r := chi.NewRouter()

    // ... middleware setup ...

    // Protected routes with permission checks
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth)

        // Posts - require specific permissions
        r.With(middleware.RequirePermission(services.Perms, rbac.ResourcePosts, rbac.ActionCreate)).
            Get(routes.PostNew, postHandler.ShowCreateForm)
        r.With(middleware.RequirePermission(services.Perms, rbac.ResourcePosts, rbac.ActionCreate)).
            Post(routes.CreatePost, postHandler.Create)

        // Edit requires ownership OR posts update permission (checked in handler)
        r.With(middleware.RequireAuth).
            Get(routes.PostEdit, postHandler.ShowEditForm)
        r.With(middleware.RequireAuth).
            Post(routes.UpdatePost, postHandler.Update)

        // Delete requires ownership OR posts delete permission (checked in handler)
        r.With(middleware.RequireAuth).
            Post(routes.DeletePost, postHandler.Delete)

        // Admin section - requires admin access permission
        r.Route("/admin", func(admin chi.Router) {
            admin.Use(middleware.RequirePermission(services.Perms, rbac.ResourceAdmin, rbac.ActionAccess))

            admin.Get("/", adminHandler.Dashboard)

            // User management
            admin.With(middleware.RequirePermission(services.Perms, rbac.ResourceUsers, rbac.ActionRead)).
                Get("/users", adminHandler.ListUsers)
            admin.With(middleware.RequirePermission(services.Perms, rbac.ResourceUsers, rbac.ActionUpdate)).
                Post("/users/{id}/edit", adminHandler.UpdateUser)

            // Role management - requires admin manage permission
            admin.With(middleware.RequirePermission(services.Perms, rbac.ResourceAdmin, rbac.ActionManage)).
                Get("/roles", adminHandler.ListRoles)
            admin.With(middleware.RequirePermission(services.Perms, rbac.ResourceAdmin, rbac.ActionManage)).
                Post("/roles/grant", adminHandler.GrantRole)
        })
    })

    return &Server{Router: r, Config: cfg, Services: services}
}
```

## Handler-Level Permission Checks

```go
// internal/http/handlers/post_handler.go
func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, routes.PostSlugParam)  // Use exported constant, no magic strings
    userID := r.Context().Value("user_id").(string)

    post, err := h.postService.GetBySlug(r.Context(), slug)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // Check ownership OR update permission
    isOwner := post.AuthorID == userID
    canUpdate, _ := h.permService.Can(r.Context(), userID, rbac.ResourcePosts, rbac.ActionUpdate)

    if !isOwner && !canUpdate {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Process update...
}
```

## Template Helpers

```go
// internal/http/views/helpers.go
package views

import (
    "context"
    "myapp/internal/rbac"
)

// Can checks if current user has permission for resource and action
func Can(ctx context.Context, resource, action string) bool {
    userID, ok := ctx.Value("user_id").(string)
    if !ok {
        return false
    }

    permSvc := ctx.Value("permission_service").(*services.PermissionService)
    can, _ := permSvc.Can(ctx, userID, resource, action)
    return can
}

// IsOwner checks if current user owns a resource
func IsOwner(ctx context.Context, ownerID string) bool {
    userID, ok := ctx.Value("user_id").(string)
    if !ok {
        return false
    }
    return userID == ownerID
}

// HasRole checks if current user has a specific role
func HasRole(ctx context.Context, role string) bool {
    userID, ok := ctx.Value("user_id").(string)
    if !ok {
        return false
    }

    permSvc := ctx.Value("permission_service").(*services.PermissionService)
    roles, _ := permSvc.GetUserRoles(ctx, userID)

    for _, r := range roles {
        if r == role {
            return true
        }
    }
    return false
}
```

## Template Usage

```go
// internal/http/views/pages/post.templ
package pages

import (
    "myapp/internal/rbac"
    "myapp/internal/http/views"
)

templ PostView(ctx context.Context, post Post) {
    <article>
        <h1>{ post.Title }</h1>
        <div>{ post.Content }</div>

        <div class="actions">
            if views.IsOwner(ctx, post.AuthorID) || views.Can(ctx, rbac.ResourcePosts, rbac.ActionUpdate) {
                <a href={ templ.URL(routes.PostEditPath(post.Slug)) }>Edit</a>
            }

            if views.IsOwner(ctx, post.AuthorID) || views.Can(ctx, rbac.ResourcePosts, rbac.ActionDelete) {
                <form method="post" action={ templ.URL(routes.DeletePost) }>
                    <input type="hidden" name="slug" value={ post.Slug }/>
                    <button type="submit">Delete</button>
                </form>
            }
        </div>
    </article>
}

templ AdminNav(ctx context.Context) {
    if views.Can(ctx, rbac.ResourceAdmin, rbac.ActionAccess) {
        <nav class="admin-nav">
            <a href="/admin">Dashboard</a>

            if views.Can(ctx, rbac.ResourceUsers, rbac.ActionRead) {
                <a href="/admin/users">Users</a>
            }

            if views.Can(ctx, rbac.ResourceAdmin, rbac.ActionManage) {
                <a href="/admin/roles">Roles</a>
            }

            if views.Can(ctx, rbac.ResourceAdmin, rbac.ActionManage) {
                <a href="/admin/audit">Audit Log</a>
            }
        </nav>
    }
}
```

## Default Roles & Permissions Setup

```go
// internal/db/seed.go
package db

import (
    "context"
    "database/sql"

    "github.com/casbin/casbin/v2"
    "github.com/gofrs/uuid/v5"  // Fixed: Using correct UUID package
    "myapp/internal/rbac"
)

// SeedDefaultRolesAndPermissions creates default RBAC structure using Casbin
func SeedDefaultRolesAndPermissions(ctx context.Context, enforcer *casbin.Enforcer, db *sql.DB) error {
    // Define role permissions
    roles := []struct {
        name        string
        description string
        permissions [][]string // [resource, action]
    }{
        {
            rbac.RoleUser,
            "Regular user - can create own content",
            [][]string{
                {rbac.ResourcePosts, rbac.ActionCreate},
                {rbac.ResourcePosts, rbac.ActionRead},
                {rbac.ResourceComments, rbac.ActionCreate},
            },
        },
        {
            rbac.RoleEditor,
            "Editor - can manage all posts",
            [][]string{
                {rbac.ResourcePosts, rbac.ActionCreate},
                {rbac.ResourcePosts, rbac.ActionRead},
                {rbac.ResourcePosts, rbac.ActionUpdate},
                {rbac.ResourcePosts, rbac.ActionDelete},
                {rbac.ResourcePosts, rbac.ActionPublish},
                {rbac.ResourceComments, rbac.ActionCreate},
                {rbac.ResourceComments, rbac.ActionModerate},
            },
        },
        {
            rbac.RoleModerator,
            "Moderator - can moderate content",
            [][]string{
                {rbac.ResourcePosts, rbac.ActionRead},
                {rbac.ResourceComments, rbac.ActionModerate},
                {rbac.ResourceComments, rbac.ActionDelete},
                {rbac.ResourceUsers, rbac.ActionRead},
            },
        },
        {
            rbac.RoleAdmin,
            "Administrator - full access",
            [][]string{
                {rbac.ResourcePosts, rbac.ActionCreate},
                {rbac.ResourcePosts, rbac.ActionRead},
                {rbac.ResourcePosts, rbac.ActionUpdate},
                {rbac.ResourcePosts, rbac.ActionDelete},
                {rbac.ResourcePosts, rbac.ActionPublish},
                {rbac.ResourceComments, rbac.ActionCreate},
                {rbac.ResourceComments, rbac.ActionModerate},
                {rbac.ResourceComments, rbac.ActionDelete},
                {rbac.ResourceUsers, rbac.ActionRead},
                {rbac.ResourceUsers, rbac.ActionUpdate},
                {rbac.ResourceUsers, rbac.ActionDelete},
                {rbac.ResourceAdmin, rbac.ActionAccess},
                {rbac.ResourceAdmin, rbac.ActionManage},
            },
        },
    }

    // Add permissions for each role to Casbin
    for _, role := range roles {
        // Store role metadata in optional roles table
        _, err := db.ExecContext(ctx, `
            INSERT INTO roles (id, name, description)
            VALUES (?, ?, ?)
            ON CONFLICT (name) DO NOTHING
        `, uuid.Must(uuid.NewV7()).String(), role.name, role.description)  // Fixed: Using UUIDv7
        if err != nil {
            return err
        }

        // Add permissions to Casbin
        for _, perm := range role.permissions {
            resource := perm[0]
            action := perm[1]

            // Add policy: role can perform action on resource
            if _, err := enforcer.AddPolicy(role.name, resource, action); err != nil {
                return err
            }
        }
    }

    // Save policies to database
    return enforcer.SavePolicy()
}

// AssignDefaultRole assigns the "user" role to new users
func AssignDefaultRole(ctx context.Context, enforcer *casbin.Enforcer, userID string) error {
    // Add grouping policy: user belongs to role
    _, err := enforcer.AddGroupingPolicy(userID, rbac.RoleUser)
    return err
}
```

## User Service Integration

```go
// internal/app/user_service.go
func (s *UserService) Create(ctx context.Context, dto CreateUserDTO) (*user.User, error) {
    u := &user.User{
        ID:          uuid.Must(uuid.NewV7()).String(),  // Fixed: Using UUIDv7
        Username:    slug.Sanitize(dto.Username),
        Email:       dto.Email,
        DisplayName: dto.DisplayName,
    }

    // Create user
    if err := s.repo.Create(ctx, u); err != nil {
        return nil, err
    }

    // Assign default "user" role via Casbin
    if err := db.AssignDefaultRole(ctx, s.enforcer, u.ID); err != nil {
        // Log error but don't fail user creation
        slog.Error("failed to assign default role", "error", err, "user_id", u.ID)
    }

    // ... rest of user creation ...

    return u, nil
}
```

## Testing Authorization

```go
// internal/rbac/enforcer_test.go
package rbac_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/casbin/casbin/v2"
)

func TestPermissions(t *testing.T) {
    e := setupTestEnforcer(t)

    // Add test policies
    e.AddPolicy("alice", "posts", "create")
    e.AddGroupingPolicy("bob", "editor")
    e.AddPolicy("editor", "posts", "update")

    // Test direct permission
    ok, _ := e.Enforce("alice", "posts", "create")
    assert.True(t, ok)

    // Test role-based permission
    ok, _ = e.Enforce("bob", "posts", "update")
    assert.True(t, ok)

    // Test denied permission
    ok, _ = e.Enforce("alice", "posts", "delete")
    assert.False(t, ok)
}
```

## Best Practices

1. **Always check permissions, not roles** - Roles can change, permissions are what matter
2. **Use middleware for route-level protection** - Fail fast at the route level
3. **Double-check in handlers for ownership** - Some actions require ownership OR permission
4. **Hide UI elements users can't access** - Better UX when users only see what they can use
5. **Audit all permission changes** - Security requires accountability
6. **Use constants for resources and actions** - Prevents typos and makes refactoring easier
7. **Test authorization thoroughly** - Permission bugs are security bugs

## Troubleshooting

### Permission Denied Errors

- Check if user has the required role
- Verify role has the necessary permissions
- Check Casbin policies are loaded correctly
- Review audit logs for recent changes

### Performance Issues

- Casbin caches policies in memory
- Use batch operations when assigning multiple permissions
- Consider permission inheritance for complex hierarchies

## Next Steps

- Continue to [Web Layer →](./5_web_layer.md)
- Back to [← Authentication](./3_authentication.md)
- Return to [Summary](./0_summary.md)
