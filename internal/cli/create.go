package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/constants"
	"tempus/internal/testutil"

	"github.com/spf13/cobra"
)

func NewCreateCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [event-name]",
		Short: "Create a new ICS calendar event",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(app, cmd, args)
		},
	}

	cmd.Flags().StringP("start", "s", "", "Start date/time (YYYY-MM-DD HH:MM)")
	cmd.Flags().StringP("end", "e", "", "End date/time (YYYY-MM-DD HH:MM) or duration (e.g. 60m, 1h30m, 1:00, 90)")
	cmd.Flags().String("duration", "", "Duration (e.g. 45m, 1h30m, 90)")
	cmd.Flags().StringP("location", "L", "", "Event location")
	cmd.Flags().StringP("description", "d", "", "Event description")
	cmd.Flags().StringP("start-tz", "", "", "Start timezone")
	cmd.Flags().StringP("end-tz", "", "", "End timezone")
	cmd.Flags().StringP("output", "o", "", "Output file path")
	cmd.Flags().BoolP("all-day", "a", false, "All-day event")
	cmd.Flags().String("rrule", "", "Recurrence rule (RRULE), e.g. FREQ=DAILY;COUNT=10")
	cmd.Flags().StringArray("exdate", []string{}, "Exclude date/time (EXDATE). Repeat flag for multiple values (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	cmd.Flags().StringArray("alarm", []string{}, "Reminder (VALARM). Repeat for multiple values (e.g. 15m, trigger=-30m,description=Boarding Pass)")
	cmd.Flags().StringArray("category", []string{}, "Category label(s) to attach to the event (repeat flag for multiple values)")
	cmd.Flags().StringArray("attendee", []string{}, "Attendee email address (repeat flag for multiple values)")
	cmd.Flags().Int("priority", 0, "Event priority (1-9, 0 to omit)")
	cmd.Flags().BoolP("interactive", "i", false, "Create an event using an interactive questionnaire")

	return cmd
}

func runCreate(app *App, cmd *cobra.Command, args []string) error {
	interactive, _ := cmd.Flags().GetBool("interactive")
	if interactive {
		return runInteractive(app, cmd)
	}

	if len(args) == 0 {
		_ = cmd.Help()
		return nil
	}

	opts, err := parseCreateFlags(cmd, args)
	if err != nil {
		return err
	}

	startTime, endTime, err := parseCreateTimes(opts)
	if err != nil {
		return err
	}

	cal := createCalendarWithEvent(opts, startTime, endTime)
	return writeCalendarOutput(app, cal, opts.output)
}

type createOptions struct {
	summary     string
	startStr    string
	endStr      string
	durStr      string
	location    string
	description string
	startTZ     string
	endTZ       string
	output      string
	allDay      bool
	rrule       string
	exdates     []string
	alarms      []string
	categories  []string
	attendees   []string
	priority    int
}

func parseCreateFlags(cmd *cobra.Command, args []string) (*createOptions, error) {
	opts := &createOptions{summary: args[0]}
	opts.startStr, _ = cmd.Flags().GetString("start")
	opts.endStr, _ = cmd.Flags().GetString("end")
	opts.durStr, _ = cmd.Flags().GetString("duration")
	opts.location, _ = cmd.Flags().GetString("location")
	opts.description, _ = cmd.Flags().GetString("description")
	opts.startTZ, _ = cmd.Flags().GetString("start-tz")
	opts.endTZ, _ = cmd.Flags().GetString("end-tz")
	opts.output, _ = cmd.Flags().GetString("output")
	opts.allDay, _ = cmd.Flags().GetBool("all-day")
	opts.rrule, _ = cmd.Flags().GetString("rrule")
	opts.exdates, _ = cmd.Flags().GetStringArray("exdate")
	opts.alarms, _ = cmd.Flags().GetStringArray("alarm")
	opts.categories, _ = cmd.Flags().GetStringArray("category")
	opts.attendees, _ = cmd.Flags().GetStringArray("attendee")
	opts.priority, _ = cmd.Flags().GetInt("priority")

	if opts.priority < 0 || opts.priority > 9 {
		return nil, fmt.Errorf("priority must be between 0 and 9")
	}

	if strings.TrimSpace(opts.startStr) == "" {
		return nil, fmt.Errorf("start time is required (use --start)")
	}

	opts.startStr = normalizeTimeInput(opts.startStr, opts.startTZ, opts.endTZ)
	opts.endStr = normalizeTimeInput(opts.endStr, opts.startTZ, opts.endTZ)

	return opts, nil
}

func normalizeTimeInput(timeStr, startTZ, endTZ string) string {
	if timeStr != "" && LooksLikeClock(timeStr) {
		return PrependToday(timeStr, FirstNonEmpty(startTZ, endTZ, ""))
	}
	return timeStr
}

func parseCreateTimes(opts *createOptions) (startTime, endTime time.Time, err error) {
	if opts.allDay {
		return parseAllDayTimes(opts.startStr, opts.endStr)
	}
	return parseTimedEventTimes(opts.startStr, opts.endStr, opts.durStr)
}

func parseAllDayTimes(startStr, endStr string) (startTime, endTime time.Time, err error) {
	startTime, err = time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
	}

	if strings.TrimSpace(endStr) == "" {
		endTime = startTime.AddDate(0, 0, 1)
	} else {
		endDate, parseErr := time.Parse("2006-01-02", endStr)
		if parseErr != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", parseErr)
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

func parseTimedEventTimes(startStr, endStr, durStr string) (startTime, endTime time.Time, err error) {
	startTime, err = time.Parse("2006-01-02 15:04", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf(testutil.ErrMsgInvalidStartTimeFormat, err)
	}

	switch {
	case strings.TrimSpace(endStr) != "":
		endTime, err = parseEndTime(startTime, endStr)
	case strings.TrimSpace(durStr) != "":
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

func parseEndTime(startTime time.Time, endStr string) (time.Time, error) {
	if d, derr := calendar.ParseHumanDuration(endStr); derr == nil {
		if d <= 0 {
			return time.Time{}, fmt.Errorf(testutil.ErrMsgDurationGreaterThanZero)
		}
		return startTime.Add(d), nil
	}

	endTime, err := time.Parse("2006-01-02 15:04", endStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end time: %w", err)
	}
	return endTime, nil
}

func parseDurationEnd(startTime time.Time, durStr string) (time.Time, error) {
	d, err := calendar.ParseHumanDuration(durStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration: %v", err)
	}
	if d <= 0 {
		return time.Time{}, fmt.Errorf(testutil.ErrMsgDurationGreaterThanZero)
	}
	return startTime.Add(d), nil
}

func createCalendarWithEvent(opts *createOptions, startTime, endTime time.Time) *calendar.Calendar {
	cal := calendar.NewCalendar()
	cal.IncludeVTZ = true
	cal.Name = opts.summary
	if tz := FirstNonEmpty(opts.startTZ, opts.endTZ); strings.TrimSpace(tz) != "" {
		cal.SetDefaultTimezone(tz)
	}

	event := calendar.NewEvent(opts.summary, startTime, endTime)
	configureEvent(event, opts)
	cal.AddEvent(event)

	return cal
}

func configureEvent(event *calendar.Event, opts *createOptions) {
	event.AllDay = opts.allDay
	if opts.location != "" {
		event.Location = opts.location
	}
	if opts.description != "" {
		event.Description = opts.description
	}

	setEventTimezones(event, opts.startTZ, opts.endTZ)

	if strings.TrimSpace(opts.rrule) != "" {
		event.RRule = strings.TrimSpace(opts.rrule)
	}

	addExDates(event, opts.exdates, opts.startTZ, opts.allDay)
	addEventAlarms(event, opts.alarms, opts.startTZ)
	addEventCategories(event, opts.categories)
	addEventAttendees(event, opts.attendees)

	if opts.priority > 0 {
		event.Priority = opts.priority
	}
}

func setEventTimezones(event *calendar.Event, startTZ, endTZ string) {
	if startTZ != "" {
		event.SetStartTimezone(startTZ)
	}
	if endTZ != "" {
		event.SetEndTimezone(endTZ)
	} else if startTZ != "" {
		event.SetEndTimezone(startTZ)
	}
}

func addEventAlarms(event *calendar.Event, alarms []string, startTZ string) {
	if len(alarms) == 0 {
		return
	}

	defaultAlarmTZ := strings.TrimSpace(event.StartTZ)
	if defaultAlarmTZ == "" {
		defaultAlarmTZ = strings.TrimSpace(startTZ)
	}

	parsed, err := calendar.ParseAlarmSpecs(alarms, defaultAlarmTZ)
	if err == nil && len(parsed) > 0 {
		event.Alarms = append(event.Alarms, parsed...)
	}
}

func addEventCategories(event *calendar.Event, categories []string) {
	for _, cat := range categories {
		if c := strings.TrimSpace(cat); c != "" {
			event.AddCategory(c)
		}
	}
}

func addEventAttendees(event *calendar.Event, attendees []string) {
	for _, attendee := range attendees {
		if a := strings.TrimSpace(attendee); a != "" {
			event.AddAttendee(a)
		}
	}
}

func writeCalendarOutput(app *App, cal *calendar.Calendar, output string) error {
	icsContent := cal.ToICS()

	if output == "" {
		fmt.Print(icsContent)
		return nil
	}

	if err := os.WriteFile(output, []byte(icsContent), 0600); err != nil {
		PrintErr(app.Stdout, constants.ErrMsgFailedToWriteFile, err)
		return err
	}
	PrintOK(app.Stdout, constants.MsgCreatedFile, output)
	return nil
}
