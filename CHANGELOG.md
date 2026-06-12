# Changelog

All notable changes to Tempus are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project
adheres to [Semantic Versioning](https://semver.org/) (0.x: minor bumps may
include breaking changes, listed explicitly below).

## [v0.6.0] - 2026-06-12

### Breaking Changes

- **Invalid input now fails loudly.** Unparseable `--alarm` specs, unknown
  `profile:` names, invalid `--exdate` values, unresolvable timezones, and
  an unreadable `--config` file all exit non-zero with a message naming the
  offending value. Previous versions silently dropped these and reported
  success — scripts relying on that behavior will now see failures.
- **Wizard exit codes.** Ctrl+C or a form error in `create -i`, `init`, and
  the `quick` confirmation exits non-zero (previously exit 0).
- **`create -i` rejects incompatible flags.** Only `-o/--output` may be
  combined with `--interactive`; any other flag is an error instead of
  being silently ignored.
- **Timezone semantics corrected.** `-t/--timezone`, `TEMPUS_TIMEZONE`, and
  the config-file timezone are now applied by `create`, `quick`, and
  `batch`. Events that previous versions stamped as UTC (`DTSTART:...Z`)
  while a timezone was configured now carry `DTSTART;TZID=<zone>` with an
  embedded VTIMEZONE. A configured `UTC` keeps the `Z` form.
- **EMAIL alarm action rejected.** RFC 5545 requires DESCRIPTION, SUMMARY,
  and ATTENDEE on EMAIL alarms, which tempus cannot supply; `action=email`
  specs now error instead of emitting invalid VALARMs. DISPLAY and AUDIO
  remain supported.

### Added

- End-to-end golden regression suite: the compiled binary runs full CLI
  flows and the final ICS is diffed against checked-in goldens; every
  generated golden must pass `tempus lint`. E2E coverage is instrumented
  in CI and shipped to Codecov/Sonar alongside unit coverage.
- Programmatic VTIMEZONE generation from the embedded tzdata for any IANA
  zone (STANDARD/DAYLIGHT blocks with derived RRULEs; no-DST zones get a
  single STANDARD block). Replaces the previous 5 hardcoded zones.
- City-alias timezone resolution everywhere (`--start-tz madrid` →
  `Europe/Madrid`) through a single validation chokepoint.
- Locale parity: 127 identical keys across all 4 languages in both locale
  trees, enforced by a build-failing parity test. Wizard alarm-profile
  labels are generated from the config profile definitions (anti-drift
  test included) with localized display names.
- `tempus lint` hardened to the RFC: requires DTSTAMP, detects unclosed
  VEVENTs, validates UNTIL/DTSTART value-type agreement and CATEGORIES
  escaping, warns on TZID references without VTIMEZONE, accepts positional
  file arguments.
- `batch template` accepts `--format csv|yaml|json` and the content always
  matches; all 10 templates are generated from a single data source and
  every generated file is guaranteed readable by `tempus batch`
  (roundtrip-tested for all 30 type×format combinations).
- `--config/-c` is wired: explicit path that cannot be read is fatal;
  default-path load failures warn on stderr instead of silently using
  defaults.

### Fixed

- Wizard alarm profiles now write VALARMs — selecting `adhd-default`
  previously produced an event with no reminders at all (B1).
- Default timezone honored outside the wizard (B2) and no raw TZID string
  reaches the ICS without IANA validation (B3).
- Alarm and exdate parse errors are no longer swallowed in create, batch,
  or the wizard (B4); `--config` is no longer a dead flag (B5); the
  template flow no longer shows hardcoded Spanish prompts to en/ga/pt
  users (B6).
- CATEGORIES values are TEXT-escaped; METHOD is omitted when attendees are
  present (RFC 5545 §3.8.4.1); date-only UNTIL on timed events is
  normalized to end-of-day UTC.
- `quick -t` reinterprets the parsed wall clock in the target zone instead
  of shifting the instant.
- QUICK_START documentation matched to the real CLI surface (`--summary`
  never existed; `--category`/`--alarm` are repeatable).

### CI

- Trunk-based flow (develop/main sync workflows removed); windows test
  suite repaired after being silently red and added to required checks.
- Security gates enforce (no `continue-on-error`) — govulncheck caught
  GO-2026-5037 the day it was enabled, fixed by pinning toolchain
  go1.26.4.
- errcheck enabled repo-wide with committed no-truncation lint caps;
  Sonar workflow fails loudly when coverage or lint reports are missing;
  released Docker images report the real tag version instead of "dev".

## [v0.5.0] - 2025-12-29

Last release before the remediation milestone. See git history.
