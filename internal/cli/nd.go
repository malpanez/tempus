package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/config"

	"github.com/google/uuid"
)

func NormalizeAndSpellCheck(text string) string {
	if text == "" {
		return text
	}

	cfg, _ := config.Load()
	corrections := make(map[string]string)
	if cfg != nil && cfg.SpellCorrections != nil {
		corrections = cfg.SpellCorrections
	}

	words := strings.Fields(text)
	for i, word := range words {
		lower := strings.ToLower(word)
		if corrected, exists := corrections[lower]; exists {
			if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
				words[i] = strings.Title(corrected)
			} else {
				words[i] = corrected
			}
		}
	}

	return strings.Join(words, " ")
}

func ValidateCategoryWithSuggestion(category string) string {
	commonCategories := map[string]string{
		"work":          "Work",
		"meeting":       "Meeting",
		"health":        "Health",
		"medication":    "Medication",
		"meds":          "Medication",
		"medical":       "Medical",
		"therapy":       "Therapy",
		"mental health": "Mental Health",
		"exercise":      "Exercise",
		"workout":       "Workout",
		"food":          "Food",
		"meal":          "Meal",
		"travel":        "Travel",
		"flight":        "Flight",
		"hotel":         "Accommodation",
		"accommodation": "Accommodation",
		"family":        "Family",
		"kids":          "Kids",
		"personal":      "Personal",
		"focus":         "Focus",
		"deep work":     "Focus",
		"break":         "Break",
		"rest":          "Rest",
		"transition":    "Transition",
		"urgent":        "Urgent",
		"important":     "Important",
		"fun":           "Fun",
		"leisure":       "Leisure",
		"learning":      "Learning",
		"education":     "Education",
		"sleep":         "Sleep",
	}

	lower := strings.ToLower(category)

	if corrected, exists := commonCategories[lower]; exists {
		return corrected
	}

	bestMatch := category
	bestDistance := 999
	threshold := 2

	for known, canonical := range commonCategories {
		dist := levenshteinDistance(lower, known)
		if dist <= threshold && dist < bestDistance {
			bestDistance = dist
			bestMatch = canonical
		}
	}

	return bestMatch
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = minInt(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func minInt(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func AddEmojiToSummary(summary string, categories []string) string {
	if len(summary) > 0 && summary[0] > 127 {
		return summary
	}

	categoryLower := make([]string, len(categories))
	for i, cat := range categories {
		categoryLower[i] = strings.ToLower(strings.TrimSpace(cat))
	}

	for _, cat := range categoryLower {
		switch cat {
		case "medication", "meds":
			return "\U0001f48a " + summary
		case "health", "medical":
			return "\U0001f3e5 " + summary
		case "therapy", "mental health":
			return "\U0001f9e0 " + summary
		case "exercise", "workout", "fitness":
			return "\U0001f4aa " + summary
		case "food", "meal", "restaurant":
			return "\U0001f37d\ufe0f " + summary
		case "travel", "flight":
			return "\u2708\ufe0f " + summary
		case "accommodation", "hotel":
			return "\U0001f3e8 " + summary
		case "work", "meeting":
			return "\U0001f4bc " + summary
		case "focus", "deep work":
			return "\U0001f3af " + summary
		case "break", "rest":
			return "\u2615 " + summary
		case "transition":
			return "\U0001f504 " + summary
		case "family", "kids":
			return "\U0001f468\u200d\U0001f469\u200d\U0001f467 " + summary
		case "personal":
			return "\U0001f31f " + summary
		case "urgent", "important":
			return "\U0001f525 " + summary
		case "fun", "leisure":
			return "\U0001f389 " + summary
		case "learning", "education":
			return "\U0001f4da " + summary
		case "sleep":
			return "\U0001f634 " + summary
		}
	}

	summaryLower := strings.ToLower(summary)
	if strings.Contains(summaryLower, "med") || strings.Contains(summaryLower, "pill") {
		return "\U0001f48a " + summary
	}
	if strings.Contains(summaryLower, "breakfast") || strings.Contains(summaryLower, "lunch") || strings.Contains(summaryLower, "dinner") {
		return "\U0001f37d\ufe0f " + summary
	}
	if strings.Contains(summaryLower, "doctor") || strings.Contains(summaryLower, "dentist") || strings.Contains(summaryLower, "appointment") {
		return "\U0001f3e5 " + summary
	}
	if strings.Contains(summaryLower, "meeting") {
		return "\U0001f4bc " + summary
	}
	if strings.Contains(summaryLower, "focus") {
		return "\U0001f3af " + summary
	}

	return summary
}

func GetSmartDefaultDuration(summary string, startTime time.Time) time.Duration {
	summaryLower := strings.ToLower(summary)
	hour := startTime.Hour()

	if strings.Contains(summaryLower, "med") || strings.Contains(summaryLower, "pill") {
		return 5 * time.Minute
	}

	if strings.Contains(summaryLower, "breakfast") {
		return 30 * time.Minute
	}
	if strings.Contains(summaryLower, "lunch") {
		return 45 * time.Minute
	}
	if strings.Contains(summaryLower, "dinner") || strings.Contains(summaryLower, "supper") {
		return 1 * time.Hour
	}

	if strings.Contains(summaryLower, "standup") || strings.Contains(summaryLower, "stand-up") {
		return 15 * time.Minute
	}
	if strings.Contains(summaryLower, "break") || strings.Contains(summaryLower, "transition") {
		return 15 * time.Minute
	}

	if strings.Contains(summaryLower, "therapy") || strings.Contains(summaryLower, "therapist") {
		return 1 * time.Hour
	}
	if strings.Contains(summaryLower, "doctor") || strings.Contains(summaryLower, "dentist") {
		return 30 * time.Minute
	}

	if strings.Contains(summaryLower, "focus") || strings.Contains(summaryLower, "deep work") {
		return 2 * time.Hour
	}

	switch {
	case hour >= 6 && hour < 9:
		return 30 * time.Minute
	case hour >= 12 && hour < 14:
		return 1 * time.Hour
	case hour >= 18 && hour < 21:
		return 1*time.Hour + 30*time.Minute
	case hour >= 21 || hour < 6:
		return 30 * time.Minute
	default:
		return 1 * time.Hour
	}
}

func DetectEventConflicts(events []calendar.Event) []string {
	var conflicts []string

	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			ev1, ev2 := events[i], events[j]

			if ev1.AllDay || ev2.AllDay {
				continue
			}

			if ev1.EndTime.After(ev2.StartTime) && ev2.EndTime.After(ev1.StartTime) {
				conflict := fmt.Sprintf("%s (%s-%s) overlaps with %s (%s-%s)",
					ev1.Summary,
					ev1.StartTime.Format("15:04"),
					ev1.EndTime.Format("15:04"),
					ev2.Summary,
					ev2.StartTime.Format("15:04"),
					ev2.EndTime.Format("15:04"))
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

func GeneratePrepTimeEvents(events []calendar.Event) []*calendar.Event {
	var prepEvents []*calendar.Event

	for _, ev := range events {
		if ev.AllDay {
			continue
		}

		if transitionEvent := createTransitionEventIfNeeded(ev); transitionEvent != nil {
			prepEvents = append(prepEvents, transitionEvent)
			continue
		}

		if prepEvent := createPrepEventIfNeeded(ev); prepEvent != nil {
			prepEvents = append(prepEvents, prepEvent)
		}
	}

	return prepEvents
}

func createTransitionEventIfNeeded(ev calendar.Event) *calendar.Event {
	if !needsFocusTransition(ev.Summary) {
		return nil
	}

	return &calendar.Event{
		UID:        GenerateUID(),
		Summary:    "\U0001f504 Transition: " + StripEmoji(ev.Summary),
		StartTime:  ev.EndTime,
		EndTime:    ev.EndTime.Add(5 * time.Minute),
		StartTZ:    ev.StartTZ,
		EndTZ:      ev.EndTZ,
		AllDay:     false,
		Categories: []string{"Transition"},
		Status:     "CONFIRMED",
		Created:    time.Now().UTC(),
		LastMod:    time.Now().UTC(),
	}
}

func createPrepEventIfNeeded(ev calendar.Event) *calendar.Event {
	duration, description := determinePrepTime(ev.Summary)
	if duration == 0 {
		return nil
	}

	return &calendar.Event{
		UID:        GenerateUID(),
		Summary:    "\u23f0 " + description + ": " + StripEmoji(ev.Summary),
		StartTime:  ev.StartTime.Add(-duration),
		EndTime:    ev.StartTime,
		StartTZ:    ev.StartTZ,
		EndTZ:      ev.EndTZ,
		AllDay:     false,
		Categories: []string{"Preparation"},
		Status:     "CONFIRMED",
		Created:    time.Now().UTC(),
		LastMod:    time.Now().UTC(),
	}
}

func needsFocusTransition(summary string) bool {
	summaryLower := strings.ToLower(summary)
	focusKeywords := []string{"focus", "deep work", "coding", "writing"}

	for _, keyword := range focusKeywords {
		if strings.Contains(summaryLower, keyword) {
			return true
		}
	}
	return false
}

func determinePrepTime(summary string) (time.Duration, string) {
	summaryLower := strings.ToLower(summary)

	if containsAny(summaryLower, []string{"doctor", "m\u00e9dico", "dentist", "therapy", "hospital", "clinic"}) {
		return 20 * time.Minute, "Travel & arrival buffer"
	}

	if containsAny(summaryLower, []string{"meeting", "reunion", "appointment", "cita", "interview", "call"}) {
		return 15 * time.Minute, "Preparation"
	}

	return 0, ""
}

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func StripEmoji(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 0 {
		firstRune := []rune(s)[0]
		if firstRune > 127 {
			runes := []rune(s)
			if len(runes) > 1 {
				return strings.TrimSpace(string(runes[1:]))
			}
		}
	}
	return s
}

func GenerateUID() string {
	return uuid.New().String() + "@tempus"
}

func DetectOverwhelmDays(events []calendar.Event, maxPerDay int) []string {
	if maxPerDay == 0 {
		maxPerDay = 8
	}

	eventsByDay := make(map[string]int)
	for _, ev := range events {
		dateKey := ev.StartTime.Format("2006-01-02")
		eventsByDay[dateKey]++
	}

	var warnings []string
	for date, count := range eventsByDay {
		if count > maxPerDay {
			t, _ := time.Parse("2006-01-02", date)
			dayName := t.Format("Monday, Jan 2")
			warnings = append(warnings, fmt.Sprintf("%s: %d events (threshold: %d)", dayName, count, maxPerDay))
		}
	}

	sort.Strings(warnings)
	return warnings
}

func ExpandAlarmProfiles(alarmSpecs []string) []string {
	cfg, err := config.Load()
	if err != nil {
		return alarmSpecs
	}

	expanded := make([]string, 0, len(alarmSpecs))
	for _, spec := range alarmSpecs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		if strings.HasPrefix(spec, "profile:") {
			profileName := strings.TrimPrefix(spec, "profile:")
			profileName = strings.TrimSpace(profileName)

			profile := cfg.GetAlarmProfile(profileName)
			if profile != nil {
				expanded = append(expanded, profile...)
			} else {
				expanded = append(expanded, spec)
			}
		} else {
			expanded = append(expanded, spec)
		}
	}

	return expanded
}
