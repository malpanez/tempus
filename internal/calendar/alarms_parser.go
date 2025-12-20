package calendar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tempus/internal/testutil"
)

const (
	actionDisplay   = "DISPLAY"
	defaultDescText = "Reminder"
)

var (
	alarmHHMMRe    = regexp.MustCompile(`^\s*(\d{1,2})\s*:\s*([0-5]?\d)\s*$`)
	alarmHMRe      = regexp.MustCompile(`^\s*(?:(\d+)\s*h\s*)?(?:(\d+)\s*m\s*)?$`)
	alarmMinutesRe = regexp.MustCompile(`^\s*\d+\s*$`)
	icsDurationRe  = regexp.MustCompile(`(?i)^P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)
)

// ParseHumanDuration converts human-friendly strings (e.g., "1h30m", "90", "1:30", "1d", "1w") into time.Duration.
func ParseHumanDuration(s string) (time.Duration, error) {
	x := strings.ToLower(strings.TrimSpace(s))
	if x == "" {
		return 0, fmt.Errorf(testutil.ErrMsgEmptyDuration)
	}

	if dur, ok := tryParseDaysOrWeeks(x); ok {
		return dur, nil
	}

	if dur, ok := tryParseTimeFormat(x); ok {
		return dur, nil
	}

	if dur, ok := tryParseMinutes(x); ok {
		return dur, nil
	}

	return 0, fmt.Errorf("unrecognized duration format: %q", s)
}

func tryParseDaysOrWeeks(x string) (time.Duration, bool) {
	// Try parsing days (1d, 2d, etc.)
	if strings.HasSuffix(x, "d") && len(x) > 1 {
		daysStr := strings.TrimSuffix(x, "d")
		days := atoiSafe(daysStr)
		if days > 0 {
			return time.Duration(days) * 24 * time.Hour, true
		}
	}

	// Try parsing weeks (1w, 2w, etc.)
	if strings.HasSuffix(x, "w") && len(x) > 1 {
		weeksStr := strings.TrimSuffix(x, "w")
		weeks := atoiSafe(weeksStr)
		if weeks > 0 {
			return time.Duration(weeks) * 7 * 24 * time.Hour, true
		}
	}

	return 0, false
}

func tryParseTimeFormat(x string) (time.Duration, bool) {
	// Try HH:MM format
	if m := alarmHHMMRe.FindStringSubmatch(x); m != nil {
		hh := atoiSafe(m[1])
		mm := atoiSafe(m[2])
		return time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute, true
	}

	// Try HhMm format (e.g., "1h30m")
	if m := alarmHMRe.FindStringSubmatch(x); m != nil {
		hh := atoiSafe(m[1])
		mm := atoiSafe(m[2])
		if hh == 0 && mm == 0 {
			return 0, false
		}
		return time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute, true
	}

	return 0, false
}

func tryParseMinutes(x string) (time.Duration, bool) {
	if alarmMinutesRe.MatchString(x) {
		mins := atoiSafe(x)
		if mins <= 0 {
			return 0, false
		}
		return time.Duration(mins) * time.Minute, true
	}
	return 0, false
}

// SplitAlarmInput tokenizes alarm strings separated by newlines, double pipes, or commas/semicolons.
func SplitAlarmInput(raw string) []string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}

	normalized := normalizeNewlines(s)
	lines := strings.Split(normalized, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		processAlarmLine(line, &out)
	}
	return out
}

func normalizeNewlines(s string) string {
	normalized := strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(normalized, "\r", "\n")
}

func processAlarmLine(line string, out *[]string) {
	if strings.Contains(line, "||") {
		for _, part := range strings.Split(line, "||") {
			*out = append(*out, SplitAlarmInput(part)...)
		}
		return
	}

	if strings.Contains(line, "=") {
		*out = append(*out, line)
		return
	}

	splitAndAppendParts(line, out)
}

func splitAndAppendParts(line string, out *[]string) {
	parts := strings.FieldsFunc(line, func(r rune) bool {
		return r == ',' || r == ';' || r == '|'
	})
	if len(parts) == 0 {
		return
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			*out = append(*out, part)
		}
	}
}

// ParseAlarmsFromString parses a single raw string into alarms.
func ParseAlarmsFromString(raw, defaultTZ string) ([]Alarm, error) {
	specs := SplitAlarmInput(raw)
	if len(specs) == 0 {
		return nil, nil
	}
	return ParseAlarmSpecs(specs, defaultTZ)
}

// ParseAlarmSpecs converts alarm specs into calendar.Alarm definitions.
func ParseAlarmSpecs(values []string, defaultTZ string) ([]Alarm, error) {
	out := make([]Alarm, 0, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		al, err := parseAlarmSpec(raw, defaultTZ)
		if err != nil {
			return nil, err
		}
		out = append(out, al)
	}
	return out, nil
}

func parseAlarmSpec(raw, defaultTZ string) (Alarm, error) {
	if strings.Contains(raw, "=") {
		return parseKeyValueAlarmSpec(raw, defaultTZ)
	}
	return parseSimpleAlarmSpec(raw, defaultTZ)
}

func parseSimpleAlarmSpec(spec, defaultTZ string) (Alarm, error) {
	trigger := strings.TrimSpace(spec)
	if trigger == "" {
		return Alarm{}, fmt.Errorf("alarm trigger cannot be empty")
	}

	if dur, err := parseRelativeAlarmDuration(trigger, -1); err == nil {
		return Alarm{
			Action:            actionDisplay,
			Description:       defaultDescText,
			TriggerIsRelative: true,
			TriggerDuration:   dur,
		}, nil
	}

	ts, err := parseAlarmAbsolute(trigger, defaultTZ)
	if err != nil {
		return Alarm{}, fmt.Errorf("invalid alarm %q: %v", spec, err)
	}
	return Alarm{
		Action:            actionDisplay,
		Description:       defaultDescText,
		TriggerIsRelative: false,
		TriggerTime:       ts.UTC(),
	}, nil
}

func parseKeyValueAlarmSpec(spec, defaultTZ string) (Alarm, error) {
	params, err := parseAlarmKeyValueParams(spec)
	if err != nil {
		return Alarm{}, err
	}

	trigger := strings.TrimSpace(firstNonEmpty(params["trigger"], params["offset"]))
	if trigger == "" {
		return Alarm{}, fmt.Errorf("alarm %q is missing trigger= value", spec)
	}

	al := createAlarmFromParams(params)
	triggerMode := determineAlarmTriggerMode(params)

	repeat, repeatDur, err := parseAlarmRepeatParams(params, spec)
	if err != nil {
		return Alarm{}, err
	}

	if err := setAlarmTrigger(&al, trigger, triggerMode, defaultTZ, spec); err != nil {
		return Alarm{}, err
	}

	if repeat > 0 || repeatDur > 0 {
		if repeat <= 0 || repeatDur <= 0 {
			return Alarm{}, fmt.Errorf("repeat count and repeat duration must both be positive in alarm %q", spec)
		}
		al.Repeat = repeat
		al.RepeatDuration = repeatDur
	}

	if al.TriggerIsRelative && al.TriggerDuration == 0 {
		return Alarm{}, fmt.Errorf("alarm %q has zero relative duration", spec)
	}
	return al, nil
}

func parseAlarmKeyValueParams(spec string) (map[string]string, error) {
	parts := strings.FieldsFunc(spec, func(r rune) bool {
		return r == ',' || r == ';'
	})
	params := make(map[string]string, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid alarm segment %q", part)
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])
		if key != "" {
			params[key] = val
		}
	}
	return params, nil
}

func createAlarmFromParams(params map[string]string) Alarm {
	action := strings.ToUpper(strings.TrimSpace(firstNonEmpty(params["action"], "")))
	if action == "" {
		action = actionDisplay
	}

	description := strings.TrimSpace(firstNonEmpty(params["description"], params["message"], params["text"]))
	summary := strings.TrimSpace(firstNonEmpty(params["summary"], params["title"]))

	al := Alarm{
		Action:      action,
		Summary:     summary,
		Description: description,
	}
	if strings.TrimSpace(al.Description) == "" && al.Action == actionDisplay {
		al.Description = defaultDescText
	}
	return al
}

type alarmTriggerMode struct {
	forceRelative    bool
	forceAbsolute    bool
	defaultDirection int
}

func determineAlarmTriggerMode(params map[string]string) alarmTriggerMode {
	mode := alarmTriggerMode{defaultDirection: -1}

	dirHint := strings.ToLower(strings.TrimSpace(firstNonEmpty(params["direction"], params["when"])))
	switch dirHint {
	case "after", "post", "later", "follow", "following", "plus":
		mode.defaultDirection = 1
	case "before", "prior", "pre", "minus":
		mode.defaultDirection = -1
	}

	kind := strings.ToLower(strings.TrimSpace(params["kind"]))
	switch kind {
	case "relative":
		mode.forceRelative = true
	case "before":
		mode.forceRelative = true
		mode.defaultDirection = -1
	case "after":
		mode.forceRelative = true
		mode.defaultDirection = 1
	case "absolute", "at", "on":
		mode.forceAbsolute = true
	}

	relativeHint := strings.TrimSpace(firstNonEmpty(params["relative"], params["is_relative"]))
	if relativeHint != "" {
		if parseBoolish(relativeHint) {
			mode.forceRelative = true
			mode.forceAbsolute = false
		} else {
			mode.forceAbsolute = true
			mode.forceRelative = false
		}
	}

	return mode
}

func parseAlarmRepeatParams(params map[string]string, spec string) (int, time.Duration, error) {
	repeatStr := strings.TrimSpace(firstNonEmpty(params["repeat"], params["repetitions"]))
	repeat := 0
	if repeatStr != "" {
		val, err := strconv.Atoi(repeatStr)
		if err != nil || val <= 0 {
			return 0, 0, fmt.Errorf("invalid repeat count %q in alarm %q", repeatStr, spec)
		}
		repeat = val
	}

	repeatDurStr := strings.TrimSpace(firstNonEmpty(params["repeat_duration"], params["repeat_interval"]))
	var repeatDur time.Duration
	if repeatDurStr != "" {
		dur, err := parseAlarmDurationValue(repeatDurStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid repeat duration %q in alarm %q: %v", repeatDurStr, spec, err)
		}
		if dur <= 0 {
			return 0, 0, fmt.Errorf("repeat duration must be positive in alarm %q", spec)
		}
		repeatDur = dur
	}

	return repeat, repeatDur, nil
}

func setAlarmTrigger(al *Alarm, trigger string, mode alarmTriggerMode, defaultTZ, spec string) error {
	var relDur time.Duration
	var relErr error

	if !mode.forceAbsolute {
		relDur, relErr = parseRelativeAlarmDuration(trigger, mode.defaultDirection)
		if relErr == nil {
			al.TriggerIsRelative = true
			al.TriggerDuration = relDur
			return nil
		}
	}

	if mode.forceRelative {
		if relErr != nil {
			return fmt.Errorf("invalid relative trigger %q in alarm %q: %v", trigger, spec, relErr)
		}
		return nil
	}

	ts, err := parseAlarmAbsolute(trigger, defaultTZ)
	if err != nil {
		if relErr != nil && !mode.forceAbsolute {
			return fmt.Errorf("invalid alarm %q: %v; also failed to parse relative offset (%v)", spec, err, relErr)
		}
		return fmt.Errorf("invalid alarm %q: %v", spec, err)
	}
	al.TriggerIsRelative = false
	al.TriggerTime = ts.UTC()
	return nil
}

func parseAlarmAbsolute(raw, defaultTZ string) (time.Time, error) {
	val := strings.TrimSpace(raw)
	if val == "" {
		return time.Time{}, fmt.Errorf("empty absolute trigger")
	}

	if t, ok := tryParseRFC3339(val); ok {
		return t, nil
	}

	loc := loadTimezoneLocation(defaultTZ)
	return tryParseCommonLayouts(val, loc, raw)
}

func tryParseRFC3339(val string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t, true
	}

	// Try RFC3339 with space instead of T
	if strings.Count(val, " ") == 1 && strings.HasSuffix(strings.ToUpper(val), "Z") {
		if t, err := time.Parse(time.RFC3339, strings.Replace(val, " ", "T", 1)); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func loadTimezoneLocation(defaultTZ string) *time.Location {
	if tz := strings.TrimSpace(defaultTZ); tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			return l
		}
	}
	return nil
}

func tryParseCommonLayouts(val string, loc *time.Location, raw string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}

	for _, layout := range layouts {
		if t, ok := tryParseWithLayout(val, layout, loc); ok {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized absolute date/time %q", raw)
}

func tryParseWithLayout(val, layout string, loc *time.Location) (time.Time, bool) {
	if loc != nil {
		if t, err := time.ParseInLocation(layout, val, loc); err == nil {
			return t.In(time.UTC), true
		}
	}
	if t, err := time.Parse(layout, val); err == nil {
		return t.In(time.UTC), true
	}
	return time.Time{}, false
}

func parseRelativeAlarmDuration(raw string, defaultDirection int) (time.Duration, error) {
	val := strings.TrimSpace(raw)
	if val == "" {
		return 0, fmt.Errorf(testutil.ErrMsgEmptyDuration)
	}

	sign := 0
	if strings.HasPrefix(val, "+") {
		sign = 1
		val = strings.TrimSpace(val[1:])
	} else if strings.HasPrefix(val, "-") {
		sign = -1
		val = strings.TrimSpace(val[1:])
	}

	dur, err := parseAlarmDurationValue(val)
	if err != nil {
		return 0, err
	}

	if sign == 0 {
		sign = defaultDirection
	}
	if sign == 0 {
		sign = -1
	}

	if sign < 0 {
		return -dur, nil
	}
	return dur, nil
}

func parseAlarmDurationValue(raw string) (time.Duration, error) {
	val := strings.TrimSpace(raw)
	if val == "" {
		return 0, fmt.Errorf(testutil.ErrMsgEmptyDuration)
	}
	if strings.HasPrefix(val, "+") {
		val = strings.TrimSpace(val[1:])
	}
	if strings.HasPrefix(val, "-") {
		return 0, fmt.Errorf(testutil.ErrMsgDurationMustBePositive)
	}

	if d, err := ParseHumanDuration(val); err == nil {
		return d, nil
	}
	if d, err := time.ParseDuration(val); err == nil {
		if d < 0 {
			return 0, fmt.Errorf(testutil.ErrMsgDurationMustBePositive)
		}
		return d, nil
	}
	if strings.HasPrefix(strings.ToUpper(val), "P") {
		return parseICSDuration(val)
	}
	return 0, fmt.Errorf("unrecognized duration format %q", raw)
}

func parseICSDuration(raw string) (time.Duration, error) {
	val := strings.ToUpper(strings.TrimSpace(raw))
	if val == "" {
		return 0, fmt.Errorf(testutil.ErrMsgEmptyDuration)
	}
	if strings.HasPrefix(val, "+") {
		val = strings.TrimSpace(val[1:])
	}
	if strings.HasPrefix(val, "-") {
		return 0, fmt.Errorf(testutil.ErrMsgDurationMustBePositive)
	}
	if !strings.HasPrefix(val, "P") {
		return 0, fmt.Errorf(testutil.ErrMsgInvalidICSDuration, raw)
	}

	matches := icsDurationRe.FindStringSubmatch(val)
	if matches == nil {
		return 0, fmt.Errorf(testutil.ErrMsgInvalidICSDuration, raw)
	}

	var total time.Duration
	if matches[1] != "" {
		total += time.Duration(atoiSafe(matches[1])) * 7 * 24 * time.Hour
	}
	if matches[2] != "" {
		total += time.Duration(atoiSafe(matches[2])) * 24 * time.Hour
	}
	if matches[3] != "" {
		total += time.Duration(atoiSafe(matches[3])) * time.Hour
	}
	if matches[4] != "" {
		total += time.Duration(atoiSafe(matches[4])) * time.Minute
	}
	if matches[5] != "" {
		total += time.Duration(atoiSafe(matches[5])) * time.Second
	}

	if total == 0 {
		return 0, fmt.Errorf(testutil.ErrMsgInvalidICSDuration, raw)
	}
	return total, nil
}

func parseBoolish(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func atoiSafe(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
