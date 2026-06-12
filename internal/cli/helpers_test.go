package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"tempus/internal/i18n"
	"tempus/internal/parsing"
	"tempus/internal/testutil"
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
	missing := filepath.Join(t.TempDir(), "nonexistent-unique-test-file.ics")
	if got := EnsureUniquePath(missing); got != missing {
		t.Errorf("got %q, want original path for nonexistent file", got)
	}

	t.Run("existing file gets suffix", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "test.ics")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := EnsureUniquePath(path)
		expected := filepath.Join(tmp, "test-2.ics")
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("multiple existing files", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "test.ics")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		path2 := filepath.Join(tmp, "test-2.ics")
		if err := os.WriteFile(path2, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := EnsureUniquePath(path)
		expected := filepath.Join(tmp, "test-3.ics")
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("no extension", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "test")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := EnsureUniquePath(path)
		expected := filepath.Join(tmp, "test-2.ics")
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

func TestValueAsStringSliceEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{"int slice", []interface{}{1, 2, 3}, []string{"1", "2", "3"}},
		{"mixed types", []interface{}{"a", 1, true, 2.5}, []string{"a", "1", "true", "2.5"}},
		{"float numbers", []interface{}{1.5, 2.0, 3.7}, []string{"1.5", "2", "3.7"}},
		{"with empty strings in slice", []interface{}{"a", "", "b", "  "}, []string{"a", "b"}},
		{"int to string", 123, []string{"123"}},
		{"bool to string", true, []string{"true"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValueAsStringSlice(tt.input)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("ValueAsStringSlice(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValueAsAlarmSliceComplexInputs(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int
	}{
		{"single alarm", "15m", 1},
		{"multiple alarms in string", []interface{}{"15m\n30m", "1h"}, 3},
		{"complex alarm specs", []string{"trigger=-15m,description=Test", "20m"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValueAsAlarmSlice(tt.input)
			if len(got) != tt.want {
				t.Errorf("ValueAsAlarmSlice(%v) returned %d items, want %d", tt.input, len(got), tt.want)
			}
		})
	}
}

func TestParseDateTimeWithTZ(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		timeStr string
		tz      string
		wantErr bool
	}{
		{"date only UTC", testutil.Date20250501, "", "", false},
		{"datetime UTC", testutil.Date20250501, "14:30", "", false},
		{testutil.TestNameWithTimezone, testutil.Date20250501, "14:30", testutil.TZEuropeMadrid, false},
		{"invalid timezone", testutil.Date20250501, "14:30", "Invalid/Zone", true},
		{"invalid date", "invalid", "14:30", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsing.ParseDateTimeWithTZ(tt.dateStr, tt.timeStr, tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTimeWithTZ(%q, %q, %q) error = %v, wantErr %v",
					tt.dateStr, tt.timeStr, tt.tz, err, tt.wantErr)
			}
		})
	}
}

func TestAddDurationToStart(t *testing.T) {
	tests := []struct {
		name      string
		start     string
		tz        string
		duration  time.Duration
		wantEmpty bool
	}{
		{"valid datetime", testutil.DateTime20250501_1000, "", 30 * time.Minute, false},
		{testutil.TestNameWithTimezone, testutil.DateTime20250501_1000, testutil.TZEuropeMadrid, 1 * time.Hour, false},
		{"invalid datetime", "invalid", "", 30 * time.Minute, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsing.AddDurationToStart(tt.start, tt.tz, tt.duration)
			isEmpty := got == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("AddDurationToStart(%q, %q, %v) empty = %v, want %v",
					tt.start, tt.tz, tt.duration, isEmpty, tt.wantEmpty)
			}
		})
	}
}

func TestParseExDateValues(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		tz      string
		allDay  bool
		wantLen int
		wantErr bool
	}{
		{"single date", []string{testutil.Date20250501}, "UTC", true, 1, false},
		{"multiple dates", []string{testutil.Date20250501, "2025-05-02", testutil.Date20250503}, "UTC", true, 3, false},
		{"datetime values", []string{testutil.DateTime20250501_1000, "2025-05-02 14:00"}, testutil.TZEuropeMadrid, false, 2, false},
		{"mixed date and datetime", []string{testutil.Date20250501, "2025-05-02 10:00"}, "UTC", false, 2, false},
		{"with T separator", []string{"2025-05-01T10:00"}, "UTC", false, 1, false},
		{"empty values", []string{"", "  "}, "UTC", true, 0, false},
		{"invalid date", []string{"invalid"}, "UTC", true, 0, true},
		{testutil.TestNameWithTimezone, []string{testutil.DateTime20250501_1000}, testutil.TZAmericaNewYork, false, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsing.ParseExDateValues(tt.values, tt.tz, tt.allDay)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExDateValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("ParseExDateValues() returned %d dates, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestNormalizeClockOnlyDateTimes(t *testing.T) {
	tests := []struct {
		name      string
		values    map[string]string
		startKey  string
		endKey    string
		tzKey     string
		checkFunc func(*testing.T, map[string]string)
	}{
		{
			name: "clock only start",
			values: map[string]string{
				"start": "14:30",
				"tz":    testutil.TZEuropeMadrid,
			},
			startKey: "start",
			endKey:   "end",
			tzKey:    "tz",
			checkFunc: func(t *testing.T, m map[string]string) {
				if !strings.Contains(m["start"], " 14:30") {
					t.Errorf("expected start to have time component, got %q", m["start"])
				}
			},
		},
		{
			name: "clock only end but might be duration",
			values: map[string]string{
				"start": testutil.DateTime20250501_1400,
				"end":   "15:30",
				"tz":    "UTC",
			},
			startKey: "start",
			endKey:   "end",
			tzKey:    "tz",
			checkFunc: func(t *testing.T, m map[string]string) {
				if m["end"] == "" {
					t.Error("end should not be empty")
				}
			},
		},
		{
			name: "no normalization needed",
			values: map[string]string{
				"start": testutil.DateTime20250501_1400,
				"end":   "2025-05-01 15:00",
			},
			startKey: "start",
			endKey:   "end",
			tzKey:    "tz",
			checkFunc: func(t *testing.T, m map[string]string) {
				if m["start"] != testutil.DateTime20250501_1400 {
					t.Errorf("start should not be modified, got %q", m["start"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsing.NormalizeClockOnlyDateTimes(tt.values, tt.startKey, tt.endKey, tt.tzKey)
			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.values)
			}
		})
	}
}

func TestNormalizeEndTimeFromDuration(t *testing.T) {
	tests := []struct {
		name            string
		values          map[string]string
		startKey        string
		endKey          string
		durationKey     string
		tzKey           string
		defaultDuration string
		checkFunc       func(*testing.T, map[string]string)
	}{
		{
			name: "use duration field",
			values: map[string]string{
				"start":    testutil.DateTime20250501_1000,
				"duration": "90m",
			},
			startKey:        "start",
			endKey:          "end",
			durationKey:     "duration",
			tzKey:           "tz",
			defaultDuration: "30m",
			checkFunc: func(t *testing.T, m map[string]string) {
				if m["end"] == "" {
					t.Error("end should be set from duration")
				}
			},
		},
		{
			name: "use default duration",
			values: map[string]string{
				"start": testutil.DateTime20250501_1000,
			},
			startKey:        "start",
			endKey:          "end",
			durationKey:     "duration",
			tzKey:           "tz",
			defaultDuration: "1h",
			checkFunc: func(t *testing.T, m map[string]string) {
				if m["end"] == "" {
					t.Error("end should be set from default duration")
				}
			},
		},
		{
			name: "end as duration",
			values: map[string]string{
				"start": testutil.DateTime20250501_1000,
				"end":   "45m",
			},
			startKey:        "start",
			endKey:          "end",
			durationKey:     "duration",
			tzKey:           "tz",
			defaultDuration: "30m",
			checkFunc: func(t *testing.T, m map[string]string) {
				if !strings.Contains(m["end"], testutil.Date20250501) {
					t.Errorf("end should be converted to datetime, got %q", m["end"])
				}
				if m["duration"] == "" {
					t.Error("duration should be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsing.NormalizeEndTimeFromDuration(tt.values, tt.startKey, tt.endKey, tt.durationKey, tt.tzKey, tt.defaultDuration)
			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.values)
			}
		})
	}
}

func TestNormalizeDateTimeInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"already normalized datetime", "2025-12-16 14:00", "2025-12-16 14:00"},
		{"already normalized date", "2025-12-16", "2025-12-16"},
		{"slash separators datetime", "2025/12/16 14:00", "2025-12-16 14:00"},
		{"slash separators date only", "2025/12/16", "2025-12-16"},
		{"missing leading zeros", "2025-1-5 9:00", "2025-01-05 09:00"},
		{"missing leading zeros date only", "2025-1-5", "2025-01-05"},
		{"time without colon", "2025-01-05 0900", "2025-01-05 09:00"},
		{"slash and missing zeros", "2025/1/5 9:00", "2025-01-05 09:00"},
		{"whitespace padding", "  2025-12-16 14:00  ", "2025-12-16 14:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsing.NormalizeDateTimeInput(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeDateTimeInput(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnsureUniquePathMultipleCollisions(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, testutil.FilenameEventICS)

	if err := os.WriteFile(basePath, []byte("test"), 0644); err != nil {
		t.Fatalf(testutil.ErrMsgFailedToCreateTestFile, err)
	}

	result := EnsureUniquePath(basePath)
	expected := filepath.Join(tmpDir, "event-2.ics")
	if result != expected {
		t.Errorf("EnsureUniquePath() = %q, want %q", result, expected)
	}

	if err := os.WriteFile(result, []byte("test"), 0644); err != nil {
		t.Fatalf(testutil.ErrMsgFailedToCreateTestFile, err)
	}

	result2 := EnsureUniquePath(basePath)
	expected2 := filepath.Join(tmpDir, "event-3.ics")
	if result2 != expected2 {
		t.Errorf("EnsureUniquePath() = %q, want %q", result2, expected2)
	}
}

func TestAlarmPromptI18nKeys(t *testing.T) {
	t.Skip("alarm_prompt_* keys not yet defined in locale files")
	keys := []string{
		"alarm_prompt_suggested",
		"alarm_prompt_keep_or_change",
		"alarm_prompt_yes_short",
		"alarm_prompt_yes_long",
		"alarm_prompt_add_up_to",
		"alarm_prompt_help_hint",
		"alarm_prompt_reminder_n",
		"alarm_prompt_examples_header",
		"alarm_prompt_example_before",
		"alarm_prompt_example_after",
		"alarm_prompt_example_trigger",
		"alarm_prompt_example_absolute",
		"alarm_prompt_optional_desc",
	}
	locales := []string{"en", "es", "pt", "ga"}
	for _, lang := range locales {
		tr, err := i18n.NewTranslator(lang)
		if err != nil {
			t.Fatalf("failed to create translator for %s: %v", lang, err)
		}
		for _, key := range keys {
			val := tr.T(key)
			if val == key || val == "" {
				t.Errorf("locale %s: key %q not translated (got %q)", lang, key, val)
			}
		}
	}
}
