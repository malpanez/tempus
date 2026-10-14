---
phase: quick-260805-di2
plan: 01
subsystem: ci
tags: [github-actions, branch-sync, ci]
requires: []
provides:
  - "One-way main -> develop fast-forward sync on every push to main"
affects:
  - ".github/workflows/"
tech-stack:
  added: []
  patterns:
    - "SHA-pinned actions/checkout (repo-wide pin 34e1148)"
    - "Job-level least-privilege permissions block"
    - "concurrency group with cancel-in-progress for ref-mutating jobs"
key-files:
  created:
    - .github/workflows/sync-branches.yml
  modified: []
decisions:
  - "One-way only: the auto-promote develop->main job from the deleted 06cca03 workflow is deliberately NOT ported (it caused the historical --ff-only deadlock)"
  - "Fast-forward failure exits non-zero with a ::error:: annotation; never --force, never a real merge or rebase"
metrics:
  duration: 2min
  tasks: 2
  files: 1
  completed: 2026-08-05
---

# Quick Task 260805-di2: Recreate one-way sync-branches.yml Summary

Added `.github/workflows/sync-branches.yml`, a single-job workflow that fast-forwards `develop` to `main` after every push to `main`, resurrecting only the safe half of the bidirectional workflow deleted in `06cca03`.

## What Was Built

One new file, `.github/workflows/sync-branches.yml` (38 lines):

- **Trigger:** `on: push: branches: [ main ]` and nothing else. `develop` appears nowhere in the `on:` block, so the workflow cannot self-trigger and no `pull_request`/`workflow_dispatch`/`schedule` entry points exist (D-02).
- **Concurrency:** `group: sync-branches`, `cancel-in-progress: true`. Batched Renovate merges landing back to back would otherwise race, and the older run's push would be rejected as non-fast-forward.
- **Job:** single `sync` job, `runs-on: ubuntu-slim` (D-05), `timeout-minutes: 5`, job-level `permissions: contents: write` only.
- **Steps:** checkout pinned to the repo-wide SHA `actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5 # v4` with `fetch-depth: 0` and `token: ${{ secrets.GITHUB_TOKEN }}`, then one run step: bot identity -> `git checkout develop` -> `git merge --ff-only origin/main` with a brace-block failure handler emitting `::error::` and `exit 1` (D-04) -> `git push origin develop`.

No existing file was touched. `ci.yml`, `security.yml`, and `release.yml` are byte-identical to their pre-plan state (confirmed via `git status --porcelain` on each path returning empty).

## Task Completion

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create the one-way sync-branches workflow | a63c1a6 | .github/workflows/sync-branches.yml |
| 2 | Gate the workflow against every stated constraint | a63c1a6 (verification-only, no file change needed) | — |

Task 2 required no edits — the Task 1 file passed all 15 assertions on the first run, so both tasks landed in one commit as specified by the plan's `<output>` block.

## Verification Results

All automated checks from `<verification>` items 1-8 passed:

1. **YAML parses** — `yaml.safe_load` succeeded.
2. **Trigger** — `on:` has exactly one key `push` with `{"branches": ["main"]}`; `develop` absent from the block.
3. **Job shape** — exactly 1 job, `runs-on: ubuntu-slim`, job-level `permissions == {"contents": "write"}`, no workflow-level `permissions` key.
4. **Forbidden patterns** — none of `checkout main`, `push origin main`, `--force`, `push -f`, `rebase`, `--no-ff`, `-X ours`, `-X theirs` present in the run body (comment-only lines stripped before matching).
5. **Required commands** — `git checkout develop`, `git merge --ff-only origin/main`, `exit 1`, `git push origin develop` all present.
6. **Action pin** — `uses` list is exactly `["actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5"]` with `fetch-depth: 0`.
7. **Fast-forward precondition** — `git fetch origin main develop` succeeded (online, live refs not cached approximations); `git merge-base --is-ancestor origin/develop origin/main` returned 0. `origin/develop` is at `3910848`, `origin/main` at `020d5fb`, develop 6 commits behind. The first real run will fast-forward cleanly.
8. **Working tree** — the only tracked change outside `.planning/` is the new `.github/workflows/sync-branches.yml`.

Gate output: `OK: sync-branches.yml satisfies all 15 constraints`.

## Deviations from Plan

None - plan executed exactly as written. The target file contract in `<interfaces>` was reproduced verbatim.

## Threat Model Coverage

All `mitigate` dispositions from the plan's threat register are implemented and asserted by the gate:

| Threat ID | Status |
|-----------|--------|
| T-QUICK-01 (EoP, permissions) | Mitigated — job-level `contents: write` only, asserted |
| T-QUICK-02 (Tampering, git commands) | Mitigated — forbidden-pattern scan passes; only `--ff-only` present |
| T-QUICK-03 (Supply chain, checkout) | Mitigated — immutable SHA pin asserted, no floating tag |
| T-QUICK-04 (DoS, trigger graph) | Mitigated — main-only trigger + `concurrency` with `cancel-in-progress` |
| T-QUICK-05 (Repudiation) | Accepted per plan — ff-only creates no commit objects |
| T-QUICK-SC (Package legitimacy) | N/A — no packages installed |

No new security-relevant surface was introduced beyond what the plan modeled.

## Known Stubs

None.

## Open Follow-Ups

Post-merge observation, deliberately not blocking (per `<verification>`):

- After this branch merges to `main`, confirm the "Sync Branches" run appears in the Actions tab and finishes green, then verify `git ls-remote origin refs/heads/develop refs/heads/main` reports the same SHA for both refs.
- The branch `feat/sync-develop-from-main` is committed but **not pushed** — pushing is left to the orchestrator per this session's no-push-without-confirmation policy.

## Self-Check: PASSED

- `.github/workflows/sync-branches.yml` — FOUND
- Commit `a63c1a6` — FOUND in `git log`, 1 file changed, 38 insertions, 0 deletions, no `Co-Authored-By` trailer
