# Road to 1.0

Short, verifiable criteria. 1.0 ships when every box is checked — nothing
else gates it.

- [x] **Phase 5 merged** — programmatic VTIMEZONE for any IANA zone, zero
  hardcoded zones (v0.6.0).
- [ ] **Manual import verified once in the three importing clients** —
  Outlook classic, Google Calendar, and Apple Calendar, with one ICS of
  each golden type (simple, TZID+VTIMEZONE, alarms/profile, batch with
  exdates, recurring with UNTIL, all-day).
- [ ] **Wizard has real E2E coverage via PTY** — huh v1.0 accessible mode
  swallows piped stdin (first prompt's scanner buffers everything; gap
  documented in `internal/testutil/golden/harness.go`). B1 lived exactly
  there: the flagship flow needs tests that drive a real terminal.
- [ ] **CLI surface frozen** — every flag and output format that exists
  works or doesn't exist. No decorative options, no silently ignored
  input. (Mostly done as of v0.6.0; freeze means no new exceptions.)
- [ ] **One real dogfooding cycle** — weeks of actual day-to-day use
  without new bugs filed against existing features.
