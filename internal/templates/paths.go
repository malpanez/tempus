package templates

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultTemplateDirs returns the standard directories Tempus scans for templates.
func DefaultTemplateDirs() []string {
	dirs := make([]string, 0, 2)
	if cdir, err := os.UserConfigDir(); err == nil && strings.TrimSpace(cdir) != "" {
		dirs = append(dirs, filepath.Join(cdir, "tempus", "templates"))
	}
	dirs = append(dirs, filepath.Join("internal", "templates", "json"))
	return dedupeStrings(dirs)
}

// ResolveTemplateDirs combines a custom directory (if provided) with defaults.
func ResolveTemplateDirs(custom string) []string {
	custom = strings.TrimSpace(custom)
	if custom != "" {
		return dedupeStrings([]string{filepath.Clean(custom)})
	}
	return DefaultTemplateDirs()
}

func dedupeStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
