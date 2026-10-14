---
phase: quick
plan: 260805-arn
type: execute
wave: 1
depends_on: []
files_modified:
  - renovate.json
  - .github/renovate.json
autonomous: true
requirements: []
must_haves:
  truths:
    - "Renovate discovers exactly one config file in the repo, at .github/renovate.json"
    - "Root renovate.json no longer exists in the working tree nor in the git index"
    - "Renovate cross-checks the OSV database directly, so advisories GitHub has not reviewed (e.g. GHSA-jpjm-c3r5-q96r) still raise a PR"
    - "The active config contains no unrecognized Renovate options (no 'automet')"
    - "The active config has no baseBranches key, so Renovate scans the repo default branch (main)"
    - "Every pre-existing rule in .github/renovate.json is byte-identical after the change"
  artifacts:
    - path: ".github/renovate.json"
      provides: "Single authoritative Renovate config with OSV vulnerability alerts enabled"
      contains: "osvVulnerabilityAlerts"
  key_links:
    - from: ".github/renovate.json"
      to: "OSV database (osv.dev)"
      via: "top-level osvVulnerabilityAlerts flag"
      pattern: "\"osvVulnerabilityAlerts\"\\s*:\\s*true"
    - from: ".github/renovate.json"
      to: "repo default branch (main)"
      via: "absence of baseBranches key — Renovate falls back to the default branch"
      pattern: "^(?!.*baseBranches)"
---

<objective>
Collapse Tempus's two competing Renovate configs into one, and close the vulnerability-detection gap that let GO-2026-5970 slip through unannounced.

Purpose: Renovate's config discovery uses the FIRST match in its priority list, which is the root `renovate.json`. That file is the one actually driving the bot today, and it carries two defects: an unrecognized `"automet": true` option (a typo of `automergeType`, silently ignored by Renovate) and `"baseBranches": ["develop"]` pointing at the wrong branch. Meanwhile `.github/renovate.json` — the better-written config with grouping, digest pinning and gomodTidy — is silently dead. Deleting the root file promotes the good config to active. Adding `osvVulnerabilityAlerts` makes Renovate query the OSV database directly instead of relying on GitHub's advisory review queue, which is what suppressed the alert for GHSA-jpjm-c3r5-q96r.

Output: Exactly one Renovate config file at `.github/renovate.json`, with `osvVulnerabilityAlerts: true` added and every other key untouched.
</objective>

<execution_context>
@/home/malpanez/.claude/plugins/cache/gsd-plugin/gsd/4.0.4/workflows/execute-plan.md
@/home/malpanez/.claude/plugins/cache/gsd-plugin/gsd/4.0.4/templates/summary.md
</execution_context>

<context>
@.github/renovate.json
@renovate.json
</context>

<preflight_facts>
Verified before planning — do not re-derive:

- Both files are git-tracked: `git ls-files | grep -i renovate` returns `.github/renovate.json` and `renovate.json`.
- No workflow, script, doc, or Go source references the root `renovate.json` path. A repo-wide grep across `*.yml`, `*.yaml`, `*.md`, `*.sh`, `*.go` (excluding `.planning/`) returns zero hits. Deleting it breaks nothing.
- `jq` 1.7 and `python3` 3.12.3 are both on PATH; the verify commands below use `jq`.
- `.github/renovate.json` currently has 18 top-level keys and 5 `packageRules` entries. After this plan it must have 19 top-level keys and still exactly 5 `packageRules` entries.
- Fingerprint of the blocks that must NOT change (`jq -S '{packageRules, golang, extends, vulnerabilityAlerts}' | sha256sum`):
  `9e5f771b6997e199eea8b2c151d528dd4a0c4be06d987733a7cd9ff2fdefb8b3`
  This exact hash is re-asserted in Task 3 as the regression gate.

Out of scope — already handled this session, do NOT touch:
- GitHub default branch was changed from `develop` to `main`.
- `develop` was fast-forwarded to match `main`. It is a synced historical pointer, not stale. Do not delete it, do not repoint it, do not touch any git remote or branch setting.
</preflight_facts>

<tasks>

<task type="auto">
  <name>Task 1: Delete the duplicate root renovate.json</name>
  <files>renovate.json</files>
  <action>
Remove the root-level `renovate.json` from both the working tree and the git index using `git rm renovate.json`. Use `git rm`, not the Write tool and not `rm` — the file is tracked, and leaving it staged-but-present or present-but-untracked would keep Renovate's discovery pointing at it on the next default-branch scan.

Do NOT port any content from this file into `.github/renovate.json`. Everything it contains is either already present in better form in `.github/renovate.json` (labels, assignees, reviewers, timezone, schedule, semantic commits, vulnerability alert labels, Go module grouping) or is defective and must not survive: the `"automet": true` typo and the `"baseBranches": ["develop"]` key. Specifically, do NOT add a `baseBranches` or `baseBranchPatterns` key anywhere — with the repo default branch now `main`, the correct configuration is the absence of that key, which makes Renovate scan the default branch.

The root file's `matchPackagePatterns` rules for `uuid` and `spf13` are deliberately dropped: `.github/renovate.json` already groups all Go modules via `matchDatasources: ["go"]` with `groupName: "Go modules"`, which subsumes them. `matchPackagePatterns` is also deprecated in current Renovate in favour of `matchPackageNames` with glob syntax, so re-adding it would introduce a new config warning.
  </action>
  <verify>
    <automated>test ! -e renovate.json && test -z "$(git ls-files renovate.json)" && echo "PASS: root renovate.json removed from tree and index"</automated>
  </verify>
  <done>`renovate.json` does not exist at the repo root and `git ls-files renovate.json` returns empty. `.github/renovate.json` still exists and is unmodified by this task.</done>
</task>

<task type="auto">
  <name>Task 2: Enable OSV vulnerability alerts in .github/renovate.json</name>
  <files>.github/renovate.json</files>
  <action>
Add a single top-level key to `.github/renovate.json`:

`osvVulnerabilityAlerts` set to boolean `true`.

Placement: insert it immediately after the `timezone` line and immediately before the `vulnerabilityAlerts` object, so the two vulnerability-related settings sit together. Use the file's existing 2-space indentation and keep the trailing newline at end of file.

This is a top-level Renovate option (boolean, defaults to `false`). When enabled, Renovate downloads the OSV database and queries it locally for every dependency, independent of GitHub's Security Advisory Database review status. That is the whole point of this change: GHSA-jpjm-c3r5-q96r (GO-2026-5970 / CVE-2026-56852 in `golang.org/x/text`) is currently marked "unreviewed" by GitHub, so GitHub-native Dependabot alerts never fired for it. OSV carries the advisory regardless of GitHub's triage state.

The existing `vulnerabilityAlerts` object stays exactly as it is — it configures how vulnerability PRs are labelled and assigned, and it applies to OSV-sourced alerts too. Do not merge the two, and do not add `automerge` to it.

Change nothing else. Do not reorder keys, do not reformat, do not touch `extends`, `packageRules`, `golang`, `prConcurrentLimit`, `prCreation`, `prHourlyLimit`, `semanticCommits`, `commitMessagePrefix`, `commitMessageAction`, `rebaseWhen`, or `platformAutomerge`. Task 3 asserts a sha256 fingerprint over the untouched blocks, so any incidental edit will fail the gate.
  </action>
  <verify>
    <automated>jq -e '.osvVulnerabilityAlerts == true' .github/renovate.json >/dev/null && jq -e 'keys | length == 19' .github/renovate.json >/dev/null && echo "PASS: osvVulnerabilityAlerts enabled, 19 top-level keys"</automated>
  </verify>
  <done>`.github/renovate.json` parses as valid JSON, contains `"osvVulnerabilityAlerts": true` at the top level, and has exactly 19 top-level keys (18 original + 1 new).</done>
</task>

<task type="auto">
  <name>Task 3: Validate the resulting single-config state</name>
  <files>.github/renovate.json</files>
  <action>
Run the full validation gate below as one command chain. Every clause must pass. If any clause fails, fix the underlying file rather than relaxing the check.

The chain asserts six things: (1) `.github/renovate.json` is syntactically valid JSON; (2) OSV alerts are on; (3) no `baseBranches` key exists at the top level, so Renovate targets the default branch `main`; (4) the string `automet` appears nowhere in any object in the file, at any nesting depth; (5) the pre-existing `extends`, `vulnerabilityAlerts`, `packageRules` and `golang` blocks still hash to the fingerprint recorded in preflight, proving Task 2 was surgical; (6) the root `renovate.json` is gone from both the working tree and the git index.

Note on the `automet` check: it walks every object recursively via `..`, so it catches the typo inside a nested `packageRules` entry, which is where it lived in the deleted root file. A flat `grep` on the top level would not.

Optionally, if network is available, `npx --yes --package renovate -- renovate-config-validator .github/renovate.json` gives Renovate's own schema verdict. This is a nice-to-have only — the tool is not installed, requires network, and pulls a large package. Do not treat its absence or failure-to-fetch as a plan failure, and do not add it to any CI workflow as part of this plan. The jq gate below is the authoritative check.
  </action>
  <verify>
    <automated>jq -e . .github/renovate.json >/dev/null && jq -e '.osvVulnerabilityAlerts == true' .github/renovate.json >/dev/null && jq -e 'has("baseBranches") | not' .github/renovate.json >/dev/null && jq -e '[.. | objects | has("automet")] | any | not' .github/renovate.json >/dev/null && jq -e '.packageRules | length == 5' .github/renovate.json >/dev/null && [ "$(jq -S '{packageRules, golang, extends, vulnerabilityAlerts}' .github/renovate.json | sha256sum | cut -d' ' -f1)" = "9e5f771b6997e199eea8b2c151d528dd4a0c4be06d987733a7cd9ff2fdefb8b3" ] && test ! -e renovate.json && test -z "$(git ls-files renovate.json)" && echo "PASS: single valid config at .github/renovate.json, OSV enabled, no automet, no baseBranches, existing rules unchanged, root config deleted"</automated>
  </verify>
  <done>All six assertions pass and the chain prints the PASS line. Exactly one Renovate config file exists in the repo.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| upstream Go module registry → repo `go.mod` | Third-party code enters the build via Renovate-authored PRs |
| advisory feeds (GitHub GHSA / OSV) → Renovate | Vulnerability signal Tempus depends on to know it must patch |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-arn-01 | Information Disclosure | vulnerability detection pipeline | mitigate | Known-vulnerable dependency stays undetected because the advisory is "unreviewed" in GitHub's database. Task 2 sets `osvVulnerabilityAlerts: true` so Renovate queries OSV directly, bypassing GitHub's review gate. This is the exact failure mode observed with GHSA-jpjm-c3r5-q96r. |
| T-arn-02 | Tampering | Renovate config discovery | mitigate | Two configs in the repo means the file a maintainer edits may not be the file Renovate reads, so a hardening change can appear applied while having zero effect. Task 1 deletes the shadowing root file, making the single remaining config unambiguously authoritative. |
| T-arn-03 | Elevation of Privilege | dead `"automet": true` option | mitigate | An unrecognized option is silently ignored by Renovate, so a maintainer reading the config forms a false belief about automerge behaviour. Deleted with the root file in Task 1; Task 3 asserts recursively that the token appears nowhere. |
| T-arn-04 | Tampering | `automerge: true` + `platformAutomerge: true` for minor/patch in `.github/renovate.json` | accept | A compromised upstream minor/patch release could land on `main` without human review. Accepted for now: this is pre-existing behaviour explicitly held out of scope for this plan, Tempus is a single-maintainer offline CLI with a small dependency surface, and `pinDigests` already constrains the GitHub Actions supply chain. **Flagged for the developer**: if this risk is not acceptable, narrow the automerge rule in a follow-up — do not fold it into this plan. |
| T-arn-SC | Tampering | npm/pip/cargo installs | n/a | No package-manager install tasks in this plan. The optional `renovate-config-validator` invocation in Task 3 is explicitly non-required and must not be added to CI here. |
</threat_model>

<verification>
Single gate, identical to Task 3's automated check:

```bash
jq -e . .github/renovate.json >/dev/null \
  && jq -e '.osvVulnerabilityAlerts == true' .github/renovate.json >/dev/null \
  && jq -e 'has("baseBranches") | not' .github/renovate.json >/dev/null \
  && jq -e '[.. | objects | has("automet")] | any | not' .github/renovate.json >/dev/null \
  && jq -e '.packageRules | length == 5' .github/renovate.json >/dev/null \
  && [ "$(jq -S '{packageRules, golang, extends, vulnerabilityAlerts}' .github/renovate.json | sha256sum | cut -d' ' -f1)" = "9e5f771b6997e199eea8b2c151d528dd4a0c4be06d987733a7cd9ff2fdefb8b3" ] \
  && test ! -e renovate.json \
  && test -z "$(git ls-files renovate.json)" \
  && echo PASS
```

Config-file inventory — must return exactly one line, `.github/renovate.json`:

```bash
git ls-files | grep -iE '(^|/)renovate\.(json|json5)$|^\.renovaterc'
```

No Go code changes in this plan, so the Tempus test gates (`go vet ./...`, `go test ./...`) are unaffected and need not be re-run.
</verification>

<success_criteria>
- Root `renovate.json` is deleted from the working tree and staged as a deletion in git
- `git ls-files | grep -i renovate` returns exactly one path: `.github/renovate.json`
- `.github/renovate.json` is valid JSON with 19 top-level keys
- `.github/renovate.json` contains `"osvVulnerabilityAlerts": true`
- `.github/renovate.json` has no `baseBranches` and no `baseBranchPatterns` key
- The token `automet` appears nowhere in the repo's Renovate config
- The `extends`, `vulnerabilityAlerts`, `packageRules` and `golang` blocks hash to `9e5f771b6997e199eea8b2c151d528dd4a0c4be06d987733a7cd9ff2fdefb8b3`
- No git branch, remote, or default-branch setting was modified
- Commit follows conventional format, e.g. `chore(renovate): consolidate config and enable OSV vulnerability alerts`
</success_criteria>

<output>
After completion, create `.planning/quick/260805-arn-fix-renovate-setup-single-config-valid-o/260805-arn-SUMMARY.md`
</output>
