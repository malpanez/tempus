# Roadmap: Tempus

**Milestone:** v1.0 -- Tempus Mejorado
**Created:** 2026-03-29
**Granularity:** Standard

## Overview

Tempus is a working CLI with ~79% test coverage and a 3,900-line monolith. This roadmap fixes data-integrity bugs first, then builds the first-run experience (init wizard + templates + env vars), delivers the flagship interactive mode with a proper CLI package structure, polishes conflict resolution and prep time UX, and finishes with ND feature extraction and performance optimization. Refactor requirements are distributed across phases where they touch the same code, not lumped at the end.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Bug Fixes & Test Infrastructure** - Fix data-integrity bugs and establish testable output patterns
- [ ] **Phase 2: First-Run Experience** - Init wizard, env vars, config validation, and practical templates
- [ ] **Phase 3: Interactive Mode & CLI Structure** - Flagship --interactive mode with charmbracelet/huh and monolith split into internal/cli/
- [ ] **Phase 4: UX Polish** - Conflict resolution guidance and customizable prep time
- [ ] **Phase 5: ND Extraction & Performance** - Extract neurodivergent features to internal/nd/ and optimize batch performance

## Phase Details

### Phase 1: Bug Fixes & Test Infrastructure
**Goal**: Users get correct ICS output for all locales, proper error messages, and consistent input normalization across create and batch paths
**Depends on**: Nothing (first phase)
**Requirements**: BUG-01, BUG-02, BUG-03, BUG-04, BUG-05, REF-06, REF-02
**Note**: REF-02 scope in Phase 1 = only apply normalizeDateTimeInput() in create path. Full unification of 13 parsing functions into internal/parsing → Phase 3.
**Success Criteria** (what must be TRUE):
  1. User creates an event titled "Reunion de equipo" and the accented characters survive emoji processing intact
  2. User runs `tempus create --interactive --language en` and all alarm prompts appear in English (not hardcoded Spanish)
  3. User references a non-existent alarm profile and receives an error listing available profiles
  4. User enters an unrecognized city as timezone and receives an error suggesting `tempus timezone list --search`
  5. User runs `tempus create --start 2025/12/16 --time 09:00` and the date/time is parsed correctly (same as batch)
**Plans:** 3 plans

Plans:
- [ ] 01-01-PLAN.md — Unicode/emoji fix (BUG-01) + testable output infrastructure (REF-06)
- [ ] 01-02-PLAN.md — Error handling (BUG-03, BUG-04) + alarm prompt i18n (BUG-02)
- [ ] 01-03-PLAN.md — Input normalization in create path (BUG-05, REF-02)

**UI hint**: no

---

### Phase 2: First-Run Experience
**Goal**: A new user can install Tempus, run `tempus init`, and be fully configured with working env var overrides, validated config, and practical templates for their real workflows
**Depends on**: Phase 1
**Requirements**: CONF-01, CONF-02, CONF-03, UX-01, TMPL-01, TMPL-02, TMPL-03
**Success Criteria** (what must be TRUE):
  1. User runs `tempus init` and completes a wizard that auto-detects timezone/language and sets up config file with alarm profile
  2. User sets `TEMPUS_TIMEZONE=America/New_York` and Tempus uses that timezone without any config file change
  3. User runs `tempus config set timezone Invalid/Zone` and receives a validation error before config is saved
  4. User runs `tempus config set output_dir /nonexistent` and receives a validation error about the directory
  5. User runs `tempus batch template school-event`, `recruiter-meeting`, or `travel-day` and gets a usable CSV/YAML template with relevant fields
**Plans**: TBD

Plans:
- [ ] 02-01: TBD
- [ ] 02-02: TBD
- [ ] 02-03: TBD

**UI hint**: yes

---

### Phase 3: Interactive Mode & CLI Structure
**Goal**: Users can create events step-by-step with `tempus create --interactive` using a guided form with progress indicators, powered by charmbracelet/huh, with the command code living in a proper internal/cli/ package
**Depends on**: Phase 2
**Requirements**: UX-02, REF-01
**Success Criteria** (what must be TRUE):
  1. User runs `tempus create --interactive` and is guided through event creation with visible step progress ("Step 2/7") ending with a summary confirmation before ICS generation
  2. All existing CLI commands (`create`, `batch`, `lint`, `config`, `template`, `timezone`, `rrule`) continue to work with identical flags and output after the monolith split
  3. `main.go` is reduced to approximately 100 lines of wiring code, with command logic in `internal/cli/<command>.go` files
  4. Test coverage remains at or above 79% after the restructure
**Plans**: TBD

Plans:
- [ ] 03-01: TBD
- [ ] 03-02: TBD
- [ ] 03-03: TBD

**UI hint**: yes

---

### Phase 4: UX Polish
**Goal**: Users get actionable guidance when conflicts are detected and can customize prep time event naming
**Depends on**: Phase 3
**Requirements**: UX-03, UX-04
**Success Criteria** (what must be TRUE):
  1. User runs `tempus batch --check-conflicts` with a file containing overlapping events and sees exactly which events conflict (names, times, overlap duration) — no false promise of reading external calendar state
  2. User sets `prep_time_prefix` in config and new prep time events use that custom prefix instead of "Preparation"
  3. User runs `tempus batch` with `--prep-label "Setup"` and prep time events use "Setup" as their prefix
**Plans**: TBD

Plans:
- [ ] 04-01: TBD
- [ ] 04-02: TBD

**UI hint**: no

---

### Phase 5: ND Extraction & Performance
**Goal**: Neurodivergent features (spellcheck, conflicts, prep time, emoji) live in their own testable package and batch processing runs significantly faster on large datasets
**Depends on**: Phase 3 (needs cli/ structure from REF-01)
**Requirements**: REF-03, REF-04, REF-05
**Success Criteria** (what must be TRUE):
  1. `internal/nd/` package exists with extracted spellcheck, conflict detection, prep time, and emoji functions, each with their own tests
  2. Conflict detection on 1000+ events completes in O(n log n) time using sweep-line algorithm (measurably faster than current O(n^2))
  3. Batch spell checking with 100+ records reuses a precomputed distance matrix instead of recalculating per record
  4. Test coverage remains at or above 79% after extraction
**Plans**: TBD

Plans:
- [ ] 05-01: TBD
- [ ] 05-02: TBD
- [ ] 05-03: TBD

**UI hint**: no

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Bug Fixes & Test Infrastructure | 0/3 | Planning complete | - |
| 2. First-Run Experience | 0/3 | Not started | - |
| 3. Interactive Mode & CLI Structure | 0/3 | Not started | - |
| 4. UX Polish | 0/2 | Not started | - |
| 5. ND Extraction & Performance | 0/3 | Not started | - |
