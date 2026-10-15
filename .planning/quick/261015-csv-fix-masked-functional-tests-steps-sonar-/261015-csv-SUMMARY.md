---
quick_id: 261015-csv
status: complete
date: 2026-10-15
branch: fix/ci-functional-tests-sonar-config
commits: [f1bfbe0, 00569dd, 8938e7b]
---

# 261015-csv: Masked CI steps, Sonar config, docs and planning drift

Executed directly by the orchestrator at the user's request, after a repo review on 2026-10-15.

## Findings and fixes

| Commit | Finding | Fix |
|--------|---------|-----|
| f1bfbe0 | The required `Functional Tests` job had three steps ending in `\|\| true`. `./tempus --version` always failed (no such flag) and was masked. The `--check-conflicts` step asserted nothing. | The step runs `tempus version`. The conflict step now requires the warning and 2 VEVENTs and runs under `shell: bash` (pipefail). The `rrule --help` fallbacks were dropped. |
| 00569dd | SonarCloud Automatic Analysis ignores `sonar-project.properties`. All 36 `*_test.go` files were indexed as production code, and 14 of the 15 files with the most duplicated lines are tests. This caused the 4.3% new-code duplication gate failure. | Added `.sonarcloud.properties` (`sonar.tests=.` + `sonar.test.inclusions=**/*_test.go`, tests and testdata excluded from sources). Marked `sonar-project.properties` as CI-only. |
| 8938e7b | The README documented a multi-line `tempus version` output that never existed. | Documented the real outputs (release: `tempus vX.Y.Z`; `make build`: plus commit and date). |
| (docs commit) | `STATE.md` claimed 5/5 phases complete but also "Phase 03 EXECUTING"; `ROADMAP.md` showed phases 2/4/5 open; `MILESTONE.md` had unchecked tasks shipped in PRs #4/#5/#7. | Reconciled against code and the phase VERIFICATION.md files. Superseded tasks (4.1, 4.2) are annotated; 4.3 (partial) and 4.6 (open) stay unchecked. |

## Verification

- Each edited Functional Tests step extracted from ci.yml and run from the repo root with `bash --noprofile --norc -eo pipefail` and an empty HOME: all exit 0. Running outside the repo fails `template describe medical` (templates are read from the checkout), as in CI.
- `ci.yml` parses as YAML; actionlint not installed locally.
- Planning ticks backed by evidence: 01/02/04/05 VERIFICATION.md `status: passed`; 02-02/02-03 SUMMARY present; `addEventAlarms`/`addExDates` return `error`, and create/wizard/batch share `addEventAlarms` → `expandAlarmProfiles`; negative tests for exdate, alarm spec and missing `-c` file; PTY Ctrl+C test; no `continue-on-error`/`dependency-check` left in workflows.

## Not verified

- Whether SonarCloud honors the `**/*_test.go` patterns. The docs say "Wildcard patterns are not allowed" without scoping it. Check after merge: `api/components/tree?...&q=_test.go` should report qualifier `UTS`, not `FIL`.

## Left for the user

- Release v0.7.1 (v0.7.0 binaries were built with go1.26.8 and x/text 0.40).
- Switch tempus to Interactive in the Mend portal.
- SonarCloud UI: rename the main branch from `develop` and set the new-code definition to a number of days. Automatic Analysis sends no project version, so "previous version" never advances.
- Enable "Automatically delete head branches" in the repo settings.
