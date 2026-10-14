---
quick_id: 261013-tdw
status: complete
date: 2026-10-14
branch: chore/security-cleanup
commits: [cad5305, a8e0faa, 93bbf26, 7422b10, 1269b82]
---

# 261013-tdw: Security cleanup

Executed directly by the orchestrator at the user's request. A planner subagent was started and then cancelled by the user; its PLAN.md described a superseded approach (toolchain bump in-repo, broad indirect refresh) and was discarded, not executed.

## Context

- Scheduled Security Scan on main failed 2026-10-12: govulncheck found GO-2026-6617/6613/6612/6611/6603 in go1.26.8 stdlib net/http.
- Renovate had never opened a PR: the Mend account default was **Silent** mode. Confirmed in the Mend Developer Portal; the user approved Renovate PRs from there.
- Renovate PR #19 (toolchain go1.26.8 -> go1.26.9) and #20 (x/vuln v1.7.0 -> v1.9.0) merged to main. The Security Scan on main is green again.

## Commits

| Commit | Change |
|--------|--------|
| cad5305 | `golang.org/x/text` v0.40.0 -> v0.41.0 (GO-2026-6629; Renovate skips indirect modules) |
| a8e0faa | `scripts/install.bash` rewritten: the old script asked for a nonexistent asset and installed the 404 page as the binary. It now downloads the release tarball and its `.sha256`, verifies the checksum, extracts the binary and installs it with mode 0755. curl runs with `--proto =https --tlsv1.2 -f` (Sonar shell:S6506). |
| 93bbf26 | Golden harness resolves `go` via `exec.LookPath` (Sonar go:S4036, new-code issue behind Security Rating B) |
| 7422b10 | Renovate: legacy top-level `golang` block migrated to a `packageRules` entry |
| 1269b82 | `.gitignore`: `.venv/`, `.vscode/settings.json` |

## Verification (local, Go 1.26.9)

- `gofmt -l .` empty; `go vet ./...` clean; `go test ./...` green; total coverage 88.6% (gate 79%)
- `go tool govulncheck ./...`: "No vulnerabilities found", exit 0
- `bash -n` + `shellcheck` clean. Installer exercised against the real v0.7.0 release (latest, v0.7.0, 0.7.0 -> `tempus v0.7.0`); a missing tag exits 22 and installs nothing; a malformed tag is rejected.
- `renovate-config-validator --strict` exit 0
- `tempus create` + `tempus lint` on the generated ICS: passed

## Not verified locally

- Whether SonarCloud clears go:S4036 and shell:S6506. Visible only in the PR analysis.
- golangci-lint (not installed locally; runs in CI `Lint`).

## Out of scope / follow-ups

- SonarCloud project: main branch still labelled `develop`; new-code period anchored at 2025-12-15 (`previous_version`); duplication 4.3% > 3%.
- `actions/dependency-review-action` pin `46a3c49… # v4`: Mend warns it cannot determine a new digest (no floating `v4` tag upstream). The job still passes; the major Actions group in Renovate will move it to v5.
- Renovate items left for the user in the Mend portal: Actions (non-major), Docker digests, Actions (major), huh v2 (do not auto-apply).
- Remote branch `fix/cyber-neo-security-audit` (merged in PR #18) still exists on origin.
