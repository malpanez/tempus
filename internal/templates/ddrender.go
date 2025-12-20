package templates

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/i18n"
	"tempus/internal/testutil"
	"tempus/internal/utils"
)

// RenderTmpl is a tiny mustache-like renderer used for filenames and text.
// Supported:
//   - {{key}}
//   - {{slug key}}
//   - {{date key}}
//   - {{#key}} ... {{/key}}  (render block only if key is non-empty)
//
// NOTE: Go's regexp (RE2) doesn't support backreferences like \1, so we capture
// the closing tag as a third group and compare it in code.
func RenderTmpl(tmpl string, values map[string]string, _ *i18n.Translator) (string, error) {
	if tmpl == "" {
		return "", nil
	}
	out := tmpl

	// Conditionals: {{#key}}...{{/key}}  (no backrefs)
	condRe := regexp.MustCompile(`\{\{\#([a-zA-Z0-9_\-\.]+)\}\}([\s\S]*?)\{\{\/([a-zA-Z0-9_\-\.]+)\}\}`)
	out = condRe.ReplaceAllStringFunc(out, func(m string) string {
		sub := condRe.FindStringSubmatch(m)
		if len(sub) < 4 {
			return m
		}
		open, body, close := sub[1], sub[2], sub[3]
		if open != close {
			// mismatched tags, leave unchanged
			return m
		}
		v := strings.TrimSpace(values[open])
		if v == "" {
			return ""
		}
		return simpleReplace(body, values)
	})

	// Simple replacements: {{key}} and {{slug key}}
	out = simpleReplace(out, values)

	return out, nil
}

func simpleReplace(s string, values map[string]string) string {
	// {{date key}}
	dateRe := regexp.MustCompile(`\{\{date\s+([a-zA-Z0-9_\-\.]+)\}\}`)
	s = dateRe.ReplaceAllStringFunc(s, func(m string) string {
		key := dateRe.FindStringSubmatch(m)[1]
		return extractDate(values[key])
	})

	// {{slug key}}
	slugRe := regexp.MustCompile(`\{\{slug\s+([a-zA-Z0-9_\-\.]+)\}\}`)
	s = slugRe.ReplaceAllStringFunc(s, func(m string) string {
		key := slugRe.FindStringSubmatch(m)[1]
		return slugify(values[key])
	})

	// {{key}}
	keyRe := regexp.MustCompile(`\{\{([a-zA-Z0-9_\-\.]+)\}\}`)
	return keyRe.ReplaceAllStringFunc(s, func(m string) string {
		key := keyRe.FindStringSubmatch(m)[1]
		return values[key]
	})
}

func slugify(s string) string {
	return utils.Slugify(s)
}

func extractDate(value string) string {
	v := strings.TrimSpace(strings.ReplaceAll(value, "T", " "))
	if v == "" {
		return ""
	}
	if len(v) >= 10 {
		datePart := v[:10]
		if _, err := time.Parse("2006-01-02", datePart); err == nil {
			return datePart
		}
	}
	if t, _, err := parseDateOrDateTimeInLocation(v, ""); err == nil {
		return t.Format("2006-01-02")
	}
	return slugify(v)
}

// -----------------------------
// Event rendering (data-driven)
// -----------------------------

// renderDDToEvent builds a calendar.Event from a data-driven template + user values.
// Duration logic: if EndField is empty and DurationField has a value, compute end = start + duration.
func (tm *TemplateManager) renderDDToEvent(dd *DataDrivenTemplate, values map[string]string, tr *i18n.Translator) (*calendar.Event, error) {
	if dd == nil {
		return nil, errors.New("nil template")
	}

	out := dd.Output

	// Resolve time zone names (strings). We label DTSTART/DTEND with these TZIDs.
	startTzName := strings.TrimSpace(values[out.StartTZField])
	endTzName := strings.TrimSpace(values[out.EndTZField])
	if startTzName == "" && endTzName != "" {
		startTzName = endTzName
	}
	if endTzName == "" && startTzName != "" {
		endTzName = startTzName
	}

	// Parse start
	startStr := strings.TrimSpace(values[out.StartField])
	if startStr == "" {
		return nil, fmt.Errorf("missing required start field %q", out.StartField)
	}
	startTime, allDayStart, err := parseDateOrDateTimeInLocation(startStr, startTzName)
	if err != nil {
		return nil, fmt.Errorf("invalid start time %q: %w", startStr, err)
	}

	// Decide all-day for this event (template setting wins; if not set, infer from start format)
	allDay := out.AllDay || allDayStart

	// Choose end:
	var endTime time.Time
	endStr := strings.TrimSpace(values[out.EndField])

	if allDay {
		// Expect date only; if not provided, default to next day
		if endStr == "" {
			endTime = startTime.AddDate(0, 0, 1) // all-day DTEND is exclusive
		} else {
			et, isDateOnly, eErr := parseDateOrDateTimeInLocation(endStr, endTzName)
			if eErr != nil {
				return nil, fmt.Errorf("invalid end date %q: %w", endStr, eErr)
			}
			if !isDateOnly {
				// Normalize to midnight if time was provided
				et = time.Date(et.Year(), et.Month(), et.Day(), 0, 0, 0, 0, et.Location())
			}
			endTime = et.AddDate(0, 0, 1) // exclusive
		}
	} else {
		// Timed event:
		switch {
		case endStr != "":
			et, _, eErr := parseDateOrDateTimeInLocation(endStr, endTzName)
			if eErr != nil {
				return nil, fmt.Errorf("invalid end time %q: %w", endStr, eErr)
			}
			endTime = et

		case strings.TrimSpace(values[out.DurationField]) != "":
			durStr := strings.TrimSpace(values[out.DurationField])
			dur, dErr := parseHumanDuration(durStr)
			if dErr != nil {
				return nil, fmt.Errorf("invalid duration %q: %w", durStr, dErr)
			}
			if dur <= 0 {
				return nil, fmt.Errorf("duration must be > 0: %s", durStr)
			}
			endTime = startTime.Add(dur)

		default:
			endTime = startTime.Add(1 * time.Hour)
		}

		if !endTime.After(startTime) {
			return nil, fmt.Errorf("end time must be after start time")
		}
	}

	// Build the event
	summary, _ := RenderTmpl(out.SummaryTmpl, values, tr)
	location, _ := RenderTmpl(out.LocationTmpl, values, tr)
	description, _ := RenderTmpl(out.DescriptionTmpl, values, tr)

	ev := calendar.NewEvent(summary, startTime, endTime)
	ev.AllDay = allDay
	if location != "" {
		ev.Location = location
	}
	if description != "" {
		ev.Description = description
	}
	// Label with TZID (do not shift wall times)
	if startTzName != "" {
		ev.SetStartTimezone(startTzName)
	}
	if endTzName != "" {
		ev.SetEndTimezone(endTzName)
	}

	for _, c := range out.Categories {
		if strings.TrimSpace(c) != "" {
			ev.AddCategory(c)
		}
	}
	if out.Priority > 0 {
		ev.Priority = out.Priority
	}

	// Optional recurrence
	if field := strings.TrimSpace(out.RRuleField); field != "" {
		if val := strings.TrimSpace(values[field]); val != "" {
			ev.RRule = val
		}
	}

	// Optional exclusions
	if field := strings.TrimSpace(out.ExDatesField); field != "" {
		if raw := strings.TrimSpace(values[field]); raw != "" {
			exDates, err := parseDDExDates(raw, startTime, allDay, startTzName)
			if err != nil {
				return nil, err
			}
			if len(exDates) > 0 {
				ev.ExDates = append(ev.ExDates, exDates...)
			}
		}
	}

	// Optional alarms
	if field := strings.TrimSpace(out.AlarmsField); field != "" {
		if raw := strings.TrimSpace(values[field]); raw != "" {
			specs := calendar.SplitAlarmInput(raw)
			if len(specs) > 0 {
				defaultAlarmTZ := strings.TrimSpace(startTzName)
				if defaultAlarmTZ == "" {
					defaultAlarmTZ = strings.TrimSpace(endTzName)
				}
				parsed, err := calendar.ParseAlarmSpecs(specs, defaultAlarmTZ)
				if err != nil {
					return nil, err
				}
				ev.Alarms = append(ev.Alarms, parsed...)
			}
		}
	}

	return ev, nil
}

func parseDDExDates(raw string, start time.Time, allDay bool, tzName string) ([]time.Time, error) {
	items := splitMultiValueList(raw)
	dates := make([]time.Time, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(strings.ReplaceAll(item, "T", " "))
		if normalized == "" {
			continue
		}
		t, isDateOnly, err := parseDateOrDateTimeInLocation(normalized, tzName)
		if err != nil {
			return nil, fmt.Errorf("invalid exdate %q: %w", normalized, err)
		}
		if allDay {
			dates = append(dates, time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()))
			continue
		}
		if isDateOnly {
			align := time.Date(t.Year(), t.Month(), t.Day(), start.Hour(), start.Minute(), start.Second(), 0, start.Location())
			dates = append(dates, align)
			continue
		}
		dates = append(dates, t)
	}
	return dates, nil
}

func splitMultiValueList(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', ';', '\n':
			return true
		default:
			return false
		}
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if s := strings.TrimSpace(f); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// parseDateOrDateTimeInLocation accepts:
//   - "YYYY-MM-DD" (date-only) -> returns date at midnight, isDateOnly=true
//   - "YYYY-MM-DD HH:MM" -> returns time, isDateOnly=false
//
// If tzName provided, parse in that location; else use time.Local.
func parseDateOrDateTimeInLocation(s, tzName string) (t time.Time, isDateOnly bool, err error) {
	s = strings.TrimSpace(s)

	// Date-time first
	if len(s) >= len("2006-01-02 15:04") && strings.Contains(s, " ") {
		layout := "2006-01-02 15:04"
		if tzName != "" {
			if loc, lerr := time.LoadLocation(tzName); lerr == nil {
				t, e := time.ParseInLocation(layout, s, loc)
				return t, false, e
			}
		}
		t2, e2 := time.ParseInLocation(layout, s, time.Local)
		return t2, false, e2
	}

	// Date-only
	layout := "2006-01-02"
	if tzName != "" {
		if loc, lerr := time.LoadLocation(tzName); lerr == nil {
			d, e := time.ParseInLocation(layout, s, loc)
			return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc), true, e
		}
	}
	d, e := time.ParseInLocation(layout, s, time.Local)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location()), true, e
}

// parseHumanDuration parses common human-friendly durations:
//
//	"60"         -> 60 minutes
//	"45m","90m"  -> minutes
//	"1h","2h"    -> hours
//	"1h30m"      -> 1 hour 30 minutes
//	"PT45M","PT1H30M" (ISO-8601 subset) -> supported
//	"45min","30 minutes" -> supported
func parseHumanDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, fmt.Errorf(testutil.ErrMsgEmptyDuration)
	}

	// Plain number => minutes
	if n, err := strconv.Atoi(s); err == nil {
		return time.Duration(n) * time.Minute, nil
	}

	// Normalize words
	s = strings.ReplaceAll(s, "minutes", "m")
	s = strings.ReplaceAll(s, "minute", "m")
	s = strings.ReplaceAll(s, "mins", "m")
	s = strings.ReplaceAll(s, "min", "m")
	s = strings.ReplaceAll(s, " ", "")

	// ISO 8601 'PT' subset
	if strings.HasPrefix(s, "pt") {
		s2 := strings.TrimPrefix(s, "pt")
		return parseHhMmCompact(s2)
	}

	return parseHhMmCompact(s)
}

func parseHhMmCompact(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("invalid duration")
	}
	var hours, mins int
	var err error

	if strings.Contains(s, "h") {
		parts := strings.SplitN(s, "h", 2)
		if parts[0] != "" {
			hours, err = strconv.Atoi(parts[0])
			if err != nil {
				return 0, fmt.Errorf("invalid hours in duration")
			}
		}
		s = parts[1]
	}
	if strings.HasSuffix(s, "m") {
		mstr := strings.TrimSuffix(s, "m")
		if mstr != "" {
			m, err := strconv.Atoi(mstr)
			if err != nil {
				return 0, fmt.Errorf("invalid minutes in duration")
			}
			mins = m
		}
	} else if s != "" {
		// If no trailing 'm' but digits remain, treat as minutes
		m, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid duration tail: %s", s)
		}
		mins = m
	}

	total := time.Duration(hours)*time.Hour + time.Duration(mins)*time.Minute
	if total <= 0 {
		return 0, fmt.Errorf("duration must be > 0")
	}
	return total, nil
}

// ParseDurationString exposes the duration parser for other packages.
func ParseDurationString(s string) (time.Duration, error) {
	return parseHumanDuration(s)
}
