package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewLintCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Validate ICS files for common issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(app, cmd, args)
		},
	}
	cmd.Flags().StringArray("file", []string{}, "ICS file(s) to lint (repeat flag for multiple files)")
	return cmd
}

func runLint(app *App, cmd *cobra.Command, _ []string) error {
	w := stdoutWriter(app)
	paths, _ := cmd.Flags().GetStringArray("file")
	if len(paths) == 0 {
		return fmt.Errorf("--file is required (repeat flag for multiple files)")
	}

	var errs []string
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if err := lintICSFile(path); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		PrintOK(w, "Lint passed: %s\n", path)
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
	eventIndex   int
	eventFields  map[string]string
	eventIssues  []string
}

func newLintState() lintState {
	return lintState{
		eventFields: make(map[string]string, 8),
	}
}

func lintICSFile(path string) error {
	lines, err := loadAndValidateICSFile(path)
	if err != nil {
		return err
	}

	state := newLintState()
	for _, line := range lines {
		processLintLine(&state, line)
	}

	return validateLintResults(state)
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
	case strings.EqualFold(line, "BEGIN:VEVENT"):
		handleBeginEvent(state)
	case strings.EqualFold(line, "END:VEVENT"):
		handleEndEvent(state)
	default:
		handleEventProperty(state, line)
	}
}

func handleBeginEvent(state *lintState) {
	state.inEvent = true
	state.eventSeen = true
	state.eventIndex++
	state.eventFields = make(map[string]string, 8)
}

func handleEndEvent(state *lintState) {
	if !state.inEvent {
		state.eventIssues = append(state.eventIssues, "unexpected END:VEVENT without matching BEGIN:VEVENT")
		return
	}
	state.inEvent = false

	label := buildEventLabel(state.eventIndex, state.eventFields)
	validateEventFields(state, label)
}

func buildEventLabel(index int, fields map[string]string) string {
	label := fmt.Sprintf("VEVENT #%d", index)
	if summary := strings.TrimSpace(fields["SUMMARY"]); summary != "" {
		label = fmt.Sprintf("%s (%s)", label, summary)
	}
	return label
}

func validateEventFields(state *lintState, label string) {
	requiredFields := []string{"UID", "SUMMARY", "DTSTART"}
	for _, key := range requiredFields {
		if strings.TrimSpace(state.eventFields[key]) == "" {
			state.eventIssues = append(state.eventIssues, fmt.Sprintf("%s missing %s", label, key))
		}
	}

	_, hasEnd := state.eventFields["DTEND"]
	_, hasDuration := state.eventFields["DURATION"]
	if !hasEnd && !hasDuration {
		state.eventIssues = append(state.eventIssues, fmt.Sprintf("%s missing DTEND or DURATION", label))
	}
}

func handleEventProperty(state *lintState, line string) {
	if !state.inEvent {
		return
	}

	name, value, ok := parseICSProperty(line)
	if ok {
		state.eventFields[name] = value
	}
}

func validateLintResults(state lintState) error {
	if !state.calendarSeen {
		return fmt.Errorf("missing BEGIN:VCALENDAR")
	}
	if !state.eventSeen {
		return fmt.Errorf("no VEVENT blocks found")
	}
	if len(state.eventIssues) > 0 {
		return fmt.Errorf("%s", strings.Join(state.eventIssues, "; "))
	}
	return nil
}

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
			current.WriteString(strings.TrimLeft(raw, " \t"))
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
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", false
	}
	if idx := strings.IndexRune(key, ';'); idx != -1 {
		key = key[:idx]
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	val := strings.TrimSpace(parts[1])
	return key, val, true
}
