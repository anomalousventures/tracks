# Contributing to Tracks

Thank you for your interest in contributing to Tracks! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## How to Contribute

### Reporting Issues

- **Search existing issues** before creating a new one
- Use the issue templates when available
- Provide clear, detailed descriptions with steps to reproduce
- Include Go version, OS, and relevant environment details

### Suggesting Features

- Open an issue with the "feature request" label
- Describe the use case and expected behavior
- Explain how it aligns with Tracks' philosophy
- Consider implementation complexity and maintenance burden

### Contributing Code

1. **Fork the repository** and create a feature branch
2. **Make your changes** following our coding standards
3. **Write tests** for new functionality
4. **Update documentation** as needed
5. **Submit a pull request** with a clear description

## Development Setup

### Prerequisites

- Go 1.25.3 or later
- Node.js 24+ and pnpm
- Git
- Make

### Initial Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/tracks.git
cd tracks

# Install dependencies
pnpm install

# Run tests to verify setup
make test
```

### Development Workflow

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Make your changes
# ...

# If you modified interfaces, regenerate mocks
make generate-mocks

# Run pre-commit validation (MANDATORY)
make lint              # Must pass with zero errors
make test              # Must pass with zero failures

# Commit your changes
git add .
git commit -m "feat: add your feature description"

# Push to your fork
git push origin feature/your-feature-name
```

## ⚠️ REQUIRED: Pre-Commit Validation

**Before making ANY commit, you MUST successfully complete:**

1. **`make generate-mocks`** - Generate test mocks from interfaces (required after interface changes)
2. **`make lint`** - All linters must pass with zero errors
3. **`make test`** - All tests must pass with zero failures

**Failure to complete these steps successfully means the code is NOT ready to commit.**

### Why This Matters

- **Mocks:** Generated mocks are checked in CI. If mocks are out of date, CI will fail
- **Linting:** Ensures code quality and consistency across the project
- **Tests:** Validates that changes don't break existing functionality

See [CLAUDE.md](./CLAUDE.md) for detailed development guidance.

## Coding Standards

### Go Code Style

- Follow standard Go conventions (gofmt, golint)
- Use meaningful variable and function names
- Write self-documenting code - comments explain WHY, not WHAT
- Keep functions small and focused
- Use interfaces for dependency injection
- Handle errors explicitly, never ignore them

### Code Comments

```go
// GOOD: Explains why
// Cache user data to avoid repeated database queries during the request lifecycle
cache := NewUserCache()

// BAD: Explains what (code already shows this)
// Create a new user cache
cache := NewUserCache()
```

### Error Handling

```go
// GOOD: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// BAD: Generic error
if err != nil {
    return errors.New("error occurred")
}
```

### Testing

- Write table-driven tests when appropriate
- Use t.Run() for subtests
- Mock external dependencies
- Test edge cases and error paths
- Aim for >80% coverage on new code

```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserDTO
        want    *User
        wantErr bool
    }{
        {
            name: "valid user",
            input: CreateUserDTO{Email: "test@example.com"},
            want: &User{Email: "test@example.com"},
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Documentation

- Update docs/ when adding features
- Keep examples up to date
- Use godoc comments for exported functions
- Document non-obvious behavior

## Pull Request Process

### Before Submitting

- [ ] Tests pass locally (`make test`)
- [ ] Linters pass (`make lint`)
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] Branch is up to date with main

### PR Requirements

1. **Clear description** - Explain what and why
2. **Link related issues** - Use "Fixes #123" syntax
3. **Small, focused changes** - One feature per PR
4. **Tests included** - Prove your code works
5. **Documentation updated** - Keep docs in sync

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:

```text
feat(auth): add OAuth provider support

Implement OAuth authentication using goth library.
Supports GitHub, Google, and GitLab providers.

Fixes #123
```

```text
fix(db): handle connection timeouts gracefully

Add retry logic and better error messages for database
connection failures.
```

### PR Title and Description

Since we use squash merging, your **PR title** becomes the commit message in the main branch.
Use Conventional Commits format for PR titles.

When you create a PR, our template provides a structure to fill out:

**Example PR:**

```text
Title: feat(mcp): add session management support

## What

Add session management to the MCP server to maintain state across multiple
client requests. Implements session creation, retrieval, and cleanup.

## Why

Users need persistent state when working with multi-step workflows. Without
session management, each request starts from scratch which breaks complex
operations.

## Testing

- [x] Tests pass locally
- [x] Linting passes (`make lint`)
- [x] Added unit tests for SessionManager
- [x] Added integration test for session lifecycle

## Notes

- Uses in-memory storage for now, will add Redis support in future PR
- Session timeout defaults to 30 minutes (configurable)
- Fixes #42
```

### Review Process

1. Maintainers will review within 3-5 business days
2. Address feedback with new commits (don't force push)
3. Once approved, maintainers will squash and merge
4. Your contribution will be credited in release notes

## Project Structure

```text
tracks/
├── cmd/                         # Executables
│   ├── tracks/                 # Main CLI tool
│   └── tracks-mcp/             # MCP server
├── internal/                    # Internal packages
│   ├── cli/                    # CLI implementation
│   │   ├── commands/          # Command implementations
│   │   ├── interfaces/        # Interface definitions
│   │   ├── renderer/          # Output formatting
│   │   └── ui/                # Mode detection, theming
│   ├── generator/             # Code generators
│   │   └── template/          # Template rendering
│   ├── templates/             # Embedded project templates
│   └── testutil/              # Test utilities
├── tests/                      # Test suites
│   ├── integration/           # Integration tests
│   └── mocks/                 # Generated mocks
├── docs/                       # Documentation
│   ├── prd/                   # Product requirements
│   ├── roadmap/               # Phase planning
│   └── adr/                   # Architecture Decision Records
├── website/                    # Docusaurus documentation site
│   ├── docs/                  # User documentation
│   ├── blog/                  # Blog posts
│   └── static/                # Static assets
└── examples/                   # Example generated apps
```

## Development Tools

### Available Make Commands

#### Code Generation

```bash
make generate-mocks   # Generate test mocks from interfaces (REQUIRED before commit)
```

#### Testing

```bash
make test                # Run unit tests (-v -short with race detector)
make test-integration    # Run integration tests
make test-all           # Run both unit and integration tests
make test-coverage       # Run all tests with coverage reports
make test-e2e-local     # Test E2E workflow locally (mimics CI)
make test-docker-local  # Test Docker workflow locally (mimics CI)
```

#### Linting

```bash
make lint              # Run all linters (Go, markdown, mocks, JavaScript)
make lint-go           # Run golangci-lint on Go code
make lint-md           # Lint markdown files
make lint-md-fix       # Auto-fix markdown linting issues
make lint-js           # Lint JavaScript/TypeScript (website)
make lint-js-fix       # Auto-fix JavaScript linting issues
make lint-mocks        # Verify mocks are up-to-date
make format            # Format code with Prettier
make format-check      # Check code formatting
```

#### Building

```bash
make help                  # Show all available commands
make build                 # Build tracks CLI for current platform
make build-mcp             # Build tracks-mcp server for current platform
make build-all             # Build both binaries for current platform
make build-all-platforms   # Build for all platforms (Linux, macOS, Windows)
make install               # Install tracks locally
make clean                 # Clean build artifacts
```

#### Website

```bash
make website-dev       # Start Docusaurus development server
make website-build     # Build website for production
make website-serve     # Serve built website locally
```

#### Utilities

```bash
make deps              # Download and tidy Go dependencies
```

### Code Generation

Tracks uses [mockery](https://vektra.github.io/mockery/) to generate test mocks from interfaces. This is documented in [ADR-004](./docs/adr/004-mockery-for-test-mock-generation.md).

**When to regenerate mocks:**

- After adding or modifying interface definitions
- Before committing changes that touch interfaces

**How to generate:**

```bash
make generate-mocks
```

Generated mocks are stored in `tests/mocks/` and **must be committed** to the repository. CI checks that mocks are up-to-date via `make lint-mocks`.

### Linting

We use multiple linters to maintain code quality:

- **golangci-lint** - Go code linting and formatting
- **markdownlint-cli2** - Markdown documentation
- **ESLint** - JavaScript/TypeScript (website code)
- **Prettier** - Code formatting (JavaScript, TypeScript, JSON, YAML)
- **Mock freshness check** - Verifies generated mocks are up-to-date

**Run all linters before committing:**

```bash
make lint  # Runs ALL linters (Go, markdown, mocks, JavaScript)
```

**Individual linters:**

```bash
make lint-go       # Go code only
make lint-md       # Markdown only
make lint-js       # JavaScript/TypeScript only
make lint-mocks    # Check mocks are current
```

**Auto-fix issues:**

```bash
make lint-md-fix   # Fix markdown linting issues
make lint-js-fix   # Fix JavaScript linting issues
make format        # Format with Prettier
```

### Testing

Tracks uses a multi-layered testing approach:

**Test Types:**

- **Unit Tests** - Fast, isolated tests colocated with source code. Run with `-short` flag and race detector.
- **Integration Tests** - Test component integration without external services (file generation, validation, git operations). Run on all platforms.
- **Docker E2E Tests** - Full end-to-end tests with databases (Postgres, LibSQL) via Docker Compose. Run only on Ubuntu in CI.

**Running tests:**

```bash
make test                # Unit tests only (fast, with race detector)
make test-integration    # Integration tests
make test-all            # Both unit and integration tests
make test-coverage       # All tests with coverage reports
make test-e2e-local      # E2E workflow locally (mimics CI)
make test-docker-local   # Docker workflow locally (mimics CI)
```

**Platform notes:**

- Unit and integration tests run on all platforms (Linux, macOS, Windows)
- Docker E2E tests require Docker and only run on Ubuntu in CI
- Use `test-e2e-local` and `test-docker-local` for local CI validation

See [CLAUDE.md](./CLAUDE.md) for detailed testing strategy and architecture tests.

## Documentation Guidelines

### Markdown Style

- Use ATX-style headers (`#` not underlines)
- One sentence per line in paragraphs
- Code blocks must have language specifiers
- Keep line length reasonable (no hard limit)
- Use tables for structured data

### Code Examples

- Must be complete and runnable
- Include necessary imports
- Show realistic use cases
- Add comments for clarity

## Release Process

See [docs/RELEASE_PROCESS.md](./docs/RELEASE_PROCESS.md) for details on:

- Versioning strategy (Semantic Versioning)
- Tag creation and management
- Release notes generation
- Publishing process

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Create an issue with the bug template
- **Chat**: Join our community (link TBD)
- **Email**: contribute@anomalousventures.com

## Recognition

Contributors are recognized in:

- Release notes
- [CONTRIBUTORS.md](./CONTRIBUTORS.md) file
- Annual contribution reports

Thank you for making Tracks better!

---

**Note**: This is a living document. Suggest improvements via pull request.
