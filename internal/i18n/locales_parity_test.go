package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const rootLocalesDir = "../../locales"

func loadLocaleTreeFromDisk(t *testing.T, dir string) map[string]map[string]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read locale dir %s: %v", dir, err)
	}
	tree := make(map[string]map[string]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		lang := strings.TrimSuffix(entry.Name(), ".json")
		data, err := os.ReadFile(filepath.Join(dir, entry.Name())) // #nosec G304 -- test fixture path
		if err != nil {
			t.Fatalf("failed to read %s: %v", entry.Name(), err)
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("invalid JSON in %s/%s: %v", dir, entry.Name(), err)
		}
		tree[lang] = m
	}
	return tree
}

func loadEmbeddedLocaleTree(t *testing.T) map[string]map[string]string {
	t.Helper()
	entries, err := embeddedLocales.ReadDir(embeddedLocalesRoot)
	if err != nil {
		t.Fatalf("failed to read embedded locales: %v", err)
	}
	tree := make(map[string]map[string]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		lang := strings.TrimSuffix(entry.Name(), ".json")
		data, err := embeddedLocales.ReadFile(embeddedLocalesRoot + "/" + entry.Name())
		if err != nil {
			t.Fatalf("failed to read embedded %s: %v", entry.Name(), err)
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("invalid JSON in embedded %s: %v", entry.Name(), err)
		}
		tree[lang] = m
	}
	return tree
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func diffKeySets(a, b map[string]string) (onlyA, onlyB []string) {
	for k := range a {
		if _, ok := b[k]; !ok {
			onlyA = append(onlyA, k)
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			onlyB = append(onlyB, k)
		}
	}
	sort.Strings(onlyA)
	sort.Strings(onlyB)
	return onlyA, onlyB
}

// TestLocaleKeyParity is the anti-drift guard for the two locale trees.
// It fails when any language's key set differs from en within a tree, or
// when the embedded catalog differs from the root locales/ catalog.
func TestLocaleKeyParity(t *testing.T) {
	root := loadLocaleTreeFromDisk(t, rootLocalesDir)
	embedded := loadEmbeddedLocaleTree(t)

	if len(root) == 0 {
		t.Fatal("no JSON locales found in root locales/ directory")
	}
	if len(embedded) == 0 {
		t.Fatal("no JSON locales found in embedded catalog")
	}

	rootLangs := sortedKeys(root)
	embeddedLangs := sortedKeys(embedded)
	if strings.Join(rootLangs, ",") != strings.Join(embeddedLangs, ",") {
		t.Fatalf("language sets differ: root=%v embedded=%v", rootLangs, embeddedLangs)
	}

	for _, tree := range []struct {
		name    string
		locales map[string]map[string]string
	}{
		{"root", root},
		{"embedded", embedded},
	} {
		en, ok := tree.locales["en"]
		if !ok {
			t.Fatalf("%s tree is missing en.json", tree.name)
		}
		for _, lang := range sortedKeys(tree.locales) {
			missing, extra := diffKeySets(en, tree.locales[lang])
			if len(missing) > 0 {
				t.Errorf("%s/%s.json is missing keys present in en: %v", tree.name, lang, missing)
			}
			if len(extra) > 0 {
				t.Errorf("%s/%s.json has keys absent from en: %v", tree.name, lang, extra)
			}
		}
	}

	for _, lang := range rootLangs {
		missing, extra := diffKeySets(root[lang], embedded[lang])
		if len(missing) > 0 {
			t.Errorf("embedded %s.json is missing keys present in root: %v", lang, missing)
		}
		if len(extra) > 0 {
			t.Errorf("embedded %s.json has keys absent from root: %v", lang, extra)
		}
		for _, key := range sortedKeys(root[lang]) {
			ev, ok := embedded[lang][key]
			if !ok {
				continue // already reported above
			}
			if rv := root[lang][key]; rv != ev {
				t.Errorf("value drift for %s key %q: root=%q embedded=%q", lang, key, rv, ev)
			}
		}
	}
}
