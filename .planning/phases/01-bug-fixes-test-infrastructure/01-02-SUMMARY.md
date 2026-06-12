---
phase: 01-bug-fixes-test-infrastructure
plan: 02
subsystem: cli
tags: [error-handling, i18n, alarm-profiles, timezone, go]

requires:
  - phase: 01-01
    provides: "var stdout io.Writer, printOK/printErr using fmt.Fprintf(stdout)"
provides:
  - "expandAlarmProfiles returns ([]string, error) with descriptive profile listing"
  - "cityToIANA returns (string, error) with search suggestion"
  - "promptAlarmField internationalized with 13 i18n keys in 4 locales"
  - "Error propagation through create and batch alarm paths"
affects: [02-core-features, 03-ux-improvements]

tech-stack:
  added: []
  patterns:
    - "Error propagation through void-to-error signature changes"
    - "i18n keys for interactive prompt strings"

key-files:
  created: []
  modified:
    - main.go
    - main_alarm_test.go
    - main_test.go
    - locales/en.json
    - locales/es.json
    - locales/pt.json
    - locales/ga.json

key-decisions:
  - "Changed configureEvent and createCalendarWithEvent to return error for alarm propagation in create path"
  - "Unknown city is fatal in runTZInfo -- aborts immediately before fuzzy fallback per D-02"

patterns-established:
  - "Error return pattern: functions that were void now return error when calling expandAlarmProfiles"

requirements-completed: [BUG-02, BUG-03, BUG-04]

duration: 7min
completed: 2026-03-29
---

# Phase 01 Plan 02: Alarm Profiles Error Handling + i18n Summary

**expandAlarmProfiles and cityToIANA return errors with actionable messages; promptAlarmField internationalized across 4 locales with 13 i18n keys**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-29T22:32:55Z
- **Completed:** 2026-03-29T22:40:24Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- expandAlarmProfiles returns descriptive error listing available profiles when profile not found (BUG-03)
- cityToIANA returns error with "tempus timezone list --search" suggestion for unknown cities (BUG-04)
- promptAlarmField uses i18n Translator for all 13 user-facing strings across en/es/pt/ga (BUG-02)
- Both create and batch paths now expand alarm profiles and propagate errors (D-01)
- Unknown city is fatal -- aborts before fuzzy fallback (D-02)

## Task Commits

Each task was committed atomically:

1. **Task 1: BUG-03 + BUG-04 error returns** - `119caed` (test: RED), `e935b8d` (feat: GREEN)
2. **Task 2: BUG-02 i18n promptAlarmField** - `ace84df` (feat)

**Plan metadata:** pending (docs: complete plan)

_Note: Task 1 followed TDD with RED/GREEN commits_

## Files Created/Modified
- `main.go` - expandAlarmProfiles, cityToIANA, addBatchAlarms, configureBatchEvent, addEventAlarms, configureEvent, createCalendarWithEvent signature changes + error propagation; promptAlarmField i18n
- `main_alarm_test.go` - TestExpandAlarmProfiles with error cases
- `main_test.go` - TestCityToIANA updated for (string, error) return; TestAlarmPromptI18nKeys for all 13 keys in 4 locales
- `locales/en.json` - 13 alarm_prompt_* keys (English)
- `locales/es.json` - 13 alarm_prompt_* keys (Spanish)
- `locales/pt.json` - 13 alarm_prompt_* keys (Portuguese)
- `locales/ga.json` - 13 alarm_prompt_* keys (Galician)

## Decisions Made
- Changed configureEvent and createCalendarWithEvent to return error -- necessary to propagate alarm profile errors up to runCreate in the create path
- Unknown city is fatal per D-02: runTZInfo returns error immediately on cityErr, before any fuzzy fallback runs

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Changed configureEvent and createCalendarWithEvent signatures**
- **Found during:** Task 1 (addEventAlarms error propagation)
- **Issue:** Plan specified changing addEventAlarms to return error, but its caller configureEvent returned void, and configureEvent's caller createCalendarWithEvent also returned void
- **Fix:** Changed both to return error, updated caller chain up to runCreate
- **Files modified:** main.go
- **Verification:** go build succeeds, all tests pass
- **Committed in:** e935b8d

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary for error propagation correctness. No scope creep.

## Issues Encountered
None

## Known Stubs
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All BUG-02, BUG-03, BUG-04 fixes complete
- Error propagation chains verified in both create and batch paths
- i18n coverage for alarm prompts complete across all supported locales
- 1435 tests passing, no regressions

---
*Phase: 01-bug-fixes-test-infrastructure*
*Completed: 2026-03-29*

## Self-Check: PASSED
