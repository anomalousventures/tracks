# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [v0.3.0] - 2025-11-28

**Phase 1 (Core Web Layer) Complete** - Generated applications now include a production-ready web stack with Chi router, templ templates, HTMX v2, TemplUI components, and comprehensive middleware.

**Highlights:**

- Complete middleware stack with 10 middleware (security headers, CSP, CORS, compression, request ID, logging)
- TemplUI integration with 100+ shadcn-style components and `tracks ui` CLI commands
- Asset pipeline with TailwindCSS v4, HTMX v2, and hashfs content-addressed URLs
- Air live reload for .templ, .css, and .js files
- Working HTMX counter example demonstrating server-side patterns

### Bug Fixes

- HTMX integration and move test apps to /tmp ([#455](https://github.com/anomalousventures/tracks/issues/455))
- allow claude bot in code review workflow ([#368](https://github.com/anomalousventures/tracks/issues/368))

### Code Refactoring

- consolidate asset directories under internal/assets ([#453](https://github.com/anomalousventures/tracks/issues/453))
- remove WHAT comment from SafeHTML ([#355](https://github.com/anomalousventures/tracks/issues/355))
- separate path and parameter constants in routing ([#300](https://github.com/anomalousventures/tracks/issues/300))

### Features

- add middleware stack for generated projects ([#499](https://github.com/anomalousventures/tracks/issues/499))
- complete Epic 1.5 documentation and integration tests ([#498](https://github.com/anomalousventures/tracks/issues/498))
- add tracks ui subcommands for templUI integration ([#497](https://github.com/anomalousventures/tracks/issues/497))
- add tracks ui command for templUI integration ([#496](https://github.com/anomalousventures/tracks/issues/496))
- migrate nav.templ to use templUI components ([#494](https://github.com/anomalousventures/tracks/issues/494))
- add script injection markers to base.templ ([#495](https://github.com/anomalousventures/tracks/issues/495))
- migrate error.templ to use templUI Alert ([#493](https://github.com/anomalousventures/tracks/issues/493))
- migrate about.templ to use templUI components ([#492](https://github.com/anomalousventures/tracks/issues/492))
- migrate counter and home templates to templUI components ([#491](https://github.com/anomalousventures/tracks/issues/491))
- wire templui execution into ProjectGenerator ([#465](https://github.com/anomalousventures/tracks/issues/465)) ([#490](https://github.com/anomalousventures/tracks/issues/490))
- add templUI animation keyframes and remove redundant CSS ([#488](https://github.com/anomalousventures/tracks/issues/488))
- add .templui.json configuration template ([#485](https://github.com/anomalousventures/tracks/issues/485))
- add templUI directory structure for component organization ([#487](https://github.com/anomalousventures/tracks/issues/487))
- add templui tool directive to go.mod template ([#486](https://github.com/anomalousventures/tracks/issues/486))
- Air live reload configuration for asset and template changes ([#459](https://github.com/anomalousventures/tracks/issues/459))
- add HTTP caching middleware for hashed assets ([#458](https://github.com/anomalousventures/tracks/issues/458))
- add HTTP compression middleware for generated projects ([#457](https://github.com/anomalousventures/tracks/issues/457))
- add hashfs content-addressed asset serving ([#456](https://github.com/anomalousventures/tracks/issues/456))
- Epic 1.3 Phases 2 & 3 - Build Pipeline and HTMX Integration ([#454](https://github.com/anomalousventures/tracks/issues/454))
- add basic assets.go template for static file serving ([#394](https://github.com/anomalousventures/tracks/issues/394))
- update .gitignore template for asset pipeline ([#393](https://github.com/anomalousventures/tracks/issues/393))
- create web/ directory structure template (issue [#386](https://github.com/anomalousventures/tracks/issues/386)) ([#392](https://github.com/anomalousventures/tracks/issues/392))
- add make targets for working with generated test projects ([#385](https://github.com/anomalousventures/tracks/issues/385))
- add make targets for test app generation ([#384](https://github.com/anomalousventures/tracks/issues/384))
- wire component test templates into project generator ([#381](https://github.com/anomalousventures/tracks/issues/381))
- wire footer, meta, counter component templates ([#344](https://github.com/anomalousventures/tracks/issues/344)) ([#380](https://github.com/anomalousventures/tracks/issues/380))
- add error_test.go.tmpl with table-driven tests for error pages ([#378](https://github.com/anomalousventures/tracks/issues/378))
- add component test templates with goquery assertions ([#379](https://github.com/anomalousventures/tracks/issues/379))
- add goquery for accessible test queries ([#376](https://github.com/anomalousventures/tracks/issues/376))
- create nav.templ component template (issue [#323](https://github.com/anomalousventures/tracks/issues/323)) ([#366](https://github.com/anomalousventures/tracks/issues/366))
- add HTMX partial rendering pattern templates (issue [#336](https://github.com/anomalousventures/tracks/issues/336)) ([#365](https://github.com/anomalousventures/tracks/issues/365))
- add HTMX attribute helper templates (Task 31, [#333](https://github.com/anomalousventures/tracks/issues/333)) ([#364](https://github.com/anomalousventures/tracks/issues/364))
- add footer.templ component template (issue [#324](https://github.com/anomalousventures/tracks/issues/324)) ([#363](https://github.com/anomalousventures/tracks/issues/363))
- add route registration and page rendering integration tests ([#367](https://github.com/anomalousventures/tracks/issues/367))
- create handler helpers and page handlers (issues [#314](https://github.com/anomalousventures/tracks/issues/314)-318) ([#361](https://github.com/anomalousventures/tracks/issues/361))
- create meta.templ component template (issue [#312](https://github.com/anomalousventures/tracks/issues/312)) ([#360](https://github.com/anomalousventures/tracks/issues/360))
- create page templates (about.templ and error.templ) for Epic 1.2 (issues [#306](https://github.com/anomalousventures/tracks/issues/306), [#307](https://github.com/anomalousventures/tracks/issues/307)) ([#358](https://github.com/anomalousventures/tracks/issues/358))
- create home.templ page template (issue [#305](https://github.com/anomalousventures/tracks/issues/305)) ([#357](https://github.com/anomalousventures/tracks/issues/357))
- create base.templ layout template (issue [#304](https://github.com/anomalousventures/tracks/issues/304)) ([#356](https://github.com/anomalousventures/tracks/issues/356))
- create view directory structure template (issue [#303](https://github.com/anomalousventures/tracks/issues/303)) ([#354](https://github.com/anomalousventures/tracks/issues/354))
- enhance README template with architecture, routing, and troubleshooting docs (issues [#283](https://github.com/anomalousventures/tracks/issues/283), [#284](https://github.com/anomalousventures/tracks/issues/284), [#285](https://github.com/anomalousventures/tracks/issues/285), [#297](https://github.com/anomalousventures/tracks/issues/297)) ([#302](https://github.com/anomalousventures/tracks/issues/302))
- organize route templates into active and example directories (issue [#275](https://github.com/anomalousventures/tracks/issues/275)) ([#301](https://github.com/anomalousventures/tracks/issues/301))
- add users route test template (issue [#288](https://github.com/anomalousventures/tracks/issues/288)) ([#298](https://github.com/anomalousventures/tracks/issues/298))
- add users route example template (issue [#287](https://github.com/anomalousventures/tracks/issues/287)) ([#296](https://github.com/anomalousventures/tracks/issues/296))
- add health route test template (issue [#286](https://github.com/anomalousventures/tracks/issues/286)) ([#295](https://github.com/anomalousventures/tracks/issues/295))
- refactor routes to domain-based structure ([#294](https://github.com/anomalousventures/tracks/issues/294))
- **templates:** add server_test.go.tmpl with integration tests ([#291](https://github.com/anomalousventures/tracks/issues/291))


## [v0.2.0] - 2025-11-13

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


## Unreleased - 2025-10-24

### Bug Fixes

- **release:** remove duplicate changelog from release notes ([#50](https://github.com/anomalousventures/tracks/issues/50))
- **release:** use TRACKS_RELEASER_TOKEN for Homebrew/Scoop publishing ([#49](https://github.com/anomalousventures/tracks/issues/49))
- **release:** add QEMU and Buildx for multi-arch Docker builds ([#48](https://github.com/anomalousventures/tracks/issues/48))
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


[v0.3.0]: https://github.com/anomalousventures/tracks/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/anomalousventures/tracks/compare/v0.1.0...v0.2.0
