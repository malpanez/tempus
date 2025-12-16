package calendar

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	// VTIMEZONE block delimiters
	vtzBegin = "BEGIN:VTIMEZONE\r\n"
	vtzEnd   = "END:VTIMEZONE\r\n"
)

//
// Calendar & Event model
//

// Calendar represents an ICS calendar.
type Calendar struct {
	ProdID   string
	Version  string
	CalScale string
	Events   []Event

	// Optional extras (safe defaults)
	// METHOD:PUBLISH is ideal for imported .ics files (not interactive invites)
	Method string
	// X-WR-CALNAME (many clients show this as calendar name)
	Name string
	// X-WR-TIMEZONE helps calendar imports (e.g., Google Calendar) infer the default TZ
	DefaultTZ string
	// If true, embed minimal VTIMEZONE blocks for a few known TZIDs
	// (helps older Outlook variants). Modern clients do not require this.
	IncludeVTZ bool
}

// Event represents an ICS calendar event
type Event struct {
	UID         string
	Summary     string
	Description string
	Location    string
	StartTime   time.Time
	EndTime     time.Time
	StartTZ     string
	EndTZ       string
	AllDay      bool
	Attendees   []string
	Categories  []string
	Priority    int
	Status      string
	Created     time.Time
	LastMod     time.Time

	// RFC niceties / recurrence / alarms (optional)
	Sequence int         // bump on updates (0 => omit)
	RRule    string      // e.g. FREQ=WEEKLY;BYDAY=MO
	ExDates  []time.Time // cancellations; must match DTSTART type/TZ
	Alarms   []Alarm     // VALARM blocks
}

// Alarm models a VALARM block (DISPLAY is most portable)
type Alarm struct {
	Action            string        // DISPLAY/EMAIL (prefer DISPLAY unless you implement EMAIL properly)
	Summary           string        // optional (useful for EMAIL)
	Description       string        // recommended for DISPLAY (Outlook prefers this)
	TriggerIsRelative bool          // true => use TriggerDuration; false => use TriggerTime (absolute UTC)
	TriggerDuration   time.Duration // negative for "before", positive for "after"
	TriggerTime       time.Time     // absolute UTC trigger if not relative
	Repeat            int           // optional repeats count
	RepeatDuration    time.Duration // optional interval between repeats
}

//
// Constructors
//

// NewCalendar creates a new calendar instance
func NewCalendar() *Calendar {
	return &Calendar{
		ProdID:   "-//Tempus//Tempus Calendar Generator//EN",
		Version:  "2.0",
		CalScale: "GREGORIAN",
		Method:   "PUBLISH", // safe default for exported files
		Events:   make([]Event, 0),
	}
}

// NewEvent creates a new event with required fields
func NewEvent(summary string, start, end time.Time) *Event {
	now := time.Now().UTC()
	return &Event{
		UID:       generateUID(),
		Summary:   summary,
		StartTime: start,
		EndTime:   end,
		Created:   now,
		LastMod:   now,
		Status:    "CONFIRMED",
		Priority:  0,
	}
}

// AddEvent adds an event to the calendar
func (c *Calendar) AddEvent(event *Event) {
	c.Events = append(c.Events, *event)
	if strings.TrimSpace(c.DefaultTZ) == "" {
		if tz := strings.TrimSpace(event.StartTZ); tz != "" && strings.TrimSpace(event.EndTZ) == tz {
			c.DefaultTZ = tz
		}
	}
}

// SetDefaultTimezone sets the default timezone for the calendar metadata.
// This value is emitted as X-WR-TIMEZONE for better compatibility with Google Calendar.
func (c *Calendar) SetDefaultTimezone(tz string) {
	c.DefaultTZ = strings.TrimSpace(tz)
}

//
// Public helpers (kept compatible)
//

// SetTimezone sets the timezone for both start and end times
func (e *Event) SetTimezone(tz string) {
	e.StartTZ = tz
	e.EndTZ = tz
}

// SetStartTimezone sets only the start timezone (useful for flights)
func (e *Event) SetStartTimezone(tz string) {
	e.StartTZ = tz
}

// SetEndTimezone sets only the end timezone (useful for flights)
func (e *Event) SetEndTimezone(tz string) {
	e.EndTZ = tz
}

// AddAttendee adds an attendee email
func (e *Event) AddAttendee(email string) {
	e.Attendees = append(e.Attendees, email)
}

// AddCategory adds a category
func (e *Event) AddCategory(category string) {
	e.Categories = append(e.Categories, category)
}

//
// ToICS (Calendar)
//

func (c *Calendar) ToICS() string {
	var b strings.Builder

	writeLine(&b, "BEGIN:VCALENDAR")
	writeProp(&b, "PRODID", c.ProdID)
	writeProp(&b, "VERSION", c.Version)
	writeProp(&b, "CALSCALE", c.CalScale)
	if strings.TrimSpace(c.Method) != "" {
		writeProp(&b, "METHOD", c.Method)
	}
	if strings.TrimSpace(c.Name) != "" {
		writeProp(&b, "X-WR-CALNAME", escapeText(c.Name))
	}
	if strings.TrimSpace(c.DefaultTZ) != "" {
		writeProp(&b, "X-WR-TIMEZONE", c.DefaultTZ)
	}

	// Optional VTIMEZONE blocks for common TZIDs (only if requested)
	if c.IncludeVTZ {
		for _, tz := range uniqueTZIDs(c.Events) {
			if vtz := knownVTZ(tz); vtz != "" {
				b.WriteString(vtz)
			}
		}
	}

	for _, event := range c.Events {
		b.WriteString(event.ToICS())
	}

	writeLine(&b, "END:VCALENDAR")
	return b.String()
}

//
// ToICS (Event)
//

func (e *Event) ToICS() string {
	var b strings.Builder
	writeLine(&b, "BEGIN:VEVENT")

	e.writeBasicProperties(&b)
	e.writeDateTimeProperties(&b)
	e.writeRecurrenceProperties(&b)
	e.writeOptionalProperties(&b)
	e.writeAlarms(&b)
	e.writeTimestamps(&b)

	writeLine(&b, "END:VEVENT")
	return b.String()
}

func (e *Event) writeBasicProperties(b *strings.Builder) {
	const layoutUTC = "20060102T150405Z"

	writeProp(b, "UID", e.UID)

	// DTSTAMP (UTC); use Created if available, else now
	dtstamp := e.Created
	if dtstamp.IsZero() {
		dtstamp = time.Now().UTC()
	}
	writeProp(b, "DTSTAMP", dtstamp.UTC().Format(layoutUTC))

	if s := strings.TrimSpace(e.Summary); s != "" {
		writeProp(b, "SUMMARY", escapeText(s))
	}

	if d := strings.TrimSpace(e.Description); d != "" {
		writeProp(b, "DESCRIPTION", escapeText(normalizeUserNewlines(d)))
	}

	if l := strings.TrimSpace(e.Location); l != "" {
		writeProp(b, "LOCATION", escapeText(normalizeUserNewlines(l)))
	}
}

func (e *Event) writeDateTimeProperties(b *strings.Builder) {
	const layoutUTC = "20060102T150405Z"
	const layoutLocal = "20060102T150405"

	if e.AllDay {
		writeProp(b, "DTSTART;VALUE=DATE", e.StartTime.Format("20060102"))
		writeProp(b, "DTEND;VALUE=DATE", e.EndTime.Format("20060102"))
		return
	}

	if tz := strings.TrimSpace(e.StartTZ); tz != "" {
		writeProp(b, "DTSTART;TZID="+tz, e.StartTime.Format(layoutLocal))
	} else {
		writeProp(b, "DTSTART", e.StartTime.UTC().Format(layoutUTC))
	}

	if tz := strings.TrimSpace(e.EndTZ); tz != "" {
		writeProp(b, "DTEND;TZID="+tz, e.EndTime.Format(layoutLocal))
	} else {
		writeProp(b, "DTEND", e.EndTime.UTC().Format(layoutUTC))
	}
}

func (e *Event) writeRecurrenceProperties(b *strings.Builder) {
	if strings.TrimSpace(e.RRule) != "" {
		writeProp(b, "RRULE", e.RRule)
	}

	if len(e.ExDates) > 0 {
		e.writeExDates(b)
	}
}

func (e *Event) writeExDates(b *strings.Builder) {
	const layoutUTC = "20060102T150405Z"
	const layoutLocal = "20060102T150405"

	if e.AllDay {
		var parts []string
		for _, x := range e.ExDates {
			parts = append(parts, x.Format("20060102"))
		}
		writeProp(b, "EXDATE;VALUE=DATE", strings.Join(parts, ","))
		return
	}

	if strings.TrimSpace(e.StartTZ) != "" {
		var parts []string
		for _, x := range e.ExDates {
			parts = append(parts, x.Format(layoutLocal))
		}
		writeProp(b, "EXDATE;TZID="+e.StartTZ, strings.Join(parts, ","))
		return
	}

	var parts []string
	for _, x := range e.ExDates {
		parts = append(parts, x.UTC().Format(layoutUTC))
	}
	writeProp(b, "EXDATE", strings.Join(parts, ","))
}

func (e *Event) writeOptionalProperties(b *strings.Builder) {
	if len(e.Attendees) > 0 {
		for _, a := range e.Attendees {
			a = strings.TrimSpace(a)
			if a == "" {
				continue
			}
			writeProp(b, "ATTENDEE", "mailto:"+a)
		}
	}

	if len(e.Categories) > 0 {
		writeProp(b, "CATEGORIES", strings.Join(e.Categories, ","))
	}

	if e.Priority > 0 {
		writeProp(b, "PRIORITY", fmt.Sprintf("%d", e.Priority))
	}

	// STATUS (default to CONFIRMED if empty for consistency)
	if s := strings.TrimSpace(e.Status); s == "" {
		writeProp(b, "STATUS", "CONFIRMED")
	} else {
		writeProp(b, "STATUS", s)
	}
}

func (e *Event) writeAlarms(b *strings.Builder) {
	const layoutUTC = "20060102T150405Z"

	for _, al := range e.Alarms {
		writeLine(b, "BEGIN:VALARM")

		action := strings.ToUpper(strings.TrimSpace(al.Action))
		if action == "" {
			action = "DISPLAY"
		}
		writeProp(b, "ACTION", action)

		e.writeAlarmTrigger(b, al, layoutUTC)
		e.writeAlarmDetails(b, al, action)

		writeLine(b, "END:VALARM")
	}
}

func (e *Event) writeAlarmTrigger(b *strings.Builder, al Alarm, layoutUTC string) {
	if al.TriggerIsRelative {
		writeProp(b, "TRIGGER", formatICSDuration(al.TriggerDuration))
	} else {
		writeProp(b, "TRIGGER;VALUE=DATE-TIME", time.Time(al.TriggerTime.UTC()).Format(layoutUTC))
	}
}

func (e *Event) writeAlarmDetails(b *strings.Builder, al Alarm, action string) {
	if action == "DISPLAY" {
		desc := strings.TrimSpace(al.Description)
		if desc == "" {
			desc = "Reminder"
		}
		writeProp(b, "DESCRIPTION", escapeText(desc))
	}

	if strings.TrimSpace(al.Summary) != "" {
		writeProp(b, "SUMMARY", escapeText(al.Summary))
	}

	if al.Repeat > 0 && al.RepeatDuration > 0 {
		writeProp(b, "REPEAT", fmt.Sprintf("%d", al.Repeat))
		writeProp(b, "DURATION", formatICSDuration(al.RepeatDuration))
	}
}

func (e *Event) writeTimestamps(b *strings.Builder) {
	if e.Sequence > 0 {
		writeProp(b, "SEQUENCE", fmt.Sprintf("%d", e.Sequence))
	}
	writeProp(b, "CREATED", e.Created.UTC().Format("20060102T150405Z"))
	writeProp(b, "LAST-MODIFIED", e.LastMod.UTC().Format("20060102T150405Z"))
}

//
// Helpers (escaping / folding)
//

// escapeText escapes TEXT per RFC 5545
//   - Backslash → \\
//   - Semicolon → \;
//   - Comma     → \,
//   - Newline   → \n
//
// Also strips CR and normalizes CRLF to LF first.
func escapeText(text string) string {
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ReplaceAll(text, `\`, `\\`)
	text = strings.ReplaceAll(text, ";", `\;`)
	text = strings.ReplaceAll(text, ",", `\,`)
	text = strings.ReplaceAll(text, "\n", `\n`)
	return text
}

// normalizeUserNewlines converts user-typed "\n" sequences into real newlines
// before we apply escapeText(), so you don't end up with "\\n" in output.
func normalizeUserNewlines(s string) string {
	if s == "" {
		return s
	}
	return strings.ReplaceAll(s, `\n`, "\n")
}

// writeProp writes "KEY:VALUE" with folding and CRLF.
func writeProp(b *strings.Builder, key, value string) {
	writeLine(b, key+":"+value)
}

// writeLine writes a single logical iCalendar line applying RFC 5545 folding.
// Lines longer than 75 octets are folded by inserting CRLF + space.
func writeLine(b *strings.Builder, line string) {
	folded := foldICalLine(line, 75)
	for i, seg := range folded {
		if i == 0 {
			b.WriteString(seg + "\r\n")
		} else {
			b.WriteString(" " + seg + "\r\n")
		}
	}
}

// foldICalLine splits a string into segments of at most limit octets.
// We approximate octets by counting UTF-8 bytes per rune.
// Returns the segments WITHOUT CRLF or leading spaces; writeLine() adds those.
func foldICalLine(s string, limit int) []string {
	if limit <= 0 || len(s) <= limit {
		return []string{s}
	}
	var segments []string
	var cur strings.Builder
	curBytes := 0

	for _, r := range s {
		rl := utf8.RuneLen(r)
		if rl < 0 {
			rl = 1
		}
		if curBytes+rl > limit {
			segments = append(segments, cur.String())
			cur.Reset()
			curBytes = 0
		}
		cur.WriteRune(r)
		curBytes += rl
	}
	if cur.Len() > 0 {
		segments = append(segments, cur.String())
	}
	return segments
}

// generateUID generates a unique identifier for events
func generateUID() string {
	// Use UUID v4 to ensure uniqueness even when generating events in parallel
	return fmt.Sprintf("%s@tempus", uuid.New().String())
}

// formatICSDuration converts a Go duration to an RFC 5545 DURATION (e.g., -PT15M, PT1H30M).
func formatICSDuration(d time.Duration) string {
	if d == 0 {
		return "PT0S"
	}
	neg := d < 0
	if neg {
		d = -d
	}
	total := int64(d.Seconds())
	days := total / 86400
	rem := total % 86400
	hours := rem / 3600
	rem = rem % 3600
	mins := rem / 60
	secs := rem % 60

	var sb strings.Builder
	if neg {
		sb.WriteByte('-')
	}
	sb.WriteString("P")
	if days > 0 {
		sb.WriteString(fmt.Sprintf("%dD", days))
	}
	if hours > 0 || mins > 0 || secs > 0 {
		sb.WriteByte('T')
		if hours > 0 {
			sb.WriteString(fmt.Sprintf("%dH", hours))
		}
		if mins > 0 {
			sb.WriteString(fmt.Sprintf("%dM", mins))
		}
		if secs > 0 {
			sb.WriteString(fmt.Sprintf("%dS", secs))
		}
	}
	out := sb.String()
	if out == "P" {
		return "PT0S"
	}
	return out
}

//
// Parsing helpers used by other packages (kept for compatibility)
//

// ParseDateTime parses datetime strings in various formats.
// If timeStr is empty, parse date-only. If timezone is empty, interpret in local.
func ParseDateTime(dateStr, timeStr string, timezone string) (time.Time, error) {
	var layout string
	var fullStr string

	if timeStr == "" {
		layout = "2006-01-02"
		fullStr = dateStr
	} else {
		layout = "2006-01-02 15:04"
		fullStr = fmt.Sprintf("%s %s", dateStr, timeStr)
	}

	if timezone == "" {
		return time.ParseInLocation(layout, fullStr, time.Local)
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}
	return time.ParseInLocation(layout, fullStr, loc)
}

// CommonTimezones returns a list of commonly used timezones
func CommonTimezones() map[string]string {
	return map[string]string{
		"madrid":      "Europe/Madrid",
		"dublin":      "Europe/Dublin",
		"london":      "Europe/London",
		"canarias":    "Atlantic/Canary",
		"paris":       "Europe/Paris",
		"berlin":      "Europe/Berlin",
		"rome":        "Europe/Rome",
		"lisbon":      "Europe/Lisbon",
		"new_york":    "America/New_York",
		"los_angeles": "America/Los_Angeles",
		"tokyo":       "Asia/Tokyo",
		"sydney":      "Australia/Sydney",
		"utc":         "UTC",
	}
}

//
// Optional: minimal VTIMEZONE support for a few common TZIDs
// (Only used if Calendar.IncludeVTZ == true)
//

func uniqueTZIDs(events []Event) []string {
	seen := map[string]struct{}{}
	add := func(s string) {
		if strings.TrimSpace(s) == "" {
			return
		}
		seen[s] = struct{}{}
	}
	for _, e := range events {
		if !e.AllDay {
			add(e.StartTZ)
			add(e.EndTZ)
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

func knownVTZ(tzid string) string {
	switch tzid {
	case "Europe/Madrid":
		return vtzBegin +
			"TZID:Europe/Madrid\r\n" +
			"X-LIC-LOCATION:Europe/Madrid\r\n" +
			"BEGIN:DAYLIGHT\r\n" +
			"TZOFFSETFROM:+0100\r\n" +
			"TZOFFSETTO:+0200\r\n" +
			"TZNAME:CEST\r\n" +
			"DTSTART:19700329T020000\r\n" +
			"RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=-1SU\r\n" +
			"END:DAYLIGHT\r\n" +
			"BEGIN:STANDARD\r\n" +
			"TZOFFSETFROM:+0200\r\n" +
			"TZOFFSETTO:+0100\r\n" +
			"TZNAME:CET\r\n" +
			"DTSTART:19701025T030000\r\n" +
			"RRULE:FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU\r\n" +
			"END:STANDARD\r\n" +
			vtzEnd
	case "Europe/Dublin":
		return vtzBegin +
			"TZID:Europe/Dublin\r\n" +
			"X-LIC-LOCATION:Europe/Dublin\r\n" +
			"BEGIN:DAYLIGHT\r\n" +
			"TZOFFSETFROM:+0000\r\n" +
			"TZOFFSETTO:+0100\r\n" +
			"TZNAME:IST\r\n" +
			"DTSTART:19700329T010000\r\n" +
			"RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=-1SU\r\n" +
			"END:DAYLIGHT\r\n" +
			"BEGIN:STANDARD\r\n" +
			"TZOFFSETFROM:+0100\r\n" +
			"TZOFFSETTO:+0000\r\n" +
			"TZNAME:GMT\r\n" +
			"DTSTART:19701025T020000\r\n" +
			"RRULE:FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU\r\n" +
			"END:STANDARD\r\n" +
			vtzEnd
	case "Europe/London":
		return vtzBegin +
			"TZID:Europe/London\r\n" +
			"X-LIC-LOCATION:Europe/London\r\n" +
			"BEGIN:DAYLIGHT\r\n" +
			"TZOFFSETFROM:+0000\r\n" +
			"TZOFFSETTO:+0100\r\n" +
			"TZNAME:BST\r\n" +
			"DTSTART:19700329T010000\r\n" +
			"RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=-1SU\r\n" +
			"END:DAYLIGHT\r\n" +
			"BEGIN:STANDARD\r\n" +
			"TZOFFSETFROM:+0100\r\n" +
			"TZOFFSETTO:+0000\r\n" +
			"TZNAME:GMT\r\n" +
			"DTSTART:19701025T020000\r\n" +
			"RRULE:FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU\r\n" +
			"END:STANDARD\r\n" +
			vtzEnd
	case "America/Sao_Paulo":
		return vtzBegin +
			"TZID:America/Sao_Paulo\r\n" +
			"X-LIC-LOCATION:America/Sao_Paulo\r\n" +
			"BEGIN:STANDARD\r\n" +
			"TZOFFSETFROM:-0300\r\n" +
			"TZOFFSETTO:-0300\r\n" +
			"TZNAME:BRT\r\n" +
			"DTSTART:19700101T000000\r\n" +
			"END:STANDARD\r\n" +
			vtzEnd
	case "Atlantic/Canary":
		return vtzBegin +
			"TZID:Atlantic/Canary\r\n" +
			"X-LIC-LOCATION:Atlantic/Canary\r\n" +
			"BEGIN:DAYLIGHT\r\n" +
			"TZOFFSETFROM:+0000\r\n" +
			"TZOFFSETTO:+0100\r\n" +
			"TZNAME:WEST\r\n" +
			"DTSTART:19700329T010000\r\n" +
			"RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=-1SU\r\n" +
			"END:DAYLIGHT\r\n" +
			"BEGIN:STANDARD\r\n" +
			"TZOFFSETFROM:+0100\r\n" +
			"TZOFFSETTO:+0000\r\n" +
			"TZNAME:WET\r\n" +
			"DTSTART:19701025T020000\r\n" +
			"RRULE:FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU\r\n" +
			"END:STANDARD\r\n" +
			vtzEnd
	default:
		return ""
	}
}
