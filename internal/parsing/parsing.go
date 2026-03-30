package parsing

import (
	"fmt"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/cli"
	"tempus/internal/testutil"
)

type ParseOptions struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	Duration  string
	Timezone  string
	EndTZ     string
	AllDay    bool
	Summary   string
}

type ParseResult struct {
	Start time.Time
	End   time.Time
}

func Parse(opts ParseOptions) (ParseResult, error) {
	opts.StartDate = NormalizeDateTimeInput(opts.StartDate)
	opts.StartTime = NormalizeDateTimeInput(opts.StartTime)
	opts.EndDate = NormalizeDateTimeInput(opts.EndDate)
	opts.EndTime = NormalizeDateTimeInput(opts.EndTime)

	opts.StartDate = normalizeTimeInput(opts.StartDate, opts.Timezone, opts.EndTZ)
	opts.EndDate = normalizeTimeInput(opts.EndDate, opts.Timezone, opts.EndTZ)

	if opts.AllDay {
		start, end, err := parseAllDay(opts)
		if err != nil {
			return ParseResult{}, err
		}
		return ParseResult{Start: start, End: end}, nil
	}

	start, end, err := parseTimed(opts)
	if err != nil {
		return ParseResult{}, err
	}
	return ParseResult{Start: start, End: end}, nil
}

func normalizeTimeInput(timeStr, startTZ, endTZ string) string {
	if timeStr != "" && cli.LooksLikeClock(timeStr) {
		return cli.PrependToday(timeStr, cli.FirstNonEmpty(startTZ, endTZ, ""))
	}
	return timeStr
}

func parseAllDay(opts ParseOptions) (startTime, endTime time.Time, err error) {
	startStr := strings.TrimSpace(opts.StartDate)
	if startStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start date is required")
	}
	startDateStr := cli.ExtractDate(startStr)
	startTime, err = time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date %q: %w", startStr, err)
	}

	endStr := strings.TrimSpace(opts.EndDate)
	if endStr == "" {
		endStr = strings.TrimSpace(opts.EndTime)
	}
	if endStr == "" {
		endTime = startTime.AddDate(0, 0, 1)
	} else {
		endDateStr := cli.ExtractDate(endStr)
		endDate, parseErr := time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end date %q: %w", endStr, parseErr)
		}
		if endDate.Before(startTime) {
			return time.Time{}, time.Time{}, fmt.Errorf(testutil.ErrMsgEndDateAfterStart)
		}
		endTime = endDate.AddDate(0, 0, 1)
	}

	if !endTime.After(startTime) {
		return time.Time{}, time.Time{}, fmt.Errorf(testutil.ErrMsgEndDateAfterStart)
	}

	return startTime, endTime, nil
}

func parseTimed(opts ParseOptions) (startTime, endTime time.Time, err error) {
	startStr := buildDateTimeStr(opts.StartDate, opts.StartTime)
	if startStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start date/time is required")
	}

	if cli.LooksLikeClock(startStr) {
		startStr = cli.PrependToday(startStr, opts.Timezone)
	}

	startTime, err = time.Parse("2006-01-02 15:04", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf(testutil.ErrMsgInvalidStartTimeFormat, err)
	}

	endStr := buildDateTimeStr(opts.EndDate, opts.EndTime)
	durStr := strings.TrimSpace(opts.Duration)

	switch {
	case endStr != "":
		endTime, err = parseEndTime(startTime, endStr, opts.EndTZ)
	case durStr != "":
		endTime, err = parseDurationEnd(startTime, durStr)
	default:
		endTime = startTime.Add(1 * time.Hour)
	}

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if !endTime.After(startTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("end time must be after start time")
	}

	return startTime, endTime, nil
}

func buildDateTimeStr(datePart, timePart string) string {
	d := strings.TrimSpace(datePart)
	t := strings.TrimSpace(timePart)
	if d == "" && t == "" {
		return ""
	}
	if t == "" {
		return d
	}
	if d == "" {
		return t
	}
	if strings.Contains(d, " ") {
		return d
	}
	return d + " " + t
}

func parseEndTime(startTime time.Time, endStr, endTZ string) (time.Time, error) {
	if cli.LooksLikeClock(endStr) {
		endStr = cli.PrependToday(endStr, endTZ)
	}

	if d, derr := calendar.ParseHumanDuration(endStr); derr == nil {
		if d <= 0 {
			return time.Time{}, fmt.Errorf(testutil.ErrMsgDurationGreaterThanZero)
		}
		return startTime.Add(d), nil
	}

	endTime, err := time.Parse("2006-01-02 15:04", endStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end time %q: %w", endStr, err)
	}
	return endTime, nil
}

func parseDurationEnd(startTime time.Time, durStr string) (time.Time, error) {
	d, err := calendar.ParseHumanDuration(durStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration %q: %v", durStr, err)
	}
	if d <= 0 {
		return time.Time{}, fmt.Errorf(testutil.ErrMsgDurationGreaterThanZero)
	}
	return startTime.Add(d), nil
}

func NormalizeDateTimeInput(input string) string {
	if input == "" {
		return input
	}

	input = strings.TrimSpace(input)

	input = strings.ReplaceAll(input, "/", "-")

	parts := strings.Split(input, " ")
	if len(parts) >= 1 {
		datePart := parts[0]
		dateComponents := strings.Split(datePart, "-")
		if len(dateComponents) == 3 {
			for i, comp := range dateComponents {
				if len(comp) == 1 && i > 0 {
					dateComponents[i] = "0" + comp
				}
			}
			parts[0] = strings.Join(dateComponents, "-")
		}
	}

	if len(parts) >= 2 {
		timePart := parts[1]
		if len(timePart) == 4 && !strings.Contains(timePart, ":") {
			parts[1] = timePart[:2] + ":" + timePart[2:]
		}
		timeComponents := strings.Split(parts[1], ":")
		if len(timeComponents) == 2 && len(timeComponents[0]) == 1 {
			parts[1] = "0" + parts[1]
		}
	}

	return strings.Join(parts, " ")
}

func NormalizeClockOnlyDateTimes(values map[string]string, startKey, endKey, tzKey string) {
	if strings.TrimSpace(startKey) == "" {
		return
	}

	tz := strings.TrimSpace(values[tzKey])

	if st := strings.TrimSpace(values[startKey]); st != "" && cli.LooksLikeClock(st) {
		values[startKey] = cli.PrependToday(st, tz)
	}

	if strings.TrimSpace(endKey) == "" {
		return
	}

	if et := strings.TrimSpace(values[endKey]); et != "" && cli.LooksLikeClock(et) {
		if _, err := calendar.ParseHumanDuration(et); err != nil {
			values[endKey] = cli.PrependToday(et, tz)
		}
	}
}

func NormalizeEndTimeFromDuration(values map[string]string, startKey, endKey, durationKey, tzKey, defaultDuration string) {
	start := strings.TrimSpace(values[startKey])
	if start == "" {
		return
	}

	end := GetValueIfKeyNotEmpty(values, endKey)
	dur := GetValueIfKeyNotEmpty(values, durationKey)
	tz := strings.TrimSpace(values[tzKey])

	if TrySetEndFromDurationInEnd(values, start, tz, end, endKey, durationKey) {
		return
	}

	if end == "" {
		if dur == "" {
			dur = strings.TrimSpace(defaultDuration)
		}
		SetEndFromDuration(values, start, tz, dur, endKey, durationKey)
	}
}

func GetValueIfKeyNotEmpty(values map[string]string, key string) string {
	if strings.TrimSpace(key) != "" {
		return strings.TrimSpace(values[key])
	}
	return ""
}

func TrySetEndFromDurationInEnd(values map[string]string, start, tz, end, endKey, durationKey string) bool {
	if end == "" {
		return false
	}
	d, err := calendar.ParseHumanDuration(end)
	if err != nil || d <= 0 {
		return false
	}
	endDT := AddDurationToStart(start, tz, d)
	if endDT == "" {
		return false
	}
	SetEndAndDuration(values, endKey, durationKey, endDT, d)
	return true
}

func SetEndFromDuration(values map[string]string, start, tz, dur, endKey, durationKey string) {
	d, err := calendar.ParseHumanDuration(dur)
	if err != nil || d <= 0 {
		return
	}
	endDT := AddDurationToStart(start, tz, d)
	if endDT == "" {
		return
	}
	SetEndAndDuration(values, endKey, durationKey, endDT, d)
}

func SetEndAndDuration(values map[string]string, endKey, durationKey, endDT string, d time.Duration) {
	if strings.TrimSpace(endKey) != "" {
		values[endKey] = endDT
	}
	if strings.TrimSpace(durationKey) != "" {
		values[durationKey] = cli.FmtDurationHuman(d)
	}
}

func AddDurationToStart(start, tz string, d time.Duration) string {
	datePart, timePart := cli.SplitDateTime(start)
	st, err := ParseDateTimeWithTZ(datePart, timePart, tz)
	if err != nil {
		if t2, e2 := time.Parse("2006-01-02 15:04", start); e2 == nil {
			st = t2
		} else {
			return ""
		}
	}
	end := st.Add(d)
	return end.Format("2006-01-02 15:04")
}

func ParseDateTimeWithTZ(dateStr, timeStr, tz string) (time.Time, error) {
	layout := "2006-01-02"
	val := strings.TrimSpace(dateStr)
	if strings.TrimSpace(timeStr) != "" {
		layout = "2006-01-02 15:04"
		val = fmt.Sprintf("%s %s", strings.TrimSpace(dateStr), strings.TrimSpace(timeStr))
	}
	if strings.TrimSpace(tz) == "" {
		return time.ParseInLocation(layout, val, time.Local)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	return time.ParseInLocation(layout, val, loc)
}

func ParseExDateValues(values []string, tz string, allDay bool) ([]time.Time, error) {
	out := make([]time.Time, 0, len(values))
	for _, raw := range values {
		normalized := strings.TrimSpace(raw)
		if normalized == "" {
			continue
		}
		normalized = strings.ReplaceAll(normalized, "T", " ")

		datePart, timePart := cli.SplitDateTime(normalized)
		isDateOnly := strings.TrimSpace(timePart) == ""

		if allDay || isDateOnly {
			t, err := ParseDateTimeWithTZ(datePart, "", tz)
			if err != nil {
				if fallback, err2 := time.Parse("2006-01-02", datePart); err2 == nil {
					t = fallback
				} else {
					return nil, fmt.Errorf("invalid exdate %q: %w", raw, err)
				}
			}
			out = append(out, t)
			continue
		}

		t, err := ParseDateTimeWithTZ(datePart, timePart, tz)
		if err != nil {
			if fallback, err2 := time.Parse("2006-01-02 15:04", normalized); err2 == nil {
				t = fallback
			} else {
				return nil, fmt.Errorf("invalid exdate %q: %w", raw, err)
			}
		}
		out = append(out, t)
	}
	return out, nil
}
