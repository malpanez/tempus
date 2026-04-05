---
phase: quick
plan: 260405-ixt
subsystem: cli
tags: [refactor, cpd, deduplication]
dependency_graph:
  requires: []
  provides: [loadBatchFromStructured, unified-parseDurationEnd]
  affects: [batch, create, helpers]
tech_stack:
  added: []
  patterns: [generic-unmarshal-delegate, single-source-helper]
key_files:
  created: []
  modified:
    - internal/cli/batch.go
    - internal/cli/create.go
    - internal/cli/helpers.go
    - internal/cli/batch_test.go
    - internal/cli/coverage_gaps_test.go
decisions:
  - "Used %q format for parseDurationEnd error messages (batch's version, more informative than create's)"
metrics:
  duration: 4min
  completed: 2026-04-05
---

# Quick Task 260405-ixt: Fix Remaining SonarCloud CPD Duplication Summary

Extract loadBatchFromStructured and unify parseDurationEnd to eliminate CPD duplication between batch.go and create.go.

## What Changed

### Task 1: Extract loadBatchFromStructured and reuse setEventTimezones
- **Commit:** 183e542
- Added `loadBatchFromStructured` generic helper accepting an unmarshal function
- `loadBatchFromJSON` and `loadBatchFromYAML` reduced to 1-liner delegations
- Replaced inline timezone block in `configureBatchEvent` with `setEventTimezones` call

### Task 2: Unify parseDurationEnd into helpers.go
- **Commit:** d6d5187
- Moved `parseDurationEnd` to helpers.go (single source of truth)
- Removed `parseDurationEnd` from create.go
- Removed `parseBatchDurationEnd` from batch.go
- Updated call site in batch.go (argument order swap)
- Updated all test files to use unified function name and signature

## Deviations from Plan

None - plan executed exactly as written.

## Verification

- `go build ./...` passes
- `go test ./... -count=1` passes (1828 tests across 14 packages)
- `go vet ./internal/cli/` clean
- `parseBatchDurationEnd` count in `*.go`: 0
- `loadBatchFromStructured` count in batch.go: 3 (definition + 2 call sites)

## Known Stubs

None.

## Self-Check: PASSED
