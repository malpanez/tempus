---
phase: 04-ux-polish
plan: 02
subsystem: cli
tags: [config, viper, cobra-flags, prep-time, neurodivergent-ux]

requires:
  - phase: 03-interactive-mode-cli-structure
    provides: "Exported GeneratePrepTimeEvents, App struct with Config injection, batchOptions struct"
provides:
  - "PrepTimePrefix config field with mapstructure binding"
  - "resolvePrepLabel priority chain: flag > config > default"
  - "--prep-label flag on tempus batch"
  - "GeneratePrepTimeEvents accepts prepLabel string parameter"
affects: [05-hardening]

tech-stack:
  added: []
  patterns: ["config field + flag override with priority chain resolution"]

key-files:
  created: []
  modified:
    - internal/config/config.go
    - internal/cli/nd.go
    - internal/cli/batch.go
    - internal/cli/nd_test.go

key-decisions:
  - "resolvePrepLabel placed in batch.go (co-located with flag parsing) -- tests in nd_test.go access it via same package"
  - "Medical prep description 'Travel & arrival buffer' protected by checking description == 'Preparation' before overriding"

patterns-established:
  - "Config field + CLI flag + resolution function pattern for user-customizable defaults"

requirements-completed: [UX-04]

duration: 2min
completed: 2026-03-30
---

# Phase 4 Plan 2: Customizable Prep Time Label Summary

**Configurable prep time event prefix via `prep_time_prefix` config key and `--prep-label` batch flag with flag > config > default priority chain**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-30T18:38:40Z
- **Completed:** 2026-03-30T18:40:44Z
- **Tasks:** 1
- **Files modified:** 4

## Accomplishments
- Added `PrepTimePrefix` field to Config struct with viper default "Preparation"
- Added `--prep-label` flag to `tempus batch` command
- Implemented `resolvePrepLabel` with priority: flag > config > default ("Preparation")
- Updated `GeneratePrepTimeEvents` and `createPrepEventIfNeeded` to accept and apply custom label
- Medical prep events ("Travel & arrival buffer") remain unaffected by custom labels
- All 1486 tests pass, coverage at 79.2%

## Task Commits

Each task was committed atomically:

1. **Task 1: Config field, flag, signature change, and prep label resolution** - `f3efae9` (feat)

## Files Created/Modified
- `internal/config/config.go` - Added PrepTimePrefix field, default value, viper.SetDefault
- `internal/cli/nd.go` - Updated GeneratePrepTimeEvents and createPrepEventIfNeeded signatures with prepLabel parameter
- `internal/cli/batch.go` - Added prepLabel to batchOptions, --prep-label flag, resolvePrepLabel function, resolution in runBatch
- `internal/cli/nd_test.go` - Tests already added by plan 04-01 (TestGeneratePrepTimeEventsCustomLabel, TestResolvePrepLabel)

## Decisions Made
- Placed `resolvePrepLabel` in batch.go co-located with flag parsing; tests access it from nd_test.go (same package)
- Medical prep protection uses `description == "Preparation"` check rather than event type, keeping determinePrepTime's return contract intact

## Deviations from Plan

None - plan executed exactly as written. Tests were already present from plan 04-01 execution.

## Issues Encountered
None

## Known Stubs
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Prep time customization complete
- Phase 04 UX polish ready for verification
- Coverage stable at 79.2%

---
*Phase: 04-ux-polish*
*Completed: 2026-03-30*
