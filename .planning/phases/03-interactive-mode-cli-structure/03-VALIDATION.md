---
phase: 3
slug: interactive-mode-cli-structure
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-30
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None needed (Go convention) |
| **Quick run command** | `go test ./... -count=1` |
| **Full suite command** | `go test ./... -count=1 -coverprofile=coverage.out && go tool cover -func=coverage.out \| tail -1` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1`
- **After every plan wave:** Run `go test ./... -count=1 -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1`
- **Before `/gsd:verify-work`:** Full suite must be green + coverage ≥ 79%
- **Max feedback latency:** ~45 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| TBD-01 | 03-01 | 1 | REF-01 | integration | `go test ./... -count=1` | Yes (move) | pending |
| TBD-02 | 03-01 | 1 | REF-01 | unit | `go test ./internal/cli/ -run TestAppWiring -count=1` | Wave 0 | pending |
| TBD-03 | 03-02 | 1 | REF-01 | metric | `wc -l main.go` (expect ≤ 120) | N/A | pending |
| TBD-04 | 03-02 | 1 | REF-01 | metric | `grep survey go.mod` (expect no match) | N/A | pending |
| TBD-05 | 03-03 | 2 | UX-02 | unit | `go test ./internal/cli/ -run TestRunInteractive -count=1` | Wave 0 | pending |
| TBD-06 | 03-03 | 2 | UX-02 | unit | `go test ./internal/cli/ -run TestInteractiveFormStructure -count=1` | Wave 0 | pending |
| TBD-07 | 03-03 | 2 | UX-02 | unit | `go test ./internal/cli/ -run TestInteractiveDefaults -count=1` | Wave 0 | pending |
| TBD-08 | 03-03 | 2 | UX-02 | unit | `go test ./internal/cli/ -run TestInteractiveCancel -count=1` | Wave 0 | pending |
| TBD-09 | all | final | Coverage | metric | `go test ./... -coverprofile=c.out && go tool cover -func=c.out \| tail -1` | N/A | pending |

---

## Wave 0 Requirements

- [ ] `internal/cli/create_test.go` — TestRunInteractive, TestInteractiveFormStructure, TestInteractiveDefaults, TestInteractiveCancel (UX-02)
- [ ] `internal/cli/app_test.go` — TestAppWiring verifies PersistentPreRunE loads config and translator (REF-01)
- [ ] `internal/parsing/parsing_test.go` — migrated and extended parsing tests (REF-01 internal/parsing)
- [ ] Test helper `testApp()` function in `internal/cli/` for constructing App instances with injected stdout/stderr
- [ ] All existing `main_*_test.go` files split and moved to `internal/cli/` alongside their command sources

*Existing test infrastructure (stdlib) requires no new framework installation.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `tempus create --interactive` full wizard end-to-end | UX-02 | huh interactive prompts cannot be driven in unit tests | Run wizard, fill 7 steps, verify .ics created with correct DTSTART/DTEND/VALARM |
| Step progress visible in terminal | UX-02 | Terminal rendering cannot be asserted in unit tests | Observe "Step 2/7" header appears above each group |
| Ctrl+C exits cleanly mid-wizard | UX-02 | Signal handling in tests is unreliable | Press Ctrl+C at various steps, verify no partial file written |
| survey/v2 fully absent from binary | REF-01 | Binary size check | `go build -o tempus . && strings tempus \| grep -i survey` (expect no match) |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 45s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
