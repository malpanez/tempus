# ⚡ Quick Start Guide - Your First Event in 30 Seconds

> **For ADHD/ASD/Dyslexia users**: This guide uses visual diagrams and step-by-step instructions with clear examples.

## 🎯 Table of Contents

1. [The 30-Second Quick Event](#30-second-quick-event)
2. [The Simple Way (Create Command)](#simple-way-create-command)
3. [The Batch Way (Multiple Events)](#batch-way-multiple-events)
4. [Visual Workflow](#visual-workflow)
5. [Common Patterns](#common-patterns)

---

## 30-Second Quick Event

The **fastest** way to create a calendar event:

```bash
tempus quick "Team meeting tomorrow at 3pm for 1 hour"
```

**What you get:**
```
✅ Created: meeting-2025-12-28.ics

Summary:   Team meeting
Start:     Sat, 28 Dec 2025 15:00 CET
End:       Sat, 28 Dec 2025 16:00 CET
Location:
Timezone:  Europe/Madrid
```

### Visual Process Flow

```
┌─────────────────────────────────────┐
│  Type ONE sentence                  │
│  "meeting tomorrow at 3pm for 1hr"  │
└────────────┬────────────────────────┘
             │
             ▼
     ┌───────────────┐
     │ AI parses it  │
     └───────┬───────┘
             │
             ▼
     ┌──────────────────┐
     │ Shows preview    │
     │ Asks: OK? (y/n)  │
     └───────┬──────────┘
             │
             ▼
     ┌──────────────────┐
     │ Creates .ics file│
     └──────────────────┘
```

---

## Simple Way (Create Command)

For **more control** over your event:

### Step 1: Basic Event

```bash
tempus create "Doctor appointment" \
  --start "2025-12-28 14:00" \
  --duration 30m \
  --output doctor.ics
```

### Step 2: Add Details

```bash
tempus create "Doctor appointment" \
  --start "2025-12-28 14:00" \
  --duration 30m \
  --location "Medical Center, 5th Ave" \
  --category health --category appointment \
  --alarm -24h --alarm -1h --alarm -15m \
  --output doctor.ics
```

### Visual Comparison

```
┌─────────────────────────────────────────────────────────────────┐
│                       BASIC vs DETAILED                          │
├─────────────────────────────┬───────────────────────────────────┤
│         BASIC               │         DETAILED                   │
├─────────────────────────────┼───────────────────────────────────┤
│ ✓ Summary                   │ ✓ Summary                         │
│ ✓ Start time                │ ✓ Start time                      │
│ ✓ Duration                  │ ✓ Duration                        │
│                             │ ✓ Location                        │
│                             │ ✓ Categories (with emoji 🏥)      │
│                             │ ✓ Reminders (3 alarms)            │
└─────────────────────────────┴───────────────────────────────────┘
```

---

## Batch Way (Multiple Events)

For creating **many events at once**:

### Step 1: Create CSV File

Create `events.csv`:

```csv
summary,start,duration,categories,alarms
Morning medication,2025-12-28 08:00,5m,medication,profile:adhd-triple
Team meeting,2025-12-28 10:00,1h,work,-15m
Lunch break,2025-12-28 13:00,45m,meal,
Doctor appointment,2025-12-28 14:00,30m,health,-1h;-15m
```

### Step 2: Preview (Dry Run)

```bash
tempus batch --dry-run -i events.csv
```

**Output:**
```
✅ ✓ Validation passed: 4 events ready to create

Event summary:
  1. 💊 Morning medication - 2025/12/28 08:00
  2. 💼 Team meeting - 2025/12/28 10:00
  3. 🍽️ Lunch break - 2025/12/28 13:00
  4. 🏥 Doctor appointment - 2025/12/28 14:00

To create the calendar file, run:
  tempus batch -i events.csv -o calendar.ics
```

### Step 3: Create Calendar

```bash
tempus batch -i events.csv -o calendar.ics
```

**Output:**
```
✅ Created: calendar.ics (4 events)
```

### Visual Process Flow

```
┌──────────────┐
│ Create CSV   │
│ (Excel/Text) │
└──────┬───────┘
       │
       ▼
┌─────────────────┐
│ Dry Run         │
│ (preview/check) │
└──────┬──────────┘
       │
       ▼
┌──────────────────┐
│ Fix any issues   │
│ (edit CSV)       │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Create calendar  │
│ (batch command)  │
└──────────────────┘
```

---

## Visual Workflow

### Complete Decision Tree

```
                    Start Here
                        │
                        ▼
            ┌───────────────────────┐
            │  How many events?     │
            └───────┬───────────────┘
                    │
         ┌──────────┴──────────┐
         │                     │
         ▼                     ▼
    ┌─────────┐         ┌──────────┐
    │  ONE    │         │ MULTIPLE │
    └────┬────┘         └─────┬────┘
         │                    │
    ┌────┴─────┐             │
    │          │             │
    ▼          ▼             ▼
┌────────┐ ┌────────┐  ┌──────────┐
│ Quick  │ │ Create │  │  Batch   │
│        │ │        │  │          │
│ Fast!  │ │ Control│  │ Powerful │
└────────┘ └────────┘  └──────────┘
    │          │             │
    │          │             │
    └──────────┴─────────────┘
                │
                ▼
         ┌─────────────┐
         │ .ics file   │
         │ created!    │
         └─────────────┘
```

---

## Common Patterns

### Pattern 1: Daily Medication

**Problem**: Need to remember medication 3x per day

**Solution**:
```csv
summary,start,duration,categories,alarms,rrule
Morning meds,2025-12-28 08:00,5m,medication,profile:adhd-triple,FREQ=DAILY
Midday meds,2025-12-28 13:00,5m,medication,profile:adhd-triple,FREQ=DAILY
Evening meds,2025-12-28 20:00,5m,medication,profile:adhd-triple,FREQ=DAILY
```

**Visual Timeline**:
```
08:00 ──┬── -5m alarm
        ├── -1m alarm
        └── 0m alarm (event time)
        💊 Take medication (5 min)

13:00 ──┬── -5m alarm
        ├── -1m alarm
        └── 0m alarm
        💊 Take medication (5 min)

20:00 ──┬── -5m alarm
        ├── -1m alarm
        └── 0m alarm
        💊 Take medication (5 min)
```

### Pattern 2: Work Day with Breaks

**Problem**: Forget to take breaks, get overwhelmed

**Solution**:
```csv
summary,start,duration,categories,alarms
Morning routine,2025-12-28 08:00,30m,personal,
Focus block,2025-12-28 09:00,2h,work,-5m
Break,2025-12-28 11:00,15m,break,-1m
Focus block,2025-12-28 11:15,1h45m,work,-5m
Lunch,2025-12-28 13:00,1h,meal,-5m
Afternoon work,2025-12-28 14:00,2h,work,-5m
Wrap-up,2025-12-28 16:00,30m,work,
```

**Visual Timeline**:
```
08:00 ─────────────  🌟 Morning routine (30m)
08:30
09:00 ═════════════  💼 Focus block (2h) ⚠️ -5m alarm
10:00
11:00 ─────────────  ☕ Break (15m) ⚠️ -1m alarm
11:15 ═════════════  💼 Focus block (1h45m) ⚠️ -5m alarm
12:00
13:00 ─────────────  🍽️ Lunch (1h) ⚠️ -5m alarm
14:00 ═════════════  💼 Afternoon work (2h) ⚠️ -5m alarm
15:00
16:00 ─────────────  💼 Wrap-up (30m)
16:30
```

### Pattern 3: Appointment with Travel Time

**Problem**: Always late because forget about travel time

**Solution**:
```csv
summary,start,duration,categories,alarms,location
Prepare to leave,2025-12-28 13:30,15m,transition,-5m,Home
Travel to doctor,2025-12-28 13:45,30m,travel,-1m,
Doctor appointment,2025-12-28 14:15,45m,health,"-1d,-1h,-15m",Medical Center
```

**Visual Timeline**:
```
13:30 ─────  🚀 Prepare to leave (15m)
             ⚠️ -5m alarm at 13:25

13:45 ─────  🚗 Travel to doctor (30m)
             ⚠️ -1m alarm at 13:44

14:15 ─────  🏥 Doctor appointment (45m)
             ⚠️ -1d alarm (day before)
             ⚠️ -1h alarm at 13:15
             ⚠️ -15m alarm at 14:00

15:00 ────── (Appointment ends)
```

---

## Time Format Examples

### Dates (All Valid)

```
✅ 2025-12-28
✅ 2025/12/28
✅ 2025-1-5       (auto-pads to 2025-01-05)
✅ 2025/1/5
```

### Times (All Valid)

```
✅ 14:30
✅ 14:30:00
✅ 1430           (auto-formats to 14:30)
✅ 9:00           (auto-pads to 09:00)
✅ 900            (auto-formats to 09:00)
```

### Durations (All Valid)

```
✅ 30m            (30 minutes)
✅ 1h             (1 hour)
✅ 1h30m          (1 hour 30 minutes)
✅ 90             (90 minutes - plain number)
✅ 1:30           (1 hour 30 minutes - HH:MM format)
✅ 1d             (1 day = 24 hours)
✅ 1w             (1 week = 7 days)
```

---

## Emoji Legend

Tempus automatically adds emojis based on categories:

```
💊 medication       🏥 health          💼 work/meeting
📚 school/study     🏃 exercise        🍽️ food/meal
👥 social           👨‍👩‍👧‍👦 family         ✈️ travel
🚗 transport        🛒 shopping        📅 appointment
🌟 personal         🎨 hobby
```

---

## Next Steps

### 1. Learn More Commands

- [`tempus create --help`](../README.md#create-command) - Full control over single events
- [`tempus batch --help`](../README.md#batch-command) - Multiple events at once
- [`tempus template list`](../README.md#templates) - Pre-made templates
- [`tempus rrule`](../README.md#rrule) - Interactive recurrence builder

### 2. Configure Your Defaults

```bash
# Set your timezone (like git config)
tempus config set timezone Europe/Madrid

# Set your language
tempus config set language en
```

### 3. Explore Neurodivergent Features

Read the [Neurodivergent Features Guide](./NEURODIVERGENT_FEATURES.md) for:
- Conflict detection
- Overwhelm prevention
- Spell checking
- Alarm profiles
- And more!

---

## Troubleshooting

### ❌ "invalid time format"

**Problem**: Date/time not recognized

**Solution**: Use standard formats
```
✅ GOOD: 2025-12-28 14:30
❌ BAD:  Dec 28 2025 2:30pm
```

### ❌ "end must be after start"

**Problem**: End time is before or equal to start time

**Solution**: Check your times
```
Start:    2025-12-28 14:00
Duration: 30m
End:      2025-12-28 14:30  ✅ (after start)
```

### ❌ "file not found"

**Problem**: Can't find your CSV file

**Solution**: Use full path or check location
```bash
# Relative path
tempus batch -i ./events.csv -o calendar.ics

# Full path
tempus batch -i /home/user/events.csv -o calendar.ics
```

---

## Visual Cheat Sheet

```
┌─────────────────────────────────────────────────────────────┐
│                    TEMPUS CHEAT SHEET                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  QUICK EVENT (fastest):                                      │
│    tempus quick "meeting tomorrow at 3pm"                    │
│                                                              │
│  SINGLE EVENT (controlled):                                  │
│    tempus create "Meeting" \                                 │
│                  --start "2025-12-28 15:00" \               │
│                  --duration 1h                              │
│                                                              │
│  MULTIPLE EVENTS (powerful):                                 │
│    tempus batch -i events.csv -o calendar.ics               │
│                                                              │
│  PREVIEW FIRST (recommended):                                │
│    tempus batch --dry-run -i events.csv                     │
│                                                              │
│  CHECK CONFLICTS:                                            │
│    tempus batch --check-conflicts -i events.csv             │
│                                                              │
│  PREVENT OVERWHELM:                                          │
│    tempus batch --max-events-per-day 6 -i events.csv       │
│                                                              │
│  LIST TEMPLATES:                                             │
│    tempus template list                                      │
│                                                              │
│  BUILD RECURRENCE:                                           │
│    tempus rrule                                             │
│                                                              │
│  TIMEZONE INFO:                                              │
│    tempus timezone madrid                                    │
│                                                              │
│  VALIDATE ICS:                                               │
│    tempus lint calendar.ics                                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

Made with ❤️ for the neurodivergent community.

**Need help?** Open an issue: https://github.com/malpanez/tempus/issues
