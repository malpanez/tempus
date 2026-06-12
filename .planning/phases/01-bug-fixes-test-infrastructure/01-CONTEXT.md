# Phase 1: Bug Fixes & Test Infrastructure - Context

**Gathered:** 2026-03-29
**Status:** Ready for planning

<domain>
## Phase Boundary

Fix 5 data-integrity bugs that produce incorrect or confusing output, plus two targeted infrastructure improvements (testable output, normalization parity). No new features, no architectural changes. The monolith stays intact — that's Phase 3.

**In scope:**
- BUG-01: `stripEmoji()` destroys accented Latin characters
- BUG-02: `promptAlarmField()` hardcoded Spanish strings
- BUG-03: alarm profile not found → silent corruption
- BUG-04: `cityToIANA()` returns empty string silently
- BUG-05: `normalizeDateTimeInput()` not called in `create` path
- REF-06: `printOK`/`printErr` not testable (write to stdout directly)
- REF-02 (reduced scope): only apply normalization to `create` path — full unification deferred to Phase 3

**Out of scope:**
- `internal/parsing` package creation (Phase 3)
- `internal/cli` package split (Phase 3)
- Any new features or commands

</domain>

<decisions>
## Implementation Decisions

### Error Handling (BUG-03, BUG-04)

- **D-01:** Alarm profile not found → **fatal error**, abort execution. Error message must list available profiles: `"Profile 'X' not found. Available: adhd-default, medication, work-meeting"`. Same behavior in both `create` and `batch` — no permissive mode.
- **D-02:** City not recognized by `cityToIANA()` → **fatal error** with actionable suggestion: `"Unknown city 'X'. Use 'tempus timezone list --search X' to find the IANA identifier."` Consistent with D-01 (both fatal, both with guidance).

### REF-02 Scope (reduced for Phase 1)

- **D-03:** REF-02 in Phase 1 means only: call `normalizeDateTimeInput()` before `time.Parse()` in `parseTimedEventTimes()` (and `parseAllDayTimes()` for slash-separated dates). **No new package, no signature changes.** Full unification of 13 functions into `internal/parsing.Parse(ParseOptions)` is deferred to Phase 3 where the monolith split happens.

### REF-06 io.Writer Pattern

- **D-04:** Use **package-level variable**: `var stdout io.Writer = os.Stdout` at the top of main.go. `printOK`/`printErr`/`printDryRunSummary` write to `stdout` instead of `fmt.Printf`. Tests override `stdout` to capture output. Zero changes to callers. Compatible with the future Phase 3 migration where `stdout` becomes a field on the `App` struct.

### Claude's Discretion

- Unicode emoji detection approach for BUG-01: use `unicode.Is(unicode.So, r)` (Symbol, Other category). No external dependency needed — Go stdlib `unicode` package covers this.
- i18n keys for BUG-02: add ~15 keys to all 4 locale files (en, es, pt, ga). Pass `*i18n.Translator` as parameter to `promptAlarmField()`. Caller already has translator context.
- `expandAlarmProfiles()` error return: change signature to `([]string, error)`. Single call site in `addEventAlarms()` — propagate error up to command runner.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Context
- `.planning/PROJECT.md` — Project goals, constraints (no API compat breaks, coverage ≥79%)
- `.planning/REQUIREMENTS.md` — BUG-01..05, REF-02, REF-06 acceptance criteria

### Primary Code Files
- `main.go` — All functions to modify live here (monolith). Key locations:
  - L371: `parseCreateTimes()` — add normalization call here (BUG-05)
  - L404: `parseTimedEventTimes()` — target for normalization (BUG-05)
  - L1269: `normalizeDateTimeInput()` — existing normalization function to reuse
  - L1427: `addEmojiToSummary()` — also has `summary[0] > 127` bug (same pattern as BUG-01)
  - L1692: `stripEmoji()` — BUG-01 fix here
  - L1743: `expandAlarmProfiles()` — BUG-03 fix, change return to `([]string, error)`
  - L3409: `promptAlarmField()` — BUG-02 fix, add translator parameter
  - L3826: `cityToIANA()` — BUG-04 fix, return error instead of empty string
  - L3881: `printOK()` / `printErr()` — REF-06 fix

### i18n System
- `internal/i18n/i18n.go` — `Translator.T(key string, args ...interface{}) string` (L106)
- `locales/en.json`, `locales/es.json`, `locales/pt.json`, `locales/ga.json` — add ~15 keys for alarm prompt strings

### Tests
- `main_test.go` — main integration tests; must stay passing after all changes
- `main_utils_test.go` — covers stripEmoji, spellcheck, conflicts; REF-06 tests go here
- `main_alarm_test.go` — alarm-specific tests; BUG-03 tests go here

No external specs — requirements fully captured in decisions above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `normalizeDateTimeInput()` (L1269): already handles `/` separators, missing zeros, 24h without colon. Reuse directly in `parseTimedEventTimes()` — no modification needed.
- `i18n.Translator.T()` (i18n.go L106): existing translation mechanism. `promptAlarmField()` just needs to receive a `*Translator` parameter.
- `cfg.ListAlarmProfiles()` (config package): exists and returns profile names. Use in BUG-03 error message.

### Established Patterns
- Error propagation: functions return `error` as last return value. `expandAlarmProfiles()` is anomalous in not returning error — BUG-03 fix aligns it with the pattern.
- `var stdout io.Writer = os.Stdout` is not yet in codebase — this is a new pattern for REF-06. Consistent with how Go testing overrides package state.

### Integration Points
- `addEventAlarms()` calls `expandAlarmProfiles()` — needs to handle the new error return
- `runCreate()` → `parseCreateTimes()` → `parseTimedEventTimes()` — normalization inserted at `parseTimedEventTimes` level
- `addEmojiToSummary()` (L1427) has the SAME `summary[0] > 127` bug as `stripEmoji()` — fix both in the same commit

</code_context>

<specifics>
## Specific Ideas

- BUG-01 fix is two places: `stripEmoji()` L1692 AND `addEmojiToSummary()` L1429 (both use `> 127` check)
- BUG-05 normalization: apply at `parseTimedEventTimes()` entry point (affects both `create` and any other callers) rather than only in `parseCreateTimes()` — this is the safer, more complete fix
- REF-06: the variable should be named `stdout` (lowercase, unexported) to match Go convention for package-level test overrides

</specifics>

<deferred>
## Deferred Ideas

- Full `internal/parsing.Parse(ParseOptions)` unification — Phase 3 (monolith split)
- `internal/cli` package creation — Phase 3
- Migrating `promptAlarmField()` to use `charmbracelet/huh` — Phase 3 (interactive mode)

</deferred>

---

*Phase: 01-bug-fixes-test-infrastructure*
*Context gathered: 2026-03-29*
