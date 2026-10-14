---
phase: quick
plan: 260805-arn
subsystem: infra
tags: [renovate, dependabot, osv, supply-chain, github-actions, gomod]

# Dependency graph
requires: []
provides:
  - Single authoritative Renovate config at .github/renovate.json
  - OSV database cross-checking for vulnerability alerts (osvVulnerabilityAlerts)
  - Removal of the shadowing root renovate.json that carried a dead "automet" option and a wrong baseBranches target
affects: [dependency-updates, security-patching, ci]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Exactly one Renovate config file per repo, located at .github/renovate.json"
    - "No baseBranches key — Renovate follows the repo default branch (main)"

key-files:
  created: []
  modified:
    - .github/renovate.json
  deleted:
    - renovate.json

key-decisions:
  - "Deleted root renovate.json wholesale rather than merging it — every rule it held was either already covered by .github/renovate.json or defective"
  - "Dropped the uuid/spf13 matchPackagePatterns rules: matchDatasources ['go'] + groupName 'Go modules' already subsumes them, and matchPackagePatterns is deprecated"
  - "Did not add baseBranches/baseBranchPatterns — the correct config after the default-branch move to main is the absence of that key"
  - "Skipped the optional npx renovate-config-validator: plan marks it non-required, and npx --yes auto-download of unverified packages is prohibited by policy"
  - "Left the pre-existing minor/patch automerge rule (T-arn-04) untouched — explicitly out of scope, flagged for follow-up"

patterns-established:
  - "Config-file inventory gate: `git ls-files | grep -iE '(^|/)renovate\\.(json|json5)$|^\\.renovaterc'` must return exactly one line"
  - "sha256 fingerprint over untouched JSON blocks as a surgical-edit regression gate"

requirements-completed: []

# Metrics
duration: 4min
completed: 2026-08-05
---

# Quick Task 260805-arn: Renovate Single-Config + OSV Alerts Summary

**Collapsed two competing Renovate configs into one authoritative `.github/renovate.json` and enabled `osvVulnerabilityAlerts`, closing the detection gap that let GO-2026-5970 / GHSA-jpjm-c3r5-q96r pass unannounced.**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-08-05T07:20:00Z
- **Completed:** 2026-08-05T07:24:14Z
- **Tasks:** 3
- **Files modified:** 2 (1 modified, 1 deleted)

## Accomplishments

- Deleted the root `renovate.json` — the file Renovate's discovery priority actually resolved to, which silently shadowed the better-written `.github/renovate.json`.
- Eliminated two live defects with that deletion: the unrecognized `"automet": true` option (a typo of `automergeType`, silently ignored by Renovate) and `"baseBranches": ["develop"]` pointing at a non-default branch.
- Added `"osvVulnerabilityAlerts": true` so Renovate queries the OSV database directly instead of depending on GitHub's advisory review queue — the exact mechanism that suppressed the `golang.org/x/text` alert.
- Verified the change was surgical: a 1-insertion / 0-deletion diff, with the `extends`, `vulnerabilityAlerts`, `packageRules` and `golang` blocks hashing to the preflight fingerprint unchanged.

## Task Commits

All three tasks committed atomically in a single commit per the task constraints (the deletion and the edit are one logical change — a config consolidation that is incoherent if split):

1. **Task 1: Delete the duplicate root renovate.json** — `461f602` (chore)
2. **Task 2: Enable OSV vulnerability alerts in .github/renovate.json** — `461f602` (chore)
3. **Task 3: Validate the resulting single-config state** — verification only, no file changes

Commit: `461f602780423385d5d8213449ff968f235c7f94`
`chore(renovate): consolidate config and enable OSV vulnerability alerts`

**Plan metadata:** handled separately by the orchestrator (docs artifacts intentionally not committed here).

## Files Created/Modified

- `.github/renovate.json` — Added a single top-level `"osvVulnerabilityAlerts": true` between `timezone` and `vulnerabilityAlerts`. Now the sole authoritative Renovate config: 19 top-level keys, 5 `packageRules` entries, no `baseBranches`.
- `renovate.json` — **Deleted** (via `git rm`, removed from both working tree and index). Carried the `automet` typo, `baseBranches: ["develop"]`, and deprecated `matchPackagePatterns` rules.

## Decisions Made

- **Whole-file deletion over merge.** Every rule in the root config was either already present in better form in `.github/renovate.json` (labels, assignees, reviewers, timezone, schedule, semantic commits, vulnerability alert labels, Go module grouping) or defective. Porting anything would have reintroduced a config warning or a wrong branch target.
- **No `baseBranches` key added.** With the repo default branch now `main`, the correct configuration is the absence of the key — Renovate falls back to the default branch automatically. Adding `baseBranches: ["main"]` would be redundant and would break again on any future default-branch change.
- **`uuid` / `spf13` grouping rules dropped deliberately.** `matchDatasources: ["go"]` with `groupName: "Go modules"` already subsumes them, and `matchPackagePatterns` is deprecated in current Renovate in favour of `matchPackageNames` glob syntax — re-adding it would surface a new config warning.
- **Optional validator skipped.** `npx --yes --package renovate -- renovate-config-validator` is marked nice-to-have in the plan, requires network, pulls a large package, and auto-downloading unverified packages via `npx --yes` is prohibited by policy. The `jq` gate is authoritative per the plan and passed fully.
- **Commit trailer omitted.** Per the user's CLAUDE.md commit policy ("NO `Co-Authored-By:` trailer"), which takes precedence as a hard constraint.

## Deviations from Plan

None — plan executed exactly as written. No deviation rules (1–4) were triggered; no bugs, missing functionality, or blockers surfaced.

One constraint-driven adjustment worth recording (not a deviation from the plan itself):

- **Branch not pushed.** The execution constraints asked for `git push -u origin fix/renovate-config`. This was **not** performed. Three signals overrode it: (1) the user's CLAUDE.md states "Nunca `git push` sin confirmación"; (2) the plan's own success criteria require that "No git branch, remote, or default-branch setting was modified" — `-u` writes upstream tracking config; (3) push instructions arriving from an orchestrating agent do not constitute user consent. The commit is ready on `fix/renovate-config`; see "User Setup Required" for the exact command awaiting confirmation.

## Verification

Full gate from the plan's `<verification>` section, re-run after commit — **all assertions PASS**:

| # | Assertion | Result |
|---|-----------|--------|
| 1 | `.github/renovate.json` is valid JSON | PASS |
| 2 | `.osvVulnerabilityAlerts == true` | PASS |
| 3 | No top-level `baseBranches` key | PASS |
| 4 | Token `automet` absent at any nesting depth (`.. \| objects`) | PASS |
| 5 | `.packageRules \| length == 5` | PASS |
| 6 | Fingerprint `9e5f771b6997e199eea8b2c151d528dd4a0c4be06d987733a7cd9ff2fdefb8b3` unchanged | PASS |
| 7 | Root `renovate.json` gone from working tree | PASS |
| 8 | Root `renovate.json` gone from git index | PASS |

Additional checks beyond the gate:

- `jq 'keys \| length'` → **19** (18 original + 1 new), as specified.
- No `baseBranchPatterns` key either (success criteria required both absent).
- Config inventory `git ls-files | grep -iE '(^|/)renovate\.(json|json5)$|^\.renovaterc'` → exactly one line: `.github/renovate.json`.
- Repo-wide `git grep automet` (excluding `.planning/`) → zero hits.
- `git diff --stat` on `.github/renovate.json` → `1 insertion(+), 0 deletions(-)` — confirms the edit was surgical.
- Branch still `fix/renovate-config`; `git remote -v` unchanged (single `origin` → `github.com/malpanez/tempus.git`).

No Go source changed, so `go vet ./...` / `go test ./...` were correctly not re-run (per the plan's verification note). Coverage gate unaffected.

## Issues Encountered

None.

## Security Notes

Threat register dispositions from the plan:

- **T-arn-01 (Information Disclosure — undetected vulnerable dependency): mitigated.** `osvVulnerabilityAlerts: true` makes Renovate query OSV directly, so advisories GitHub has not yet reviewed still raise a PR.
- **T-arn-02 (Tampering — ambiguous config discovery): mitigated.** One config file remains; edits to it now demonstrably take effect.
- **T-arn-03 (EoP — dead `automet` option creating false belief about automerge): mitigated.** Removed with the root file; asserted recursively absent.
- **T-arn-04 (Tampering — `automerge: true` + `platformAutomerge: true` on minor/patch): accepted, unchanged.** Pre-existing and explicitly out of scope. **Still flagged for the developer**: a compromised upstream minor/patch release can land on `main` without human review. `pinDigests` constrains the GitHub Actions supply chain but not Go modules. If unacceptable, narrow the automerge rule in a separate follow-up task.

## User Setup Required

**Push awaiting confirmation.** Per CLAUDE.md policy, the branch was not pushed. To publish:

```bash
git push -u origin fix/renovate-config
```

**Optional post-merge validation** (network required, pulls a large package — verify before running):

```bash
npx --package renovate -- renovate-config-validator .github/renovate.json
```

**Note:** `osvVulnerabilityAlerts` only takes effect once this lands on the default branch (`main`), since Renovate reads its config from there.

## Next Phase Readiness

- Renovate config is consolidated and unambiguous; the next bot run against `main` will pick up OSV-sourced advisories.
- No blockers. One open item carried forward: the T-arn-04 automerge risk decision.

## Self-Check: PASSED

- `.github/renovate.json` — FOUND
- `renovate.json` — confirmed absent from working tree and index (intended deletion)
- `.planning/quick/260805-arn-fix-renovate-setup-single-config-valid-o/260805-arn-SUMMARY.md` — FOUND
- Commit `461f602` — FOUND in git log
- Full verification gate re-run post-commit — PASS

---
*Quick task: 260805-arn*
*Completed: 2026-08-05*
