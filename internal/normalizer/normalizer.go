package normalizer

import (
	"fmt"
	"regexp"
	"strings"
	"time"
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
	return fmt.Sprintf("%s %s", now.Format("2006-01-02"), input)
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
		return "", fmt.Errorf("invalid start time: %w", err)
	}

	// Parse duration
	dur, err := ParseHumanDuration(duration)
	if err != nil {
		return "", fmt.Errorf("invalid duration: %w", err)
	}

	// Calculate end time
	endTime := startTime.Add(dur)

	// Format based on whether start has time component
	if strings.Contains(start, ":") {
		return endTime.Format("2006-01-02 15:04"), nil
	}
	return endTime.Format("2006-01-02"), nil
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
		"2006-01-02 15:04",
		"2006-01-02",
		"15:04",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, input, loc); err == nil {
			// If it's time-only format, prepend today
			if format == "15:04" {
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
		return 0, fmt.Errorf("empty duration")
	}

	// Try parsing days (1d, 2d, etc.)
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		if days, err := fmt.Sscanf(daysStr, "%d", new(int)); err == nil && days == 1 {
			var d int
			fmt.Sscanf(daysStr, "%d", &d)
			return time.Duration(d) * 24 * time.Hour, nil
		}
	}

	// Try parsing weeks (1w, 2w, etc.)
	if strings.HasSuffix(s, "w") {
		weeksStr := strings.TrimSuffix(s, "w")
		if weeks, err := fmt.Sscanf(weeksStr, "%d", new(int)); err == nil && weeks == 1 {
			var w int
			fmt.Sscanf(weeksStr, "%d", &w)
			return time.Duration(w) * 7 * 24 * time.Hour, nil
		}
	}

	// Try parsing as Go duration first (45m, 1h30m, etc.)
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Try parsing as HH:MM format
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {
			var hours, minutes int
			_, err := fmt.Sscanf(s, "%d:%d", &hours, &minutes)
			if err == nil {
				return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute, nil
			}
		}
	}

	// Try parsing as plain number (minutes)
	var minutes int
	_, err := fmt.Sscanf(s, "%d", &minutes)
	if err == nil {
		return time.Duration(minutes) * time.Minute, nil
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
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
