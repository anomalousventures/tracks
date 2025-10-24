# Phase 0: Foundation

[← Back to Roadmap](../README.md)

## Overview

This phase establishes the basic CLI tool and project scaffolding capabilities. The goal is to create a working `tracks` command that can generate new Go projects with the proper structure.

**Target Version:** v0.1.0
**Estimated Duration:** 2-3 weeks
**Status:** In Progress (Epic 1 Complete)

## Goals

- Working CLI with Cobra
- `tracks new` command for project creation
- Basic project structure generation
- Build system with Makefile
- Initial documentation

## Epic Breakdown

This phase has been broken down into 5 epics for implementation:

1. ✅ [Epic 1: CLI Infrastructure](./epics/1-cli-infrastructure.md) - Foundation CLI with Cobra, version tracking, help system (COMPLETE)
2. [Epic 2: Template Engine & Embedding](./epics/2-template-engine.md) - Go embed system, template rendering, variable substitution
3. [Epic 3: Project Generation](./epics/3-project-generation.md) - `tracks new` command, directory structure, config generation
4. [Epic 4: Generated Project Tooling](./epics/4-generated-tooling.md) - Makefile, Air config, linting, Docker, CI/CD
5. [Epic 5: Documentation & Installation](./epics/5-documentation.md) - README templates, Docusaurus updates, guides

Each epic contains detailed task breakdowns that will become GitHub issues.

## Features

### 0.1 CLI with Cobra

**Description:** Create the main CLI application using Cobra framework

**Acceptance Criteria:**

- [x] `tracks` command exists and runs
- [x] `--version` flag works
- [x] `--help` provides usage information
- [x] Launches TUI when run without arguments (placeholder message until Phase 4)

**PRD Reference:** [Core Architecture - CLI Tool Structure](../../prd/1_core_architecture.md#cli-tool-structure)

**Implementation Notes:**

- Use `cmd/tracks/main.go` as entry point
- Set up Cobra command structure
- Implement version tracking

### 0.2 Project Scaffolding

**Description:** Implement `tracks new` command to create new projects

**Acceptance Criteria:**

- [ ] Creates proper directory structure
- [ ] Generates go.mod with specified module name
- [ ] Creates basic configuration files
- [ ] Initializes git repository (optional)
- [ ] Supports database driver selection

**PRD Reference:** [Core Architecture - Generated Application Structure](../../prd/1_core_architecture.md#generated-application-structure)

**Implementation Notes:**

- Support flags: `--db`, `--no-git`, `--module`
- Use embed for template files
- Create all directories as specified in PRD

### 0.3 Build System

**Description:** Set up build automation and tooling

**Acceptance Criteria:**

- [ ] Makefile with standard targets
- [ ] Go module management
- [ ] Linting setup (golangci-lint)
- [ ] Test runner configuration

**PRD Reference:** [Core Architecture - Build System](../../prd/1_core_architecture.md#build-system)

**Implementation Notes:**

- Reuse existing Makefile patterns
- Set up CI-friendly commands
- Include help target

### 0.4 Basic Documentation

**Description:** Initial documentation and examples

**Acceptance Criteria:**

- [ ] README for generated projects
- [ ] Basic usage documentation
- [ ] Docusaurus site updates
- [ ] Installation instructions

**Implementation Notes:**

- Leverage existing Docusaurus setup
- Create getting started guide
- Document CLI commands

## Dependencies

### Prerequisites

- Go 1.25+ installed
- Git installed (for project init)
- Make installed (optional but recommended)

### External Dependencies

- github.com/spf13/cobra
- github.com/spf13/viper (for config)

### Internal Dependencies

- None (this is the foundation)

## Success Criteria

1. Developer can install Tracks CLI
2. `tracks new myapp` creates a working project structure
3. Generated project has proper Go module setup
4. Build commands work in generated project
5. Basic documentation exists

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Template complexity | High | Start with minimal templates, iterate |
| Cross-platform paths | Medium | Test on Windows, Mac, Linux |
| Version management | Low | Use git tags and embed version |

## Testing Requirements

- Unit tests for CLI commands
- Integration test for project generation
- Cross-platform testing
- Template rendering tests

## Documentation Requirements

- CLI command reference
- Installation guide
- Quick start tutorial
- Troubleshooting guide

## Future Considerations

Features that depend on this phase:

- All subsequent phases require the CLI to exist
- Code generation commands will extend the CLI
- TUI mode will be added to the main command

## Adjustments Log

This section tracks changes made to the original plan.

| Date | Change | Reason |
|------|--------|--------|
| - | - | No changes yet |

## Notes

- Keep the initial scaffolding minimal - just enough to have a valid Go project
- Focus on getting the CLI structure right as everything builds on it
- The TUI launch (when no args) can be a placeholder initially
- Existing Makefile and project structure can guide implementation

## Next Phase

[Phase 1: Core Web →](./1-core-web.md)
