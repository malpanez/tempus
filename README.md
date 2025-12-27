# Tempus

**A neurodivergent-friendly calendar tool that actually gets it.**

Create [RFC 5545](https://www.rfc-editor.org/rfc/rfc5545)-compliant ICS calendars with smart timezone handling, batch operations, and features specifically built to reduce cognitive load, fight time blindness, and support executive function.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/malpanez/tempus)](https://goreportcard.com/report/github.com/malpanez/tempus)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=malpanez_tempus&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=malpanez_tempus)
[![Coverage](https://img.shields.io/badge/coverage-78.8%25-brightgreen)](https://github.com/malpanez/tempus/actions)
[![CI Status](https://github.com/malpanez/tempus/actions/workflows/ci.yml/badge.svg)](https://github.com/malpanez/tempus/actions/workflows/ci.yml)
[![Security Scan](https://github.com/malpanez/tempus/actions/workflows/security.yml/badge.svg)](https://github.com/malpanez/tempus/actions/workflows/security.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/malpanez/tempus/badge)](https://securityscorecards.dev/viewer/?uri=github.com/malpanez/tempus)
[![Go Version](https://img.shields.io/github/go-mod/go-version/malpanez/tempus)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/malpanez/tempus)](https://github.com/malpanez/tempus/releases/latest)

[Why Tempus?](#-why-tempus) • [Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Commands](#-command-reference) • [Documentation](#-documentation) • [Contributing](#-contributing)

---

## 🧠 Why Tempus?

Traditional calendar tools are built for neurotypical brains. **Tempus is different.**

### Built for ADHD, ASD, and Dyslexia

**Time Blindness Solutions:**
- ⏰ **Multiple countdown reminders** with alarm profiles (adhd-default: -2h, -1h, -30m, -10m)
- 📊 **Automatic prep time buffers** (15min before meetings, 20min before medical)
- ⚡ **Focus block transitions** (5min decompression after deep work)

**Reduce Cognitive Load:**
- ✨ **Auto-emoji categories** - Visual icons without thinking (💊 medication, 💼 work, 🏥 health)
- 🔧 **Smart spell checking** - Common typos fixed automatically (meetting→meeting, docter→doctor)
- 📝 **Flexible input** - Type `10:30` instead of `2025-12-20 10:30:00 Europe/Madrid`

**Prevent Overwhelm:**
- 🚦 **Conflict detection** - Catch overlapping events before they happen
- 📉 **Daily event limits** - Warnings when you over-schedule (customizable threshold)
- 👁️ **Dry-run mode** - Preview everything before creating

**Executive Function Support:**
- 📋 **Batch templates** - Pre-filled CSVs for common scenarios (medication, routines, meetings)
- 🎯 **Smart duration defaults** - Medication=5m, breakfast=30m, focus=2h (auto-detected)
- 🔄 **Reusable alarm profiles** - Type `profile:medication` instead of `-5m,-1m,0m` every time

### Why CLI?

Many neurodivergent individuals prefer keyboard-driven workflows:
- **Fewer distractions** than GUI apps with infinite click paths
- **Faster input** once you learn the patterns
- **Scriptable** for automation and consistency
- **Works anywhere** - local, private, no subscription

---

## ✨ Features

### Core Functionality
- **ADHD-friendly UX**: time-only input, human durations (`45m`, `1h30m`, `1:15`, `-1d`, `-1w`), multiple alarms, required prompts marked with `*`.
- **Multilingual**: English (`en`), Spanish (`es`), Portuguese (`pt`), Irish/Gaeilge (`ga`).
- **Smart timezones**: start/end can use different TZs; timezone explorer with search and country filters.
- **Batch mode**: create one calendar from many events via CSV, JSON, or YAML.
- **Templates**: built-in (flight, meeting, holiday, medical, ADHD-friendly focus/medication/transition/deadline) plus external JSON/YAML.
- **Universal compatibility**: ICS files work with Google Calendar, Outlook, Apple Calendar, and any [RFC 5545](https://www.rfc-editor.org/rfc/rfc5545)-compliant app.
- **[RFC 5545](https://www.rfc-editor.org/rfc/rfc5545) compliance**: proper `TZID`, `VALARM`, recurrence (`RRULE`/`EXDATE`), and line folding for maximum compatibility.

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

## 📘 Command Reference

### `tempus create` - Single Event Creation

Create a single calendar event with full control over all properties.

**Basic usage:**
```bash
tempus create "Event Name" \
  --start "2025-03-15 10:00" \
  --duration "1h" \
  --start-tz "Europe/Madrid" \
  -o event.ics
```

**All flags:**
- `--start`, `-s` **(required)**: Start date/time (YYYY-MM-DD HH:MM) or time-only (HH:MM for today)
- `--end`, `-e`: End date/time OR duration (e.g. 1h30m, 90m, 1:15)
- `--duration`: Duration (alternative to --end, e.g. 45m, 1h30m, 90)
- `--start-tz`: Start timezone (e.g. Europe/Madrid, America/New_York)
- `--end-tz`: End timezone (for events spanning multiple timezones)
- `--all-day`, `-a`: All-day event (ignores time components)
- `--location`, `-L`: Event location
- `--description`, `-d`: Event description (multi-line supported with \n)
- `--category`: Category labels (repeat flag for multiple, e.g. --category work --category meeting)
- `--attendee`: Attendee email addresses (repeat for multiple)
- `--alarm`: Reminders (repeat for multiple, see alarm formats below)
- `--rrule`: Recurrence rule (e.g. FREQ=WEEKLY;COUNT=10)
- `--exdate`: Exclude specific dates (repeat for multiple)
- `--priority`: Event priority (1-9, where 1=highest)
- `--interactive`, `-i`: Launch interactive mode with prompts
- `--output`, `-o`: Output file path (default: stdout)

**Alarm formats:**
```bash
# Simple duration before event
--alarm 15m        # 15 minutes before
--alarm 1h         # 1 hour before
--alarm 1d         # 1 day before

# With custom description
--alarm "trigger=-30m,description=Boarding Pass"

# Absolute time
--alarm "trigger=2025-03-01 09:15,description=Check-in"

# After event (positive trigger)
--alarm "trigger=+10m,description=Wrap up"
```

**Examples:**

Time-only input (defaults to today):
```bash
tempus create "Focus Block" \
  --start "10:30" \
  --duration "2h" \
  --start-tz "Europe/Dublin" \
  -o focus.ics
```

All-day event:
```bash
tempus create "Holiday" \
  --start "2025-07-01" \
  --end   "2025-07-03" \
  --all-day \
  --start-tz "Europe/Dublin" \
  -o holiday.ics
```

Multi-timezone event (flight):
```bash
tempus create "Flight MAD→NYC" \
  --start "2025-03-15 10:00" \
  --start-tz "Europe/Madrid" \
  --end "2025-03-15 13:00" \
  --end-tz "America/New_York" \
  --location "Airport" \
  -o flight.ics
```

Weekly recurring with exceptions:
```bash
tempus create "Weekly Retro" \
  --start "2025-04-01 16:00" \
  --duration "1h" \
  --start-tz "Europe/Madrid" \
  --rrule "FREQ=WEEKLY;COUNT=6" \
  --exdate "2025-04-29 16:00" \
  --alarm 15m \
  -o retro.ics
```

Interactive mode (prompts for all fields):
```bash
tempus create --interactive
```

---

### `tempus lint` - Validate ICS Files

Validate ICS calendar files for common issues and RFC 5545 compliance.

**Usage:**
```bash
tempus lint --file calendar.ics
```

**Multiple files:**
```bash
tempus lint --file calendar1.ics --file calendar2.ics --file events.ics
```

**What it checks:**
- RFC 5545 compliance (required fields, proper formatting)
- Valid timezone identifiers (TZID)
- Proper date/time formats
- Valid RRULE syntax
- VALARM consistency
- Line folding correctness

**Example output:**
```
✅ calendar.ics: Valid (12 events, 0 errors, 0 warnings)

⚠️  events.ics: Issues found
  Line 15: Invalid TZID "US/Eastern" - use "America/New_York"
  Line 42: RRULE missing required FREQ parameter
  Line 58: VALARM trigger format invalid

❌ broken.ics: Critical errors
  Missing required VCALENDAR component
  Invalid VEVENT: missing DTSTART
```

---

### `tempus locale` - Inspect Available Locales

View available languages and locale information.

**List all available locales:**
```bash
tempus locale list
```

**Example output:**
```
Available locales:
  en - English
  es - Spanish (Español)
  pt - Portuguese (Português)
  ga - Irish (Gaeilge)

Current locale: en

To change language:
  tempus config set language es
  # or use --language flag:
  tempus create --language es ...
```

**Use different language for a single command:**
```bash
tempus create "Reunión" --start "2025-03-15 10:00" --language es
tempus template create meeting --language pt
```

---

### `tempus version` - Show Version Information

Display version information and build details.

**Usage:**
```bash
tempus version
```

**Example output:**
```
tempus version 0.5.0
Built: 2025-03-14 10:23:45
Commit: a1b2c3d
Go: go1.24.0
Platform: linux/amd64
```

---

### `tempus completion` - Shell Autocompletion

Generate shell completion scripts for faster command-line usage.

**Bash:**
```bash
# Generate completion script
tempus completion bash > /etc/bash_completion.d/tempus

# Or add to your ~/.bashrc:
source <(tempus completion bash)
```

**Zsh:**
```bash
# Generate completion script
tempus completion zsh > "${fpath[1]}/_tempus"

# Or add to your ~/.zshrc:
source <(tempus completion zsh)
```

**Fish:**
```bash
tempus completion fish > ~/.config/fish/completions/tempus.fish
```

**PowerShell (Windows):**
```powershell
tempus completion powershell | Out-String | Invoke-Expression

# Or add to your PowerShell profile:
tempus completion powershell >> $PROFILE
```

**What autocompletion provides:**
- Tab-complete command names (`tempus cre<TAB>` → `tempus create`)
- Tab-complete flag names (`--sta<TAB>` → `--start`)
- Timezone suggestions (`--start-tz Europe/<TAB>` shows European timezones)
- Template name completion
- File path completion

---

### `tempus config` - Manage Configuration

Manage persistent configuration settings (stored in `~/.config/tempus/config.yaml`).

**List all configuration:**
```bash
tempus config list
```

**Example output:**
```
Configuration:
  timezone: Europe/Madrid
  language: es

Config file: /home/user/.config/tempus/config.yaml
```

**Set configuration values:**
```bash
# Set default timezone (like git config)
tempus config set timezone "Europe/Madrid"

# Set default language
tempus config set language "es"

# Values persist across all tempus commands
```

**View available alarm profiles:**
```bash
tempus config alarm-profiles
```

**Example output:**
```
Available alarm profiles:

  adhd-default:
    -2h, -1h, -30m, -10m
    (Optimal spacing for regular events - recommended)

  adhd-countdown:
    -1d, -1h, -15m, -5m
    (For important deadlines/appointments)

  medication:
    -5m, -1m, 0m
    (Triple reminder for medication adherence)

  single:
    -15m
    (Standard single reminder)

  none:
    (No alarms)

Use in batch files: alarms: [profile:adhd-default]
```

**Advanced configuration (manual editing):**

For advanced settings, edit `~/.config/tempus/config.yaml` directly:

```yaml
# Default settings
timezone: Europe/Madrid
language: es

# Custom alarm profiles
alarm_profiles:
  my-default: ["-15m", "-5m", "-1m"]
  urgent: ["-1d", "-12h", "-1h", "-15m", "-5m"]

# Custom spell corrections
spell_corrections:
  # Built-in corrections are included automatically
  # Add your own:
  focusblock: focus block
  standup: stand-up
  tmrw: tomorrow
  # Language-specific:
  reunión: reunion
  médico: medico
```

**Configuration file locations:**
- Linux/macOS: `~/.config/tempus/config.yaml`
- Windows: `%APPDATA%\tempus\config.yaml`

**Priority order (highest to lowest):**
1. Command-line flags (`--timezone`, `--language`)
2. Environment variables (`TEMPUS_TIMEZONE`, `TEMPUS_LANGUAGE`)
3. Config file (`~/.config/tempus/config.yaml`)
4. Built-in defaults

---

## Importing to Calendar Apps

Tempus generates standard ICS files that work with any calendar application. Simply import the `.ics` file:

**Google Calendar:**
1. Open [Google Calendar](https://calendar.google.com/)
2. Click ⚙️ (Settings) → Import & Export
3. Select your `.ics` file → Choose destination calendar → Import

**Outlook:**
1. File → Open & Export → Import/Export
2. Select "Import an iCalendar (.ics) file"
3. Browse to your file → Import

**Apple Calendar:**
1. File → Import
2. Select your `.ics` file → Choose calendar → Import

**No API setup, no OAuth, no complexity** - just create and import!

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

## 📚 Documentation

Comprehensive guides for all user types:

### Getting Started
- **[Quick Start Guide](docs/QUICK_START.md)** - Visual, step-by-step guide with diagrams and flowcharts
  - Perfect for ADHD/ASD/Dyslexia users
  - ASCII diagrams and decision trees
  - Common patterns with visual examples
  - Troubleshooting with clear solutions

### Feature Guides
- **[Neurodivergent Features Guide](docs/NEURODIVERGENT_FEATURES.md)** - Complete feature documentation
  - Conflict detection and overwhelm prevention
  - Input normalization and spell checking
  - Alarm profiles and smart defaults
  - Visual aids and batch templates
  - Configuration examples

### Reference
- **[Command Reference](#-command-reference)** - Complete command documentation (this README)
  - All commands with flags and examples
  - `create`, `batch`, `quick`, `template`, `rrule`, `timezone`, `lint`, `locale`, `config`, `version`, `completion`
- **[Configuration Example](config.example.yaml)** - YAML config template with comments
- **[Batch Examples](examples/README.md)** - Ready-to-use CSV/JSON/YAML templates

### Advanced Topics
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development guidelines
- **[SECURITY.md](SECURITY.md)** - Security policy and responsible disclosure
- **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** - Community guidelines

---

## 🤝 Contributing

We welcome contributions that help neurodivergent users!

**Ways to contribute:**
- 🐛 Report bugs or usability issues
- ✨ Suggest neurodivergent-friendly features
- 📝 Improve documentation
- 🌐 Add translations
- 🧪 Write tests
- 💼 Share batch templates

**Before contributing:**
- Read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
- Follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) (neurodivergent-friendly)
- Check existing issues/PRs to avoid duplicates

**Security issues:** See [SECURITY.md](SECURITY.md) for responsible disclosure.

---

## 🙏 Acknowledgments

This project was built with [Claude Code](https://claude.com/claude-code), combining lived experience with neurodivergence and modern AI-assisted development.

**Research & Inspiration:**
- [ADHD Prospective Memory Research](https://www.nature.com/articles/s41598-025-08944-w) - Optimal reminder spacing
- [Time Blocking for ADHD](https://akiflow.com/blog/time-blocking-adhd) - Prep time buffers
- [ADHD Time Management](https://www.healthline.com/health/adhd/how-to-time-block-with-adhd) - 15-minute transitions

---

## 📄 License

[MIT License](LICENSE) - Use freely, commercially or personally.

---

## ⭐ Support

If Tempus helps you manage your calendar better, please consider:
- ⭐ Starring the repo on GitHub
- 🐛 Reporting bugs or usability issues
- 💬 Sharing your experience (Reddit, Twitter, Hacker News)
- 🤝 Contributing features or translations

**Made with ❤️ for the neurodivergent community.**

Even if only a few people use it, we've made their lives a little easier. That's success.
