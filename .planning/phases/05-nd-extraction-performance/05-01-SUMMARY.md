---
phase: 05-nd-extraction-performance
plan: 01
subsystem: refactoring
tags: [go-packages, domain-extraction, dependency-injection]

requires:
  - phase: 03-interactive-mode-cli-structure
    provides: internal/cli/nd.go with 19 ND functions, internal/cli/batch.go call sites
provides:
  - internal/nd/ package with all 19 ND domain functions decoupled from CLI
  - NormalizeAndSpellCheck with injected corrections map (no config.Load)
  - ExpandAlarmProfiles with injected profileLookup func (no config.Load)
affects: [05-02, batch-processing, nd-features]

tech-stack:
  added: []
  patterns: [dependency-injection-for-config, domain-package-extraction]

key-files:
  created:
    - internal/nd/nd.go
    - internal/nd/nd_test.go
  modified:
    - internal/cli/batch.go
    - internal/cli/batch_test.go

key-decisions:
  - "NormalizeAndSpellCheck accepts corrections map[string]string param -- caller threads from app.Config"
  - "ExpandAlarmProfiles accepts func(string) []string profileLookup -- more flexible than passing *config.Config"
  - "expandAlarmProfilesWithError stays in batch.go unchanged -- CLI-specific error wrapper with config.Load"
  - "corrections threaded through buildBatchCalendar and validateBatchRecord signatures"

patterns-established:
  - "Domain extraction: internal/nd/ is a leaf package with no cli or config imports"
  - "Dependency injection: config-dependent functions accept params instead of calling config.Load()"

requirements-completed: [REF-03]

duration: 5min
completed: 2026-03-30
---

# Phase 5 Plan 1: ND Extraction Summary

**Extracted 19 neurodivergent-domain functions from internal/cli/ to internal/nd/ with dependency-injected signatures for NormalizeAndSpellCheck and ExpandAlarmProfiles**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-30T20:16:55Z
- **Completed:** 2026-03-30T20:22:54Z
- **Tasks:** 2
- **Files modified:** 4 (2 created, 2 modified, 2 deleted)

## Accomplishments
- All 19 ND functions extracted to internal/nd/ package with zero config coupling
- NormalizeAndSpellCheck and ExpandAlarmProfiles decoupled from config.Load() via parameter injection
- LevenshteinDistance and FormatDuration exported for external use
- All 1491 tests pass with 79.5% coverage (above 79% gate)

## Task Commits

Each task was committed atomically:

1. **Task 1: Create internal/nd/ package with all extracted functions and migrated tests** - `e8b9e9e` (feat)
2. **Task 2: Update batch.go call sites, delete cli/nd.go and cli/nd_test.go** - `cf7d2e8` (feat)

## Files Created/Modified
- `internal/nd/nd.go` - All 19 ND domain functions with decoupled signatures
- `internal/nd/nd_test.go` - All migrated tests plus new ExpandAlarmProfiles and NormalizeAndSpellCheckWithCorrections tests
- `internal/cli/batch.go` - Updated 7 call sites to nd.* prefix, threaded corrections param
- `internal/cli/batch_test.go` - Updated buildEventFromBatch calls with new corrections param
- `internal/cli/nd.go` - Deleted
- `internal/cli/nd_test.go` - Deleted

## Decisions Made
- NormalizeAndSpellCheck accepts `corrections map[string]string` -- nil corrections returns text unchanged
- ExpandAlarmProfiles accepts `func(string) []string` -- nil lookup returns specs unchanged
- expandAlarmProfilesWithError stays in batch.go as CLI-specific error wrapper (calls config.Load directly)
- corrections threaded through buildBatchCalendar -> buildEventFromBatch -> validateBatchRecord

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added TestNormalizeAndSpellCheckWithCorrections and TestExpandAlarmProfiles**
- **Found during:** Task 1
- **Issue:** Original tests only tested NormalizeAndSpellCheck with nil corrections (config.Load default). New signature needs tests with actual corrections map.
- **Fix:** Added TestNormalizeAndSpellCheckWithCorrections with correction map and TestExpandAlarmProfiles with lookup function
- **Files modified:** internal/nd/nd_test.go
- **Verification:** go test ./internal/nd/ -v -count=1 passes all tests
- **Committed in:** e8b9e9e (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** Added test coverage for new signatures. No scope creep.

## Issues Encountered
None

## Known Stubs
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- internal/nd/ package ready for Plan 02 optimizations (sweep-line conflict detection, spellcheck cache)
- No blockers

## Self-Check: PASSED

All artifacts verified: internal/nd/nd.go, internal/nd/nd_test.go created; internal/cli/nd.go, internal/cli/nd_test.go deleted; commits e8b9e9e and cf7d2e8 exist; SUMMARY.md created.

---
*Phase: 05-nd-extraction-performance*
*Completed: 2026-03-30*
