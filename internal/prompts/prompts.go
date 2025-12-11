package prompts

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Scanner is the global scanner for reading user input
var Scanner *bufio.Scanner

func init() {
	Scanner = bufio.NewScanner(os.Stdin)
}

// Input prompts the user for input with an optional default value.
// If the scanner encounters an error or EOF, it handles it gracefully.
func Input(prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	if !Scanner.Scan() {
		// Check for scanner error
		if err := Scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error reading input: %v\n", err)
			os.Exit(1)
		}
		// EOF reached (Ctrl+D / end of input)
		if defaultValue != "" {
			return defaultValue
		}
		return ""
	}

	input := strings.TrimSpace(Scanner.Text())

	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}

// InputWithHelp prompts the user for input and shows help if they type '?'
func InputWithHelp(prompt, defaultValue string, helpFn func()) string {
	for {
		input := Input(prompt, defaultValue)
		if input == "?" {
			helpFn()
			continue
		}
		return input
	}
}

// Confirm asks a yes/no question and returns true if yes
func Confirm(prompt string) bool {
	response := strings.ToLower(strings.TrimSpace(Input(prompt+" (y/n)", "")))
	return response == "y" || response == "yes" || response == "s" || response == "si"
}

// Choose presents a numbered list and lets the user select one
func Choose(prompt string, options []string) (int, string) {
	fmt.Printf("\n%s\n", prompt)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}

	for {
		input := Input("Select option", "")
		if input == "" {
			return -1, ""
		}

		var choice int
		_, err := fmt.Sscanf(input, "%d", &choice)
		if err == nil && choice >= 1 && choice <= len(options) {
			return choice - 1, options[choice-1]
		}

		fmt.Println("Invalid choice, please try again.")
	}
}

// MultiInput collects multiple lines of input until empty line or max reached
func MultiInput(prompt string, max int) []string {
	fmt.Printf("\n%s (empty line to finish, max %d)\n", prompt, max)

	var results []string
	for i := 0; i < max; i++ {
		input := Input(fmt.Sprintf("  %d)", i+1), "")
		if input == "" {
			break
		}
		results = append(results, input)
	}

	return results
}
