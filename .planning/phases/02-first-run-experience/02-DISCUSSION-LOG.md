# Phase 2: First-Run Experience - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-30
**Phase:** 02-first-run-experience
**Areas discussed:** tempus init wizard flow, Template field contents, Config validation feedback, Env var priority & scope

---

## `tempus init` Wizard Flow

### Auto-detection approach

| Option | Description | Selected |
|--------|-------------|----------|
| Auto-detect + confirm | Detect timezone and language, show to user, ask to confirm or change | ✓ |
| Always ask | Detected values as defaults, always require active input | |
| Auto-detect silently | No prompts for detectable fields | |

### Fields configured

All four: Timezone, Language, Output directory, Default alarm profile ✓

### Existing config handling

| Option | Description | Selected |
|--------|-------------|----------|
| Warn + confirm overwrite | "Config exists. Overwrite? [y/N]" | ✓ |
| Merge/update | Only change wizard-set fields | |
| Always overwrite silently | No prompt | |

### Completion output

| Option | Description | Selected |
|--------|-------------|----------|
| Summary + next steps | Table of saved values + `tempus create` hint | ✓ |
| Just 'Config saved' | Minimal | |
| Full config preview | Print entire config.yaml | |

### Non-interactive mode

| Option | Description | Selected |
|--------|-------------|----------|
| Interactive only | No --yes flag; scripted setup uses env vars | ✓ |
| --yes flag | Accept all auto-detected values | |

### Wizard language

| Option | Description | Selected |
|--------|-------------|----------|
| Always English | Avoids chicken-and-egg with language config | ✓ |
| Detected language | Use LANG env var | |
| Ask first | First wizard question is language | |

---

## Template Field Contents

### TMPL-01: school-event

All fields selected: event name + dates, category, school/child name, alarm.
User note: "Creo que todo importa, ayuda a no olvidarte de las cosas como por ejemplo la salida al colegio o la recogida de los escolares" (school drop-off and pick-up are key use cases).

Final columns: summary, start_date, end_date, category, location, alarm, notes

### TMPL-02: recruiter-meeting

All fields selected: company + role, prep time auto-generated, triple ADHD alarms, recruiter name/contact.

Final columns: summary, start_date, time, duration, timezone, alarm, add_prep_time, company, role, recruiter, notes

### TMPL-03: travel-day

All fields selected: flight info, origin + destination timezones, accommodation, buffer/transfer time.

Final columns: summary, start_date, time, end_time, timezone, destination_timezone, category, location, add_prep_time, alarm, notes

### Template format

| Option | Description | Selected |
|--------|-------------|----------|
| CSV only | Consistent with existing templates | |
| YAML only | More readable | |
| Both (--format flag) | CSV default, --format yaml for YAML | ✓ |

---

## Config Validation Feedback

### Error message style

| Option | Description | Selected |
|--------|-------------|----------|
| Error + suggest search | "Invalid timezone. Use tempus timezone list --search" | ✓ |
| Error only | Short, no guidance | |
| Error + show nearest matches | Fuzzy search candidates | |

### config set UX

| Option | Description | Selected |
|--------|-------------|----------|
| Show before+after | "timezone: UTC → Europe/Madrid" | ✓ |
| Just confirm new value | "timezone set to Europe/Madrid" | |
| Silent on success | No output | |

---

## Env Var Priority & Scope

### Which env vars

All 5 selected: TEMPUS_TIMEZONE, TEMPUS_LANGUAGE, TEMPUS_OUTPUT_DIR, TEMPUS_DATE_FORMAT, TEMPUS_TIME_FORMAT

### Priority

| Option | Description | Selected |
|--------|-------------|----------|
| Env > config file > defaults | Standard Viper AutomaticEnv() behavior | ✓ |
| CLI flag > env > config > defaults | Explicit (same behavior) | |
| Env > CLI flag > config > defaults | Non-standard, not recommended | |
