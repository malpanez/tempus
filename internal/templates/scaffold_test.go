package templates

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateScaffoldYAML(t *testing.T) {
	data, err := GenerateScaffold(ScaffoldOptions{
		Name:     "custom",
		Language: "en",
		Format:   "yaml",
	})
	if err != nil {
		t.Fatalf("GenerateScaffold: %v", err)
	}

	var dd DataDrivenTemplate
	if err := yaml.Unmarshal(data, &dd); err != nil {
		t.Fatalf("yaml unmarshal: %v\n%s", err, string(data))
	}

	if dd.Name != "custom" {
		t.Fatalf("expected name custom, got %s", dd.Name)
	}
	if dd.SchemaVersion != 1 {
		t.Fatalf("expected schema version 1, got %d", dd.SchemaVersion)
	}
	if dd.Output.StartField != "start_time" {
		t.Fatalf("unexpected start_field: %s", dd.Output.StartField)
	}
	if len(dd.Fields) == 0 {
		t.Fatalf("expected fields to be populated")
	}
}

func TestGenerateScaffoldJSON(t *testing.T) {
	data, err := GenerateScaffold(ScaffoldOptions{
		Name:     "custom-json",
		Language: "es",
		Format:   "json",
	})
	if err != nil {
		t.Fatalf("GenerateScaffold JSON: %v", err)
	}

	var dd DataDrivenTemplate
	if err := json.Unmarshal(data, &dd); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, string(data))
	}

	if !strings.Contains(dd.Fields[0].Name, "Título") && !strings.Contains(dd.Fields[0].Name, "Titulo") {
		t.Fatalf("expected spanish title label, got %s", dd.Fields[0].Name)
	}
}

func TestResolveTemplateDirs(t *testing.T) {
	custom := filepath.Join("testdata", "templates")
	dirs := ResolveTemplateDirs(custom)
	if len(dirs) != 1 || dirs[0] != filepath.Clean(custom) {
		t.Fatalf("unexpected dirs for custom: %v", dirs)
	}

	defaults := DefaultTemplateDirs()
	if len(defaults) == 0 {
		t.Fatalf("expected default dirs to be non-empty")
	}
}
