---
phase: quick-260907-q7c
plan: 01
subsystem: security-hardening
tags: [security, cwe-93, cwe-22, supply-chain, least-privilege, dependencies]
requires: []
provides:
  - calendar.SanitizeStructuralValue (CRLF guard for verbatim ICS property values)
  - basename-collapsed filename_tmpl rendering in deriveTemplateFilename
  - 0600 config.yaml on POSIX
  - explicit GITHUB_TOKEN scope on every ci.yml and security.yml job
  - digest-pinned Docker base images
affects:
  - internal/calendar
  - internal/cli
  - internal/config
  - .github/workflows
  - Dockerfile
tech-stack:
  added: [go-viper/mapstructure/v2 v2.4.0 (indirect, via viper v1.21.0)]
  patterns:
    - "Narrow structural sanitizer distinct from escapeText for values with their own micro-syntax"
    - "SetConfigPermissions + explicit os.Chmod for idempotent file-mode tightening"
    - "Manifest-index digest pins (not single-platform) for multi-arch Docker releases"
key-files:
  created: []
  modified:
    - internal/calendar/calendar.go
    - internal/calendar/calendar_test.go
    - internal/cli/batch.go
    - internal/cli/create.go
    - internal/cli/template.go
    - internal/cli/coverage_gaps_test.go
    - internal/config/config.go
    - internal/config/config_test.go
    - .github/workflows/ci.yml
    - .github/workflows/security.yml
    - Dockerfile
    - .dockerignore
    - go.mod
    - go.sum
decisions:
  - "SanitizeStructuralValue collapses CR/LF to a space rather than escaping — RRULE and mailto: values carry their own ; / , / : micro-syntax that escapeText would corrupt"
  - "CRLF replaced before bare CR and LF so \\r\\n collapses to one space, not two"
  - "Sanitize-then-trim at ingestion so a whitespace-only RRULE collapses to empty, not to a lone space"
  - "Interactive path's user-typed filename deliberately NOT basename-collapsed — local operator input, and collapsing would break a deliberate subdir/name.ics"
  - "Explicit os.Chmod after WriteConfigAs — viper's SetConfigPermissions only applies at file creation, so a pre-existing 0644 config.yaml would never be tightened"
  - "ci.yml gets a top-level permissions block (uniformly read-only workflow); security.yml gets per-job blocks (gosec/codeql legitimately need security-events: write)"
  - "Docker pins use the manifest INDEX digest, not a single-platform image digest — release.yml builds linux/amd64,linux/arm64"
metrics:
  duration: ~35min
  completed: 2026-09-07
  tasks: 4
  files: 14
  commits: 4
  coverage: 89.1%
---

# Quick Task 260907-q7c: Cyber Neo Security Audit Remediation Summary

Closed CN-001 through CN-009 from the 2026-09-07 Cyber Neo audit in four atomic commits on `fix/cyber-neo-security-audit`: a CRLF-injection guard for RRULE/ATTENDEE, basename collapse of `filename_tmpl`, 0600 config permissions, least-privilege `GITHUB_TOKEN` scopes on all eleven CI jobs, a pinned govulncheck, digest-pinned Docker bases, and cobra/pflag/viper brought current.

## Commits

| # | Hash | Scope | Findings closed |
|---|------|-------|-----------------|
| 1 | `5c87cda` | `fix(security)` | CN-002, CN-005, CN-008 |
| 2 | `e2ea83d` | `ci(security)` | CN-001, CN-003, CN-007 |
| 3 | `6ae1ddc` | `build(docker)` | CN-004, CN-006 |
| 4 | `baee611` | `chore(deps)` | CN-009 |

`git log --oneline main..HEAD` shows exactly these four, in this order, with the four required scopes. Nothing was committed to `main`; nothing was pushed; `--no-verify` was never used.

CN-010 (gosec/trufflehog actions pinned to a full commit SHA with an imprecise trailing `# master` / `# main` comment) was left untouched, as the plan explicitly scoped it out. The Info-level unbounded-batch-ingestion item was likewise left alone.

## Task 1 — CRLF injection, path traversal, config mode (`5c87cda`)

Executed TDD. All new tests were written first and confirmed failing (`undefined: SanitizeStructuralValue` in `calendar`; 10 failures across `cli` and `config`) before any source edit.

**CN-002 (CWE-93, Medium).** Added `calendar.SanitizeStructuralValue(s string) string` immediately after `escapeText`. It replaces `\r\n`, then `\r`, then `\n` with a single space — CRLF first, so `a\r\nb` becomes `a b` rather than `a  b`. It is deliberately *not* `escapeText`: `FREQ=WEEKLY;BYDAY=MO,TU` must survive byte-for-byte, and `escapeText` would mangle it to `FREQ=WEEKLY\;BYDAY=MO\,TU`. Applied at both raw-write sinks (`writeRecurrenceProperties`, `writeOptionalProperties`) and, as defense in depth, at ingestion in `batch.go` `configureBatchEvent`, `create.go` `configureEvent`, and `create.go` `addEventAttendees`. Ingestion sanitizes *then* trims, so a whitespace-only value collapses to empty rather than to a lone space.

**CN-005 (CWE-22, Medium).** `deriveTemplateFilename` returned the rendered `filename_tmpl` verbatim. Backslashes are now normalised to `/` first (so a Linux run also collapses `..\..\etc\passwd`), then `filepath.Base(filepath.Clean(...))` reduces it to a bare basename. A degenerate result (`.`, `..`, `/`) falls through to the existing `Slugify` fallbacks instead of being returned. Neither `internal/templates/ddrender.go` nor the interactive path's user-typed `finalName` was touched, per the plan.

**CN-008 (Low).** `Config.Save()` now calls `viper.SetConfigPermissions(0o600)` and follows `WriteConfigAs` with an explicit `os.Chmod(configFile, 0o600)`, both errors wrapped with `fmt.Errorf(...: %w)`. The Chmod is load-bearing: viper hands the mode to `os.OpenFile`, which only honours it at *creation*, so a `config.yaml` predating this change would have stayed 0644 forever.

**Tests added:** `TestSanitizeStructuralValue` (9 table cases), `TestToICSRejectsStructuralInjection`, `TestDeriveTemplateFilenameRejectsTraversal` (4 cases), `TestDeriveTemplateFilenameRejectsDegenerateBase` (5 cases), `TestSaveConfigPermissions` (fresh Save + rewrite, `runtime.GOOS == "windows"` skip since the CI matrix runs windows-latest on PRs).

## Task 2 — Workflow token scopes and govulncheck pin (`e2ea83d`)

**CN-001.** `ci.yml` declared no `permissions:` at all, so all five jobs inherited the repository default. Added a top-level `permissions: contents: read`. Top-level is right here because the workflow is uniformly read-only; Codecov authenticates with `CODECOV_TOKEN` and its PR comments come from the Codecov GitHub App, not `GITHUB_TOKEN`.

**CN-007.** Added job-level `contents: read` to the four `security.yml` jobs that lacked it: `dependency-review`, `govulncheck`, `secrets-scan`, `security-summary`. `gosec` and `codeql` were left alone — both correctly retain `security-events: write` for their SARIF uploads. `dependency-review` has no `comment-summary-in-pr`, so it never needs `pull-requests: write`; `security-summary` writes only to `$GITHUB_STEP_SUMMARY`, a file descriptor that consumes no token scope.

**CN-003.** `go install golang.org/x/vuln/cmd/govulncheck@latest` → `@v1.7.0`. Independently re-confirmed against the Go module proxy during execution, not merely trusted from the plan: `proxy.golang.org/golang.org/x/vuln/@latest` returns `{"Version":"v1.7.0","Time":"2026-08-13T18:01:04Z"}`, and v1.7.0 is the highest entry in the version list. Pinning matters because `@latest` re-resolves on every run, so a compromised upstream release would execute in CI with no review gate.

## Task 3 — Docker digest pins and build context (`6ae1ddc`)

**CN-004. Resolved digests, date resolved 2026-09-07:**

| Image | Digest |
|-------|--------|
| `golang:1.26-alpine` | `sha256:ce864e7223ac17b1775e6fd0b4c0db580c2eb50e7953a427916379e4b92a1628` |
| `alpine:3.21` | `sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d` |

Both are **manifest-index** digests, not single-platform image digests — `release.yml` builds `linux/amd64,linux/arm64`, so a single-arch pin would break every future arm64 release. Verified two independent ways, which agreed exactly on both images:

1. `docker buildx imagetools inspect <image>:<tag> --format '{{json .Manifest.Digest}}'`
2. `docker pull <image>:<tag>` followed by `docker inspect --format='{{index .RepoDigests 0}}'`

Both indexes were then re-inspected *by digest* and confirmed to still serve `linux/amd64` and `linux/arm64/v8` (alongside 386, arm/v6, arm/v7, ppc64le, riscv64, s390x). No Dockerfile change beyond the two `FROM` lines and their explanatory comments: the multi-stage layout, non-root user and labels are untouched.

**CN-006.** Appended `.planning/`, `.venv/` and `.coverage/` to `.dockerignore` under a new comment header. None reaches the final image today thanks to the multi-stage build, so this is defense in depth against a future single-stage regression, plus a smaller context upload.

Verified with `docker build -t tempus:security-audit-verify .` and `docker run --rm tempus:security-audit-verify --help`. **Nothing was pushed to any registry; no registry login was performed.**

## Task 4 — Dependency bumps (`baee611`)

Bumped one at a time, in the mandated order, with explicit pinned targets (`go get <mod>@<version>`, never `go get -u`). pflag first because cobra v1.10.2 requires pflag ≥ v1.0.9; doing cobra first would have dragged pflag along silently and muddied attribution.

| Module | Before | After | `go build` | `go test` | Reverted? |
|--------|--------|-------|-----------|-----------|-----------|
| `github.com/spf13/pflag` | v1.0.5 | **v1.0.10** | pass | pass | No |
| `github.com/spf13/cobra` | v1.8.0 | **v1.10.2** | pass | pass | No |
| `github.com/spf13/viper` | v1.18.2 | **v1.21.0** | pass | pass | No |

**No module was reverted. There is no failure text to report.** Each bump was individually gated on `go build ./...` && `go test ./...` before proceeding to the next; all three passed on the first attempt (1976 tests, 15 packages). `go mod tidy` ran after all three, and build + test were re-run clean afterwards.

Both plan-flagged risks were checked explicitly rather than left to surface as a red test:

- **viper's decoder swap** happened exactly as predicted: `github.com/mitchellh/mapstructure v1.5.0` is gone from the indirect block, replaced by `github.com/go-viper/mapstructure/v2 v2.4.0`. The `mapstructure:"..."` tags on `config.Config` (config.go:17-27) needed no change — the fork reads them identically. Collateral indirect upgrades: fsnotify v1.7.0→v1.9.0, go-toml/v2 v2.1.0→v2.2.4, locafero v0.4.0→v0.11.0, afero v1.11.0→v1.15.0, cast v1.6.0→v1.10.0, conc v0.3.0→v0.3.1-pre.
- **`viper.SetConfigPermissions`** (added to `Save()` in Task 1) is still present in v1.21.0 — `go build ./...` succeeds and `TestSaveConfigPermissions` passes against the new module.

## Verification

Full `<verification>` block from the plan, run from the repo root at final state:

| # | Check | Result |
|---|-------|--------|
| 1 | `gofmt -l .` prints nothing | **PASS** |
| 2 | `go vet ./...` exits 0 | **PASS** |
| 3 | `go build ./...` exits 0 | **PASS** |
| 4 | `go test ./...` exits 0 incl. new tests | **PASS** (1976 tests, 15 packages; all 21 new subtests PASS) |
| 5 | Total coverage ≥ 79% | **PASS — 89.1%** |
| 6 | PyYAML parses both workflows; every security.yml job declares `permissions` | **PASS** (all 6 jobs) |
| 7 | `grep -c 'govulncheck@latest'` is 0 | **PASS** |
| 8 | Both Dockerfile FROM lines match `^FROM (golang\|alpine):[^@]+@sha256:[0-9a-f]{64}` | **PASS** (2/2) |
| 9 | `docker build` succeeds; nothing pushed | **PASS** (image also runs `--help`) |
| 10 | `git log --oneline main..HEAD` shows 4 commits with the required scopes | **PASS** |
| 11 | Negative control on the RRULE sink | **PASS — see below** |

**Coverage** was measured with the plan's specified command shape, matching ci.yml:288-292: `go test ./... -coverprofile=.coverage/profile.out -covermode=atomic` then `go tool cover -func`. Result **89.1%**, comfortably above the 79% gate. `.coverage/` is already in `.gitignore` and was not committed. Note that CI's number will be at least this: ci.yml:70 additionally merges instrumented golden-E2E coverage via `GOLDEN_COVERDIR`, which only adds.

### Verification step 11 — negative control: PERFORMED (twice)

Confirmed the CN-002 regression test is capable of failing, so it proves something.

Procedure: temporarily reverted the RRULE sink in `internal/calendar/calendar.go:268` from `writeProp(b, "RRULE", SanitizeStructuralValue(e.RRule))` back to `writeProp(b, "RRULE", e.RRule)`, ran `go test ./internal/calendar/ -run TestToICSRejectsStructuralInjection`, then restored the fix from a pre-edit copy and re-ran.

- **Run A — during Task 1, before committing.** With the fix reverted: `FAIL` at `calendar_test.go:535` ("injected value started a new line") and `:538` ("expected sanitized RRULE value on one line"). Restored → `ok`.
- **Run B — after all four commits, at final state (post-dependency-bump).** With the fix reverted: `--- FAIL: TestToICSRejectsStructuralInjection`, same two assertions. Restored → `ok tempus/internal/calendar`.

After restoring, `git status --short --untracked-files=no` shows only the pre-existing, unrelated ` M .planning/STATE.md` — the source tree is identical to `HEAD`, so no negative-control residue was left behind.

## Deviations from Plan

**None affecting behavior.** Every live line number, function signature, ordering constraint and edge case described in the plan's `<interfaces>` matched HEAD exactly; no source drift was found. The following are additions *within* the plan's stated intent, not departures from it:

1. **Extra test cases beyond the `<behavior>` minimum.** `TestSanitizeStructuralValue` gained four cases past the six specified (`mailto:` value preserved, literal backslash-n preserved, bare newline, trailing CRLF). `TestDeriveTemplateFilenameRejectsTraversal` was made table-driven with an absolute-path case (`/etc/passwd` → `passwd.ics`) and a benign-name control proving the collapse is not over-broad; the degenerate-basename test covers all five of `..`, `.`, `/`, `../..`, `../` rather than just `..`. Strictly additive coverage of the same behavior.
2. **`base != "/"` alongside `base != string(filepath.Separator)`** in the traversal guard. Redundant on POSIX, but keeps the guard correct after the `\` → `/` normalisation on a Windows build, where `filepath.Separator` is `\`.
3. **Independent re-verification of govulncheck v1.7.0** against the module proxy rather than trusting the planning-time claim. It confirmed the plan (v1.7.0, tagged 2026-08-13, current latest). A supply-chain pin is not worth taking on faith.

No Rule 4 (architectural) situations arose. No auth gates. No package-install failures — Task 4 upgraded three already-vetted existing direct dependencies, so the Package Legitimacy Gate did not apply.

## Known Stubs

None. No placeholder values, empty returns, or unwired data paths were introduced.

## Threat Flags

None. No new network endpoints, auth paths, file-access patterns, or trust-boundary schema changes were introduced — every change in this plan *narrows* existing surface.

## Self-Check: PASSED

Files (all modified, none created):

- FOUND: `internal/calendar/calendar.go` — contains `func SanitizeStructuralValue`
- FOUND: `internal/calendar/calendar_test.go` — contains `SanitizeStructuralValue`
- FOUND: `internal/cli/batch.go` — contains `calendar.SanitizeStructuralValue`
- FOUND: `internal/cli/create.go` — contains `calendar.SanitizeStructuralValue`
- FOUND: `internal/cli/template.go` — contains `filepath.Base(filepath.Clean`
- FOUND: `internal/cli/coverage_gaps_test.go`
- FOUND: `internal/config/config.go` — contains `SetConfigPermissions(0o600)` and `os.Chmod(configFile, 0o600)`
- FOUND: `internal/config/config_test.go`
- FOUND: `.github/workflows/ci.yml` — contains top-level `permissions:`
- FOUND: `.github/workflows/security.yml` — contains `govulncheck@v1.7.0`
- FOUND: `Dockerfile` — both FROM lines contain `@sha256:`
- FOUND: `.dockerignore` — contains `.planning/`
- FOUND: `go.mod`, `go.sum`

Commits (all present in `git log main..HEAD`):

- FOUND: `5c87cda` fix(security): sanitize CRLF injection, filename traversal, and config mode
- FOUND: `e2ea83d` ci(security): scope workflow token permissions and pin govulncheck
- FOUND: `6ae1ddc` build(docker): pin base images to digests and harden build context
- FOUND: `baee611` chore(deps): bump cobra, pflag and viper to latest

All `must_haves.key_links` patterns from the plan frontmatter were grep-confirmed against the working tree. No file deletions occurred in any of the four commits.
