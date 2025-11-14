# ADR-009: Templ-UI for UI Component Library

**Status:** Accepted
**Date:** 2025-11-13
**Deciders:** Development Team
**Context:** Phase 1 - Core Web Layer

## Context

Phase 1 (Core Web Layer) requires a comprehensive UI component library for generated web applications. The PRD specifies templ + TailwindCSS + Alpine.js for the frontend stack.

**Requirements:**

- Production-ready UI components (forms, layouts, feedback elements)
- Type-safe integration with Go's templ templating system
- Customizable components that users can modify
- Consistent design system out of the box
- Dark mode support
- Accessibility best practices
- Minimal JavaScript footprint

**Alternatives Considered:**

1. **Build components from scratch**
   - Pros: Full control, exactly what we need
   - Cons: Time-consuming, maintenance burden, requires design expertise

2. **DaisyUI (Tailwind component library)**
   - Pros: Large component set, mature ecosystem
   - Cons: Not templ-specific, requires adaptation, CSS-only

3. **Shadcn/ui**
   - Pros: "Copy code, own components" philosophy, excellent design
   - Cons: React-specific, doesn't work with templ

4. **Templ-UI** ✅
   - Pros: Purpose-built for templ, matches our exact stack, shadcn-inspired ownership model
   - Cons: Newer ecosystem, smaller community

## Decision

We will adopt **templ-ui** (https://templui.io/) as the core UI component library for all generated Tracks applications.

**Key Principles:**

1. **Core Dependency** - Every `tracks new` project includes templ-ui by default
2. **Full Starter Set** - Install comprehensive component set (forms + layout + feedback)
3. **Component Ownership** - Users own the copied code, can customize freely
4. **CLI Integration** - Wrap templui CLI for easy component addition

**Component Set Included by Default:**

During `tracks new`, we automatically run `templui init` and install these starter components:

```bash
# Forms
- button, input, textarea, label
- checkbox, radio, select
- form (with validation patterns)

# Layout
- card, modal, dialog
- sidebar, tabs, accordion
- sheet (slide-over panel)

# Feedback
- alert, toast
- progress, spinner, skeleton
```

**Integration Points:**

1. **Project Generation** - `tracks new` automatically runs `templui init` and installs the starter set above
2. **Tool Dependency** - Add `github.com/templui/templui` to go.mod tool block
3. **CLI Commands** - `tracks ui add <component>` and `tracks ui list` wrap `templui` CLI
4. **Documentation** - Guide users on component customization workflow

**Architecture:**

```text
myapp/
├── internal/
│   └── http/
│       └── views/
│           ├── components/     # templ-ui components (user-owned)
│           │   ├── ui/         # Core UI primitives
│           │   │   ├── button.templ
│           │   │   ├── input.templ
│           │   │   └── card.templ
│           │   └── layout/     # Layout components
│           │       ├── sidebar.templ
│           │       └── modal.templ
│           ├── layouts/        # Page layouts
│           │   └── base.templ
│           └── pages/          # Page templates
│               └── home.templ
├── web/
│   ├── css/
│   │   └── app.css            # Tailwind with templ-ui styles
│   └── js/
│       └── app.js             # Alpine.js + component scripts
└── templui.yaml               # Templ-UI configuration
```

## Rationale

**Why templ-ui aligns with Tracks philosophy:**

1. **"Own Your Code"** - Components are copied into the project, not imported as dependencies. Users can modify freely without worrying about breaking upstream changes.

2. **Type Safety** - Built specifically for Go's templ, leveraging compile-time type checking for props and component composition.

3. **Production-Tested** - 40+ components used in real applications, following accessibility and security best practices.

4. **Stack Alignment** - Uses our exact stack: templ + Tailwind + vanilla JS (no framework lock-in).

5. **Developer Experience** - CLI-first workflow matches Tracks' approach. Users can add components incrementally as needed.

6. **CSP-Compliant** - No inline scripts, follows security best practices that align with our generated apps.

## Consequences

**Positive:**

- **Faster Time-to-Value** - Users get production-ready UI immediately
- **Consistent Design** - Professional look-and-feel out of the box
- **Lower Maintenance** - Don't have to build/maintain component library
- **Better DX** - Type-safe components with excellent tooling
- **Customization Freedom** - Users can modify components without constraint

**Negative:**

- **Additional Dependency** - Projects depend on templui CLI being installed
- **Update Strategy** - Users must decide when/how to update components from upstream
- **Learning Curve** - Users need to learn templ-ui conventions and component APIs
- **Bundle Size** - Full starter set adds initial code volume (mitigated by tree-shaking)

**Migration Path:**

For existing projects (pre-Phase 1), users can opt into templ-ui via:

```bash
cd myapp
go install github.com/templui/templui@latest
templui init
templui add "*"  # Install all components
```

**Version Management:**

- Generated projects pin specific templui version in `templui.yaml`
- Users can update via `templui upgrade` when desired
- Breaking changes handled by user at their pace (they own the code)

## Implementation

This decision affects the following Phase 1 epics:

- **Epic 1.2 (Templ Template System)** - Update to include views directory structure
- **Epic 1.4 (Web Build Pipeline)** - Include templui styles in Tailwind config
- **Epic 1.5 (Templ-UI Integration)** - NEW epic for templui installation and configuration
- **Epic 1.6 (Documentation)** - Add component usage guide and customization workflow

See `docs/roadmap/phases/1-core-web.md` for detailed epic breakdowns.

## References

- Templ-UI Documentation: https://templui.io/
- Templ Documentation: https://templ.guide/
- Shadcn/ui (inspiration): https://ui.shadcn.com/
- ADR-005: HTTP Layer Architecture (views directory structure)
- PRD-007: Templates & Assets (TailwindCSS + Alpine.js stack)
