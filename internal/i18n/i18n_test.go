package i18n

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testErrCreateLocalesDir = "failed to create locales directory: %v"
	testErrNewTranslator    = "NewTranslator() error = %v"
	testDate                = "15/03/2024"
	testDateTime            = "15/03/2024 14:30"
	testDateFormatDDMMYYYY  = "DD/MM/YYYY"
)

func TestSupportedLanguagesIncludesEmbedded(t *testing.T) {
	langs := SupportedLanguages()
	required := []string{"en", "es", "ga", "pt"}
	for _, lang := range required {
		if !contains(langs, lang) {
			t.Fatalf("expected supported languages to contain %q; got %v", lang, langs)
		}
	}
}

func TestIsSupportedLanguageDetectsDiskOverride(t *testing.T) {
	dir := "locales"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf(testErrCreateLocalesDir, err)
	}

	path := filepath.Join(dir, "test-precommit.yaml")
	if err := os.WriteFile(path, []byte("hello: world\n"), 0o644); err != nil {
		t.Fatalf("failed to create temporary locale: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(path)
	})

	if !IsSupportedLanguage("test-precommit") {
		t.Fatalf("expected disk locale test-precommit to be detected")
	}

	found := false
	for _, loc := range Locales() {
		if loc.Code == "test-precommit" {
			if len(loc.DiskPaths) == 0 {
				t.Fatalf("expected disk path for test-precommit locale")
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("Locales did not report test-precommit")
	}
}

func TestNewTranslatorLoadsSpanish(t *testing.T) {
	tr, err := NewTranslator("es")
	if err != nil {
		t.Fatalf("failed to load spanish translator: %v", err)
	}

	const sample = "Formato de fecha inválido: foo"
	got := tr.T(KeyInvalidDate, "foo")
	if got != sample {
		t.Fatalf("unexpected translation: got %q want %q", got, sample)
	}
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

// TestNewTranslatorWithDifferentLanguages tests creating translators for all supported languages
func TestNewTranslatorWithDifferentLanguages(t *testing.T) {
	tests := []struct {
		name     string
		language string
		wantErr  bool
	}{
		{
			name:     "English",
			language: "en",
			wantErr:  false,
		},
		{
			name:     "Spanish",
			language: "es",
			wantErr:  false,
		},
		{
			name:     "Irish",
			language: "ga",
			wantErr:  false,
		},
		{
			name:     "Portuguese",
			language: "pt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTranslator(tt.language)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTranslator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tr.GetLanguage() != tt.language {
				t.Errorf("GetLanguage() = %v, want %v", tr.GetLanguage(), tt.language)
			}
		})
	}
}

// TestNewTranslatorWithInvalidLanguage tests that invalid language falls back to English
func TestNewTranslatorWithInvalidLanguage(t *testing.T) {
	tr, err := NewTranslator("invalid-lang")
	if err != nil {
		t.Fatalf("NewTranslator() unexpected error = %v", err)
	}
	if tr.GetLanguage() != "invalid-lang" {
		t.Errorf("GetLanguage() = %v, want invalid-lang", tr.GetLanguage())
	}

	// Should still be able to translate using English fallback
	result := tr.T(KeyEventCreated, "test-event")
	if result == KeyEventCreated {
		t.Errorf("T() returned key instead of translation, expected English fallback")
	}
}

// TestTranslatorT tests the T() method with various scenarios
func TestTranslatorT(t *testing.T) {
	tr, err := NewTranslator("en")
	if err != nil {
		t.Fatalf(testErrNewTranslator, err)
	}

	tests := []struct {
		name     string
		key      string
		args     []interface{}
		wantText string
	}{
		{
			name:     "ExistingKeyNoArgs",
			key:      KeyConfigSaved,
			args:     nil,
			wantText: "Configuration saved",
		},
		{
			name:     "ExistingKeyWithArgs",
			key:      KeyEventCreated,
			args:     []interface{}{"my-event.ics"},
			wantText: "Event created successfully: my-event.ics",
		},
		{
			name:     "ExistingKeyMultipleArgs",
			key:      KeyInvalidDate,
			args:     []interface{}{"2023-99-99"},
			wantText: "Invalid date format: 2023-99-99",
		},
		{
			name:     "MissingKey",
			key:      "non_existent_key",
			args:     nil,
			wantText: "non_existent_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tr.T(tt.key, tt.args...)
			if result != tt.wantText {
				t.Errorf("T() = %q, want %q", result, tt.wantText)
			}
		})
	}
}

// TestTranslatorTFallback tests that T() falls back to English when translation is missing
func TestTranslatorTFallback(t *testing.T) {
	// Create a custom locale file with missing keys
	dir := "locales"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf(testErrCreateLocalesDir, err)
	}

	path := filepath.Join(dir, "partial.json")
	// Only include one key
	content := `{"config_saved": "Config guardado"}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create partial locale: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(path)
	})

	tr, err := NewTranslator("partial")
	if err != nil {
		t.Fatalf(testErrNewTranslator, err)
	}

	// Test that existing key uses the partial translation
	result := tr.T(KeyConfigSaved)
	if result != "Config guardado" {
		t.Errorf("T(KeyConfigSaved) = %q, want 'Config guardado'", result)
	}

	// Test that missing key falls back to English
	result = tr.T(KeyEventCreated, "test")
	if result == KeyEventCreated || result == "" {
		t.Errorf("T(KeyEventCreated) should fall back to English, got %q", result)
	}
}

// TestGetLanguage tests the GetLanguage method
func TestGetLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
	}{
		{"English", "en"},
		{"Spanish", "es"},
		{"Irish", "ga"},
		{"Portuguese", "pt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTranslator(tt.language)
			if err != nil {
				t.Fatalf(testErrNewTranslator, err)
			}
			if got := tr.GetLanguage(); got != tt.language {
				t.Errorf("GetLanguage() = %v, want %v", got, tt.language)
			}
		})
	}
}

// TestFormatDateTime tests date/time formatting for different locales
func TestFormatDateTime(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		language string
		dateOnly bool
		want     string
	}{
		{
			name:     "EnglishDateOnly",
			language: "en",
			dateOnly: true,
			want:     "03/15/2024",
		},
		{
			name:     "EnglishDateTime",
			language: "en",
			dateOnly: false,
			want:     "03/15/2024 14:30",
		},
		{
			name:     "SpanishDateOnly",
			language: "es",
			dateOnly: true,
			want:     testDate,
		},
		{
			name:     "SpanishDateTime",
			language: "es",
			dateOnly: false,
			want:     testDateTime,
		},
		{
			name:     "IrishDateOnly",
			language: "ga",
			dateOnly: true,
			want:     testDate,
		},
		{
			name:     "IrishDateTime",
			language: "ga",
			dateOnly: false,
			want:     testDateTime,
		},
		{
			name:     "PortugueseDateOnly",
			language: "pt",
			dateOnly: true,
			want:     testDate,
		},
		{
			name:     "PortugueseDateTime",
			language: "pt",
			dateOnly: false,
			want:     testDateTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTranslator(tt.language)
			if err != nil {
				t.Fatalf(testErrNewTranslator, err)
			}
			got := tr.FormatDateTime(testTime, tt.dateOnly)
			if got != tt.want {
				t.Errorf("FormatDateTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetDateFormat tests the GetDateFormat method
func TestGetDateFormat(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     string
	}{
		{
			name:     "English",
			language: "en",
			want:     "MM/DD/YYYY",
		},
		{
			name:     "Spanish",
			language: "es",
			want:     testDateFormatDDMMYYYY,
		},
		{
			name:     "Irish",
			language: "ga",
			want:     testDateFormatDDMMYYYY,
		},
		{
			name:     "Portuguese",
			language: "pt",
			want:     testDateFormatDDMMYYYY,
		},
		{
			name:     "Unknown",
			language: "fr",
			want:     "MM/DD/YYYY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTranslator(tt.language)
			if err != nil {
				t.Fatalf(testErrNewTranslator, err)
			}
			got := tr.GetDateFormat()
			if got != tt.want {
				t.Errorf("GetDateFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetTimeFormat tests the GetTimeFormat method
func TestGetTimeFormat(t *testing.T) {
	tr, err := NewTranslator("en")
	if err != nil {
		t.Fatalf(testErrNewTranslator, err)
	}

	want := "HH:MM"
	got := tr.GetTimeFormat()
	if got != want {
		t.Errorf("GetTimeFormat() = %v, want %v", got, want)
	}

	// Test for other languages - should all return the same
	tr2, err := NewTranslator("es")
	if err != nil {
		t.Fatalf(testErrNewTranslator, err)
	}
	got2 := tr2.GetTimeFormat()
	if got2 != want {
		t.Errorf("GetTimeFormat() for Spanish = %v, want %v", got2, want)
	}
}

// TestLoadEmbeddedTranslation tests loading embedded translations
func TestLoadEmbeddedTranslation(t *testing.T) {
	tests := []struct {
		name       string
		language   string
		wantLoaded bool
	}{
		{
			name:       "EnglishEmbedded",
			language:   "en",
			wantLoaded: true,
		},
		{
			name:       "SpanishEmbedded",
			language:   "es",
			wantLoaded: true,
		},
		{
			name:       "NonExistent",
			language:   "xx",
			wantLoaded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, ok := loadEmbeddedTranslation(tt.language)
			if ok != tt.wantLoaded {
				t.Errorf("loadEmbeddedTranslation() ok = %v, want %v", ok, tt.wantLoaded)
			}
			if tt.wantLoaded && len(data) == 0 {
				t.Errorf("loadEmbeddedTranslation() returned empty data for %s", tt.language)
			}
		})
	}
}

// TestCloneStringMap tests the cloneStringMap function
func TestCloneStringMap(t *testing.T) {
	original := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	cloned := cloneStringMap(original)

	// Verify contents match
	if len(cloned) != len(original) {
		t.Errorf("cloneStringMap() length = %d, want %d", len(cloned), len(original))
	}

	for k, v := range original {
		if cloned[k] != v {
			t.Errorf("cloneStringMap() key %s = %v, want %v", k, cloned[k], v)
		}
	}

	// Verify it's a copy, not a reference
	cloned["key1"] = "modified"
	if original["key1"] == "modified" {
		t.Errorf("cloneStringMap() did not create a copy, original was modified")
	}

	// Test with empty map
	emptyClone := cloneStringMap(map[string]string{})
	if len(emptyClone) != 0 {
		t.Errorf("cloneStringMap() with empty map should return empty map, got length %d", len(emptyClone))
	}
}

// TestDedupe tests the dedupe function
func TestDedupe(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "NoDuplicates",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "WithDuplicates",
			input: []string{"a", "b", "a", "c", "b"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "WithEmptyStrings",
			input: []string{"a", "", "b", "  ", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "AllEmpty",
			input: []string{"", "  ", "   "},
			want:  []string{},
		},
		{
			name:  "EmptyInput",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "WithWhitespace",
			input: []string{" a ", "a", "b ", " b"},
			want:  []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedupe(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("dedupe() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("dedupe() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

// TestIsSupportedLocaleExt tests the isSupportedLocaleExt function
func TestIsSupportedLocaleExt(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		want bool
	}{
		{
			name: "JSON",
			ext:  ".json",
			want: true,
		},
		{
			name: "YAML",
			ext:  ".yaml",
			want: true,
		},
		{
			name: "YML",
			ext:  ".yml",
			want: true,
		},
		{
			name: "JSONUpperCase",
			ext:  ".JSON",
			want: false,
		},
		{
			name: "TXT",
			ext:  ".txt",
			want: false,
		},
		{
			name: "XML",
			ext:  ".xml",
			want: false,
		},
		{
			name: "Empty",
			ext:  "",
			want: false,
		},
		{
			name: "NoExtension",
			ext:  "json",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSupportedLocaleExt(tt.ext)
			if got != tt.want {
				t.Errorf("isSupportedLocaleExt(%q) = %v, want %v", tt.ext, got, tt.want)
			}
		})
	}
}

// TestDecodeLocaleBytes tests the decodeLocaleBytes function
func TestDecodeLocaleBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		ext     string
		wantErr bool
		wantLen int
	}{
		{
			name:    "ValidJSON",
			data:    []byte(`{"key1": "value1", "key2": "value2"}`),
			ext:     ".json",
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "ValidYAML",
			data:    []byte("key1: value1\nkey2: value2"),
			ext:     ".yaml",
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "ValidYML",
			data:    []byte("key1: value1\nkey2: value2"),
			ext:     ".yml",
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "InvalidJSON",
			data:    []byte(`{invalid json}`),
			ext:     ".json",
			wantErr: true,
		},
		{
			name:    "InvalidYAML",
			data:    []byte(":\ninvalid: yaml: content:"),
			ext:     ".yaml",
			wantErr: true,
		},
		{
			name:    "UnsupportedExtension",
			data:    []byte("some data"),
			ext:     ".txt",
			wantErr: true,
		},
		{
			name:    "EmptyJSON",
			data:    []byte("{}"),
			ext:     ".json",
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeLocaleBytes(tt.data, tt.ext)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeLocaleBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("decodeLocaleBytes() returned map length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestLoadFromDiskWithDifferentFormats tests loading locale files in different formats
func TestLoadFromDiskWithDifferentFormats(t *testing.T) {
	dir := "locales"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf(testErrCreateLocalesDir, err)
	}

	tests := []struct {
		name     string
		filename string
		content  string
		wantErr  bool
	}{
		{
			name:     "JSONFormat",
			filename: "test-json.json",
			content:  `{"hello": "world", "test": "value"}`,
			wantErr:  false,
		},
		{
			name:     "YAMLFormat",
			filename: "test-yaml.yaml",
			content:  "hello: world\ntest: value",
			wantErr:  false,
		},
		{
			name:     "YMLFormat",
			filename: "test-yml.yml",
			content:  "hello: world\ntest: value",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.filename)
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}
			t.Cleanup(func() {
				_ = os.Remove(path)
			})

			lang := tt.filename[:len(tt.filename)-len(filepath.Ext(tt.filename))]
			data, err := loadFromDisk(lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadFromDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) == 0 {
				t.Errorf("loadFromDisk() returned empty data")
			}
		})
	}
}

// TestLoadFromDiskWithInvalidContent tests error handling for invalid file contents
func TestLoadFromDiskWithInvalidContent(t *testing.T) {
	dir := "locales"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf(testErrCreateLocalesDir, err)
	}

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "InvalidJSON",
			filename: "invalid-json.json",
			content:  `{invalid json content`,
		},
		{
			name:     "InvalidYAML",
			filename: "invalid-yaml.yaml",
			content:  ":\ninvalid:\nyaml:\ncontent:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.filename)
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}
			t.Cleanup(func() {
				_ = os.Remove(path)
			})

			lang := tt.filename[:len(tt.filename)-len(filepath.Ext(tt.filename))]
			_, err := loadFromDisk(lang)
			if err == nil {
				t.Errorf("loadFromDisk() expected error for invalid content, got nil")
			}
		})
	}
}

// TestLoadTranslationsPrefersDisk tests that disk translations are preferred over embedded
func TestLoadTranslationsPrefersDisk(t *testing.T) {
	dir := "locales"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf(testErrCreateLocalesDir, err)
	}

	// Use a custom language name to avoid interfering with en.json
	path := filepath.Join(dir, "custom-disk-test.json")
	customContent := `{"config_saved": "Custom disk translation", "event_created": "Test created: %s"}`
	if err := os.WriteFile(path, []byte(customContent), 0o644); err != nil {
		t.Fatalf("failed to create custom locale: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(path)
	})

	data, err := loadTranslations("custom-disk-test")
	if err != nil {
		t.Fatalf("loadTranslations() error = %v", err)
	}

	// Should load the custom disk version
	if data["config_saved"] != "Custom disk translation" {
		t.Errorf("loadTranslations() did not load from disk, got %q", data["config_saved"])
	}
}

// TestIsSupportedLanguageWithEmptyString tests IsSupportedLanguage with empty/whitespace strings
func TestIsSupportedLanguageWithEmptyString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "EmptyString",
			input: "",
			want:  false,
		},
		{
			name:  "Whitespace",
			input: "   ",
			want:  false,
		},
		{
			name:  "ValidLanguage",
			input: "en",
			want:  true,
		},
		{
			name:  "ValidLanguageWithSpaces",
			input: "  en  ",
			want:  true,
		},
		{
			name:  "InvalidLanguage",
			input: "zz",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSupportedLanguage(tt.input)
			if got != tt.want {
				t.Errorf("IsSupportedLanguage(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestLocalesWithMultipleDiskPaths tests Locales() when multiple disk paths exist for same locale
func TestLocalesWithMultipleDiskPaths(t *testing.T) {
	dir := "locales"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf(testErrCreateLocalesDir, err)
	}

	// Create the same locale in multiple formats
	jsonPath := filepath.Join(dir, "multi.json")
	yamlPath := filepath.Join(dir, "multi.yaml")

	if err := os.WriteFile(jsonPath, []byte(`{"key": "json"}`), 0o644); err != nil {
		t.Fatalf("failed to create json file: %v", err)
	}
	if err := os.WriteFile(yamlPath, []byte("key: yaml"), 0o644); err != nil {
		t.Fatalf("failed to create yaml file: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(jsonPath)
		_ = os.Remove(yamlPath)
	})

	locales := Locales()
	var multiLocale *LocaleInfo
	for _, loc := range locales {
		if loc.Code == "multi" {
			multiLocale = &loc
			break
		}
	}

	if multiLocale == nil {
		t.Fatalf("Expected to find 'multi' locale")
	}

	if len(multiLocale.DiskPaths) < 2 {
		t.Errorf("Expected at least 2 disk paths for 'multi' locale, got %d", len(multiLocale.DiskPaths))
	}

	// Verify paths are sorted
	if len(multiLocale.DiskPaths) >= 2 {
		for i := 1; i < len(multiLocale.DiskPaths); i++ {
			if multiLocale.DiskPaths[i-1] > multiLocale.DiskPaths[i] {
				t.Errorf("DiskPaths not sorted: %v", multiLocale.DiskPaths)
				break
			}
		}
	}
}

// TestLocalesReturnsAllSources tests that Locales() returns both embedded and disk locales
func TestLocalesReturnsAllSources(t *testing.T) {
	locales := Locales()

	if len(locales) == 0 {
		t.Fatal("Locales() returned empty list")
	}

	// Check that at least English is present and marked as embedded
	var foundEn bool
	for _, loc := range locales {
		if loc.Code == "en" {
			foundEn = true
			if !loc.Embedded {
				t.Errorf("Expected English locale to be marked as embedded")
			}
			break
		}
	}

	if !foundEn {
		t.Errorf("Expected to find English in Locales()")
	}

	// Verify locales are sorted
	for i := 1; i < len(locales); i++ {
		if locales[i-1].Code > locales[i].Code {
			t.Errorf("Locales() not sorted by code")
			break
		}
	}
}

// TestTranslatorWithEmptyArgs tests T() with empty format arguments
func TestTranslatorWithEmptyArgs(t *testing.T) {
	tr, err := NewTranslator("en")
	if err != nil {
		t.Fatalf(testErrNewTranslator, err)
	}

	// Test key with format string but no args
	result := tr.T(KeyEventCreated)
	// Should still work but format placeholder will be visible
	if result == KeyEventCreated {
		t.Errorf("T() returned key instead of translation")
	}
}

// TestLoadFromDiskNonExistent tests loadFromDisk with non-existent language
func TestLoadFromDiskNonExistent(t *testing.T) {
	_, err := loadFromDisk("nonexistent-language-xyz")
	if err == nil {
		t.Errorf("loadFromDisk() expected error for non-existent language, got nil")
	}
}

// TestFormatDateTimeEdgeCases tests FormatDateTime with edge case dates
func TestFormatDateTimeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		language string
		dateOnly bool
	}{
		{
			name:     "ZeroTime",
			time:     time.Time{},
			language: "en",
			dateOnly: true,
		},
		{
			name:     "LeapYear",
			time:     time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC),
			language: "es",
			dateOnly: false,
		},
		{
			name:     "NewYear",
			time:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			language: "pt",
			dateOnly: true,
		},
		{
			name:     "EndOfYear",
			time:     time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			language: "ga",
			dateOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTranslator(tt.language)
			if err != nil {
				t.Fatalf(testErrNewTranslator, err)
			}
			result := tr.FormatDateTime(tt.time, tt.dateOnly)
			if result == "" {
				t.Errorf("FormatDateTime() returned empty string")
			}
		})
	}
}
