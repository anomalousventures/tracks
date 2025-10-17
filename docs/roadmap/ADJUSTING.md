# How to Adjust the Roadmap

[← Back to Roadmap](./README.md)

## Overview

This guide explains how to update and maintain the Tracks roadmap as development progresses and priorities change. The roadmap is designed to be a living document that evolves with the project.

## Core Principles

### 1. Flexibility Over Rigidity

- The roadmap is a guide, not a contract
- Adjust based on learnings and feedback
- Prioritize value delivery over following the plan

### 2. Transparency

- Document all changes
- Explain reasoning for adjustments
- Keep history of major revisions

### 3. Dependencies Matter

- Respect technical dependencies
- Don't skip prerequisites
- Document new dependencies discovered

## When to Adjust

### Regular Review Points

- **Phase Completion** - After completing each phase
- **Monthly Reviews** - Regular progress check
- **Blockers** - When encountering significant obstacles
- **Discoveries** - When learning invalidates assumptions

### Triggers for Adjustment

- Technical discoveries that change approach
- User feedback requiring priority changes
- Resource constraints
- External dependencies changing
- Security issues requiring immediate attention

## How to Update

### 1. Update Progress Table

In `README.md`, update the status and dates:

```markdown
| Phase | Feature | Status | Started | Completed | PRD Link | Notes |
|-------|---------|--------|---------|-----------|----------|-------|
| 0.1 | CLI with Cobra | Complete | 2025-10-20 | 2025-10-25 | [Link](../prd/...) | Completed ahead of schedule |
```

Status values:

- **Not Started** - Not begun
- **In Progress** - Active development
- **Complete** - Finished and tested
- **Blocked** - Cannot proceed
- **Revised** - Plan changed
- **Deferred** - Postponed to later phase

### 2. Update Phase Documents

In the relevant `phases/*.md` file:

1. Update the phase status
2. Check off completed acceptance criteria
3. Add notes to Adjustments Log
4. Update risks if new ones discovered

Example adjustment log entry:

```markdown
## Adjustments Log

| Date | Change | Reason |
|------|--------|--------|
| 2025-10-25 | Moved i18n to Phase 6 | Not needed for MVP, reduces complexity |
| 2025-10-28 | Added WebSocket support to Phase 4 | Required for real-time features |
```

### 3. Version the Roadmap

When making significant changes:

1. Update version number in `README.md`
2. Add entry to Version History table
3. Commit with clear message

```markdown
### Version History

| Version | Date | Changes | Reason |
|---------|------|---------|--------|
| 1.1.0 | 2025-10-25 | Reordered Phase 3 & 4 | Authentication needed earlier than expected |
```

## Types of Adjustments

### Moving Features Between Phases

When moving a feature:

1. Remove from original phase
2. Add to new phase
3. Update dependencies
4. Document in both phase files

### Adding New Features

When adding features:

1. Determine correct phase based on dependencies
2. Add to phase document
3. Update main progress table
4. Link to PRD if applicable

### Removing Features

When removing features:

1. Move to "Deferred" or "Cancelled" section
2. Document reason for removal
3. Update dependent features
4. Keep historical record

### Changing Timeline

When adjusting timeline:

1. Update estimated duration
2. Document cause of delay/acceleration
3. Adjust dependent phase timelines
4. Communicate to stakeholders

## Documentation Standards

### Commit Messages

Use clear commit messages for roadmap updates:

```text
docs(roadmap): move OAuth to Phase 3

- OAuth needed before code generation
- Updates dependencies accordingly
- Ref: #issue-number
```

### Pull Request Template

When updating via PR:

```markdown
## Roadmap Adjustment

### What Changed
- Moved feature X from Phase Y to Phase Z

### Why
- Discovered dependency on feature A
- User feedback prioritized this feature

### Impact
- Phase Z now 1 week longer
- Phase Y can complete sooner

### Dependencies Updated
- Feature B now depends on X
```

## Review Process

### Who Can Adjust

- **Minor Updates** (status, dates): Any contributor
- **Feature Moves**: Team discussion required
- **Phase Reordering**: Architecture review needed
- **Major Revisions**: Project lead approval

### Review Checklist

Before committing roadmap changes:

- [ ] Dependencies still valid?
- [ ] Timelines realistic?
- [ ] Documentation updated?
- [ ] Version history updated?
- [ ] Stakeholders notified?

## Common Adjustments

### Scenario: Feature Too Complex

**Symptom:** Feature taking much longer than estimated

**Adjustment:**

1. Break feature into smaller parts
2. Move parts to later phases
3. Implement MVP version first
4. Document lessons learned

### Scenario: New Dependency Discovered

**Symptom:** Feature requires unexpected prerequisite

**Adjustment:**

1. Add prerequisite to earlier phase
2. Move dependent feature if needed
3. Update dependency documentation
4. Adjust timelines

### Scenario: Priority Change

**Symptom:** User feedback requires different focus

**Adjustment:**

1. Re-evaluate phase goals
2. Move high-priority items earlier
3. Defer less critical features
4. Update success criteria

## Maintaining Integrity

### Don't

- Skip documenting changes
- Break dependencies without analysis
- Make changes without communication
- Remove history of changes
- Ignore technical debt

### Do

- Keep changes traceable
- Maintain dependency graph
- Communicate early and often
- Learn from adjustments
- Balance flexibility with stability

## Tools for Tracking

### GitHub Issues

- Tag issues with phase labels
- Link PRs to roadmap items
- Track blockers

### GitHub Projects

- Create project board per phase
- Track feature progress
- Visualize dependencies

### Metrics to Track

- Velocity per phase
- Accuracy of estimates
- Number of adjustments
- Completion percentage

## Communication

### When to Communicate

Communicate roadmap changes when:

- Phase completion changes by >1 week
- Features moved between phases
- New blockers identified
- Major revisions needed

### How to Communicate

1. **Team:** Update in next standup/meeting
2. **Stakeholders:** Email with impact summary
3. **Community:** Blog post or changelog
4. **Documentation:** Update immediately

## Learning from Adjustments

### Post-Phase Review

After each phase:

1. What went as planned?
2. What required adjustment?
3. What did we learn?
4. How can we improve estimates?

### Patterns to Watch

- Consistent underestimation?
- Dependency surprises?
- Scope creep?
- Technical debt impact?

Use these patterns to improve future planning.

## Summary

The roadmap is a tool to guide development, not a rigid contract. Adjust it based on reality, learn from changes, and always document the journey. The goal is delivering value, not following a plan perfectly.

Remember: A roadmap that changes based on learning is a sign of a healthy project, not a failure of planning.
