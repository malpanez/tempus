---
phase: quick
plan: 260405-2eo
subsystem: internal/cli
tags: [refactor, cpd, sonarcloud, helpers]
dependency_graph:
  requires: []
  provides: [valueAsSlice private helper]
  affects: [ValueAsStringSlice, ValueAsAlarmSlice]
tech_stack:
  added: []
  patterns: [thin-wrapper, strategy-function]
key_files:
  modified:
    - internal/cli/helpers.go
decisions:
  - "Used eachFn callback pattern (func(string, *[]string)) rather than itemFn (func(string) string) because ValueAsAlarmSlice needs to expand one input element into multiple output elements via calendar.SplitAlarmInput"
metrics:
  duration: "5m"
  completed: "2026-04-05"
  tasks_completed: 2
  files_modified: 1
---

# Phase quick Plan 260405-2eo: Extract valueAsSlice to fix SonarCloud CPD Summary

Private `valueAsSlice` helper with callback strategy pattern eliminates duplicated type-switch in `ValueAsStringSlice` and `ValueAsAlarmSlice`, removing ~22 duplicated lines.

## Tasks Completed

| # | Name | Commit | Files |
|---|------|--------|-------|
| 1 | Extract valueAsSlice and rewrite wrappers | ce5102a | internal/cli/helpers.go |
| 2 | Full test suite — no regressions | ce5102a | — |

## Changes Made

`internal/cli/helpers.go` — replaced the two near-identical 33-line type-switch bodies with:

1. `valueAsSlice(v interface{}, eachFn func(string, *[]string), splitFn func(string) []string) []string` — single private helper containing the type-switch once
2. `ValueAsStringSlice` — 3-line thin wrapper passing `strings.TrimSpace` append and `SplitDelimited`
3. `ValueAsAlarmSlice` — 5-line thin wrapper passing `calendar.SplitAlarmInput` expansion and `calendar.SplitAlarmInput`

Net diff: 22 insertions, 43 deletions (-21 lines net).

## Decisions Made

- `eachFn func(string, *[]string)` was chosen over `itemFn func(string) string` because `ValueAsAlarmSlice` must fan out one element to multiple parts. A pointer-to-slice append callback handles both the single-append (strings) and multi-append (alarms) cases uniformly.

## Verification

- `go build ./...` — no errors
- `go test ./internal/cli/... -count=1 -race` — 634 tests passed
- `go test ./... -count=1 -race` — 1828 tests passed across 14 packages
- Exported function signatures unchanged

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check: PASSED

- `internal/cli/helpers.go` exists and contains `valueAsSlice`
- Commit `ce5102a` verified in git log
- 1828 tests pass
