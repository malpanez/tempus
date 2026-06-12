# Road to 1.0

Short, verifiable criteria. 1.0 ships when every box is checked — nothing
else gates it.

- [x] **Phase 5 merged** — programmatic VTIMEZONE for any IANA zone, zero
  hardcoded zones (v0.6.0).
- [x] **Manual import verified once in the three importing clients** —
  Outlook classic, Google Calendar, and Apple Calendar, with one ICS of
  each golden type (verified 2026-06-12).
- [x] **Wizard has real E2E coverage via PTY** — expect-style harness
  drives the binary through a pseudo-terminal (TERM=dumb accessible mode,
  answers written only after each prompt appears). First run immediately
  caught a real bug: huh v1.0 accessible mode ignores WithHideFunc, so the
  wizard demanded custom alarm offsets even when a profile was chosen —
  fixed by giving the wizard explicit staged control flow.
- [ ] **CLI surface frozen** — every flag and output format that exists
  works or doesn't exist. No decorative options, no silently ignored
  input. (Mostly done as of v0.6.0; freeze means no new exceptions.)
- [ ] **One real dogfooding cycle** — weeks of actual day-to-day use
  without new bugs filed against existing features.
