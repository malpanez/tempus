# Phase 3: Interactive Mode & CLI Structure - Context

**Gathered:** 2026-03-30
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver `tempus create --interactive` — a 7-step guided wizard using charmbracelet/huh — and split the 4,223-line `main.go` monolith into `internal/cli/` packages (REF-01), while keeping all existing CLI commands working with identical flags and output.

**In scope:**
- UX-02 — `tempus create --interactive`: 7-step form using charmbracelet/huh
- REF-01 — Monolith split: all command logic moves to `internal/cli/<command>.go`; `main.go` ~100 lines
- Replace survey/v2 (archived 2023) with charmbracelet/huh everywhere it's used (`runInit`, `promptAlarmField`, new interactive wizard)
- Full `internal/parsing.Parse(ParseOptions)` unification (deferred from Phase 1 REF-02)
- `App` struct with `PersistentPreRunE` config/translator injection replacing package-level globals

**Out of scope:**
- Any new commands or features beyond --interactive
- changes to ICS output format
- `internal/nd/` extraction (Phase 5)
- Performance optimization (Phase 5)

</domain>

<decisions>
## Implementation Decisions

### D-01: Interactive wizard — 7-step structure

Extended 7-step flow:

| Step | Label | Fields | Notes |
|------|-------|--------|-------|
| 1/7 | Summary | Event name (text input) | Required |
| 2/7 | Date, Time & Duration | Start date, start time, duration (or all-day toggle) | Required |
| 3/7 | Timezone | Select from config default + search | Pre-filled from config |
| 4/7 | Alarms | Profile selector + optional custom offsets | Pre-filled from config DefaultAlarmProfile |
| 5/7 | Categories | Multi-select: work / health / personal / travel / school / finance + Other | Optional, skippable |
| 6/7 | Details | Location (text), Description (text) | Optional, skippable |
| 7/7 | Confirm | Summary table of all values → write ICS | |

After confirmation: write ICS to output_dir, show `"> Created: ~/calendars/event.ics"` with next-step hint. Same output path logic as non-interactive `create`.

### D-02: Alarms step — profile + custom

Step 4 shows a `huh.Select` with profiles (adhd-default, adhd-countdown, medication, single, none, custom). If user picks "custom", a follow-up `huh.Input` appears for entering comma-separated offsets (e.g., `-2h,-30m,-5m`). Pre-filled with the user's `DefaultAlarmProfile` from config.

### D-03: Categories step — multi-select from preset list

Step 5: `huh.MultiSelect` with options: work, health, personal, travel, school, finance, other. "Other" lets the user type a custom category. Step is optional — user can skip without selecting any.

### D-04: huh migration scope — replace survey everywhere

charmbracelet/huh replaces survey/v2 in **all usages**:
- New `--interactive` wizard (UX-02)
- `runInit` wizard (currently survey — migrate to huh)
- `promptAlarmField` / `confirmQuickEvent` (any remaining survey calls)

After Phase 3: `survey/v2` removed from `go.mod`. Single prompt library across the codebase.

### D-05: Monolith split — all commands move to internal/cli/

REF-01 requires `main.go` → ~100 lines. All command logic moves:

```
internal/cli/
  create.go      ← newCreateCmd, runCreate, parseCreateFlags, runInteractive
  batch.go       ← newBatchCmd, runBatch, loadBatchInput
  quick.go       ← newQuickCmd, runQuick
  lint.go        ← newLintCmd, lintICSFile
  config.go      ← newConfigCmd, runConfigSet, runConfigList
  init.go        ← newInitCmd, runInit
  template.go    ← newTemplateCmd, getBatchTemplateContent
  timezone.go    ← newTimezoneCmd, cityToIANA
  rrule.go       ← newRRuleHelperCmd, promptRRule*
  version.go     ← newVersionCmd
  locale.go      ← newLocaleCmd
```

`main.go` becomes ~100 lines: package declaration, imports, `main()`, `newRootCmd()` that assembles commands from `cli/`.

### D-06: App struct and dependency injection

`App` struct lives in `internal/cli/app.go`:

```go
type App struct {
    Config     *config.Config
    Translator *i18n.Translator
    Stdout     io.Writer
    Stderr     io.Writer
}
```

- `PersistentPreRunE` on root command: loads config once, initializes translator, wires `App` into each command via closure or context
- Package-level `var stdout io.Writer = os.Stdout` in `main.go` is replaced by `App.Stdout`
- Tests construct `App` directly with a `bytes.Buffer` for `Stdout` — no package-level globals to override

### D-07: internal/parsing full unification

The 13 date/time parsing functions are unified into `internal/parsing/` package with a single `Parse(ParseOptions) (time.Time, error)` entry point. This was scoped to Phase 3 in Phase 1 D-03. All callers in `internal/cli/` use the new package.

### Claude's Discretion

- Step progress display style: researcher should determine best huh API for persistent "Step N/7" header (huh.Form groups or custom note — use what renders cleanly without extra dependencies)
- huh version to add to go.mod: researcher verifies latest stable release and Go 1.23 compatibility (noted as concern in STATE.md blockers)
- `internal/parsing.ParseOptions` struct field names and signature: researcher determines from existing function analysis
- `confirmQuickEvent` (survey.Confirm for the `quick` command): migrate to huh or to simple y/N prompt — researcher recommends
- Whether `internal/cli/` commands receive `*App` via function parameter or cobra context: researcher recommends based on cobra patterns

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Context
- `.planning/PROJECT.md` — Core value, constraints (no API compat breaks, coverage ≥79%)
- `.planning/REQUIREMENTS.md` — UX-02, REF-01 acceptance criteria

### Prior Phase Decisions
- `.planning/phases/01-bug-fixes-test-infrastructure/01-CONTEXT.md` — D-04: `var stdout io.Writer` pattern; D-03: `internal/parsing` deferred from Phase 1
- `.planning/phases/02-first-run-experience/02-CONTEXT.md` — D-07: Viper env setup; survey usage in runInit to be migrated

### Primary Code Files
- `main.go` — Current monolith (4,223 lines); `newRootCmd()` L55, `newCreateCmd()` L372, `--interactive` stub L401-403, survey usages L102/113/128/139
- `internal/prompts/prompts.go` — Existing raw bufio.Scanner prompt infrastructure (will coexist or be replaced)
- `internal/config/config.go` — `App.Config` source; `Load()`, `ConfigDir()`, `DetectTimezone()`, `DetectLanguage()`
- `internal/i18n/i18n.go` — `Translator.T(key, args...)` — `App.Translator` source

### State Blockers
- `.planning/STATE.md` — "Verify charmbracelet/huh Go 1.23 compatibility before planning" — researcher must check this first

No external specs — requirements fully captured in decisions above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `var stdout io.Writer = os.Stdout` (main.go L37): becomes `App.Stdout` — pattern already proven in Phase 1
- `internal/config.Load()`: single entry point for config — wire via `PersistentPreRunE`
- `internal/i18n.NewTranslator()`: wire alongside config in `PersistentPreRunE`
- Existing `internal/prompts/` package: raw bufio.Scanner — keep for batch non-interactive input; huh for interactive forms

### Established Patterns
- Phase 1+2 command factory pattern: `newXxxCmd()` returns `*cobra.Command` — keep this pattern, just move files
- Phase 1 D-04: test override via struct field, not package global — `App.Stdout = &buf` in tests
- `runInit` in main.go: already uses survey; becomes the migration reference point for huh conversion

### Integration Points
- `newRootCmd()` (L55): becomes the App wiring point — creates App, wires PersistentPreRunE
- All command `RunE` functions need access to `*App` — decide injection approach (closure vs context)
- `var scanner *bufio.Scanner` (main.go L40): batch CSV reading — keep, not replaced by huh
- `var clockOnlyRe` (main.go L41): utility regex — moves to internal/parsing or internal/utils

</code_context>

<specifics>
## Specific Ideas

- Wizard step 7 confirm output: match the style from `tempus init` summary (Phase 2 D-03) — consistent table format
- huh form groups map naturally to wizard steps: each step is a `huh.Group`, progress shown as group header
- Categories step: "other" entry in MultiSelect opens an Input follow-up, similar to D-02 custom alarm pattern
- If user Ctrl+C during --interactive wizard: clean exit (same behavior as runInit in Phase 2)
- REF-01 split: do NOT break test coverage — move test files alongside their command files (e.g., `internal/cli/create_test.go`)

</specifics>

<deferred>
## Deferred Ideas

- `charmbracelet/huh` theming (colors, high-contrast mode for neurodivergent users) — could be Phase 4 polish
- `tempus create --interactive` supporting `--rrule` (recurrence wizard) — out of scope for Phase 3; rrule wizard already exists as separate command
- Attendees and priority fields in the wizard — not included in 7-step scope; available via flags
- Full `internal/cli/` sub-packages per domain (parsing/, output/, etc.) — REF-01 only requires splitting by command file, not by sub-domain

</deferred>

---

*Phase: 03-interactive-mode-cli-structure*
*Context gathered: 2026-03-30*
