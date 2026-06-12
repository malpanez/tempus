package golden

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "regenerate golden files")

var binPath string

func TestMain(m *testing.M) {
	flag.Parse()
	dir, err := os.MkdirTemp("", "tempus-golden-bin-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir) //nolint:errcheck // temp dir cleanup

	binPath, err = BuildBinary(dir, repoRootFromHere())
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func repoRootFromHere() string {
	abs, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		panic(err)
	}
	return abs
}

// runAndReadICS executes the CLI writing to out.ics in a fresh temp dir and
// returns the normalized ICS content.
func runAndReadICS(t *testing.T, extraEnv []string, args ...string) string {
	t.Helper()
	workDir := t.TempDir()
	outPath := filepath.Join(workDir, "out.ics")
	args = append(args, "-o", outPath)
	res := RunCLI(t, binPath, workDir, extraEnv, "", args...)
	if res.ExitCode != 0 {
		t.Fatalf("exit code %d, stdout: %s, stderr: %s", res.ExitCode, res.Stdout, res.Stderr)
	}
	raw, err := os.ReadFile(filepath.Clean(outPath))
	if err != nil {
		t.Fatalf("reading %s: %v (stdout: %s, stderr: %s)", outPath, err, res.Stdout, res.Stderr)
	}
	return NormalizeICS(string(raw))
}

func TestGoldenCreateSimple(t *testing.T) {
	ics := runAndReadICS(t, nil,
		"create", "Simple Event",
		"--start", "2030-05-10 10:30",
		"--duration", "45m",
	)
	CompareGolden(t, *update, "create_simple", ics)
}

func TestGoldenCreateStartTZ(t *testing.T) {
	ics := runAndReadICS(t, nil,
		"create", "Madrid Meeting",
		"--start", "2030-05-10 10:30",
		"--duration", "1h",
		"--start-tz", "Europe/Madrid",
	)
	CompareGolden(t, *update, "create_start_tz", ics)
}

// TestGoldenCreateProfileAlarm pins the two acceptance criteria of the
// remediation milestone: the adhd-default profile emits its 4 VALARM blocks
// (fixed in phase 1) and the global -t flag yields TZID=Europe/Madrid local
// times with an embedded VTIMEZONE (fixed in phase 2).
func TestGoldenCreateProfileAlarm(t *testing.T) {
	ics := runAndReadICS(t, nil,
		"-t", "Europe/Madrid",
		"create", "Med Reminder",
		"--start", "2030-05-10 10:30",
		"--duration", "45m",
		"--alarm", "profile:adhd-default",
	)
	CompareGolden(t, *update, "create_profile_alarm", ics)
}

// TestGoldenDefaultTZEnv: TEMPUS_TIMEZONE must be honored by create.
func TestGoldenDefaultTZEnv(t *testing.T) {
	ics := runAndReadICS(t, []string{"TEMPUS_TIMEZONE=Europe/Madrid"},
		"create", "Env TZ Event",
		"--start", "2030-05-10 10:30",
		"--duration", "45m",
	)
	CompareGolden(t, *update, "create_default_tz_env", ics)
}

// TestGoldenAliasTZ: city aliases resolve to IANA zones before emission.
func TestGoldenAliasTZ(t *testing.T) {
	ics := runAndReadICS(t, nil,
		"create", "Alias TZ Event",
		"--start", "2030-05-10 10:30",
		"--duration", "45m",
		"--start-tz", "madrid",
	)
	CompareGolden(t, *update, "create_alias_tz", ics)
}

func TestGoldenBatchExdates(t *testing.T) {
	workDir := t.TempDir()
	outPath := filepath.Join(workDir, "out.ics")
	csvPath, err := filepath.Abs(filepath.Join("testdata", "batch_exdates.csv"))
	if err != nil {
		t.Fatal(err)
	}
	res := RunCLI(t, binPath, workDir, nil, "",
		"batch", "-i", csvPath, "-o", outPath)
	if res.ExitCode != 0 {
		t.Fatalf("exit code %d, stdout: %s, stderr: %s", res.ExitCode, res.Stdout, res.Stderr)
	}
	raw, err := os.ReadFile(filepath.Clean(outPath))
	if err != nil {
		t.Fatalf("reading %s: %v", outPath, err)
	}
	CompareGolden(t, *update, "batch_exdates", NormalizeICS(string(raw)))
}

// TestGoldenRRuleUntil captures current behavior for a recurring event with
// a date-only UNTIL on a timed DTSTART (RFC 5545 §3.3.10 requires matching
// value types — known gap, phase 3 will change this golden).
func TestGoldenRRuleUntil(t *testing.T) {
	ics := runAndReadICS(t, nil,
		"create", "Weekly Sync",
		"--start", "2030-05-10 10:30",
		"--duration", "30m",
		"--rrule", "FREQ=WEEKLY;UNTIL=20301231",
		"--exdate", "2030-05-17 10:30",
	)
	CompareGolden(t, *update, "rrule_until", ics)
}

// TestGoldensPassOwnLint: every golden the tool generates must pass the
// tool's own RFC 5545 gate. This closes the loop between emission and lint.
func TestGoldensPassOwnLint(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("testdata", "golden", "*.ics"))
	if err != nil || len(matches) == 0 {
		t.Fatalf("no goldens found: %v", err)
	}
	for _, m := range matches {
		abs, err := filepath.Abs(m)
		if err != nil {
			t.Fatal(err)
		}
		res := RunCLI(t, binPath, t.TempDir(), nil, "", "lint", abs)
		if res.ExitCode != 0 {
			t.Errorf("golden %s fails own lint: %s%s", m, res.Stdout, res.Stderr)
		}
	}
}

func TestGoldenAllDay(t *testing.T) {
	ics := runAndReadICS(t, nil,
		"create", "Conference",
		"--start", "2030-05-10",
		"--end", "2030-05-12",
		"--all-day",
	)
	CompareGolden(t, *update, "all_day", ics)
}
