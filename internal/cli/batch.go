package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/config"
	"tempus/internal/nd"
	"tempus/internal/parsing"
	"tempus/internal/testutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type batchFormat string

const (
	batchFormatCSV  batchFormat = "csv"
	batchFormatJSON batchFormat = "json"
	batchFormatYAML batchFormat = "yaml"
)

type batchRecord struct {
	Summary     string
	Start       string
	End         string
	Duration    string
	StartTZ     string
	EndTZ       string
	Location    string
	Description string
	AllDay      bool
	RRule       string
	ExDates     []string
	Categories  []string
	Alarms      []string
}

type batchOptions struct {
	input           string
	output          string
	formatFlag      string
	name            string
	defaultTZ       string
	dryRun          bool
	checkConflicts  bool
	maxEventsPerDay int
	addPrepTime     bool
	prepLabel       string
}

func NewBatchCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Create multiple ICS events from CSV, JSON, or YAML",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatch(app, cmd, args)
		},
	}

	cmd.Flags().StringP("input", "i", "", "Input file path (CSV, JSON, or YAML)")
	cmd.Flags().StringP("output", "o", "batch.ics", "Output ICS file path")
	cmd.Flags().String("format", "auto", "Input format: auto, csv, json, or yaml")
	cmd.Flags().String("name", "", "Calendar name (X-WR-CALNAME)")
	cmd.Flags().String("default-tz", "", "Default timezone for rows without start_tz")
	cmd.Flags().Bool("dry-run", false, "Validate batch file without creating output")
	cmd.Flags().Bool("check-conflicts", false, "Detect and warn about overlapping events")
	cmd.Flags().Int("max-events-per-day", 0, "Warn if any day exceeds this number of events (0=unlimited)")
	cmd.Flags().Bool("add-prep-time", false, "Auto-add preparation/transition time buffers (ADHD time boxing)")
	cmd.Flags().String("prep-label", "", "Custom prefix for prep time events (overrides config)")

	cmd.AddCommand(NewBatchTemplateCmd(app))

	return cmd
}

func runBatch(app *App, cmd *cobra.Command, _ []string) error {
	opts, err := parseBatchFlags(cmd)
	if err != nil {
		return err
	}

	opts.prepLabel = resolvePrepLabel(opts.prepLabel, app.Config)

	configTZ := ""
	if app.Config != nil {
		configTZ = app.Config.Timezone
	}
	defaultTZ, err := ResolveTimezone(FirstNonEmpty(opts.defaultTZ, configTZ))
	if err != nil {
		return err
	}
	if defaultTZ == "UTC" {
		defaultTZ = ""
	}
	opts.defaultTZ = defaultTZ
	warnMissingVTZ(app.Stderr, defaultTZ)

	records, _, err := loadBatchInput(opts)
	if err != nil {
		return err
	}

	corrections := make(map[string]string)
	if app.Config != nil && app.Config.SpellCorrections != nil {
		corrections = app.Config.SpellCorrections
	}
	spellCache := nd.NewSpellCheckCache(corrections)
	catCache := nd.NewCategoryCache()

	cal, validationErrors, err := buildBatchCalendar(records, opts, spellCache, catCache, app.Config)
	if err != nil {
		return err
	}

	warnings := collectBatchWarnings(cal.Events, opts)

	if opts.dryRun {
		return handleDryRun(app, validationErrors, warnings, records, opts.input, opts.output)
	}

	return writeBatchOutput(app, cal, warnings, opts.output, len(records))
}

func parseBatchFlags(cmd *cobra.Command) (*batchOptions, error) {
	opts := &batchOptions{}
	opts.input, _ = cmd.Flags().GetString("input")
	opts.output, _ = cmd.Flags().GetString("output")
	opts.formatFlag, _ = cmd.Flags().GetString("format")
	opts.name, _ = cmd.Flags().GetString("name")
	opts.defaultTZ, _ = cmd.Flags().GetString("default-tz")
	opts.dryRun, _ = cmd.Flags().GetBool("dry-run")
	opts.checkConflicts, _ = cmd.Flags().GetBool("check-conflicts")
	opts.maxEventsPerDay, _ = cmd.Flags().GetInt("max-events-per-day")
	opts.addPrepTime, _ = cmd.Flags().GetBool("add-prep-time")
	opts.prepLabel, _ = cmd.Flags().GetString("prep-label")

	opts.input = strings.TrimSpace(opts.input)
	if opts.input == "" {
		return nil, fmt.Errorf("--input is required")
	}

	return opts, nil
}

func resolvePrepLabel(flagValue string, cfg *config.Config) string {
	if flagValue != "" {
		return flagValue
	}
	if cfg != nil && cfg.PrepTimePrefix != "" {
		return cfg.PrepTimePrefix
	}
	return "Preparation"
}

func loadBatchInput(opts *batchOptions) ([]batchRecord, batchFormat, error) {
	format, err := detectBatchFormat(opts.formatFlag, opts.input)
	if err != nil {
		return nil, "", err
	}

	records, err := loadBatchRecords(opts.input, format)
	if err != nil {
		return nil, "", err
	}

	if len(records) == 0 {
		return nil, "", fmt.Errorf("no events found in %s", opts.input)
	}

	return records, format, nil
}

func buildBatchCalendar(records []batchRecord, opts *batchOptions, spellCache *nd.SpellCheckCache, catCache *nd.CategoryCache, cfg *config.Config) (*calendar.Calendar, []string, error) {
	cal := calendar.NewCalendar()
	cal.IncludeVTZ = true

	if strings.TrimSpace(opts.name) != "" {
		cal.Name = opts.name
	}
	if strings.TrimSpace(opts.defaultTZ) != "" {
		cal.SetDefaultTimezone(opts.defaultTZ)
	}

	var validationErrors []string
	for i, rec := range records {
		ev, err := buildEventFromBatch(rec, opts.defaultTZ, spellCache, catCache, cfg)
		if err != nil {
			if opts.dryRun {
				validationErrors = append(validationErrors, fmt.Sprintf("Row %d: %v", i+1, err))
				continue
			}
			return nil, nil, fmt.Errorf(testutil.ErrMsgRowFormat, i+1, err)
		}
		cal.AddEvent(ev)
	}

	if opts.addPrepTime {
		prepEvents := nd.GeneratePrepTimeEvents(cal.Events, opts.prepLabel)
		for _, prepEv := range prepEvents {
			cal.AddEvent(prepEv)
		}
	}

	return cal, validationErrors, nil
}

func collectBatchWarnings(events []calendar.Event, opts *batchOptions) []string {
	var warnings []string

	if opts.checkConflicts || opts.dryRun {
		conflicts := nd.DetectEventConflicts(events)
		if len(conflicts) > 0 {
			warnings = append(warnings, fmt.Sprintf("⚠️  Found %d time conflict(s):", len(conflicts)))
			for _, conflict := range conflicts {
				warnings = append(warnings, fmt.Sprintf("  • %s", conflict))
			}
		}
	}

	if opts.maxEventsPerDay > 0 || opts.dryRun {
		overwhelmDays := nd.DetectOverwhelmDays(events, opts.maxEventsPerDay)
		if len(overwhelmDays) > 0 {
			warnings = append(warnings, "⚠️  Days with high event load:")
			for _, day := range overwhelmDays {
				warnings = append(warnings, fmt.Sprintf("  • %s", day))
			}
		}
	}

	return warnings
}

func handleDryRun(app *App, validationErrors, warnings []string, records []batchRecord, input, output string) error {
	w := stdoutWriter(app)
	if len(validationErrors) > 0 {
		PrintErr(w, "Validation failed with %d error(s):\n", len(validationErrors))
		for _, errMsg := range validationErrors {
			fmt.Fprintf(w, "  ❌ %s\n", errMsg)
		}
		return fmt.Errorf("validation failed")
	}

	PrintOK(w, "✓ Validation passed: %d events ready to create\n", len(records))

	if len(warnings) > 0 {
		fmt.Fprintf(w, "\n")
		for _, warning := range warnings {
			fmt.Fprintln(w, warning)
		}
	}

	printDryRunSummary(w, records, input, output)
	return nil
}

func printDryRunSummary(w io.Writer, records []batchRecord, input, output string) {
	fmt.Fprintf(w, "\nEvent summary:\n")
	for i, rec := range records {
		summary := rec.Summary
		if summary == "" {
			summary = "(no summary)"
		}
		start := rec.Start
		if start == "" {
			start = "(no start)"
		}
		fmt.Fprintf(w, "  %d. %s - %s\n", i+1, summary, start)
	}
	fmt.Fprintf(w, "\nTo create the calendar file, run:\n")
	fmt.Fprintf(w, "  tempus batch -i %s -o %s\n", input, output)
}

func writeBatchOutput(app *App, cal *calendar.Calendar, warnings []string, output string, eventCount int) error {
	w := stdoutWriter(app)
	if len(warnings) > 0 {
		fmt.Fprintf(w, "\n")
		for _, warning := range warnings {
			fmt.Fprintln(w, warning)
		}
		fmt.Fprintf(w, "\n")
	}

	if err := EnsureDirForFile(output); err != nil {
		return err
	}

	if err := os.WriteFile(output, []byte(cal.ToICS()), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", output, err)
	}

	PrintOK(w, "Created: %s (%d events)\n", output, eventCount)
	return nil
}

func detectBatchFormat(flag, path string) (batchFormat, error) {
	switch strings.ToLower(strings.TrimSpace(flag)) {
	case "auto", "":
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".csv":
			return batchFormatCSV, nil
		case ".json":
			return batchFormatJSON, nil
		case ".yaml", ".yml":
			return batchFormatYAML, nil
		default:
			return "", fmt.Errorf("cannot infer format from %s; use --format csv|json|yaml", path)
		}
	case "csv":
		return batchFormatCSV, nil
	case "json":
		return batchFormatJSON, nil
	case "yaml", "yml":
		return batchFormatYAML, nil
	default:
		return "", fmt.Errorf("unsupported format %q (use csv, json, or yaml)", flag)
	}
}

func loadBatchRecords(path string, format batchFormat) ([]batchRecord, error) {
	switch format {
	case batchFormatCSV:
		return loadBatchFromCSV(path)
	case batchFormatJSON:
		return loadBatchFromJSON(path)
	case batchFormatYAML:
		return loadBatchFromYAML(path)
	default:
		return nil, fmt.Errorf("unknown batch format %q", format)
	}
}

func loadBatchFromCSV(path string) ([]batchRecord, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only file, close error is harmless

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	index := make(map[string]int, len(header))
	for i, col := range header {
		index[strings.ToLower(strings.TrimSpace(col))] = i
	}

	var records []batchRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(row) == 0 {
			continue
		}

		rec := batchRecord{
			Summary:     csvValue(row, index, "summary"),
			Start:       csvValue(row, index, "start"),
			End:         csvValue(row, index, "end"),
			Duration:    csvValue(row, index, "duration"),
			StartTZ:     csvValue(row, index, "start_tz"),
			EndTZ:       csvValue(row, index, "end_tz"),
			Location:    csvValue(row, index, "location"),
			Description: csvValue(row, index, "description"),
			RRule:       csvValue(row, index, "rrule"),
		}
		rec.AllDay = ParseBoolish(csvValue(row, index, "all_day"))

		if ex := csvValue(row, index, "exdate"); ex != "" {
			rec.ExDates = SplitDelimited(ex)
		}
		if cats := csvValue(row, index, "categories"); cats != "" {
			rec.Categories = SplitDelimited(cats)
		}
		if alarms := csvValue(row, index, "alarms"); alarms != "" {
			rec.Alarms = calendar.SplitAlarmInput(alarms)
		}

		records = append(records, rec)
	}

	return records, nil
}

func csvValue(row []string, index map[string]int, key string) string {
	if pos, ok := index[key]; ok {
		if pos < len(row) {
			return strings.TrimSpace(row[pos])
		}
	}
	return ""
}

func loadBatchFromStructured(path string, unmarshal func([]byte, interface{}) error) ([]batchRecord, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}
	var raw []map[string]interface{}
	if err := unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return parseMapsToRecords(raw), nil
}

func loadBatchFromJSON(path string) ([]batchRecord, error) {
	return loadBatchFromStructured(path, json.Unmarshal)
}

func loadBatchFromYAML(path string) ([]batchRecord, error) {
	return loadBatchFromStructured(path, yaml.Unmarshal)
}

func parseMapsToRecords(raw []map[string]interface{}) []batchRecord {
	records := make([]batchRecord, 0, len(raw))
	for _, item := range raw {
		rec := batchRecord{
			Summary:     ValueAsString(item["summary"]),
			Start:       ValueAsString(item["start"]),
			End:         ValueAsString(item["end"]),
			Duration:    ValueAsString(item["duration"]),
			StartTZ:     ValueAsString(item["start_tz"]),
			EndTZ:       ValueAsString(item["end_tz"]),
			Location:    ValueAsString(item["location"]),
			Description: ValueAsString(item["description"]),
			RRule:       ValueAsString(item["rrule"]),
			AllDay:      ValueAsBool(item["all_day"]),
			ExDates:     ValueAsStringSlice(item["exdate"]),
			Categories:  ValueAsStringSlice(item["categories"]),
			Alarms:      ValueAsAlarmSlice(item["alarms"]),
		}
		records = append(records, rec)
	}
	return records
}

func buildEventFromBatch(rec batchRecord, fallbackTZ string, spellCache *nd.SpellCheckCache, catCache *nd.CategoryCache, cfg *config.Config) (*calendar.Event, error) {
	summary, startStr, err := validateBatchRecord(rec, spellCache)
	if err != nil {
		return nil, err
	}

	startTZ, endTZ, err := resolveBatchTimezones(rec, fallbackTZ)
	if err != nil {
		return nil, err
	}
	startTime, endTime, err := parseBatchTimes(rec, startStr, startTZ, endTZ, summary)
	if err != nil {
		return nil, err
	}

	summaryWithEmoji := nd.AddEmojiToSummary(summary, rec.Categories)
	event := calendar.NewEvent(summaryWithEmoji, startTime, endTime)
	if err := configureBatchEvent(event, rec, startTZ, endTZ, catCache, cfg); err != nil {
		return nil, err
	}

	return event, nil
}

func validateBatchRecord(rec batchRecord, spellCache *nd.SpellCheckCache) (summary, startStr string, err error) {
	summary = spellCache.NormalizeAndCheck(strings.TrimSpace(rec.Summary))
	if summary == "" {
		return "", "", fmt.Errorf("summary is required")
	}

	startStr = parsing.NormalizeDateTimeInput(strings.TrimSpace(rec.Start))
	if startStr == "" {
		return "", "", fmt.Errorf("start is required")
	}

	return summary, startStr, nil
}

func resolveBatchTimezones(rec batchRecord, fallbackTZ string) (startTZ, endTZ string, err error) {
	startTZ, err = ResolveTimezone(FirstNonEmpty(rec.StartTZ, fallbackTZ))
	if err != nil {
		return "", "", err
	}
	endTZ, err = ResolveTimezone(rec.EndTZ)
	if err != nil {
		return "", "", err
	}
	if startTZ == "UTC" {
		startTZ = ""
	}
	if endTZ == "UTC" {
		endTZ = ""
	}
	if endTZ == "" {
		endTZ = startTZ
	}
	return startTZ, endTZ, nil
}

func parseBatchTimes(rec batchRecord, startStr, startTZ, endTZ, summary string) (startTime, endTime time.Time, err error) {
	if rec.AllDay {
		return parseBatchAllDayTimes(startStr, rec.End)
	}
	return parseBatchTimedEventTimes(rec, startStr, startTZ, endTZ, summary)
}

func parseBatchAllDayTimes(startStr, endStr string) (startTime, endTime time.Time, err error) {
	startDateStr := ExtractDate(startStr)
	startTime, err = time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date %q: %w", startStr, err)
	}

	if strings.TrimSpace(endStr) == "" {
		endTime = startTime.AddDate(0, 0, 1)
	} else {
		endDateStr := ExtractDate(endStr)
		endDate, parseErr := time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end date %q: %w", endStr, parseErr)
		}
		if endDate.Before(startTime) {
			return time.Time{}, time.Time{}, fmt.Errorf(testutil.ErrMsgEndDateAfterStart)
		}
		endTime = endDate.AddDate(0, 0, 1)
	}

	return startTime, endTime, nil
}

func parseBatchTimedEventTimes(rec batchRecord, startStr, startTZ, endTZ, summary string) (startTime, endTime time.Time, err error) {
	if LooksLikeClock(startStr) {
		startStr = PrependToday(startStr, startTZ)
	}
	startTime, err = time.Parse("2006-01-02 15:04", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start time %q: %w", rec.Start, err)
	}

	endTime, err = parseBatchEndTime(rec, startTime, endTZ, summary)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if !endTime.After(startTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("end time must be after start time")
	}

	return startTime, endTime, nil
}

func parseBatchEndTime(rec batchRecord, startTime time.Time, endTZ, summary string) (time.Time, error) {
	endStr := strings.TrimSpace(rec.End)

	switch {
	case endStr != "":
		return parseBatchExplicitEnd(endStr, startTime, endTZ, rec.End)
	case strings.TrimSpace(rec.Duration) != "":
		return parseDurationEnd(startTime, rec.Duration)
	default:
		return startTime.Add(nd.GetSmartDefaultDuration(summary, startTime)), nil
	}
}

func parseBatchExplicitEnd(endStr string, startTime time.Time, endTZ, originalEnd string) (time.Time, error) {
	if LooksLikeClock(endStr) {
		endStr = PrependToday(endStr, endTZ)
	}

	if dur, derr := calendar.ParseHumanDuration(endStr); derr == nil {
		if dur <= 0 {
			return time.Time{}, fmt.Errorf(testutil.ErrMsgDurationGreaterThanZero)
		}
		return startTime.Add(dur), nil
	}

	endTime, err := time.Parse("2006-01-02 15:04", endStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end time %q: %w", originalEnd, err)
	}
	return endTime, nil
}

func configureBatchEvent(event *calendar.Event, rec batchRecord, startTZ, endTZ string, catCache *nd.CategoryCache, cfg *config.Config) error {
	event.AllDay = rec.AllDay

	setEventTimezones(event, startTZ, endTZ)

	event.Location = strings.TrimSpace(rec.Location)
	event.Description = strings.TrimSpace(rec.Description)

	if strings.TrimSpace(rec.RRule) != "" {
		event.RRule = NormalizeRRuleUntil(strings.TrimSpace(rec.RRule), rec.AllDay)
	}

	addBatchCategories(event, rec.Categories, catCache)
	if err := addExDates(event, rec.ExDates, startTZ, rec.AllDay); err != nil {
		return err
	}
	return addEventAlarms(event, rec.Alarms, startTZ, cfg)
}

func addBatchCategories(event *calendar.Event, categories []string, catCache *nd.CategoryCache) {
	for _, cat := range categories {
		cat = strings.TrimSpace(cat)
		if cat != "" {
			validated := catCache.Validate(cat)
			event.AddCategory(validated)
		}
	}
}

// ========================================================================
// Batch Template Generator
// ========================================================================

// batchTemplateEvent is the single source of truth for batch template
// content. Its fields and yaml/json tags mirror the batch loader columns
// exactly, so every rendered format (csv, yaml, json) round-trips through
// `tempus batch -i <file>` without translation.
type batchTemplateEvent struct {
	Summary     string   `yaml:"summary" json:"summary"`
	Start       string   `yaml:"start" json:"start"`
	End         string   `yaml:"end,omitempty" json:"end,omitempty"`
	Duration    string   `yaml:"duration,omitempty" json:"duration,omitempty"`
	StartTZ     string   `yaml:"start_tz,omitempty" json:"start_tz,omitempty"`
	EndTZ       string   `yaml:"end_tz,omitempty" json:"end_tz,omitempty"`
	Location    string   `yaml:"location,omitempty" json:"location,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	RRule       string   `yaml:"rrule,omitempty" json:"rrule,omitempty"`
	AllDay      bool     `yaml:"all_day,omitempty" json:"all_day,omitempty"`
	ExDate      []string `yaml:"exdate,omitempty" json:"exdate,omitempty"`
	Categories  []string `yaml:"categories,omitempty" json:"categories,omitempty"`
	Alarms      []string `yaml:"alarms,omitempty" json:"alarms,omitempty"`
}

// batchTemplateColumns is the exact header understood by loadBatchFromCSV.
var batchTemplateColumns = []string{
	"summary", "start", "end", "duration", "start_tz", "end_tz",
	"location", "description", "rrule", "all_day", "exdate",
	"categories", "alarms",
}

const batchTemplateTypesList = "basic, adhd-routine, medication, work-meetings, medical, travel, family, school-event, recruiter-meeting, travel-day"

const (
	templateTZMadrid   = "Europe/Madrid"
	templateTZLondon   = "Europe/London"
	templateRRWeekdays = "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;COUNT=20"
	templateRRDaily30  = "FREQ=DAILY;COUNT=30"
)

func NewBatchTemplateCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template [type]",
		Short: "Generate a pre-filled template file for batch mode",
		Long: `Generate template files to quickly start creating events.

Available template types:
  basic              - Simple 3-event example
  adhd-routine       - Daily ADHD routine with medication and focus blocks
  medication         - Medication schedule with triple alarms
  work-meetings      - Recurring team meetings
  medical            - Healthcare appointments with prep reminders
  travel             - Travel itinerary with flights and hotels
  family             - Family calendar with mixed events
  school-event       - School term dates, pickups, and activities
  recruiter-meeting  - Job interview calls with prep reminders
  travel-day         - Single-day travel itinerary across timezones

Every template can be generated as CSV, YAML, or JSON (--format), and the
generated file is always loadable with: tempus batch -i <file>

Examples:
  tempus batch template basic -o my-events.csv
  tempus batch template adhd-routine -o routine.csv
  tempus batch template medication -o meds.yaml -f yaml
  tempus batch template travel -o trip.json -f json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchTemplate(app, cmd, args)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (required)")
	_ = cmd.MarkFlagRequired("output")
	cmd.Flags().StringP("format", "f", "csv", "Template format: csv, yaml, or json")

	return cmd
}

func runBatchTemplate(app *App, cmd *cobra.Command, args []string) error {
	w := stdoutWriter(app)
	templateType := strings.ToLower(strings.TrimSpace(args[0]))
	output, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")

	if output == "" {
		return fmt.Errorf("--output is required")
	}

	content, err := getBatchTemplateContent(templateType, format)
	if err != nil {
		return err
	}

	if err := os.WriteFile(output, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	PrintOK(w, "Template created: %s\n", output)
	fmt.Fprintf(w, "Edit the file and run: tempus batch -i %s -o calendar.ics\n", output)

	return nil
}

func getBatchTemplateContent(templateType, format string) (string, error) {
	events, err := batchTemplateEvents(templateType)
	if err != nil {
		return "", err
	}

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "csv":
		return renderBatchTemplateCSV(events)
	case "yaml", "yml":
		return renderBatchTemplateYAML(events)
	case "json":
		return renderBatchTemplateJSON(events)
	default:
		return "", fmt.Errorf("--format must be 'csv', 'yaml', or 'json', got %q", format)
	}
}

func batchTemplateEvents(templateType string) ([]batchTemplateEvent, error) {
	switch templateType {
	case "basic":
		return basicTemplateEvents(), nil
	case "adhd-routine":
		return adhdRoutineTemplateEvents(), nil
	case "medication", "meds":
		return medicationTemplateEvents(), nil
	case "work-meetings", "work":
		return workMeetingsTemplateEvents(), nil
	case "medical", "health":
		return medicalTemplateEvents(), nil
	case "travel":
		return travelTemplateEvents(), nil
	case "family":
		return familyTemplateEvents(), nil
	case "school-event":
		return schoolEventTemplateEvents(), nil
	case "recruiter-meeting":
		return recruiterMeetingTemplateEvents(), nil
	case "travel-day":
		return travelDayTemplateEvents(), nil
	default:
		return nil, fmt.Errorf("unknown template type: %s\nAvailable: %s", templateType, batchTemplateTypesList)
	}
}

func renderBatchTemplateCSV(events []batchTemplateEvent) (string, error) {
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	if err := w.Write(batchTemplateColumns); err != nil {
		return "", fmt.Errorf("writing template header: %w", err)
	}
	for _, ev := range events {
		allDay := ""
		if ev.AllDay {
			allDay = "true"
		}
		row := []string{
			ev.Summary, ev.Start, ev.End, ev.Duration, ev.StartTZ, ev.EndTZ,
			ev.Location, ev.Description, ev.RRule, allDay,
			strings.Join(ev.ExDate, ";"),
			strings.Join(ev.Categories, ";"),
			// "||" is the loader's multi-alarm separator; ";" cannot be
			// used here because key=value alarm specs contain ";" internally.
			strings.Join(ev.Alarms, "||"),
		}
		if err := w.Write(row); err != nil {
			return "", fmt.Errorf("writing template row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("rendering CSV template: %w", err)
	}
	return buf.String(), nil
}

func renderBatchTemplateYAML(events []batchTemplateEvent) (string, error) {
	data, err := yaml.Marshal(events)
	if err != nil {
		return "", fmt.Errorf("rendering YAML template: %w", err)
	}
	return string(data), nil
}

func renderBatchTemplateJSON(events []batchTemplateEvent) (string, error) {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return "", fmt.Errorf("rendering JSON template: %w", err)
	}
	return string(data) + "\n", nil
}

func basicTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "Team Meeting", Start: "2030-12-16 10:00", Duration: "1h",
			StartTZ: templateTZMadrid, Location: "Conference Room",
			Description: "Weekly sync",
			Categories:  []string{"Work", "Meeting"},
			Alarms:      []string{"-15m"},
		},
		{
			Summary: "Lunch Break", Start: "2030-12-16 13:00", Duration: "1h",
			StartTZ:    templateTZMadrid,
			Categories: []string{"Break"},
		},
		{
			Summary: "Doctor Appointment", Start: "2030-12-17 09:00", Duration: "45m",
			StartTZ: templateTZMadrid, Location: "Medical Center",
			Categories: []string{"Health"},
			Alarms:     []string{"trigger=-1d;description=Confirm appointment", "-2h"},
		},
	}
}

func adhdRoutineTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "Morning Medication", Start: "2030-12-16 08:00", Duration: "5m",
			StartTZ: templateTZMadrid, RRule: templateRRDaily30,
			Description: "Take morning meds with food",
			Categories:  []string{"Health", "Medication"},
			Alarms:      []string{"profile:medication"},
		},
		{
			Summary: "Deep Focus Block", Start: "2030-12-16 09:00", Duration: "2h",
			StartTZ: templateTZMadrid, RRule: templateRRWeekdays,
			Description: "NO meetings - deep work only",
			Categories:  []string{"Work", "Focus"},
			Alarms: []string{
				"trigger=-10m;description=Prepare workspace and eliminate distractions",
				"-1m",
				"trigger=2030-12-16 10:30;description=Halfway - stay focused",
			},
		},
		{
			Summary: "Transition Buffer", Start: "2030-12-16 11:00", Duration: "15m",
			StartTZ: templateTZMadrid, RRule: templateRRWeekdays,
			Description: "Stretch and reset before next task",
			Categories:  []string{"Break", "Transition"},
			Alarms:      []string{"-1m"},
		},
		{
			Summary: "Lunch + Walk", Start: "2030-12-16 13:00", Duration: "1h",
			StartTZ: templateTZMadrid, RRule: templateRRDaily30,
			Description: "Eat away from desk - go outside",
			Categories:  []string{"Break", "Health"},
			Alarms: []string{
				"-5m",
				"trigger=2030-12-16 13:30;description=Time to walk",
			},
		},
		{
			Summary: "Evening Medication", Start: "2030-12-16 20:00", Duration: "5m",
			StartTZ: templateTZMadrid, RRule: templateRRDaily30,
			Description: "Take evening meds",
			Categories:  []string{"Health", "Medication"},
			Alarms:      []string{"profile:medication"},
		},
		{
			Summary: "Wind Down Routine", Start: "2030-12-16 22:00", Duration: "30m",
			StartTZ: templateTZMadrid, RRule: templateRRDaily30,
			Description: "No screens - prepare for sleep",
			Categories:  []string{"Health", "Sleep"},
			Alarms:      []string{"-15m", "-5m"},
		},
	}
}

func medicationTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "Morning Meds - Methylphenidate 20mg", Start: "2030-12-16 08:00", Duration: "5m",
			StartTZ: templateTZMadrid, RRule: templateRRDaily30,
			Description: "Take with food. Don't skip.",
			Categories:  []string{"Health", "Medication"},
			Alarms:      []string{"profile:medication"},
		},
		{
			Summary: "Afternoon Meds - Methylphenidate 10mg", Start: "2030-12-16 14:00", Duration: "5m",
			StartTZ: templateTZMadrid, RRule: templateRRDaily30,
			Description: "Booster dose",
			Categories:  []string{"Health", "Medication"},
			Alarms:      []string{"profile:medication"},
		},
		{
			Summary: "Evening Meds - Melatonin 3mg", Start: "2030-12-16 21:00", Duration: "5m",
			StartTZ: templateTZMadrid, RRule: templateRRDaily30,
			Description: "Take 1 hour before bed",
			Categories:  []string{"Health", "Medication", "Sleep"},
			Alarms:      []string{"profile:medication"},
		},
	}
}

func workMeetingsTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "Team Standup", Start: "2030-12-16 09:30", Duration: "30m",
			StartTZ: templateTZMadrid, Location: "Video call - Zoom",
			RRule:       templateRRWeekdays,
			Description: "Daily sync with team",
			Categories:  []string{"Work", "Meeting"},
			Alarms:      []string{"-5m"},
		},
		{
			Summary: "Weekly 1:1 with Manager", Start: "2030-12-16 14:00", Duration: "45m",
			StartTZ: templateTZMadrid, Location: "Office - Meeting Room 3",
			RRule:       "FREQ=WEEKLY;BYDAY=MO;COUNT=12",
			Description: "Discuss progress and blockers",
			ExDate:      []string{"2030-12-23 14:00", "2030-12-30 14:00"},
			Categories:  []string{"Work", "1on1"},
			Alarms: []string{
				"trigger=-1d;description=Prepare agenda and questions",
				"-15m",
			},
		},
		{
			Summary: "Sprint Planning", Start: "2030-12-17 10:00", Duration: "2h",
			StartTZ: templateTZMadrid, Location: "Conference Room A",
			RRule:       "FREQ=WEEKLY;BYDAY=TU;COUNT=6",
			Description: "Plan next 2-week sprint",
			Categories:  []string{"Work", "Meeting", "Planning"},
			Alarms: []string{
				"trigger=-1h;description=Review backlog",
				"-15m",
			},
		},
		{
			Summary: "Client Demo", Start: "2030-12-19 16:00", Duration: "90m",
			StartTZ: templateTZMadrid, Location: "Video call - Google Meet",
			RRule:       "FREQ=WEEKLY;BYDAY=TH;COUNT=8",
			Description: "Demo progress to stakeholders",
			Categories:  []string{"Work", "Client", "Demo"},
			Alarms: []string{
				"trigger=-1d;description=Prepare demo script",
				"trigger=-2h;description=Test demo environment",
				"-30m",
			},
		},
	}
}

func medicalTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "Dentist - 6 Month Checkup", Start: "2030-12-20 10:00", Duration: "30m",
			StartTZ: templateTZMadrid, Location: "Dental Clinic - Main Street",
			Description: "Routine cleaning and checkup",
			Categories:  []string{"Health", "Dental"},
			Alarms: []string{
				"trigger=-1d;description=Call to confirm appointment",
				"trigger=-2h;description=Time to leave (30min drive)",
				"-5m",
			},
		},
		{
			Summary: "Therapy Session", Start: "2030-12-18 17:00", Duration: "1h",
			StartTZ: templateTZMadrid, Location: "Downtown Office - 3rd Floor",
			Description: "Weekly therapy appointment",
			Categories:  []string{"Health", "Mental Health"},
			Alarms: []string{
				"trigger=-1d;description=Think about topics to discuss",
				"trigger=-30m;description=Prepare - bring notebook",
				"-5m",
			},
		},
		{
			Summary: "General Practitioner Checkup", Start: "2031-01-10 09:00", Duration: "45m",
			StartTZ: templateTZMadrid, Location: "Health Center - Room 12",
			Description: "Annual physical examination",
			Categories:  []string{"Health", "Checkup"},
			Alarms: []string{
				"trigger=-1w;description=Schedule blood work if needed",
				"trigger=-1d;description=Confirm appointment",
				"trigger=-2h;description=Leave now (traffic)",
				"-15m",
			},
		},
		{
			Summary: "Lab Tests (Fasting Required)", Start: "2030-12-22 08:00", Duration: "15m",
			StartTZ: templateTZMadrid, Location: "Hospital Lab - Building C",
			Description: "Blood work - MUST FAST",
			Categories:  []string{"Health", "Tests"},
			Alarms: []string{
				"trigger=-1d;description=No food after 10pm tonight",
				"trigger=-12h;description=Fasting period begins - no food",
				"trigger=-1h;description=Drink water only",
				"-15m",
			},
		},
	}
}

func travelTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "Flight MAD → DUB", Start: "2030-12-25 08:30", End: "2030-12-25 10:00",
			StartTZ: templateTZMadrid, EndTZ: "Europe/Dublin",
			Location:    "Madrid Barajas Airport - Terminal 1",
			Description: "Ryanair FR1234 - Gate closes 08:00. Confirmation: ABC123",
			Categories:  []string{"Travel", "Flight"},
			Alarms: []string{
				"trigger=-1d;description=Check-in online opens",
				"trigger=-3h;description=Wake up and get ready",
				"trigger=2030-12-25 06:30;description=Leave for airport now (traffic)",
				"trigger=2030-12-25 07:45;description=Security checkpoint - gate closes at 08:00",
			},
		},
		{
			Summary: "Hotel Check-in", Start: "2030-12-25 12:00", Duration: "30m",
			StartTZ:     "Europe/Dublin",
			Location:    "Dublin City Hotel - 123 O'Connell Street",
			Description: "Confirmation: XYZ789. Room 305. Check-in after 14:00.",
			Categories:  []string{"Travel", "Accommodation"},
			Alarms: []string{
				"trigger=-1h;description=Head to hotel from airport",
			},
		},
		{
			Summary: "Return Flight DUB → MAD", Start: "2030-12-28 18:30", End: "2030-12-28 22:00",
			StartTZ: "Europe/Dublin", EndTZ: templateTZMadrid,
			Location:    "Dublin Airport - Terminal 2",
			Description: "Ryanair FR5678. Gate closes 18:00.",
			Categories:  []string{"Travel", "Flight"},
			Alarms: []string{
				"trigger=-1d;description=Check-in online",
				"trigger=-4h;description=Leave hotel for airport (bus takes 45min)",
				"trigger=2030-12-28 17:45;description=Final boarding call",
			},
		},
	}
}

func familyTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "School Drop-off", Start: "2030-12-16 08:15", Duration: "15m",
			StartTZ: templateTZMadrid, Location: "Elementary School",
			RRule:       templateRRWeekdays,
			Description: "Drop kids at school",
			Categories:  []string{"Family", "Kids"},
			Alarms: []string{
				"trigger=-30m;description=Kids breakfast and get ready",
				"trigger=-10m;description=Leave now",
			},
		},
		{
			Summary: "Soccer Practice", Start: "2030-12-17 17:00", Duration: "1h",
			StartTZ: templateTZMadrid, Location: "Sports Complex Field 3",
			RRule:       "FREQ=WEEKLY;BYDAY=TU,TH;COUNT=12",
			Description: "Lucas soccer practice",
			Categories:  []string{"Family", "Kids", "Sports"},
			Alarms: []string{
				"trigger=-1h;description=Prepare soccer bag and snacks",
				"-15m",
			},
		},
		{
			Summary: "Piano Lesson", Start: "2030-12-18 16:30", Duration: "45m",
			StartTZ: templateTZMadrid, Location: "Music Academy",
			RRule:       "FREQ=WEEKLY;BYDAY=WE;COUNT=10",
			Description: "Emma piano lesson",
			Categories:  []string{"Family", "Kids", "Music"},
			Alarms: []string{
				"trigger=-2h;description=Practice today",
				"-30m",
			},
		},
		{
			Summary: "Pediatrician Checkup", Start: "2030-12-20 10:00", Duration: "30m",
			StartTZ: templateTZMadrid, Location: "Pediatric Clinic",
			Description: "Annual checkup for both kids",
			Categories:  []string{"Family", "Kids", "Health"},
			Alarms: []string{
				"trigger=-1d;description=Confirm appointment",
				"-2h",
				"-30m",
			},
		},
		{
			Summary: "Date Night", Start: "2030-12-21 20:00", Duration: "2h",
			StartTZ: templateTZMadrid, Location: "Restaurant Downtown",
			Description: "Dinner reservation - babysitter confirmed",
			Categories:  []string{"Family", "Personal"},
			Alarms: []string{
				"trigger=-1d;description=Confirm babysitter",
				"trigger=-4h;description=Start getting ready",
				"-1h",
			},
		},
	}
}

func schoolEventTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "School starts Q3", Start: "2030-09-02", AllDay: true,
			Location:   "IES Cervantes",
			Categories: []string{"School", "Trimester"},
			Alarms:     []string{"-1d"},
		},
		{
			Summary: "Half-term holiday", Start: "2030-10-28", End: "2030-11-01", AllDay: true,
			Description: "Emma and Leo - no school",
			Categories:  []string{"School", "Vacation"},
		},
		{
			Summary: "Parent-teacher meeting", Start: "2030-11-15 17:00",
			Location:    "IES Cervantes",
			Description: "Emma - bring report card",
			Categories:  []string{"School", "Activity"},
			Alarms:      []string{"profile:adhd-default"},
		},
		{
			Summary: "Emma pickup 17:00", Start: "2030-09-03 17:00",
			Location:    "IES Cervantes",
			Description: "Gate 2",
			Categories:  []string{"School", "Transport"},
			Alarms:      []string{"profile:single"},
		},
		{
			Summary: "End of year concert", Start: "2031-06-20 18:00",
			Location:    "School auditorium",
			Description: "Bring camera",
			Categories:  []string{"School", "Activity"},
			Alarms:      []string{"-2h"},
		},
	}
}

func recruiterMeetingTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "Call with Sarah @ Acme Corp", Start: "2030-12-16 10:00", Duration: "30m",
			StartTZ:     templateTZMadrid,
			Description: "Acme Corp - Senior Developer. Recruiter: Sarah Jones. LinkedIn: linkedin.com/in/sarah",
			Categories:  []string{"Work", "Interview"},
			Alarms:      []string{"profile:adhd-default"},
		},
		{
			Summary: "Technical interview @ StartupX", Start: "2030-12-18 15:00", Duration: "1h",
			StartTZ:     "America/New_York",
			Description: "StartupX - Backend Engineer. Recruiter: Mike Chen. Phone: +1-555-0123",
			Categories:  []string{"Work", "Interview"},
			Alarms:      []string{"profile:adhd-default"},
		},
	}
}

func travelDayTemplateEvents() []batchTemplateEvent {
	return []batchTemplateEvent{
		{
			Summary: "MAD -> LHR BA456", Start: "2030-12-20 08:30", End: "2030-12-20 11:00",
			StartTZ: templateTZMadrid, EndTZ: templateTZLondon,
			Location:    "Madrid Barajas T4",
			Description: "Booking: ABC123",
			Categories:  []string{"Travel", "Flight"},
			Alarms:      []string{"-2h"},
		},
		{
			Summary: "Arrive London Heathrow", Start: "2030-12-20 11:00",
			StartTZ:     templateTZLondon,
			Location:    "Heathrow T5",
			Description: "Take Heathrow Express to Paddington",
			Categories:  []string{"Travel", "Transfer"},
		},
		{
			Summary: "Hotel check-in Hilton London", Start: "2030-12-20 15:00",
			StartTZ:     templateTZLondon,
			Location:    "Hilton London Paddington",
			Description: "Booking ref: HIL789",
			Categories:  []string{"Travel", "Accommodation"},
			Alarms:      []string{"-1h"},
		},
		{
			Summary: "Walking tour South Bank", Start: "2030-12-20 17:00", End: "2030-12-20 19:00",
			StartTZ:     templateTZLondon,
			Location:    "Waterloo Bridge",
			Description: "Comfortable shoes",
			Categories:  []string{"Travel", "Activity"},
			Alarms:      []string{"-30m"},
		},
	}
}
