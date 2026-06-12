package cli

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tempus/internal/calendar"
	"tempus/internal/config"
	"tempus/internal/constants"
	"tempus/internal/i18n"
	"tempus/internal/parsing"
	"tempus/internal/prompts"
	tpl "tempus/internal/templates"
	"tempus/internal/testutil"

	"github.com/spf13/cobra"
)

type templateCreateParams struct {
	templateName string
	inputPath    string
	formatFlag   string
	outputDir    string
}

func NewTemplateCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage event templates",
	}

	cmd.PersistentFlags().String("templates-dir", "", "Directory with JSON templates (default: user config dir)")

	createCmd := &cobra.Command{
		Use:   "create <template-name>",
		Short: "Create event from template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateCreate(app, cmd, args)
		},
	}
	createCmd.Flags().String("output-dir", "", "Directory where generated ICS files will be stored")
	createCmd.Flags().String("input", "", "CSV or JSON file with template data (creates one ICS per row)")
	createCmd.Flags().String("format", "auto", "Input format: auto, csv, or json")
	createCmd.Flags().String("templates-dir", "", "Directory with JSON templates (overrides defaults)")

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available templates",
			RunE: func(cmd *cobra.Command, args []string) error {
				return runTemplateList(app, cmd, args)
			},
		},
		createCmd,
		&cobra.Command{
			Use:   "describe <template-name>",
			Short: "Show details for a template",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runTemplateDescribe(app, cmd, args)
			},
		},
		&cobra.Command{
			Use:   "validate",
			Short: "Validate data-driven templates",
			RunE: func(cmd *cobra.Command, args []string) error {
				return runTemplateValidate(app, cmd, args)
			},
		},
		newTemplateInitCmd(app),
	)

	return cmd
}

func newTemplateInitCmd(app *App) *cobra.Command {
	defaultDir := "."
	if dirs := tpl.DefaultTemplateDirs(); len(dirs) > 0 {
		defaultDir = dirs[0]
	}

	cmd := &cobra.Command{
		Use:   "init <template-name>",
		Short: "Generate a template scaffold file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateInit(app, cmd, args)
		},
	}
	cmd.Flags().String("dir", defaultDir, "Directory where the scaffold will be written")
	cmd.Flags().String("format", "yaml", "Output format (yaml or json)")
	cmd.Flags().String("lang", "en", "Language for sample labels (en, es, pt)")
	cmd.Flags().Bool("force", false, "Overwrite the file if it already exists")
	return cmd
}

func runTemplateList(app *App, cmd *cobra.Command, _ []string) error {
	w := stdoutWriter(app)
	tm, _, err := loadTemplateManager(cmd)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "Available templates:")
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
		fmt.Fprintf(w, "  %-12s  %s\n", name, desc)
	}
	return nil
}

func runTemplateCreate(app *App, cmd *cobra.Command, args []string) error {
	w := stdoutWriter(app)
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
		return runTemplateCreateFromFile(app, tm, tr, tmpl, dd, params)
	}

	values := map[string]string{}
	for _, f := range tmpl.Fields {
		if isAlarmField(f) {
			values[f.Key] = promptAlarmField(labelForField(f), f.Default)
			continue
		}
		v := prompts.Input(labelForField(f), f.Default)
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
	userOutput := prompts.Input("Output filename", defaultName)
	finalName := strings.TrimSpace(userOutput)
	if finalName == "" {
		finalName = defaultName
	}
	finalName = EnsureICSExtension(finalName)
	if dir := strings.TrimSpace(outputDir); dir != "" && !filepath.IsAbs(finalName) {
		finalName = filepath.Join(dir, finalName)
	}
	finalName = EnsureUniquePath(finalName)

	if err := EnsureDirForFile(finalName); err != nil {
		return err
	}
	if err := os.WriteFile(finalName, []byte(cal.ToICS()), 0600); err != nil {
		PrintErr(w, constants.ErrMsgFailedToWriteFile, err)
		return err
	}
	PrintOK(w, "Created: %s\n", finalName)
	return nil
}

func runTemplateCreateFromFile(app *App, tm *tpl.TemplateManager, tr *i18n.Translator, tmpl *tpl.Template, dd tpl.DataDrivenTemplate, params templateCreateParams) error {
	w := stdoutWriter(app)
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
		filename = EnsureICSExtension(filename)
		if params.outputDir != "" && !filepath.IsAbs(filename) {
			filename = filepath.Join(params.outputDir, filename)
		}
		filename = EnsureUniquePath(filename)

		if err := EnsureDirForFile(filename); err != nil {
			return fmt.Errorf(testutil.ErrMsgRowFormat, idx+1, err)
		}
		if err := os.WriteFile(filename, []byte(cal.ToICS()), 0600); err != nil {
			return fmt.Errorf("row %d: failed to write file: %w", idx+1, err)
		}
		PrintOK(w, "Created: %s\n", filename)
	}

	return nil
}

func runTemplateDescribe(app *App, cmd *cobra.Command, args []string) error {
	w := stdoutWriter(app)
	name := args[0]
	tm, _, err := loadTemplateManager(cmd)
	if err != nil {
		return err
	}

	tmpl, err := tm.GetTemplate(name)
	if err != nil {
		return err
	}

	printTemplateBasicInfo(w, tmpl)
	dd := printTemplateTypeInfo(w, tm, name)
	printTemplateFields(w, tmpl.Fields)
	printTemplateOutput(w, dd)
	return nil
}

func printTemplateBasicInfo(w io.Writer, tmpl *tpl.Template) {
	fmt.Fprintf(w, "Name: %s\n", tmpl.Name)
	if desc := strings.TrimSpace(tmpl.Description); desc != "" {
		fmt.Fprintf(w, "Description: %s\n", desc)
	} else {
		fmt.Fprintln(w, "Description: -")
	}
}

func printTemplateTypeInfo(w io.Writer, tm *tpl.TemplateManager, name string) tpl.DataDrivenTemplate {
	var dd tpl.DataDrivenTemplate
	if ddt, ok := tm.DataTemplate(name); ok {
		fmt.Fprintf(w, "Type: data-driven (schema v%d)\n", ddt.SchemaVersion)
		if strings.TrimSpace(ddt.Source) != "" {
			fmt.Fprintf(w, "Source: %s\n", ddt.Source)
		} else {
			fmt.Fprintln(w, "Source: embedded")
		}
		if strings.TrimSpace(ddt.FilenameTemplate) != "" {
			fmt.Fprintf(w, "Filename template: %s\n", ddt.FilenameTemplate)
		}
		dd = ddt
	} else {
		fmt.Fprintln(w, "Type: built-in (compiled)")
	}
	return dd
}

func printTemplateFields(w io.Writer, fields []tpl.Field) {
	fmt.Fprintln(w, "Fields:")
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
		fmt.Fprintln(w, line)
	}
}

func printTemplateOutput(w io.Writer, dd tpl.DataDrivenTemplate) {
	if dd.Name == "" {
		return
	}
	fmt.Fprintln(w, "Output:")
	fmt.Fprintf(w, "  start_field: %s\n", dd.Output.StartField)
	printIfNotEmpty(w, "  end_field: %s\n", dd.Output.EndField)
	printIfNotEmpty(w, "  duration_field: %s\n", dd.Output.DurationField)
	printIfNotEmpty(w, "  start_tz_field: %s\n", dd.Output.StartTZField)
	if strings.TrimSpace(dd.Output.EndTZField) != "" && dd.Output.EndTZField != dd.Output.StartTZField {
		fmt.Fprintf(w, "  end_tz_field: %s\n", dd.Output.EndTZField)
	}
	printIfNotEmpty(w, "  summary_tmpl: %s\n", dd.Output.SummaryTmpl)
	printIfNotEmpty(w, "  location_tmpl: %s\n", dd.Output.LocationTmpl)
	printIfNotEmpty(w, "  description_tmpl: %s\n", dd.Output.DescriptionTmpl)
	if len(dd.Output.Categories) > 0 {
		fmt.Fprintf(w, "  categories: %s\n", strings.Join(dd.Output.Categories, ", "))
	}
	if dd.Output.Priority > 0 {
		fmt.Fprintf(w, "  priority: %d\n", dd.Output.Priority)
	}
}

func printIfNotEmpty(w io.Writer, format, value string) {
	if strings.TrimSpace(value) != "" {
		fmt.Fprintf(w, format, value)
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
	defer f.Close() //nolint:errcheck // read-only file, close error is harmless

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
		parsing.NormalizeClockOnlyDateTimes(values, "start_time", "end_time", "timezone")
		parsing.NormalizeEndTimeFromDuration(values, "start_time", "end_time", "duration", "timezone", FirstNonEmpty(values["duration"], durationDefault, "30m"))
		return
	}

	startField := strings.TrimSpace(dd.Output.StartField)
	endField := strings.TrimSpace(dd.Output.EndField)
	durationField := strings.TrimSpace(dd.Output.DurationField)
	tzField := strings.TrimSpace(dd.Output.StartTZField)
	if tzField == "" {
		tzField = strings.TrimSpace(dd.Output.EndTZField)
	}

	parsing.NormalizeClockOnlyDateTimes(values, startField, endField, tzField)
	var durationDefault string
	if durationField != "" {
		durationDefault = FirstNonEmpty(values[durationField], templateFieldDefault(tmpl, durationField))
	}
	parsing.NormalizeEndTimeFromDuration(values, startField, endField, durationField, tzField, durationDefault)
}

func buildTemplateCalendar(ev *calendar.Event) *calendar.Calendar {
	cal := calendar.NewCalendar()
	cal.IncludeVTZ = true
	cal.AddEvent(ev)
	cal.Name = ev.Summary
	if tz := FirstNonEmpty(ev.StartTZ, ev.EndTZ); strings.TrimSpace(tz) != "" {
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
		base := Slugify(ev.Summary)
		if base == "" {
			base = Slugify(templateName)
		}
		return fmt.Sprintf("%s-%s.ics", base, ev.StartTime.Format("2006-01-02"))
	}
	if sdate, ok := tm.GuessStartDate(templateName, values); ok && strings.TrimSpace(sdate) != "" {
		return fmt.Sprintf("%s-%s.ics", Slugify(templateName), sdate)
	}
	return fmt.Sprintf("%s.ics", Slugify(templateName))
}

func runTemplateValidate(app *App, cmd *cobra.Command, _ []string) error {
	w := stdoutWriter(app)
	templatesDirFlag, _ := cmd.Flags().GetString("templates-dir")
	dirs := tpl.ResolveTemplateDirs(templatesDirFlag)
	if len(dirs) == 0 {
		fmt.Fprintln(w, "No template directories configured.")
		return nil
	}

	fmt.Fprintln(w, "Validating template directories:")
	var validationErr bool
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(w, " - %s: not found (skipped)\n", dir)
				continue
			}
			fmt.Fprintf(w, " - %s: error: %v\n", dir, err)
			validationErr = true
			continue
		}
		if !info.IsDir() {
			fmt.Fprintf(w, " - %s: not a directory (skipped)\n", dir)
			continue
		}

		defs, err := tpl.LoadDDTemplates(dir)
		if err != nil {
			fmt.Fprintf(w, " - %s: error: %v\n", dir, err)
			validationErr = true
			continue
		}
		names := make([]string, 0, len(defs))
		for name := range defs {
			names = append(names, name)
		}
		sort.Strings(names)
		if len(names) == 0 {
			fmt.Fprintf(w, " - %s: ok (no templates found)\n", dir)
			continue
		}
		fmt.Fprintf(w, " - %s: ok (%d template(s): %s)\n", dir, len(names), strings.Join(names, ", "))
	}

	if validationErr {
		return fmt.Errorf("template validation failed")
	}
	fmt.Fprintln(w, "Validation completed successfully.")
	return nil
}

func runTemplateInit(app *App, cmd *cobra.Command, args []string) error {
	w := stdoutWriter(app)
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

	filename := filepath.Join(dir, Slugify(name)+ext)
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

	PrintOK(w, "Created scaffold: %s\n", filename)
	return nil
}

func loadTemplateManager(cmd *cobra.Command) (*tpl.TemplateManager, *i18n.Translator, error) {
	cfg, _ := config.Load()
	langFlag, _ := cmd.Root().Flags().GetString("language")
	templatesDirFlag, _ := cmd.Flags().GetString("templates-dir")

	cfgLang := ""
	if cfg != nil {
		if v, err := cfg.Get("language"); err == nil {
			cfgLang = v
		}
	}
	lang := FirstNonEmpty(langFlag, cfgLang, "en")

	tr, err := newTranslator(lang)
	if err != nil {
		return nil, nil, err
	}

	tm := tpl.NewTemplateManager()

	for _, dir := range tpl.ResolveTemplateDirs(templatesDirFlag) {
		tm.LoadDDDir(dir)
	}

	return tm, tr, nil
}

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

func promptAlarmField(label, defaultValue string) string {
	fmt.Printf("\n%s\n", label)
	existing := calendar.SplitAlarmInput(defaultValue)
	if len(existing) > 0 {
		fmt.Println("Recordatorios sugeridos:")
		for i, spec := range existing {
			fmt.Printf("  %d) %s\n", i+1, spec)
		}
		keep := strings.ToLower(strings.TrimSpace(prompts.Input("Pulsa Enter para mantenerlos o escribe 'n' para cambiarlos", "")))
		if keep == "" || keep == "s" || keep == "si" {
			return strings.Join(existing, "\n")
		}
		fmt.Println("")
	}

	fmt.Println("Añade hasta 4 recordatorios. Usa formatos como -15m, +10m, 2025-03-01 09:15 o trigger=-15m,description=Texto.")
	fmt.Println("Escribe '?' para ver ejemplos o deja vacío para terminar.")

	specs := make([]string, 0, 4)
	for len(specs) < 4 {
		p := fmt.Sprintf("Recordatorio #%d (-15m, +10m, trigger=..., ? para ayuda)", len(specs)+1)
		input := strings.TrimSpace(prompts.Input(p, ""))
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
			desc := strings.TrimSpace(prompts.Input("Descripción opcional (Enter para usar la genérica)", ""))
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
