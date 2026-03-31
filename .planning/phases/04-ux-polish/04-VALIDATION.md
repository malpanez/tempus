---
phase: 4
slug: ux-polish
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-30
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None needed (Go convention) |
| **Quick run command** | `go test ./internal/cli/ -run "TestDetectEventConflicts\|TestGeneratePrepTimeEvents\|TestResolvePrepLabel" -v -count=1` |
| **Full suite command** | `go test ./... -count=1 -coverprofile=coverage.out && go tool cover -func=coverage.out \| tail -1` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/cli/ -count=1`
- **After every plan wave:** Full suite + coverage check
- **Before `/gsd:verify-work`:** Full suite must be green + coverage ≥ 79%
- **Max feedback latency:** ~30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| T01 | 04-01 | 1 | UX-03 | unit | `go test ./internal/cli/ -run TestDetectEventConflicts -v -count=1` | Exists (nd_test.go:107) — needs new assertions | pending |
| T02 | 04-01 | 1 | UX-03 | integration | `go test ./internal/cli/ -run TestBatch -count=1` | Exists | pending |
| T03 | 04-02 | 1 | UX-04 | unit | `go test ./internal/cli/ -run TestGeneratePrepTimeEvents -v -count=1` | Exists (nd_test.go:194) — needs new cases | pending |
| T04 | 04-02 | 1 | UX-04 | unit | `go test ./internal/cli/ -run TestResolvePrepLabel -v -count=1` | New test needed | pending |
| T05 | all | final | Coverage | metric | `go test ./... -coverprofile=c.out && go tool cover -func=c.out \| tail -1` | N/A | pending |

---

## Wave 0 Requirements

- [ ] `TestDetectEventConflicts` needs assertions for overlap duration string and move suggestion string
- [ ] `TestGeneratePrepTimeEvents` needs test cases with custom label parameter
- [ ] New `TestResolvePrepLabel` for flag > config > default priority resolution
- [ ] `TestCollectBatchWarnings` (batch_test.go:670) updated for new conflict output format

*Existing test infrastructure (stdlib) requires no new framework installation.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `tempus batch -i file.csv --check-conflicts` conflict output readable | UX-03 | Terminal output review | Run with overlapping events CSV, verify names/times/duration/suggestion visible |
| `tempus batch -i file.csv --prep-label "Buffer"` naming | UX-04 | ICS file content review | Open generated ICS, verify prep events named "Buffer: Event Title" |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 45s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
