---
phase: quick
plan: 260405-ida
subsystem: internal/cli, internal/nd
tags: [refactor, cpd, sonarcloud, deduplication]
dependency_graph:
  requires: []
  provides: [parseMapsToRecords-helper, newGeneratedEvent-helper]
  affects: [internal/cli/batch.go, internal/nd/nd.go]
tech_stack:
  added: []
  patterns: [shared-helper-extraction]
key_files:
  created: []
  modified:
    - internal/cli/batch.go
    - internal/nd/nd.go
decisions: []
metrics:
  duration: 91s
  completed: "2026-04-05"
  tasks: 2
  files: 2
---

# Quick Task 260405-ida: Fix Remaining SonarCloud CPD Duplication

Extract shared helpers in batch.go and nd.go to eliminate CPD-detected duplicated struct literal blocks, reducing SonarCloud CPD metric.

## Changes

### Task 1: Extract parseMapsToRecords in batch.go (5f55eb2)

Extracted the 14-line record-mapping loop duplicated in `loadBatchFromJSON` and `loadBatchFromYAML` into a shared `parseMapsToRecords` function. Both loaders now call `parseMapsToRecords(raw)` instead of inlining the loop. Net reduction: 15 lines removed.

### Task 2: Extract newGeneratedEvent in nd.go (91c6042)

Extracted the common `calendar.Event` struct construction duplicated in `createTransitionEventIfNeeded` and `createPrepEventIfNeeded` into a shared `newGeneratedEvent` function. Both callers now delegate to `newGeneratedEvent` with their specific parameters. Net change: cleaner separation of event-specific logic from shared struct fields.

## Verification

- `go build ./...` -- passed
- `go vet ./...` -- clean
- `go test ./...` -- 1828 tests passed across 14 packages
- No exported function signatures changed

## Deviations from Plan

None -- plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED
