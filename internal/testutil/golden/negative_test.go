package golden

import (
	"path/filepath"
	"strings"
	"testing"
)

// Negative tests: invalid user input must fail loudly with exit code != 0
// and an error message naming the offending value. These pin the phase 1
// fail-loud contract (.planning/tempus-remediation/MILESTONE.md).

func runExpectingFailure(t *testing.T, args ...string) Result {
	t.Helper()
	workDir := t.TempDir()
	outPath := filepath.Join(workDir, "out.ics")
	args = append(args, "-o", outPath)
	res := RunCLI(t, binPath, workDir, nil, "", args...)
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0\nstdout: %s\nstderr: %s", res.Stdout, res.Stderr)
	}
	return res
}

func TestNegativeInvalidExDate(t *testing.T) {
	res := runExpectingFailure(t,
		"create", "Recurring Thing",
		"--start", "2030-05-10 10:30",
		"--duration", "30m",
		"--rrule", "FREQ=WEEKLY;COUNT=4",
		"--exdate", "2030-13-99",
	)
	if !strings.Contains(res.Stderr+res.Stdout, "2030-13-99") {
		t.Errorf("error should name the offending exdate value, got stderr: %s stdout: %s", res.Stderr, res.Stdout)
	}
}

func TestNegativeInvalidAlarmSpec(t *testing.T) {
	res := runExpectingFailure(t,
		"create", "Event",
		"--start", "2030-05-10 10:30",
		"--duration", "30m",
		"--alarm", "not-a-real-alarm-spec",
	)
	if !strings.Contains(res.Stderr+res.Stdout, "not-a-real-alarm-spec") {
		t.Errorf("error should name the offending alarm spec, got stderr: %s stdout: %s", res.Stderr, res.Stdout)
	}
}

func TestNegativeUnknownAlarmProfile(t *testing.T) {
	res := runExpectingFailure(t,
		"create", "Event",
		"--start", "2030-05-10 10:30",
		"--duration", "30m",
		"--alarm", "profile:does-not-exist",
	)
	combined := res.Stderr + res.Stdout
	if !strings.Contains(combined, "does-not-exist") || !strings.Contains(combined, "Available") {
		t.Errorf("error should name the unknown profile and list available ones, got: %s", combined)
	}
}

func TestNegativeMissingExplicitConfig(t *testing.T) {
	workDir := t.TempDir()
	res := RunCLI(t, binPath, workDir, nil, "",
		"-c", filepath.Join(workDir, "nonexistent.yaml"),
		"create", "Event",
		"--start", "2030-05-10 10:30",
		"--duration", "30m",
		"-o", filepath.Join(workDir, "out.ics"),
	)
	if res.ExitCode == 0 {
		t.Fatalf("explicit -c pointing at a missing file must fail, got exit 0\nstdout: %s\nstderr: %s", res.Stdout, res.Stderr)
	}
	if !strings.Contains(res.Stderr+res.Stdout, "nonexistent.yaml") {
		t.Errorf("error should name the missing config file, got stderr: %s stdout: %s", res.Stderr, res.Stdout)
	}
}

func TestNegativeInvalidStartTZ(t *testing.T) {
	res := runExpectingFailure(t,
		"create", "Event",
		"--start", "2030-05-10 10:30",
		"--duration", "30m",
		"--start-tz", "narnia",
	)
	if !strings.Contains(res.Stderr+res.Stdout, "narnia") {
		t.Errorf("error should name the invalid timezone, got stderr: %s stdout: %s", res.Stderr, res.Stdout)
	}
}

func TestNegativeInvalidGlobalTZ(t *testing.T) {
	res := runExpectingFailure(t,
		"-t", "mordor",
		"create", "Event",
		"--start", "2030-05-10 10:30",
		"--duration", "30m",
	)
	if !strings.Contains(res.Stderr+res.Stdout, "mordor") {
		t.Errorf("error should name the invalid timezone, got stderr: %s stdout: %s", res.Stderr, res.Stdout)
	}
}

// TestNegativeInteractiveWithIncompatibleFlags: the wizard must never
// silently ignore flags the user typed — anything other than -o errors out.
func TestNegativeInteractiveWithIncompatibleFlags(t *testing.T) {
	workDir := t.TempDir()
	res := RunCLI(t, binPath, workDir, nil, "",
		"create", "-i",
		"--start", "2030-05-10 10:30",
		"--duration", "45m",
	)
	if res.ExitCode == 0 {
		t.Fatalf("create -i with --start must fail, got exit 0\nstdout: %s\nstderr: %s", res.Stdout, res.Stderr)
	}
	combined := res.Stderr + res.Stdout
	if !strings.Contains(combined, "--start") || !strings.Contains(combined, "--duration") {
		t.Errorf("error should name the incompatible flags, got: %s", combined)
	}
}

func TestNegativeBatchInvalidExDateRow(t *testing.T) {
	workDir := t.TempDir()
	csvPath := filepath.Join(workDir, "bad.csv")
	csv := "summary,start,duration,exdate\nTherapy,2030-06-03 17:00,1h,2030-99-99\n"
	if err := writeFile(csvPath, csv); err != nil {
		t.Fatal(err)
	}
	res := RunCLI(t, binPath, workDir, nil, "",
		"batch", "-i", csvPath, "-o", filepath.Join(workDir, "out.ics"))
	if res.ExitCode == 0 {
		t.Fatalf("batch with invalid exdate must fail, got exit 0\nstdout: %s\nstderr: %s", res.Stdout, res.Stderr)
	}
	combined := res.Stderr + res.Stdout
	if !strings.Contains(combined, "2030-99-99") {
		t.Errorf("error should name the offending exdate, got: %s", combined)
	}
}
