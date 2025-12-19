package utils

import (
	"tempus/internal/testutil"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "hello", "hello"},
		{"uppercase to lowercase", "HELLO", "hello"},
		{"spaces to hyphens", "hello world", testutil.TemplateHelloWorld},
		{"multiple spaces", "hello   world", testutil.TemplateHelloWorld},
		{"special characters", "Meeting @ 3pm", "meeting-3pm"},
		{"mixed punctuation", "hello, world! 2024", "hello-world-2024"},
		{"underscores", "hello_world", testutil.TemplateHelloWorld},
		{"slashes", "hello/world\\test", "hello-world-test"},
		{"dots", "hello.world.test", "hello-world-test"},
		{"leading/trailing spaces", "  hello world  ", testutil.TemplateHelloWorld},
		{"leading/trailing hyphens", "-hello-world-", testutil.TemplateHelloWorld},
		{"only special chars", "@#$%^", "event"},
		{testutil.TestNameEmptyString, "", ""},
		{"accented characters", "múltiple espacios", "m-ltiple-espacios"},
		{"numbers", "event 123 test 456", "event-123-test-456"},
		{"consecutive hyphens", "hello---world", testutil.TemplateHelloWorld},
		{"mixed case with numbers", "Event2024Test", "event2024test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSlugifyEdgeCases(t *testing.T) {
	// Test that single characters work
	if result := Slugify("a"); result != "a" {
		t.Errorf("Slugify('a') = %q, want 'a'", result)
	}

	// Test that numbers work
	if result := Slugify("123"); result != "123" {
		t.Errorf("Slugify('123') = %q, want '123'", result)
	}

	// Test very long string
	longString := strings.Repeat("hello world ", 100)
	result := Slugify(longString)
	if !strings.Contains(result, testutil.TemplateHelloWorld) {
		t.Error("Slugify should handle long strings")
	}
	if strings.HasPrefix(result, "-") || strings.HasSuffix(result, "-") {
		t.Error("Slugify should not have leading or trailing hyphens")
	}
}
