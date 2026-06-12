package nd

import (
	"strings"
)

type SpellCheckCache struct {
	corrections map[string]string
	cache       map[string]string
}

func NewSpellCheckCache(corrections map[string]string) *SpellCheckCache {
	if corrections == nil {
		corrections = make(map[string]string)
	}
	return &SpellCheckCache{
		corrections: corrections,
		cache:       make(map[string]string),
	}
}

func (c *SpellCheckCache) NormalizeAndCheck(text string) string {
	if text == "" {
		return text
	}

	words := strings.Fields(text)
	for i, word := range words {
		lower := strings.ToLower(word)
		if result, ok := c.cache[lower]; ok {
			words[i] = applyWordCase(word, result)
			continue
		}
		if corrected, exists := c.corrections[lower]; exists {
			c.cache[lower] = corrected
			words[i] = applyWordCase(word, corrected)
		} else {
			c.cache[lower] = word
		}
	}

	return strings.Join(words, " ")
}

type CategoryCache struct {
	cache map[string]string
}

func NewCategoryCache() *CategoryCache {
	return &CategoryCache{
		cache: make(map[string]string),
	}
}

func (c *CategoryCache) Validate(category string) string {
	lower := strings.ToLower(category)
	if result, ok := c.cache[lower]; ok {
		return result
	}
	result := ValidateCategoryWithSuggestion(category)
	c.cache[lower] = result
	return result
}
