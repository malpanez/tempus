---
phase: quick
plan: 260805-d0t
subsystem: infra
tags: [ci, github-actions, sonarcloud, static-analysis, supply-chain]

# Dependency graph
requires:
  - phase: quick/260403-skg
    provides: "sonar.exclusions tuning in sonar-project.properties (left byte-identical here)"
  - phase: quick/260405-ixt3
    provides: "sonar.cpd.exclusions + minimumTokens tuning (left byte-identical here)"
provides:
  - "CI workflow set reduced to the three that can actually pass: ci.yml, release.yml, security.yml"
  - "SonarCloud analysis consolidated on Automatic Analysis (single mode, no CI conflict)"
  - "SonarSource/sonarqube-scan-action and its GITHUB_TOKEN + SONAR_TOKEN exposure removed from the CI supply chain"
affects: [ci, release-engineering, sonarcloud-config]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Single SonarCloud analysis mode — Automatic Analysis only, never alongside CI-based scanning"

key-files:
  created: []
  modified:
    - .github/workflows/sonarcloud.yml  # deleted

key-decisions:
  - "Delete sonarcloud.yml rather than disabling Automatic Analysis in the SonarCloud UI — decision locked pre-flight, alternative explicitly declined"
  - "Leave the two now-inert report-path keys in sonar-project.properties (lines 14, 17) — single-file-deletion constraint; SonarCloud ignores absent report paths"
  - "Leave the dormant SONAR_TOKEN repo secret in place — revocation is a GitHub repo-settings action, out of scope (threat T-D0T-01, disposition accept)"

patterns-established:
  - "Byte-identity proof: git hash-object on untouched files before and after the change, compared explicitly rather than assumed"
  - "Comment-stripped grep gate: full-line YAML comments removed before counting, so a stale comment cannot satisfy the residual-reference check"

requirements-completed: []

# Metrics
duration: 2min
completed: 2026-08-05
---

# Quick 260805-d0t: Remove Redundant SonarCloud CI Workflow Summary

**Deleted `.github/workflows/sonarcloud.yml`, the CI-based Sonar scan that failed 71/71 runs since 2025-12-15 because SonarCloud Automatic Analysis was already active and the two modes are mutually exclusive.**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-08-05T08:29:12Z
- **Completed:** 2026-08-05T08:31:00Z
- **Tasks:** 2 (1 mutating, 1 read-only audit)
- **Files modified:** 1 (deleted)

## Accomplishments

- Removed a workflow with a 100% failure rate from day one — noise that could never succeed while Automatic Analysis stays on.
- Proved by audit that no surviving workflow had a hidden dependency on the deleted job (`needs:`, `SONAR_TOKEN` reuse, or composite action call).
- Dropped a third-party action (`SonarSource/sonarqube-scan-action`) and its `GITHUB_TOKEN` + `SONAR_TOKEN` exposure from the CI supply chain — a net reduction in attack surface (threat T-D0T-02, disposition mitigate).
- Kept quality reporting intact: Automatic Analysis still posts the passing `SonarCloud Code Analysis` check, and the README badge resolves by project key, independent of analysis mode.

## Task Commits

1. **Task 1: Remove sonarcloud.yml from the working tree and the git index** — `d153614` (chore)
2. **Task 2: Audit surviving workflows for residual SonarCloud CI wiring** — read-only audit, no commit (correct: this task modifies nothing by design)

Single commit on branch `fix/remove-sonarcloud-workflow`, based on `origin/main` at `3910848`. Not pushed — left for the orchestrator/user per the no-push-without-confirmation policy.

**Plan metadata:** handled separately by the orchestrator (SUMMARY.md, STATE.md, PLAN.md deliberately excluded from `d153614`).

## Files Created/Modified

- `.github/workflows/sonarcloud.yml` — **deleted** (53 lines). Ran the Sonar scan on push/PR to `main` plus `workflow_dispatch`; generated coverage + golangci-lint reports purely as scanner input.

Untouched and verified byte-identical via `git hash-object` before and after:

| File | Hash (unchanged) |
|------|------------------|
| `.github/workflows/ci.yml` | `86c93171b5416c7dd4789c96d7823c86a9b621fc` |
| `.github/workflows/release.yml` | `25386b9363c91bab478e04870d816282ce5426ce` |
| `.github/workflows/security.yml` | `6ed536dccc807f8f3c69df5db4457c95337decef` |
| `sonar-project.properties` | `e52740c7a77d51a12073027cc09111be7c741251` |
| `README.md` | `96a9755bf04b9a8e578cfdb590121153d7104264` |

All `.planning/quick/*` files untouched (accurate history of past CPD fixes — records, not config).

## Verification

**Task 1 gates:**
- `WORKTREE_OK` — file absent from disk.
- `INDEX_OK` — `git ls-files --error-unmatch` no longer matches the path.
- `git diff --cached --name-status` — exactly one line: `D .github/workflows/sonarcloud.yml`.

**Task 2 gates:**
- Hard gate printed `0` — zero functional sonar references across surviving workflows, with full-line YAML comments stripped first.
- Informational grep printed `NO_SONAR_REFERENCES_REMAIN` — not even a comment mentions Sonar.
- `ls -1 .github/workflows/` lists exactly `ci.yml`, `release.yml`, `security.yml`.

**Post-commit:**
- `git diff --diff-filter=D HEAD~1 HEAD` — the deletion is exactly the intended path, nothing collateral.
- Commit message byte-exact against the plan's specified text (em dash stored as correct UTF-8 `e2 80 94`); no `Co-Authored-By` trailer, per the developer's global commit policy.

**Go gates deliberately skipped.** Per the plan's `<verification>` section, this change touches zero Go files — the diff is one deleted YAML workflow, so `go vet ./...` and `go test ./...` cannot be affected. CI runs both against the resulting commit anyway, and the 79% coverage floor stays enforced by `ci.yml`'s `Coverage Check` job.

## Decisions Made

None beyond the plan — the decision was locked pre-flight and executed as written.

## Deviations from Plan

None — plan executed exactly as written. No deviation rules fired.

## Issues Encountered

**One benign discrepancy, surfaced rather than silently absorbed:**

Task 1's `<done>` criterion expects `git status --porcelain --untracked-files=no` to show the deletion "and nothing else". It also shows ` M .planning/STATE.md`. That modification **pre-dates this task** — it was present in the working tree before `git rm` ran (confirmed by the status check taken before any mutation) and is orchestrator-owned docs territory, explicitly excluded from my commit by the task constraints.

It was left unstaged and is absent from `d153614`. The load-bearing criterion — the *staged* diff being exactly one deletion (success_criteria line 178) — holds without qualification. No action taken.

## Known Consequences (accepted, carried forward from the plan)

`sonar-project.properties` lines 14 and 17 now point at artifacts nothing generates:

- `sonar.go.coverage.reportPaths=.coverage/profile.out,.coverage/e2e.out`
- `sonar.go.golangci-lint.reportPaths=golangci-lint-report.xml`

Both keys are inert. **Nothing regresses:** because the workflow failed 71/71 times, SonarCloud never actually received these reports — the config was already dead. Neither metric was ever a merge gate (coverage lives in `ci.yml`'s `Coverage Check`, lint in its `Lint` job, both required checks). SonarCloud ignores report paths whose files are absent, so the dead keys are harmless. Removing them is optional future cleanup, kept out of scope here to preserve the single-file-deletion constraint and the live `sonar.exclusions` / `sonar.cpd.exclusions` values.

## Threat Flags

None. This change adds no network endpoints, auth paths, file access patterns, or schema changes — it only removes CI surface. The plan's register was fully addressed:

- **T-D0T-01** (dormant `SONAR_TOKEN`) — accepted. Now referenced by no workflow. Revoking it is a GitHub repo-settings action; noted for the developer's hygiene backlog.
- **T-D0T-02** (third-party action privilege) — mitigated by the deletion itself.
- **T-D0T-03** (half-removed wiring) — mitigated; Task 2's comment-stripped gate returned `0`.
- **T-D0T-SC** (package legitimacy) — n/a; no packages installed, no dependencies added.

## Known Stubs

None — no code was written.

## User Setup Required

No file-level setup. Two optional GitHub repo-settings actions for the developer, both out of scope here:

1. **Revoke the `SONAR_TOKEN` repo secret** — dormant now that no workflow reads it (T-D0T-01).
2. **Confirm branch protection is unaffected** — `SonarCloud Scan` was never in the required-checks list for `main` (required: Lint, Coverage Check, Functional Tests, Test (ubuntu-latest, 1.26), Build (ubuntu-latest, 1.26), Test (windows-latest, 1.26)), so no merge behavior changes.

## Post-Merge Observations

Not verifiable locally. On the next PR, expect:

- The `SonarCloud Scan` check no longer appears in the checks list.
- The external `SonarCloud Code Analysis` check still appears and still passes.
- The README Sonar quality-gate badge still renders a real status.

## Next Readiness

- Branch `fix/remove-sonarcloud-workflow` holds one clean commit, ready to push and open a PR (mirroring the flow used for quick task 260805-arn / PR #14). **Not pushed** — awaiting explicit confirmation.
- No blockers.

## Self-Check: PASSED

- `.github/workflows/sonarcloud.yml` absent from worktree — verified.
- Path absent from git index — verified via `git ls-files --error-unmatch`.
- Commit `d153614` exists in `git log` — verified.
- Commit contains exactly one deletion of the target path — verified.
- Five untouched files byte-identical by hash — verified.
- SUMMARY.md written to the path specified in the plan's `<output>` — verified.

---
*Quick task: 260805-d0t*
*Completed: 2026-08-05*
