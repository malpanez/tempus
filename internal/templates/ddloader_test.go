package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDDTemplatesJSONAndYAML(t *testing.T) {
	dir := t.TempDir()

	jsonTemplate := `{
		"schema_version": 1,
		"name": "json-demo",
		"fields": [
			{"key": "title", "name": "Title", "type": "text", "required": true},
			{"key": "start_time", "name": "Start", "type": "datetime", "required": true},
			{"key": "duration", "name": "Duration", "type": "text"},
			{"key": "timezone", "name": "Timezone", "type": "timezone"}
		],
		"output": {
			"start_field": "start_time",
			"duration_field": "duration",
			"start_tz_field": "timezone",
			"summary_tmpl": "Event: {{title}}"
		}
	}`

	yamlTemplate := `
schema_version: 1
name: yaml-demo
fields:
  - key: title
    name: Title
    type: text
    required: true
  - key: start_time
    name: Start
    type: datetime
    required: true
  - key: duration
    name: Duration
    type: text
  - key: timezone
    name: Timezone
    type: timezone
output:
  start_field: start_time
  duration_field: duration
  start_tz_field: timezone
  summary_tmpl: "{{title}}"
`

	if err := os.WriteFile(filepath.Join(dir, "demo.json"), []byte(jsonTemplate), 0o644); err != nil {
		t.Fatalf("failed to write json template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "demo.yaml"), []byte(yamlTemplate), 0o644); err != nil {
		t.Fatalf("failed to write yaml template: %v", err)
	}

	templates, err := LoadDDTemplates(dir)
	if err != nil {
		t.Fatalf("LoadDDTemplates failed: %v", err)
	}

	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	jsonTmpl, ok := templates["json-demo"]
	if !ok {
		t.Fatalf("expected template json-demo to be registered")
	}
	if jsonTmpl.Output.StartField != "start_time" {
		t.Fatalf("unexpected start field: %s", jsonTmpl.Output.StartField)
	}
	if jsonTmpl.SchemaVersion != 1 {
		t.Fatalf("unexpected schema version: %d", jsonTmpl.SchemaVersion)
	}

	yamlTmpl, ok := templates["yaml-demo"]
	if !ok {
		t.Fatalf("expected template yaml-demo to be registered")
	}
	if yamlTmpl.Output.SummaryTmpl != "{{title}}" {
		t.Fatalf("unexpected summary template: %s", yamlTmpl.Output.SummaryTmpl)
	}
}

func TestLoadDDTemplatesInvalid(t *testing.T) {
	dir := t.TempDir()
	invalid := `{
		"name": "invalid",
		"fields": [
			{"key": "title", "name": "Title", "type": "text"}
		],
		"output": {
			"summary_tmpl": "{{title}}"
		}
	}`

	if err := os.WriteFile(filepath.Join(dir, "invalid.json"), []byte(invalid), 0o644); err != nil {
		t.Fatalf("failed to write invalid template: %v", err)
	}

	if _, err := LoadDDTemplates(dir); err == nil {
		t.Fatalf("expected LoadDDTemplates to fail for invalid template")
	}
}
