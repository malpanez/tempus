package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"tempus/internal/config"
	"tempus/internal/parsing"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

type interactiveVars struct {
	summary     string
	startDate   string
	startTime   string
	duration    string
	allDay      bool
	timezone    string
	alarmProf   string
	customAlarm string
	categories  []string
	customCat   string
	location    string
	description string
	confirmed   bool
}

func buildInteractiveForm(vars *interactiveVars) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Event name").
				Value(&vars.summary).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("event name is required")
					}
					return nil
				}),
		).Title("Step 1/7 - Summary"),

		huh.NewGroup(
			huh.NewInput().
				Title("Start date (YYYY-MM-DD)").
				Value(&vars.startDate).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("start date is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Start time (HH:MM, leave empty for all-day)").
				Value(&vars.startTime),
			huh.NewInput().
				Title("Duration (e.g. 1h, 1h30m, 45m)").
				Value(&vars.duration),
			huh.NewConfirm().
				Title("All-day event?").
				Value(&vars.allDay),
		).Title("Step 2/7 - Date, Time & Duration"),

		huh.NewGroup(
			huh.NewInput().
				Title("Timezone (IANA, e.g. America/New_York)").
				Value(&vars.timezone).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("timezone is required")
					}
					return config.ValidateTimezone(s)
				}),
		).Title("Step 3/7 - Timezone"),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Alarm profile").
				Options(
					huh.NewOption("ADHD Default (-2h, -30m, -5m)", "adhd-default"),
					huh.NewOption("ADHD Countdown (-1h, -45m, -30m, -15m, -5m)", "adhd-countdown"),
					huh.NewOption("Medication (-30m, -5m)", "medication"),
					huh.NewOption("Single (-15m)", "single"),
					huh.NewOption("None", "none"),
					huh.NewOption("Custom...", "custom"),
				).
				Value(&vars.alarmProf),
		).Title("Step 4/7 - Alarms"),

		huh.NewGroup(
			huh.NewInput().
				Title("Custom alarm offsets (comma-separated, e.g. -2h,-30m,-5m)").
				Value(&vars.customAlarm).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("at least one alarm offset is required")
					}
					return nil
				}),
		).Title("Step 4/7 - Custom Alarms").WithHideFunc(func() bool {
			return vars.alarmProf != "custom"
		}),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Categories (optional, use space to select)").
				Options(
					huh.NewOption("Work", "work"),
					huh.NewOption("Health", "health"),
					huh.NewOption("Personal", "personal"),
					huh.NewOption("Travel", "travel"),
					huh.NewOption("School", "school"),
					huh.NewOption("Finance", "finance"),
					huh.NewOption("Other...", "other"),
				).
				Value(&vars.categories),
		).Title("Step 5/7 - Categories"),

		huh.NewGroup(
			huh.NewInput().
				Title("Custom category name").
				Value(&vars.customCat),
		).Title("Step 5/7 - Custom Category").WithHideFunc(func() bool {
			for _, c := range vars.categories {
				if c == "other" {
					return false
				}
			}
			return true
		}),

		huh.NewGroup(
			huh.NewInput().Title("Location (optional)").Value(&vars.location),
			huh.NewInput().Title("Description (optional)").Value(&vars.description),
		).Title("Step 6/7 - Details"),

		huh.NewGroup(
			huh.NewNote().
				Title("Event Summary").
				DescriptionFunc(func() string {
					return buildSummaryDescription(vars)
				}, vars),
			huh.NewConfirm().
				Title("Create this event?").
				Affirmative("Yes, create it").
				Negative("No, cancel").
				Value(&vars.confirmed),
		).Title("Step 7/7 - Confirm"),
	)
}

func buildSummaryDescription(vars *interactiveVars) string {
	cats := strings.Join(processCategories(vars.categories, vars.customCat), ", ")
	if cats == "" {
		cats = "(none)"
	}
	alarms := strings.Join(resolveAlarmSpecs(vars.alarmProf, vars.customAlarm), ", ")
	if alarms == "" {
		alarms = "(none)"
	}
	startStr := buildStartStr(vars.startDate, vars.startTime)
	return fmt.Sprintf(
		"Name:        %s\nDate/Time:   %s\nDuration:    %s\nTimezone:    %s\nAlarms:      %s\nCategories:  %s\nLocation:    %s\nDescription: %s",
		vars.summary,
		startStr,
		vars.duration,
		vars.timezone,
		alarms,
		cats,
		vars.location,
		vars.description,
	)
}

func createEventFromWizard(app *App, vars *interactiveVars) error {
	if strings.TrimSpace(vars.startTime) == "" {
		vars.allDay = true
	}

	processedCategories := processCategories(vars.categories, vars.customCat)
	alarmSpecs := resolveAlarmSpecs(vars.alarmProf, vars.customAlarm)

	outputDir := app.Config.OutputDir
	if outputDir == "" {
		outputDir = "."
	}
	output := filepath.Join(outputDir, Slugify(vars.summary)+".ics")

	opts := &createOptions{
		summary:     vars.summary,
		startStr:    buildStartStr(vars.startDate, vars.startTime),
		durStr:      vars.duration,
		startTZ:     vars.timezone,
		allDay:      vars.allDay,
		categories:  processedCategories,
		alarms:      alarmSpecs,
		location:    vars.location,
		description: vars.description,
		output:      output,
	}

	parseOpts := parsing.ParseOptions{
		StartDate: vars.startDate,
		StartTime: vars.startTime,
		Duration:  vars.duration,
		Timezone:  vars.timezone,
		AllDay:    vars.allDay,
		Summary:   vars.summary,
	}
	result, err := parsing.Parse(parseOpts)
	if err != nil {
		return fmt.Errorf("invalid date/time: %w", err)
	}

	cal := createCalendarWithEvent(opts, result.Start, result.End)
	return writeCalendarOutput(app, cal, opts.output)
}

func runInteractive(app *App, cmd *cobra.Command) error {
	vars := &interactiveVars{
		timezone:  app.Config.Timezone,
		alarmProf: app.Config.DefaultAlarmProfile,
	}
	if vars.alarmProf == "" {
		vars.alarmProf = "adhd-default"
	}

	form := buildInteractiveForm(vars)
	if err := form.Run(); err != nil {
		return nil
	}

	if !vars.confirmed {
		fmt.Fprintf(app.Stdout, "Event creation cancelled.\n")
		return nil
	}

	return createEventFromWizard(app, vars)
}

func processCategories(categories []string, customCat string) []string {
	result := make([]string, 0, len(categories))
	hasOther := false
	for _, c := range categories {
		if c == "other" {
			hasOther = true
			continue
		}
		result = append(result, c)
	}
	if hasOther && strings.TrimSpace(customCat) != "" {
		result = append(result, strings.TrimSpace(customCat))
	}
	return result
}

func resolveAlarmSpecs(profile, custom string) []string {
	switch profile {
	case "none":
		return []string{}
	case "custom":
		if strings.TrimSpace(custom) == "" {
			return []string{}
		}
		parts := strings.Split(custom, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				result = append(result, t)
			}
		}
		return result
	default:
		if profile == "" {
			return []string{}
		}
		return []string{"profile:" + profile}
	}
}

func buildStartStr(date, startTime string) string {
	d := strings.TrimSpace(date)
	t := strings.TrimSpace(startTime)
	if t == "" {
		return d
	}
	return d + " " + t
}
