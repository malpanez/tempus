package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func NewRRuleHelperCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "rrule",
		Short: "Interactive helper to build recurrence rules (RRULE)",
		Long: `Generate RRULE strings for recurring events without memorizing the syntax.

Examples of what you can create:
  - Every weekday (Monday-Friday)
  - Every 2 weeks on Tuesday and Thursday
  - Monthly on the 15th
  - Yearly on March 1st
  - Custom patterns with end dates or occurrence counts`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRRuleHelper(app, cmd, args)
		},
	}
}

func runRRuleHelper(app *App, _ *cobra.Command, _ []string) error {
	w := stdoutWriter(app)
	fmt.Fprintln(w, "RRULE Builder - Create recurring event patterns")
	fmt.Fprintln(w)

	freq, err := promptRRuleFrequency()
	if err != nil {
		return err
	}

	parts := []string{fmt.Sprintf("FREQ=%s", freq)}

	if interval := promptRRuleInterval(); interval != "" {
		parts = append(parts, interval)
	}

	if freq == "WEEKLY" {
		if days := promptRRuleWeeklyDays(); days != "" {
			parts = append(parts, days)
		}
	}

	if endCond := promptRRuleEndCondition(); endCond != "" {
		parts = append(parts, endCond)
	}

	rrule := strings.Join(parts, ";")

	fmt.Fprintln(w)
	PrintOK(w, "Generated RRULE:\n")
	fmt.Fprintln(w, rrule)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Usage examples:")
	fmt.Fprintf(w, "  CSV batch file:  rrule column = %s\n", rrule)
	fmt.Fprintf(w, "  JSON batch file: \"rrule\": \"%s\"\n", rrule)
	fmt.Fprintf(w, "  YAML batch file: rrule: %s\n", rrule)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "This means:")
	fmt.Fprintf(w, "  %s\n", interpretRRule(rrule))

	return nil
}

func interpretRRule(rrule string) string {
	parts := strings.Split(rrule, ";")
	var freq, interval, byday, count, until string

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, val := kv[0], kv[1]
		switch key {
		case "FREQ":
			freq = strings.ToLower(val)
		case "INTERVAL":
			interval = val
		case "BYDAY":
			byday = val
		case "COUNT":
			count = val
		case "UNTIL":
			until = val
		}
	}

	var result string

	if interval != "" && interval != "1" {
		result = fmt.Sprintf("Every %s %ss", interval, freq)
	} else {
		result = fmt.Sprintf("Every %s", freq)
	}

	if byday != "" {
		result += fmt.Sprintf(" on %s", byday)
	}

	if count != "" {
		result += fmt.Sprintf(", %s times", count)
	} else if until != "" {
		result += fmt.Sprintf(", until %s", until)
	} else {
		result += ", forever"
	}

	return result
}

func promptRRuleFrequency() (string, error) {
	fmt.Println("Select frequency:")
	fmt.Println("  1. Daily")
	fmt.Println("  2. Weekly")
	fmt.Println("  3. Monthly")
	fmt.Println("  4. Yearly")
	fmt.Print("Enter choice (1-4): ")

	var freqChoice int
	if _, err := fmt.Scanln(&freqChoice); err != nil || freqChoice < 1 || freqChoice > 4 {
		return "", fmt.Errorf("invalid choice")
	}

	frequencies := map[int]string{1: "DAILY", 2: "WEEKLY", 3: "MONTHLY", 4: "YEARLY"}
	return frequencies[freqChoice], nil
}

func promptRRuleInterval() string {
	fmt.Print("\nRepeat every N occurrences (default 1): ")
	var intervalStr string
	_, _ = fmt.Scanln(&intervalStr)
	intervalStr = strings.TrimSpace(intervalStr)
	if intervalStr != "" && intervalStr != "1" {
		interval := AtoiSafe(intervalStr)
		if interval > 0 {
			return fmt.Sprintf("INTERVAL=%d", interval)
		}
	}
	return ""
}

func promptRRuleWeeklyDays() string {
	fmt.Println("\nSelect days of week (comma-separated):")
	fmt.Println("  MO, TU, WE, TH, FR, SA, SU")
	fmt.Print("Days (e.g., 'MO,WE,FR' or leave empty for all): ")
	var daysStr string
	_, _ = fmt.Scanln(&daysStr)
	daysStr = strings.TrimSpace(daysStr)
	if daysStr != "" {
		return fmt.Sprintf("BYDAY=%s", strings.ToUpper(daysStr))
	}
	return ""
}

func promptRRuleEndCondition() string {
	fmt.Println("\nHow should the recurrence end?")
	fmt.Println("  1. Never (infinite)")
	fmt.Println("  2. After N occurrences")
	fmt.Println("  3. On a specific date")
	fmt.Print("Enter choice (1-3): ")

	var endChoice int
	if _, err := fmt.Scanln(&endChoice); err != nil || endChoice < 1 || endChoice > 3 {
		endChoice = 1
	}

	switch endChoice {
	case 2:
		fmt.Print("Number of occurrences: ")
		var countStr string
		_, _ = fmt.Scanln(&countStr)
		count := AtoiSafe(strings.TrimSpace(countStr))
		if count > 0 {
			return fmt.Sprintf("COUNT=%d", count)
		}
	case 3:
		fmt.Print("End date (YYYY-MM-DD): ")
		var untilStr string
		_, _ = fmt.Scanln(&untilStr)
		untilStr = strings.TrimSpace(untilStr)
		if untilStr != "" {
			if _, err := time.Parse("2006-01-02", untilStr); err == nil {
				untilStr = strings.ReplaceAll(untilStr, "-", "")
				return fmt.Sprintf("UNTIL=%s", untilStr)
			}
		}
	}
	return ""
}
