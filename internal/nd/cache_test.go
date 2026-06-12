package nd

import (
	"testing"
)

func TestSpellCheckCacheNormalizeAndCheck(t *testing.T) {
	corrections := map[string]string{
		"meetng": "meeting",
		"doctr":  "doctor",
		"lunhc":  "lunch",
	}

	t.Run("corrects known typo", func(t *testing.T) {
		cache := NewSpellCheckCache(corrections)
		got := cache.NormalizeAndCheck("meetng")
		if got != "meeting" {
			t.Errorf("NormalizeAndCheck(\"meetng\") = %q, want \"meeting\"", got)
		}
	})

	t.Run("cache hit returns same result", func(t *testing.T) {
		cache := NewSpellCheckCache(corrections)
		first := cache.NormalizeAndCheck("meetng")
		second := cache.NormalizeAndCheck("meetng")
		if first != second {
			t.Errorf("cache hit mismatch: first=%q, second=%q", first, second)
		}
		if second != "meeting" {
			t.Errorf("second call = %q, want \"meeting\"", second)
		}
	})

	t.Run("no correction returns input unchanged", func(t *testing.T) {
		cache := NewSpellCheckCache(corrections)
		got := cache.NormalizeAndCheck("hello")
		if got != "hello" {
			t.Errorf("NormalizeAndCheck(\"hello\") = %q, want \"hello\"", got)
		}
	})

	t.Run("empty corrections returns input unchanged", func(t *testing.T) {
		cache := NewSpellCheckCache(nil)
		got := cache.NormalizeAndCheck("meetng")
		if got != "meetng" {
			t.Errorf("NormalizeAndCheck with nil corrections = %q, want \"meetng\"", got)
		}
	})

	t.Run("preserves capitalization", func(t *testing.T) {
		cache := NewSpellCheckCache(corrections)
		got := cache.NormalizeAndCheck("Meetng")
		if got != "Meeting" {
			t.Errorf("NormalizeAndCheck(\"Meetng\") = %q, want \"Meeting\"", got)
		}
	})

	t.Run("multi-word text", func(t *testing.T) {
		cache := NewSpellCheckCache(corrections)
		got := cache.NormalizeAndCheck("meetng with doctr")
		if got != "meeting with doctor" {
			t.Errorf("NormalizeAndCheck multi-word = %q, want \"meeting with doctor\"", got)
		}
	})

	t.Run("empty string returns empty", func(t *testing.T) {
		cache := NewSpellCheckCache(corrections)
		got := cache.NormalizeAndCheck("")
		if got != "" {
			t.Errorf("NormalizeAndCheck(\"\") = %q, want \"\"", got)
		}
	})
}

func TestCategoryCacheValidate(t *testing.T) {
	t.Run("exact match lowercase", func(t *testing.T) {
		cache := NewCategoryCache()
		got := cache.Validate("work")
		if got != "Work" {
			t.Errorf("Validate(\"work\") = %q, want \"Work\"", got)
		}
	})

	t.Run("Levenshtein match", func(t *testing.T) {
		cache := NewCategoryCache()
		got := cache.Validate("wrk")
		if got != "Work" {
			t.Errorf("Validate(\"wrk\") = %q, want \"Work\"", got)
		}
	})

	t.Run("cache hit returns same result", func(t *testing.T) {
		cache := NewCategoryCache()
		first := cache.Validate("wrk")
		second := cache.Validate("wrk")
		if first != second {
			t.Errorf("cache hit mismatch: first=%q, second=%q", first, second)
		}
		if second != "Work" {
			t.Errorf("second call = %q, want \"Work\"", second)
		}
	})

	t.Run("unknown category returns input", func(t *testing.T) {
		cache := NewCategoryCache()
		got := cache.Validate("xyzabc123")
		if got != "xyzabc123" {
			t.Errorf("Validate(\"xyzabc123\") = %q, want \"xyzabc123\"", got)
		}
	})

	t.Run("empty category", func(t *testing.T) {
		cache := NewCategoryCache()
		got := cache.Validate("")
		if got != "" {
			t.Errorf("Validate(\"\") = %q, want \"\"", got)
		}
	})
}
