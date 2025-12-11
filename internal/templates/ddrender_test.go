package templates

import (
	"strings"
	"testing"
	"time"
)

// TestRenderTmpl tests template rendering
func TestRenderTmpl(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name     string
		tmpl     string
		values   map[string]string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple variable replacement",
			tmpl:     "Hello {{name}}",
			values:   map[string]string{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "multiple variables",
			tmpl:     "{{first}} {{last}}",
			values:   map[string]string{"first": "John", "last": "Doe"},
			expected: "John Doe",
		},
		{
			name:     "slug helper",
			tmpl:     "{{slug title}}",
			values:   map[string]string{"title": "Hello World!"},
			expected: "hello-world",
		},
		{
			name:     "date helper",
			tmpl:     "{{date start}}",
			values:   map[string]string{"start": "2025-12-01 10:00"},
			expected: "2025-12-01",
		},
		{
			name:     "conditional block - value present",
			tmpl:     "{{#name}}Hello {{name}}{{/name}}",
			values:   map[string]string{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "conditional block - value absent",
			tmpl:     "{{#name}}Hello {{name}}{{/name}}",
			values:   map[string]string{},
			expected: "",
		},
		{
			name:     "conditional block - empty value",
			tmpl:     "{{#name}}Hello {{name}}{{/name}}",
			values:   map[string]string{"name": ""},
			expected: "",
		},
		{
			name:     "mismatched conditional tags",
			tmpl:     "{{#name}}Hello{{/other}}",
			values:   map[string]string{"name": "World"},
			expected: "{{#name}}Hello{{/other}}", // Should remain unchanged
		},
		{
			name:     "nested variables in conditional",
			tmpl:     "{{#title}}Title: {{title}}, Author: {{author}}{{/title}}",
			values:   map[string]string{"title": "Book", "author": "Smith"},
			expected: "Title: Book, Author: Smith",
		},
		{
			name:     "empty template",
			tmpl:     "",
			values:   map[string]string{"name": "World"},
			expected: "",
		},
		{
			name:     "no variables",
			tmpl:     "Static text",
			values:   map[string]string{"name": "World"},
			expected: "Static text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTmpl(tt.tmpl, tt.values, tr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("RenderTmpl() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

// TestSimpleReplace tests the simpleReplace function
func TestSimpleReplace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		values   map[string]string
		expected string
	}{
		{
			name:     "basic replacement",
			input:    "{{name}}",
			values:   map[string]string{"name": "John"},
			expected: "John",
		},
		{
			name:     "slug function",
			input:    "{{slug title}}",
			values:   map[string]string{"title": "Hello World"},
			expected: "hello-world",
		},
		{
			name:     "date function",
			input:    "{{date when}}",
			values:   map[string]string{"when": "2025-12-01"},
			expected: "2025-12-01",
		},
		{
			name:     "date with time",
			input:    "{{date when}}",
			values:   map[string]string{"when": "2025-12-01 14:30"},
			expected: "2025-12-01",
		},
		{
			name:     "multiple replacements",
			input:    "{{first}}-{{last}}",
			values:   map[string]string{"first": "John", "last": "Doe"},
			expected: "John-Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simpleReplace(tt.input, tt.values)
			if result != tt.expected {
				t.Errorf("simpleReplace() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestExtractDate tests date extraction
func TestExtractDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "date only",
			input:    "2025-12-01",
			expected: "2025-12-01",
		},
		{
			name:     "datetime",
			input:    "2025-12-01 14:30",
			expected: "2025-12-01",
		},
		{
			name:     "datetime with T separator",
			input:    "2025-12-01T14:30",
			expected: "2025-12-01",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "too short",
			input:    "2025",
			expected: "2025", // Falls back to slugify
		},
		{
			name:     "invalid date",
			input:    "not-a-date",
			expected: "not-a-date", // Falls back to slugify
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDate(tt.input)
			if result != tt.expected {
				t.Errorf("extractDate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestParseDateOrDateTimeInLocation tests date/time parsing
func TestParseDateOrDateTimeInLocation(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		tzName       string
		wantDateOnly bool
		wantErr      bool
	}{
		{
			name:         "date only",
			input:        "2025-12-01",
			tzName:       "",
			wantDateOnly: true,
			wantErr:      false,
		},
		{
			name:         "date with time",
			input:        "2025-12-01 14:30",
			tzName:       "",
			wantDateOnly: false,
			wantErr:      false,
		},
		{
			name:         "date with timezone",
			input:        "2025-12-01",
			tzName:       "America/New_York",
			wantDateOnly: true,
			wantErr:      false,
		},
		{
			name:         "datetime with timezone",
			input:        "2025-12-01 14:30",
			tzName:       "America/New_York",
			wantDateOnly: false,
			wantErr:      false,
		},
		{
			name:    "invalid date",
			input:   "not-a-date",
			tzName:  "",
			wantErr: true,
		},
		{
			name:         "invalid timezone",
			input:        "2025-12-01",
			tzName:       "Invalid/Timezone",
			wantDateOnly: true, // Falls back to local, still date only
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, isDateOnly, err := parseDateOrDateTimeInLocation(tt.input, tt.tzName)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if isDateOnly != tt.wantDateOnly {
					t.Errorf("isDateOnly = %v, want %v", isDateOnly, tt.wantDateOnly)
				}
				if result.IsZero() {
					t.Error("result time is zero")
				}
			}
		})
	}
}

// TestParseHumanDuration tests duration parsing
func TestParseHumanDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "plain number (minutes)",
			input:    "60",
			expected: 60 * time.Minute,
		},
		{
			name:     "minutes with m",
			input:    "45m",
			expected: 45 * time.Minute,
		},
		{
			name:     "hours",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "hours and minutes",
			input:    "1h30m",
			expected: 90 * time.Minute,
		},
		{
			name:     "ISO format minutes",
			input:    "PT45M",
			expected: 45 * time.Minute,
		},
		{
			name:     "ISO format hours",
			input:    "PT2H",
			expected: 2 * time.Hour,
		},
		{
			name:     "ISO format hours and minutes",
			input:    "PT1H30M",
			expected: 90 * time.Minute,
		},
		{
			name:     "word 'minutes'",
			input:    "30 minutes",
			expected: 30 * time.Minute,
		},
		{
			name:     "word 'mins'",
			input:    "45 mins",
			expected: 45 * time.Minute,
		},
		{
			name:     "word 'minute'",
			input:    "1 minute",
			expected: 1 * time.Minute,
		},
		{
			name:     "word 'min'",
			input:    "15min",
			expected: 15 * time.Minute,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:     "zero duration",
			input:    "0",
			expected: 0,
			wantErr:  false, // 0 is parsed but rejected later in usage
		},
		{
			name:    "zero hours and minutes",
			input:   "0h0m",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseHumanDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseHumanDuration(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestParseHhMmCompact tests the compact duration parser
func TestParseHhMmCompact(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "just minutes",
			input:    "45m",
			expected: 45 * time.Minute,
		},
		{
			name:     "just hours",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "hours and minutes",
			input:    "1h30m",
			expected: 90 * time.Minute,
		},
		{
			name:     "hours only no m",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "zero duration",
			input:   "0m",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseHhMmCompact(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseHhMmCompact(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestParseDurationString tests the exported duration parser
func TestParseDurationString(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"60", 60 * time.Minute, false},
		{"1h30m", 90 * time.Minute, false},
		{"PT45M", 45 * time.Minute, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseDurationString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("ParseDurationString(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestSplitMultiValueList tests multi-value list splitting
func TestSplitMultiValueList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "comma separated",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "semicolon separated",
			input:    "a;b;c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "newline separated",
			input:    "a\nb\nc",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "mixed separators",
			input:    "a,b;c\nd",
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "with whitespace",
			input:    " a , b , c ",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty values",
			input:    "a,,b,,,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "CRLF line endings",
			input:    "a\r\nb\r\nc",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitMultiValueList(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestRenderDDToEvent tests data-driven event rendering
func TestRenderDDToEvent(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	tests := []struct {
		name    string
		dd      DataDrivenTemplate
		values  map[string]string
		wantErr bool
		check   func(*testing.T, interface{})
	}{
		{
			name: "basic timed event",
			dd: DataDrivenTemplate{
				Name: "test-event",
				Fields: []Field{
					{Key: "title", Name: "Title", Type: "text", Required: true},
					{Key: "start", Name: "Start", Type: "datetime", Required: true},
				},
				Output: OutputTemplate{
					StartField:  "start",
					SummaryTmpl: "{{title}}",
				},
			},
			values: map[string]string{
				"title": "Test Event",
				"start": "2025-12-01 10:00",
			},
			wantErr: false,
		},
		{
			name: "all-day event",
			dd: DataDrivenTemplate{
				Name: "all-day-event",
				Fields: []Field{
					{Key: "title", Name: "Title", Type: "text", Required: true},
					{Key: "date", Name: "Date", Type: "datetime", Required: true},
				},
				Output: OutputTemplate{
					StartField:  "date",
					SummaryTmpl: "{{title}}",
					AllDay:      true,
				},
			},
			values: map[string]string{
				"title": "All Day Event",
				"date":  "2025-12-01",
			},
			wantErr: false,
		},
		{
			name: "event with duration",
			dd: DataDrivenTemplate{
				Name: "duration-event",
				Fields: []Field{
					{Key: "title", Name: "Title", Type: "text", Required: true},
					{Key: "start", Name: "Start", Type: "datetime", Required: true},
					{Key: "duration", Name: "Duration", Type: "text", Required: false},
				},
				Output: OutputTemplate{
					StartField:    "start",
					DurationField: "duration",
					SummaryTmpl:   "{{title}}",
				},
			},
			values: map[string]string{
				"title":    "Meeting",
				"start":    "2025-12-01 10:00",
				"duration": "1h30m",
			},
			wantErr: false,
		},
		{
			name: "event with explicit end",
			dd: DataDrivenTemplate{
				Name: "end-event",
				Fields: []Field{
					{Key: "title", Name: "Title", Type: "text", Required: true},
					{Key: "start", Name: "Start", Type: "datetime", Required: true},
					{Key: "end", Name: "End", Type: "datetime", Required: false},
				},
				Output: OutputTemplate{
					StartField:  "start",
					EndField:    "end",
					SummaryTmpl: "{{title}}",
				},
			},
			values: map[string]string{
				"title": "Meeting",
				"start": "2025-12-01 10:00",
				"end":   "2025-12-01 11:30",
			},
			wantErr: false,
		},
		{
			name: "event with categories and priority",
			dd: DataDrivenTemplate{
				Name: "categorized-event",
				Fields: []Field{
					{Key: "title", Name: "Title", Type: "text", Required: true},
					{Key: "start", Name: "Start", Type: "datetime", Required: true},
				},
				Output: OutputTemplate{
					StartField:  "start",
					SummaryTmpl: "{{title}}",
					Categories:  []string{"Work", "Important"},
					Priority:    1,
				},
			},
			values: map[string]string{
				"title": "Important Meeting",
				"start": "2025-12-01 10:00",
			},
			wantErr: false,
		},
		{
			name: "event with timezones",
			dd: DataDrivenTemplate{
				Name: "tz-event",
				Fields: []Field{
					{Key: "title", Name: "Title", Type: "text", Required: true},
					{Key: "start", Name: "Start", Type: "datetime", Required: true},
					{Key: "tz", Name: "Timezone", Type: "timezone", Required: false},
				},
				Output: OutputTemplate{
					StartField:   "start",
					StartTZField: "tz",
					EndTZField:   "tz",
					SummaryTmpl:  "{{title}}",
				},
			},
			values: map[string]string{
				"title": "Meeting",
				"start": "2025-12-01 10:00",
				"tz":    "America/New_York",
			},
			wantErr: false,
		},
		{
			name: "missing required start field",
			dd: DataDrivenTemplate{
				Name: "no-start",
				Output: OutputTemplate{
					StartField:  "start",
					SummaryTmpl: "Test",
				},
			},
			values:  map[string]string{},
			wantErr: true,
		},
		{
			name: "invalid start time",
			dd: DataDrivenTemplate{
				Name: "bad-start",
				Output: OutputTemplate{
					StartField:  "start",
					SummaryTmpl: "Test",
				},
			},
			values: map[string]string{
				"start": "invalid-date",
			},
			wantErr: true,
		},
		{
			name: "invalid duration",
			dd: DataDrivenTemplate{
				Name: "bad-duration",
				Output: OutputTemplate{
					StartField:    "start",
					DurationField: "duration",
					SummaryTmpl:   "Test",
				},
			},
			values: map[string]string{
				"start":    "2025-12-01 10:00",
				"duration": "invalid",
			},
			wantErr: true,
		},
		{
			name: "zero duration",
			dd: DataDrivenTemplate{
				Name: "zero-duration",
				Output: OutputTemplate{
					StartField:    "start",
					DurationField: "duration",
					SummaryTmpl:   "Test",
				},
			},
			values: map[string]string{
				"start":    "2025-12-01 10:00",
				"duration": "0",
			},
			wantErr: true,
		},
		{
			name: "end before start",
			dd: DataDrivenTemplate{
				Name: "bad-end",
				Output: OutputTemplate{
					StartField:  "start",
					EndField:    "end",
					SummaryTmpl: "Test",
				},
			},
			values: map[string]string{
				"start": "2025-12-01 15:00",
				"end":   "2025-12-01 14:00",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := tm.renderDDToEvent(&tt.dd, tt.values, tr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("event is nil")
				}
				if tt.check != nil {
					var iface interface{} = event
					tt.check(t, iface)
				}
			}
		})
	}
}

// TestRenderDDToEventWithRecurrence tests recurrence rules
func TestRenderDDToEventWithRecurrence(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	dd := DataDrivenTemplate{
		Name: "recurring-event",
		Fields: []Field{
			{Key: "title", Name: "Title", Type: "text", Required: true},
			{Key: "start", Name: "Start", Type: "datetime", Required: true},
			{Key: "rrule", Name: "Recurrence", Type: "text", Required: false},
		},
		Output: OutputTemplate{
			StartField:  "start",
			RRuleField:  "rrule",
			SummaryTmpl: "{{title}}",
		},
	}

	values := map[string]string{
		"title": "Weekly Meeting",
		"start": "2025-12-01 10:00",
		"rrule": "FREQ=WEEKLY;BYDAY=MO",
	}

	event, err := tm.renderDDToEvent(&dd, values, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.RRule != "FREQ=WEEKLY;BYDAY=MO" {
		t.Errorf("RRule = %q, want %q", event.RRule, "FREQ=WEEKLY;BYDAY=MO")
	}
}

// TestParseDDExDates tests exception date parsing
func TestParseDDExDates(t *testing.T) {
	startTime := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		raw       string
		allDay    bool
		wantCount int
		wantErr   bool
	}{
		{
			name:      "single date",
			raw:       "2025-12-15",
			allDay:    false,
			wantCount: 1,
		},
		{
			name:      "multiple dates comma separated",
			raw:       "2025-12-15, 2025-12-22, 2025-12-29",
			allDay:    false,
			wantCount: 3,
		},
		{
			name:      "multiple dates newline separated",
			raw:       "2025-12-15\n2025-12-22\n2025-12-29",
			allDay:    false,
			wantCount: 3,
		},
		{
			name:      "dates with time",
			raw:       "2025-12-15 10:00, 2025-12-22 10:00",
			allDay:    false,
			wantCount: 2,
		},
		{
			name:      "all-day dates",
			raw:       "2025-12-15, 2025-12-22",
			allDay:    true,
			wantCount: 2,
		},
		{
			name:      "empty string",
			raw:       "",
			allDay:    false,
			wantCount: 0,
		},
		{
			name:    "invalid date",
			raw:     "not-a-date",
			allDay:  false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDDExDates(tt.raw, startTime, tt.allDay, "UTC")
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != tt.wantCount {
					t.Errorf("count = %d, want %d", len(result), tt.wantCount)
				}
			}
		})
	}
}

// TestRenderDDToEventWithAlarms tests alarm parsing
func TestRenderDDToEventWithAlarms(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	dd := DataDrivenTemplate{
		Name: "event-with-alarms",
		Fields: []Field{
			{Key: "title", Name: "Title", Type: "text", Required: true},
			{Key: "start", Name: "Start", Type: "datetime", Required: true},
			{Key: "alarms", Name: "Alarms", Type: "text", Required: false},
		},
		Output: OutputTemplate{
			StartField:  "start",
			AlarmsField: "alarms",
			SummaryTmpl: "{{title}}",
		},
	}

	tests := []struct {
		name      string
		alarms    string
		wantCount int
	}{
		{
			name:      "single alarm",
			alarms:    "-15m",
			wantCount: 1,
		},
		{
			name:      "multiple alarms",
			alarms:    "-15m, -1h, -2h",
			wantCount: 3,
		},
		{
			name:      "no alarms",
			alarms:    "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{
				"title":  "Test Event",
				"start":  "2025-12-01 10:00",
				"alarms": tt.alarms,
			}

			event, err := tm.renderDDToEvent(&dd, values, tr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(event.Alarms) != tt.wantCount {
				t.Errorf("alarm count = %d, want %d", len(event.Alarms), tt.wantCount)
			}
		})
	}
}

// TestRenderDDToEventWithExDates tests exception dates
func TestRenderDDToEventWithExDates(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	dd := DataDrivenTemplate{
		Name: "recurring-with-exceptions",
		Fields: []Field{
			{Key: "title", Name: "Title", Type: "text", Required: true},
			{Key: "start", Name: "Start", Type: "datetime", Required: true},
			{Key: "exdates", Name: "Exception Dates", Type: "text", Required: false},
		},
		Output: OutputTemplate{
			StartField:   "start",
			ExDatesField: "exdates",
			SummaryTmpl:  "{{title}}",
		},
	}

	values := map[string]string{
		"title":   "Weekly Event",
		"start":   "2025-12-01 10:00",
		"exdates": "2025-12-08, 2025-12-15",
	}

	event, err := tm.renderDDToEvent(&dd, values, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(event.ExDates) != 2 {
		t.Errorf("exdates count = %d, want 2", len(event.ExDates))
	}
}

// TestRenderDDToEventAllDayWithEndDate tests all-day event end date handling
func TestRenderDDToEventAllDayWithEndDate(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	dd := DataDrivenTemplate{
		Name: "all-day-multi",
		Fields: []Field{
			{Key: "title", Name: "Title", Type: "text", Required: true},
			{Key: "start", Name: "Start Date", Type: "datetime", Required: true},
			{Key: "end", Name: "End Date", Type: "datetime", Required: false},
		},
		Output: OutputTemplate{
			StartField:  "start",
			EndField:    "end",
			SummaryTmpl: "{{title}}",
			AllDay:      true,
		},
	}

	tests := []struct {
		name   string
		values map[string]string
	}{
		{
			name: "with end date",
			values: map[string]string{
				"title": "Vacation",
				"start": "2025-12-20",
				"end":   "2025-12-27",
			},
		},
		{
			name: "without end date",
			values: map[string]string{
				"title": "Single Day",
				"start": "2025-12-20",
			},
		},
		{
			name: "end with time (should normalize)",
			values: map[string]string{
				"title": "Vacation",
				"start": "2025-12-20",
				"end":   "2025-12-27 14:00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := tm.renderDDToEvent(&dd, tt.values, tr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !event.AllDay {
				t.Error("event should be all-day")
			}
			if !event.EndTime.After(event.StartTime) {
				t.Error("end time should be after start time")
			}
		})
	}
}

// TestRenderDDToEventWithTemplatedFields tests template rendering in fields
func TestRenderDDToEventWithTemplatedFields(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	dd := DataDrivenTemplate{
		Name: "templated-event",
		Fields: []Field{
			{Key: "name", Name: "Name", Type: "text", Required: true},
			{Key: "location", Name: "Location", Type: "text", Required: false},
			{Key: "start", Name: "Start", Type: "datetime", Required: true},
		},
		Output: OutputTemplate{
			StartField:      "start",
			SummaryTmpl:     "Meeting with {{name}}",
			LocationTmpl:    "{{location}}",
			DescriptionTmpl: "Attendee: {{name}}\nLocation: {{location}}",
		},
	}

	values := map[string]string{
		"name":     "John Doe",
		"location": "Conference Room A",
		"start":    "2025-12-01 10:00",
	}

	event, err := tm.renderDDToEvent(&dd, values, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(event.Summary, "John Doe") {
		t.Error("summary should contain name")
	}
	if event.Location != "Conference Room A" {
		t.Errorf("location = %q, want %q", event.Location, "Conference Room A")
	}
	if !strings.Contains(event.Description, "John Doe") {
		t.Error("description should contain name")
	}
}
