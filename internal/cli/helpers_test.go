package cli

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestParseBoolish(t *testing.T) {
	trueVals := []string{"1", "true", "yes", "y", "on", "TRUE", "Yes", "Y", "ON"}
	for _, v := range trueVals {
		if !ParseBoolish(v) {
			t.Errorf("ParseBoolish(%q) = false, want true", v)
		}
	}
	falseVals := []string{"0", "false", "no", "n", "off", "", "random"}
	for _, v := range falseVals {
		if ParseBoolish(v) {
			t.Errorf("ParseBoolish(%q) = true, want false", v)
		}
	}
}

func TestSplitDelimited(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"a,b,c", 3},
		{"a;b|c", 3},
		{"", 0},
		{"  ", 0},
		{"single", 1},
	}
	for _, tc := range tests {
		got := SplitDelimited(tc.input)
		if len(got) != tc.want {
			t.Errorf("SplitDelimited(%q) len = %d, want %d", tc.input, len(got), tc.want)
		}
	}
}

func TestExtractDate(t *testing.T) {
	if got := ExtractDate("2025-01-15 10:00"); got != "2025-01-15" {
		t.Errorf("got %q, want %q", got, "2025-01-15")
	}
	if got := ExtractDate("short"); got != "short" {
		t.Errorf("got %q, want %q", got, "short")
	}
}

func TestEnsureDirForFile(t *testing.T) {
	if err := EnsureDirForFile("file.txt"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := EnsureDirForFile("./file.txt"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	tmp := t.TempDir()
	if err := EnsureDirForFile(tmp + "/sub/dir/file.txt"); err != nil {
		t.Errorf("unexpected error creating nested dir: %v", err)
	}
	if _, err := os.Stat(tmp + "/sub/dir"); err != nil {
		t.Errorf("directory should have been created: %v", err)
	}
}

func TestValueAsString(t *testing.T) {
	if got := ValueAsString(nil); got != "" {
		t.Errorf("nil: got %q", got)
	}
	if got := ValueAsString("hello"); got != "hello" {
		t.Errorf("string: got %q", got)
	}
	if got := ValueAsString(42.5); got != "42.5" {
		t.Errorf("float: got %q", got)
	}
	if got := ValueAsString(true); got != "true" {
		t.Errorf("bool: got %q", got)
	}
	if got := ValueAsString(false); got != "false" {
		t.Errorf("bool false: got %q", got)
	}
	if got := ValueAsString(123); got != "123" {
		t.Errorf("int: got %q", got)
	}
}

func TestValueAsBool(t *testing.T) {
	if ValueAsBool(nil) {
		t.Error("nil should be false")
	}
	if !ValueAsBool(true) {
		t.Error("true should be true")
	}
	if ValueAsBool(float64(0)) {
		t.Error("0.0 should be false")
	}
	if !ValueAsBool(float64(1)) {
		t.Error("1.0 should be true")
	}
	if !ValueAsBool("yes") {
		t.Error("yes should be true")
	}
	if !ValueAsBool(1) {
		t.Error("int 1 should be true")
	}
}

func TestValueAsStringSlice(t *testing.T) {
	if got := ValueAsStringSlice(nil); got != nil {
		t.Error("nil should return nil")
	}
	got := ValueAsStringSlice([]interface{}{"a", "b"})
	if len(got) != 2 {
		t.Errorf("[]interface{}: len = %d, want 2", len(got))
	}
	got = ValueAsStringSlice([]string{"x", "y"})
	if len(got) != 2 {
		t.Errorf("[]string: len = %d, want 2", len(got))
	}
	got = ValueAsStringSlice("a,b,c")
	if len(got) != 3 {
		t.Errorf("string: len = %d, want 3", len(got))
	}
	got = ValueAsStringSlice(42)
	if len(got) != 1 {
		t.Errorf("int: len = %d, want 1", len(got))
	}
}

func TestValueAsAlarmSlice(t *testing.T) {
	if got := ValueAsAlarmSlice(nil); got != nil {
		t.Error("nil should return nil")
	}
	got := ValueAsAlarmSlice("-15m")
	if len(got) == 0 {
		t.Error("string alarm should return non-empty")
	}
	got = ValueAsAlarmSlice([]string{"-15m"})
	if len(got) == 0 {
		t.Error("[]string alarm should return non-empty")
	}
	got = ValueAsAlarmSlice([]interface{}{"-15m"})
	if len(got) == 0 {
		t.Error("[]interface{} alarm should return non-empty")
	}
	_ = ValueAsAlarmSlice(42)
	got = ValueAsAlarmSlice([]interface{}{"", "-10m"})
	if len(got) == 0 {
		t.Error("mixed alarm should return non-empty")
	}
	got = ValueAsAlarmSlice([]string{"", "-5m"})
	if len(got) == 0 {
		t.Error("mixed string alarm should return non-empty")
	}
}

func TestEnsureICSExtension(t *testing.T) {
	if got := EnsureICSExtension(""); got != "event.ics" {
		t.Errorf("empty: got %q", got)
	}
	if got := EnsureICSExtension("test"); got != "test.ics" {
		t.Errorf("no ext: got %q", got)
	}
	if got := EnsureICSExtension("test.ics"); got != "test.ics" {
		t.Errorf("with ext: got %q", got)
	}
	if got := EnsureICSExtension("test.ICS"); got != "test.ICS" {
		t.Errorf("uppercase ext: got %q", got)
	}
}

func TestEnsureUniquePath(t *testing.T) {
	got := EnsureUniquePath("/tmp/nonexistent-unique-test-file.ics")
	if got != "/tmp/nonexistent-unique-test-file.ics" {
		t.Errorf("got %q, want original path for nonexistent file", got)
	}

	t.Run("existing file gets suffix", func(t *testing.T) {
		tmp := t.TempDir()
		path := tmp + "/test.ics"
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := EnsureUniquePath(path)
		expected := tmp + "/test-2.ics"
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("multiple existing files", func(t *testing.T) {
		tmp := t.TempDir()
		path := tmp + "/test.ics"
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		path2 := tmp + "/test-2.ics"
		if err := os.WriteFile(path2, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := EnsureUniquePath(path)
		expected := tmp + "/test-3.ics"
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("no extension", func(t *testing.T) {
		tmp := t.TempDir()
		path := tmp + "/test"
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := EnsureUniquePath(path)
		expected := tmp + "/test-2.ics"
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})
}

func TestSlugify(t *testing.T) {
	got := Slugify("Hello World!")
	if got == "" {
		t.Error("slugify should return non-empty")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := FirstNonEmpty("", "", "third"); got != "third" {
		t.Errorf("got %q, want %q", got, "third")
	}
	if got := FirstNonEmpty("first", "second"); got != "first" {
		t.Errorf("got %q, want %q", got, "first")
	}
	if got := FirstNonEmpty("", ""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestAtoiSafe(t *testing.T) {
	if got := AtoiSafe("42"); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
	if got := AtoiSafe(""); got != 0 {
		t.Errorf("empty: got %d, want 0", got)
	}
	if got := AtoiSafe("abc"); got != 0 {
		t.Errorf("abc: got %d, want 0", got)
	}
}

func TestPrintOK(t *testing.T) {
	var buf bytes.Buffer
	PrintOK(&buf, "test %s", "msg")
	if buf.Len() == 0 {
		t.Error("PrintOK should write output")
	}
}

func TestPrintErr(t *testing.T) {
	var buf bytes.Buffer
	PrintErr(&buf, "test %s", "msg")
	if buf.Len() == 0 {
		t.Error("PrintErr should write output")
	}
}

func TestCleanDisplay(t *testing.T) {
	if got := CleanDisplay("Europe/Madrid (CET)"); got != "Europe/Madrid" {
		t.Errorf("got %q, want %q", got, "Europe/Madrid")
	}
	if got := CleanDisplay("simple"); got != "simple" {
		t.Errorf("got %q, want %q", got, "simple")
	}
}

func TestLooksLikeClock(t *testing.T) {
	if !LooksLikeClock("10:00") {
		t.Error("10:00 should look like clock")
	}
	if !LooksLikeClock("9:00") {
		t.Error("9:00 should look like clock")
	}
	if LooksLikeClock("2025-01-15 10:00") {
		t.Error("datetime should not look like clock")
	}
	if LooksLikeClock("") {
		t.Error("empty should not look like clock")
	}
}

func TestFmtDurationHuman(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0m"},
		{30 * time.Minute, "30m"},
		{1 * time.Hour, "1h"},
		{90 * time.Minute, "1h30m"},
		{-1 * time.Minute, "0m"},
	}
	for _, tc := range tests {
		if got := FmtDurationHuman(tc.d); got != tc.want {
			t.Errorf("FmtDurationHuman(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestSplitDateTime(t *testing.T) {
	d, tm := SplitDateTime("2025-01-15 10:00")
	if d != "2025-01-15" || tm != "10:00" {
		t.Errorf("got (%q, %q), want (%q, %q)", d, tm, "2025-01-15", "10:00")
	}
	d, tm = SplitDateTime("2025-01-15")
	if d != "2025-01-15" || tm != "" {
		t.Errorf("got (%q, %q), want (%q, %q)", d, tm, "2025-01-15", "")
	}
	d, tm = SplitDateTime("")
	if d != "" || tm != "" {
		t.Errorf("got (%q, %q), want (%q, %q)", d, tm, "", "")
	}
}

func TestPrependToday(t *testing.T) {
	got := PrependToday("10:00", "UTC")
	if got == "10:00" {
		t.Error("should prepend today's date")
	}
}
