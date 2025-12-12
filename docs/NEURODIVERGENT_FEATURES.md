# Neurodivergent-Friendly Features Guide

Tempus is designed with neurodivergent users in mind—especially those with ADHD, ASD (Autism Spectrum Disorder), and Dyslexia. This guide explains all neurodivergent-friendly features and how to use them effectively to reduce cognitive load, manage time blindness, and support executive function.

## Table of Contents
- [Conflict Detection](#conflict-detection)
- [Overwhelm Prevention](#overwhelm-prevention)
- [Input Normalization](#input-normalization)
- [Spell Checking](#spell-checking)
- [Alarm Profiles](#alarm-profiles)
- [Smart Defaults](#smart-defaults)
- [Visual Aids](#visual-aids)
- [Batch Templates](#batch-templates)
- [Dry-Run Validation](#dry-run-validation)

---

## Conflict Detection

### What It Does
Automatically detects overlapping events when creating calendars in batch mode.

### Why It Helps
- **ADHD**: Prevents double-booking and scheduling conflicts that can cause overwhelm
- **Executive Function**: Catches planning mistakes before they happen
- **Time Blindness**: Visual warnings about time overlaps

### How to Use

```bash
# Check for conflicts while creating calendar
tempus batch --check-conflicts -i my-events.csv -o calendar.ics
```

**Example Output:**
```
⚠️  Found 2 time conflict(s):
  • 💼 Team meeting (09:00-10:00) overlaps with 🏥 Doctor appointment (09:45-11:00)
  • 💼 Afternoon meeting (14:00-16:00) overlaps with 💼 Late meeting (15:00-16:00)

✅ Created: calendar.ics (9 events)
```

### Features
- Compares all events pairwise
- Skips all-day events (they don't cause time conflicts)
- Shows exact times and event names
- Works in both regular and dry-run mode

---

## Overwhelm Prevention

### What It Does
Warns when any single day has too many events scheduled.

### Why It Helps
- **ADHD**: Prevents over-scheduling and burnout
- **Energy Management**: Helps maintain sustainable schedules
- **Decision Fatigue**: Highlights when you might be planning too much

### How to Use

```bash
# Set a maximum of 6 events per day
tempus batch --max-events-per-day 6 -i my-events.csv -o calendar.ics
```

**Example Output:**
```
⚠️  Days with high event load:
  • Tuesday, Dec 16: 9 events (threshold: 6)
  • Friday, Dec 19: 8 events (threshold: 6)

✅ Created: calendar.ics (25 events)
```

### Default Behavior
- In **dry-run mode**: Automatically checks with threshold of 8 events/day
- In **regular mode**: Only checks if you specify `--max-events-per-day`
- Set to `0` to disable: `--max-events-per-day 0`

### Recommended Thresholds
- **High Executive Function Needs**: 4-5 events/day
- **Moderate**: 6-8 events/day
- **Light Monitoring**: 10-12 events/day

---

## Input Normalization

### What It Does
Automatically fixes common date and time formatting variations.

### Why It Helps
- **Dyslexia**: Don't worry about exact formatting
- **Cognitive Load**: Focus on content, not format rules
- **Consistency**: Mix different date formats in the same file

### Formats Supported

**Date Normalization:**
```
Input:          Output:
2025/12/16  →   2025-12-16
2025-1-5    →   2025-01-05
2025/1/5    →   2025-01-05
```

**Time Normalization:**
```
Input:          Output:
0900        →   09:00
900         →   09:00  (pads to 09:00)
9:00        →   09:00  (pads single digit)
14:30       →   14:30  (already correct)
```

**Combined:**
```
Input:                  Output:
2025/12/16 9:00     →   2025-12-16 09:00
2025-1-5 0830       →   2025-01-05 08:30
```

### How It Works
All normalization happens **automatically** when you use batch mode. No configuration needed!

```csv
summary,start,duration,start_tz,categories
Morning meds,2025/12/16 0800,5m,Europe/Madrid,medication
Doctor,2025-1-5 14:30,1h,Europe/Madrid,health
```

Both formats work perfectly and are normalized internally.

---

## Spell Checking

### What It Does
Automatically corrects common spelling errors in event summaries.

### Why It Helps
- **Dyslexia**: Don't stress about spelling
- **Fast Input**: Type quickly, let Tempus fix typos
- **Consistency**: All events have correctly spelled summaries

### Built-in Corrections

**Common Event Words:**
```
meetting    → meeting
meting      → meeting
meetng      → meeting

docter      → doctor
doctr       → doctor

medicaton   → medication
mediction   → medication
medikation  → medication

appointmnt  → appointment
apointment  → appointment

brekfast    → breakfast
breakfst    → breakfast
brek        → break
brk         → break

therepy     → therapy
theraphy    → therapy
sesion      → session
sesson      → session

excersize   → exercise
excercise   → exercise

prepartion  → preparation
preperation → preparation

dinr        → dinner
diner       → dinner
```

### How to Use

Just type naturally - corrections happen automatically:

```csv
summary,start,duration,start_tz,categories
Team meetting,2025-12-16 09:00,1h,Europe/Madrid,work
Docter appointmnt,2025-12-16 14:00,30m,Europe/Madrid,health
Morning medicaton,2025-12-16 08:00,5m,Europe/Madrid,medication
```

**Output in ICS:**
```
SUMMARY:💼 Team meeting
SUMMARY:🏥 Doctor appointment
SUMMARY:💊 Morning medication
```

### Customizing the Dictionary

Add your own corrections in `~/.config/tempus/config.yaml`:

```yaml
spell_corrections:
  # All built-in corrections are included automatically
  # Add your own:
  focusblock: focus block
  standup: stand-up
  one2one: one-to-one

  # Language-specific (if you type in multiple languages):
  reunión: reunion
  médico: medico

  # Personal shortcuts:
  tmrw: tomorrow
  appt: appointment
```

**Example with Custom Corrections:**

```csv
summary,start,duration,start_tz,categories
Focusblock,2025-12-16 10:00,2h,Europe/Madrid,work
Standup,2025-12-16 09:00,15m,Europe/Madrid,work
```

With the config above, these become:
```
SUMMARY:💼 Focus block
SUMMARY:💼 Stand-up
```

### Capitalization
Spell checking **preserves capitalization**:
```
Meetting  → Meeting   (capital M preserved)
meetting  → meeting   (lowercase)
MEETTING  → Meeting   (converts to title case)
```

---

## Alarm Profiles

### What It Does
Reusable alarm presets for common scenarios.

### Why It Helps
- **ADHD**: Multiple reminders reduce forgetting
- **Time Blindness**: Countdown warnings help with transitions
- **Consistency**: Same alarm pattern for similar events

### Built-in Profiles

**`adhd-triple`** (Medication, quick tasks):
```
-5m, -1m, 0m
```
Three reminders: 5 minutes before, 1 minute before, and at event time.

**`adhd-countdown`** (Important events, appointments):
```
-1d, -1h, -15m, -5m
```
Four reminders: 1 day before, 1 hour before, 15 minutes, and 5 minutes.

**`medication`** (Same as adhd-triple):
```
-5m, -1m, 0m
```

**`single`** (Standard reminder):
```
-15m
```

**`none`** (No alarms):
```
(empty)
```

### How to Use

**In CSV:**
```csv
summary,start,duration,start_tz,categories,alarms
Morning meds,2025-12-16 08:00,5m,Europe/Madrid,medication,profile:adhd-triple
Doctor appointment,2025-12-16 14:00,1h,Europe/Madrid,health,profile:adhd-countdown
```

**In JSON:**
```json
{
  "summary": "Morning meds",
  "start": "2025-12-16 08:00",
  "duration": "5m",
  "start_tz": "Europe/Madrid",
  "categories": ["medication"],
  "alarms": ["profile:adhd-triple"]
}
```

**In YAML:**
```yaml
- summary: Morning meds
  start: "2025-12-16 08:00"
  duration: 5m
  start_tz: Europe/Madrid
  categories: [medication]
  alarms: [profile:adhd-triple]
```

### List Available Profiles

```bash
tempus config alarm-profiles
```

---

## Smart Defaults

### What It Does
Automatically sets sensible durations when you don't specify one.

### Why It Helps
- **Cognitive Load**: One less thing to think about
- **Time Estimation**: Based on event type and time of day
- **Consistency**: Similar events get similar durations

### How It Works

**Based on Category:**
```
medication     → 5 minutes
breakfast      → 30 minutes
lunch          → 45 minutes
dinner         → 1 hour
exercise       → 1 hour
work (general) → 1 hour
focus block    → 2 hours
```

**Based on Keywords:**
```
"focus"        → 2 hours
"medication"   → 5 minutes
"breakfast"    → 30 minutes
"meeting"      → 1 hour
```

**Based on Time of Day:**
```
Before 10:00   → 30 minutes (breakfast/morning routine)
10:00-12:00    → 1 hour (morning meetings/work)
12:00-14:00    → 45 minutes (lunch)
14:00-17:00    → 1 hour (afternoon meetings)
After 17:00    → 1 hour (evening activities)
```

### Example

```csv
summary,start,start_tz,categories
Morning medication,2025-12-16 08:00,Europe/Madrid,medication
Team meeting,2025-12-16 10:00,Europe/Madrid,work
Lunch,2025-12-16 13:00,Europe/Madrid,personal
Focus block,2025-12-16 15:00,Europe/Madrid,work
```

**Automatic Durations:**
```
Morning medication  → 5 minutes   (08:00-08:05)
Team meeting        → 1 hour      (10:00-11:00)
Lunch               → 45 minutes  (13:00-13:45)
Focus block         → 2 hours     (15:00-17:00)
```

---

## Visual Aids

### What It Does
Automatically adds emoji icons based on event categories.

### Why It Helps
- **Visual Processing**: Icons easier to scan than text
- **Color Coding**: Works across all calendar apps
- **Quick Recognition**: Identify event types at a glance

### Emoji Mapping

```
Medication  → 💊
Health      → 🏥
Work        → 💼
Meeting     → 💼
School      → 📚
Study       → 📚
Exercise    → 🏃
Fitness     → 🏃
Food        → 🍽️
Meal        → 🍽️
Social      → 👥
Family      → 👨‍👩‍👧‍👦
Travel      → ✈️
Transport   → 🚗
Shopping    → 🛒
Appointment → 📅
Personal    → 🌟
Hobby       → 🎨
```

### How to Use

Just add categories - emojis are added automatically:

```csv
summary,start,duration,start_tz,categories
Team meeting,2025-12-16 09:00,1h,Europe/Madrid,work
Doctor visit,2025-12-16 14:00,30m,Europe/Madrid,health
Morning meds,2025-12-16 08:00,5m,Europe/Madrid,medication
Gym session,2025-12-16 18:00,1h,Europe/Madrid,exercise
```

**Output:**
```
💼 Team meeting
🏥 Doctor visit
💊 Morning meds
🏃 Gym session
```

---

## Batch Templates

### What It Does
Pre-filled templates for common scheduling scenarios.

### Why It Helps
- **Executive Function**: Starting is the hardest part - templates help
- **Consistency**: Standard structure reduces decision fatigue
- **Examples**: Learn by example, modify what works

### Available Templates

```bash
tempus batch template list

# Templates:
# - basic          Simple event list
# - adhd-routine   ADHD-optimized daily routine
# - medication     Medication schedule with triple alarms
# - work-meetings  Work meetings with recurrence
# - medical        Medical appointments with prep time
# - travel         Travel itinerary
# - family         Family calendar
```

### Generate a Template

```bash
tempus batch template adhd-routine -o my-routine.csv
```

**Generated File (my-routine.csv):**
```csv
summary,start,duration,start_tz,categories,alarms
Morning medication,2025-12-16 08:00,5m,Europe/Madrid,medication,profile:adhd-triple
Breakfast,2025-12-16 08:15,30m,Europe/Madrid,meal,
Transition: prepare for work,2025-12-16 08:45,15m,Europe/Madrid,transition,-5m
Focus block (morning),2025-12-16 09:00,2h,Europe/Madrid,work,-5m
Break,2025-12-16 11:00,15m,Europe/Madrid,break,
# ... more events
```

### Edit and Create

1. Generate template: `tempus batch template adhd-routine -o routine.csv`
2. Edit file in your preferred editor
3. Create calendar: `tempus batch -i routine.csv -o calendar.ics`

---

## Dry-Run Validation

### What It Does
Preview and validate events before creating the calendar file.

### Why It Helps
- **Anxiety Reduction**: See what you'll get before committing
- **Error Catching**: Find typos and mistakes early
- **Confidence**: Verify everything looks correct

### How to Use

```bash
tempus batch --dry-run -i my-events.csv
```

**Example Output:**
```
✅ ✓ Validation passed: 9 events ready to create

⚠️  Found 2 time conflict(s):
  • 💼 Team meeting (09:00-10:00) overlaps with 🏥 Doctor appointment (09:45-11:00)
  • 💼 Afternoon meeting (14:00-16:00) overlaps with 💼 Late meeting (15:00-16:00)

⚠️  Days with high event load:
  • Tuesday, Dec 16: 9 events (threshold: 8)

Event summary:
  1. Team meeting - 2025/12/16 09:00
  2. Morning medication - 2025/12/16 08:30
  3. Doctor appointment - 2025/12/16 09:45
  4. Lunch break - 2025/12/16 12:00
  5. Afternoon meeting - 2025/12/16 14:00
  6. Late meeting - 2025/12/16 15:00
  7. Therapy session - 2025/12/16 17:00
  8. Dinner preparation - 2025/12/16 19:00
  9. Evening exercise - 2025/12/16 20:00

To create the calendar file, run:
  tempus batch -i my-events.csv -o calendar.ics
```

### What It Checks

- ✅ Event validity (required fields, date formats)
- ⚠️ Time conflicts (overlapping events)
- ⚠️ Overwhelm (too many events per day)
- 📋 Event summary with normalized data
- 🔧 Suggests command to create calendar

---

## Combining Features

All features work together seamlessly:

```bash
# Full neurodivergent-friendly workflow
tempus batch \
  --dry-run \
  --check-conflicts \
  --max-events-per-day 6 \
  -i my-events.csv \
  -o calendar.ics
```

This command:
1. ✅ Validates all events
2. ✅ Fixes date/time formats automatically
3. ✅ Corrects spelling errors
4. ✅ Detects conflicts
5. ✅ Warns about overwhelm
6. ✅ Shows preview before creating

---

## Tips for Success

### ADHD Users
- Use `--dry-run` first to preview and catch mistakes
- Set `--max-events-per-day` to prevent over-scheduling
- Use alarm profiles (`adhd-triple`, `adhd-countdown`) for important events
- Let spell checking handle typos - type fast, don't worry

### Dyslexia Users
- Don't worry about date formats - use what feels natural
- Spell checking handles most common errors
- Visual emoji icons help with quick scanning
- Use templates to reduce typing

### Executive Function Challenges
- Start with templates instead of blank files
- Use dry-run to reduce decision anxiety
- Let smart defaults handle durations
- Conflict detection catches planning mistakes

### Autism Spectrum
- Consistent structure via templates
- Predictable formatting via normalization
- Visual categorization via emojis
- Explicit validation via dry-run

---

## Configuration Examples

### Minimal Config (use all defaults)
```yaml
# ~/.config/tempus/config.yaml
language: en
timezone: Europe/Madrid
```

### ADHD-Optimized Config
```yaml
language: en
timezone: Europe/Madrid

alarm_profiles:
  # Override defaults with your preferences
  my-default: ["-15m", "-5m", "-1m"]
  urgent: ["-1d", "-12h", "-1h", "-15m", "-5m"]

spell_corrections:
  # Add personal shortcuts
  tmrw: tomorrow
  appt: appointment
  focusblock: focus block
```

### Multi-Language Config
```yaml
language: es
timezone: Europe/Madrid

spell_corrections:
  # Spanish corrections
  reunión: reunion
  médico: medico
  cita: appointment
  # English corrections still work
  meetting: meeting
  docter: doctor
```

---

## Getting Help

- **Documentation**: [README.md](../README.md)
- **Examples**: [examples/](../examples/)
- **Issues**: [GitHub Issues](https://github.com/malpanez/tempus/issues)
- **Questions**: Open a discussion or issue

Made with ❤️ for the neurodivergent community.
