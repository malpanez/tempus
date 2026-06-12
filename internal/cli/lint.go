package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func NewLintCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint [file ...]",
		Short: "Validate ICS files for common issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(app, cmd, args)
		},
	}
	cmd.Flags().StringArray("file", []string{}, "ICS file(s) to lint (repeat flag for multiple files)")
	return cmd
}

func runLint(app *App, cmd *cobra.Command, args []string) error {
	w := stdoutWriter(app)
	paths, _ := cmd.Flags().GetStringArray("file")
	paths = append(paths, args...)

	var errs []string
	linted := 0
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		linted++
		warnings, err := lintICSFile(path)
		for _, warning := range warnings {
			fmt.Fprintf(w, "⚠️  %s: %s\n", path, warning)
		}
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		PrintOK(w, "Lint passed: %s\n", path)
	}

	if linted == 0 {
		return fmt.Errorf("no files to lint: pass paths as arguments or with --file")
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

type lintState struct {
	calendarSeen bool
	eventSeen    bool
	inEvent      bool
	inVTimezone  bool
	eventIndex   int
	eventFields  map[string]string
	eventParams  map[string]string
	eventIssues  []string
	warnings     []string
	vtimezones   map[string]bool
	tzidRefs     map[string]string
}

func newLintState() lintState {
	return lintState{
		eventFields: make(map[string]string, 8),
		eventParams: make(map[string]string, 8),
		vtimezones:  make(map[string]bool, 2),
		tzidRefs:    make(map[string]string, 2),
	}
}

func lintICSFile(path string) ([]string, error) {
	lines, err := loadAndValidateICSFile(path)
	if err != nil {
		return nil, err
	}

	state := newLintState()
	for _, line := range lines {
		processLintLine(&state, line)
	}
	missingVTZWarnings(&state)

	return state.warnings, validateLintResults(state)
}

func loadAndValidateICSFile(path string) ([]string, error) {
	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory, expected file", path)
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	lines := unfoldICSLines(string(data))
	if len(lines) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	return lines, nil
}

func processLintLine(state *lintState, raw string) {
	line := strings.TrimSpace(raw)
	if line == "" {
		return
	}

	switch {
	case strings.EqualFold(line, "BEGIN:VCALENDAR"):
		state.calendarSeen = true
	case strings.EqualFold(line, "END:VCALENDAR"):
	case strings.EqualFold(line, "BEGIN:VTIMEZONE"):
		state.inVTimezone = true
	case strings.EqualFold(line, "END:VTIMEZONE"):
		state.inVTimezone = false
	case strings.EqualFold(line, "BEGIN:VEVENT"):
		handleBeginEvent(state)
	case strings.EqualFold(line, "END:VEVENT"):
		handleEndEvent(state)
	default:
		handleEventProperty(state, line)
	}
}

func handleBeginEvent(state *lintState) {
	if state.inEvent {
		state.eventIssues = append(state.eventIssues,
			fmt.Sprintf("VEVENT #%d not closed before next BEGIN:VEVENT", state.eventIndex))
	}
	state.inEvent = true
	state.eventSeen = true
	state.eventIndex++
	state.eventFields = make(map[string]string, 8)
	state.eventParams = make(map[string]string, 8)
}

func handleEndEvent(state *lintState) {
	if !state.inEvent {
		state.eventIssues = append(state.eventIssues, "unexpected END:VEVENT without matching BEGIN:VEVENT")
		return
	}
	state.inEvent = false

	label := buildEventLabel(state.eventIndex, state.eventFields)
	validateEventFields(state, label)
	validateUntilValueType(state, label)
	validateCategoriesEscaping(state, label)
	collectTZIDRefs(state, label)
}

func buildEventLabel(index int, fields map[string]string) string {
	label := fmt.Sprintf("VEVENT #%d", index)
	if summary := strings.TrimSpace(fields["SUMMARY"]); summary != "" {
		label = fmt.Sprintf("%s (%s)", label, summary)
	}
	return label
}

// validateEventFields enforces RFC 5545 §3.6.1: UID and DTSTAMP are
// REQUIRED; DTSTART is required when the calendar has no METHOD (the only
// mode tempus emits). SUMMARY and DTEND/DURATION are optional and no longer
// flagged.
func validateEventFields(state *lintState, label string) {
	requiredFields := []string{"UID", "DTSTAMP", "DTSTART"}
	for _, key := range requiredFields {
		if strings.TrimSpace(state.eventFields[key]) == "" {
			state.eventIssues = append(state.eventIssues, fmt.Sprintf("%s missing %s", label, key))
		}
	}
}

var rruleUntilRe = regexp.MustCompile(`(?i)(?:^|;)UNTIL=([0-9TZ]+)`)

// validateUntilValueType enforces RFC 5545 §3.3.10: UNTIL must have the
// same value type as DTSTART, and for timezone-aware date-times it must be
// in UTC (Z suffix).
func validateUntilValueType(state *lintState, label string) {
	rrule := state.eventFields["RRULE"]
	if rrule == "" {
		return
	}
	m := rruleUntilRe.FindStringSubmatch(rrule)
	if m == nil {
		return
	}
	until := m[1]
	dtstart := state.eventFields["DTSTART"]
	if dtstart == "" {
		return
	}

	startIsDate := !strings.Contains(dtstart, "T")
	untilIsDate := !strings.Contains(strings.ToUpper(until), "T")

	switch {
	case startIsDate && !untilIsDate:
		state.eventIssues = append(state.eventIssues,
			fmt.Sprintf("%s RRULE UNTIL is a date-time but DTSTART is a date (RFC 5545 §3.3.10 requires matching value types)", label))
	case !startIsDate && untilIsDate:
		state.eventIssues = append(state.eventIssues,
			fmt.Sprintf("%s RRULE UNTIL is a date but DTSTART is a date-time (RFC 5545 §3.3.10 requires matching value types)", label))
	case !startIsDate && strings.Contains(state.eventParams["DTSTART"], "TZID=") && !strings.HasSuffix(strings.ToUpper(until), "Z"):
		state.eventIssues = append(state.eventIssues,
			fmt.Sprintf("%s RRULE UNTIL must be UTC (Z suffix) when DTSTART carries a TZID (RFC 5545 §3.3.10)", label))
	}
}

var unescapedSemicolonRe = regexp.MustCompile(`(^|[^\\]);`)

func validateCategoriesEscaping(state *lintState, label string) {
	cats := state.eventFields["CATEGORIES"]
	if cats == "" {
		return
	}
	if unescapedSemicolonRe.MatchString(cats) {
		state.eventIssues = append(state.eventIssues,
			fmt.Sprintf("%s CATEGORIES contains an unescaped ';' (must be \\; per RFC 5545 §3.8.1.2)", label))
	}
}

var tzidParamRe = regexp.MustCompile(`TZID=([^;:]+)`)

func collectTZIDRefs(state *lintState, label string) {
	for _, prop := range []string{"DTSTART", "DTEND", "EXDATE"} {
		params := state.eventParams[prop]
		if m := tzidParamRe.FindStringSubmatch(params); m != nil {
			tzid := strings.TrimSpace(m[1])
			if _, seen := state.tzidRefs[tzid]; !seen {
				state.tzidRefs[tzid] = label
			}
		}
	}
}

func handleEventProperty(state *lintState, line string) {
	name, params, value, ok := parseICSPropertyFull(line)
	if !ok {
		return
	}
	if state.inVTimezone {
		if name == "TZID" {
			state.vtimezones[strings.TrimSpace(value)] = true
		}
		return
	}
	if !state.inEvent {
		return
	}
	state.eventFields[name] = value
	state.eventParams[name] = params
}

func validateLintResults(state lintState) error {
	if !state.calendarSeen {
		return fmt.Errorf("missing BEGIN:VCALENDAR")
	}
	if state.inEvent {
		return fmt.Errorf("VEVENT #%d is never closed (missing END:VEVENT)", state.eventIndex)
	}
	if !state.eventSeen {
		return fmt.Errorf("no VEVENT blocks found")
	}
	if len(state.eventIssues) > 0 {
		return fmt.Errorf("%s", strings.Join(state.eventIssues, "; "))
	}
	return nil
}

// missingVTZWarnings reports TZID references that have no matching
// VTIMEZONE component. Modern clients resolve IANA names anyway, so this
// is a warning rather than an error (strict clients like Outlook classic
// may reject the file).
func missingVTZWarnings(state *lintState) {
	for tzid, label := range state.tzidRefs {
		if tzid != "" && !state.vtimezones[tzid] {
			state.warnings = append(state.warnings,
				fmt.Sprintf("%s references TZID=%s with no matching VTIMEZONE component", label, tzid))
		}
	}
}

// unfoldICSLines reverses RFC 5545 §3.1 folding: a CRLF immediately
// followed by a single space or tab is removed, deleting ONLY that first
// whitespace character — any further whitespace is content.
func unfoldICSLines(data string) []string {
	data = strings.ReplaceAll(data, "\r\n", "\n")
	rawLines := strings.Split(data, "\n")
	lines := make([]string, 0, len(rawLines))

	var current strings.Builder
	for _, raw := range rawLines {
		if raw == "" && current.Len() == 0 {
			continue
		}
		if strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t") {
			current.WriteString(strings.TrimRight(raw[1:], "\r"))
			continue
		}
		if current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
		current.WriteString(strings.TrimRight(raw, "\r"))
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

func parseICSProperty(line string) (name, value string, ok bool) {
	name, _, value, ok = parseICSPropertyFull(line)
	return name, value, ok
}

// parseICSPropertyFull splits "NAME;PARAM=X:VALUE" into its three parts.
func parseICSPropertyFull(line string) (name, params, value string, ok bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", "", false
	}
	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", "", false
	}
	if idx := strings.IndexRune(key, ';'); idx != -1 {
		params = key[idx+1:]
		key = key[:idx]
	}
	name = strings.ToUpper(strings.TrimSpace(key))
	value = strings.TrimSpace(parts[1])
	return name, params, value, true
}
