package i18n

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed locales/*.json
var embeddedLocales embed.FS

var (
	embeddedOnce sync.Once
	embeddedData map[string]map[string]string
)

var localeExtensions = []string{".json", ".yaml", ".yml"}

const localeDirRelative = "locales"

// LocaleInfo describes the availability of a locale.
type LocaleInfo struct {
	Code      string
	Embedded  bool
	DiskPaths []string
}

const embeddedLocalesRoot = "locales"

// Translator handles internationalization
type Translator struct {
	language     string
	translations map[string]string
	fallback     map[string]string // English fallback
}

// Translation keys
const (
	KeyEventCreated       = "event_created"
	KeyInvalidDate        = "invalid_date"
	KeyInvalidTimezone    = "invalid_timezone"
	KeyConfigSaved        = "config_saved"
	KeyTemplateNotFound   = "template_not_found"
	KeyFlightTemplate     = "flight_template"
	KeyMeetingTemplate    = "meeting_template"
	KeyHolidayTemplate    = "holiday_template"
	KeyEventSummary       = "event_summary"
	KeyEventDescription   = "event_description"
	KeyEventLocation      = "event_location"
	KeyStartTime          = "start_time"
	KeyEndTime            = "end_time"
	KeyDuration           = "duration"
	KeyAttendees          = "attendees"
	KeyCategories         = "categories"
	KeyTimezone           = "timezone"
	KeyAllDay             = "all_day"
	KeyFlightFrom         = "flight_from"
	KeyFlightTo           = "flight_to"
	KeyFlightNumber       = "flight_number"
	KeyMeetingWith        = "meeting_with"
	KeyMeetingTopic       = "meeting_topic"
	KeyHolidayDestination = "holiday_destination"
)

// NewTranslator creates a new translator instance
func NewTranslator(language string) (*Translator, error) {
	t := &Translator{
		language: language,
	}

	// Load English as fallback
	fallback, err := loadTranslations("en")
	if err != nil {
		return nil, fmt.Errorf("failed to load fallback translations: %w", err)
	}
	t.fallback = fallback

	// Load requested language
	if language != "en" {
		translations, err := loadTranslations(language)
		if err != nil {
			// If we can't load the requested language, use English
			t.translations = fallback
		} else {
			t.translations = translations
		}
	} else {
		t.translations = fallback
	}

	return t, nil
}

// T translates a key to the current language
func (t *Translator) T(key string, args ...interface{}) string {
	// Try current language first
	if text, exists := t.translations[key]; exists {
		return fmt.Sprintf(text, args...)
	}

	// Fall back to English
	if text, exists := t.fallback[key]; exists {
		return fmt.Sprintf(text, args...)
	}

	// Return the key if no translation found
	return key
}

// GetLanguage returns the current language
func (t *Translator) GetLanguage() string {
	return t.language
}

// SupportedLanguages returns the list of available locales (embedded + disk overrides).
func SupportedLanguages() []string {
	infos := Locales()
	out := make([]string, 0, len(infos))
	for _, info := range infos {
		out = append(out, info.Code)
	}
	return out
}

// Locales returns a summary of supported locales, including source metadata.
func Locales() []LocaleInfo {
	ensureEmbeddedLocales()

	infoByCode := make(map[string]*LocaleInfo, len(embeddedData))
	for code := range embeddedData {
		infoByCode[code] = &LocaleInfo{Code: code, Embedded: true}
	}

	for _, dir := range localeSearchPaths() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if !isSupportedLocaleExt(ext) {
				continue
			}
			code := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			info, ok := infoByCode[code]
			if !ok {
				info = &LocaleInfo{Code: code}
				infoByCode[code] = info
			}
			info.DiskPaths = append(info.DiskPaths, filepath.Join(dir, entry.Name()))
		}
	}

	keys := make([]string, 0, len(infoByCode))
	for code := range infoByCode {
		keys = append(keys, code)
	}
	sort.Strings(keys)

	out := make([]LocaleInfo, 0, len(keys))
	for _, code := range keys {
		info := infoByCode[code]
		if len(info.DiskPaths) > 1 {
			sort.Strings(info.DiskPaths)
		}
		out = append(out, *info)
	}
	return out
}

// IsSupportedLanguage reports whether lang has an embedded or on-disk translation.
func IsSupportedLanguage(lang string) bool {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return false
	}

	ensureEmbeddedLocales()
	if _, ok := embeddedData[lang]; ok {
		return true
	}
	for _, ext := range localeExtensions {
		if _, err := os.Stat(filepath.Join("locales", lang+ext)); err == nil {
			return true
		}
	}
	return false
}

// loadTranslations loads translation data for a language.
func loadTranslations(language string) (map[string]string, error) {
	if data, err := loadFromDisk(language); err == nil {
		return data, nil
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	if data, ok := loadEmbeddedTranslation(language); ok {
		return data, nil
	}

	return nil, fmt.Errorf("translation data not found for language %s", language)
}

func loadFromDisk(language string) (map[string]string, error) {
	for _, base := range localeSearchPaths() {
		for _, ext := range localeExtensions {
			path := filepath.Join(base, language+ext)
			data, err := os.ReadFile(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return nil, fmt.Errorf("failed to read translation file %s: %w", path, err)
			}

			m, err := decodeLocaleBytes(data, ext)
			if err != nil {
				return nil, fmt.Errorf("failed to parse translation file %s: %w", path, err)
			}
			return m, nil
		}
	}
	return nil, fs.ErrNotExist
}

func loadEmbeddedTranslation(language string) (map[string]string, bool) {
	ensureEmbeddedLocales()
	m, ok := embeddedData[language]
	if !ok {
		return nil, false
	}
	return cloneStringMap(m), true
}

func ensureEmbeddedLocales() {
	embeddedOnce.Do(func() {
		embeddedData = make(map[string]map[string]string)
		entries, err := embeddedLocales.ReadDir(embeddedLocalesRoot)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if !isSupportedLocaleExt(ext) {
				continue
			}
			lang := strings.TrimSuffix(name, filepath.Ext(name))
			// Do not overwrite when multiple files refer to same language.
			if _, exists := embeddedData[lang]; exists {
				continue
			}
			data, err := embeddedLocales.ReadFile(path.Join(embeddedLocalesRoot, name))
			if err != nil {
				continue
			}
			m, err := decodeLocaleBytes(data, ext)
			if err != nil {
				continue
			}
			embeddedData[lang] = m
		}
	})
}

func isSupportedLocaleExt(ext string) bool {
	switch ext {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func decodeLocaleBytes(data []byte, ext string) (map[string]string, error) {
	var (
		out map[string]string
		err error
	)
	switch strings.ToLower(ext) {
	case ".json":
		err = json.Unmarshal(data, &out)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &out)
	default:
		return nil, fmt.Errorf("unsupported locale format: %s", ext)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// FormatDateTime formats a datetime according to locale preferences
func (t *Translator) FormatDateTime(dt time.Time, dateOnly bool) string {
	switch t.language {
	case "es", "ga", "pt":
		if dateOnly {
			return dt.Format("02/01/2006")
		}
		return dt.Format("02/01/2006 15:04")
	default: // en and others
		if dateOnly {
			return dt.Format("01/02/2006")
		}
		return dt.Format("01/02/2006 15:04")
	}
}

func localeSearchPaths() []string {
	paths := make([]string, 0, 2)
	if cdir, err := os.UserConfigDir(); err == nil && strings.TrimSpace(cdir) != "" {
		paths = append(paths, filepath.Join(cdir, "tempus", localeDirRelative))
	}
	paths = append(paths, localeDirRelative)
	return dedupe(paths)
}

func dedupe(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

// GetDateFormat returns the date format string for the current locale
func (t *Translator) GetDateFormat() string {
	switch t.language {
	case "es", "ga", "pt":
		return "DD/MM/YYYY"
	default:
		return "MM/DD/YYYY"
	}
}

// GetTimeFormat returns the time format string for the current locale
func (t *Translator) GetTimeFormat() string {
	return "HH:MM" // 24-hour format for all locales
}
