---
phase: 1
slug: bug-fixes-test-infrastructure
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-29
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None needed (Go convention) |
| **Quick run command** | `go test ./... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1 -race` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1 -short`
- **After every plan wave:** Run `go test ./... -count=1 -race`
- **Before `/gsd:verify-work`:** Full suite must be green + coverage ≥ 79%
- **Max feedback latency:** ~30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| TBD-01 | 01-01 | 0 | BUG-01 | unit | `go test -run TestStripEmoji -v` | Yes (update) | pending |
| TBD-02 | 01-01 | 0 | BUG-01 | unit | `go test -run TestAddEmojiToSummary -v` | Yes (update) | pending |
| TBD-03 | 01-02 | 0 | BUG-02 | unit | `go test -run TestAlarmPromptI18nKeys -v` | No (Wave 0) | pending |
| TBD-04 | 01-02 | 0 | BUG-03 | unit | `go test -run TestExpandAlarmProfiles -v` | No (Wave 0) | pending |
| TBD-05 | 01-02 | 0 | BUG-04 | unit | `go test -run TestCityToIANA -v` | Yes (update) | pending |
| TBD-06 | 01-03 | 0 | BUG-05 | unit | `go test -run TestParseTimedEventTimes -v` | No (Wave 0) | pending |
| TBD-07 | 01-03 | 0 | REF-06 | unit | `go test -run TestPrintOK -v` | Partial (update) | pending |
| TBD-08 | all | final | Coverage | coverage | `go test ./... -coverprofile=cover.out && go tool cover -func=cover.out \| grep total` | N/A | pending |

---

## Wave 0 Gaps (tests to create before fixing bugs)

- [ ] `TestExpandAlarmProfiles` in `main_alarm_test.go` — covers BUG-03 (error on missing profile)
- [ ] `TestParseTimedEventTimesNormalization` in `main_test.go` — covers BUG-05 (slash dates, missing zeros)
- [ ] `TestAlarmPromptI18nKeys` in `main_test.go` — covers BUG-02 (i18n keys exist in all locales)
- [ ] Update `TestStripEmoji` — assert accented chars preserved (currently asserts buggy behavior)
- [ ] Update `TestCityToIANA` — assert error return for unknown city

---

## Coverage Gate

```bash
go test ./... -coverprofile=cover.out && go tool cover -func=cover.out | grep total
# Must be ≥ 79.0%
```
