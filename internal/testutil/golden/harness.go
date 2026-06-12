// Package golden provides an end-to-end golden-file harness that builds the
// tempus binary, runs full CLI flows, and compares the generated ICS output
// against testdata/golden/*.ics after normalizing volatile fields.
//
// Goldens capture CURRENT behavior, bugs included — they are a regression
// net, not a correctness oracle. When a bug fix intentionally changes output,
// regenerate with:
//
//	go test ./internal/testutil/golden/ -run TestGolden -update
//
// and review the diff by hand before committing.
//
// Note on the interactive wizard: huh v1.0.0 accessible mode creates a new
// bufio.Scanner per prompt, so the first prompt swallows all piped stdin and
// every later field silently falls back to its default. The wizard therefore
// cannot be driven through a pipe; the alarm-profile path it shares with
// `create --alarm profile:<name>` is covered through that flag instead.
package golden

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var (
	uidRe     = regexp.MustCompile(`(?m)^UID:.*$`)
	dtstampRe = regexp.MustCompile(`(?m)^DTSTAMP:.*$`)
	createdRe = regexp.MustCompile(`(?m)^CREATED:.*$`)
	modifRe   = regexp.MustCompile(`(?m)^LAST-MODIFIED:.*$`)
)

// RepoRoot returns the repository root, derived from this source file's
// location (internal/testutil/golden → three levels up).
func RepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	return root
}

// BuildBinary compiles the tempus binary once into dir and returns its path.
// When GOLDEN_COVERDIR is set, the binary is built with -cover so CLI runs
// emit coverage counters into that directory (via GOCOVERDIR in RunCLI).
func BuildBinary(dir, repoRoot string) (string, error) {
	name := "tempus-golden"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(dir, name)
	args := []string{"build"}
	if os.Getenv("GOLDEN_COVERDIR") != "" {
		args = append(args, "-cover")
	}
	args = append(args, "-o", bin, ".")
	cmd := exec.Command("go", args...)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("building tempus: %v\n%s", err, out)
	}
	return bin, nil
}

// Result holds the outcome of one CLI invocation.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// RunCLI executes the binary with a hermetic environment: HOME points at an
// empty temp dir (no user config leaks), TZ is fixed to UTC, and all TEMPUS_*
// variables are scrubbed unless explicitly passed in extraEnv.
func RunCLI(t *testing.T, bin, workDir string, extraEnv []string, stdin string, args ...string) Result {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = workDir
	env := []string{
		"HOME=" + workDir,
		"USERPROFILE=" + workDir,
		"APPDATA=" + workDir,
		"TZ=UTC",
		"TERM=dumb",
		"PATH=" + os.Getenv("PATH"),
	}
	if coverDir := os.Getenv("GOLDEN_COVERDIR"); coverDir != "" {
		env = append(env, "GOCOVERDIR="+coverDir)
	}
	cmd.Env = append(env, extraEnv...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("running %s %v: %v", bin, args, err)
		}
	}
	return Result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}
}

// NormalizeICS replaces volatile fields (UID, DTSTAMP, CREATED,
// LAST-MODIFIED) with stable placeholders so goldens are deterministic.
// CRLF line endings are preserved — they are part of RFC 5545.
func NormalizeICS(ics string) string {
	ics = uidRe.ReplaceAllString(ics, "UID:NORMALIZED")
	ics = dtstampRe.ReplaceAllString(ics, "DTSTAMP:NORMALIZED")
	ics = createdRe.ReplaceAllString(ics, "CREATED:NORMALIZED")
	ics = modifRe.ReplaceAllString(ics, "LAST-MODIFIED:NORMALIZED")
	return ics
}

// writeFile is a small helper for tests that need fixture files.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}

// CompareGolden checks got against testdata/golden/<name>.ics. With -update,
// it (re)writes the golden file instead and fails the test as a reminder that
// goldens changed and the diff needs human review.
func CompareGolden(t *testing.T, update bool, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".ics")
	if update {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("creating golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("writing golden %s: %v", path, err)
		}
		t.Logf("golden %s updated — review the diff before committing", path)
		return
	}
	want, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("reading golden %s (run with -update to create it): %v", path, err)
	}
	if got != string(want) {
		t.Errorf("output differs from golden %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}
