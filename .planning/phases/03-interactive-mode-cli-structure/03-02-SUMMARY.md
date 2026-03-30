---
phase: 03-interactive-mode-cli-structure
plan: 02
subsystem: cli
tags: [cobra, huh, factory-pattern, command-migration, survey-replacement]

requires:
  - phase: 03-01
    provides: App struct, helpers, PersistentPreRunE, internal/parsing package
provides:
  - Cobra command factories for create, quick, init, config, version in internal/cli/
  - ND feature functions (spell check, conflict detection, prep time, emoji) in internal/cli/nd.go
  - huh-based interactive prompts replacing survey/v2 in quick and init commands
  - DetectTimezone, DetectLanguage, ValidateOutputDir in config package
affects: [03-03, 03-04, batch-commands, main-cleanup]

tech-stack:
  added: [github.com/charmbracelet/huh v1.0.0]
  patterns: [command-factory-with-app-injection, thin-wrapper-delegation]

key-files:
  created:
    - internal/cli/create.go
    - internal/cli/create_test.go
    - internal/cli/quick.go
    - internal/cli/config.go
    - internal/cli/config_test.go
    - internal/cli/version.go
    - internal/cli/nd.go
    - internal/cli/nd_test.go
    - internal/cli/init.go
    - internal/cli/init_test.go
  modified:
    - main.go
    - go.mod
    - go.sum
    - internal/config/config.go
    - internal/config/config_test.go
    - main_alarm_test.go
    - main_categories_test.go
    - main_coverage_test.go
    - main_recurrence_test.go
    - main_test.go
    - main_utils_test.go

key-decisions:
  - "Inlined exdate parsing in create.go to avoid import cycle (internal/parsing imports internal/cli)"
  - "Used thin wrapper pattern in main.go for batch code calling ND functions"
  - "huh forms return error in non-TTY (tests) which triggers clean exit via return nil"

patterns-established:
  - "Command factory: func NewXxxCmd(app *App) *cobra.Command with RunE closure capturing app"
  - "Thin wrapper: main.go lowercase functions delegate to exported cli.Function for backward compat"
  - "huh form pattern: var result; form := huh.NewForm(huh.NewGroup(...)); form.Run()"

requirements-completed: [REF-01, UX-02]

duration: 120min
completed: 2026-03-30
---

# Phase 3 Plan 2: Command Migration and Survey-to-Huh Replacement Summary

**Moved create/quick/init/config/version commands to internal/cli/ with factory pattern and replaced survey/v2 with charmbracelet/huh v1.0.0 for interactive prompts**

## Performance

- **Duration:** ~120 min
- **Started:** 2026-03-30
- **Completed:** 2026-03-30
- **Tasks:** 3/3
- **Files modified:** 21

## Accomplishments
- Moved 5 Cobra commands (create, quick, init, config, version) from main.go to internal/cli/ using factory pattern with App injection
- Moved all ND feature functions (spell check, conflict detection, prep time, emoji, overwhelm detection) to internal/cli/nd.go
- Replaced all survey/v2 usage with charmbracelet/huh v1.0.0 (Confirm, Input with validation, Select with named options)
- Reduced main.go from ~3550 to ~2500 lines while maintaining backward compatibility via thin wrappers
- Coverage maintained at 79.1% with cross-package tracking

## Task Commits

Each task was committed atomically:

1. **Task 1: Move commands and ND functions to internal/cli/** - `ad8ee62` (feat)
2. **Task 2: Move tests to internal/cli/ package** - `0f3907f` (test)
3. **Task 3: Migrate runInit from survey to huh** - `ebf3c4f` (feat)

## Files Created/Modified
- `internal/cli/create.go` - NewCreateCmd factory with all create logic, flag parsing, calendar output
- `internal/cli/quick.go` - NewQuickCmd factory with huh confirm prompt replacing survey
- `internal/cli/init.go` - NewInitCmd factory with 5 huh prompts (confirm, 2x input, 2x select)
- `internal/cli/config.go` - NewConfigCmd factory with set/list/alarm-profiles subcommands
- `internal/cli/version.go` - NewVersionCmd factory with build info injection
- `internal/cli/nd.go` - All neurodivergent feature functions (exported)
- `internal/cli/nd_test.go` - Tests for ND functions including new spell check and category tests
- `internal/cli/create_test.go` - Tests for create, version, quick commands and parse functions
- `internal/cli/config_test.go` - Tests for config command and subcommands
- `internal/cli/init_test.go` - Tests for init command (registered, help, existing config, fresh config)
- `internal/config/config.go` - Added DetectTimezone, DetectLanguage, ValidateOutputDir
- `internal/config/config_test.go` - Tests for new config functions and alarm profiles
- `main.go` - Slim root command using cli factories, thin wrappers for batch code
- `go.mod` / `go.sum` - Added charmbracelet/huh v1.0.0 and transitive dependencies

## Decisions Made
- **Inlined exdate parsing**: create.go cannot import internal/parsing (would create import cycle since parsing imports cli). Inlined simpler time.Parse-based exdate handling.
- **Thin wrapper pattern**: main.go batch functions still call ND helpers. Added lowercase wrappers (e.g., `addEmojiToSummary -> cli.AddEmojiToSummary`) to avoid breaking batch code.
- **huh non-TTY behavior**: huh forms return error when no TTY is available. Init command handles this by returning nil (clean exit), which makes the existing-config-no-overwrite test work correctly.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Import cycle between internal/parsing and internal/cli**
- **Found during:** Task 1
- **Issue:** Plan 03-01 made internal/parsing import internal/cli for helper functions. create.go importing internal/parsing would create a cycle.
- **Fix:** Inlined a simpler exdate parsing in create.go using time.Parse directly instead of importing parsing.ParseExDateValues
- **Files modified:** internal/cli/create.go
- **Verification:** Build succeeds, exdate tests pass
- **Committed in:** ad8ee62

**2. [Rule 3 - Blocking] Missing config functions (DetectTimezone, DetectLanguage, ValidateOutputDir)**
- **Found during:** Task 3
- **Issue:** Init command requires config.DetectTimezone(), config.DetectLanguage(), config.ValidateOutputDir() which were added in Phase 2 but don't exist in this worktree
- **Fix:** Added the three functions to internal/config/config.go with corresponding tests
- **Files modified:** internal/config/config.go, internal/config/config_test.go
- **Verification:** Tests pass, init command compiles
- **Committed in:** ebf3c4f

---

**Total deviations:** 2 auto-fixed (2 blocking issues)
**Impact on plan:** Both fixes necessary to complete the migration. No scope creep.

## Issues Encountered
- Worktree was behind main branch (missing Plan 01 outputs) -- resolved with git merge
- Coverage dipped below 79% after adding new uncovered init code -- resolved by adding tests for config functions, alarm profiles, and ND helpers

## Known Stubs
None -- all moved functions are fully wired to their original callers.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- internal/cli/ package now contains create, quick, init, config, version commands with factory pattern
- Batch, lint, template, locale, timezone, rrule commands remain in main.go for Plan 03-03
- huh dependency established for future interactive prompt work
- Thin wrappers in main.go can be removed once batch code is also migrated

## Self-Check: PASSED

- All 10 created files verified present
- All 3 task commits verified (ad8ee62, 0f3907f, ebf3c4f)
- All tests pass (1528 tests across 13 packages)
- Coverage at 79.1% (cross-package)

---
*Phase: 03-interactive-mode-cli-structure*
*Completed: 2026-03-30*
