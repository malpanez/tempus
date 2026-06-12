# Phase 2: First-Run Experience - Context

**Gathered:** 2026-03-30
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver the first-run experience: `tempus init` interactive wizard, env var overrides (CONF-01), config validation (CONF-02/03), and 3 practical batch templates tuned to real user workflows (school calendar, recruiter meetings, travel days).

**In scope:**
- `tempus init` — new command, interactive wizard using `survey/v2`, always in English
- CONF-01 — Viper `SetEnvPrefix("TEMPUS")` + `AutomaticEnv()` + `SetEnvKeyReplacer` for 5 env vars
- CONF-02 — `tempus config set timezone` validates IANA identifier before saving
- CONF-03 — `tempus config set output_dir` validates directory exists and is writable before saving
- TMPL-01 — `tempus batch template school-event` (CSV + YAML via `--format`)
- TMPL-02 — `tempus batch template recruiter-meeting` (CSV + YAML via `--format`)
- TMPL-03 — `tempus batch template travel-day` (CSV + YAML via `--format`)

**Out of scope:**
- `charmbracelet/huh` migration — Phase 3 (interactive mode)
- `tempus create --interactive` — Phase 3
- Config validation for fields other than timezone and output_dir

</domain>

<decisions>
## Implementation Decisions

### D-01: `tempus init` — Auto-detection + confirm
Auto-detect timezone from the OS (e.g., `timedatectl` on Linux or reading `/etc/timezone`) and language from the `LANG` environment variable. Present detected values to the user and ask them to confirm or change. The wizard always runs in **English** regardless of detected/configured language (avoids chicken-and-egg: language isn't configured yet).

Fields configured by the wizard (in order):
1. Timezone — auto-detected, user confirms or overrides
2. Language — detected from LANG env var, user selects from: en / es / pt / ga
3. Output directory — default: current dir (`.`), user types a path
4. Default alarm profile — user selects from available profiles (adhd-default, adhd-countdown, medication, single, none)

### D-02: `tempus init` — Existing config handling
If a config file already exists at the standard path (`~/.config/tempus/config.yaml` or equivalent), show:
`"Config already exists at {path}. Overwrite? [y/N]"`
If yes → run the full wizard. If no → exit cleanly with a message pointing to `tempus config set` for individual changes.

### D-03: `tempus init` — Completion output
After the wizard completes, show a summary table of saved values and a next-step hint:
```
✓ Config saved to ~/.config/tempus/config.yaml

  Timezone:      Europe/Madrid
  Language:      es
  Output dir:    ~/calendars
  Alarm profile: adhd-default

Next: tempus create --start today --duration 1h
```

### D-04: `tempus init` — Interactive only (no --yes flag)
No non-interactive mode for the wizard. Scripted/headless setups use env vars (CONF-01) or edit the config file directly. This keeps the Phase 2 scope clean.

### D-05: Config validation error messages
Consistent with Phase 1 error style (D-01/D-02 from 01-CONTEXT.md — fatal errors with actionable guidance):
- **CONF-02 (timezone):** `"Invalid timezone: 'Invalid/Zone'. Use 'tempus timezone list --search <name>' to find a valid IANA identifier."`
- **CONF-03 (output_dir):** `"Directory '/nonexistent' does not exist or is not writable."`

### D-06: `tempus config set` — show before+after
On successful `config set`, output: `"timezone: UTC → Europe/Madrid"`. Shows what changed at a glance.

### D-07: Env vars — all 5 exposed
Use Viper `SetEnvPrefix("TEMPUS")`, `AutomaticEnv()`, and `SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))` in `config.Load()`. The replacer maps dots and hyphens in config keys to underscores so Viper finds e.g. `TEMPUS_OUTPUT_DIR` → `output_dir`. Expose all 5:

| Env Var | Config Key | Type |
|---------|-----------|------|
| `TEMPUS_TIMEZONE` | `timezone` | string |
| `TEMPUS_LANGUAGE` | `language` | string |
| `TEMPUS_OUTPUT_DIR` | `output_dir` | string |
| `TEMPUS_DATE_FORMAT` | `date_format` | string |
| `TEMPUS_TIME_FORMAT` | `time_format` | string |

Priority (highest to lowest): CLI flags > env vars > config file > defaults. This is the standard Viper+Cobra behavior — no custom overriding needed.

### D-08: TMPL-01 — school-event template
Target use case: kids' school calendar — drop-off, pick-up, exams, vacations, trimester dates.

**CSV columns:**
```
summary,start_date,end_date,category,location,alarm,notes
```

- `summary`: Event name (e.g., "School starts Q3", "Emma pickup 17:00")
- `start_date`: ISO date or datetime
- `end_date`: ISO date or datetime (optional for single-day events)
- `category`: trimester | vacation | exam | activity | transport | holiday
- `location`: School name or address (e.g., "IES Cervantes")
- `alarm`: reminder, e.g., `single` or `-1d` (day-before for school-starts events)
- `notes`: Child name, additional context (e.g., "Emma — bring swimming kit")

**YAML format:** Same fields, YAML list-of-maps. Enabled via `tempus batch template school-event --format yaml`.

### D-09: TMPL-02 — recruiter-meeting template
Target use case: recruiter calls / job interviews — needs prep time, triple alarms, company context.

**CSV columns:**
```
summary,start_date,time,duration,timezone,alarm,add_prep_time,company,role,recruiter,notes
```

- `summary`: e.g., "Call with Sarah @ Acme Corp"
- `start_date`, `time`, `duration`: standard event time fields
- `timezone`: defaults to config timezone (recruiter may be in different TZ)
- `alarm`: pre-filled `adhd-default` (triple reminder pattern)
- `add_prep_time`: `true` (auto-generates prep time event)
- `company`: company name
- `role`: job title being discussed
- `recruiter`: recruiter name
- `notes`: contact info, LinkedIn, phone

**YAML format:** Available via `--format yaml`.

### D-10: TMPL-03 — travel-day template
Target use case: trip days — flights, transfers, hotel check-in, multi-timezone.

**CSV columns (multi-row template — one row per travel event):**
```
summary,start_date,time,end_time,timezone,destination_timezone,category,location,add_prep_time,alarm,notes
```

- `summary`: e.g., "MAD → LHR BA456", "Hotel check-in Hilton London"
- `timezone`: origin timezone for departure events, destination for arrival
- `destination_timezone`: arrival timezone (for timezone-crossing flights)
- `category`: flight | transfer | accommodation | activity
- `add_prep_time`: `true` for departure events (auto-generates "get to airport" buffer)
- `alarm`: pre-filled `-2h` for flights
- `location`: airport name / hotel address
- `notes`: flight number, booking reference, hotel address

Pre-filled rows in template: departure event, arrival event, hotel check-in, optional activity row.

**YAML format:** Available via `--format yaml`.

### D-11: `--format` flag for batch template
Add `--format` flag to `tempus batch template` with values: `csv` (default) | `yaml`.
Existing templates (basic, adhd-routine, etc.) keep CSV-only behavior unless `--format yaml` is specified. The 3 new templates support both from the start.

### Claude's Discretion
- Timezone auto-detection implementation (timedatectl / /etc/localtime symlink / go-tzlocal library — researcher should determine best approach without adding a heavy dependency)
- Survey prompt style (select vs input vs confirm — use what's already used in main.go L171)
- Config file path detection (already handled by `config.Load()` — reuse existing logic)
- YAML template formatting details (block style vs flow style, comment headers)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Context
- `.planning/PROJECT.md` — Core value, constraints (no API compat breaks, coverage ≥79%)
- `.planning/REQUIREMENTS.md` — CONF-01..03, UX-01, TMPL-01..03 acceptance criteria

### Primary Code Files
- `main.go` — `newBatchTemplateCmd()` (L1810), `getBatchTemplateContent()` (L1862), `runConfigSet` area — add `tempus init` command here
- `internal/config/config.go` — `Load()` function (L74) — add `AutomaticEnv()`, `SetEnvPrefix`, validation functions
- `internal/config/config_test.go` — existing tests; add validation tests here
- `internal/templates/` — existing template infrastructure (ddrender.go, scaffold.go, templates.go)

### Prior Phase Decisions
- `.planning/phases/01-bug-fixes-test-infrastructure/01-CONTEXT.md` — D-04: `var stdout io.Writer` pattern for output; D-01/D-02: error style (fatal + actionable guidance)

### i18n System
- `internal/i18n/i18n.go` — `Translator.T(key, args...)` — `tempus init` runs in English, no translation needed for wizard text, but new config keys may need i18n

### Existing Batch Templates (reference implementations)
- `main.go` L1862+ — `getBatchTemplateContent()` switch statement; new templates follow same pattern

No external specs — requirements fully captured in decisions above.

</canonical_refs>

<specifics>
## Specific Ideas

- School-event rows: include example rows for common school events (trimestre Q1 start, holiday, Emma pickup) so the template is immediately usable without editing headers
- Recruiter template: include a commented header row explaining each column, and 1 pre-filled example row
- Travel template: include 4 pre-filled rows (departure, arrival, hotel check-in, activity) as a realistic sample trip
- `tempus init` auto-detection: detect timezone from `/etc/localtime` symlink target (works on most Linux/macOS systems without external deps); fallback to asking explicitly
- Config validation for timezone: reuse `time.LoadLocation(tz)` which returns error for invalid IANA identifiers — stdlib, no deps
- Config validation for output_dir: `os.Stat(dir)` + check `d.Mode().IsDir()` + test write with `os.CreateTemp` — stdlib only
- `tempus config set` before+after output: write to `stdout` (var from Phase 1, D-04)

</specifics>

<deferred>
## Deferred Ideas

- Non-interactive `tempus init --yes` flag — user explicitly declined; use env vars for scripted setup
- `charmbracelet/huh` for the init wizard — Phase 3 migration
- Config validation for fields beyond timezone and output_dir (language, date_format, etc.) — can be added in Phase 4 if needed
- `tempus init --update` to selectively update one field — out of scope; `tempus config set` handles this

</deferred>

---

*Phase: 02-first-run-experience*
*Context gathered: 2026-03-30*
