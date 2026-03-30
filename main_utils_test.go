package main

import (
	"strings"
	"tempus/internal/testutil"
	"testing"
)

// ============================================================================
// Utility function tests - covering 0% functions
// ============================================================================

// TestLevenshteinDistance, TestMin, TestStripEmoji, TestGenerateUID moved to internal/cli/nd_test.go

func TestAtoiSafe(t *testing.T) {
	// atoiSafe only handles positive integers, returns 0 for invalid/negative
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"valid positive", "123", 123},
		{"valid negative returns 0", "-45", 0}, // No negative support
		{"zero", "0", 0},
		{"invalid", "abc", 0},
		{"empty", "", 0},
		{"float", "3.14", 0},
		{"spaces with number", "  42  ", 42}, // Trims spaces
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := atoiSafe(tt.s)
			if got != tt.want {
				t.Errorf("atoiSafe(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestPrintErr(t *testing.T) {
	// This function prints to stderr, so we just test it doesn't panic
	printErr("test error message")
	printErr("")
	printErr("error with special chars: 💊 😀")
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", testutil.EventTitleHelloWorld, testutil.TemplateHelloWorld},
		{"uppercase", "TEST STRING", "test-string"},
		{"with underscores", "test_name_here", "test-name-here"},
		{"multiple spaces", "hello    world", testutil.TemplateHelloWorld},
		{"leading/trailing spaces", "  hello world  ", testutil.TemplateHelloWorld},
		{"special chars", "hello@world!test", "hello-world-test"},
		{"numbers", "test123", "test123"},
		{"mixed", "Test_123 Hello!", "test-123-hello"},
		{"empty", "", ""},
		{"only special chars", "@#$%", "event"}, // Returns "event" for empty/special-only
		{"hyphen already exists", testutil.TemplateHelloWorld, testutil.TemplateHelloWorld},
		{"consecutive hyphens", "hello--world", testutil.TemplateHelloWorld},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestDetectEventConflicts, TestDetectOverwhelmDays, TestGeneratePrepTimeEvents,
// TestGetSmartDefaultDuration, TestAddEmojiToSummary moved to internal/cli/nd_test.go

// ============================================================================
// Command creation tests for 0% coverage commands
// ============================================================================

func TestNewTimezoneCmd(t *testing.T) {
	cmd := newTimezoneCmd()
	if cmd == nil {
		t.Fatal("newTimezoneCmd() returned nil")
	}
	if cmd.Use != "timezone" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "timezone")
	}

	// Check subcommands
	subcommands := cmd.Commands()
	if len(subcommands) != 2 {
		t.Errorf("expected 2 subcommands, got %d", len(subcommands))
	}

	var hasList, hasInfo bool
	for _, sub := range subcommands {
		if strings.HasPrefix(sub.Use, "list") {
			hasList = true
		}
		if strings.HasPrefix(sub.Use, "info") {
			hasInfo = true
		}
	}
	if !hasList {
		t.Error("timezone command missing 'list' subcommand")
	}
	if !hasInfo {
		t.Error("timezone command missing 'info' subcommand")
	}
}

func TestNewRRuleHelperCmd(t *testing.T) {
	cmd := newRRuleHelperCmd()
	if cmd == nil {
		t.Fatal("newRRuleHelperCmd() returned nil")
	}
	if cmd.Use != "rrule" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "rrule")
	}
	if cmd.RunE == nil {
		t.Error("rrule command should have RunE function")
	}
}

func TestNewLocaleCmd(t *testing.T) {
	cmd := newLocaleCmd()
	if cmd == nil {
		t.Fatal("newLocaleCmd() returned nil")
	}
	if cmd.Use != "locale" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "locale")
	}

	// Check subcommands
	subcommands := cmd.Commands()
	if len(subcommands) != 1 {
		t.Errorf("expected 1 subcommand, got %d", len(subcommands))
	}

	listCmd := subcommands[0]
	if !strings.HasPrefix(listCmd.Use, "list") {
		t.Error("locale command should have 'list' subcommand")
	}
}

func TestNewTemplateCmd(t *testing.T) {
	cmd := newTemplateCmd()
	if cmd == nil {
		t.Fatal("newTemplateCmd() returned nil")
	}
	if cmd.Use != "template" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "template")
	}

	// Check subcommands exist
	subcommands := cmd.Commands()
	if len(subcommands) < 5 {
		t.Errorf("expected at least 5 subcommands, got %d", len(subcommands))
	}

	var hasList, hasDescribe, hasCreate, hasValidate, hasInit bool
	for _, sub := range subcommands {
		use := sub.Use
		if strings.HasPrefix(use, "list") {
			hasList = true
		}
		if strings.HasPrefix(use, "describe") {
			hasDescribe = true
		}
		if strings.HasPrefix(use, "create") {
			hasCreate = true
		}
		if strings.HasPrefix(use, "validate") {
			hasValidate = true
		}
		if strings.HasPrefix(use, "init") {
			hasInit = true
		}
	}

	if !hasList {
		t.Error("template command missing 'list' subcommand")
	}
	if !hasDescribe {
		t.Error("template command missing 'describe' subcommand")
	}
	if !hasCreate {
		t.Error("template command missing 'create' subcommand")
	}
	if !hasValidate {
		t.Error("template command missing 'validate' subcommand")
	}
	if !hasInit {
		t.Error("template command missing 'init' subcommand")
	}
}

// ============================================================================
// Batch template functions (0% coverage)
// ============================================================================

func TestGetBasicTemplate(t *testing.T) {
	content := getBasicTemplate()
	if content == "" {
		t.Error("getBasicTemplate() returned empty string")
	}
	if !strings.Contains(content, "summary") {
		t.Error("basic template should contain 'summary' field")
	}
	if !strings.Contains(content, "start") {
		t.Error("basic template should contain 'start' field")
	}
}

func TestGetADHDRoutineTemplate(t *testing.T) {
	content := getADHDRoutineTemplate()
	if content == "" {
		t.Error("getADHDRoutineTemplate() returned empty string")
	}
	// Should contain typical ADHD routine events
	if !strings.Contains(content, "Morning") && !strings.Contains(content, "medication") {
		t.Error("ADHD routine template should contain morning or medication events")
	}
}

func TestGetMedicationTemplate(t *testing.T) {
	content := getMedicationTemplate()
	if content == "" {
		t.Error("getMedicationTemplate() returned empty string")
	}
	if !strings.Contains(content, "medication") && !strings.Contains(content, "Medication") {
		t.Error("medication template should contain medication-related content")
	}
}

func TestGetWorkMeetingsTemplate(t *testing.T) {
	content := getWorkMeetingsTemplate()
	if content == "" {
		t.Error("getWorkMeetingsTemplate() returned empty string")
	}
	if !strings.Contains(content, "meeting") && !strings.Contains(content, "Meeting") {
		t.Error("work meetings template should contain meeting-related content")
	}
}

func TestGetMedicalTemplate(t *testing.T) {
	content := getMedicalTemplate()
	if content == "" {
		t.Error("getMedicalTemplate() returned empty string")
	}
	if !strings.Contains(content, "appointment") && !strings.Contains(content, "medical") {
		t.Error("medical template should contain appointment or medical content")
	}
}

func TestGetTravelTemplate(t *testing.T) {
	content := getTravelTemplate()
	if content == "" {
		t.Error("getTravelTemplate() returned empty string")
	}
	// Travel template might contain flight or travel-related terms
	if !strings.Contains(content, "flight") && !strings.Contains(content, "Flight") &&
		!strings.Contains(content, "travel") && !strings.Contains(content, "Travel") {
		t.Error("travel template should contain flight or travel-related content")
	}
}

func TestGetFamilyTemplate(t *testing.T) {
	content := getFamilyTemplate()
	if content == "" {
		t.Error("getFamilyTemplate() returned empty string")
	}
	// Family template might contain family-related events
}

func TestGetBatchTemplateContent(t *testing.T) {
	tests := []struct {
		name        string
		templateKey string
		wantEmpty   bool
		wantErr     bool
	}{
		{"basic", "basic", false, false},
		{"adhd-routine", "adhd-routine", false, false},
		{"medication", "medication", false, false},
		{"work-meetings", "work-meetings", false, false},
		{"medical", "medical", false, false},
		{"travel", "travel", false, false},
		{"family", "family", false, false},
		{"unknown", "unknown-template", true, true},
		{"empty", "", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBatchTemplateContent(tt.templateKey, "csv")
			isEmpty := got == ""
			hasErr := err != nil
			if isEmpty != tt.wantEmpty {
				t.Errorf("getBatchTemplateContent(%q) isEmpty = %v, want %v",
					tt.templateKey, isEmpty, tt.wantEmpty)
			}
			if hasErr != tt.wantErr {
				t.Errorf("getBatchTemplateContent(%q) hasErr = %v, want %v",
					tt.templateKey, hasErr, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// RRULE interpretation (0% coverage)
// ============================================================================

func TestInterpretRRule(t *testing.T) {
	tests := []struct {
		name      string
		rrule     string
		wantEmpty bool
	}{
		{"daily", "FREQ=DAILY", false},
		{"weekly", "FREQ=WEEKLY;BYDAY=MO,WE,FR", false},
		{"monthly", "FREQ=MONTHLY;BYMONTHDAY=15", false},
		{"yearly", "FREQ=YEARLY", false},
		{"with count", "FREQ=DAILY;COUNT=10", false},
		{"with until", "FREQ=WEEKLY;UNTIL=20251231", false},
		{"complex", "FREQ=MONTHLY;BYDAY=1MO;COUNT=12", false},
		{"empty", "", false},          // Returns default message even for empty
		{"invalid", "INVALID", false}, // Should still return something
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpretRRule(tt.rrule)
			isEmpty := got == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("interpretRRule(%q) isEmpty = %v, want %v, got %q",
					tt.rrule, isEmpty, tt.wantEmpty, got)
			}
		})
	}
}
