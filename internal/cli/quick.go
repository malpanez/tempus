package cli

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/constants"

	"github.com/charmbracelet/huh"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/spf13/cobra"
)

func NewQuickCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quick [natural language event description]",
		Short: "Create a new event from a single sentence (experimental)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuick(app, cmd, args)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (optional)")
	cmd.Flags().StringP("timezone", "t", "", "Default timezone (overrides config)")

	return cmd
}

type quickParsedEvent struct {
	Summary   string
	StartTime time.Time
	EndTime   time.Time
	Location  string
	InputText string
}

func runQuick(app *App, cmd *cobra.Command, args []string) error {
	details, err := parseQuickInput(args[0])
	if err != nil {
		return err
	}

	finalTZ, err := resolveQuickTimezone(app, cmd)
	if err != nil {
		return err
	}
	applyTimezoneToDetails(&details, finalTZ)

	confirmed, err := confirmQuickEvent(app, details, finalTZ)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Operation cancelled.")
		return nil
	}

	output := getQuickOutput(cmd, details.Summary)
	return writeQuickCalendar(app, details, finalTZ, output)
}

func parseQuickInput(text string) (quickParsedEvent, error) {
	w := when.New(nil)
	w.Add(en.All...)

	res, err := w.Parse(text, timeNow())
	if err != nil || res == nil {
		return quickParsedEvent{}, fmt.Errorf("could not understand the date/time in your request. Please be more specific, e.g., 'tomorrow at 3pm'")
	}

	return extractEventDetails(text, res), nil
}

// resolveQuickTimezone resolves the -t flag (or the configured default)
// through the ResolveTimezone chokepoint. A resolved "UTC" keeps the empty
// form so naive times stay machine-local wall clock stamped per calendar
// rules, matching create's semantics.
func resolveQuickTimezone(app *App, cmd *cobra.Command) (string, error) {
	flagTZ, _ := cmd.Flags().GetString("timezone")
	defaultTZ := ""
	if app.Config != nil {
		defaultTZ = app.Config.Timezone
	}
	tz, err := ResolveTimezone(FirstNonEmpty(flagTZ, defaultTZ))
	if err != nil {
		return "", err
	}
	if tz == "UTC" {
		return "", nil
	}
	return tz, nil
}

// applyTimezoneToDetails reinterprets the parsed wall-clock time in the
// target zone ("3pm" means 3pm in that zone), instead of converting the
// instant with In(), which would shift the clock the user asked for.
func applyTimezoneToDetails(details *quickParsedEvent, tz string) {
	if tz == "" {
		return
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return
	}
	details.StartTime = rebuildInLocation(details.StartTime, loc)
	details.EndTime = rebuildInLocation(details.EndTime, loc)
}

func rebuildInLocation(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

func confirmQuickEvent(app *App, details quickParsedEvent, tz string) (bool, error) {
	fmt.Println("I understood the following event:")
	fmt.Printf("  Summary:   %s\n", details.Summary)
	fmt.Printf("  Start:     %s\n", details.StartTime.Format(constants.DateTimeFormatRFC1123))
	fmt.Printf("  End:       %s\n", details.EndTime.Format(constants.DateTimeFormatRFC1123))
	if details.Location != "" {
		fmt.Printf("  Location:  %s\n", details.Location)
	}
	if tz != "" {
		fmt.Printf("  Timezone:  %s\n", tz)
	}

	var confirmed bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Does this look correct?").
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, fmt.Errorf("event creation aborted")
		}
		return false, fmt.Errorf("interactive form: %w", err)
	}
	_ = app
	return confirmed, nil
}

func getQuickOutput(cmd *cobra.Command, summary string) string {
	output, _ := cmd.Flags().GetString("output")
	if output == "" {
		output = fmt.Sprintf("%s.ics", Slugify(summary))
	}
	return output
}

func writeQuickCalendar(app *App, details quickParsedEvent, tz, output string) error {
	cal := calendar.NewCalendar()
	cal.IncludeVTZ = true
	cal.Name = details.Summary
	if tz != "" {
		cal.SetDefaultTimezone(tz)
	}

	event := calendar.NewEvent(details.Summary, details.StartTime, details.EndTime)
	if details.Location != "" {
		event.Location = details.Location
	}
	if tz != "" {
		event.SetStartTimezone(tz)
		event.SetEndTimezone(tz)
	}

	cal.AddEvent(event)
	icsContent := cal.ToICS()

	if err := os.WriteFile(output, []byte(icsContent), 0600); err != nil {
		PrintErr(app.Stdout, constants.ErrMsgFailedToWriteFile, err)
		return err
	}
	PrintOK(app.Stdout, constants.MsgCreatedFile, output)

	return nil
}

func extractEventDetails(text string, res *when.Result) quickParsedEvent {
	summary := strings.TrimSpace(strings.Replace(text, res.Text, "", 1))

	durRegex := regexp.MustCompile(`(?i)\b(?:for|duration)\s+((?:\d+\s*)?(?:h|hr|hour|m|min|minute)s?)`)
	locRegex := regexp.MustCompile(`(?i)\b(?:at|in)\s+([\w\s\d]+)`)

	var duration time.Duration
	if matches := durRegex.FindStringSubmatch(text); len(matches) > 1 {
		summary = strings.Replace(summary, matches[0], "", 1)
		if d, err := calendar.ParseHumanDuration(matches[1]); err == nil {
			duration = d
		}
	}

	var location string
	if matches := locRegex.FindStringSubmatch(text); len(matches) > 1 {
		if !strings.Contains(res.Text, matches[1]) {
			location = strings.TrimSpace(matches[1])
			summary = strings.Replace(summary, matches[0], "", 1)
		}
	}

	summary = strings.TrimSpace(summary)
	summary = strings.Trim(summary, ",. ")

	endTime := res.Time.Add(time.Hour)
	if duration > 0 {
		endTime = res.Time.Add(duration)
	}

	return quickParsedEvent{
		Summary:   summary,
		StartTime: res.Time,
		EndTime:   endTime,
		Location:  location,
		InputText: text,
	}
}
