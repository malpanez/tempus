package prompts

import (
	"bufio"
	"strings"
	"testing"
)

func TestInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		prompt       string
		defaultValue string
		expected     string
	}{
		{
			name:         "user provides value",
			input:        "hello\n",
			prompt:       "Name",
			defaultValue: "",
			expected:     "hello",
		},
		{
			name:         "user provides empty with default",
			input:        "\n",
			prompt:       "Name",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "user provides value overrides default",
			input:        "custom\n",
			prompt:       "Name",
			defaultValue: "default",
			expected:     "custom",
		},
		{
			name:         "user provides whitespace",
			input:        "  spaced  \n",
			prompt:       "Name",
			defaultValue: "",
			expected:     "spaced",
		},
		{
			name:         "EOF with default",
			input:        "", // EOF
			prompt:       "Name",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "EOF without default",
			input:        "", // EOF
			prompt:       "Name",
			defaultValue: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and replace global scanner
			prevScanner := Scanner
			Scanner = bufio.NewScanner(strings.NewReader(tt.input))
			defer func() { Scanner = prevScanner }()

			result := Input(tt.prompt, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Input() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestInputWithHelp(t *testing.T) {
	t.Run("shows help on ?", func(t *testing.T) {
		input := "?\nactual answer\n"
		prevScanner := Scanner
		Scanner = bufio.NewScanner(strings.NewReader(input))
		defer func() { Scanner = prevScanner }()

		helpCalled := false
		helpFn := func() {
			helpCalled = true
		}

		result := InputWithHelp("Question", "", helpFn)

		if !helpCalled {
			t.Error("Help function was not called")
		}
		if result != "actual answer" {
			t.Errorf("InputWithHelp() = %q, want %q", result, "actual answer")
		}
	})

	t.Run("returns answer without help", func(t *testing.T) {
		input := "direct answer\n"
		prevScanner := Scanner
		Scanner = bufio.NewScanner(strings.NewReader(input))
		defer func() { Scanner = prevScanner }()

		helpCalled := false
		helpFn := func() {
			helpCalled = true
		}

		result := InputWithHelp("Question", "", helpFn)

		if helpCalled {
			t.Error("Help function was called when it shouldn't be")
		}
		if result != "direct answer" {
			t.Errorf("InputWithHelp() = %q, want %q", result, "direct answer")
		}
	})
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"yes lowercase", "y\n", true},
		{"yes full", "yes\n", true},
		{"yes uppercase", "Y\n", true},
		{"yes mixed case", "Yes\n", true},
		{"si spanish", "si\n", true},
		{"si uppercase", "Si\n", true},
		{"s short", "s\n", true},
		{"no", "n\n", false},
		{"no full", "no\n", false},
		{"empty", "\n", false},
		{"invalid", "maybe\n", false},
		{"whitespace yes", "  yes  \n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevScanner := Scanner
			Scanner = bufio.NewScanner(strings.NewReader(tt.input))
			defer func() { Scanner = prevScanner }()

			result := Confirm("Are you sure")
			if result != tt.expected {
				t.Errorf("Confirm() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChoose(t *testing.T) {
	options := []string{"Option A", "Option B", "Option C"}

	tests := []struct {
		name           string
		input          string
		expectedIndex  int
		expectedOption string
	}{
		{
			name:           "select first option",
			input:          "1\n",
			expectedIndex:  0,
			expectedOption: "Option A",
		},
		{
			name:           "select second option",
			input:          "2\n",
			expectedIndex:  1,
			expectedOption: "Option B",
		},
		{
			name:           "select third option",
			input:          "3\n",
			expectedIndex:  2,
			expectedOption: "Option C",
		},
		{
			name:           "empty input",
			input:          "\n",
			expectedIndex:  -1,
			expectedOption: "",
		},
		{
			name:           "invalid then valid",
			input:          "99\n2\n",
			expectedIndex:  1,
			expectedOption: "Option B",
		},
		{
			name:           "text then valid",
			input:          "abc\n3\n",
			expectedIndex:  2,
			expectedOption: "Option C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevScanner := Scanner
			Scanner = bufio.NewScanner(strings.NewReader(tt.input))
			defer func() { Scanner = prevScanner }()

			index, option := Choose("Select an option", options)

			if index != tt.expectedIndex {
				t.Errorf("Choose() index = %d, want %d", index, tt.expectedIndex)
			}
			if option != tt.expectedOption {
				t.Errorf("Choose() option = %q, want %q", option, tt.expectedOption)
			}
		})
	}
}

func TestMultiInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected []string
	}{
		{
			name:     "collect three items",
			input:    "first\nsecond\nthird\n\n",
			max:      5,
			expected: []string{"first", "second", "third"},
		},
		{
			name:     "stop at max",
			input:    "1\n2\n3\n4\n5\n6\n",
			max:      3,
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "empty line stops early",
			input:    "one\n\n",
			max:      5,
			expected: []string{"one"},
		},
		{
			name:     "no input",
			input:    "\n",
			max:      5,
			expected: []string{},
		},
		{
			name:     "whitespace trimmed",
			input:    "  first  \n  second  \n\n",
			max:      5,
			expected: []string{"first", "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevScanner := Scanner
			Scanner = bufio.NewScanner(strings.NewReader(tt.input))
			defer func() { Scanner = prevScanner }()

			result := MultiInput("Enter items", tt.max)

			if len(result) != len(tt.expected) {
				t.Errorf("MultiInput() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("MultiInput()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
