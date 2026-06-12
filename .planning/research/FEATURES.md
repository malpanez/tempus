# Features Research

**Project:** Tempus (ICS calendar CLI for neurodivergent users)
**Researched:** 2026-03-29
**Overall confidence:** MEDIUM (training data only -- no live web verification possible)

---

## ICS/Calendar CLI Ecosystem

### Direct Competitors / Comparable Tools

| Tool | Language | What It Does | Key Features Tempus Lacks | Notes |
|------|----------|------------|---------------------------|-------|
| **khal** | Python | TUI calendar client, reads/writes ICS, syncs via vdirsyncer | Two-way sync, TUI browsing, event editing, recurring event display | Most mature CLI calendar tool. Different scope (client vs generator) but sets user expectations. |
| **calcurse** | C | TUI calendar/todo app with ICS import/export | Full TUI, todo integration, appointment editing, CalDAV sync | Desktop-oriented, not a generator. Shows what "calendar CLI" means to most users. |
| **gcalcli** | Python | Google Calendar CLI | Cloud sync, agenda view, natural language input | Vendor-locked (Google), but its NLP input and agenda view are UX benchmarks. |
| **ical-generator** | Node.js | Library for generating ICS programmatically | Not a CLI -- library only. Shows the programmatic ICS generation space. | |
| **rem2ics** / **remind** | C | Remind format to ICS converter | Scripting-friendly, piping model | Unix philosophy approach. Different paradigm. |
| **khard** | Python | vCard CLI (not calendar, but same ecosystem) | Contact management pattern similar to what a full calendar CLI would look like | |

### What Tempus Already Does Better Than Most

- **Neurodivergent-specific features** -- no competitor has ADHD alarm profiles, prep time buffers, overwhelm detection, or smart duration defaults. This is a genuine differentiator with no equivalent in the ecosystem.
- **Batch mode from CSV/JSON/YAML** -- most tools are single-event or require scripting.
- **Spell-checking and input normalization** -- unique in the ICS generator space.
- **Offline-first, stateless** -- unlike gcalcli or CalDAV clients.

### What Tempus Lacks vs Ecosystem Expectations

1. **No event viewing/listing** -- users who create events want to see them. Even a simple `tempus list *.ics` that parses and displays would bridge this gap.
2. **No import/edit of existing ICS** -- marked out of scope, and correctly so for this cycle, but users will ask.
3. **No CalDAV sync** -- correctly out of scope (offline-first is the value prop).
4. **No agenda/upcoming view** -- gcalcli's `agenda` command is a UX benchmark.

**Confidence:** MEDIUM. Based on training data up to early 2025. Tool landscape for ICS CLIs is slow-moving; unlikely to have changed significantly.

---

## Table Stakes Features

Features users expect from any calendar CLI tool. Missing = product feels incomplete.

### Must-Have (Tempus has these)

| Feature | Status | Notes |
|---------|--------|-------|
| Create single event with title, date, time, duration | Implemented | Core `tempus create` |
| Timezone handling | Implemented | Explorer + smart defaults |
| Recurring events (RRULE) | Implemented | Interactive helper |
| Multiple reminders/alarms | Implemented | Profiles + custom |
| Output valid RFC 5545 ICS | Implemented | `tempus lint` validates |
| Configuration persistence | Implemented | `tempus config` |
| Dry-run / preview | Implemented | `--dry-run` flag |

### Must-Have (Tempus needs these)

| Feature | Priority | Why Expected | Complexity |
|---------|----------|--------------|------------|
| **Interactive mode for event creation** | HIGH | Users who forget flags need guided input. Every major CLI (gh, npm, cargo) has this. | Med -- survey library already in go.mod |
| **First-run wizard (init)** | HIGH | New users need guided setup (timezone, language, defaults). Without it, first experience is reading docs. | Med |
| **Input normalization in `create`** | HIGH | Already works in batch -- inconsistency is a bug, not a feature gap. | Low -- port existing logic |
| **Environment variable support** | MED | Standard for CI/CD and power users. Documented but broken. | Low -- Viper AutomaticEnv |
| **Conflict resolution guidance** | MED | Detecting conflicts without suggesting fixes is half the job. | Med |

### Differentiators (Tempus's unique value)

| Feature | Value Proposition | Status |
|---------|-------------------|--------|
| ADHD alarm profiles | Pre-configured multi-alarm patterns for time blindness | Implemented |
| Prep time auto-generation | Buffer time before events, category-aware | Implemented |
| Overwhelm detection | Daily event limit warnings | Implemented |
| Smart duration defaults | Category-based sensible durations | Implemented |
| Auto-emoji categories | Visual markers without cognitive effort | Implemented |
| Spell-checking | Catches common typos automatically | Implemented |
| Batch templates | Pre-filled scenarios for common use cases | Implemented |

---

## ADHD UX Patterns for CLIs

### Core Principles (from ADHD/accessibility research)

**1. Progressive Disclosure -- Show Only What's Needed Now**
- Never present all options at once. This causes decision paralysis.
- Default path should require zero flags. Advanced options discovered incrementally.
- Pattern: `tempus create` with no flags should launch interactive mode (or prompt "run with --interactive?").
- Anti-pattern: `--help` showing 30 flags with no grouping.

**2. Sensible Defaults Over Configuration**
- Every prompt should have a pre-filled default that works for 80% of cases.
- "Just press Enter" should produce a valid result.
- Pattern: Duration defaults by category (meeting=1h, medication=5m). Timezone from config. Language from config.

**3. Immediate Feedback -- Reduce Uncertainty**
- After every input, confirm what was understood. ADHD users second-guess themselves.
- Pattern: After entering a date, show "Tuesday, December 16, 2025" -- the human-readable version.
- Pattern: Show a summary before final confirmation: "Create: Team standup, Dec 16 09:00-09:30 Europe/Madrid, 3 alarms. Correct? [Y/n]"
- Anti-pattern: Silent success. Always show what was created and where.

**4. Forgiving Input -- Accept Messy Input**
- ADHD users type fast, make typos, use inconsistent formats.
- Accept: `10:30`, `1030`, `10.30` for times. `2025-12-16`, `2025/12/16`, `12/16/2025`, `dec 16` for dates.
- Already partially implemented in batch normalizer. Must extend to `create`.

**5. Escape Hatches -- Easy to Abort, Easy to Redo**
- Ctrl+C should always work cleanly (no partial files left behind).
- `--dry-run` is already an escape hatch. Good.
- Interactive mode: allow going back to previous question.

**6. Reduce Working Memory Load**
- Don't ask for information the tool already knows (timezone from config, language from config).
- Show context in prompts: "Start time (timezone: Europe/Madrid):" not just "Start time:".
- Number steps: "Step 2/6: When does it start?" -- progress indicators reduce anxiety.

**7. Visual Anchors**
- Use color, emoji, and formatting to make output scannable.
- Group related information. Use whitespace generously.
- Mark required fields explicitly (Tempus already does this with `*`).

**8. Error Recovery Over Error Prevention**
- Instead of "Invalid date format", show "Could not parse '16 dec'. Try: 2025-12-16 or 'dec 16 2025'"
- Offer to re-enter rather than aborting the entire flow.

**Confidence:** MEDIUM-HIGH. These patterns are well-established in accessibility literature and neurodivergent UX research. Not specific to CLIs but directly applicable.

---

## Interactive Mode Patterns (--interactive)

### What Users Expect from `--interactive`

Based on patterns from gh (GitHub CLI), npm init, cargo init, and similar tools:

1. **Step-by-step prompted input** -- one question at a time, not a form dump.
2. **Smart defaults pre-filled** -- press Enter to accept.
3. **Validation at each step** -- don't let me proceed with invalid input.
4. **Summary before commit** -- show everything, ask for confirmation.
5. **Ability to skip optional fields** -- Enter on empty = skip.
6. **Context-aware prompts** -- show timezone in time prompts, show parsed date in confirmation.

### Go Library Options for Interactive Prompts

| Library | Status | Pros | Cons | Recommendation |
|---------|--------|------|------|----------------|
| **AlecAivazis/survey/v2** | Already in go.mod | Familiar, works, in dependencies | Archived/unmaintained since 2023. No new features. | Use for now -- already a dependency, switching mid-cycle adds risk |
| **charmbracelet/huh** | Active, well-maintained | Modern, accessible, Bubble Tea ecosystem, great UX | New dependency. Different API from survey. | Consider for future cycle. Better long-term choice. |
| **charmbracelet/bubbletea** | Active | Full TUI framework, maximum flexibility | Overkill for prompted input. Complex model. | Not needed -- too heavy for this use case. |
| **manifoldco/promptui** | Maintained | Simple, lightweight | Less flexible than survey/huh | No advantage over survey which is already in deps. |

**Recommendation:** Use `survey/v2` for this cycle since it is already a dependency (PROJECT.md constraint: "no unnecessary dependencies"). Plan migration to `charmbracelet/huh` in a future cycle when survey becomes a blocker.

**Confidence on survey status:** MEDIUM. Survey/v2 was archived by early 2024 in my training data. The `huh` library from Charm was the recommended successor.

### Recommended Interactive Flow for `tempus create --interactive`

```
Step 1/7: Event title *
> Team standup

Step 2/7: Date * (formats: 2025-12-16, dec 16, tomorrow)
> dec 16
  --> Tuesday, December 16, 2025

Step 3/7: Start time * (timezone: Europe/Madrid)
> 9:00
  --> 09:00 CET

Step 4/7: Duration (default: 1h for meetings)
> 30m

Step 5/7: Description (optional, Enter to skip)
>

Step 6/7: Category (optional: work, health, medication, focus, personal)
> work

Step 7/7: Alarms (default: profile:adhd-default = -2h, -1h, -30m, -10m)
> [Enter]

--- Summary ---
  Title:    Team standup
  Date:     2025-12-16 (Tuesday)
  Time:     09:00 - 09:30 Europe/Madrid
  Category: work
  Alarms:   -2h, -1h, -30m, -10m (adhd-default)
  Output:   ./team-standup.ics

  Create this event? [Y/n]
```

Key UX details:
- Step counter ("2/7") reduces ADHD anxiety about "how long is this?"
- Immediate parsing feedback ("Tuesday, December 16")
- Defaults shown inline, accepted with Enter
- Summary uses clean formatting, not raw ICS fields
- Required fields marked with `*`

---

## Init Wizard Best Practices

### What Good Init Wizards Do (examples from npm init, cargo init, gh auth login, eslint --init)

**1. Detect before asking**
- Check if config already exists. If yes: "Config found at ~/.config/tempus/config.yaml. Reconfigure? [y/N]"
- Auto-detect timezone from system (`timedatectl` on Linux, system timezone on macOS).
- Auto-detect locale/language from `$LANG` or `$LC_ALL`.

**2. Minimal questions (3-5 max)**
- `npm init` asks 7 questions max. `gh auth login` asks 3.
- For Tempus init, recommended questions:
  1. **Timezone** -- pre-filled from system detection. "Detected: Europe/Madrid. Correct? [Y/n]"
  2. **Language** -- pre-filled from locale detection. "Detected: es. Correct? [Y/n]"
  3. **Output directory** -- default: current directory. "Where to save ICS files? [./]"
  4. **Default alarm profile** -- "Which alarm style? (1) ADHD default (2) Minimal (3) Custom"
  5. **Done** -- "Config saved to ~/.config/tempus/config.yaml"

**3. Show what was created**
- Print the config file path.
- Print a "next steps" hint: "Try: tempus create --interactive"

**4. Idempotent -- safe to re-run**
- Running `tempus init` again should show current values as defaults.
- Never silently overwrite without asking.

### Recommended `tempus init` Flow

```
Welcome to Tempus -- calendar events without the cognitive tax.

Detecting your system settings...

  Timezone: Europe/Madrid (from system)
  Language: es (from LANG=es_ES.UTF-8)

Step 1/4: Timezone
  Detected: Europe/Madrid. Use this? [Y/n] y

Step 2/4: Language
  Detected: Spanish (es). Use this? [Y/n] y

Step 3/4: Default output directory
  Where to save .ics files? [./]: ~/calendar/

Step 4/4: Default alarm profile
  How should reminders work?
  > (1) ADHD friendly   -- 2h, 1h, 30m, 10m before (recommended)
    (2) Standard         -- 15m before
    (3) Minimal          -- no alarms by default
    (4) Custom           -- configure your own

Config saved to ~/.config/tempus/config.yaml

Next steps:
  tempus create --interactive    Create your first event
  tempus batch template          Generate a batch template
  tempus config list             View your settings
```

Key patterns:
- Auto-detection reduces questions from 4 to 2 confirmations.
- Numbered options with descriptions for alarm profiles.
- "Recommended" marker guides choice without forcing it.
- Next steps section prevents the "now what?" moment.

**Confidence:** HIGH. Init wizard patterns are well-established across ecosystems (npm, cargo, gh, eslint, prettier, commitlint). The patterns are stable and well-documented.

---

## Conflict Resolution UX

### Current State in Tempus

Tempus detects conflicts (`--check-conflicts`) and warns about overloaded days (`--max-events-per-day`). But it only reports problems without suggesting solutions. The PROJECT.md lists "F4: Conflict resolution guidance" as an active requirement.

### What Good Conflict Resolution Looks Like

**1. Show the conflict clearly**

```
CONFLICT DETECTED:
  Existing:  Team standup      09:00-09:30
  New:       Doctor appointment 09:15-10:15
  Overlap:   15 minutes (09:15-09:30)
```

**2. Suggest concrete alternatives**

```
Suggestions:
  (1) Move new event to 09:30 (right after standup)
  (2) Move new event to 10:00 (30min gap for transition)
  (3) Keep as-is (overlapping)
  (4) Cancel creation
```

Key UX principles for conflict resolution:
- **Don't just report -- guide.** "Conflict found" without alternatives triggers ADHD paralysis.
- **Offer the easiest fix first.** "Move to 09:30" is simpler than "reschedule both."
- **Include a transition buffer option.** ADHD users need context-switching time. Suggesting "10:00 (30min gap)" acknowledges this.
- **Always allow override.** Some conflicts are intentional (travel overlap, etc.).
- **In batch mode: report all conflicts upfront**, don't fail on first one. Show a summary table.

**3. Batch mode conflict summary**

```
Conflict Report (3 conflicts in 47 events):

  Dec 16: Team standup (09:00-09:30) overlaps Doctor (09:15-10:15)
          Suggestion: Move Doctor to 09:30 or 10:00

  Dec 17: Focus block (14:00-16:00) overlaps Meeting (15:00-16:00)
          Suggestion: Shorten focus block to 14:00-15:00

  Dec 20: 6 events scheduled (limit: 5)
          Suggestion: Consider moving Gym (18:00) to Dec 21

  Proceed anyway? [y/N]
```

**4. What NOT to do**

- Don't auto-resolve conflicts. The user must choose.
- Don't block creation entirely -- always offer "keep as-is" option.
- Don't require re-entering all event details to adjust time -- just the time field.
- Don't show conflicts in a wall of text -- use structured formatting.

**Confidence:** MEDIUM. Conflict resolution UX patterns are well-understood in calendar applications (Google Calendar, Outlook), but CLI-specific implementations are rare. Adapting GUI patterns to CLI is the novel part.

---

## Anti-features (things to NOT build)

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Full TUI calendar view** | Scope creep. Tempus is a generator, not a client. khal/calcurse already do this. Building a TUI is a multi-month effort that dilutes the core value prop. | Recommend `khal` for viewing. Focus on `tempus list` (simple text output) if needed later. |
| **CalDAV / Google Calendar sync** | Vendor lock-in, authentication complexity, shifts from offline-first to cloud-dependent. Fundamentally changes the tool's identity. | Stay offline-first. Users import ICS files into their preferred client. |
| **Event editing / `tempus edit`** | Parsing arbitrary ICS files is complex (VTIMEZONE variants, proprietary extensions). High effort, moderate value. | If needed, `tempus create` with same filename + `--overwrite` flag is simpler. |
| **Natural language date parsing (beyond basic)** | `olebedev/when` is already in go.mod for basic NLP. Going beyond basic ("next Tuesday at 3") into complex NLP ("the Tuesday after Thanksgiving") adds fragility and locale complexity. | Keep basic NLP via `olebedev/when`. Accept ISO dates, short formats (dec 16), and relative (tomorrow, next monday). Don't try to be a full NLP engine. |
| **Plugin / extension system** | Premature abstraction. User base is small. Maintenance burden of a plugin API is disproportionate. | Config-based customization (spell corrections, alarm profiles) covers power user needs. |
| **Notification / reminder daemon** | Running a background process is a fundamentally different tool. Calendar apps handle reminders. | Generate ICS with proper VALARM. Let the calendar app handle notifications. |
| **Multi-user / sharing features** | Tempus is a personal tool. Sharing adds auth, permissions, conflict resolution between users. | Generate ICS files. Users share files however they want (email, Slack, etc.). |
| **Database / event storage** | Tempus is stateless by design. Adding a DB means migrations, backup, corruption recovery. | ICS files on disk ARE the storage. Filesystem is the database. |

---

## Feature Dependencies (for roadmap ordering)

```
B1 (stripEmoji fix) --> standalone, no deps
B2 (promptAlarmField i18n) --> standalone, no deps
B3 (alarm profile error) --> standalone, no deps
B4 (cityToIANA fix) --> standalone, no deps
B5 (normalize in create) --> R2 (unified date parsing) benefits this but not required

F1 (tempus init) --> F3 (env vars) should land first or simultaneously
F2 (--interactive) --> B5 (normalization) should land first, uses survey/v2
F3 (env vars) --> standalone, Viper AutomaticEnv
F4 (conflict guidance) --> R4 (O(n log n) optimization) is nice-to-have but not required
F5 (prep time customize) --> standalone
F6 (config set validation) --> standalone

R1 (refactor main.go) --> enables all other work, highest leverage
R2 (unify date parsing) --> B5 depends on this conceptually
R3 (centralize defaults) --> F1 and F2 benefit from this
R4 (conflict optimization) --> F4 benefits from this
R5 (Levenshtein cache) --> standalone perf improvement
R6 (abstract output) --> improves testability for everything
```

### Recommended Priority Order

1. **Bugs first** (B1-B5) -- broken things before new things. ADHD users abandon tools that feel broken.
2. **R1 (refactor)** -- in parallel with bugs, as code is touched. Enables cleaner feature work.
3. **F3 (env vars)** -- low effort, high value, unblocks F1.
4. **F1 (tempus init)** -- first-run experience is make-or-break for adoption.
5. **F2 (--interactive)** -- the flagship ADHD feature. Biggest UX improvement.
6. **F4 (conflict guidance)** -- completes the conflict detection story.
7. **F5 (prep time customize)** -- polish feature, lower urgency.
8. **R2-R6** -- ongoing refactors woven into feature work.

---

## Sources

All findings are based on training data (cutoff early 2025). Live verification was not possible due to tool access restrictions.

- khal: https://github.com/pimutils/khal (Python CLI calendar)
- calcurse: https://calcurse.org/ (TUI calendar/todo)
- gcalcli: https://github.com/insanum/gcalcli (Google Calendar CLI)
- AlecAivazis/survey: https://github.com/AlecAivazis/survey (Go prompt library -- archived)
- charmbracelet/huh: https://github.com/charmbracelet/huh (Go forms library -- active)
- olebedev/when: https://github.com/olebedev/when (Natural language date parsing for Go)
- ADHD UX: W3C WCAG cognitive accessibility guidelines, ADDitude magazine UX research
- Init wizard patterns: npm init, cargo init, gh auth login, eslint --init
