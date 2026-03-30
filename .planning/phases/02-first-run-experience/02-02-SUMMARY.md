---
plan: 02-02
phase: 02-first-run-experience
status: complete
completed: 2026-03-30
coverage_after: 79.5%
---

# Plan 02-02 Summary — tempus init wizard

## What was built

`tempus init` interactive wizard (UX-01) — 4-step configuration flow using `survey/v2`:
1. Timezone — auto-detected from `/etc/localtime`, validated via `config.ValidateTimezone`
2. Language — auto-detected from `LANG` env var, select from en/es/pt/ga
3. Output directory — defaults to `.`, validated via `config.ValidateOutputDir`
4. Default alarm profile — select from adhd-default/adhd-countdown/medication/single/none

## Key decisions implemented

- D-01: Auto-detect + confirm (system timezone + LANG detection pre-filled)
- D-02: Existing config → overwrite prompt `[y/N]` (default No)
- D-03: Summary table + `Next: tempus create --start today --duration 1h` hint
- D-04: No `--yes` flag — interactive only

## Deviations from plan

- **Overwrite error handling**: Any error from survey on overwrite prompt (including EOF in non-interactive mode) defaults to "no overwrite". Plan spec said only handle `terminal.InterruptErr`, but EOF from pipe stdin also needs graceful handling. This is safer behavior and enables the unit test.

## Tests added

- `TestInitCmdRegistered` — verifies `init` appears in root command tree
- `TestInitCmdHelp` — verifies `cmd.Use == "init"` and Short description
- `TestInitCmdExistingConfigNoOverwrite` — verifies config file unchanged when overwrite declined (or stdin EOF)

## Files modified

- `main.go` — added `newInitCmd()`, `runInit()`, registered in `newRootCmd()`; added `terminal` import
- `main_test.go` — 3 new unit tests

## Human verification

Passed manually:
- Auto-detection works (Europe/Dublin pre-filled)
- Language select works (en/es/pt/ga)
- Summary table shows correct values
- Overwrite prompt defaults to No
- Invalid timezone shows error and re-prompts
- Ctrl+C exits cleanly mid-wizard
