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

// TestGoldenCreateProfileAlarm captures two known bugs as current behavior:
// B1 — profile: alarm specs are silently dropped, so the output has NO
// VALARM blocks despite the user asking for the adhd-default profile.
// B2 — the global -t flag is ignored by create, so DTSTART/DTEND carry the
// wall-clock time stamped as UTC (Z) instead of TZID=Europe/Madrid.
// When phase 1/2 fix these, this golden MUST change (VALARMs appear, TZID
// appears) — regenerate with -update and review.
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

func TestGoldenAllDay(t *testing.T) {
	ics := runAndReadICS(t, nil,
		"create", "Conference",
		"--start", "2030-05-10",
		"--end", "2030-05-12",
		"--all-day",
	)
	CompareGolden(t, *update, "all_day", ics)
}
