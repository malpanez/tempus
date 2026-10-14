---
phase: quick-260805-du1
plan: 01
subsystem: infra
tags: [github-actions, ci, workflow_run, branch-protection, git]

# Dependency graph
requires:
  - phase: quick-260805-di2
    provides: the one-way main -> develop ff-only sync-branches.yml this plan repairs
provides:
  - CI-gated sync trigger — sync-branches.yml fires on `workflow_run` completion of CI, never on the raw push
  - TOCTOU-proof merge target — checkout ref and merge target both pinned to workflow_run.head_sha
  - Trigger-source gating — sync only fires for CI runs whose own event was a push to main
affects: [ci, release-workflow, branch-protection]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "workflow_run chaining to sequence a write-scoped workflow behind a validation workflow"
    - "Pin downstream git operations to workflow_run.head_sha, never a moving branch ref"
    - "Thread event payload values into shell via step-level env:, never ${{ }} interpolation in run:"

key-files:
  created: []
  modified:
    - .github/workflows/sync-branches.yml

key-decisions:
  - "Trigger is workflow_run on [\"CI\"] types [completed] and nothing else — no push, no dispatch, no develop (D-01)"
  - "Job gated on all three of conclusion=='success', head_branch=='main', event=='push' — the event clause excludes ci.yml's weekly cron and workflow_dispatch runs (D-02)"
  - "Checkout ref and merge target both pin to workflow_run.head_sha, never bare origin/main (D-03)"
  - "head_sha reaches the shell via step-level env HEAD_SHA, quoted as \"$HEAD_SHA\" — no ${{ }} in any run body (T-QUICK-06)"

patterns-established:
  - "workflow_run gating: a write-scoped workflow that must observe another workflow's status checks triggers on that workflow's completion, not on the shared upstream event"
  - "SHA pinning: chained workflows operate on the validated head_sha, not the branch tip, so a later unvalidated push cannot ride along"

requirements-completed: [QUICK-260805-du1]

# Metrics
duration: 4min
completed: 2026-08-05
---

# Quick Task 260805-du1: Fix sync-branches.yml race condition Summary

**sync-branches.yml now triggers on CI's `workflow_run` completion instead of the raw push to main, and pins both checkout and merge to `workflow_run.head_sha`, so develop's required-status-check gate is already satisfied when the push lands and no unvalidated commit can ride along.**

> **Scope note — this plan is NOT fully complete.** Tasks 1 and 2 are done and committed.
> **Task 3 (realigning `develop`'s required status check contexts) was deliberately NOT executed** — see
> [Task 3: pending owner decision](#task-3-pending-owner-decision--not-executed) below. Until Task 3 is
> applied, the sync push will still be rejected; only the error message changes.

## Performance

- **Duration:** ~4 min
- **Started:** 2026-08-05T09:06Z
- **Completed:** 2026-08-05T09:10Z
- **Tasks:** 2 of 3 (Task 3 intentionally deferred — checkpoint:decision)
- **Files modified:** 1

## Accomplishments

- Replaced the `push: branches: [main]` trigger with `workflow_run: workflows: ["CI"], types: [completed]`, so the sync starts only after every CI check has posted a final conclusion for the commit. This removes the timing race that made commit `c8ad278` fail with `GH006 ... Required status check "Lint" is in progress.`
- Added a three-condition job gate (`conclusion == 'success'` && `head_branch == 'main'` && `event == 'push'`). The `event` clause is what keeps ci.yml's weekly Sunday cron, its `workflow_dispatch` runs and every `pull_request` run from firing a pointless sync.
- Closed the TOCTOU hole: both `actions/checkout`'s `ref:` and the merge target now resolve to `github.event.workflow_run.head_sha`. Under `workflow_run`, a checkout with no `ref` would land on the default branch's *current* tip — potentially a newer commit whose CI has not finished, which is the exact class of bug being fixed.
- Hardened the shell boundary: `head_sha` is threaded through step-level `env: HEAD_SHA:` and referenced as `"$HEAD_SHA"`; no `${{ }}` appears in any run body.
- Preserved byte-for-byte everything reviewed in PR #16 (D-05): the `concurrency` block and its three-line comment, `runs-on: ubuntu-slim`, `timeout-minutes: 5`, job-level `permissions: contents: write` with no workflow-level block, the `actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5 # v4` pin, `fetch-depth: 0`, `token: ${{ secrets.GITHUB_TOKEN }}`, and the `--ff-only` brace-block failure handler.

## Task Commits

1. **Task 1: Retrigger the sync off CI completion and pin it to the validated SHA** — `5dc715f` (fix)
   - `fix(ci): trigger sync-branches on CI completion and pin to the validated SHA`
   - 1 file changed, 18 insertions(+), 4 deletions(-). No `Co-Authored-By` trailer, per repo commit policy.
2. **Task 2: Gate the workflow against every stated constraint** — no commit (verification-only task; the gate validates Task 1's file and produced no edits).
3. **Task 3: develop branch protection realignment** — NOT EXECUTED. See below.

**Plan metadata:** committed separately by the orchestrator (this executor was scoped to the workflow file only).

## Files Created/Modified

- `.github/workflows/sync-branches.yml` — trigger switched from `push` to `workflow_run` on CI; three-condition job `if:` gate added; checkout pinned to `head_sha`; merge target changed from `origin/main` to `"$HEAD_SHA"` via step-level env.

## Verification

Every structural assertion in the plan's Task 2 gate passed on the first run (items 1-11 and 14 of `<verification>`; items 12-13 belong to Task 3 and were not run):

| # | Assertion | Result |
|---|-----------|--------|
| 1 | `sync-branches.yml` parses as valid YAML | PASS |
| 2 | `on:` has exactly one key `workflow_run`, `workflows: ["CI"]`, `types: [completed]`; no `push`, no `develop` | PASS |
| 3 | `ci.yml`'s top-level `name:` confirmed to be `CI` (read-only cross-check) | PASS |
| 4 | Exactly one job, `if:` contains all three required conditions | PASS |
| 5 | Checkout pinned to `actions/checkout@34e1148...` with `fetch-depth: 0` and `ref: ${{ github.event.workflow_run.head_sha }}` | PASS |
| 6 | Merge step declares `env: HEAD_SHA:`; no run body contains `${{` | PASS |
| 7 | Run body contains `git merge --ff-only "$HEAD_SHA"` and does NOT contain `origin/main` | PASS |
| 8 | No `git checkout main`, `git push origin main`, `--force`, `push -f`, `rebase`, `--no-ff`, `-X ours/theirs` | PASS |
| 9 | Run body contains `git checkout develop`, `exit 1` on ff failure, `git push origin develop` | PASS |
| 10 | `runs-on: ubuntu-slim`, `timeout-minutes: 5`, job-level `permissions: contents: write`, no workflow-level `permissions`, `concurrency` preserved | PASS |
| 11 | `origin/develop` (`3910848`) is an ancestor of `origin/main` (`c8ad278`) | PASS |
| 14 | `git status --porcelain -- .github/` lists only `sync-branches.yml` | PASS |

Post-commit checks: the commit contains no file deletions, and `ci.yml`, `security.yml`, `release.yml` and every other file are byte-identical to their pre-plan state.

## Task 3: pending owner decision — NOT executed

**Task 3 is a `checkpoint:decision` with `gate="blocking"`. This executor was explicitly instructed not to run it; the repo owner's choice is being collected separately by the orchestrator. No `gh api` call was made by this executor — `develop`'s and `main`'s branch protection are untouched.**

Why it still matters (from the plan's `<blocking_finding>`): the race is only half the reason the push was rejected. `develop`'s `required_status_checks` list contains

```
Test (ubuntu-latest, 1.26)
Build (ubuntu-latest, 1.26)
```

but `ci.yml`'s conditional runner matrix (lines 21 and 113) evaluates to `["ubuntu-slim"]` on a `push` event, so a push-to-main commit only ever reports `Test (ubuntu-slim, 1.26)` / `Build (ubuntu-slim, 1.26)`. The two `ubuntu-latest` contexts are producible only by `pull_request` and tag runs, so on a push-to-main commit they sit permanently in the "Expected" state and block the ref update.

**Consequence of the current state:** with Tasks 1 and 2 alone, the sync's rejection message changes from
`Required status check "Lint" is in progress.` to
`Required status check "Test (ubuntu-latest, 1.26)" is expected.`
and the push still fails on every run. **Tasks 1 and 2 are necessary but not sufficient.**

The two options awaiting the owner's selection:

- **option-a** — realign `develop`'s contexts to the five a push-to-main actually produces (`PATCH .../branches/develop/protection/required_status_checks` with `strict: true` and contexts `Lint`, `Coverage Check`, `Functional Tests`, `Test (ubuntu-slim, 1.26)`, `Build (ubuntu-slim, 1.26)`). Keeps the safety property; couples `develop`'s protection to `ci.yml`'s `ubuntu-slim` matrix label.
- **option-b** — drop required status checks from `develop` entirely (`DELETE .../branches/develop/protection/required_status_checks`). Cannot go stale; loses the belt-and-braces re-gate on a commit that is already green on `main`.

**Recording obligation carried forward:** branch protection changes leave no trace in git history. Whoever executes Task 3 MUST append the chosen option and the exact `gh api` call to this SUMMARY — that is the only record. If **option-a** is chosen, also record the known follow-up: `develop`'s required contexts become coupled to `ci.yml`'s `ubuntu-slim` matrix label, and a future change to that label silently re-breaks the sync with no error surfaced anywhere.

**Timing constraint (from the plan):** Task 3 must be applied BEFORE this change merges to `main`. The merge itself triggers CI, whose completion fires the new `workflow_run` sync — that first run is the live proof, and it will fail if the contexts are still misaligned.

<!-- TASK 3 RECORD:
Option chosen: option-b (drop required status checks from develop entirely)
gh api call made: gh api --method DELETE /repos/malpanez/tempus/branches/develop/protection/required_status_checks
Applied at (UTC): 2026-08-05T09:XX (same session, before merging fix/sync-branches-race-condition)
Verification output:
  - GET .../branches/develop/protection/required_status_checks -> 404 "Required status checks not enabled" (confirms deletion)
  - GET .../branches/develop/protection -> allow_deletions=false, allow_force_pushes=false, enforce_admins=false (unchanged)
  - GET .../branches/main/protection/required_status_checks -> unchanged, still 6 contexts (Lint, Coverage Check, Functional Tests,
    Test (ubuntu-latest, 1.26), Build (ubuntu-latest, 1.26), Test (windows-latest, 1.26))
Rationale for option-b over option-a: develop is a single-writer passive mirror (nothing merges into it directly, by
design). Every commit reaching it already passed this exact CI suite on main moments earlier. Re-gating an
already-gated commit adds no real safety, only a coupling to ci.yml's ubuntu-slim matrix label that would silently
re-break the sync if that label ever changes. Force-push and deletion protection (the guards that actually matter
for a mirror branch) remain enabled.
-->

## Decisions Made

None beyond the plan — D-01 through D-06 were locked before planning and were followed exactly. The file was written to match the `<interfaces>` target contract byte-for-byte.

## Deviations from Plan

None - plan executed exactly as written (for the two tasks in scope). No auto-fixes were needed; the Task 1 and Task 2 gates both passed on the first run.

## Issues Encountered

None during execution. The pre-existing blocker documented in the plan's `<blocking_finding>` (develop's unsatisfiable required contexts) is unresolved by design — it is Task 3's scope.

## User Setup Required

**Yes — two manual steps remain, in this order:**

1. **Decide and apply Task 3** (repo owner): select option-a or option-b and run the corresponding `gh api` call. Without this the sync will still fail.
2. **Push and merge** (deliberately not done by this executor, per the session's no-push-without-confirmation policy): branch `fix/sync-branches-race-condition` is committed locally at `5dc715f` but has not been pushed. A `workflow_run`-triggered workflow always runs the copy of the file on the **default branch**, so this fix cannot be exercised from the feature branch — the PR will show no "Sync Branches" run at all. That is expected, not a failure.

## Next Phase Readiness

**Post-merge observation checklist** (cannot be run before this lands):

- After merge, the CI run for the merge commit finishes first, then a "Sync Branches" run appears and finishes green.
- `git ls-remote origin refs/heads/main refs/heads/develop` then reports the same SHA for both refs.
- If it fails again, read the new `remote:` lines: `is in progress` means the trigger change did not take effect; `is expected` means a required context is still misnamed (i.e. Task 3 was not applied or was applied incorrectly).

**Known follow-up (independent of which Task 3 option is chosen):** `ci.yml`'s conditional runner matrix means the set of check names a commit carries depends on the event that produced it. Any future change to that conditional, or to the `ubuntu-slim` label, changes the emitted context names and can silently invalidate branch protection contexts on either branch. There is no automated tripwire for this today; the Task 2 gate's `ci.yml` name assertion catches only a workflow *rename*, not a matrix-label change.

**Current remote state (unchanged by this plan):** `main` = `c8ad2782af19a323173f1262e4644d678f7561fb`, `develop` = `39108482edec09c75da8ca8531954c6e13f5b3ca`. `develop` is still a strict ancestor of `main` and still needs the fast-forward.

## Self-Check: PASSED

- `.github/workflows/sync-branches.yml` — FOUND (modified, 52 lines, contains `workflow_run`; satisfies the plan's `min_lines: 40` artifact contract)
- Commit `5dc715f` — FOUND in `git log`, subject `fix(ci): trigger sync-branches on CI completion and pin to the validated SHA`, 1 file changed, no `Co-Authored-By` trailer
- `git status --porcelain -- .github/` — confirms `sync-branches.yml` is the only `.github` change
- No branch protection API calls were made by this executor (Task 3 not executed)

---
*Quick task: 260805-du1*
*Completed (Tasks 1-2): 2026-08-05*
*Task 3: BLOCKED on repo owner decision*
