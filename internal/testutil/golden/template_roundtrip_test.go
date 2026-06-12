package golden

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestBatchTemplateRoundtrip is the contract that every generated batch
// template, in every supported format, is directly consumable by
// `tempus batch -i` without manual editing. This guards against template
// headers or field semantics drifting away from the batch loader.
func TestBatchTemplateRoundtrip(t *testing.T) {
	templateTypes := []string{
		"basic", "adhd-routine", "medication", "work-meetings", "medical",
		"travel", "family", "school-event", "recruiter-meeting", "travel-day",
	}
	formats := []string{"csv", "yaml", "json"}

	for _, typ := range templateTypes {
		for _, format := range formats {
			t.Run(typ+"_"+format, func(t *testing.T) {
				workDir := t.TempDir()
				templateFile := filepath.Join(workDir, "events."+format)

				res := RunCLI(t, binPath, workDir, nil, "",
					"batch", "template", typ, "-o", templateFile, "-f", format)
				if res.ExitCode != 0 {
					t.Fatalf("batch template %s -f %s failed (exit %d)\nstdout: %s\nstderr: %s",
						typ, format, res.ExitCode, res.Stdout, res.Stderr)
				}

				outFile := filepath.Join(workDir, "out.ics")
				res = RunCLI(t, binPath, workDir, nil, "",
					"batch", "--dry-run", "-i", templateFile, "-o", outFile)
				if res.ExitCode != 0 {
					t.Fatalf("batch --dry-run on generated %s template (%s) failed (exit %d)\nstdout: %s\nstderr: %s",
						typ, format, res.ExitCode, res.Stdout, res.Stderr)
				}
				if !strings.Contains(res.Stdout, "Validation passed") {
					t.Errorf("expected 'Validation passed' for %s (%s), got:\n%s", typ, format, res.Stdout)
				}
			})
		}
	}
}
