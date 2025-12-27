# Troubleshooting Guide

Visual guide to solving common Tempus issues with clear examples and decision trees.

> **For neurodivergent users**: This guide uses visual diagrams, step-by-step fixes, and clear examples.

## Table of Contents

1. [Quick Diagnosis Flow](#quick-diagnosis-flow)
2. [Common Errors](#common-errors)
3. [Date/Time Issues](#datetime-issues)
4. [Timezone Problems](#timezone-problems)
5. [Batch Import Errors](#batch-import-errors)
6. [Calendar App Import Issues](#calendar-app-import-issues)
7. [Configuration Problems](#configuration-problems)
8. [Performance Issues](#performance-issues)

---

## Quick Diagnosis Flow

```
                    Got an error?
                         │
                         ▼
              ┌──────────────────────┐
              │ What type of error?  │
              └──────┬───────────────┘
                     │
        ┌────────────┼────────────┬──────────────┐
        │            │            │              │
        ▼            ▼            ▼              ▼
   ┌────────┐  ┌────────┐  ┌─────────┐   ┌──────────┐
   │ Date/  │  │ Time-  │  │ Batch   │   │ Import   │
   │ Time   │  │ zone   │  │ File    │   │ to App   │
   └───┬────┘  └───┬────┘  └────┬────┘   └─────┬────┘
       │           │             │              │
       ▼           ▼             ▼              ▼
   Section 3   Section 4     Section 5      Section 6
```

---

## Common Errors

### Error: "required flag(s) not set"

**Problem**: Missing required command-line arguments.

**Visual Indicator**:
```
❌ Error: required flag(s) "start" not set
```

**Solution**: All `create` and `quick` commands need at minimum:
```bash
# Minimum for create:
tempus create "Event Name" --start "2025-12-28 10:00" --duration 1h -o event.ics

# Minimum for quick:
tempus quick "Meeting tomorrow at 3pm for 1 hour"
```

**Decision Tree**:
```
Missing required flag?
       │
       ▼
   ┌───────────────────┐
   │ Using 'create'?   │
   └────┬──────────────┘
        │
   ┌────┴────┐
   │         │
   ▼         ▼
 YES        NO
   │         │
   │         ▼
   │    Using 'batch'? → Need --input and --output
   │         │
   ▼         ▼
Need:    Using 'quick'? → Need event description
 --start
 --duration (or --end)
 -o (output file)
```

---

### Error: "file not found"

**Problem**: Batch input file doesn't exist or path is wrong.

**Visual Indicator**:
```
❌ Error: open events.csv: no such file or directory
```

**Solution Steps**:

**Step 1**: Check if file exists
```bash
# List files in current directory
ls -l *.csv
# or
ls -l *.json
ls -l *.yaml
```

**Step 2**: Use correct path
```bash
# ✅ CORRECT - file in current directory
tempus batch -i events.csv -o calendar.ics

# ✅ CORRECT - file in subdirectory
tempus batch -i ./data/events.csv -o calendar.ics

# ✅ CORRECT - absolute path
tempus batch -i /home/user/events.csv -o calendar.ics

# ❌ WRONG - file in different directory without path
tempus batch -i events.csv -o calendar.ics
# (when events.csv is actually in ~/Documents/)
```

**Visual Path Debug**:
```
Where is your file?
       │
       ▼
  ┌─────────────────────┐
  │ Current directory?  │
  └────┬────────────────┘
       │
  ┌────┴────┐
  │         │
  ▼         ▼
 YES       NO
  │         │
  │         ▼
  │    In subdirectory? → Use ./subdir/file.csv
  │         │
  │         ▼
  │    In parent dir? → Use ../file.csv
  │         │
  │         ▼
  │    Elsewhere? → Use full path /path/to/file.csv
  │
  ▼
Use: tempus batch -i file.csv -o out.ics
```

---

## Date/Time Issues

### Error: "invalid date format"

**Problem**: Date format not recognized.

**Visual Indicator**:
```
❌ Error: invalid date format "Dec 28 2025"
❌ Error: time: cannot parse "28-12-2025" as "2006"
```

**CORRECT Formats**:
```
✅ Dates:
   2025-12-28
   2025-1-5        (auto-pads to 2025-01-05)
   2025/12/28      (auto-converts to 2025-12-28)

✅ Times:
   14:30
   14:30:00
   1430            (auto-formats to 14:30)
   9:00            (auto-pads to 09:00)

✅ Date + Time:
   2025-12-28 14:30
   2025-12-28 14:30:00
   2025/12/28 1430
```

**WRONG Formats**:
```
❌ Dec 28 2025
❌ 28-12-2025
❌ 12/28/2025      (American format - use ISO)
❌ 2:30pm          (use 24-hour format: 14:30)
```

**Fix Decision Tree**:
```
Got date/time error?
       │
       ▼
  Check format:
       │
       ├─ Date part: YYYY-MM-DD ✅ or YYYY/MM/DD ✅
       │           NOT: DD-MM-YYYY ❌ or MM/DD/YYYY ❌
       │
       ├─ Time part: HH:MM ✅ (24-hour)
       │           NOT: HH:MMam/pm ❌
       │
       └─ Together: "2025-12-28 14:30" ✅
                   NOT: "Dec 28, 2025 2:30pm" ❌
```

---

### Error: "end must be after start"

**Problem**: Event ends before or at the same time it starts.

**Visual Indicator**:
```
❌ Error: end time must be after start time
```

**Visual Timeline**:
```
❌ WRONG:
Start: 2025-12-28 14:00
End:   2025-12-28 14:00  ← Same time!

Start: 2025-12-28 14:00
End:   2025-12-28 13:00  ← Earlier!


✅ CORRECT:
Start: 2025-12-28 14:00
End:   2025-12-28 15:00  ← After start

Or use duration:
Start:    2025-12-28 14:00
Duration: 1h
End:      2025-12-28 15:00  ← Auto-calculated
```

**Solution**:
```bash
# ✅ Use duration instead of end time (easier!)
tempus create "Meeting" \
  --start "2025-12-28 14:00" \
  --duration "1h" \
  -o meeting.ics

# ✅ Or ensure end is after start
tempus create "Meeting" \
  --start "2025-12-28 14:00" \
  --end   "2025-12-28 15:00" \
  -o meeting.ics
```

---

### Error: "duration must be positive"

**Problem**: Duration is zero or negative.

**Visual Indicator**:
```
❌ Error: duration must be positive, got 0
```

**Visual Duration Guide**:
```
✅ Valid durations:
   5m         → 5 minutes
   30m        → 30 minutes
   1h         → 1 hour
   1h30m      → 1 hour 30 minutes
   2h         → 2 hours
   90         → 90 minutes (plain number)
   1:30       → 1 hour 30 minutes

❌ Invalid:
   0m         → Zero duration
   0          → Zero duration
   -30m       → Negative (use --alarm for reminders)
```

**Fix**:
```bash
# ❌ WRONG
tempus create "Event" --start "2025-12-28 10:00" --duration 0m

# ✅ CORRECT
tempus create "Event" --start "2025-12-28 10:00" --duration 30m
```

---

## Timezone Problems

### Error: "unknown timezone"

**Problem**: Timezone identifier not recognized.

**Visual Indicator**:
```
❌ Error: unknown time zone US/Eastern
❌ Error: cannot find timezone "PST"
```

**Why it happens**:
```
❌ Common mistakes:
   US/Eastern  → Old format (deprecated)
   PST         → Abbreviation (ambiguous)
   GMT+1       → Offset format (not IANA)

✅ Use IANA timezone names:
   America/New_York   ← Instead of US/Eastern
   America/Los_Angeles← Instead of PST
   Europe/London      ← Instead of GMT
   Europe/Madrid      ← Instead of CET
```

**Solution - Find correct timezone**:
```bash
# Search for your city/country
tempus timezone list --country "United States"
tempus timezone list --country Spain
tempus timezone search "New York"

# Get timezone info
tempus timezone info Europe/Madrid
```

**Common timezone mappings**:
```
❌ Old/Wrong          ✅ Correct IANA name
────────────────────────────────────────────
US/Eastern         → America/New_York
US/Pacific         → America/Los_Angeles
US/Central         → America/Chicago
PST                → America/Los_Angeles
EST                → America/New_York
CET                → Europe/Madrid (or Paris, Berlin)
GMT                → Europe/London
CST (China)        → Asia/Shanghai
IST (India)        → Asia/Kolkata
AEST (Australia)   → Australia/Sydney
```

**Fix Decision Tree**:
```
Got timezone error?
       │
       ▼
  ┌────────────────────┐
  │ Using abbreviation?│  (PST, EST, CET)
  └─────┬──────────────┘
        │
     ┌──┴──┐
     │     │
    YES   NO
     │     │
     │     ▼
     │   Using old format?  (US/Eastern)
     │     │
     │  ┌──┴──┐
     │  │     │
     │ YES   NO
     │  │     │
     │  │     ▼
     │  │   Invalid format? → Check for typos
     │  │
     │  ▼
     │ Use IANA name:
     │ America/New_York
     │ Europe/London
     │ Asia/Tokyo
     │
     ▼
   Use timezone search:
   tempus timezone search "your city"
```

---

### Timezone not applied correctly

**Problem**: Events showing wrong time in calendar app.

**Visual Timeline Example**:
```
You created:
  Start: 2025-12-28 14:00
  Timezone: Europe/Madrid

Calendar shows:
  Start: 2025-12-28 13:00  ← Wrong! Off by 1 hour

Why? You forgot --start-tz flag!
```

**Solution**:
```bash
# ❌ WRONG - No timezone specified
tempus create "Meeting" \
  --start "2025-12-28 14:00" \
  --duration "1h" \
  -o meeting.ics

# ✅ CORRECT - Timezone specified
tempus create "Meeting" \
  --start "2025-12-28 14:00" \
  --start-tz "Europe/Madrid" \
  --duration "1h" \
  -o meeting.ics

# ✅ EVEN BETTER - Set default timezone
tempus config set timezone "Europe/Madrid"
# Now all events use this timezone by default
```

---

## Batch Import Errors

### Error: CSV parsing failed

**Problem**: CSV file has formatting issues.

**Visual Indicator**:
```
❌ Error: record on line 3: wrong number of fields
❌ Error: invalid CSV format
```

**Common CSV mistakes**:

**Problem 1: Missing header**
```csv
❌ WRONG (no header):
Team meeting,2025-12-28 09:00,1h,Europe/Madrid,work
Doctor,2025-12-29 14:00,30m,Europe/Madrid,health

✅ CORRECT (with header):
summary,start,duration,start_tz,categories
Team meeting,2025-12-28 09:00,1h,Europe/Madrid,work
Doctor,2025-12-29 14:00,30m,Europe/Madrid,health
```

**Problem 2: Commas inside values**
```csv
❌ WRONG:
summary,location,start,duration,start_tz
Team meeting,Conference Room A, Building 2,2025-12-28 09:00,1h,Europe/Madrid
                                ^ Extra comma breaks parsing!

✅ CORRECT (quote the field):
summary,location,start,duration,start_tz
Team meeting,"Conference Room A, Building 2",2025-12-28 09:00,1h,Europe/Madrid
             ^                              ^
```

**Problem 3: Missing columns**
```csv
❌ WRONG (missing duration on line 2):
summary,start,duration,start_tz,categories
Team meeting,2025-12-28 09:00,1h,Europe/Madrid,work
Doctor,2025-12-29 14:00,Europe/Madrid,health
       ^ Missing duration field

✅ CORRECT (all columns present):
summary,start,duration,start_tz,categories
Team meeting,2025-12-28 09:00,1h,Europe/Madrid,work
Doctor,2025-12-29 14:00,30m,Europe/Madrid,health
```

**Fix Steps**:
```
1. Open CSV in text editor (NOT Excel - it may auto-format)
2. Check line with error (error message shows line number)
3. Verify:
   ✓ Header row exists
   ✓ Same number of commas in each row
   ✓ Values with commas are quoted
   ✓ No missing columns
4. Save and try again
```

---

### Error: Required field missing

**Problem**: Event missing required data (summary or start).

**Visual Indicator**:
```
❌ Error: event on line 5: required field 'start' is missing
❌ Error: event on line 3: required field 'summary' is missing
```

**Required vs Optional Fields**:
```
┌─────────────────────────────────────────┐
│          FIELD REQUIREMENTS             │
├─────────────────────────────────────────┤
│ REQUIRED (must have):                   │
│  ✓ summary     (event name)            │
│  ✓ start       (start date/time)       │
│                                          │
│ OPTIONAL (can be empty):                │
│  • duration/end (smart default if empty)│
│  • start_tz    (uses default config)   │
│  • end_tz      (uses start_tz)         │
│  • location                             │
│  • description                          │
│  • categories                           │
│  • alarms                               │
│  • rrule                                │
│  • exdate                               │
│  • all_day                              │
└─────────────────────────────────────────┘
```

**Fix**:
```csv
❌ WRONG (missing summary):
summary,start,duration,start_tz,categories
,2025-12-28 09:00,1h,Europe/Madrid,work
^ Empty summary!

✅ CORRECT:
summary,start,duration,start_tz,categories
Team meeting,2025-12-28 09:00,1h,Europe/Madrid,work

❌ WRONG (missing start):
summary,start,duration,start_tz,categories
Team meeting,,1h,Europe/Madrid,work
             ^ Empty start!

✅ CORRECT:
summary,start,duration,start_tz,categories
Team meeting,2025-12-28 09:00,1h,Europe/Madrid,work
```

---

### Dry-run shows warnings

**Problem**: Conflicts or overwhelm detected in dry-run.

**Visual Indicator**:
```
⚠️  Found 2 time conflict(s):
  • 💼 Team meeting (09:00-10:00) overlaps with 🏥 Doctor (09:45-11:00)

⚠️  Days with high event load:
  • Tuesday, Dec 16: 9 events (threshold: 8)
```

**Conflict Resolution Decision Tree**:
```
       Got conflicts?
              │
              ▼
     ┌────────────────┐
     │ Intended?      │
     └───┬────────────┘
         │
    ┌────┴────┐
    │         │
   YES       NO
    │         │
    │         ▼
    │    Fix the times in CSV:
    │      - Move one event earlier/later
    │      - Shorten duration
    │      - Remove one event
    │
    ▼
  Continue with:
  tempus batch -i events.csv -o calendar.ics
```

**Fix Examples**:

**Conflict**:
```csv
❌ CONFLICT DETECTED:
summary,start,duration,start_tz,categories
Team meeting,2025-12-28 09:00,1h,Europe/Madrid,work
Doctor appointment,2025-12-28 09:45,30m,Europe/Madrid,health
                                    ↑ Overlaps with meeting!

✅ OPTION 1 - Move doctor to later:
summary,start,duration,start_tz,categories
Team meeting,2025-12-28 09:00,1h,Europe/Madrid,work
Doctor appointment,2025-12-28 10:30,30m,Europe/Madrid,health

✅ OPTION 2 - Shorten meeting:
summary,start,duration,start_tz,categories
Team meeting,2025-12-28 09:00,30m,Europe/Madrid,work
Doctor appointment,2025-12-28 09:45,30m,Europe/Madrid,health
```

**Overwhelm**:
```csv
⚠️  9 events on Tuesday (threshold: 8)

Options:
1. Accept it (ignore warning)
2. Spread events across multiple days
3. Combine similar events
4. Increase threshold: --max-events-per-day 10
```

---

## Calendar App Import Issues

### Events not showing in calendar

**Problem**: ICS file imported but events don't appear.

**Diagnosis Steps**:

**Step 1: Validate the ICS file**
```bash
tempus lint --file calendar.ics
```

If errors found, fix them and regenerate.

**Step 2: Check import destination**
```
Google Calendar:
  ┌─────────────────────────┐
  │ Did you select correct  │
  │ destination calendar?   │
  └──────────┬──────────────┘
             │
        ┌────┴────┐
        │         │
       YES       NO
        │         │
        │         └→ Re-import, choose correct calendar
        │
        ▼
   Check calendar visibility (enabled?)
```

**Step 3: Refresh the view**
```
Apple Calendar:
  1. Restart the app
  2. View → Refresh All

Google Calendar:
  1. Hard refresh browser (Ctrl+Shift+R)
  2. Check "Show" checkbox for calendar

Outlook:
  1. Close and reopen
  2. Check calendar is visible in sidebar
```

---

### Events showing wrong time

**Problem**: Times shifted by hours when imported.

**Visual Example**:
```
Created:  14:00 Europe/Madrid
Shows in: 13:00 (off by 1 hour)
          OR
          08:00 (off by 6 hours)
```

**Likely causes**:
```
Cause 1: Missing timezone in creation
  ❌ Created without --start-tz
  ✅ Fix: Recreate with --start-tz "Europe/Madrid"

Cause 2: Calendar app using different timezone
  ❌ Your app is set to different timezone
  ✅ Fix: Change app timezone to match event timezone

Cause 3: All-day event interpreted as timed
  ❌ Created all-day without --all-day flag
  ✅ Fix: Recreate with --all-day flag
```

**Fix**:
```bash
# ✅ ALWAYS specify timezone
tempus create "Meeting" \
  --start "2025-12-28 14:00" \
  --start-tz "Europe/Madrid" \
  --duration "1h" \
  -o meeting.ics

# ✅ Or set default
tempus config set timezone "Europe/Madrid"
```

---

### Recurrence not working

**Problem**: Event shows only once instead of recurring.

**Visual Indicator**:
```
Expected: Event every Monday for 4 weeks
Got:      Event only on first Monday
```

**Check RRULE syntax**:
```bash
# Validate your RRULE
tempus rrule
# Use interactive wizard to build correct RRULE

# Common RRULE examples:
FREQ=DAILY;COUNT=10              → Daily for 10 days
FREQ=WEEKLY;BYDAY=MO,WE,FR       → Mon, Wed, Fri forever
FREQ=WEEKLY;BYDAY=MO;COUNT=4     → 4 Mondays
FREQ=MONTHLY;BYMONTHDAY=1        → 1st of every month
FREQ=YEARLY;BYMONTH=12;BYMONTHDAY=25  → Every Dec 25
```

**Common mistakes**:
```
❌ WRONG:
FREQ=WEEKLY;DAYS=MO,TU    → "DAYS" is wrong
COUNT=4;FREQ=WEEKLY       → Wrong order

✅ CORRECT:
FREQ=WEEKLY;BYDAY=MO,TU   → Use "BYDAY"
FREQ=WEEKLY;COUNT=4       → FREQ comes first
```

---

## Configuration Problems

### Config file not found

**Problem**: Can't read config file.

**Visual Indicator**:
```
❌ Warning: config file not found at /home/user/.config/tempus/config.yaml
```

**This is usually OK!** Config file is optional.

**If you want to create one**:
```bash
# Create config directory
mkdir -p ~/.config/tempus/

# Copy example config
# (Assuming you have config.example.yaml in project)
cp config.example.yaml ~/.config/tempus/config.yaml

# Or create minimal config:
cat > ~/.config/tempus/config.yaml <<EOF
timezone: Europe/Madrid
language: en
EOF

# Verify
tempus config list
```

**Config file locations**:
```
Linux/macOS:  ~/.config/tempus/config.yaml
Windows:      %APPDATA%\tempus\config.yaml

Alternative locations (in order):
  1. --config flag: tempus --config /path/to/config.yaml
  2. TEMPUS_CONFIG env var
  3. ~/.config/tempus/config.yaml (default)
```

---

### Settings not persisting

**Problem**: Config changes don't stick between commands.

**Diagnosis**:
```
Did you use 'tempus config set'?
       │
       ▼
  ┌────────┐
  │  YES   │
  └───┬────┘
      │
      ▼
  Check file was created:
  ls -la ~/.config/tempus/config.yaml
      │
  ┌───┴───┐
  │       │
 YES     NO
  │       │
  │       └→ Check permissions:
  │          chmod 755 ~/.config/tempus/
  │
  ▼
Settings should persist
```

**Fix**:
```bash
# Ensure directory exists
mkdir -p ~/.config/tempus/

# Set values
tempus config set timezone "Europe/Madrid"
tempus config set language "en"

# Verify
tempus config list
# Should show:
#   timezone: Europe/Madrid
#   language: en

# Check file
cat ~/.config/tempus/config.yaml
```

---

## Performance Issues

### Batch processing slow

**Problem**: Large CSV files take long to process.

**Expected performance**:
```
Events         Time
───────────────────────
10             < 1 sec
100            < 2 sec
1,000          < 10 sec
10,000         < 60 sec
```

**If slower**:

**Step 1: Use dry-run to validate first**
```bash
# Fast validation without creating file
tempus batch --dry-run -i large.csv
```

**Step 2: Check for complex recurrence**
```
Complex RRULE rules slow down processing.

If possible:
  - Simplify RRULE
  - Reduce COUNT
  - Split into multiple files
```

**Step 3: Profile**
```bash
# Time the operation
time tempus batch -i events.csv -o calendar.ics

# Check file size
wc -l events.csv
# Very large? Consider splitting:
# split -l 1000 events.csv events_part_
```

---

### High memory usage

**Problem**: Tempus using too much RAM.

**Typical usage**:
```
Events         RAM
─────────────────────
100            < 10 MB
1,000          < 50 MB
10,000         < 200 MB
```

**If higher**:
```
Likely cause: Very large description fields

Fix:
  - Trim long descriptions
  - Split into multiple smaller batch files
  - Use external description files (links)
```

---

## Getting More Help

### Check logs for details

**Enable verbose output**:
```bash
# Most errors show helpful details already
# If you need more info, check the error message

# Validate ICS files for issues
tempus lint --file calendar.ics
```

---

### Report a bug

**Before reporting**:
1. Check this troubleshooting guide
2. Verify your input format
3. Try with minimal example

**What to include**:
```
1. Exact command you ran:
   tempus create "Event" --start "2025-12-28 10:00" ...

2. Full error message:
   ❌ Error: ...

3. Your environment:
   - OS: Linux/macOS/Windows
   - Tempus version: tempus version
   - Config: tempus config list

4. Minimal example that reproduces the issue:
   - Sample CSV (if batch)
   - Exact flags (if create)
```

**Where to report**:
- GitHub Issues: https://github.com/malpanez/tempus/issues
- Include "troubleshooting" label

---

## Visual Summary - Error Categories

```
┌─────────────────────────────────────────────────────────────┐
│                    ERROR QUICK REFERENCE                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  DATE/TIME ERRORS:                                           │
│    • Use YYYY-MM-DD format                                  │
│    • Use 24-hour time (HH:MM)                               │
│    • End must be after start                                │
│                                                              │
│  TIMEZONE ERRORS:                                            │
│    • Use IANA names (Europe/Madrid)                         │
│    • NOT abbreviations (CET, PST)                           │
│    • Search: tempus timezone search "city"                  │
│                                                              │
│  CSV/BATCH ERRORS:                                           │
│    • Check header row exists                                │
│    • Quote values with commas                               │
│    • Verify all required fields (summary, start)            │
│                                                              │
│  IMPORT ERRORS:                                              │
│    • Validate with: tempus lint --file calendar.ics        │
│    • Check calendar app timezone                            │
│    • Refresh calendar view                                  │
│                                                              │
│  CONFIG ERRORS:                                              │
│    • Create config: tempus config set timezone ...         │
│    • Check location: ~/.config/tempus/config.yaml          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

Made with ❤️ for the neurodivergent community.

**Still stuck?** Open an issue: https://github.com/malpanez/tempus/issues
