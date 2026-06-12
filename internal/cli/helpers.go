package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/config"
	"tempus/internal/normalizer"
	"tempus/internal/parsing"
	"tempus/internal/testutil"
	"tempus/internal/utils"
)

// timeNow is the package clock; tests may override it for determinism.
var timeNow = time.Now

func EnsureDirForFile(path string) error {
	dir := strings.TrimSpace(filepath.Dir(path))
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o750)
}

func ExtractDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func ParseBoolish(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func SplitDelimited(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func ValueAsString(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(x)
	case fmt.Stringer:
		return strings.TrimSpace(x.String())
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%g", x))
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", x))
	}
}

func ValueAsBool(v interface{}) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case float64:
		return x != 0
	case string:
		return ParseBoolish(x)
	default:
		return ParseBoolish(fmt.Sprintf("%v", x))
	}
}

func valueAsSlice(v interface{}, eachFn func(string, *[]string), splitFn func(string) []string) []string {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []interface{}:
		var out []string
		for _, item := range x {
			val := strings.TrimSpace(ValueAsString(item))
			if val != "" {
				eachFn(val, &out)
			}
		}
		return out
	case []string:
		var out []string
		for _, item := range x {
			if strings.TrimSpace(item) != "" {
				eachFn(item, &out)
			}
		}
		return out
	case string:
		return splitFn(x)
	default:
		val := strings.TrimSpace(fmt.Sprintf("%v", x))
		if val == "" {
			return nil
		}
		return splitFn(val)
	}
}

func ValueAsStringSlice(v interface{}) []string {
	return valueAsSlice(v,
		func(s string, out *[]string) { *out = append(*out, strings.TrimSpace(s)) },
		SplitDelimited,
	)
}

func ValueAsAlarmSlice(v interface{}) []string {
	return valueAsSlice(v,
		func(s string, out *[]string) {
			for _, part := range calendar.SplitAlarmInput(s) {
				if strings.TrimSpace(part) != "" {
					*out = append(*out, part)
				}
			}
		},
		calendar.SplitAlarmInput,
	)
}

func EnsureICSExtension(name string) string {
	n := strings.TrimSpace(name)
	if n == "" {
		return "event.ics"
	}
	if strings.HasSuffix(strings.ToLower(n), ".ics") {
		return n
	}
	return n + ".ics"
}

func EnsureUniquePath(path string) string {
	clean := filepath.Clean(path)
	if _, err := os.Stat(clean); errors.Is(err, os.ErrNotExist) {
		return clean
	}

	dir := filepath.Dir(clean)
	base := filepath.Base(clean)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if ext == "" {
		ext = ".ics"
	}

	for i := 2; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%d%s", name, i, ext))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}

func Slugify(s string) string {
	return utils.Slugify(s)
}

func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func AtoiSafe(s string) int {
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

func PrintOK(w io.Writer, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(w, "\u2705 %s", msg)
}

func PrintErr(w io.Writer, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(w, "\u274c %s", msg)
}

var reParen = regexp.MustCompile(`\s*\([^(]*\)\s*$`)

func CleanDisplay(s string) string {
	return reParen.ReplaceAllString(s, "")
}

var clockOnlyRe = regexp.MustCompile(`^\d{1,2}:\d{2}$`)

func LooksLikeClock(s string) bool {
	return clockOnlyRe.MatchString(strings.TrimSpace(s))
}

func PrependToday(clock, tz string) string {
	return normalizer.PrependToday(clock, tz)
}

func FmtDurationHuman(d time.Duration) string {
	return parsing.FmtDurationHuman(d)
}

func SplitDateTime(s string) (string, string) {
	return parsing.SplitDateTime(s)
}

func stdoutWriter(app *App) io.Writer {
	if app.Stdout != nil {
		return app.Stdout
	}
	return os.Stdout
}

func parseDurationEnd(startTime time.Time, durStr string) (time.Time, error) {
	dur, err := calendar.ParseHumanDuration(durStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration %q: %v", durStr, err)
	}
	if dur <= 0 {
		return time.Time{}, fmt.Errorf(testutil.ErrMsgDurationGreaterThanZero)
	}
	return startTime.Add(dur), nil
}

func addExDates(event *calendar.Event, exdates []string, startTZ string, allDay bool) error {
	if len(exdates) == 0 {
		return nil
	}
	tzForEx := strings.TrimSpace(event.StartTZ)
	if tzForEx == "" {
		tzForEx = strings.TrimSpace(startTZ)
	}
	for _, raw := range exdates {
		parsed, err := parsing.ParseExDateValues([]string{raw}, tzForEx, allDay)
		if err != nil {
			return fmt.Errorf("invalid exdate %q: %w", raw, err)
		}
		event.ExDates = append(event.ExDates, parsed...)
	}
	return nil
}

// ResolveTimezone is the single validation chokepoint for user timezone
// input. It trims, accepts valid IANA zones as-is, resolves known city
// aliases (madrid → Europe/Madrid), and rejects anything else — no raw
// string ever reaches DTSTART;TZID=. Empty input resolves to "".
func ResolveTimezone(input string) (string, error) {
	tz := strings.TrimSpace(input)
	if tz == "" {
		return "", nil
	}
	if _, err := time.LoadLocation(tz); err == nil {
		return tz, nil
	}
	if iana, err := cityToIANA(tz); err == nil {
		return iana, nil
	}
	return "", fmt.Errorf("invalid timezone %q: not an IANA zone or known city alias; try 'tempus timezone list --search %s'", input, input)
}

// warnMissingVTZ tells the user when a zone has no embedded VTIMEZONE
// definition: the ICS is still usable in modern clients, but strict ones
// (Outlook classic) may not resolve the TZID. UTC needs no VTIMEZONE.
func warnMissingVTZ(w io.Writer, tz string) {
	if tz == "" || tz == "UTC" || calendar.HasVTZDefinition(tz) {
		return
	}
	fmt.Fprintf(w, "Warning: no VTIMEZONE definition embedded for %q — most clients resolve IANA names, but strict ones (e.g. Outlook classic) may not\n", tz)
}

// expandAlarmProfiles resolves profile:<name> specs against the loaded
// configuration. Unknown profiles are an error naming the available ones.
func expandAlarmProfiles(cfg *config.Config, alarmSpecs []string) ([]string, error) {
	expanded := make([]string, 0, len(alarmSpecs))
	for _, spec := range alarmSpecs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		if name, ok := strings.CutPrefix(spec, "profile:"); ok {
			name = strings.TrimSpace(name)
			profile := cfg.GetAlarmProfile(name)
			if profile == nil {
				return nil, fmt.Errorf("alarm profile %q not found. Available: %s",
					name, strings.Join(cfg.ListAlarmProfiles(), ", "))
			}
			expanded = append(expanded, profile...)
			continue
		}
		expanded = append(expanded, spec)
	}
	return expanded, nil
}
