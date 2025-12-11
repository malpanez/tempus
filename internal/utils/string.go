package utils

import (
	"strings"
)

// Slugify converts a string to a URL-friendly slug.
// It converts to lowercase, replaces spaces and special characters with hyphens,
// and removes any characters that are not alphanumeric or hyphens.
// Returns "event" if the input is empty or only contains special characters.
//
// Examples:
//   - Slugify("Hello World") -> "hello-world"
//   - Slugify("Meeting @ 3pm") -> "meeting-3pm"
//   - Slugify("Múltiple   Espacios") -> "m-ltiple-espacios"
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}

	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevHyphen = false
		case r == '-' || r == '_' || r == ' ' || r == '.' || r == '/' || r == '\\':
			if !prevHyphen && b.Len() > 0 {
				b.WriteRune('-')
				prevHyphen = true
			}
		default:
			if !prevHyphen && b.Len() > 0 {
				b.WriteRune('-')
				prevHyphen = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "event"
	}
	return out
}
