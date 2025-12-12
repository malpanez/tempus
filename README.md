# Tempus

**A neurodivergent-friendly calendar tool that actually gets it.**

Create RFC 5545-compliant ICS calendars with smart timezone handling, batch operations, and features specifically built to reduce cognitive load, fight time blindness, and support executive function.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/malpanez/tempus)](https://goreportcard.com/report/github.com/malpanez/tempus)
[![Coverage](https://img.shields.io/badge/coverage-78.8%25-brightgreen)](https://github.com/malpanez/tempus/actions)

[Why Tempus?](#-why-tempus) • [Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Contributing](#-contributing)

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
- **Universal compatibility**: ICS files work with Google Calendar, Outlook, Apple Calendar, and any RFC 5545-compliant app.
- **RFC 5545 compliance**: proper `TZID`, `VALARM`, recurrence (`RRULE`/`EXDATE`), and line folding for maximum compatibility.

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
