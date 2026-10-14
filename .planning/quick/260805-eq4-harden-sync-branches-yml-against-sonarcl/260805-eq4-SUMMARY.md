---
phase: quick-260805-eq4
plan: 01
subsystem: infra
tags: [github-actions, ci, workflow_run, sonarcloud, supply-chain, yaml]

# Dependency graph
requires:
  - phase: quick-260805-du1
    provides: "sync-branches.yml on the workflow_run trigger with the three-clause if gate (5dc715f)"
provides:
  - "Four-clause jobs.sync.if gate with an explicit repo-origin check on workflow_run.head_repository"
  - "Inline rationale citing githubactions:S7631 so the intent survives future edits"
affects: [ci, branch-sync, sonarcloud-quality-gate]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "workflow_run jobs gate on head_repository.full_name == github.repository as explicit fork-origin denial"

key-files:
  created: []
  modified:
    - .github/workflows/sync-branches.yml

key-decisions:
  - "Placed the new clause last in the folded scalar to keep the diff to one modified + one added line"
  - "Kept the clause despite being logically redundant with event == 'push' — defense-in-depth and legibility to static analysis"
  - "Condensed the rationale to two comment lines (per plan) while still naming both head_repository and S7631"

patterns-established:
  - "Fork-origin gating: privileged workflow_run jobs pin origin with head_repository.full_name == github.repository"

requirements-completed: [SEC-S7631]

# Metrics
duration: 4min
completed: 2026-08-05
---

# Quick Task 260805-eq4: Harden sync-branches.yml against SonarCloud S7631 Summary

**Added an explicit `head_repository.full_name == github.repository` clause to the `sync-branches.yml` job gate, making it provably unreachable from any fork-originated CI run.**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-08-05T09:37Z
- **Completed:** 2026-08-05T09:41Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments

- `jobs.sync.if` now gates on four `&&`-joined clauses: `conclusion == 'success'`, `head_branch == 'main'`, `event == 'push'`, and `head_repository.full_name == github.repository`.
- Closes SonarCloud `githubactions:S7631` (HIGH / VULNERABILITY, "Make sure that no untrusted code is executed from a fork") flagged on the `workflow_run` trigger.
- Blast radius proven minimal: head region (line 1 → `name: Sync main to develop`) and tail region (`runs-on: ubuntu-slim` → EOF) both diff byte-identical against `5dc715f`. The only change is 2 added comment lines + 1 modified + 1 added clause line.

## Task Commits

Both tasks landed in a single atomic commit — Task 1 was the edit and Task 2 was the proof-and-commit step for that same edit, so splitting would have produced an unverified intermediate commit.

1. **Task 1: Add head_repository clause to the sync job's if gate** — `1338fcf` (fix)
2. **Task 2: Prove no collateral edits, then commit onto the existing branch** — `1338fcf` (fix)

Committed onto the pre-existing branch `fix/sync-branches-race-condition` (PR #17), on top of `5dc715f`. No new branch, no new PR. Not pushed — left for the orchestrator.

## Files Created/Modified

- `.github/workflows/sync-branches.yml` — added the fourth `if` clause pinning the upstream CI run's origin to this repository, plus two comment lines explaining why and citing the rule id.

Resulting gate (parsed by PyYAML into a single well-formed expression):

```
${{ github.event.workflow_run.conclusion == 'success' && github.event.workflow_run.head_branch == 'main' && github.event.workflow_run.event == 'push' && github.event.workflow_run.head_repository.full_name == github.repository }}
```

## Decisions Made

- **Clause placed last** in the folded `>-` scalar. Moving the closing `}}` down one line keeps the diff to exactly one modified line plus one added line.
- **Two comment lines, not three.** The plan specified "two more lines"; the rationale was condensed to fit while still naming both `head_repository` and `githubactions:S7631`.
- **Acknowledged-redundant by design.** The existing `event == 'push'` clause already implies same-repo origin, and the job never executes checked-out tree content (pure `checkout` / `merge --ff-only` / `push`). The new clause is retained anyway as defense-in-depth and as the literal answer to the rule's stated concern.

## Deviations from Plan

None — plan executed exactly as written. No deviation rules fired.

## Issues Encountered

- **`go vet` / `go test` could not run:** Go is not on `PATH` in this environment. Not a blocker and not a deviation — the commit touches exactly one file (`.github/workflows/sync-branches.yml`, confirmed via `git show --stat`) and zero Go files, so the project CLAUDE.md Go gate is trivially unaffected.
- **`actionlint` and `yq` unavailable** (anticipated by the plan). YAML validity was gated with `python3` + PyYAML instead, which also confirmed the folded scalar joins into one valid `${{ ... }}` expression.
- **`.planning/STATE.md` was already modified** in the working tree by the orchestrator (quick-task bookkeeping rows for arn/d0t/di2/du1 + session continuity). Inspected and confirmed out of scope; deliberately NOT staged, per the task constraint that the orchestrator handles docs artifacts. It therefore remains the sole tracked modification in the working tree.

## Verification Results

All six plan verification items pass. Region diffs were run against `5dc715f` explicitly rather than `HEAD`, since post-commit `HEAD` would make them pass trivially.

| # | Check | Result |
|---|-------|--------|
| 1 | `python3 -c "import yaml; yaml.safe_load(...)"` | `YAML_OK` |
| 2 | Four clauses counted outside comment lines | `FOUR_CLAUSES_OK` |
| 3 | Exact expression present exactly once outside comments | `EXACT_EXPR_OK` |
| 4a | Head region diff vs `5dc715f` | `HEAD_REGION_UNCHANGED` |
| 4b | Tail region diff vs `5dc715f` | `TAIL_REGION_UNCHANGED` |
| 5 | Commit touched exactly one file | `ONE_FILE_IN_COMMIT_OK` |
| 6 | Branch still `fix/sync-branches-race-condition` | `BRANCH_OK` |
| + | No `Co-Authored-By:` trailer (repo policy) | `NO_TRAILER_OK` |

Untouched and byte-identical to `5dc715f`: `name:`, the `on:`/`workflow_run` block, the concurrency comment and `concurrency:` block, `jobs.sync.name`, `runs-on`, `timeout-minutes`, `permissions`, the checkout step (including its pinned SHA `34e114876b0b11c390a56381ad16ebd13914f8d5` and inline comment), and the `env:`/`run:` body of the fast-forward step.

## Threat Model Outcome

| Threat ID | Disposition | Outcome |
|-----------|-------------|---------|
| T-eq4-01 | mitigate | **Mitigated.** `head_repository.full_name == github.repository` added to the gate; `jobs.sync` (which holds `contents: write`) is now unreachable from any fork-originated CI run. |
| T-eq4-02 | accept | Unchanged. Job still executes no tree content — no build, no install, no script invocation. |
| T-eq4-03 | accept | Unchanged. `head_repository` is GitHub-generated event metadata, not attacker-supplied input. |
| T-eq4-SC | accept | Unchanged. YAML-only edit; no `go.mod` / lockfile touched, no package-manager installs. |

No new threat surface introduced — the change only narrows an existing gate.

## Known Stubs

None.

## User Setup Required

None — no external service configuration required. PR #17 picks up the new commit automatically once pushed; no PR-level action needed.

## Next Phase Readiness

- Branch `fix/sync-branches-race-condition` now carries two commits (`5dc715f`, `1338fcf`) and is ready to push.
- **Not pushed** — deliberately left to the orchestrator per this session's "no push without explicit confirmation" policy.
- On merge, SonarCloud should clear `githubactions:S7631` on `.github/workflows/sync-branches.yml`.

## Self-Check: PASSED

- `.github/workflows/sync-branches.yml` — FOUND (modified, four-clause gate present)
- Commit `1338fcf` — FOUND on `fix/sync-branches-race-condition`
- `.planning/quick/260805-eq4-harden-sync-branches-yml-against-sonarcl/260805-eq4-SUMMARY.md` — FOUND

---
*Quick task: 260805-eq4*
*Completed: 2026-08-05*
