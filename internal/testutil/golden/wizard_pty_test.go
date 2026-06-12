//go:build !windows

package golden

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
)

// wizardStep pairs a prompt substring with the line to type when it appears.
type wizardStep struct {
	expect string
	send   string
}

// runWizardPTY drives the binary through a real pseudo-terminal. TERM=dumb
// forces huh's accessible (line-based) mode, and each answer is written only
// after its prompt has been printed — which sidesteps the huh v1.0 issue
// where the first prompt's scanner swallows all pre-buffered input (the gap
// documented in harness.go where bug B1 hid).
func runWizardPTY(t *testing.T, workDir string, steps []wizardStep, args ...string) (transcript string, exitErr error) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Dir = workDir
	cmd.Env = []string{
		"HOME=" + workDir,
		"TZ=UTC",
		"TERM=dumb",
		"PATH=" + os.Getenv("PATH"),
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("starting pty: %v", err)
	}
	defer ptmx.Close() //nolint:errcheck // test pty cleanup

	var mu sync.Mutex
	var out strings.Builder
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 4096)
		for {
			n, rerr := ptmx.Read(buf)
			if n > 0 {
				mu.Lock()
				out.Write(buf[:n])
				mu.Unlock()
			}
			if rerr != nil {
				close(done)
				return
			}
		}
	}()

	snapshot := func() string {
		mu.Lock()
		defer mu.Unlock()
		return out.String()
	}

	deadline := time.Now().Add(60 * time.Second)
	consumed := 0
	for _, step := range steps {
		for {
			if time.Now().After(deadline) {
				_ = cmd.Process.Kill()
				t.Fatalf("timeout waiting for prompt %q\ntranscript so far:\n%s", step.expect, snapshot())
			}
			text := snapshot()
			if idx := strings.Index(text[consumed:], step.expect); idx != -1 {
				consumed += idx + len(step.expect)
				break
			}
			time.Sleep(25 * time.Millisecond)
		}
		if _, err := ptmx.Write([]byte(step.send + "\n")); err != nil {
			t.Fatalf("writing %q: %v", step.send, err)
		}
	}

	waitErr := cmd.Wait()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	return snapshot(), waitErr
}

// TestWizardPTYCreatesEventWithProfileAlarms is the real-terminal E2E for
// the flagship flow: wizard + adhd-default profile must produce an ICS with
// the profile's VALARMs (bug B1's exact home) and the validated timezone.
func TestWizardPTYCreatesEventWithProfileAlarms(t *testing.T) {
	workDir := t.TempDir()
	outPath := filepath.Join(workDir, "wizard.ics")

	steps := []wizardStep{
		{"Event name", "Med Reminder"},
		{"Start date", "2030-05-10"},
		{"Start time", "10:30"},
		{"Duration", "45m"},
		{"All-day event?", "n"},
		{"Timezone", "Europe/Madrid"},
		{"Enter a number between 1 and", "1"}, // alarm profile: adhd-default
		{"Enter a number between 0 and", "0"}, // categories: confirm empty
		{"Location (optional)", ""},
		{"Description (optional)", ""},
		{"Create this event?", "y"},
	}

	transcript, err := runWizardPTY(t, workDir, steps, "create", "-i", "-o", outPath)
	if err != nil {
		t.Fatalf("wizard exited with error: %v\ntranscript:\n%s", err, transcript)
	}

	raw, err := os.ReadFile(filepath.Clean(outPath))
	if err != nil {
		t.Fatalf("wizard did not write %s: %v\ntranscript:\n%s", outPath, err, transcript)
	}
	ics := string(raw)

	if got := strings.Count(ics, "BEGIN:VALARM"); got != 4 {
		t.Errorf("adhd-default profile must emit 4 VALARMs, got %d\nics:\n%s", got, ics)
	}
	for _, want := range []string{
		"DTSTART;TZID=Europe/Madrid:20300510T103000",
		"BEGIN:VTIMEZONE",
		"TRIGGER:-PT2H",
		"SUMMARY:Med Reminder",
	} {
		if !strings.Contains(ics, want) {
			t.Errorf("wizard ICS missing %q", want)
		}
	}

	res := RunCLI(t, binPath, workDir, nil, "", "lint", outPath)
	if res.ExitCode != 0 {
		t.Errorf("wizard output fails own lint: %s%s", res.Stdout, res.Stderr)
	}
}

// TestWizardPTYCancelExitsNonZero: declining the final confirm reports
// cancellation (exit 0, nothing written); Ctrl+C mid-form exits non-zero.
func TestWizardPTYCtrlCExitsNonZero(t *testing.T) {
	workDir := t.TempDir()

	cmd := exec.Command(binPath, "create", "-i")
	cmd.Dir = workDir
	cmd.Env = []string{"HOME=" + workDir, "TZ=UTC", "TERM=dumb", "PATH=" + os.Getenv("PATH")}
	ptmx, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("starting pty: %v", err)
	}
	defer ptmx.Close() //nolint:errcheck // test pty cleanup

	time.Sleep(300 * time.Millisecond)
	if _, err := ptmx.Write([]byte{0x03}); err != nil { // ^C
		t.Fatalf("sending ^C: %v", err)
	}

	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()
	select {
	case werr := <-waitDone:
		if werr == nil {
			t.Error("Ctrl+C during the wizard must exit non-zero, got 0")
		}
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("wizard did not exit after Ctrl+C")
	}
}
