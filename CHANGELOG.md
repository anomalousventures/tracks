# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [v0.2.0] - 2025-11-13

### Phase 0 Complete - Foundation Ready for Production

This release marks the completion of Phase 0 (Foundation) for Tracks. The CLI tool can now generate production-ready Go web applications with clean architecture, comprehensive tooling, and multi-database support.

**Key Highlights:**

- ✅ Complete `tracks new` command - generates full project scaffolds with layered architecture
- ✅ Multi-database driver support - LibSQL, SQLite3, and PostgreSQL with type-safe queries (SQLC)
- ✅ Development tooling - Makefile, Air live reload, Docker Compose, golangci-lint, Mockery, CI/CD
- ✅ Auto-generated `.env` and automatic Docker service startup with `make dev`
- ✅ Comprehensive documentation - Getting started guides, tutorials, architecture docs, CLI references
- ✅ Production-ready templates - Health endpoints, structured logging, configuration management

**What's Next:** Phase 1 (Core Web Layer) will add code generators for resources, handlers, services, and repositories, plus an interactive TUI for guided project setup.

### Bug Fixes

- resolve LibSQL project generation bugs and add quickstart tutorial ([#249](https://github.com/anomalousventures/tracks/issues/249))
- remove broken tutorial link from footer ([#52](https://github.com/anomalousventures/tracks/issues/52))
- **ci:** add MCP inline comment tool and GitHub context to claude review ([#201](https://github.com/anomalousventures/tracks/issues/201))
- **ci:** add write permissions and bash tool allowlist for claude review ([#200](https://github.com/anomalousventures/tracks/issues/200))
- **workflow:** enable Claude to create PRs and use common tools ([#188](https://github.com/anomalousventures/tracks/issues/188))
- **workflow:** add missing checkout step ([#186](https://github.com/anomalousventures/tracks/issues/186))
- **workflow:** simplify Claude bot to use Interactive Mode ([#185](https://github.com/anomalousventures/tracks/issues/185))

### Code Refactoring

- move Renderer interface to internal/cli/interfaces per ADR-002 ([#195](https://github.com/anomalousventures/tracks/issues/195))
- move newCmd to internal/cli/commands/new.go ([#189](https://github.com/anomalousventures/tracks/issues/189))

### Features

- enhance tracks.yaml template and tests ([#113](https://github.com/anomalousventures/tracks/issues/113), [#114](https://github.com/anomalousventures/tracks/issues/114)) ([#210](https://github.com/anomalousventures/tracks/issues/210))
- add go.mod template dependencies and tests ([#111](https://github.com/anomalousventures/tracks/issues/111), [#112](https://github.com/anomalousventures/tracks/issues/112)) ([#209](https://github.com/anomalousventures/tracks/issues/209))
- add Docker E2E tests for all database drivers ([#233](https://github.com/anomalousventures/tracks/issues/233)) ([#242](https://github.com/anomalousventures/tracks/issues/242))
- add Docker templates and CI workflow for generated projects ([#232](https://github.com/anomalousventures/tracks/issues/232)) ([#241](https://github.com/anomalousventures/tracks/issues/241))
- implement ProjectGenerator ([#230](https://github.com/anomalousventures/tracks/issues/230)) ([#231](https://github.com/anomalousventures/tracks/issues/231))
- implement project directory creation ([#109](https://github.com/anomalousventures/tracks/issues/109), [#110](https://github.com/anomalousventures/tracks/issues/110)) ([#208](https://github.com/anomalousventures/tracks/issues/208))
- add .golangci.yml and Makefile templates for generated projects ([#224](https://github.com/anomalousventures/tracks/issues/224), [#138](https://github.com/anomalousventures/tracks/issues/138)) ([#228](https://github.com/anomalousventures/tracks/issues/228))
- expand README.md template with comprehensive content ([#136](https://github.com/anomalousventures/tracks/issues/136)) ([#223](https://github.com/anomalousventures/tracks/issues/223))
- add sqlc.yaml template ([#135](https://github.com/anomalousventures/tracks/issues/135)) ([#222](https://github.com/anomalousventures/tracks/issues/222))
- add db/db.go template with multi-driver support ([#134](https://github.com/anomalousventures/tracks/issues/134)) ([#221](https://github.com/anomalousventures/tracks/issues/221))
- add DB and REPOSITORIES markers with comprehensive tests ([#132](https://github.com/anomalousventures/tracks/issues/132), [#133](https://github.com/anomalousventures/tracks/issues/133)) ([#220](https://github.com/anomalousventures/tracks/issues/220))
- add config and logging system to generated projects ([#218](https://github.com/anomalousventures/tracks/issues/218))
- add server and routes templates with DI ([#217](https://github.com/anomalousventures/tracks/issues/217))
- add mockery config template with tests ([#216](https://github.com/anomalousventures/tracks/issues/216))
- add health handler template with tests ([#215](https://github.com/anomalousventures/tracks/issues/215))
- add routes constants template with ADR-006 ([#214](https://github.com/anomalousventures/tracks/issues/214))
- add health check templates and tests ([#213](https://github.com/anomalousventures/tracks/issues/213))
- enhance .env.example template with driver-specific URLs ([#117](https://github.com/anomalousventures/tracks/issues/117), [#118](https://github.com/anomalousventures/tracks/issues/118)) ([#212](https://github.com/anomalousventures/tracks/issues/212))
- generate .env file with sensible defaults ([#244](https://github.com/anomalousventures/tracks/issues/244)) ([#246](https://github.com/anomalousventures/tracks/issues/246))
- make dev auto-starts docker-compose services ([#245](https://github.com/anomalousventures/tracks/issues/245)) ([#247](https://github.com/anomalousventures/tracks/issues/247))
- add git initialization logic with comprehensive tests ([#139](https://github.com/anomalousventures/tracks/issues/139), [#140](https://github.com/anomalousventures/tracks/issues/140)) ([#229](https://github.com/anomalousventures/tracks/issues/229))
- add --db, --module, --no-git flags to tracks new command ([#104](https://github.com/anomalousventures/tracks/issues/104)-108) ([#207](https://github.com/anomalousventures/tracks/issues/207))
- implement context propagation pattern (ADR-003) ([#204](https://github.com/anomalousventures/tracks/issues/204))
- implement DI pattern and move validator to validation package ([#163](https://github.com/anomalousventures/tracks/issues/163)-172) ([#203](https://github.com/anomalousventures/tracks/issues/203))
- create internal/generator/interfaces/ directory per ADR-002 ([#196](https://github.com/anomalousventures/tracks/issues/196))
- move template.Renderer interface to generator/interfaces ([#193](https://github.com/anomalousventures/tracks/issues/193)) ([#197](https://github.com/anomalousventures/tracks/issues/197))
- move versionCmd to internal/cli/commands/version.go ([#157](https://github.com/anomalousventures/tracks/issues/157)) ([#198](https://github.com/anomalousventures/tracks/issues/198))
- **architecture:** move interfaces to consumer packages per ADR-002 ([#202](https://github.com/anomalousventures/tracks/issues/202))
- **cli:** add doc.go to interfaces package per ADR-002 ([#199](https://github.com/anomalousventures/tracks/issues/199))
- **cli:** create commands/ directory structure ([#187](https://github.com/anomalousventures/tracks/issues/187))
- **generator:** add validation with go-playground/validator ([#152](https://github.com/anomalousventures/tracks/issues/152))
- **generator:** Epic 3 Phase 1 - Interfaces & Types ([#151](https://github.com/anomalousventures/tracks/issues/151))
- **template:** add validation and integration tests ([#87](https://github.com/anomalousventures/tracks/issues/87))
- **template:** implement rendering engine with variable substitution ([#85](https://github.com/anomalousventures/tracks/issues/85))
- **template:** define core interfaces, types, and errors ([#83](https://github.com/anomalousventures/tracks/issues/83))
- **templates:** add production templates with comprehensive tests ([#86](https://github.com/anomalousventures/tracks/issues/86))
- **templates:** implement embed system for template files ([#84](https://github.com/anomalousventures/tracks/issues/84))


## [0.1.0] - 2025-10-24

### Initial Release - CLI Infrastructure Complete (Epic 1)

This is the first release of Tracks, a batteries-included toolkit for building hypermedia servers in Go. This release focuses on establishing the CLI tool infrastructure that will power project generation and code scaffolding in future releases.

**Key Highlights:**

- ✅ Complete CLI framework with Cobra
- ✅ Flexible output system supporting console, JSON, and TUI modes (TUI coming in Phase 4)
- ✅ Cross-platform builds (Linux, macOS, Windows on amd64/arm64)
- ✅ Multi-distribution support (Homebrew, Scoop, Linux packages, Docker, go install)
- ✅ Comprehensive test coverage with CI/CD automation

**What's Next:** Phase 1 (Core Web) will add the `tracks new` command to generate full-stack Go web applications with type-safe templates, SQL, routing, and built-in auth/RBAC.

### Bug Fixes

- **release:** update GoReleaser for v2 and add Linux packages ([#46](https://github.com/anomalousventures/tracks/issues/46))

### Features

- complete CLI infrastructure (Epic 1) ([#24](https://github.com/anomalousventures/tracks/issues/24)) ([#44](https://github.com/anomalousventures/tracks/issues/44))
- initial project setup with monorepo structure
- initial project setup ([#2](https://github.com/anomalousventures/tracks/issues/2))
- **build:** add cross-platform builds, CI matrix testing, and PostHog analytics ([#40](https://github.com/anomalousventures/tracks/issues/40))
- **build:** add version embedding to Makefile build targets ([#19](https://github.com/anomalousventures/tracks/issues/19)) ([#39](https://github.com/anomalousventures/tracks/issues/39))
- **cli:** define Renderer interface and core types ([#13](https://github.com/anomalousventures/tracks/issues/13)) ([#33](https://github.com/anomalousventures/tracks/issues/33))
- **cli:** wire up flag support to mode detection ([#11](https://github.com/anomalousventures/tracks/issues/11)) ([#30](https://github.com/anomalousventures/tracks/issues/30))
- **cli:** add Table rendering to ConsoleRenderer ([#16](https://github.com/anomalousventures/tracks/issues/16)) ([#36](https://github.com/anomalousventures/tracks/issues/36))
- **cli:** implement ConsoleRenderer with Title and Section ([#15](https://github.com/anomalousventures/tracks/issues/15)) ([#35](https://github.com/anomalousventures/tracks/issues/35))
- **cli:** add Lip Gloss theme system with NO_COLOR support ([#14](https://github.com/anomalousventures/tracks/issues/14)) ([#34](https://github.com/anomalousventures/tracks/issues/34))
- **cli:** implement JSONRenderer for machine-readable output ([#18](https://github.com/anomalousventures/tracks/issues/18)) ([#38](https://github.com/anomalousventures/tracks/issues/38))
- **cli:** add verbose/quiet flags and TRACKS_LOG_LEVEL support ([#12](https://github.com/anomalousventures/tracks/issues/12)) ([#31](https://github.com/anomalousventures/tracks/issues/31))
- **cli:** add Progress bar rendering using Bubbles ViewAs ([#17](https://github.com/anomalousventures/tracks/issues/17)) ([#37](https://github.com/anomalousventures/tracks/issues/37))
- **cli:** implement DetectMode with TTY and CI detection ([#10](https://github.com/anomalousventures/tracks/issues/10)) ([#29](https://github.com/anomalousventures/tracks/issues/29))
- **cli:** define UIMode enum and UIConfig struct ([#28](https://github.com/anomalousventures/tracks/issues/28))
- **cli:** add TUI placeholder message for no-args execution ([#27](https://github.com/anomalousventures/tracks/issues/27))
- **cli:** add comprehensive help text and examples ([#7](https://github.com/anomalousventures/tracks/issues/7)) ([#26](https://github.com/anomalousventures/tracks/issues/26))
- **cli:** add global flags (--json, --no-color, --interactive) ([#25](https://github.com/anomalousventures/tracks/issues/25))
- **cli:** wire up --json flag and enforce Renderer pattern ([#43](https://github.com/anomalousventures/tracks/issues/43))
- **test:** add CLI integration test framework ([#41](https://github.com/anomalousventures/tracks/issues/41))


[v0.2.0]: https://github.com/anomalousventures/tracks/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/anomalousventures/tracks/releases/tag/v0.1.0
