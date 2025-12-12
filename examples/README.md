# Tempus Batch Examples

Ready-to-use CSV, JSON, and YAML templates for creating multiple calendar events at once. Perfect for neurodivergent users who want to set up consistent routines, medication schedules, or complex itineraries without starting from scratch.

## Quick Start

1. **Choose** an example file that fits your needs
2. **Edit** the file with your dates, times, and details
3. **Preview** with dry-run (optional): `tempus batch --dry-run -i <file>`
4. **Generate** your calendar:
   ```bash
   tempus batch -i <file> -o my-calendar.ics
   ```
5. **Import** the `.ics` file to Google Calendar, Outlook, or Apple Calendar

---

## Available Examples

### 🧠 ADHD & Neurodivergent

**`adhd-weekly-routine.csv`** - Complete daily routine with executive function support
- Morning/evening medication (3 alarms each: prepare, take, confirm)
- 90-minute focus blocks with halfway reminders
- Explicit transition time between tasks (15min buffers)
- Lunch breaks with reminders (don't skip meals!)
- Wind-down time before bed (no screens)

**Use case**: Create a consistent weekly structure with automatic reminders

```bash
tempus batch -i adhd-weekly-routine.csv -o my-routine.ics
```

---

**`medication-schedule.yaml`** - Multiple daily medications (YAML format)
- Morning, midday, and evening medications
- Each has 3 alarms: -10m, 0m, +5m
- 30-day schedule with daily recurrence
- Comments and dosage notes inline

**Use case**: Never miss a dose with triple alarms

```bash
tempus batch -i medication-schedule.yaml -o medications.ics
```

---

### 💼 Work & Productivity

**`work-meetings.csv`** - Recurring work meetings for 3 months
- Daily standups (Mon-Fri)
- Weekly 1:1s with manager
- Sprint planning, demos, retros
- All-hands meetings
- Includes exceptions for holidays (using `exdate`)

**Use case**: Set up a full quarter of team meetings in one import

```bash
tempus batch -i work-meetings.csv -o team-meetings.ics
```

**Features demonstrated**:
- `FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR` - Weekday recurrence
- `exdate` column to skip specific dates (holidays)
- Multiple alarms with custom descriptions

---

### 🏥 Health & Medical

**`medical-appointments.csv`** - Various healthcare appointments
- Dentist checkups
- Therapy sessions
- General practitioner visits
- Lab tests (with fasting reminders)
- Physical therapy
- Psychiatrist follow-ups
- Eye exams

**Use case**: Keep track of all medical appointments with proper preparation reminders

```bash
tempus batch -i medical-appointments.csv -o health-calendar.ics
```

**Features demonstrated**:
- Multi-day advance reminders (e.g., `-1w` for 1 week before)
- Preparation reminders ("bring medication list", "no food after 10pm")
- Travel time reminders ("leave 2 hours early")

---

### ✈️ Travel & Trips

**`travel-itinerary.json`** - Complete 4-day Dublin trip (JSON format)
- Flights with cross-timezone support (Madrid → Dublin)
- Hotel check-in/check-out
- Tourist activities (Guinness Storehouse, Temple Bar)
- Restaurant reservations
- Packing reminders

**Use case**: Import an entire trip itinerary with all logistics

```bash
tempus batch -i travel-itinerary.json -o dublin-trip.ics
```

**Features demonstrated**:
- Different timezones for start/end (`start_tz` vs `end_tz`)
- Absolute time alarms (`trigger=2025-12-25 06:30`)
- Detailed descriptions with confirmation codes

---

### 👨‍👩‍👧‍👦 Family & Personal

**`family-calendar.csv`** - Family activities for a month
- School drop-offs (weekdays)
- Kids' activities (soccer, piano, swimming)
- Parent-teacher meetings
- Weekly grocery shopping
- Family movie night (Fridays)
- Birthday parties
- School holidays
- Pediatrician appointments
- Date night (parents' time!)

**Use case**: Manage family logistics with shared calendar events

```bash
tempus batch -i family-calendar.csv -o family-december.ics
```

**Features demonstrated**:
- Mix of recurring and one-time events
- All-day events (`all_day: true`)
- Multi-day date ranges (school breaks)
- Complex recurring patterns (soccer Mon+Wed, piano Tuesdays)

---

## File Formats Comparison

### CSV - Best for Spreadsheet Editing

**Pros**:
- Edit in Excel, Google Sheets, LibreOffice Calc
- Easy to copy/paste rows
- Compact format

**Cons**:
- Hard to read in text editor
- Limited support for multi-line descriptions
- Special characters in descriptions require escaping

**When to use**: Large number of similar events (e.g., 30 medication reminders)

---

### JSON - Best for Structured Data

**Pros**:
- Clear structure with nested objects
- Easy to generate programmatically
- Good for complex alarms with descriptions

**Cons**:
- More verbose than CSV
- Requires proper syntax (commas, brackets)

**When to use**: Complex events with many fields (travel itineraries, detailed appointments)

---

### YAML - Best for Readability

**Pros**:
- Most human-readable format
- Supports inline comments (`# This is a comment`)
- No quoting required (usually)
- Lists are cleaner than JSON

**Cons**:
- Indentation-sensitive (spaces matter)
- Less common than CSV/JSON

**When to use**: Events with descriptions, notes, or when you want to add comments

---

## Customization Tips

### Change Dates

**Option 1: Find & Replace**
```bash
# In your text editor
Find: 2025-12-16
Replace: 2026-01-15
```

**Option 2: Spreadsheet Formulas** (CSV only)
```
=TODAY()         # Today's date
=TODAY()+7       # One week from today
=TODAY()+30      # One month from today
```

---

### Change Timezone

**Option 1: Edit the file**
```bash
# Find & replace
Find: Europe/Madrid
Replace: America/New_York
```

**Option 2: Use --default-tz flag**
```bash
# Remove start_tz column from file, then:
tempus batch -i events.csv -o output.ics --default-tz "America/New_York"
```

---

### Adjust Recurrence

**Daily for 30 days**:
```
rrule: FREQ=DAILY;COUNT=30
```

**Weekdays only (Mon-Fri) for 4 weeks**:
```
rrule: FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;COUNT=20
```

**Weekly on specific days for 3 months**:
```
rrule: FREQ=WEEKLY;BYDAY=MO,WE,FR;COUNT=12
```

**Monthly on the 1st for 1 year**:
```
rrule: FREQ=MONTHLY;BYMONTHDAY=1;COUNT=12
```

---

### Skip Specific Dates (Holidays)

**In CSV/YAML**:
```csv
exdate: 2025-12-25 09:00|2026-01-01 09:00
```

**In JSON**:
```json
"exdate": ["2025-12-25 09:00", "2026-01-01 09:00"]
```

This will skip those specific occurrences of a recurring event.

---

## Testing Before Import

### Validate the ICS File
```bash
tempus lint output.ics
```

### Check Event Count
```bash
# Linux/Mac
grep "BEGIN:VEVENT" output.ics | wc -l

# Windows PowerShell
(Select-String "BEGIN:VEVENT" output.ics).Count
```

### Open in Text Editor
ICS files are plain text. Open `output.ics` in any text editor to inspect.

---

## Troubleshooting

### "Error: summary is required"
**Fix**: Every event needs a `summary` field (event title).

### "Error: start is required"
**Fix**: Every event needs a `start` field with date/time.

### "Error: invalid date format"
**Fix**: Use format `YYYY-MM-DD HH:MM` (e.g., `2025-12-10 14:00`).

### "No events found in file"
**Fix**:
- CSV needs a header row with column names
- JSON/YAML must be an array of objects

### Dates are off by one day
**Fix**: Check if you're mixing all-day events with timed events. All-day events use `YYYY-MM-DD` only.

---

## Full Field Reference

### Required Fields
- `summary` - Event title
- `start` - Start date/time

### Optional Fields
- `end` - End date/time (or use `duration`)
- `duration` - Duration (e.g., `45m`, `1h30m`, `90`)
- `start_tz` - Start timezone (e.g., `Europe/Madrid`)
- `end_tz` - End timezone (for cross-timezone events like flights)
- `location` - Where the event takes place
- `description` - Additional notes, instructions
- `all_day` - Set to `true`/`yes`/`1` for all-day events
- `rrule` - Recurrence rule (e.g., `FREQ=DAILY;COUNT=10`)
- `exdate` - Exception dates (skip specific occurrences)
- `categories` - Tags/labels (separate with `|`, `,`, or `;`)
- `alarms` - Reminders (separate multiple with `||` in CSV)

### Alarm Formats
```
-15m                                    # 15 minutes before
-1h                                     # 1 hour before
-1d                                     # 1 day before
trigger=-30m,description=Custom text    # With description
trigger=2025-12-25 08:00                # Absolute time
```

---

## Next Steps

### Import to Your Calendar

**Google Calendar**:
1. Open Google Calendar
2. Click ⚙️ (Settings) → Import & Export
3. Select your `.ics` file
4. Choose destination calendar
5. Click Import

**Outlook**:
1. File → Open & Export → Import/Export
2. Select "Import an iCalendar (.ics) file"
3. Browse to your file
4. Click OK

**Apple Calendar**:
1. File → Import
2. Select your `.ics` file
3. Choose destination calendar
4. Click Import

---

## More Examples & Documentation

- **Neurodivergent features guide**: [docs/NEURODIVERGENT_FEATURES.md](../docs/NEURODIVERGENT_FEATURES.md)
- **Main README**: [README.md](../README.md)
- **Spanish guide**: [docs/es/guia-plantillas.md](../docs/es/guia-plantillas.md)
- **Portuguese guide**: [docs/pt/guia-modelos.md](../docs/pt/guia-modelos.md)

---

## Contributing Examples

Have a useful batch template? Share it!

1. Create your example file (CSV/JSON/YAML)
2. Test it: `tempus batch -i your-file.csv -o test.ics`
3. Add description to this README
4. Submit a pull request

**Ideas for new examples**:
- Academic calendar (classes, exams, assignments)
- Fitness routine (gym, meal prep, recovery)
- Content creation schedule (video uploads, posts, streams)
- Project milestones (sprints, releases, demos)
- Religious observances (prayers, services, holidays)
- Pet care (vet visits, grooming, medication)

---

**Happy scheduling!** 📅✨
