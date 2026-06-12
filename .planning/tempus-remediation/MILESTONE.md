# Tempus Remediation — GSD Milestone Plan

## Milestone

Fix the silent-error plumbing that contradicts the product promise (reliable reminders for
neurodivergent users), put a regression net under the final ICS output, and bring CI gates
from theater to enforcement. Five phases, each one PR. Phases 0→2 are strictly sequential;
phases 3 and 4 are independent of each other and can run in parallel after phase 1.

## Context

Full code review (verified by executing the binary) found a single root pattern behind all
high-priority bugs: **errors are discarded silently, and tests assert intermediate strings
instead of the final ICS**. Key findings, with locations:

- B1 — Wizard with alarm profile writes zero VALARMs: `resolveAlarmSpecs` returns
  `"profile:adhd-default"` (interactive.go:279), `ParseAlarmSpecs` doesn't understand
  `profile:` (only batch does, batch.go:623), error discarded at create.go:263-266.
- B2 — Default timezone ignored in `create`: `app.Config.Timezone` is only read by the
  wizard (interactive.go:223). `tempus -t Europe/Madrid create --start "2026-07-01 10:30"`
  and `TEMPUS_TIMEZONE=Europe/Madrid` both produce `DTSTART:20260701T103000Z` (UTC).
- B3 — Unvalidated TZIDs pass straight to ICS: `--start-tz madrid` emits
  `DTSTART;TZID=madrid:...`. `cityToIANA` exists but is not applied in create/quick/batch.
  VTIMEZONE only emitted for 5 hardcoded zones (calendar.go:593-684).
- B4 — Alarm/exdate errors swallowed CLI-wide: `--exdate` with a typo silently ignored
  (helpers.go:271-276); same pattern at batch.go:613-620.
- B5 — `--config/-c` flag is dead: defined at main.go:35, never read.
- B6 — i18n broken in template flow: `promptAlarmField` (template.go:731-784) has
  hardcoded Spanish; `alarm_prompt_*` keys exist in `locales/` but not in
  `internal/i18n/locales/` (110 vs 123 keys), code never uses them, covering test is
  permanently `t.Skip`.
- RFC: CATEGORIES unescaped `;` (calendar.go:321); ACTION:EMAIL missing required
  DESCRIPTION/ATTENDEE; `rrule` helper emits `UNTIL=YYYYMMDD` violating value-type rule
  with timed DTSTART; ATTENDEE + METHOD:PUBLISH; `tempus lint` requires SUMMARY/DTEND
  (optional in RFC), doesn't require DTSTAMP (mandatory), misses unclosed VEVENT,
  unfolding breaks spaces in folded lines.
- Tests: coverage_gaps_test.go (3,541 lines, 187 tests) is coverage filler — mostly
  `err == nil` assertions protecting the 79% gate without verifying behavior.
- CI: `.golangci.yml` has `default: none` (govet+staticcheck only, no errcheck);
  sonarcloud.yml installs golangci-lint v1 against v2 config with `|| true`;
  sync-branches.yml promotes renovate bumps to main without waiting for CI;
  `make release` broken (`$binary` unescaped); Docker always reports version `dev`
  (missing `ARG VERSION` + build-arg); security.yml gates are all `continue-on-error`.

Solid already (do not touch): TEXT escaping, UTF-8-safe folding at 75 octets, all-day
events (DTEND exclusive), EXDATE consistent with DTSTART, UIDs, embedded tzdata.

## Decisions (resolved 2026-06-12)

1. **42 commits ahead of main** → PR #3 to main, normal merge (no squash).
2. **`--config/-c`** → wire it (flags > env > config file > defaults; explicit `-c` to a
   missing file = fatal error).
3. **5 stale quick plans** → committed as historical archive.

---

## Phase 0 — Hygiene + regression net

**Goal:** clean working state, errcheck enabled and mapped, golden-file harness green
against CURRENT behavior (bugs included). No bug fixes in this phase.

### Tasks

- [x] 0.1 `git worktree prune`; delete all `worktree-agent-*` branches; merge PR #3;
      archive stale `.planning/quick/` plans.
- [ ] 0.2 `.golangci.yml` → `default: standard`. Run `golangci-lint run ./...` and commit
      the full errcheck violation list to `.planning/tempus-remediation/errcheck-map.md`.
      Do NOT fix violations yet.
- [ ] 0.3 Create golden harness in `internal/testutil`: builds the binary once per test
      run, runs full CLI flows (create, quick, batch, wizard via scripted stdin),
      compares `.ics` output against `testdata/golden/*.ics` with DTSTAMP/UID
      normalization. `-update` flag to regenerate.
- [ ] 0.4 Minimum 6 golden scenarios capturing CURRENT (buggy) output:
      create simple / create --start-tz / wizard with adhd-default profile (zero
      VALARMs — that's the point) / batch with exdates / rrule recurring / all-day.
- [ ] 0.5 CI runs the golden tests (same job as unit tests).

### Verification

- `git worktree list` shows only the main worktree; no `worktree-agent-*` branches remain.
- `golangci-lint run ./...` executes with `default: standard`; violation map committed.
- All 6 goldens pass against current behavior, locally and in CI.

---

## Phase 1 — Stop swallowing errors (B1, B4, B5)

**Goal:** invalid input or failed alarm/exdate processing is a fatal error with a clear
message and exit code != 0. Wizard alarm profiles actually produce VALARMs.

**Depends on:** Phase 0 complete (goldens green).

### Tasks

- [ ] 1.1 `addEventAlarms` / `addExDates` return `error`; propagate at create.go:263-266
      and batch.go:613-620; CLI exits non-zero with a message naming the offending value.
- [ ] 1.2 Fix `profile:` handling — one canonical expansion path shared by create/wizard/
      batch; remove the batch-only special case (batch.go:623).
- [ ] 1.3 `--exdate` with unparseable value → fatal error (helpers.go:271-276).
- [ ] 1.4 Wire `--config/-c` (main.go:35): precedence flags > env > config file >
      defaults. Missing file pointed to explicitly by flag = fatal error.
- [ ] 1.5 Clear errcheck map for all touched packages.
- [ ] 1.6 Update affected goldens to correct behavior (`-update`, review diff by hand).
      Negative tests: invalid `--exdate` → exit != 0; invalid alarm spec → exit != 0;
      `-c /nonexistent` → exit != 0.
- [ ] 1.7 Wizard Ctrl+C exits with code != 0.

### Verification

- `tempus create -i` with profile `adhd-default` → resulting `.ics` contains the
  profile's VALARMs (golden 3 diff shows exactly this change).
- `tempus create --exdate 2026-13-99 ...` → exit != 0, message names the bad value.
- `errcheck` clean on touched packages.

---

## Phase 2 — Timezone correctness (B2, B3)

**Goal:** configured timezone honored on every path; no TZID reaches the ICS without
IANA validation; invalid zone = fatal error.

**Depends on:** Phase 1.

### Tasks

- [x] 2.1 `app.Config.Timezone` (flag `-t`, env `TEMPUS_TIMEZONE`, config file) read by
      `create`, `quick`, and `batch` — not just the wizard. A resolved "UTC" keeps the
      Z form (correct RFC 5545 representation of UTC wall time).
- [x] 2.2 Single validation chokepoint `ResolveTimezone` (helpers.go): IANA accepted
      as-is → `cityToIANA` alias → error. Applied in create, quick, batch (per-row +
      --default-tz), wizard, and template flows. Bonus: quick's TZ now reinterprets
      wall clock (`time.Date` in zone) instead of shifting the instant with `In()`.
- [x] 2.3 `warnMissingVTZ` on stderr for valid zones without embedded VTIMEZONE
      (calendar.HasVTZDefinition exported). Invalid zone → hard error.
- [x] 2.4 Package clocks (`var timeNow = time.Now`) in calendar, nd, cli — no direct
      `time.Now()` left in production code.
- [x] 2.5 New goldens: create_default_tz_env (TEMPUS_TIMEZONE), create_alias_tz
      (madrid → Europe/Madrid); create_profile_alarm regenerated with TZID+VTIMEZONE.
      Negatives: --start-tz narnia and -t mordor → exit != 0.

### Verification

- Both repro commands produce Madrid-local DTSTART with TZID.
- `--start-tz madrid` → `Europe/Madrid`; `--start-tz narnia` fails loudly.
- No `time.Now()` left in `internal/` production code.

---

## Phase 3 — RFC 5545 compliance + i18n (B6 + medium findings)

**Depends on:** Phase 1. Parallelizable with phase 4.

### Tasks

- [x] 3.1 CATEGORIES values TEXT-escaped individually before joining.
- [x] 3.2 Alarm actions restricted to DISPLAY/AUDIO at parse time; EMAIL fails loud (cannot supply required ATTENDEE).
- [x] 3.3 rrule helper asks all-day vs timed; create/batch normalize date-only UNTIL on timed events to T235959Z.
- [x] 3.4 METHOD omitted when attendees present (documented: REQUEST would need an ORGANIZER we do not model).
- [x] 3.5 `tempus lint` rewritten: DTSTAMP required, SUMMARY/DTEND optional, unclosed
      VEVENT detected, RFC unfolding, UNTIL-type and CATEGORIES-escaping errors,
      TZID↔VTIMEZONE warning, positional args; goldens must pass own lint (E2E test).
- [x] 3.6 Locale trees synced (127 keys × 8 files, identical), promptAlarmField fully
      localized via translator + App writer, skip removed, per-language assertions.
- [x] 3.8 Wizard labels generated from config profile definitions (alarmProfileOptions);
      localized display names via alarm_profile_name_* keys; anti-drift test asserts
      every label contains exactly the offsets from config. No offsets hardcoded anywhere.
- [x] 3.7 TestLocaleKeyParity: any key/value drift between languages or trees fails the build.

### Verification

- `tempus lint` flags fixtures with each defect class and passes the goldens.
- Template flow in en/ga/pt shows translated prompts; parity test green.

---

## Phase 4 — CI/infra enforcement

**Depends on:** Phase 0. Parallelizable with phases 2-3.

### Tasks

- [ ] 4.1 sonarcloud.yml: golangci-lint v2 matching local config; remove `|| true`.
- [ ] 4.2 Remove sync-branches.yml + develop/main dance → trunk-based with branch
      protection + auto-merge after checks.
- [ ] 4.3 Sonar CPD exclusions back to tests-only; extract real duplicated helpers.
- [ ] 4.4 `make release`: `$$binary` fix; `ARG VERSION` + build-arg in Dockerfile/release.yml.
- [ ] 4.5 security.yml: remove `continue-on-error` from gates; fix or delete
      dependency-check.
- [ ] 4.6 Test diet: triage coverage_gaps_test.go; goldens + negative tests become the
      meaningful gate.

### Verification

- A PR with a lint violation, failing security check, or golden diff cannot merge.
- `docker run <image> --version` reports the tagged version.
- Renovate bump → PR → checks → auto-merge; no parallel promotion path.

---

## Phase 5 (deferred, optional) — Programmatic VTIMEZONE

Generate VTIMEZONE for any IANA zone from embedded tzdata. Acceptance: golden with
`Asia/Tokyo` includes a correct VTIMEZONE; Outlook-classic import verified manually once.

---

## Global success criteria

1. The review's two repro commands (wizard profile → VALARMs; `-t Europe/Madrid` →
   Madrid-local DTSTART) produce correct ICS.
2. Every invalid user input fails loudly with exit != 0.
3. Golden suite asserts final ICS on every PR; errcheck enforced repo-wide.
4. `tempus lint` catches every defect class the review found.
5. CI gates block merges; no `|| true` / `continue-on-error` on enforcement jobs.
