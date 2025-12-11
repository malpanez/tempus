package utils

import (
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
		{"spaces to hyphens", "hello world", "hello-world"},
		{"multiple spaces", "hello   world", "hello-world"},
		{"special characters", "Meeting @ 3pm", "meeting-3pm"},
		{"mixed punctuation", "hello, world! 2024", "hello-world-2024"},
		{"underscores", "hello_world", "hello-world"},
		{"slashes", "hello/world\\test", "hello-world-test"},
		{"dots", "hello.world.test", "hello-world-test"},
		{"leading/trailing spaces", "  hello world  ", "hello-world"},
		{"leading/trailing hyphens", "-hello-world-", "hello-world"},
		{"only special chars", "@#$%^", "event"},
		{"empty string", "", ""},
		{"accented characters", "múltiple espacios", "m-ltiple-espacios"},
		{"numbers", "event 123 test 456", "event-123-test-456"},
		{"consecutive hyphens", "hello---world", "hello-world"},
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

func TestSlugify_EdgeCases(t *testing.T) {
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
	if !strings.Contains(result, "hello-world") {
		t.Error("Slugify should handle long strings")
	}
	if strings.HasPrefix(result, "-") || strings.HasSuffix(result, "-") {
		t.Error("Slugify should not have leading or trailing hyphens")
	}
}
