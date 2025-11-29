# Phase 3: Auth System

[← Back to Roadmap](../README.md) | [← Phase 2](./2-data-layer.md) | [Phase 4 →](./4-generation.md)

## Overview

This phase implements authentication and authorization with passwordless auth by default, OAuth2 support, and Casbin-based RBAC. Security-first approach with rate limiting and secure sessions.

**Target Version:** v0.5.0
**Estimated Duration:** 3-4 weeks
**Status:** Not Started

## Goals

- Secure session management
- Passwordless authentication (OTP/Magic links)
- OAuth2 provider support
- Casbin RBAC implementation
- Rate limiting and security

## Features

### 3.1 Session Management

**Description:** Implement secure session handling with scs

**Acceptance Criteria:**

- [ ] Cookie-based sessions
- [ ] Redis/memory store options
- [ ] Session middleware
- [ ] CSRF protection

**PRD Reference:** [Authentication - Session Management](../../prd/3_authentication.md#session-management)

**Implementation Notes:**

- Use alexedwards/scs/v2
- Configure secure cookie settings
- Implement session rotation

### 3.2 OTP/Magic Links

**Description:** Passwordless authentication system

**Acceptance Criteria:**

- [ ] OTP generation and validation
- [ ] Magic link generation
- [ ] Email sending integration
- [ ] Rate limiting on attempts

**PRD Reference:** [Authentication - OTP](../../prd/3_authentication.md#1-otp-one-time-password---default)

**Implementation Notes:**

- 6-digit OTP codes
- 10-minute expiry
- Maximum 3 attempts
- Secure token generation

### 3.3 OAuth2 Providers

**Description:** Dynamic OAuth provider registration

**Acceptance Criteria:**

- [ ] GitHub OAuth support
- [ ] Google OAuth support
- [ ] Dynamic provider registration
- [ ] Callback handling

**PRD Reference:** [Authentication - OAuth2 Providers](../../prd/3_authentication.md#3-oauth2-providers)

**Implementation Notes:**

- Use markbates/goth
- Configure via environment variables
- Handle user creation on first login

### 3.4 Casbin RBAC

**Description:** Role-based access control system

**Acceptance Criteria:**

- [ ] Casbin model configuration
- [ ] Default roles and policies
- [ ] Permission checking middleware
- [ ] Database persistence

**PRD Reference:** [Authorization & RBAC](../../prd/4_authorization_rbac.md)

**Implementation Notes:**

- Start with simple roles: admin, user
- Database adapter for policies
- Caching for performance

## Dependencies

### Prerequisites

- Phase 2 completed (database for users)
- Phase 1 completed (sessions need middleware)

### External Dependencies

- github.com/alexedwards/scs/v2
- github.com/markbates/goth
- github.com/casbin/casbin/v2
- Email service adapter

### Internal Dependencies

- Database layer for user storage
- Web layer for auth routes

## Success Criteria

1. Users can sign up and log in
2. OTP codes work reliably
3. OAuth providers function correctly
4. Permissions properly enforced
5. Sessions secure and persistent

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Security vulnerabilities | Critical | Security audit, rate limiting |
| OAuth complexity | Medium | Start with one provider |
| Session management bugs | High | Extensive testing |

## Testing Requirements

- Authentication flow tests
- Session security tests
- Rate limiting tests
- RBAC permission tests
- OAuth integration tests

## Documentation Requirements

- Authentication setup guide
- OAuth provider configuration
- RBAC policy documentation
- Security best practices

## Future Considerations

Features that depend on this phase:

- Any authenticated features
- Admin interfaces
- User-generated content
- API authentication

## Adjustments Log

This section tracks changes made to the original plan.

| Date | Change | Reason |
|------|--------|--------|
| - | - | No changes yet |

## Notes

- Security is paramount - no shortcuts
- Passwordless is the default, passwords optional
- Rate limiting is mandatory
- Test extensively with multiple users/roles

## Next Phase

[Phase 4: Generation →](./4-generation.md)
