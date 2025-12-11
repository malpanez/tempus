package templates

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadDDTemplates scans a directory for data-driven templates (JSON/YAML).
// It is safe to call this on a non-existing directory; it will return an empty map and nil error.
func LoadDDTemplates(dir string) (map[string]DataDrivenTemplate, error) {
	out := map[string]DataDrivenTemplate{}

	fi, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return out, err
	}
	if !fi.IsDir() {
		return out, nil
	}

	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if !isTemplateFileExt(ext) {
			return nil
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		tmpl, err := decodeDDTemplate(data, ext)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}

		if strings.TrimSpace(tmpl.Name) == "" {
			base := filepath.Base(p)
			tmpl.Name = strings.TrimSuffix(base, filepath.Ext(base))
		}
		if tmpl.SchemaVersion == 0 {
			tmpl.SchemaVersion = 1
		}
		if tmpl.SchemaVersion != 1 {
			return fmt.Errorf("%s: unsupported schema_version %d", p, tmpl.SchemaVersion)
		}

		tmpl.Source = p

		if err := ValidateDDTemplate(&tmpl); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}

		out[tmpl.Name] = tmpl
		return nil
	})
	return out, err
}

func decodeDDTemplate(data []byte, ext string) (DataDrivenTemplate, error) {
	var tmpl DataDrivenTemplate
	var err error
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &tmpl)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &tmpl)
	default:
		err = fmt.Errorf("unsupported template format %s", ext)
	}
	return tmpl, err
}

func ValidateDDTemplate(t *DataDrivenTemplate) error {
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	if len(t.Fields) == 0 {
		return fmt.Errorf("template %q must define at least one field", t.Name)
	}

	fieldKeys := make(map[string]struct{}, len(t.Fields))
	for i, f := range t.Fields {
		key := strings.TrimSpace(f.Key)
		if key == "" {
			return fmt.Errorf("template %q field %d is missing a key", t.Name, i)
		}
		if _, exists := fieldKeys[key]; exists {
			return fmt.Errorf("template %q has duplicate field key %q", t.Name, key)
		}
		fieldKeys[key] = struct{}{}
	}

	checkField := func(label, key string, allowEmpty bool) error {
		key = strings.TrimSpace(key)
		if key == "" {
			if allowEmpty {
				return nil
			}
			return fmt.Errorf("template %q missing %s", t.Name, label)
		}
		if _, ok := fieldKeys[key]; !ok {
			return fmt.Errorf("template %q references unknown %s %q", t.Name, label, key)
		}
		return nil
	}

	if err := checkField("output.start_field", t.Output.StartField, false); err != nil {
		return err
	}
	if err := checkField("output.end_field", t.Output.EndField, true); err != nil {
		return err
	}
	if err := checkField("output.duration_field", t.Output.DurationField, true); err != nil {
		return err
	}
	if err := checkField("output.start_tz_field", t.Output.StartTZField, true); err != nil {
		return err
	}
	if err := checkField("output.end_tz_field", t.Output.EndTZField, true); err != nil {
		return err
	}
	if err := checkField("output.rrule_field", t.Output.RRuleField, true); err != nil {
		return err
	}
	if err := checkField("output.exdates_field", t.Output.ExDatesField, true); err != nil {
		return err
	}
	if err := checkField("output.alarms_field", t.Output.AlarmsField, true); err != nil {
		return err
	}

	if strings.TrimSpace(t.Output.SummaryTmpl) == "" {
		return fmt.Errorf("template %q missing output.summary_tmpl", t.Name)
	}

	return nil
}

func isTemplateFileExt(ext string) bool {
	switch ext {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}
