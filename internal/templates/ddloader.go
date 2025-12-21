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

	if err := validateTemplateDir(dir); err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return out, err
	}

	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, walkErr error) error {
		return processTemplateFile(p, d, walkErr, out)
	})
	return out, err
}

// validateTemplateDir checks if the directory exists and is actually a directory.
func validateTemplateDir(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}
	return nil
}

// processTemplateFile processes a single file during the directory walk.
func processTemplateFile(p string, d fs.DirEntry, walkErr error, out map[string]DataDrivenTemplate) error {
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

	tmpl, err := loadAndDecodeTemplate(p, ext)
	if err != nil {
		return err
	}

	if err := normalizeTemplateMetadata(&tmpl, p); err != nil {
		return err
	}

	if err := ValidateDDTemplate(&tmpl); err != nil {
		return fmt.Errorf("%s: %w", p, err)
	}

	out[tmpl.Name] = tmpl
	return nil
}

// loadAndDecodeTemplate reads and decodes a template file.
func loadAndDecodeTemplate(path, ext string) (DataDrivenTemplate, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return DataDrivenTemplate{}, err
	}

	tmpl, err := decodeDDTemplate(data, ext)
	if err != nil {
		return DataDrivenTemplate{}, fmt.Errorf("%s: %w", path, err)
	}
	return tmpl, nil
}

// normalizeTemplateMetadata sets defaults and normalizes template metadata.
func normalizeTemplateMetadata(tmpl *DataDrivenTemplate, path string) error {
	if strings.TrimSpace(tmpl.Name) == "" {
		base := filepath.Base(path)
		tmpl.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if tmpl.SchemaVersion == 0 {
		tmpl.SchemaVersion = 1
	}
	if tmpl.SchemaVersion != 1 {
		return fmt.Errorf("%s: unsupported schema_version %d", path, tmpl.SchemaVersion)
	}
	tmpl.Source = path
	return nil
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
	if err := validateBasicTemplateInfo(t); err != nil {
		return err
	}

	fieldKeys, err := buildFieldKeysMap(t)
	if err != nil {
		return err
	}

	if err := validateOutputFields(t, fieldKeys); err != nil {
		return err
	}

	if strings.TrimSpace(t.Output.SummaryTmpl) == "" {
		return fmt.Errorf("template %q missing output.summary_tmpl", t.Name)
	}

	return nil
}

// validateBasicTemplateInfo validates the template name and that fields are defined.
func validateBasicTemplateInfo(t *DataDrivenTemplate) error {
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	if len(t.Fields) == 0 {
		return fmt.Errorf("template %q must define at least one field", t.Name)
	}
	return nil
}

// buildFieldKeysMap builds a map of field keys and validates they are unique and non-empty.
func buildFieldKeysMap(t *DataDrivenTemplate) (map[string]struct{}, error) {
	fieldKeys := make(map[string]struct{}, len(t.Fields))
	for i, f := range t.Fields {
		key := strings.TrimSpace(f.Key)
		if key == "" {
			return nil, fmt.Errorf("template %q field %d is missing a key", t.Name, i)
		}
		if _, exists := fieldKeys[key]; exists {
			return nil, fmt.Errorf("template %q has duplicate field key %q", t.Name, key)
		}
		fieldKeys[key] = struct{}{}
	}
	return fieldKeys, nil
}

// validateOutputFields validates all output field references.
func validateOutputFields(t *DataDrivenTemplate, fieldKeys map[string]struct{}) error {
	fieldsToCheck := []struct {
		label      string
		key        string
		allowEmpty bool
	}{
		{"output.start_field", t.Output.StartField, false},
		{"output.end_field", t.Output.EndField, true},
		{"output.duration_field", t.Output.DurationField, true},
		{"output.start_tz_field", t.Output.StartTZField, true},
		{"output.end_tz_field", t.Output.EndTZField, true},
		{"output.rrule_field", t.Output.RRuleField, true},
		{"output.exdates_field", t.Output.ExDatesField, true},
		{"output.alarms_field", t.Output.AlarmsField, true},
	}

	for _, field := range fieldsToCheck {
		if err := validateFieldReference(t.Name, field.label, field.key, field.allowEmpty, fieldKeys); err != nil {
			return err
		}
	}
	return nil
}

// validateFieldReference validates a single field reference.
func validateFieldReference(templateName, label, key string, allowEmpty bool, fieldKeys map[string]struct{}) error {
	key = strings.TrimSpace(key)
	if key == "" {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("template %q missing %s", templateName, label)
	}
	if _, ok := fieldKeys[key]; !ok {
		return fmt.Errorf("template %q references unknown %s %q", templateName, label, key)
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
