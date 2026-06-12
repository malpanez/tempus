---
phase: 04-ux-polish
plan: 01
subsystem: cli
tags: [conflict-detection, ux, neurodivergent, duration-formatting]

requires:
  - phase: 03-interactive-mode-cli-structure
    provides: "App struct, DetectEventConflicts function, batch warning system"
provides:
  - "Enhanced conflict detection with overlap duration and move suggestion"
  - "formatDuration helper for human-readable duration strings"
affects: [04-02, batch-output]

tech-stack:
  added: []
  patterns: ["overlap duration calculation via max(start)/min(end) subtraction"]

key-files:
  created: []
  modified:
    - internal/cli/nd.go
    - internal/cli/nd_test.go

key-decisions:
  - "Kept DetectEventConflicts return type as []string -- no breaking change to callers"
  - "formatDuration uses integer minutes (not float) for clean output"
  - "Batch tests needed no changes -- header line assertion already matches"

patterns-established:
  - "formatDuration: unexported helper for human-readable durations (Xm, Xh, XhYm)"

requirements-completed: [UX-03]

duration: 2min
completed: 2026-03-30
---

# Phase 4 Plan 1: Enhanced Conflict Detection Summary

**DetectEventConflicts now shows overlap duration (e.g. "by 30m") and suggests when to move the later event (e.g. "Suggestion: move Event 2 to 11:00")**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-30T18:33:17Z
- **Completed:** 2026-03-30T18:35:23Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Conflict output now includes overlap duration in human-readable format (30m, 1h30m, 2h)
- Conflict output now includes a concrete suggestion to move the later event after the earlier one ends
- Added formatDuration helper with full test coverage (table-driven: 0m, 30m, 45m, 1h30m, 2h)
- Coverage maintained at 79.1% (above 79% gate)

## Task Commits

Each task was committed atomically:

1. **Task 1: Enhanced conflict detection with overlap duration and move suggestion** - `928b210` (feat) -- TDD: RED/GREEN in single commit
2. **Task 2: Update batch warning tests and run full suite** - no commit needed (batch tests passed as-is, no file changes)

## Files Created/Modified
- `internal/cli/nd.go` - Added formatDuration helper, enhanced DetectEventConflicts with overlap duration calculation and move suggestion
- `internal/cli/nd_test.go` - Added TestFormatDuration (5 table-driven cases), added overlap duration and suggestion assertions to TestDetectEventConflicts

## Decisions Made
- Kept `DetectEventConflicts` return type as `[]string` -- no breaking change to callers (collectBatchWarnings in batch.go)
- `formatDuration` uses integer minutes via `int(d.Minutes())` for clean output without fractional seconds
- Batch tests (`TestCollectBatchWarnings`) needed no changes -- they check for the word "conflict" in the header line (`"Found %d time conflict(s):"`) not in the detail strings

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Conflict detection enhancement complete, ready for 04-02 (prep time prefix customization)
- No blockers

---
*Phase: 04-ux-polish*
*Completed: 2026-03-30*
