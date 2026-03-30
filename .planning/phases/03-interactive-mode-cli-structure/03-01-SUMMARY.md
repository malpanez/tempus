---
phase: 03-interactive-mode-cli-structure
plan: 01
subsystem: cli
tags: [go, cobra, refactor, parsing, dependency-injection]

requires:
  - phase: 02-env-vars-config-validation
    provides: config.Load() with env var support, Viper AutomaticEnv
provides:
  - App struct with Config/Translator/Stdout/Stderr DI fields
  - SetupPersistentPreRunE for config loading and translator wiring
  - TestApp() helper for test construction
  - 20 exported helper functions in internal/cli/helpers.go
  - Unified Parse(ParseOptions) entry point for date/time parsing
  - Exported normalization and parsing utilities in internal/parsing/
affects: [03-02, 03-03, 03-04]

tech-stack:
  added: []
  patterns: [factory-function-DI, thin-wrapper-migration, unified-parsing-API]

key-files:
  created:
    - internal/cli/app.go
    - internal/cli/app_test.go
    - internal/cli/helpers.go
    - internal/cli/helpers_test.go
    - internal/parsing/parsing.go
    - internal/parsing/parsing_test.go
  modified:
    - main.go

key-decisions:
  - "PrintOK/PrintErr take io.Writer parameter for testability instead of using package-level stdout"
  - "Thin wrappers in main.go delegate to new packages -- zero behavior change, all 1499 tests pass"
  - "Parse() unifies create and batch parsing paths via ParseOptions struct"
  - "Config fallback to defaults when config.Load() fails (tolerant for init/version commands)"

patterns-established:
  - "Thin wrapper migration: move logic to new package, leave wrapper in main.go, remove wrapper when command migrates"
  - "App struct DI: Config/Translator/Stdout/Stderr injected via SetupPersistentPreRunE"
  - "ParseOptions/ParseResult: single entry point for all date/time parsing"

requirements-completed: [REF-01]

duration: 15min
completed: 2026-03-30
---

# Phase 03 Plan 01: Foundation Packages Summary

**App struct with DI wiring, 20 exported CLI helpers, and unified Parse(ParseOptions) date/time API in internal/cli/ and internal/parsing/**

## Performance

- **Duration:** 15 min
- **Started:** 2026-03-30T02:06:11Z
- **Completed:** 2026-03-30T02:21:11Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Created internal/cli/ package with App struct, SetupPersistentPreRunE, TestApp(), and 20 exported helper functions
- Created internal/parsing/ package with unified Parse(ParseOptions) replacing both parseCreateTimes and parseBatchTimes
- Migrated all 17 parsing functions plus 5 supporting functions to internal/parsing/
- All 1499 tests pass, coverage maintained at 79.0%

## Task Commits

Each task was committed atomically:

1. **Task 1: Create App struct, helpers, and PersistentPreRunE wiring** - `d6f92f3` (feat)
2. **Task 2: Create internal/parsing/ with unified Parse() entry point** - `6937dc2` (feat)

## Files Created/Modified
- `internal/cli/app.go` - App struct with DI fields, SetupPersistentPreRunE, TestApp()
- `internal/cli/app_test.go` - Tests for App construction and PersistentPreRunE wiring
- `internal/cli/helpers.go` - 20 exported helper functions (ParseBoolish, SplitDelimited, ValueAsString, PrintOK, etc.)
- `internal/cli/helpers_test.go` - Comprehensive tests for all helper functions
- `internal/parsing/parsing.go` - ParseOptions/ParseResult types, Parse() entry point, all parsing and normalization functions
- `internal/parsing/parsing_test.go` - Tests for Parse(), NormalizeDateTimeInput, ParseDateTimeWithTZ, and all utilities
- `main.go` - Added cli/parsing imports, replaced 37 function bodies with thin wrappers

## Decisions Made
- PrintOK/PrintErr accept io.Writer as first parameter for testability (per D-06) -- main.go wrappers pass os.Stdout
- Config fallback to hardcoded defaults when Load() fails -- tolerant for init/version commands
- Parse() normalizes all inputs via NormalizeDateTimeInput before delegating to parseAllDay/parseTimed
- Kept originals in main.go as thin wrappers to preserve all existing test compilation

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added helpers_test.go for coverage gate**
- **Found during:** Task 2
- **Issue:** Coverage dropped to 75.1% after extracting functions to new packages (cross-package calls not tracked)
- **Fix:** Added comprehensive helpers_test.go with tests for all 20 exported functions
- **Files modified:** internal/cli/helpers_test.go
- **Verification:** Coverage restored to 79.0%
- **Committed in:** 6937dc2 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** Coverage-restoring tests were necessary to maintain 79% gate. No scope creep.

## Issues Encountered
- `ParseHumanDuration("11:00")` returns 11h -- clock-like values that are valid durations are intentionally not prepended with today's date in NormalizeClockOnlyDateTimes. Adjusted test expectation accordingly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- internal/cli/ and internal/parsing/ packages ready for Plans 02/03 to move command code
- Thin wrappers in main.go ensure backward compatibility during incremental migration
- All downstream plans can import these packages without defining their own types

---
*Phase: 03-interactive-mode-cli-structure*
*Completed: 2026-03-30*
