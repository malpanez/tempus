package normalizer

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"tempus/internal/constants"
	"tempus/internal/testutil"
)

var clockOnlyRe = regexp.MustCompile(`^\d{1,2}:\d{2}$`)

// PrependToday takes a time-only string (HH:MM) and prepends today's date in YYYY-MM-DD format.
// If the input already contains a date, it returns the input unchanged.
func PrependToday(input, timezone string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}

	// If it already has a date component, return as-is
	if !clockOnlyRe.MatchString(input) {
		return input
	}

	// Load timezone location
	loc := time.Local
	if tz := strings.TrimSpace(timezone); tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}

	// Prepend today's date
	now := time.Now().In(loc)
	return fmt.Sprintf("%s %s", now.Format(constants.DateFormatISO), input)
}

// NormalizeEndTimeFromDuration calculates end time from start + duration.
// If end is already provided (non-empty), it returns end unchanged.
// Duration should be a string like "45m", "1h30m", "90", etc.
func NormalizeEndTimeFromDuration(start, end, duration, timezone string) (string, error) {
	end = strings.TrimSpace(end)
	if end != "" {
		return end, nil
	}

	duration = strings.TrimSpace(duration)
	if duration == "" {
		return "", nil
	}

	// Parse start time
	startTime, err := ParseDateTime(start, timezone)
	if err != nil {
		return "", fmt.Errorf(testutil.ErrMsgInvalidStartTimeFormat, err)
	}

	// Parse duration
	dur, err := ParseHumanDuration(duration)
	if err != nil {
		return "", fmt.Errorf(testutil.ErrMsgInvalidDurationFormat, err)
	}

	// Calculate end time
	endTime := startTime.Add(dur)

	// Format based on whether start has time component
	if strings.Contains(start, ":") {
		return endTime.Format(constants.DateTimeFormatISO), nil
	}
	return endTime.Format(constants.DateFormatISO), nil
}

// ParseDateTime parses a datetime string in various formats, using the given timezone.
// Supports:
// - YYYY-MM-DD HH:MM
// - YYYY-MM-DD
// - HH:MM (prepends today)
func ParseDateTime(input, timezone string) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty datetime")
	}

	// Load timezone location
	loc := time.Local
	if tz := strings.TrimSpace(timezone); tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}

	// Try parsing different formats
	formats := []string{
		constants.DateTimeFormatISO,
		constants.DateFormatISO,
		constants.TimeFormatHHMM,
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, input, loc); err == nil {
			// If it's time-only format, prepend today
			if format == constants.TimeFormatHHMM {
				now := time.Now().In(loc)
				return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid datetime format: %s", input)
}

// ParseHumanDuration parses human-friendly duration strings.
// Supports:
// - "45m", "1h", "1h30m", "90m"
// - "1d", "2d", "7d" (days)
// - "1w", "2w" (weeks)
// - "90" (interpreted as minutes)
// - "1:30" (HH:MM format)
func ParseHumanDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf(testutil.ErrMsgEmptyDuration)
	}

	// Try days format
	if d, ok := parseDaysFormat(s); ok {
		return d, nil
	}

	// Try weeks format
	if d, ok := parseWeeksFormat(s); ok {
		return d, nil
	}

	// Try Go duration format
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Try HH:MM format
	if d, ok := parseHHMMFormat(s); ok {
		return d, nil
	}

	// Try plain number (minutes)
	if d, ok := parseMinutesFormat(s); ok {
		return d, nil
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}

// parseDaysFormat parses duration strings like "1d", "2d", "7d".
func parseDaysFormat(s string) (time.Duration, bool) {
	if !strings.HasSuffix(s, "d") {
		return 0, false
	}
	daysStr := strings.TrimSuffix(s, "d")
	var d int
	if _, err := fmt.Sscanf(daysStr, "%d", &d); err == nil {
		return time.Duration(d) * 24 * time.Hour, true
	}
	return 0, false
}

// parseWeeksFormat parses duration strings like "1w", "2w".
func parseWeeksFormat(s string) (time.Duration, bool) {
	if !strings.HasSuffix(s, "w") {
		return 0, false
	}
	weeksStr := strings.TrimSuffix(s, "w")
	var w int
	if _, err := fmt.Sscanf(weeksStr, "%d", &w); err == nil {
		return time.Duration(w) * 7 * 24 * time.Hour, true
	}
	return 0, false
}

// parseHHMMFormat parses duration strings like "1:30" (HH:MM).
func parseHHMMFormat(s string) (time.Duration, bool) {
	if !strings.Contains(s, ":") {
		return 0, false
	}
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, false
	}
	var hours, minutes int
	if _, err := fmt.Sscanf(s, "%d:%d", &hours, &minutes); err == nil {
		return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute, true
	}
	return 0, false
}

// parseMinutesFormat parses plain numbers as minutes (e.g., "90" = 90 minutes).
func parseMinutesFormat(s string) (time.Duration, bool) {
	var minutes int
	if _, err := fmt.Sscanf(s, "%d", &minutes); err == nil {
		return time.Duration(minutes) * time.Minute, true
	}
	return 0, false
}

// NormalizeValuesForTemplate normalizes datetime fields in a map of values.
// It prepends today's date to clock-only times and calculates end times from durations.
func NormalizeValuesForTemplate(values map[string]string, startField, endField, durationField, startTzField string) error {
	// Get timezone
	tz := strings.TrimSpace(values[startTzField])

	// Normalize start time (prepend today if clock-only)
	if startValue, ok := values[startField]; ok {
		values[startField] = PrependToday(startValue, tz)
	}

	// Normalize end time from duration if needed
	if endField != "" && durationField != "" {
		if start, ok := values[startField]; ok {
			end, err := NormalizeEndTimeFromDuration(start, values[endField], values[durationField], tz)
			if err != nil {
				return err
			}
			if end != "" {
				values[endField] = end
			}
		}
	}

	return nil
}
