---
phase: 5
slug: nd-extraction-performance
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-30
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None needed (Go convention) |
| **Quick run command** | `go test ./internal/nd/ ./internal/cli/ -count=1` |
| **Full suite command** | `go test ./... -count=1 -coverprofile=coverage.out && go tool cover -func=coverage.out \| tail -1` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/nd/ ./internal/cli/ -count=1`
- **After every plan wave:** Full suite + coverage check
- **Before `/gsd:verify-work`:** Full suite must be green + coverage ≥ 79%
- **Max feedback latency:** ~30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| T01 | 05-01 | 1 | REF-03 | unit | `go test ./internal/nd/ -count=1` | Wave 0 (new) | pending |
| T02 | 05-01 | 1 | REF-03 | integration | `go test ./internal/cli/ -run TestBatch -count=1` | Existing | pending |
| T03 | 05-02 | 2 | REF-05 | unit | `go test ./internal/nd/ -run TestDetectEventConflicts -count=1` | Wave 0 (migrated) | pending |
| T04 | 05-02 | 2 | REF-04 | unit | `go test ./internal/nd/ -run TestSpellCheckCache -count=1` | Wave 0 (new) | pending |
| T05 | 05-02 | 2 | REF-04 | unit | `go test ./internal/nd/ -run TestCategoryCache -count=1` | Wave 0 (new) | pending |
| T06 | all | final | Coverage | metric | `go test ./... -coverprofile=c.out && go tool cover -func=c.out \| tail -1` | N/A | pending |

---

## Wave 0 Requirements

- [ ] `internal/nd/nd.go` — new package with all 19 extracted ND functions
- [ ] `internal/nd/nd_test.go` — migrated tests from `internal/cli/nd_test.go`
- [ ] `internal/nd/cache.go` — `SpellCheckCache` and `CategoryCache` types
- [ ] `internal/nd/cache_test.go` — cache unit tests (`TestSpellCheckCache`, `TestCategoryCache`)
- [ ] `internal/cli/nd.go` — deleted (or empty after extraction)
- [ ] `internal/cli/nd_test.go` — deleted (tests live in `internal/nd/`)

*Existing test infrastructure (stdlib) requires no new framework installation.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Batch on 1000+ event file runs measurably faster | REF-04/REF-05 | Benchmark comparison | `go test ./internal/nd/ -bench BenchmarkDetectEventConflicts -benchtime 3s` before/after |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 45s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
