# Epic 5: Documentation & Installation

[← Back to Phase 0](../0-foundation.md) | [← Epic 4](./4-generated-tooling.md)

## Overview

Create comprehensive documentation for the Tracks CLI and generated projects. This includes README templates for generated projects, Docusaurus website updates, installation guides, and getting started tutorials. Good documentation ensures developers can successfully install, use, and understand Tracks.

## Goals

- Complete README template for generated projects
- Updated Docusaurus site with Phase 0 features
- Installation guide for all platforms
- Getting started tutorial with examples
- CLI command reference documentation
- Troubleshooting guide

## Scope

### In Scope

- README.md template for generated projects
- Docusaurus documentation pages for Phase 0
- Installation instructions (all platforms)
- Getting started tutorial (step-by-step)
- CLI command reference
- Troubleshooting common issues
- Example project walkthrough
- Screenshot/demo of basic workflow

### Out of Scope

- Video tutorials - future enhancement
- API documentation - comes with later phases
- Architecture deep-dives - later phases
- Contributing guide - separate effort
- Advanced usage patterns - future

## Task Breakdown

The following tasks will become GitHub issues:

1. **Create README.md template for generated projects**
2. **Add project description and features section to README**
3. **Add getting started section to README template**
4. **Add development commands documentation to README**
5. **Create Docusaurus installation page**
6. **Write installation instructions for macOS**
7. **Write installation instructions for Linux**
8. **Write installation instructions for Windows**
9. **Create Docusaurus getting started tutorial**
10. **Write CLI command reference documentation**
11. **Document all `tracks new` flags and options**
12. **Create troubleshooting guide for common issues**
13. **Add example project walkthrough**
14. **Create screenshots/GIFs of CLI workflow**
15. **Update main Tracks README with Phase 0 status**
16. **Review and publish documentation**

## Dependencies

### Prerequisites

- Epic 1-4 complete (documenting working features)
- Docusaurus site already set up (from initial project setup)
- Understanding of what needs to be documented

### Blocks

- User adoption (good docs enable users)
- Future documentation (establishes patterns)

## Acceptance Criteria

- [ ] README template explains what generated project does
- [ ] README shows how to run development server
- [ ] README documents all Makefile targets
- [ ] Installation page covers all platforms
- [ ] Installation instructions include CGO requirements
- [ ] Getting started tutorial works end-to-end
- [ ] CLI command reference lists all commands and flags
- [ ] Troubleshooting guide addresses common issues
- [ ] Example project walkthrough shows real usage
- [ ] Screenshots show CLI in action
- [ ] Documentation passes technical review
- [ ] Links between docs pages work correctly
- [ ] Code examples in docs are tested and work

## Technical Notes

### README Template Structure

Example template for generated project READMEs:

```text
# {{.ProjectName}}

{{.ProjectDescription}}

## Getting Started

### Prerequisites

- Go 1.25+
- Make (optional)

### Development

$ make dev  # Start development server with hot-reload

### Building

$ make build  # Build production binary

### Testing

$ make test   # Run tests
$ make lint   # Run linters

## Project Structure

[Explain directory layout]

## Configuration

[Explain tracks.yaml and environment variables]
```

### Docusaurus Pages to Create/Update

- `/docs/getting-started/installation.md`
- `/docs/getting-started/quickstart.md`
- `/docs/cli/commands.md`
- `/docs/cli/tracks-new.md`
- `/docs/troubleshooting.md`

### Installation Guide Structure

1. System requirements
2. Install Go 1.25+
3. Install Tracks CLI
4. Verify installation
5. Database driver requirements (CGO)
6. Next steps

### Getting Started Tutorial

1. Install Tracks
2. Run `tracks new myapp`
3. Explore generated structure
4. Run `make dev`
5. View application in browser
6. Make a change and see hot-reload
7. Run tests
8. Next: Add your first feature (link to Phase 1)

### Troubleshooting Topics

- CGO not enabled (for libsql/sqlite3)
- Build tools not installed
- Port already in use
- Go version too old
- Module download issues

## Testing Strategy

- Test all code examples in documentation
- Verify all links work
- Test installation instructions on clean VMs
- Review documentation for clarity and completeness
- Get feedback from fresh users (if possible)

## Success Metrics

- Developer can install Tracks following docs alone
- Developer can generate and run project using docs
- Common questions are answered in docs
- Docusaurus builds without errors
- Documentation is clear and concise

## Next Steps

After Phase 0 is complete, documentation will evolve with each phase:

- Phase 1: Document web layer features
- Phase 2: Document database layer
- etc.

Establish good documentation patterns now for future phases to follow.
