package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/cli"
	"tempus/internal/config"
	"tempus/internal/constants"
	"tempus/internal/i18n"
	"tempus/internal/parsing"
	"tempus/internal/prompts"
	tpl "tempus/internal/templates"
	"tempus/internal/testutil"
	tzpkg "tempus/internal/timezone"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	scanner *bufio.Scanner
)

func init() {
	scanner = bufio.NewScanner(os.Stdin)
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		printErr("%v\n", err)
		os.Exit(1)
	}
}

var (
	version = "dev"
	commit  = "unknown"
	date    = ""
)

func newRootCmd() *cobra.Command {
	app := &cli.App{Stdout: os.Stdout, Stderr: os.Stderr}

	cmd := &cobra.Command{
		Use:              "tempus",
		Short:            "A multilingual ICS calendar file generator",
		SilenceUsage:     true,
		PersistentPreRunE: cli.SetupPersistentPreRunE(app),
	}

	cmd.PersistentFlags().StringP("language", "l", "", "Language for output (es, en, ga, pt)")
	cmd.PersistentFlags().StringP("timezone", "t", "", "Default timezone")
	cmd.PersistentFlags().StringP("config", "c", "", "Config file path")

	cmd.AddCommand(
		cli.NewCreateCmd(app),
		cli.NewQuickCmd(app),
		cli.NewInitCmd(app),
		newBatchCmd(),
		newLintCmd(),
		cli.NewConfigCmd(app),
		cli.NewVersionCmd(app, version, commit, date),
		newTemplateCmd(),
		newLocaleCmd(),
		newTimezoneCmd(),
		newRRuleHelperCmd(),
	)

	return cmd
}

func newBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Create multiple ICS events from CSV, JSON, or YAML",
		RunE:  runBatch,
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

	cmd.AddCommand(newBatchTemplateCmd())

	return cmd
}

func runBatch(cmd *cobra.Command, _ []string) error {
	opts, err := parseBatchFlags(cmd)
	if err != nil {
		return err
	}

	records, _, err := loadBatchInput(opts)
	if err != nil {
		return err
	}

	cal, validationErrors, err := buildBatchCalendar(records, opts)
	if err != nil {
		return err
	}

	warnings := collectBatchWarnings(cal.Events, opts)

	if opts.dryRun {
		return handleDryRun(validationErrors, warnings, records, opts.input, opts.output)
	}

	return writeBatchOutput(cal, warnings, opts.output, len(records))
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

	opts.input = strings.TrimSpace(opts.input)
	if opts.input == "" {
		return nil, fmt.Errorf("--input is required")
	}

	return opts, nil
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

func buildBatchCalendar(records []batchRecord, opts *batchOptions) (*calendar.Calendar, []string, error) {
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
		ev, err := buildEventFromBatch(rec, opts.defaultTZ)
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
		prepEvents := generatePrepTimeEvents(cal.Events)
		for _, prepEv := range prepEvents {
			cal.AddEvent(prepEv)
		}
	}

	return cal, validationErrors, nil
}

func collectBatchWarnings(events []calendar.Event, opts *batchOptions) []string {
	var warnings []string

	if opts.checkConflicts || opts.dryRun {
		conflicts := detectEventConflicts(events)
		if len(conflicts) > 0 {
			warnings = append(warnings, fmt.Sprintf("⚠️  Found %d time conflict(s):", len(conflicts)))
			for _, conflict := range conflicts {
				warnings = append(warnings, fmt.Sprintf("  • %s", conflict))
			}
		}
	}

	if opts.maxEventsPerDay > 0 || opts.dryRun {
		overwhelmDays := detectOverwhelmDays(events, opts.maxEventsPerDay)
		if len(overwhelmDays) > 0 {
			warnings = append(warnings, "⚠️  Days with high event load:")
			for _, day := range overwhelmDays {
				warnings = append(warnings, fmt.Sprintf("  • %s", day))
			}
		}
	}

	return warnings
}

func handleDryRun(validationErrors, warnings []string, records []batchRecord, input, output string) error {
	if len(validationErrors) > 0 {
		printErr("Validation failed with %d error(s):\n", len(validationErrors))
		for _, errMsg := range validationErrors {
			fmt.Printf("  ❌ %s\n", errMsg)
		}
		return fmt.Errorf("validation failed")
	}

	printOK("✓ Validation passed: %d events ready to create\n", len(records))

	if len(warnings) > 0 {
		fmt.Printf("\n")
		for _, warning := range warnings {
			fmt.Println(warning)
		}
	}

	printDryRunSummary(records, input, output)
	return nil
}

func printDryRunSummary(records []batchRecord, input, output string) {
	fmt.Printf("\nEvent summary:\n")
	for i, rec := range records {
		summary := rec.Summary
		if summary == "" {
			summary = "(no summary)"
		}
		start := rec.Start
		if start == "" {
			start = "(no start)"
		}
		fmt.Printf("  %d. %s - %s\n", i+1, summary, start)
	}
	fmt.Printf("\nTo create the calendar file, run:\n")
	fmt.Printf("  tempus batch -i %s -o %s\n", input, output)
}

func writeBatchOutput(cal *calendar.Calendar, warnings []string, output string, eventCount int) error {
	if len(warnings) > 0 {
		fmt.Printf("\n")
		for _, warning := range warnings {
			fmt.Println(warning)
		}
		fmt.Printf("\n")
	}

	if err := ensureDirForFile(output); err != nil {
		return err
	}

	if err := os.WriteFile(output, []byte(cal.ToICS()), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", output, err)
	}

	printOK("Created: %s (%d events)\n", output, eventCount)
	return nil
}

func newLintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Validate ICS files for common issues",
		RunE:  runLint,
	}
	cmd.Flags().StringArray("file", []string{}, "ICS file(s) to lint (repeat flag for multiple files)")
	return cmd
}

func runLint(cmd *cobra.Command, _ []string) error {
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
		printOK("Lint passed: %s\n", path)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

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

var icsDurationRegex = regexp.MustCompile(`(?i)^P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

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
	defer f.Close()

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
		rec.AllDay = parseBoolish(csvValue(row, index, "all_day"))

		if ex := csvValue(row, index, "exdate"); ex != "" {
			rec.ExDates = splitDelimited(ex)
		}
		if cats := csvValue(row, index, "categories"); cats != "" {
			rec.Categories = splitDelimited(cats)
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

func loadBatchFromJSON(path string) ([]batchRecord, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	records := make([]batchRecord, 0, len(raw))
	for _, item := range raw {
		rec := batchRecord{
			Summary:     valueAsString(item["summary"]),
			Start:       valueAsString(item["start"]),
			End:         valueAsString(item["end"]),
			Duration:    valueAsString(item["duration"]),
			StartTZ:     valueAsString(item["start_tz"]),
			EndTZ:       valueAsString(item["end_tz"]),
			Location:    valueAsString(item["location"]),
			Description: valueAsString(item["description"]),
			RRule:       valueAsString(item["rrule"]),
			AllDay:      valueAsBool(item["all_day"]),
			ExDates:     valueAsStringSlice(item["exdate"]),
			Categories:  valueAsStringSlice(item["categories"]),
			Alarms:      valueAsAlarmSlice(item["alarms"]),
		}
		records = append(records, rec)
	}
	return records, nil
}

func loadBatchFromYAML(path string) ([]batchRecord, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}

	var raw []map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	records := make([]batchRecord, 0, len(raw))
	for _, item := range raw {
		rec := batchRecord{
			Summary:     valueAsString(item["summary"]),
			Start:       valueAsString(item["start"]),
			End:         valueAsString(item["end"]),
			Duration:    valueAsString(item["duration"]),
			StartTZ:     valueAsString(item["start_tz"]),
			EndTZ:       valueAsString(item["end_tz"]),
			Location:    valueAsString(item["location"]),
			Description: valueAsString(item["description"]),
			RRule:       valueAsString(item["rrule"]),
			AllDay:      valueAsBool(item["all_day"]),
			ExDates:     valueAsStringSlice(item["exdate"]),
			Categories:  valueAsStringSlice(item["categories"]),
			Alarms:      valueAsAlarmSlice(item["alarms"]),
		}
		records = append(records, rec)
	}
	return records, nil
}

func buildEventFromBatch(rec batchRecord, fallbackTZ string) (*calendar.Event, error) {
	summary, startStr, err := validateBatchRecord(rec)
	if err != nil {
		return nil, err
	}

	startTZ, endTZ := resolveBatchTimezones(rec, fallbackTZ)
	startTime, endTime, err := parseBatchTimes(rec, startStr, startTZ, endTZ, summary)
	if err != nil {
		return nil, err
	}

	summaryWithEmoji := addEmojiToSummary(summary, rec.Categories)
	event := calendar.NewEvent(summaryWithEmoji, startTime, endTime)
	configureBatchEvent(event, rec, startTZ, endTZ)

	return event, nil
}

func validateBatchRecord(rec batchRecord) (summary, startStr string, err error) {
	summary = normalizeAndSpellCheck(strings.TrimSpace(rec.Summary))
	if summary == "" {
		return "", "", fmt.Errorf("summary is required")
	}

	startStr = normalizeDateTimeInput(strings.TrimSpace(rec.Start))
	if startStr == "" {
		return "", "", fmt.Errorf("start is required")
	}

	return summary, startStr, nil
}

func resolveBatchTimezones(rec batchRecord, fallbackTZ string) (startTZ, endTZ string) {
	startTZ = strings.TrimSpace(firstNonEmpty(rec.StartTZ, fallbackTZ))
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
	startDateStr := extractDate(startStr)
	startTime, err = time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date %q: %w", startStr, err)
	}

	if strings.TrimSpace(endStr) == "" {
		endTime = startTime.AddDate(0, 0, 1)
	} else {
		endDateStr := extractDate(endStr)
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
	if looksLikeClock(startStr) {
		startStr = prependToday(startStr, startTZ)
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
		return parseBatchDurationEnd(rec.Duration, startTime)
	default:
		return startTime.Add(getSmartDefaultDuration(summary, startTime)), nil
	}
}

func parseBatchExplicitEnd(endStr string, startTime time.Time, endTZ, originalEnd string) (time.Time, error) {
	if looksLikeClock(endStr) {
		endStr = prependToday(endStr, endTZ)
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

func parseBatchDurationEnd(durStr string, startTime time.Time) (time.Time, error) {
	dur, err := calendar.ParseHumanDuration(durStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration %q: %v", durStr, err)
	}
	if dur <= 0 {
		return time.Time{}, fmt.Errorf(testutil.ErrMsgDurationGreaterThanZero)
	}
	return startTime.Add(dur), nil
}

func configureBatchEvent(event *calendar.Event, rec batchRecord, startTZ, endTZ string) {
	event.AllDay = rec.AllDay

	if startTZ != "" {
		event.SetStartTimezone(startTZ)
	}
	if endTZ != "" {
		event.SetEndTimezone(endTZ)
	} else if startTZ != "" {
		event.SetEndTimezone(startTZ)
	}

	event.Location = strings.TrimSpace(rec.Location)
	event.Description = strings.TrimSpace(rec.Description)

	if strings.TrimSpace(rec.RRule) != "" {
		event.RRule = strings.TrimSpace(rec.RRule)
	}

	addBatchCategories(event, rec.Categories)
	addBatchExDates(event, rec.ExDates, startTZ, rec.AllDay)
	addBatchAlarms(event, rec.Alarms, startTZ)
}

func addBatchCategories(event *calendar.Event, categories []string) {
	for _, cat := range categories {
		cat = strings.TrimSpace(cat)
		if cat != "" {
			validated := validateCategoryWithSuggestion(cat)
			event.AddCategory(validated)
		}
	}
}

func addBatchExDates(event *calendar.Event, exdates []string, startTZ string, allDay bool) {
	if len(exdates) == 0 {
		return
	}

	tzForEx := event.StartTZ
	if tzForEx == "" {
		tzForEx = startTZ
	}

	parsed, err := parseExDateValues(exdates, tzForEx, allDay)
	if err == nil {
		event.ExDates = append(event.ExDates, parsed...)
	}
}

func addBatchAlarms(event *calendar.Event, alarms []string, startTZ string) {
	if len(alarms) == 0 {
		return
	}

	defaultAlarmTZ := event.StartTZ
	if defaultAlarmTZ == "" {
		defaultAlarmTZ = startTZ
	}

	expandedAlarms, err := expandAlarmProfiles(alarms)
	if err != nil {
		return
	}
	parsed, err := calendar.ParseAlarmSpecs(expandedAlarms, defaultAlarmTZ)
	if err == nil {
		event.Alarms = append(event.Alarms, parsed...)
	}
}

func normalizeAndSpellCheck(text string) string {
	return cli.NormalizeAndSpellCheck(text)
}

func normalizeDateTimeInput(input string) string {
	return parsing.NormalizeDateTimeInput(input)
}

func validateCategoryWithSuggestion(category string) string {
	return cli.ValidateCategoryWithSuggestion(category)
}

func addEmojiToSummary(summary string, categories []string) string {
	return cli.AddEmojiToSummary(summary, categories)
}

func getSmartDefaultDuration(summary string, startTime time.Time) time.Duration {
	return cli.GetSmartDefaultDuration(summary, startTime)
}

func detectEventConflicts(events []calendar.Event) []string {
	return cli.DetectEventConflicts(events)
}

func generatePrepTimeEvents(events []calendar.Event) []*calendar.Event {
	return cli.GeneratePrepTimeEvents(events)
}

func stripEmoji(s string) string {
	return cli.StripEmoji(s)
}

func generateUID() string {
	return cli.GenerateUID()
}

func detectOverwhelmDays(events []calendar.Event, maxPerDay int) []string {
	return cli.DetectOverwhelmDays(events, maxPerDay)
}

func expandAlarmProfiles(alarmSpecs []string) ([]string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config for alarm profiles: %w", err)
	}

	expanded := make([]string, 0, len(alarmSpecs))
	for _, spec := range alarmSpecs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		if strings.HasPrefix(spec, "profile:") {
			profileName := strings.TrimPrefix(spec, "profile:")
			profileName = strings.TrimSpace(profileName)

			profile := cfg.GetAlarmProfile(profileName)
			if profile != nil {
				expanded = append(expanded, profile...)
			} else {
				names := cfg.ListAlarmProfiles()
				return nil, fmt.Errorf("profile '%s' not found. Available: %s", profileName, strings.Join(names, ", "))
			}
		} else {
			expanded = append(expanded, spec)
		}
	}

	return expanded, nil
}

// ========================================================================
// Batch Template Generator
// ========================================================================

func newBatchTemplateCmd() *cobra.Command {
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
		RunE: runBatchTemplate,
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (required)")
	_ = cmd.MarkFlagRequired("output")
	cmd.Flags().StringP("format", "f", "csv", "Template format: csv or yaml")

	return cmd
}

func runBatchTemplate(cmd *cobra.Command, args []string) error {
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

	printOK("Template created: %s\n", output)
	fmt.Printf("Edit the file and run: tempus batch -i %s -o calendar.ics\n", output)

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
		// nothing
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

func ensureDirForFile(path string) error {
	return cli.EnsureDirForFile(path)
}

func extractDate(s string) string {
	return cli.ExtractDate(s)
}

func parseBoolish(s string) bool {
	return cli.ParseBoolish(s)
}

func splitDelimited(s string) []string {
	return cli.SplitDelimited(s)
}

func valueAsString(v interface{}) string {
	return cli.ValueAsString(v)
}

func valueAsBool(v interface{}) bool {
	return cli.ValueAsBool(v)
}

func valueAsStringSlice(v interface{}) []string {
	return cli.ValueAsStringSlice(v)
}

func valueAsAlarmSlice(v interface{}) []string {
	return cli.ValueAsAlarmSlice(v)
}

// ========================================================================
// RRULE Helper Command
// ========================================================================

func newRRuleHelperCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rrule",
		Short: "Interactive helper to build recurrence rules (RRULE)",
		Long: `Generate RRULE strings for recurring events without memorizing the syntax.

Examples of what you can create:
  - Every weekday (Monday-Friday)
  - Every 2 weeks on Tuesday and Thursday
  - Monthly on the 15th
  - Yearly on March 1st
  - Custom patterns with end dates or occurrence counts`,
		RunE: runRRuleHelper,
	}
}

func runRRuleHelper(_ *cobra.Command, _ []string) error {
	fmt.Println("RRULE Builder - Create recurring event patterns")
	fmt.Println()

	freq, err := promptRRuleFrequency()
	if err != nil {
		return err
	}

	parts := []string{fmt.Sprintf("FREQ=%s", freq)}

	if interval := promptRRuleInterval(); interval != "" {
		parts = append(parts, interval)
	}

	if freq == "WEEKLY" {
		if days := promptRRuleWeeklyDays(); days != "" {
			parts = append(parts, days)
		}
	}

	if endCond := promptRRuleEndCondition(); endCond != "" {
		parts = append(parts, endCond)
	}

	rrule := strings.Join(parts, ";")

	fmt.Println()
	printOK("Generated RRULE:\n")
	fmt.Println(rrule)
	fmt.Println()

	// Show examples
	fmt.Println("Usage examples:")
	fmt.Printf("  CSV batch file:  rrule column = %s\n", rrule)
	fmt.Printf("  JSON batch file: \"rrule\": \"%s\"\n", rrule)
	fmt.Printf("  YAML batch file: rrule: %s\n", rrule)
	fmt.Println()

	// Show human-readable interpretation
	fmt.Println("This means:")
	fmt.Printf("  %s\n", interpretRRule(rrule))

	return nil
}

func interpretRRule(rrule string) string {
	parts := strings.Split(rrule, ";")
	var freq, interval, byday, count, until string

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, val := kv[0], kv[1]
		switch key {
		case "FREQ":
			freq = strings.ToLower(val)
		case "INTERVAL":
			interval = val
		case "BYDAY":
			byday = val
		case "COUNT":
			count = val
		case "UNTIL":
			until = val
		}
	}

	var result string

	// Build frequency part
	if interval != "" && interval != "1" {
		result = fmt.Sprintf("Every %s %ss", interval, freq)
	} else {
		result = fmt.Sprintf("Every %s", freq)
	}

	// Add day specification
	if byday != "" {
		result += fmt.Sprintf(" on %s", byday)
	}

	// Add end condition
	if count != "" {
		result += fmt.Sprintf(", %s times", count)
	} else if until != "" {
		result += fmt.Sprintf(", until %s", until)
	} else {
		result += ", forever"
	}

	return result
}

func promptRRuleFrequency() (string, error) {
	fmt.Println("Select frequency:")
	fmt.Println("  1. Daily")
	fmt.Println("  2. Weekly")
	fmt.Println("  3. Monthly")
	fmt.Println("  4. Yearly")
	fmt.Print("Enter choice (1-4): ")

	var freqChoice int
	if _, err := fmt.Scanln(&freqChoice); err != nil || freqChoice < 1 || freqChoice > 4 {
		return "", fmt.Errorf("invalid choice")
	}

	frequencies := map[int]string{1: "DAILY", 2: "WEEKLY", 3: "MONTHLY", 4: "YEARLY"}
	return frequencies[freqChoice], nil
}

func promptRRuleInterval() string {
	fmt.Print("\nRepeat every N occurrences (default 1): ")
	var intervalStr string
	_, _ = fmt.Scanln(&intervalStr)
	intervalStr = strings.TrimSpace(intervalStr)
	if intervalStr != "" && intervalStr != "1" {
		interval := atoiSafe(intervalStr)
		if interval > 0 {
			return fmt.Sprintf("INTERVAL=%d", interval)
		}
	}
	return ""
}

func promptRRuleWeeklyDays() string {
	fmt.Println("\nSelect days of week (comma-separated):")
	fmt.Println("  MO, TU, WE, TH, FR, SA, SU")
	fmt.Print("Days (e.g., 'MO,WE,FR' or leave empty for all): ")
	var daysStr string
	_, _ = fmt.Scanln(&daysStr)
	daysStr = strings.TrimSpace(daysStr)
	if daysStr != "" {
		return fmt.Sprintf("BYDAY=%s", strings.ToUpper(daysStr))
	}
	return ""
}

func promptRRuleEndCondition() string {
	fmt.Println("\nHow should the recurrence end?")
	fmt.Println("  1. Never (infinite)")
	fmt.Println("  2. After N occurrences")
	fmt.Println("  3. On a specific date")
	fmt.Print("Enter choice (1-3): ")

	var endChoice int
	if _, err := fmt.Scanln(&endChoice); err != nil || endChoice < 1 || endChoice > 3 {
		endChoice = 1
	}

	switch endChoice {
	case 2:
		fmt.Print("Number of occurrences: ")
		var countStr string
		_, _ = fmt.Scanln(&countStr)
		count := atoiSafe(strings.TrimSpace(countStr))
		if count > 0 {
			return fmt.Sprintf("COUNT=%d", count)
		}
	case 3:
		fmt.Print("End date (YYYY-MM-DD): ")
		var untilStr string
		_, _ = fmt.Scanln(&untilStr)
		untilStr = strings.TrimSpace(untilStr)
		if untilStr != "" {
			if _, err := time.Parse("2006-01-02", untilStr); err == nil {
				untilStr = strings.ReplaceAll(untilStr, "-", "")
				return fmt.Sprintf("UNTIL=%s", untilStr)
			}
		}
	}
	return ""
}

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage event templates",
	}

	cmd.PersistentFlags().String("templates-dir", "", "Directory with JSON templates (default: user config dir)")

	createCmd := &cobra.Command{
		Use:   "create <template-name>",
		Short: "Create event from template",
		Args:  cobra.ExactArgs(1),
		RunE:  runTemplateCreate,
	}
	createCmd.Flags().String("output-dir", "", "Directory where generated ICS files will be stored")
	createCmd.Flags().String("input", "", "CSV or JSON file with template data (creates one ICS per row)")
	createCmd.Flags().String("format", "auto", "Input format: auto, csv, or json")
	createCmd.Flags().String("templates-dir", "", "Directory with JSON templates (overrides defaults)")

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available templates",
			RunE:  runTemplateList,
		},
		createCmd,
		&cobra.Command{
			Use:   "describe <template-name>",
			Short: "Show details for a template",
			Args:  cobra.ExactArgs(1),
			RunE:  runTemplateDescribe,
		},
		&cobra.Command{
			Use:   "validate",
			Short: "Validate data-driven templates",
			RunE:  runTemplateValidate,
		},
		newTemplateInitCmd(),
	)

	return cmd
}

func newTemplateInitCmd() *cobra.Command {
	defaultDir := "."
	if dirs := tpl.DefaultTemplateDirs(); len(dirs) > 0 {
		defaultDir = dirs[0]
	}

	cmd := &cobra.Command{
		Use:   "init <template-name>",
		Short: "Generate a template scaffold file",
		Args:  cobra.ExactArgs(1),
		RunE:  runTemplateInit,
	}
	cmd.Flags().String("dir", defaultDir, "Directory where the scaffold will be written")
	cmd.Flags().String("format", "yaml", "Output format (yaml or json)")
	cmd.Flags().String("lang", "en", "Language for sample labels (en, es, pt)")
	cmd.Flags().Bool("force", false, "Overwrite the file if it already exists")
	return cmd
}

func runTemplateList(cmd *cobra.Command, _ []string) error {
	tm, _, err := loadTemplateManager(cmd)
	if err != nil {
		return err
	}

	fmt.Println("Available templates:")
	all := tm.ListTemplates()
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t := all[name]
		desc := t.Description
		if desc == "" {
			desc = "-"
		}
		fmt.Printf("  %-12s  %s\n", name, desc)
	}
	return nil
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	tm, tr, err := loadTemplateManager(cmd)
	if err != nil {
		return err
	}

	tmpl, err := tm.GetTemplate(name)
	if err != nil {
		return err
	}

	outputDir, _ := cmd.Flags().GetString("output-dir")
	inputPath, _ := cmd.Flags().GetString("input")
	formatFlag, _ := cmd.Flags().GetString("format")

	dd, _ := tm.DataTemplate(name)

	if strings.TrimSpace(inputPath) != "" {
		params := templateCreateParams{
			templateName: name,
			inputPath:    inputPath,
			formatFlag:   formatFlag,
			outputDir:    outputDir,
		}
		return runTemplateCreateFromFile(tm, tr, tmpl, dd, params)
	}

	values := map[string]string{}
	for _, f := range tmpl.Fields {
		if isAlarmField(f) {
			values[f.Key] = promptAlarmField(labelForField(f), f.Default)
			continue
		}
		v := promptInput(labelForField(f), f.Default)
		if f.Required && strings.TrimSpace(v) == "" {
			return fmt.Errorf("field %q is required", f.Key)
		}
		values[f.Key] = v
	}

	normalizeValuesForTemplate(values, tmpl, dd)

	ev, err := tm.GenerateEvent(name, values, tr)
	if err != nil {
		return err
	}

	cal := buildTemplateCalendar(ev)

	augmented := augmentValuesForFilename(values, ev)
	defaultName := deriveTemplateFilename(tm, name, augmented, ev, tr)
	userOutput := promptInput("Output filename", defaultName)
	finalName := strings.TrimSpace(userOutput)
	if finalName == "" {
		finalName = defaultName
	}
	finalName = ensureICSExtension(finalName)
	if dir := strings.TrimSpace(outputDir); dir != "" && !filepath.IsAbs(finalName) {
		finalName = filepath.Join(dir, finalName)
	}
	finalName = ensureUniquePath(finalName)

	if err := ensureDirForFile(finalName); err != nil {
		return err
	}
	if err := os.WriteFile(finalName, []byte(cal.ToICS()), 0600); err != nil {
		printErr(constants.ErrMsgFailedToWriteFile, err)
		return err
	}
	printOK("Created: %s\n", finalName)
	return nil
}

type templateCreateParams struct {
	templateName string
	inputPath    string
	formatFlag   string
	outputDir    string
}

func runTemplateCreateFromFile(tm *tpl.TemplateManager, tr *i18n.Translator, tmpl *tpl.Template, dd tpl.DataDrivenTemplate, params templateCreateParams) error {
	format, err := detectTemplateInputFormat(params.formatFlag, params.inputPath)
	if err != nil {
		return err
	}

	records, err := loadTemplateRecords(params.inputPath, format)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return fmt.Errorf("no data found in %s", params.inputPath)
	}

	params.outputDir = strings.TrimSpace(params.outputDir)
	for idx, record := range records {
		values := mergeTemplateValues(tmpl, record)
		normalizeValuesForTemplate(values, tmpl, dd)

		ev, err := tm.GenerateEvent(params.templateName, values, tr)
		if err != nil {
			return fmt.Errorf(testutil.ErrMsgRowFormat, idx+1, err)
		}

		cal := buildTemplateCalendar(ev)
		augmented := augmentValuesForFilename(values, ev)
		filename := deriveTemplateFilename(tm, params.templateName, augmented, ev, tr)
		filename = ensureICSExtension(filename)
		if params.outputDir != "" && !filepath.IsAbs(filename) {
			filename = filepath.Join(params.outputDir, filename)
		}
		filename = ensureUniquePath(filename)

		if err := ensureDirForFile(filename); err != nil {
			return fmt.Errorf(testutil.ErrMsgRowFormat, idx+1, err)
		}
		if err := os.WriteFile(filename, []byte(cal.ToICS()), 0600); err != nil {
			return fmt.Errorf("row %d: failed to write file: %w", idx+1, err)
		}
		printOK("Created: %s\n", filename)
	}

	return nil
}

func runTemplateDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	tm, _, err := loadTemplateManager(cmd)
	if err != nil {
		return err
	}

	tmpl, err := tm.GetTemplate(name)
	if err != nil {
		return err
	}

	printTemplateBasicInfo(tmpl)
	dd := printTemplateTypeInfo(tm, name)
	printTemplateFields(tmpl.Fields)
	printTemplateOutput(dd)
	return nil
}

func printTemplateBasicInfo(tmpl *tpl.Template) {
	fmt.Printf("Name: %s\n", tmpl.Name)
	if desc := strings.TrimSpace(tmpl.Description); desc != "" {
		fmt.Printf("Description: %s\n", desc)
	} else {
		fmt.Println("Description: -")
	}
}

func printTemplateTypeInfo(tm *tpl.TemplateManager, name string) tpl.DataDrivenTemplate {
	var dd tpl.DataDrivenTemplate
	if ddt, ok := tm.DataTemplate(name); ok {
		fmt.Printf("Type: data-driven (schema v%d)\n", ddt.SchemaVersion)
		if strings.TrimSpace(ddt.Source) != "" {
			fmt.Printf("Source: %s\n", ddt.Source)
		} else {
			fmt.Println("Source: embedded")
		}
		if strings.TrimSpace(ddt.FilenameTemplate) != "" {
			fmt.Printf("Filename template: %s\n", ddt.FilenameTemplate)
		}
		dd = ddt
	} else {
		fmt.Println("Type: built-in (compiled)")
	}
	return dd
}

func printTemplateFields(fields []tpl.Field) {
	fmt.Println("Fields:")
	for _, field := range fields {
		required := "optional"
		if field.Required {
			required = "required"
		}
		line := fmt.Sprintf("  - %s (%s, %s)", field.Key, field.Type, required)
		if field.Default != "" {
			line += fmt.Sprintf(", default=%q", field.Default)
		}
		if strings.TrimSpace(field.Description) != "" {
			line += fmt.Sprintf(" — %s", field.Description)
		}
		fmt.Println(line)
	}
}

func printTemplateOutput(dd tpl.DataDrivenTemplate) {
	if dd.Name == "" {
		return
	}
	fmt.Println("Output:")
	fmt.Printf("  start_field: %s\n", dd.Output.StartField)
	printIfNotEmpty("  end_field: %s\n", dd.Output.EndField)
	printIfNotEmpty("  duration_field: %s\n", dd.Output.DurationField)
	printIfNotEmpty("  start_tz_field: %s\n", dd.Output.StartTZField)
	if strings.TrimSpace(dd.Output.EndTZField) != "" && dd.Output.EndTZField != dd.Output.StartTZField {
		fmt.Printf("  end_tz_field: %s\n", dd.Output.EndTZField)
	}
	printIfNotEmpty("  summary_tmpl: %s\n", dd.Output.SummaryTmpl)
	printIfNotEmpty("  location_tmpl: %s\n", dd.Output.LocationTmpl)
	printIfNotEmpty("  description_tmpl: %s\n", dd.Output.DescriptionTmpl)
	if len(dd.Output.Categories) > 0 {
		fmt.Printf("  categories: %s\n", strings.Join(dd.Output.Categories, ", "))
	}
	if dd.Output.Priority > 0 {
		fmt.Printf("  priority: %d\n", dd.Output.Priority)
	}
}

func printIfNotEmpty(format, value string) {
	if strings.TrimSpace(value) != "" {
		fmt.Printf(format, value)
	}
}

func detectTemplateInputFormat(flag, path string) (string, error) {
	flag = strings.ToLower(strings.TrimSpace(flag))
	switch flag {
	case "", "auto":
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".csv":
			return "csv", nil
		case ".json":
			return "json", nil
		default:
			return "", fmt.Errorf("cannot infer format from %s; use --format csv|json", path)
		}
	case "csv", "json":
		return flag, nil
	default:
		return "", fmt.Errorf("unsupported format %q (use csv or json)", flag)
	}
}

func loadTemplateRecords(path, format string) ([]map[string]string, error) {
	switch format {
	case "csv":
		return loadTemplateFromCSV(path)
	case "json":
		return loadTemplateFromJSON(path)
	default:
		return nil, fmt.Errorf("unknown format %q", format)
	}
}

func loadTemplateFromCSV(path string) ([]map[string]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}

	var records []map[string]string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		record := make(map[string]string, len(header))
		empty := true
		for i, key := range header {
			if key == "" {
				continue
			}
			value := ""
			if i < len(row) {
				value = strings.TrimSpace(row[i])
			}
			if value != "" {
				empty = false
			}
			record[key] = value
		}
		if empty {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

func loadTemplateFromJSON(path string) ([]map[string]string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	records := make([]map[string]string, 0, len(raw))
	for _, item := range raw {
		record := make(map[string]string, len(item))
		empty := true
		for k, v := range item {
			value := strings.TrimSpace(fmt.Sprintf("%v", v))
			if value != "" {
				empty = false
			}
			record[strings.TrimSpace(k)] = value
		}
		if empty {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

func mergeTemplateValues(tmpl *tpl.Template, record map[string]string) map[string]string {
	values := make(map[string]string, len(record)+len(tmpl.Fields))
	for _, f := range tmpl.Fields {
		raw := strings.TrimSpace(record[f.Key])
		if raw == "" {
			values[f.Key] = f.Default
		} else {
			values[f.Key] = raw
		}
	}
	for k, v := range record {
		if _, exists := values[k]; !exists {
			values[k] = strings.TrimSpace(v)
		}
	}
	return values
}

func templateFieldDefault(tmpl *tpl.Template, key string) string {
	for _, f := range tmpl.Fields {
		if f.Key == key {
			return f.Default
		}
	}
	return ""
}

func normalizeValuesForTemplate(values map[string]string, tmpl *tpl.Template, dd tpl.DataDrivenTemplate) {
	if strings.TrimSpace(dd.Name) == "" {
		durationDefault := templateFieldDefault(tmpl, "duration")
		normalizeClockOnlyDateTimes(values, "start_time", "end_time", "timezone")
		normalizeEndTimeFromDuration(values, "start_time", "end_time", "duration", "timezone", firstNonEmpty(values["duration"], durationDefault, "30m"))
		return
	}

	startField := strings.TrimSpace(dd.Output.StartField)
	endField := strings.TrimSpace(dd.Output.EndField)
	durationField := strings.TrimSpace(dd.Output.DurationField)
	tzField := strings.TrimSpace(dd.Output.StartTZField)
	if tzField == "" {
		tzField = strings.TrimSpace(dd.Output.EndTZField)
	}

	normalizeClockOnlyDateTimes(values, startField, endField, tzField)
	var durationDefault string
	if durationField != "" {
		durationDefault = firstNonEmpty(values[durationField], templateFieldDefault(tmpl, durationField))
	}
	normalizeEndTimeFromDuration(values, startField, endField, durationField, tzField, durationDefault)
}

func buildTemplateCalendar(ev *calendar.Event) *calendar.Calendar {
	cal := calendar.NewCalendar()
	cal.IncludeVTZ = true
	cal.AddEvent(ev)
	cal.Name = ev.Summary
	if tz := firstNonEmpty(ev.StartTZ, ev.EndTZ); strings.TrimSpace(tz) != "" {
		cal.SetDefaultTimezone(tz)
	}
	return cal
}

func augmentValuesForFilename(values map[string]string, ev *calendar.Event) map[string]string {
	out := make(map[string]string, len(values)+2)
	for k, v := range values {
		out[k] = v
	}
	if !ev.StartTime.IsZero() {
		out["start_date"] = ev.StartTime.Format("2006-01-02")
		out["start_time_iso"] = ev.StartTime.Format("2006-01-02 15:04")
	}
	if !ev.EndTime.IsZero() {
		out["end_date"] = ev.EndTime.Format("2006-01-02")
		out["end_time_iso"] = ev.EndTime.Format("2006-01-02 15:04")
	}
	return out
}

func deriveTemplateFilename(tm *tpl.TemplateManager, templateName string, values map[string]string, ev *calendar.Event, tr *i18n.Translator) string {
	if ftmpl, ok := tm.FilenameTemplate(templateName); ok {
		if out, err := tpl.RenderTmpl(ftmpl, values, tr); err == nil {
			if cleaned := strings.TrimSpace(out); cleaned != "" {
				return cleaned
			}
		}
	}

	if !ev.StartTime.IsZero() {
		base := slugify(ev.Summary)
		if base == "" {
			base = slugify(templateName)
		}
		return fmt.Sprintf("%s-%s.ics", base, ev.StartTime.Format("2006-01-02"))
	}
	if sdate, ok := tm.GuessStartDate(templateName, values); ok && strings.TrimSpace(sdate) != "" {
		return fmt.Sprintf("%s-%s.ics", slugify(templateName), sdate)
	}
	return fmt.Sprintf("%s.ics", slugify(templateName))
}

func ensureICSExtension(name string) string {
	return cli.EnsureICSExtension(name)
}

func ensureUniquePath(path string) string {
	return cli.EnsureUniquePath(path)
}

func runTemplateValidate(cmd *cobra.Command, _ []string) error {
	templatesDirFlag, _ := cmd.Flags().GetString("templates-dir")
	dirs := tpl.ResolveTemplateDirs(templatesDirFlag)
	if len(dirs) == 0 {
		fmt.Println("No template directories configured.")
		return nil
	}

	fmt.Println("Validating template directories:")
	var validationErr bool
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf(" - %s: not found (skipped)\n", dir)
				continue
			}
			fmt.Printf(" - %s: error: %v\n", dir, err)
			validationErr = true
			continue
		}
		if !info.IsDir() {
			fmt.Printf(" - %s: not a directory (skipped)\n", dir)
			continue
		}

		defs, err := tpl.LoadDDTemplates(dir)
		if err != nil {
			fmt.Printf(" - %s: error: %v\n", dir, err)
			validationErr = true
			continue
		}
		names := make([]string, 0, len(defs))
		for name := range defs {
			names = append(names, name)
		}
		sort.Strings(names)
		if len(names) == 0 {
			fmt.Printf(" - %s: ok (no templates found)\n", dir)
			continue
		}
		fmt.Printf(" - %s: ok (%d template(s): %s)\n", dir, len(names), strings.Join(names, ", "))
	}

	if validationErr {
		return fmt.Errorf("template validation failed")
	}
	fmt.Println("Validation completed successfully.")
	return nil
}

func runTemplateInit(cmd *cobra.Command, args []string) error {
	name := args[0]
	dir, _ := cmd.Flags().GetString("dir")
	format, _ := cmd.Flags().GetString("format")
	lang, _ := cmd.Flags().GetString("lang")
	force, _ := cmd.Flags().GetBool("force")

	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("directory cannot be empty")
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "yaml"
	}
	ext := ".yaml"
	if format == "json" {
		ext = ".json"
	} else if format != "yaml" {
		return fmt.Errorf("unsupported format %q (use yaml or json)", format)
	}

	filename := filepath.Join(dir, slugify(name)+ext)
	if !force {
		if _, err := os.Stat(filename); err == nil {
			return fmt.Errorf("file %s already exists (use --force to overwrite)", filename)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to stat %s: %w", filename, err)
		}
	}

	content, err := tpl.GenerateScaffold(tpl.ScaffoldOptions{
		Name:     name,
		Language: lang,
		Format:   format,
	})
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, content, 0o600); err != nil {
		return fmt.Errorf("failed to write scaffold: %w", err)
	}

	printOK("Created scaffold: %s\n", filename)
	return nil
}

func newLocaleCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "locale",
		Short: "Inspect available locales",
	}

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available locales",
		RunE:  runLocaleList,
	})

	return root
}

func runLocaleList(_ *cobra.Command, _ []string) error {
	locales := i18n.Locales()
	if len(locales) == 0 {
		fmt.Println("No locales found.")
		return nil
	}
	fmt.Printf("%-8s %-9s %s\n", "Code", "Embedded", "Disk Paths")
	for _, loc := range locales {
		embedded := "no"
		if loc.Embedded {
			embedded = "yes"
		}
		paths := "-"
		if len(loc.DiskPaths) > 0 {
			paths = strings.Join(loc.DiskPaths, ", ")
		}
		fmt.Printf("%-8s %-9s %s\n", loc.Code, embedded, paths)
	}
	return nil
}

// ---------- helpers ----------

func loadTemplateManager(cmd *cobra.Command) (*tpl.TemplateManager, *i18n.Translator, error) {
	cfg, _ := config.Load() // proceed with defaults if it fails
	langFlag, _ := cmd.Root().Flags().GetString("language")
	templatesDirFlag, _ := cmd.Flags().GetString("templates-dir")

	cfgLang := ""
	if cfg != nil {
		if v, err := cfg.Get("language"); err == nil {
			cfgLang = v
		}
	}
	lang := firstNonEmpty(langFlag, cfgLang, "en")

	tr, err := newTranslator(lang)
	if err != nil {
		return nil, nil, err
	}

	tm := tpl.NewTemplateManager()

	// Load external JSON templates (optional dirs). NOTE: LoadDDDir returns no values; just call it.
	for _, dir := range tpl.ResolveTemplateDirs(templatesDirFlag) {
		tm.LoadDDDir(dir)
	}

	return tm, tr, nil
}

// Build translator with graceful fallback to "en"
func newTranslator(lang string) (*i18n.Translator, error) {
	tr, err := i18n.NewTranslator(lang)
	if err == nil {
		return tr, nil
	}
	if lang != "en" {
		if fallback, err2 := i18n.NewTranslator("en"); err2 == nil {
			return fallback, nil
		}
	}
	return nil, err
}

func firstNonEmpty(vals ...string) string {
	return cli.FirstNonEmpty(vals...)
}

func labelForField(f tpl.Field) string {
	label := f.Name
	if label == "" {
		label = f.Key
	}
	if f.Required {
		return label + " *"
	}
	return label
}

func isAlarmField(f tpl.Field) bool {
	if strings.EqualFold(f.Key, "alarms") {
		return true
	}
	if strings.EqualFold(f.Type, "alarms") {
		return true
	}
	return false
}

func slugify(s string) string {
	return cli.Slugify(s)
}

func promptInput(prompt, defaultValue string) string {
	return prompts.Input(prompt, defaultValue)
}

func promptAlarmField(label, defaultValue string) string {
	fmt.Printf("\n%s\n", label)
	existing := calendar.SplitAlarmInput(defaultValue)
	if len(existing) > 0 {
		fmt.Println("Recordatorios sugeridos:")
		for i, spec := range existing {
			fmt.Printf("  %d) %s\n", i+1, spec)
		}
		keep := strings.ToLower(strings.TrimSpace(promptInput("Pulsa Enter para mantenerlos o escribe 'n' para cambiarlos", "")))
		if keep == "" || keep == "s" || keep == "si" {
			return strings.Join(existing, "\n")
		}
		fmt.Println("")
	}

	fmt.Println("Añade hasta 4 recordatorios. Usa formatos como -15m, +10m, 2025-03-01 09:15 o trigger=-15m,description=Texto.")
	fmt.Println("Escribe '?' para ver ejemplos o deja vacío para terminar.")

	specs := make([]string, 0, 4)
	for len(specs) < 4 {
		prompt := fmt.Sprintf("Recordatorio #%d (-15m, +10m, trigger=..., ? para ayuda)", len(specs)+1)
		input := strings.TrimSpace(promptInput(prompt, ""))
		if input == "" {
			break
		}
		if input == "?" {
			fmt.Println("Ejemplos:")
			fmt.Println("  -15m                 -> 15 minutos antes")
			fmt.Println("  +5m                  -> 5 minutos después")
			fmt.Println("  trigger=-30m,description=Buscar taxi")
			fmt.Println("  trigger=2025-03-01 09:15,description=Check-in")
			continue
		}

		spec := input
		if !strings.Contains(spec, "=") {
			desc := strings.TrimSpace(promptInput("Descripción opcional (Enter para usar la genérica)", ""))
			if desc != "" {
				spec = fmt.Sprintf("trigger=%s,description=%s", input, desc)
			}
		}

		if _, err := calendar.ParseAlarmSpecs([]string{spec}, ""); err != nil {
			fmt.Printf("? %v\n", err)
			continue
		}
		specs = append(specs, spec)
	}

	if len(specs) == 0 {
		return ""
	}
	return strings.Join(specs, "\n")
}

// ------------------------------
// ND-friendly normalization
// ------------------------------

func looksLikeClock(s string) bool {
	return cli.LooksLikeClock(s)
}

func prependToday(clock, tz string) string {
	return cli.PrependToday(clock, tz)
}

func normalizeClockOnlyDateTimes(values map[string]string, startKey, endKey, tzKey string) {
	parsing.NormalizeClockOnlyDateTimes(values, startKey, endKey, tzKey)
}

func normalizeEndTimeFromDuration(values map[string]string, startKey, endKey, durationKey, tzKey, defaultDuration string) {
	parsing.NormalizeEndTimeFromDuration(values, startKey, endKey, durationKey, tzKey, defaultDuration)
}

func getValueIfKeyNotEmpty(values map[string]string, key string) string {
	return parsing.GetValueIfKeyNotEmpty(values, key)
}

func trySetEndFromDurationInEnd(values map[string]string, start, tz, end, endKey, durationKey string) bool {
	return parsing.TrySetEndFromDurationInEnd(values, start, tz, end, endKey, durationKey)
}

func setEndFromDuration(values map[string]string, start, tz, dur, endKey, durationKey string) {
	parsing.SetEndFromDuration(values, start, tz, dur, endKey, durationKey)
}

func setEndAndDuration(values map[string]string, endKey, durationKey, endDT string, d time.Duration) {
	parsing.SetEndAndDuration(values, endKey, durationKey, endDT, d)
}

func addDurationToStart(start, tz string, d time.Duration) string {
	return parsing.AddDurationToStart(start, tz, d)
}

func splitDateTime(s string) (string, string) {
	return cli.SplitDateTime(s)
}

func parseDateTimeWithTZ(dateStr, timeStr, tz string) (time.Time, error) {
	return parsing.ParseDateTimeWithTZ(dateStr, timeStr, tz)
}

func fmtDurationHuman(d time.Duration) string {
	return cli.FmtDurationHuman(d)
}

func parseExDateValues(values []string, tz string, allDay bool) ([]time.Time, error) {
	return parsing.ParseExDateValues(values, tz, allDay)
}

// ------------------------------
// Timezone commands
// ------------------------------

func newTimezoneCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "timezone",
		Short: "Timezone information and conversion",
	}

	// timezone list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List timezones (filterable)",
		RunE:  runTZList,
	}
	listCmd.Flags().String("search", "", "Filter by text (matches IANA, display name, or country)")
	listCmd.Flags().String("country", "", "Filter by country (case-insensitive contains)")
	listCmd.Flags().String("region", "", "Filter by region (supported: europe)")
	listCmd.Flags().Bool("all", false, "Show all known zones (ignores region)")

	// timezone info <name|IANA>
	infoCmd := &cobra.Command{
		Use:   "info <name-or-IANA>",
		Short: "Show details for a specific timezone",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runTZInfo,
	}

	root.AddCommand(listCmd, infoCmd)
	return root
}

func cleanDisplay(s string) string {
	return cli.CleanDisplay(s)
}

func runTZList(cmd *cobra.Command, _ []string) error {
	search, _ := cmd.Flags().GetString("search")
	country, _ := cmd.Flags().GetString("country")
	region, _ := cmd.Flags().GetString("region")
	showAll, _ := cmd.Flags().GetBool("all")

	tm := tzpkg.NewTimezoneManager()

	var zones []*tzpkg.TimezoneInfo
	switch {
	case showAll:
		zones = tm.ListTimezones()
	case strings.EqualFold(strings.TrimSpace(region), "europe"):
		zones = tm.GetEuropeanTimezones()
	default:
		zones = tm.ListTimezones()
	}

	search = strings.ToLower(strings.TrimSpace(search))
	country = strings.ToLower(strings.TrimSpace(country))

	filtered := make([]*tzpkg.TimezoneInfo, 0, len(zones))
	for _, z := range zones {
		match := true
		if search != "" {
			if !strings.Contains(strings.ToLower(z.IANA), search) &&
				!strings.Contains(strings.ToLower(z.DisplayName), search) &&
				!strings.Contains(strings.ToLower(z.Country), search) {
				match = false
			}
		}
		if match && country != "" {
			if !strings.Contains(strings.ToLower(z.Country), country) {
				match = false
			}
		}
		if match {
			filtered = append(filtered, z)
		}
	}

	// nicer columns: separate Display & Country
	fmt.Printf("%-32s  %-7s  %-3s  %-28s  %s\n", "IANA", "Offset", "DST", "Display", "Country")
	for _, z := range filtered {
		dst := "no"
		if z.DST {
			dst = "yes"
		}
		name := cleanDisplay(z.DisplayName)
		fmt.Printf("%-32s  %-7s  %-3s  %-28s  %s\n",
			z.IANA, z.Offset, dst, name, z.Country)
	}
	return nil
}

func runTZInfo(_ *cobra.Command, args []string) error {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return fmt.Errorf("please provide a timezone name or IANA identifier")
	}

	tm := tzpkg.NewTimezoneManager()

	// Try exact/alias/system
	zone, err := tm.GetTimezone(query)
	if err != nil {
		// Try city→IANA mapping
		if mapped, mapErr := cityToIANA(query); mapErr == nil && mapped != "" {
			if z2, err2 := tm.GetTimezone(mapped); err2 == nil {
				zone = z2
			}
		}
	}

	if zone == nil {
		// Last-ditch: suggest by fuzzy search
		sugs := tm.SuggestTimezone(query)
		if len(sugs) == 0 {
			fmt.Println("Timezone not found.")
			return nil
		}
		fmt.Println("Timezone not found. Did you mean:")
		for _, s := range sugs {
			fmt.Printf("  - %s (%s) [%s]\n", cleanDisplay(s.DisplayName), s.Country, s.IANA)
		}
		return nil
	}

	loc, err := time.LoadLocation(zone.IANA)
	if err != nil {
		// Still show info without current local time
		printZoneInfo(zone, "", "")
		return nil
	}

	now := time.Now().In(loc)
	printZoneInfo(zone, now.Format(constants.DateTimeFormatISOSeconds), now.Format(constants.DateTimeFormatRFC1123))
	return nil
}

func printZoneInfo(z *tzpkg.TimezoneInfo, local1, local2 string) {
	name := cleanDisplay(z.DisplayName)
	fmt.Printf("IANA:       %s\n", z.IANA)
	fmt.Printf("Display:    %s\n", name)
	fmt.Printf("Country:    %s\n", z.Country)
	fmt.Printf("Offset:     %s\n", z.Offset)
	if z.DST {
		fmt.Printf("DST:        yes\n")
	} else {
		fmt.Printf("DST:        no\n")
	}
	if local1 != "" {
		fmt.Printf("Now:        %s\n", local1)
	}
	if local2 != "" {
		fmt.Printf("Readable:   %s\n", local2)
	}
}

// Lightweight city → IANA mapping for friendlier queries.
func cityToIANA(s string) (string, error) {
	x := strings.ToLower(strings.TrimSpace(s))
	if x == "" {
		return "", fmt.Errorf("Unknown city '%s'. Use 'tempus timezone list --search %s' to find the IANA identifier", s, s)
	}

	// Spain / territories
	if x == "melilla" || x == "ceuta" {
		return "Africa/Ceuta", nil
	}
	if x == "canarias" || x == "gran canaria" || x == "tenerife" || x == "las palmas" {
		return "Atlantic/Canary", nil
	}

	// Brazil
	switch x {
	case "pelotas", "porto alegre", "porto-alegre", "florianopolis", "florianópolis":
		return "America/Sao_Paulo", nil
	case "campo grande", "campo-grande", "ponta pora", "ponta-porã", "ponta-pora", "dourados":
		return "America/Campo_Grande", nil
	case "cuiaba", "cuiabá":
		return "America/Cuiaba", nil
	case "manaus":
		return "America/Manaus", nil
	case "recife":
		return "America/Recife", nil
	case "belem", "belém":
		return "America/Belem", nil
	case "fortaleza":
		return "America/Fortaleza", nil
	case "salvador":
		return "America/Bahia", nil
	case "rio", "rio de janeiro", "rio-de-janeiro", "niteroi", "niterói":
		return "America/Sao_Paulo", nil
	case "sao paulo", "são paulo", "sao-paulo", "campinas":
		return "America/Sao_Paulo", nil
	}

	// Ireland / UK
	if x == "dublin" {
		return "Europe/Dublin", nil
	}
	if x == "london" {
		return "Europe/London", nil
	}

	// Spain common cities
	if x == "madrid" || x == "barcelona" || x == "valencia" || x == "bilbao" || x == "sevilla" {
		return "Europe/Madrid", nil
	}

	return "", fmt.Errorf("Unknown city '%s'. Use 'tempus timezone list --search %s' to find the IANA identifier", s, s)
}

// ------------------------------
// Output helpers (ND-friendly)
// ------------------------------

func printOK(format string, a ...interface{}) {
	cli.PrintOK(os.Stdout, format, a...)
}

func printErr(format string, a ...interface{}) {
	cli.PrintErr(os.Stdout, format, a...)
}

func atoiSafe(s string) int {
	return cli.AtoiSafe(s)
}
