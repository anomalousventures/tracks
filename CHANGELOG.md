# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


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


[0.1.0]: https://github.com/anomalousventures/tracks/releases/tag/v0.1.0
