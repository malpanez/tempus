package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"tempus/internal/config"
	"tempus/internal/i18n"
	"tempus/internal/parsing"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	output      string
	confirmed   bool
}

// defaultAlarmProfileOrder pins the display order of the built-in profiles;
// any user-defined profiles are appended alphabetically.
var defaultAlarmProfileOrder = []string{"adhd-default", "adhd-countdown", "medication", "single"}

// alarmProfilesOrDefault returns the configured alarm profiles, falling back
// to the config package defaults when none are configured.
func alarmProfilesOrDefault(cfg *config.Config) map[string][]string {
	if cfg != nil && len(cfg.AlarmProfiles) > 0 {
		return cfg.AlarmProfiles
	}
	if defaults, err := config.Load(); err == nil && defaults != nil {
		return defaults.AlarmProfiles
	}
	return nil
}

// alarmProfileDisplayName resolves the localized display name for a profile,
// falling back to the raw profile name when no locale key exists.
func alarmProfileDisplayName(tr *i18n.Translator, name string) string {
	if tr != nil {
		key := "alarm_profile_name_" + strings.ReplaceAll(name, "-", "_")
		if v := tr.T(key); v != key {
			return v
		}
	}
	return name
}

// alarmProfileOptions builds the wizard Select options from the profile
// definitions so the offsets shown can never drift from the config values.
func alarmProfileOptions(cfg *config.Config, tr *i18n.Translator) []huh.Option[string] {
	profiles := alarmProfilesOrDefault(cfg)

	names := make([]string, 0, len(profiles))
	seen := make(map[string]bool, len(profiles))
	for _, name := range defaultAlarmProfileOrder {
		if offsets, ok := profiles[name]; ok && len(offsets) > 0 {
			names = append(names, name)
			seen[name] = true
		}
	}
	extra := make([]string, 0, len(profiles))
	for name, offsets := range profiles {
		if seen[name] || len(offsets) == 0 {
			continue
		}
		extra = append(extra, name)
	}
	sort.Strings(extra)
	names = append(names, extra...)

	opts := make([]huh.Option[string], 0, len(names)+2)
	for _, name := range names {
		label := fmt.Sprintf("%s (%s)", alarmProfileDisplayName(tr, name), strings.Join(profiles[name], ", "))
		opts = append(opts, huh.NewOption(label, name))
	}
	return append(opts,
		huh.NewOption("None", "none"),
		huh.NewOption("Custom...", "custom"),
	)
}

// The wizard runs as a sequence of small forms with explicit control flow
// instead of one form with WithHideFunc groups: huh v1.0's accessible mode
// (forced by TERM=dumb, used by screen readers) ignores hide functions and
// would prompt for conditional fields regardless of earlier answers.

func buildInteractiveForm(app *App, vars *interactiveVars) *huh.Form {
	return huh.NewForm(append(coreWizardGroups(app, vars), confirmWizardGroup(vars))...)
}

func coreWizardGroups(app *App, vars *interactiveVars) []*huh.Group {
	return []*huh.Group{
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
				Title("Timezone (IANA or city, e.g. America/New_York, madrid)").
				Value(&vars.timezone).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("timezone is required")
					}
					_, err := ResolveTimezone(s)
					return err
				}),
		).Title("Step 3/7 - Timezone"),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Alarm profile").
				Options(alarmProfileOptions(app.Config, app.Translator)...).
				Value(&vars.alarmProf),
		).Title("Step 4/7 - Alarms"),
	}
}

func customAlarmGroup(vars *interactiveVars) *huh.Group {
	return huh.NewGroup(
		huh.NewInput().
			Title("Custom alarm offsets (comma-separated, e.g. -2h,-30m,-5m)").
			Value(&vars.customAlarm).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("at least one alarm offset is required")
				}
				return nil
			}),
	).Title("Step 4/7 - Custom Alarms")
}

func categoriesGroup(vars *interactiveVars) *huh.Group {
	return huh.NewGroup(
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
	).Title("Step 5/7 - Categories")
}

func customCategoryGroup(vars *interactiveVars) *huh.Group {
	return huh.NewGroup(
		huh.NewInput().
			Title("Custom category name").
			Value(&vars.customCat),
	).Title("Step 5/7 - Custom Category")
}

func detailsGroup(vars *interactiveVars) *huh.Group {
	return huh.NewGroup(
		huh.NewInput().Title("Location (optional)").Value(&vars.location),
		huh.NewInput().Title("Description (optional)").Value(&vars.description),
	).Title("Step 6/7 - Details")
}

func confirmWizardGroup(vars *interactiveVars) *huh.Group {
	return huh.NewGroup(
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
	).Title("Step 7/7 - Confirm")
}

// runWizardForm wraps form.Run with the shared abort/error handling.
func runWizardForm(form *huh.Form) error {
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return fmt.Errorf("event creation aborted")
		}
		return fmt.Errorf("interactive form: %w", err)
	}
	return nil
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

	tz, err := ResolveTimezone(vars.timezone)
	if err != nil {
		return err
	}
	if tz == "UTC" {
		tz = ""
	}
	vars.timezone = tz

	processedCategories := processCategories(vars.categories, vars.customCat)
	alarmSpecs := resolveAlarmSpecs(vars.alarmProf, vars.customAlarm)

	output := strings.TrimSpace(vars.output)
	if output == "" {
		outputDir := app.Config.OutputDir
		if outputDir == "" {
			outputDir = "."
		}
		output = filepath.Join(outputDir, Slugify(vars.summary)+".ics")
	}

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

	cal, err := createCalendarWithEvent(opts, result.Start, result.End, app.Config)
	if err != nil {
		return err
	}
	return writeCalendarOutput(app, cal, opts.output)
}

// wizardCompatibleFlags are the create flags that --interactive honors;
// any other changed flag is a hard error — the wizard must never silently
// ignore something the user typed.
var wizardCompatibleFlags = map[string]bool{
	"interactive": true,
	"output":      true,
}

func runInteractive(app *App, cmd *cobra.Command) error {
	var incompatible []string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if !wizardCompatibleFlags[f.Name] {
			incompatible = append(incompatible, "--"+f.Name)
		}
	})
	if len(incompatible) > 0 {
		return fmt.Errorf("flags %s cannot be combined with --interactive: the wizard collects those values itself (only -o/--output is honored)", strings.Join(incompatible, ", "))
	}

	vars := &interactiveVars{
		timezone:  app.Config.Timezone,
		alarmProf: app.Config.DefaultAlarmProfile,
	}
	vars.output, _ = cmd.Flags().GetString("output")
	if vars.alarmProf == "" {
		vars.alarmProf = "adhd-default"
	}

	if err := runWizardForm(huh.NewForm(coreWizardGroups(app, vars)...)); err != nil {
		return err
	}
	if vars.alarmProf == "custom" {
		if err := runWizardForm(huh.NewForm(customAlarmGroup(vars))); err != nil {
			return err
		}
	}
	if err := runWizardForm(huh.NewForm(categoriesGroup(vars))); err != nil {
		return err
	}
	if hasOtherCategory(vars.categories) {
		if err := runWizardForm(huh.NewForm(customCategoryGroup(vars))); err != nil {
			return err
		}
	}
	if err := runWizardForm(huh.NewForm(detailsGroup(vars), confirmWizardGroup(vars))); err != nil {
		return err
	}

	if !vars.confirmed {
		fmt.Fprintf(app.Stdout, "Event creation cancelled.\n")
		return nil
	}

	return createEventFromWizard(app, vars)
}

func hasOtherCategory(categories []string) bool {
	for _, c := range categories {
		if c == "other" {
			return true
		}
	}
	return false
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
