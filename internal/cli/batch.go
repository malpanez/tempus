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

	startTZ, endTZ := resolveBatchTimezones(rec, fallbackTZ)
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

func resolveBatchTimezones(rec batchRecord, fallbackTZ string) (startTZ, endTZ string) {
	startTZ = strings.TrimSpace(FirstNonEmpty(rec.StartTZ, fallbackTZ))
	endTZ = strings.TrimSpace(rec.EndTZ)
	if endTZ == "" {
		endTZ = startTZ
	}
	return startTZ, endTZ
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
		event.RRule = strings.TrimSpace(rec.RRule)
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

func NewBatchTemplateCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template [type]",
		Short: "Generate a pre-filled template file for batch mode",
		Long: `Generate template files to quickly start creating events.

Available template types:
  basic             - Simple 3-event example (CSV)
  adhd-routine      - Daily ADHD routine with medication and focus blocks (CSV)
  medication        - Medication schedule with triple alarms (YAML)
  work-meetings     - Recurring team meetings (CSV)
  medical           - Healthcare appointments with prep reminders (CSV)
  travel            - Travel itinerary with flights and hotels (JSON)
  family            - Family calendar with mixed events (CSV)

Examples:
  tempus batch template basic -o my-events.csv
  tempus batch template adhd-routine -o routine.csv
  tempus batch template medication -o meds.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchTemplate(app, cmd, args)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (required)")
	_ = cmd.MarkFlagRequired("output")
	cmd.Flags().StringP("format", "f", "csv", "Template format: csv or yaml")

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

	if format != "csv" && format != "yaml" {
		return fmt.Errorf("--format must be 'csv' or 'yaml', got %q", format)
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
	switch templateType {
	case "basic":
		return getBasicTemplate(), nil
	case "adhd-routine":
		return getADHDRoutineTemplate(), nil
	case "medication", "meds":
		return getMedicationTemplate(), nil
	case "work-meetings", "work":
		return getWorkMeetingsTemplate(), nil
	case "medical", "health":
		return getMedicalTemplate(), nil
	case "travel":
		return getTravelTemplate(), nil
	case "family":
		return getFamilyTemplate(), nil
	case "school-event":
		if format == "yaml" {
			return getSchoolEventTemplateYAML(), nil
		}
		return getSchoolEventTemplateCSV(), nil
	case "recruiter-meeting":
		if format == "yaml" {
			return getRecruiterMeetingTemplateYAML(), nil
		}
		return getRecruiterMeetingTemplateCSV(), nil
	case "travel-day":
		if format == "yaml" {
			return getTravelDayTemplateYAML(), nil
		}
		return getTravelDayTemplateCSV(), nil
	default:
		return "", fmt.Errorf("unknown template type: %s\nAvailable: basic, adhd-routine, medication, work-meetings, medical, travel, family, school-event, recruiter-meeting, travel-day", templateType)
	}
}

func getSchoolEventTemplateCSV() string {
	return `summary,start_date,end_date,category,location,alarm,notes
School starts Q3,2025-09-01,,trimester,IES Cervantes,-1d,
Half-term holiday,2025-10-27,2025-10-31,vacation,,,"Emma and Leo - no school"
Parent-teacher meeting,2025-11-15T17:00,,activity,IES Cervantes,adhd-default,"Emma - bring report card"
Emma pickup 17:00,2025-09-02T17:00,,transport,IES Cervantes,single,"Gate 2"
End of year concert,2025-06-20T18:00,,activity,School auditorium,-2h,"Bring camera"
`
}

func getSchoolEventTemplateYAML() string {
	return `- summary: "School starts Q3"
  start_date: "2025-09-01"
  category: trimester
  location: "IES Cervantes"
  alarm: "-1d"

- summary: "Half-term holiday"
  start_date: "2025-10-27"
  end_date: "2025-10-31"
  category: vacation
  notes: "Emma and Leo - no school"

- summary: "Parent-teacher meeting"
  start_date: "2025-11-15T17:00"
  category: activity
  location: "IES Cervantes"
  alarm: adhd-default
  notes: "Emma - bring report card"

- summary: "Emma pickup 17:00"
  start_date: "2025-09-02T17:00"
  category: transport
  location: "IES Cervantes"
  alarm: single
  notes: "Gate 2"

- summary: "End of year concert"
  start_date: "2025-06-20T18:00"
  category: activity
  location: "School auditorium"
  alarm: "-2h"
  notes: "Bring camera"
`
}

func getRecruiterMeetingTemplateCSV() string {
	return `summary,start_date,time,duration,timezone,alarm,add_prep_time,company,role,recruiter,notes
Call with Sarah @ Acme Corp,2025-12-16,10:00,30m,Europe/Madrid,adhd-default,true,Acme Corp,Senior Developer,Sarah Jones,"LinkedIn: linkedin.com/in/sarah"
Technical interview @ StartupX,2025-12-18,15:00,1h,America/New_York,adhd-default,true,StartupX,Backend Engineer,Mike Chen,"Phone: +1-555-0123"
`
}

func getRecruiterMeetingTemplateYAML() string {
	return `- summary: "Call with Sarah @ Acme Corp"
  start_date: "2025-12-16"
  time: "10:00"
  duration: 30m
  timezone: Europe/Madrid
  alarm: adhd-default
  add_prep_time: true
  company: "Acme Corp"
  role: "Senior Developer"
  recruiter: "Sarah Jones"
  notes: "LinkedIn: linkedin.com/in/sarah"

- summary: "Technical interview @ StartupX"
  start_date: "2025-12-18"
  time: "15:00"
  duration: 1h
  timezone: America/New_York
  alarm: adhd-default
  add_prep_time: true
  company: "StartupX"
  role: "Backend Engineer"
  recruiter: "Mike Chen"
  notes: "Phone: +1-555-0123"
`
}

func getTravelDayTemplateCSV() string {
	return `summary,start_date,time,end_time,timezone,destination_timezone,category,location,add_prep_time,alarm,notes
MAD -> LHR BA456,2025-12-20,08:30,11:00,Europe/Madrid,Europe/London,flight,Madrid Barajas T4,true,-2h,"Booking: ABC123"
Arrive London Heathrow,2025-12-20,11:00,,Europe/London,,transfer,Heathrow T5,false,,"Take Heathrow Express to Paddington"
Hotel check-in Hilton London,2025-12-20,15:00,,Europe/London,,accommodation,Hilton London Paddington,false,-1h,"Booking ref: HIL789"
Walking tour South Bank,2025-12-20,17:00,19:00,Europe/London,,activity,Waterloo Bridge,false,-30m,"Comfortable shoes"
`
}

func getTravelDayTemplateYAML() string {
	return `- summary: "MAD -> LHR BA456"
  start_date: "2025-12-20"
  time: "08:30"
  end_time: "11:00"
  timezone: Europe/Madrid
  destination_timezone: Europe/London
  category: flight
  location: "Madrid Barajas T4"
  add_prep_time: true
  alarm: "-2h"
  notes: "Booking: ABC123"

- summary: "Arrive London Heathrow"
  start_date: "2025-12-20"
  time: "11:00"
  timezone: Europe/London
  category: transfer
  location: "Heathrow T5"
  add_prep_time: false
  notes: "Take Heathrow Express to Paddington"

- summary: "Hotel check-in Hilton London"
  start_date: "2025-12-20"
  time: "15:00"
  timezone: Europe/London
  category: accommodation
  location: "Hilton London Paddington"
  add_prep_time: false
  alarm: "-1h"
  notes: "Booking ref: HIL789"

- summary: "Walking tour South Bank"
  start_date: "2025-12-20"
  time: "17:00"
  end_time: "19:00"
  timezone: Europe/London
  category: activity
  location: "Waterloo Bridge"
  add_prep_time: false
  alarm: "-30m"
  notes: "Comfortable shoes"
`
}

func getBasicTemplate() string {
	return `summary,start,duration,start_tz,location,description,categories,alarms
Team Meeting,2025-12-16 10:00,1h,Europe/Madrid,Conference Room,Weekly sync,Work|Meeting,-15m
Lunch Break,2025-12-16 13:00,1h,Europe/Madrid,,,Break,
Doctor Appointment,2025-12-17 09:00,45m,Europe/Madrid,Medical Center,,Health,trigger=-1d;description=Confirm appointment||-2h
`
}

func getADHDRoutineTemplate() string {
	return `summary,start,duration,start_tz,location,rrule,categories,description,alarms
Morning Medication,2025-12-16 08:00,5m,Europe/Madrid,,FREQ=DAILY;COUNT=30,Health|Medication,Take morning meds with food,trigger=-5m||trigger=-1m||trigger=2025-12-16 08:00
Deep Focus Block,2025-12-16 09:00,2h,Europe/Madrid,,FREQ=WEEKLY;BYDAY=MO;TU;WE;TH;FR;COUNT=20,Work|Focus,NO meetings - deep work only,trigger=-10m;description=Prepare workspace and eliminate distractions||trigger=-1m||trigger=2025-12-16 10:30;description=Halfway - stay focused
Transition Buffer,2025-12-16 11:00,15m,Europe/Madrid,,FREQ=WEEKLY;BYDAY=MO;TU;WE;TH;FR;COUNT=20,Break|Transition,Stretch and reset before next task,trigger=-1m
Lunch + Walk,2025-12-16 13:00,1h,Europe/Madrid,,FREQ=DAILY;COUNT=30,Break|Health,Eat away from desk - go outside,trigger=-5m||trigger=2025-12-16 13:30;description=Time to walk
Evening Medication,2025-12-16 20:00,5m,Europe/Madrid,,FREQ=DAILY;COUNT=30,Health|Medication,Take evening meds,trigger=-5m||trigger=-1m||trigger=2025-12-16 20:00
Wind Down Routine,2025-12-16 22:00,30m,Europe/Madrid,,FREQ=DAILY;COUNT=30,Health|Sleep,No screens - prepare for sleep,trigger=-15m||trigger=-5m
`
}

func getMedicationTemplate() string {
	return `# Medication Schedule Template
# Triple alarms: 5min before, 1min before, exact time

- summary: Morning Meds - Methylphenidate 20mg
  start: "2025-12-16 08:00"
  duration: 5m
  start_tz: Europe/Madrid
  rrule: FREQ=DAILY;COUNT=30
  categories: [Health, Medication]
  description: Take with food. Don't skip.
  alarms:
    - trigger=-5m
    - trigger=-1m
    - trigger=2025-12-16 08:00

- summary: Afternoon Meds - Methylphenidate 10mg
  start: "2025-12-16 14:00"
  duration: 5m
  start_tz: Europe/Madrid
  rrule: FREQ=DAILY;COUNT=30
  categories: [Health, Medication]
  description: Booster dose
  alarms:
    - trigger=-5m
    - trigger=-1m
    - trigger=2025-12-16 14:00

- summary: Evening Meds - Melatonin 3mg
  start: "2025-12-16 21:00"
  duration: 5m
  start_tz: Europe/Madrid
  rrule: FREQ=DAILY;COUNT=30
  categories: [Health, Medication, Sleep]
  description: Take 1 hour before bed
  alarms:
    - trigger=-5m
    - trigger=-1m
    - trigger=2025-12-16 21:00
`
}

func getWorkMeetingsTemplate() string {
	return `summary,start,duration,start_tz,location,rrule,exdate,categories,description,alarms
Team Standup,2025-12-16 09:30,30m,Europe/Madrid,Video call - Zoom,FREQ=WEEKLY;BYDAY=MO;TU;WE;TH;FR;COUNT=20,,Work|Meeting,Daily sync with team,-5m
Weekly 1:1 with Manager,2025-12-16 14:00,45m,Europe/Madrid,Office - Meeting Room 3,FREQ=WEEKLY;BYDAY=MO;COUNT=12,2025-12-23 14:00|2025-12-30 14:00,Work|1on1,Discuss progress and blockers,trigger=-1d;description=Prepare agenda and questions||trigger=-15m
Sprint Planning,2025-12-17 10:00,2h,Europe/Madrid,Conference Room A,FREQ=WEEKLY;BYDAY=TU;COUNT=6,,Work|Meeting|Planning,Plan next 2-week sprint,trigger=-1h;description=Review backlog||trigger=-15m
Client Demo,2025-12-19 16:00,90m,Europe/Madrid,Video call - Google Meet,FREQ=WEEKLY;BYDAY=TH;COUNT=8,,Work|Client|Demo,Demo progress to stakeholders,trigger=-1d;description=Prepare demo script||trigger=-2h;description=Test demo environment||trigger=-30m
`
}

func getMedicalTemplate() string {
	return `summary,start,duration,start_tz,location,categories,description,alarms
Dentist - 6 Month Checkup,2025-12-20 10:00,30m,Europe/Madrid,Dental Clinic - Main Street,Health|Dental,Routine cleaning and checkup,trigger=-1d;description=Call to confirm appointment||trigger=-2h;description=Time to leave (30min drive)||trigger=-5m
Therapy Session,2025-12-18 17:00,1h,Europe/Madrid,Downtown Office - 3rd Floor,Health|Mental Health,Weekly therapy appointment,trigger=-1d;description=Think about topics to discuss||trigger=-30m;description=Prepare - bring notebook||trigger=-5m
General Practitioner Checkup,2026-01-10 09:00,45m,Europe/Madrid,Health Center - Room 12,Health|Checkup,Annual physical examination,trigger=-1w;description=Schedule blood work if needed||trigger=-1d;description=Confirm appointment||trigger=-2h;description=Leave now (traffic)||trigger=-15m
Lab Tests (Fasting Required),2025-12-22 08:00,15m,Europe/Madrid,Hospital Lab - Building C,Health|Tests,Blood work - MUST FAST,trigger=-1d;description=No food after 10pm tonight||trigger=-12h;description=Fasting period begins - no food||trigger=-1h;description=Drink water only||trigger=-15m
`
}

func getTravelTemplate() string {
	return `[
  {
    "summary": "Flight MAD → DUB",
    "start": "2025-12-25 08:30",
    "end": "2025-12-25 10:00",
    "start_tz": "Europe/Madrid",
    "end_tz": "Europe/Dublin",
    "location": "Madrid Barajas Airport - Terminal 1",
    "description": "Ryanair FR1234 - Gate closes 08:00. Confirmation: ABC123",
    "categories": ["Travel", "Flight"],
    "alarms": [
      "trigger=-1d,description=Check-in online opens",
      "trigger=-3h,description=Wake up and get ready",
      "trigger=2025-12-25 06:30,description=Leave for airport now (traffic)",
      "trigger=2025-12-25 07:45,description=Security checkpoint - gate closes at 08:00"
    ]
  },
  {
    "summary": "Hotel Check-in",
    "start": "2025-12-25 12:00",
    "duration": "30m",
    "start_tz": "Europe/Dublin",
    "location": "Dublin City Hotel - 123 O'Connell Street",
    "description": "Confirmation: XYZ789. Room 305. Check-in after 14:00.",
    "categories": ["Travel", "Accommodation"],
    "alarms": [
      "trigger=-1h,description=Head to hotel from airport"
    ]
  },
  {
    "summary": "Return Flight DUB → MAD",
    "start": "2025-12-28 18:30",
    "end": "2025-12-28 22:00",
    "start_tz": "Europe/Dublin",
    "end_tz": "Europe/Madrid",
    "location": "Dublin Airport - Terminal 2",
    "description": "Ryanair FR5678. Gate closes 18:00.",
    "categories": ["Travel", "Flight"],
    "alarms": [
      "trigger=-1d,description=Check-in online",
      "trigger=-4h,description=Leave hotel for airport (bus takes 45min)",
      "trigger=2025-12-28 17:45,description=Final boarding call"
    ]
  }
]
`
}

func getFamilyTemplate() string {
	return `summary,start,duration,start_tz,location,rrule,categories,description,alarms
School Drop-off,2025-12-16 08:15,15m,Europe/Madrid,Elementary School,FREQ=WEEKLY;BYDAY=MO;TU;WE;TH;FR;COUNT=20,Family|Kids,Drop kids at school,trigger=-30m;description=Kids breakfast and get ready||trigger=-10m;description=Leave now
Soccer Practice,2025-12-17 17:00,1h,Europe/Madrid,Sports Complex Field 3,FREQ=WEEKLY;BYDAY=TU;TH;COUNT=12,Family|Kids|Sports,Lucas soccer practice,trigger=-1h;description=Prepare soccer bag and snacks||trigger=-15m
Piano Lesson,2025-12-18 16:30,45m,Europe/Madrid,Music Academy,FREQ=WEEKLY;BYDAY=WE;COUNT=10,Family|Kids|Music,Emma piano lesson,trigger=-2h;description=Practice today||trigger=-30m
Pediatrician Checkup,2025-12-20 10:00,30m,Europe/Madrid,Pediatric Clinic,Family|Kids|Health,Annual checkup for both kids,trigger=-1d;description=Confirm appointment||trigger=-2h||trigger=-30m
Date Night,2025-12-21 20:00,2h,Europe/Madrid,Restaurant Downtown,Family|Personal,Dinner reservation - babysitter confirmed,trigger=-1d;description=Confirm babysitter||trigger=-4h;description=Start getting ready||trigger=-1h
`
}
