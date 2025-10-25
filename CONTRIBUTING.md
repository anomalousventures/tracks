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

# Run linters
make lint

# Run tests
make test

# Commit your changes
git add .
git commit -m "feat: add your feature description"

# Push to your fork
git push origin feature/your-feature-name
```

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
├── cmd/                    # CLI commands
├── internal/              # Internal packages
│   ├── generator/        # Code generators
│   ├── tui/             # Terminal UI
│   └── mcp/             # MCP server
├── docs/                 # Documentation
│   └── prd/             # Product requirements
├── tests/               # Integration tests
└── examples/            # Example applications
```

## Development Tools

### Available Make Commands

```bash
make help             # Show all available commands
make test             # Run test suite
make test-integration # Run integration tests
make lint             # Run all linters
make lint-md          # Lint markdown files
make build            # Build the CLI
make install          # Install locally
```

### Linting

We use multiple linters to maintain code quality:

- **golangci-lint** - Go code linting
- **markdownlint-cli2** - Markdown documentation
- **gofmt** - Go code formatting

Run all linters before committing:

```bash
make lint
```

### Testing

```bash
# Unit tests
make test

# Integration tests
make test-integration

# With coverage
make test-coverage
```

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
