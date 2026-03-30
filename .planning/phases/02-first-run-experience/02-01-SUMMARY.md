---
phase: 02-first-run-experience
plan: 01
subsystem: config
tags: [viper, env-vars, validation, timezone, i18n]

requires:
  - phase: 01-bug-fixes
    provides: "testable output via var stdout, existing config package"
provides:
  - "TEMPUS_* env var overrides via Viper SetEnvPrefix/AutomaticEnv"
  - "ValidateOutputDir function for directory validation"
  - "DetectTimezone function for system timezone detection"
  - "DetectLanguage function for LANG env var parsing"
  - "DefaultAlarmProfile field on Config struct"
  - "Config set validation for timezone and output_dir"
  - "Before/after output on config set success"
affects: [02-02-init-wizard, 02-03-batch-templates]

tech-stack:
  added: []
  patterns: ["Viper env var binding with SetEnvPrefix/AutomaticEnv/SetEnvKeyReplacer"]

key-files:
  created: []
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - main.go

key-decisions:
  - "Used SetEnvKeyReplacer(\".\", \"_\", \"-\", \"_\") as no-op safety net per RESEARCH.md Pitfall 1"
  - "Used os.Stdout directly in runConfigSet since var stdout io.Writer does not exist in current codebase"

patterns-established:
  - "Env var override pattern: TEMPUS_<KEY> overrides config file values"
  - "Config set validation: validate before saving, show before+after on success"

requirements-completed: [CONF-01, CONF-02, CONF-03]

duration: 6min
completed: 2026-03-30
---

# Phase 02 Plan 01: Config Infrastructure Summary

**Viper env var binding (TEMPUS_*), config set validation with before/after output, and detection helpers for init wizard**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-30T00:31:05Z
- **Completed:** 2026-03-30T00:37:02Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- TEMPUS_TIMEZONE, TEMPUS_LANGUAGE, TEMPUS_OUTPUT_DIR, TEMPUS_DATE_FORMAT, TEMPUS_TIME_FORMAT env vars now override config file values
- `config set timezone Invalid/Zone` returns user-friendly error with search suggestion
- `config set output_dir /nonexistent` returns descriptive validation error
- Successful config set shows before/after: `timezone: UTC -> Europe/Madrid`
- DetectTimezone and DetectLanguage exported for init wizard (Plan 02-02)
- DefaultAlarmProfile field on Config struct with "adhd-default" default

## Task Commits

Each task was committed atomically:

1. **Task 1: Env var binding + ValidateOutputDir + DetectTimezone + DetectLanguage + DefaultAlarmProfile** - `434f305` (feat) - TDD: RED+GREEN
2. **Task 2: Wire config validation + before/after into runConfigSet** - `7cff5ef` (feat)

## Files Created/Modified
- `internal/config/config.go` - Added env var binding, ValidateOutputDir, DetectTimezone, DetectLanguage, DefaultAlarmProfile field
- `internal/config/config_test.go` - 15 new tests for env var overrides, output dir validation, timezone/language detection
- `main.go` - runConfigSet with validation switch and before/after output

## Decisions Made
- Used `os.Stdout` directly in runConfigSet instead of `var stdout io.Writer` -- the plan referenced a `stdout` variable at L37 that does not exist in the current codebase. printOK/printErr use fmt.Printf directly.
- Used SetEnvKeyReplacer(".", "_", "-", "_") per RESEARCH.md Pitfall 1 guidance, not the CONTEXT.md D-07 replacer which had a typo.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Used os.Stdout instead of non-existent stdout variable**
- **Found during:** Task 2 (runConfigSet wiring)
- **Issue:** Plan referenced `stdout` variable (plan interface said L37), but no such variable exists in main.go
- **Fix:** Used `os.Stdout` directly in `fmt.Fprintf`
- **Files modified:** main.go
- **Verification:** `go build ./...` succeeds
- **Committed in:** 7cff5ef (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Minimal -- same behavior, different variable reference.

## Issues Encountered
- Coverage at 78.1% (baseline was 78.2% before changes). The 79% gate was already not met at baseline. This plan's changes are coverage-neutral (added 15 tests alongside new code). Pre-existing issue, not caused by this plan.

## Known Stubs
None - all functions are fully implemented and wired.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- DetectTimezone, DetectLanguage, DefaultAlarmProfile are exported and tested -- ready for Plan 02-02 (init wizard)
- ValidateOutputDir and ValidateTimezone are wired into config set -- validation infrastructure complete
- Env var binding active -- TEMPUS_* vars work immediately

---
*Phase: 02-first-run-experience*
*Completed: 2026-03-30*
