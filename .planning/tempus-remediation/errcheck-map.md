# Errcheck Map — Phase 0 baseline (2026-06-12)

`golangci-lint run --max-issues-per-linter=0 --max-same-issues=0` with `default: standard`
found **81 issues** (77 errcheck + 4 unused). Note: the default `max-same-issues: 3` cap
hid 65 of them on the first run — always lift the caps when auditing.

## Disposition in phase 0 (no behavior changes)

| Category | Count | Disposition |
|----------|-------|-------------|
| `fmt.Fprint*` to console writers | 71 | `errcheck.exclude-functions` in .golangci.yml — best-effort UI output (template.go 28, timezone.go 13, batch.go 11, rrule.go 12, locale.go 3, version.go 2, helpers.go 2, interactive.go 1, config.go 1... see full list below) |
| `defer f.Close()` on read-only files | 2 | inline `//nolint:errcheck` (batch.go:336, template.go:383) |
| Writability-probe cleanup | 2 | inline `//nolint:errcheck` (config/config.go:284-285) |
| Test pipe helpers | 2 | inline `//nolint:errcheck` (create_test.go:70,72) |
| Dead code (`unused`) | 4 | deleted — `icsDurationRegex` (batch.go), `newBatchCmdForTest` (batch_test.go), `tzInfo` (coverage_gaps_test.go), `eventWithAlarms` (templates_test.go). Compiler-verified, zero behavior change. |

## THE REAL BUG MAP — what errcheck CANNOT see

The high-priority review bugs are **semantic** swallows: the error IS checked, then
discarded. errcheck is structurally blind to these. They are phase 1's worklist:

| Site | Pattern | Consequence |
|------|---------|-------------|
| internal/cli/create.go:263-266 (`addEventAlarms`) | `if err == nil { ... }` — parse error path does nothing | B1: wizard profile alarms silently dropped, zero VALARMs written |
| internal/cli/helpers.go:271-276 (`addExDates`) | parse error silently skipped per value | B4: typoed `--exdate` ignored, recurring event fires on excluded day |
| internal/cli/batch.go:613-620 (`addBatchAlarms`) | discards profile-expansion error AND ParseAlarmSpecs error | B4: batch alarms silently dropped; the helpful "profile not found. Available: ..." message is dead code |
| internal/cli/quick.go:92-96 (`applyTimezoneToDetails`) | invalid TZ skips conversion but still stamps TZID | invalid TZID in output ICS |
| internal/cli/app.go:22-31 (`SetupPersistentPreRunE`) | `config.Load()` error → silent defaults (en/UTC) | corrupt config silently ignored |
| internal/cli/interactive.go:231-233, init.go:41,59,76,89,106 | `if err := form.Run(); err != nil { return nil }` | Ctrl+C and real terminal errors both exit 0 |
| main.go:35 `--config/-c` | flag value never read | dead flag (decision: wire it, phase 1.4) |

Phase 1 exit criteria: every row above either returns the error (fatal, exit != 0) or has
an explicit, justified reason not to.

## Full raw list (81)

```
internal/cli/batch.go:236:15: fmt.Fprintf (errcheck)
internal/cli/batch.go:244:14: fmt.Fprintf (errcheck)
internal/cli/batch.go:246:16: fmt.Fprintln (errcheck)
internal/cli/batch.go:255:13: fmt.Fprintf (errcheck)
internal/cli/batch.go:264:14: fmt.Fprintf (errcheck)
internal/cli/batch.go:266:13: fmt.Fprintf (errcheck)
internal/cli/batch.go:267:13: fmt.Fprintf (errcheck)
internal/cli/batch.go:273:14: fmt.Fprintf (errcheck)
internal/cli/batch.go:276:16: fmt.Fprintln (errcheck)
internal/cli/batch.go:278:14: fmt.Fprintf (errcheck)
internal/cli/batch.go:336:15: f.Close (errcheck) → nolint
internal/cli/batch.go:61:5: icsDurationRegex (unused) → deleted
internal/cli/batch_test.go:1048:6: newBatchCmdForTest (unused) → deleted
internal/cli/coverage_gaps_test.go:984:7: tzInfo (unused) → deleted
internal/cli/create_test.go:70:11: w.Close (errcheck) → nolint
internal/cli/create_test.go:72:11: io.Copy (errcheck) → nolint
internal/cli/config.go + helpers.go + interactive.go + locale.go + rrule.go +
template.go + timezone.go + version.go: 61 more fmt.Fprint* console-output sites
(run `golangci-lint run --max-same-issues=0` without the exclude-functions block
to regenerate the exact list)
internal/config/config.go:284:9: f.Close (errcheck) → nolint
internal/config/config.go:285:11: os.Remove (errcheck) → nolint
internal/templates/templates_test.go:1272:9: eventWithAlarms (unused) → deleted
```
