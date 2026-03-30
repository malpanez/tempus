---
phase: 2
slug: first-run-experience
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-30
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None needed (Go convention) |
| **Quick run command** | `go test ./... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1 -race` |
| **Estimated runtime** | ~35 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1 -short`
- **After every plan wave:** Run `go test ./... -count=1 -race`
- **Before `/gsd:verify-work`:** Full suite must be green + coverage ≥ 79%
- **Max feedback latency:** ~35 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| TBD-01 | 02-01 | 1 | CONF-01 | unit | `go test -run TestEnvVarOverrides -v ./internal/config/` | No (Wave 0) | pending |
| TBD-02 | 02-01 | 1 | CONF-02 | unit | `go test -run TestValidateTimezone -v ./internal/config/` | Yes (update) | pending |
| TBD-03 | 02-01 | 1 | CONF-03 | unit | `go test -run TestValidateOutputDir -v ./internal/config/` | No (Wave 0) | pending |
| TBD-04 | 02-02 | 2 | UX-01 | unit | `go test -run TestDetectSystemTimezone -v` | No (Wave 0) | pending |
| TBD-05 | 02-02 | 2 | UX-01 | manual | `tempus init` smoke test (interactive — cannot automate fully) | N/A | pending |
| TBD-06 | 02-03 | 1 | TMPL-01 | unit | `go test -run TestSchoolEventTemplate -v` | No (Wave 0) | pending |
| TBD-07 | 02-03 | 1 | TMPL-02 | unit | `go test -run TestRecruiterMeetingTemplate -v` | No (Wave 0) | pending |
| TBD-08 | 02-03 | 1 | TMPL-03 | unit | `go test -run TestTravelDayTemplate -v` | No (Wave 0) | pending |
| TBD-09 | all | final | Coverage | coverage | `go test ./... -coverprofile=cover.out && go tool cover -func=cover.out \| grep total` | N/A | pending |

---

## Wave 0 Requirements

- [ ] `TestEnvVarOverrides` in `internal/config/config_test.go` — verifies TEMPUS_TIMEZONE, TEMPUS_LANGUAGE, TEMPUS_OUTPUT_DIR, TEMPUS_DATE_FORMAT, TEMPUS_TIME_FORMAT override config file values
- [ ] `TestValidateOutputDir` in `internal/config/config_test.go` — verifies error for nonexistent or non-writable directories (new function)
- [ ] `TestDetectSystemTimezone` in `main_test.go` or new test file — verifies /etc/localtime symlink detection and LANG-based language detection
- [ ] `TestSchoolEventTemplate`, `TestRecruiterMeetingTemplate`, `TestTravelDayTemplate` in `main_test.go` — verify CSV and YAML output contains expected columns

*Existing infrastructure covers env var and template testing needs via stdlib — no new framework required.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `tempus init` interactive wizard (full flow) | UX-01 | survey/v2 interactive prompts cannot be driven in unit tests | Run `tempus init`, confirm auto-detected values shown, tab through timezone/language/output_dir/alarm profile selection, verify config.yaml written |
| Existing config overwrite prompt | UX-01 | Requires existing config on disk + interactive confirm | Run `tempus init` with config already present, verify `[y/N]` prompt appears, verify 'n' exits cleanly |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 35s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
