# Phase 3: Interactive Mode & CLI Structure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-30
**Phase:** 03-interactive-mode-cli-structure
**Areas discussed:** Wizard fields & steps

---

## Wizard Fields & Steps

### Step structure

| Option | Description | Selected |
|--------|-------------|----------|
| Core 7 | Summary → Date → Time+Duration → Timezone → Alarms → Location+Description (skippable) → Confirm | |
| Extended 7 | Summary → Date+Time+Duration → Timezone → Alarms → Categories → Location+Description → Confirm | ✓ |
| Minimal 5 | Summary → Date+Time → Duration → Alarms → Confirm | |

### Categories step input method

| Option | Description | Selected |
|--------|-------------|----------|
| Free text, comma-separated | User types "health, appointment" | |
| Multi-select from common list | Checkboxes: work, health, personal, travel, school, finance, other | ✓ |
| Skipped by default | Only appears with --categories flag | |

### Alarms step

| Option | Description | Selected |
|--------|-------------|----------|
| Profile selector + custom option | Select profile, pick "custom" to enter raw offsets | ✓ |
| Only profile selector | Just the profile, no custom entry | |
| Raw alarm string input | Text field with format hint | |

### Confirm action

| Option | Description | Selected |
|--------|-------------|----------|
| Write ICS directly, show path | Same as non-interactive; shows file path + next step hint | ✓ |
| Show ICS preview first, then write | Print raw ICS before writing | |
| Dry run by default, confirm again | Extra confirmation before write | |

---

## Claude's Discretion

The following areas were NOT discussed — Claude has full discretion:
- huh migration scope (survey/v2 replacement everywhere — aligned with Phase 1+2 context)
- Monolith split scope (all commands per REF-01 spec)
- Step progress display style (huh form groups or persistent header)
- App struct injection pattern (closure vs cobra context)
- internal/parsing.ParseOptions struct design

## Deferred Ideas

- huh theming / high-contrast mode for neurodivergent users
- --rrule support in the interactive wizard
- Attendees and priority fields in wizard scope
