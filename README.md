# Tempus

A multilingual, neurodivergent-friendly **ICS calendar generator** with smart timezone handling, batch CSV/JSON support, and Google Calendar import.

[Features](#features) • [Installation](#installation) • [Quick-start](#quick-start) • [Batch](#batch) • [Templates](#templates) • [Google](#google-calendar) • [Contributing](#contributing)

---

## Features

### Core Functionality
- **ADHD-friendly UX**: time-only input, human durations (`45m`, `1h30m`, `1:15`, `-1d`, `-1w`), multiple alarms, required prompts marked with `*`.
- **Multilingual**: English (`en`), Spanish (`es`), Portuguese (`pt`), Irish/Gaeilge (`ga`).
- **Smart timezones**: start/end can use different TZs; timezone explorer with search and country filters.
- **Batch mode**: create one calendar from many events via CSV, JSON, or YAML.
- **Templates**: built-in (flight, meeting, holiday, medical, ADHD-friendly focus/medication/transition/deadline) plus external JSON/YAML.
- **Google import**: device-flow OAuth to upload any `.ics` into Google Calendar.
- **RFC 5545 compliance**: proper `TZID`, `VALARM`, recurrence (`RRULE`/`EXDATE`), and line folding for compatibility.

### Neurodivergent-Friendly Enhancements
- **Batch Template Generator**: Pre-filled templates for common scenarios (`tempus batch template`)
- **Dry-Run Validation**: Preview and validate batch files before creating (`--dry-run`)
- **Conflict Detection**: Automatically detects overlapping events in batch mode (`--check-conflicts`)
- **Overwhelm Prevention**: Warns when any day exceeds event threshold (`--max-events-per-day N`)
- **Prep Time Auto-Addition**: Automatically adds preparation/transition buffers (`--add-prep-time`) - **ADHD time boxing**
  - 15min before meetings/appointments, 20min before medical events, 5min after focus blocks
- **Input Normalization**: Auto-fixes date/time formats (2025/12/16→2025-12-16, 0900→09:00)
- **Smart Spell Checking**: Corrects common typos in event summaries (meetting→meeting, docter→doctor, medicaton→medication)
  - **Customizable Dictionary**: Add your own corrections via `spell_corrections` in config.yaml
- **Alarm Profiles**: Reusable alarm presets (adhd-default, adhd-countdown, medication) - use `profile:name` in batch files
- **Smart Duration Defaults**: Auto-detects sensible durations based on event type and time (meds=5m, focus=2h, etc.)
- **Auto-Emoji Support**: Adds visual category icons automatically (💊 medication, 💼 work, 🏥 health, etc.)
- **RRULE Helper**: Interactive wizard to build recurrence rules without memorizing syntax (`tempus rrule`)

📖 **[Complete Neurodivergent Features Guide](docs/NEURODIVERGENT_FEATURES.md)** - Detailed documentation with examples and tips for ADHD, ASD, and Dyslexia users.

---

## Installation

### Prebuilt binaries (Recommended)

Download the latest release for your platform:

**[→ Download from GitHub Releases](https://github.com/malpanez/tempus/releases)**

Available for:
- **Linux**: AMD64, ARM64
- **macOS**: Intel (AMD64), Apple Silicon (ARM64)
- **Windows**: AMD64, ARM64

All releases are automatically built and tested via GitHub Actions CI/CD.

**Installation steps:**

Linux/macOS:
```bash
# Download the binary for your platform
# Extract and move to PATH
chmod +x tempus
sudo mv tempus /usr/local/bin/
```

Windows (PowerShell):
```powershell
# Download tempus.exe
# Move to a folder on PATH, e.g., C:\Users\<you>\bin
# Add to PATH if not already there
```

### From source (Go 1.24+)

Linux/macOS:
```bash
go mod tidy
go build -trimpath -ldflags "-s -w" -o build/tempus .
```

Windows (from PowerShell):
```powershell
go mod tidy
go build -trimpath -ldflags "-s -w" -o build\tempus.exe .
```

Cross-compile for Windows from Linux/macOS:
```bash
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/tempus.exe .
```

### Docker

```bash
docker pull ghcr.io/malpanez/tempus:latest
docker run --rm -v $(pwd):/data ghcr.io/malpanez/tempus:latest --help
```

---

## Quick-start

### Configure defaults (like Git)
```bash
tempus config set timezone "Europe/Madrid"
tempus config set language "es"
tempus config list
```

**Advanced configuration**: Copy [config.example.yaml](config.example.yaml) to `~/.config/tempus/config.yaml` and customize alarm profiles, spell corrections, and more.

### Create an event
```bash
tempus create "Team Meeting" \
  --start "2025-03-15 10:00" \
  --duration "1h" \
  --start-tz "Europe/Madrid" \
  --location "Conference Room A" \
  --attendee "alice@example.com" \
  -o meeting.ics
```

Time-only + duration (auto-expands to today):
```bash
tempus create "Focus Block" \
  --start "10:30" \
  --end   "1h30m" \
  --start-tz "Europe/Dublin" \
  -o focus.ics
```

All-day / multi-day:
```bash
tempus create "Holiday" \
  --start "2025-07-01" \
  --end   "2025-07-03" \
  --all-day \
  --start-tz "Europe/Dublin" \
  -o holiday.ics
```

Recurring with exceptions:
```bash
tempus create "Weekly Retro" \
  --start "2025-04-01 16:00" \
  --end   "2025-04-01 17:00" \
  --start-tz "Europe/Madrid" \
  --rrule "FREQ=WEEKLY;COUNT=6" \
  --exdate "2025-04-29 16:00" \
  -o retro.ics
```

Reminders (VALARM):
```bash
tempus create "Boarding" \
  --start "2025-03-01 10:00" \
  --duration "1h" \
  --start-tz "Europe/Madrid" \
  --alarm 30m \
  --alarm "trigger=+10m,description=Wrap up" \
  --alarm "trigger=2025-03-01 09:15,description=Airport check-in" \
  -o boarding.ics
```

---

## Batch

Generate many events into one calendar from CSV, JSON, or YAML:
```bash
tempus batch \
  --input examples/adhd-weekly-routine.csv \
  --output my-routine.ics \
  --name "Weekly Routine"
```

### Quick Start with Templates
Generate a pre-filled template to edit:
```bash
# See available template types
tempus batch template --help

# Generate templates
tempus batch template adhd-routine -o my-routine.csv
tempus batch template medication -o meds.yaml
tempus batch template work-meetings -o meetings.csv
tempus batch template travel -o trip.json

# Edit the file, then create calendar
tempus batch -i my-routine.csv -o calendar.ics
```

Available templates: `basic`, `adhd-routine`, `medication`, `work-meetings`, `medical`, `travel`, `family`

### Validate Before Creating
Preview and check for errors without creating output:
```bash
tempus batch --dry-run -i my-events.csv
# Shows event summary and catches errors early
```

### Conflict Detection and Overwhelm Prevention
Tempus helps prevent scheduling conflicts and over-scheduling:

**Detect overlapping events:**
```bash
tempus batch --check-conflicts -i my-events.csv -o calendar.ics
# ⚠️  Found 2 time conflict(s):
#   • 💼 Team meeting (09:00-10:00) overlaps with 🏥 Doctor appointment (09:45-11:00)
#   • 💼 Afternoon meeting (14:00-16:00) overlaps with 💼 Late meeting (15:00-16:00)
```

**Prevent overwhelm by limiting events per day:**
```bash
tempus batch --max-events-per-day 6 -i my-events.csv -o calendar.ics
# ⚠️  Days with high event load:
#   • Tuesday, Dec 16: 9 events (threshold: 6)
```

**Combine both in dry-run mode** (automatically enabled):
```bash
tempus batch --dry-run -i my-events.csv
# Automatically checks for conflicts and overwhelm (default threshold: 8 events/day)
```

### Input Normalization and Spell Checking
Tempus automatically fixes common input errors:

**Date/Time format normalization:**
- Converts slashes to dashes: `2025/12/16` → `2025-12-16`
- Pads single digits: `2025-1-5` → `2025-01-05`
- Handles time without colons: `0900` → `09:00`
- Pads hours: `9:00` → `09:00`

**Automatic spell correction** for common typos:
- `meetting` → `meeting`
- `docter` → `doctor`
- `medicaton` → `medication`
- `appointmnt` → `appointment`
- `brekfast` → `breakfast`
- `therepy` → `therapy`
- And 20+ more common corrections

**Customize the spell checker** - Add your own corrections in `~/.config/tempus/config.yaml`:
```yaml
spell_corrections:
  # Built-in corrections are included by default
  # Add your own:
  focusblock: focus block
  standup: stand-up
  # Language-specific corrections:
  reunión: reunion
  médico: medico
```

### ADHD Time Boxing: Automatic Prep Time

Tempus can automatically add preparation and transition buffers based on [ADHD time boxing research](https://akiflow.com/blog/time-blocking-adhd):

```bash
tempus batch --add-prep-time -i my-events.csv -o calendar.ics
```

**What it does:**
- **15min preparation** before meetings/appointments (mental prep + setup)
- **20min buffer** before medical events (travel, parking, check-in)
- **5min transition** after focus blocks (decompression, reset)

**Example:**
```csv
summary,start,duration,start_tz,categories
Team meeting,2025-12-20 14:00,1h,Europe/Madrid,work
Doctor appointment,2025-12-21 10:00,30m,Europe/Madrid,health
Focus block,2025-12-20 09:00,2h,Europe/Madrid,work
```

**Creates:**
- ⏰ Preparation: Team meeting (13:45-14:00)
- 💼 Team meeting (14:00-15:00)
- ⏰ Travel & arrival buffer: Doctor appointment (09:40-10:00)
- 🏥 Doctor appointment (10:00-10:30)
- 💼 Focus block (09:00-11:00)
- 🔄 Transition: Focus block (11:00-11:05)

**Why 15min buffers?** [Research shows](https://www.healthline.com/health/adhd/how-to-time-block-with-adhd) that 15-minute buffers prevent task derailment in ADHD, providing time for mental context switching.

### Alarm Profiles
Use reusable alarm presets instead of typing triggers every time:
```bash
# List available profiles
tempus config alarm-profiles

# In your CSV/JSON/YAML, use profile references:
# CSV: alarms column = "profile:adhd-triple"
# JSON: "alarms": ["profile:medication"]
# YAML: alarms: [profile:adhd-countdown]
```

Built-in profiles (evidence-based, neuroscience research 2024-2025):
- `adhd-default`: -2h, -1h, -30m, -10m (optimal spacing for regular events - **recommended**)
- `adhd-countdown`: -1d, -1h, -15m, -5m (for important deadlines/appointments)
- `medication`: -5m, -1m, 0m (triple reminder for medication)
- `single`: -15m (standard single reminder)
- `none`: no alarms

**Why these intervals?** Based on [ADHD prospective memory research](https://www.nature.com/articles/s41598-025-08944-w), optimal reminder spacing helps with strategic time monitoring and working memory deficits.

### Batch Features
- **Format auto-detected** (`--format csv|json|yaml|auto`)
- **Fields**: `summary`, `start`, `end`, `duration`, `start_tz`, `end_tz`, `location`, `description`, `all_day`, `rrule`, `exdate`, `categories`, `alarms`
- **Alarms**: Support `-15m`, `-1h`, `-1d`, `-1w` formats or profile references (`profile:adhd-triple`)
- **Smart defaults**: No duration? Auto-detects based on event type (meds=5m, breakfast=30m, focus=2h)
- **Auto-emoji**: Categories auto-add visual icons (Health→🏥, Work→💼, Medication→💊)
- **Input normalization**: Auto-fixes date/time formats (2025/12/16→2025-12-16, 0900→09:00)
- **Spell checking**: Common typos corrected automatically (meetting→meeting, docter→doctor, customizable)
- **Conflict detection**: Detects overlapping events with `--check-conflicts`
- **Overwhelm prevention**: Warns when days exceed event limit with `--max-events-per-day N`
- **Dry-run validation**: Preview events and catch errors before creating with `--dry-run`

**Ready-to-use examples** in `examples/`:
- `adhd-weekly-routine.csv` - Medication + focus blocks + transitions
- `work-meetings.csv` - Team meetings with recurrence
- `medical-appointments.csv` - Healthcare visits with prep reminders
- `travel-itinerary.json` - Complete trip with flights + hotels
- `family-calendar.csv` - School + activities
- `medication-schedule.yaml` - Multi-medication with triple alarms

Full guide: [examples/README.md](examples/README.md)

---

## Templates

Built-in templates (interactive): `flight`, `meeting`, `holiday`, `medical`, `focus-block`, `medication`, `appointment`, `transition`, `deadline`.

```bash
tempus template list

tempus template create flight
# or: meeting, holiday, medical, focus-block, medication, appointment, transition, deadline
```

Use external templates (JSON/YAML):
```bash
tempus template create my-template.yaml
```

---

## RRULE Helper

Don't know RRULE syntax? Use the interactive wizard:
```bash
tempus rrule
```

The wizard guides you through:
1. Frequency (daily, weekly, monthly, yearly)
2. Interval (every N occurrences)
3. Days of week (for weekly events)
4. End condition (never, after N times, or on a date)

Example output:
```
FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;COUNT=20
```

Copy this into your batch files or use with `--rrule` flag.

---

## Google Calendar

Device-flow import for any ICS file:
```bash
export TEMPUS_CLIENT_ID="<client-id>"
export TEMPUS_CLIENT_SECRET="<client-secret>"

# First run triggers OAuth device flow
tempus google import \
  --input event.ics \
  --calendar primary \
  --token-file ~/.tempus/google_token.json
```

Flags: `--client-id`, `--client-secret`, `--token-file`, `--calendar`. Token is cached and refreshed automatically. See `docs/GOOGLE_API_SETUP.md` for the full walkthrough.

---

## Timezone Explorer
```bash
tempus timezone list --country Spain
tempus timezone info Europe/Madrid
```

---

## Development

### Project Structure

```
main.go               # CLI commands
internal/calendar     # ICS generation
internal/config       # config handling
internal/normalizer   # date/time parsing
internal/templates    # templates & prompts
internal/prompts      # user interaction
internal/utils        # shared utilities
locales               # translations
timezones             # IANA data
```

### Local Development

```bash
# Run tests
go test ./...
go test -cover ./...
go test -race ./...

# Lint
golangci-lint run

# Build
go build -o build/tempus .
```

### CI/CD Pipeline

The project uses GitHub Actions for automated testing, security scanning, and releases:

**Continuous Integration** (`.github/workflows/ci.yml`):
- Runs on every push and pull request
- Tests on Linux, macOS, and Windows
- Runs `go test ./...`, `go vet`, and `golangci-lint`
- Tests with race detector (`-race`)
- Target: 75-80% code coverage

**Security Scanning** (`.github/workflows/security.yml`):
- Weekly automated scans
- Dependency vulnerability checking (Dependabot)
- CodeQL analysis
- gosec security scanner
- nancy (dependency vulnerability scanner)
- trivy (container scanning if using Docker)

**Automated Releases** (`.github/workflows/release.yml`):
- Triggered by pushing a tag: `git tag v0.5.0 && git push origin v0.5.0`
- Builds binaries for 6 platforms:
  - Linux (AMD64, ARM64)
  - macOS (Intel, Apple Silicon)
  - Windows (AMD64, ARM64)
- Creates GitHub Release with all binaries attached
- Publishes Docker image to GitHub Container Registry

### Git Workflow

We follow a **git-flow** branching model:

```
feature/fix branch --> develop --> main
```

**Branches:**
- `main`: Production-ready code, protected
- `develop`: Integration branch for features
- `feature/*`: New features
- `fix/*`: Bug fixes

**Process:**
1. Create feature branch from `develop`: `git checkout -b feature/my-feature develop`
2. Make changes and commit
3. Push and open PR to `develop`
4. CI runs automatically (tests, linting, security)
5. After approval and passing CI, merge to `develop`
6. Automated sync between `develop` and `main` after CI passes
7. Tag `main` for releases: `git tag v0.5.0 && git push origin v0.5.0`

**Automated checks:**
- All PRs must pass CI before merging
- Branch protection on `main` and `develop`
- Renovate bot for automatic dependency updates

---

## Contributing
- Read the [Code of Conduct](CODE_OF_CONDUCT.md) and [CONTRIBUTING](CONTRIBUTING.md).
- Issues and PRs welcome: ADHD-friendly features, templates, translations, tests, docs.

---

## License
MIT — see [LICENSE](LICENSE).

---

Made with ❤️ for the neurodivergent community. If Tempus helps you, consider starring the repo.
