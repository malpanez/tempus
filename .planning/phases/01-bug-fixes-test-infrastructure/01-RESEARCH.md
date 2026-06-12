# Phase 1: Bug Fixes & Test Infrastructure - Research

**Researched:** 2026-03-29
**Domain:** Go CLI bug fixes, i18n, test infrastructure
**Confidence:** HIGH

## Summary

Phase 1 addresses 5 data-integrity bugs and 2 infrastructure improvements in a Go CLI monolith (`main.go`). All changes are localized -- no new packages, no architectural shifts. The codebase has 1404 passing tests across 11 packages with strong coverage on most target functions. The key risk is that existing tests assert the current (buggy) behavior -- specifically `TestStripEmoji` expects `"¡Hola"` to become `"Hola"`, which must be updated when BUG-01 is fixed.

The bugs are independent of each other, enabling parallel development with minimal merge conflict risk. REF-06 (var stdout) should be implemented first since it enables better test assertions for subsequent fixes. The i18n work (BUG-02) is the most tedious -- 15 hardcoded Spanish strings in `promptAlarmField()` need keys across 4 locale files.

**Primary recommendation:** Implement REF-06 first (test infrastructure), then BUG-01 (simplest fix, highest data corruption risk), then BUG-03/BUG-04 (error handling pair), then BUG-02 (i18n), then BUG-05/REF-02 (normalization).

<user_constraints>

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Alarm profile not found -> fatal error, abort execution. Error message must list available profiles: `"Profile 'X' not found. Available: adhd-default, medication, work-meeting"`. Same behavior in both `create` and `batch` -- no permissive mode.
- **D-02:** City not recognized by `cityToIANA()` -> fatal error with actionable suggestion: `"Unknown city 'X'. Use 'tempus timezone list --search X' to find the IANA identifier."` Consistent with D-01 (both fatal, both with guidance).
- **D-03:** REF-02 in Phase 1 means only: call `normalizeDateTimeInput()` before `time.Parse()` in `parseTimedEventTimes()` (and `parseAllDayTimes()` for slash-separated dates). No new package, no signature changes. Full unification deferred to Phase 3.
- **D-04:** Use package-level variable: `var stdout io.Writer = os.Stdout` at the top of main.go. `printOK`/`printErr`/`printDryRunSummary` write to `stdout` instead of `fmt.Printf`. Tests override `stdout` to capture output. Zero changes to callers.

### Claude's Discretion
- Unicode emoji detection approach for BUG-01: use `unicode.Is(unicode.So, r)` (Symbol, Other category). No external dependency needed.
- i18n keys for BUG-02: add ~15 keys to all 4 locale files (en, es, pt, ga). Pass `*i18n.Translator` as parameter to `promptAlarmField()`.
- `expandAlarmProfiles()` error return: change signature to `([]string, error)`. Single call site in `addBatchAlarms()`.

### Deferred Ideas (OUT OF SCOPE)
- Full `internal/parsing.Parse(ParseOptions)` unification -- Phase 3
- `internal/cli` package creation -- Phase 3
- Migrating `promptAlarmField()` to use `charmbracelet/huh` -- Phase 3

</user_constraints>

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| BUG-01 | `stripEmoji()` uses `unicode.Is(unicode.So, r)` instead of `rune > 127` | Fix confirmed in `stripEmoji()` L1692 AND `addEmojiToSummary()` L1429 -- both use same `> 127` pattern |
| BUG-02 | `promptAlarmField()` uses i18n system with ~15 new keys in 4 locales | 15 hardcoded Spanish strings identified at L3409-3462; caller at L2748 already has translator context |
| BUG-03 | `expandAlarmProfiles()` returns `([]string, error)` with clear error message | Single call site at `addBatchAlarms()` L1229; `cfg.ListAlarmProfiles()` exists for error message |
| BUG-04 | `cityToIANA()` returns error with suggestion | Single call site at L3773; existing test at L1896 needs negative case update |
| BUG-05 | `normalizeDateTimeInput()` applied in `parseTimedEventTimes()` and `parseAllDayTimes()` | Currently only called in batch path L1057; `parseTimedEventTimes()` L404 is the insertion point |
| REF-06 | `printOK`/`printErr` write to `var stdout io.Writer` | 18 calls in main.go (15 printOK + 3 printErr) + `printDryRunSummary` with 4 `fmt.Printf` calls |
| REF-02 | Normalization in create path (reduced scope) | Covered by BUG-05 implementation -- same fix |

</phase_requirements>

## Bug Fix Analysis

### BUG-01: stripEmoji() + addEmojiToSummary()

**Current bug (L1692-1706):**
```go
firstRune := []rune(s)[0]
if firstRune > 127 {
    runes := []rune(s)
    if len(runes) > 1 {
        return strings.TrimSpace(string(runes[1:]))
    }
}
```
Any non-ASCII first rune (accented chars like e, n, u, inverted punctuation) gets stripped. This destroys Spanish, Portuguese, Galician text.

**Fix:** Replace `firstRune > 127` with `unicode.Is(unicode.So, r)` (Symbol, Other category). This targets actual emoji codepoints, not all non-ASCII. Import: `unicode` (stdlib, already available).

**Same bug in addEmojiToSummary() (L1429):**
```go
if len(summary) > 0 && summary[0] > 127 {
    return summary
}
```
This is a byte-level check (`summary[0]`), not rune-level. For multi-byte UTF-8 characters, `summary[0]` is the first byte. Any character above ASCII (accented, emoji, CJK) triggers early return, preventing emoji addition. Fix: convert to rune, use `unicode.Is(unicode.So, r)`.

**Existing tests that need updating:**
- `main_utils_test.go:83` -- `{"leading high unicode", "¡Hola", "Hola"}` currently expects the accented char to be stripped. After fix, `"¡Hola"` should remain `"¡Hola"` (inverted exclamation is punctuation, not emoji).
- `main_utils_test.go:448` -- `{"already has emoji", "💊 Medicine", []string{"medication"}, false}` should still pass (emoji IS unicode.So).

**New test cases needed:**
1. `stripEmoji("Reunion con equipo")` -> `"Reunion con equipo"` (accented e preserved)
2. `stripEmoji("Ninos al colegio")` -> `"Ninos al colegio"` (n-tilde preserved)
3. `stripEmoji("uber fahrt")` -> `"uber fahrt"` (umlaut preserved)
4. `addEmojiToSummary("Reunion", []string{"work"})` -> should ADD emoji (not skip due to accented first char)
5. `addEmojiToSummary("💊 Medicine", []string{"medication"})` -> should still skip (already has emoji)

**Cascading changes:** None beyond the two functions. `stripEmoji` is called from batch processing. `addEmojiToSummary` is called from batch at L1044.

**Coverage:** `stripEmoji` is at 100% (but tests assert wrong behavior). `addEmojiToSummary` is at 50% -- opportunity to add more cases.

### BUG-02: promptAlarmField() i18n

**Current state (L3409-3462):** 15 hardcoded Spanish strings:

| # | Current String (Spanish) | Suggested i18n Key | English Value |
|---|---|---|---|
| 1 | `"Recordatorios sugeridos:"` | `alarm_prompt_suggested` | `"Suggested reminders:"` |
| 2 | `"Pulsa Enter para mantenerlos o escribe 'n' para cambiarlos"` | `alarm_prompt_keep_or_change` | `"Press Enter to keep them or type 'n' to change"` |
| 3 | `"Anade hasta 4 recordatorios..."` | `alarm_prompt_add_up_to` | `"Add up to 4 reminders. Use formats like -15m, +10m, 2025-03-01 09:15 or trigger=-15m,description=Text."` |
| 4 | `"Escribe '?' para ver ejemplos o deja vacio para terminar."` | `alarm_prompt_help_hint` | `"Type '?' for examples or leave empty to finish."` |
| 5 | `"Recordatorio #%d (-15m, +10m, trigger=..., ? para ayuda)"` | `alarm_prompt_reminder_n` | `"Reminder #%d (-15m, +10m, trigger=..., ? for help)"` |
| 6 | `"Ejemplos:"` | `alarm_prompt_examples_header` | `"Examples:"` |
| 7 | `"  -15m                 -> 15 minutos antes"` | `alarm_prompt_example_before` | `"  -15m                 -> 15 minutes before"` |
| 8 | `"  +5m                  -> 5 minutos despues"` | `alarm_prompt_example_after` | `"  +5m                  -> 5 minutes after"` |
| 9 | `"  trigger=-30m,description=Buscar taxi"` | `alarm_prompt_example_trigger` | `"  trigger=-30m,description=Find taxi"` |
| 10 | `"  trigger=2025-03-01 09:15,description=Check-in"` | `alarm_prompt_example_absolute` | `"  trigger=2025-03-01 09:15,description=Check-in"` |
| 11 | `"Descripcion opcional (Enter para usar la generica)"` | `alarm_prompt_optional_desc` | `"Optional description (Enter for default)"` |
| 12 | Format string `"  %d) %s\n"` | No key needed (format only) | -- |

**Actual count: 11 translatable strings** (not 15 -- some are format-only or duplicated in logic). The `"s"` and `"si"` acceptance strings (L3418) should also be localized:

| 13 | Accept values `"s"`, `"si"` | `alarm_prompt_yes_short` / `alarm_prompt_yes_long` | `"y"` / `"yes"` |

Total: **13 unique i18n keys** to add across 4 locale files.

**Signature change:** `promptAlarmField(label, defaultValue string)` -> `promptAlarmField(label, defaultValue string, t *i18n.Translator) string`

**Caller (L2748):**
```go
values[f.Key] = promptAlarmField(labelForField(f), f.Default)
```
Needs to pass translator. The caller context must already have a `*i18n.Translator` available -- verify at implementation time.

**New test cases:** Testing `promptAlarmField` is difficult because it reads stdin. Recommend testing that the correct keys exist in all 4 locale files (key existence test), rather than full interactive testing.

### BUG-03: expandAlarmProfiles() error return

**Current behavior (L1743-1775):** When profile not found (L1766-1767), silently passes the literal string `"profile:name"` through, which later causes a parse error with an unhelpful message.

**Fix:** Change signature from `func expandAlarmProfiles(alarmSpecs []string) []string` to `func expandAlarmProfiles(alarmSpecs []string) ([]string, error)`.

When profile not found:
```go
names := cfg.ListAlarmProfiles()
return nil, fmt.Errorf("profile '%s' not found. Available: %s", profileName, strings.Join(names, ", "))
```

**Single call site -- `addBatchAlarms()` (L1229):**
```go
expandedAlarms := expandAlarmProfiles(alarms)
```
Must become:
```go
expandedAlarms, err := expandAlarmProfiles(alarms)
if err != nil {
    // fatal per D-01
}
```

**Problem:** `addBatchAlarms()` currently returns nothing (`func addBatchAlarms(event *calendar.Event, alarms []string, startTZ string)`). Per D-01, this must be fatal. The function signature needs to change to return `error`, and its caller must handle it.

**`addBatchAlarms` caller chain:**
- Called from batch processing loop. Need to trace the exact caller to determine how to propagate the error up.

**Also check `addEventAlarms()` (L523):** This is the create-path equivalent. It does NOT call `expandAlarmProfiles` -- it calls `calendar.ParseAlarmSpecs` directly. Per D-01, both create and batch should have the same fatal behavior. The create path also needs profile expansion added, OR the decision means profiles only work in batch (current behavior). CONTEXT.md says "Same behavior in both create and batch" -- so `addEventAlarms()` at L523 should also call `expandAlarmProfiles()`.

**Existing tests:** No `TestExpandAlarmProfiles` exists. Coverage at 52.9% (from integration tests).

**New test cases needed:**
1. Valid profile name -> returns expanded alarms
2. Invalid profile name -> returns error with available profiles listed
3. Mixed valid specs and profile refs -> expands only profiles
4. Empty input -> returns empty, no error
5. Config load failure -> returns error (not silently pass through)

### BUG-04: cityToIANA() error return

**Current behavior (L3826-3875):** Returns empty string for unknown cities. Caller at L3773 checks `if mapped := cityToIANA(query); mapped != ""` and only uses it when non-empty. But the user gets no feedback about why their city wasn't recognized.

**Fix:** Change signature from `func cityToIANA(s string) string` to `func cityToIANA(s string) (string, error)`.

For unknown city:
```go
return "", fmt.Errorf("unknown city '%s'. Use 'tempus timezone list --search %s' to find the IANA identifier", s, s)
```

**Single call site (L3773):**
```go
if mapped := cityToIANA(query); mapped != "" {
```
Must become:
```go
mapped, err := cityToIANA(query)
if err != nil {
    // fatal per D-02
}
```

**Existing test (L1896-1934):** Tests unknown city expects empty string `{"unknown", ""}`. Must update to expect error return.

**New test cases needed:**
1. Known city -> returns IANA, nil error
2. Unknown city -> returns empty string, error with suggestion text
3. Empty string -> returns empty, error (or special handling)

### BUG-05 / REF-02: normalizeDateTimeInput() in create path

**Current state:** `normalizeDateTimeInput()` is only called at L1057 in the batch path:
```go
startStr = normalizeDateTimeInput(strings.TrimSpace(rec.Start))
```

**Fix per D-03:** Add normalization calls in:

1. **`parseTimedEventTimes()` (L404-405):** Before `time.Parse("2006-01-02 15:04", startStr)`, add:
```go
startStr = normalizeDateTimeInput(startStr)
```
And similarly for `endStr` in `parseEndTime()` at L438.

2. **`parseAllDayTimes()` (L378-379):** Before `time.Parse("2006-01-02", startStr)`, add:
```go
startStr = normalizeDateTimeInput(startStr)
```
And similarly for `endStr`.

**What becomes supported in create path:**
- `2025/12/16 14:00` (slash separators)
- `2025-1-5 9:00` (missing leading zeros)
- `2025-01-05 0900` (time without colon)

**Existing tests:** No `TestNormalizeDateTimeInput` unit test exists (function tested indirectly through batch integration). Coverage at 85.7%.

**New test cases needed:**
1. `normalizeDateTimeInput("2025/12/16 14:00")` -> `"2025-12-16 14:00"`
2. `normalizeDateTimeInput("2025-1-5 9:00")` -> `"2025-01-05 09:00"`
3. `normalizeDateTimeInput("2025-01-05 0900")` -> `"2025-01-05 09:00"`
4. Integration test: `parseTimedEventTimes("2025/12/16 14:00", "", "1h")` should succeed
5. Integration test: `parseAllDayTimes("2025/12/16", "")` should succeed

### REF-06: var stdout io.Writer

**Current state:** `printOK` (L3881) and `printErr` (L3887) use `fmt.Printf` directly. `printDryRunSummary` (L750-765) also uses `fmt.Printf` directly.

**Fix per D-04:**
1. Add at top of main.go: `var stdout io.Writer = os.Stdout`
2. Change `printOK`: `fmt.Printf(...)` -> `fmt.Fprintf(stdout, ...)`
3. Change `printErr`: `fmt.Printf(...)` -> `fmt.Fprintf(stdout, ...)`
4. Change `printDryRunSummary`: all `fmt.Printf(...)` -> `fmt.Fprintf(stdout, ...)`

**Call count in main.go:**
- `printOK(`: 15 calls (callers unchanged -- D-04 says zero caller changes)
- `printErr(`: 3 calls (callers unchanged)
- `printDryRunSummary`: 1 call site at L746
- Direct `fmt.Printf` in `printDryRunSummary`: 4 occurrences
- Direct `fmt.Printf`/`fmt.Println` in `writeBatchOutput`: additional calls at L769-773

**Test pattern:**
```go
func TestPrintOK(t *testing.T) {
    var buf bytes.Buffer
    stdout = &buf
    defer func() { stdout = os.Stdout }()

    printOK("test %s\n", "message")
    if !strings.Contains(buf.String(), "test message") {
        t.Errorf(...)
    }
}
```

**Existing tests (main_utils_test.go L143-145):** Already test `printErr` but output goes to real stdout. After REF-06, these tests can capture output properly.

## Test Coverage Plan

### Current Coverage for Target Functions

| Function | Current Coverage | Existing Tests | New Tests Needed |
|----------|-----------------|----------------|------------------|
| `stripEmoji` | 100% | `TestStripEmoji` (6 cases) | Update 1 case, add 3 accent cases |
| `addEmojiToSummary` | 50% | `TestAddEmojiToSummary` (8 cases) | Add accent-start cases |
| `expandAlarmProfiles` | 52.9% | None (indirect only) | New `TestExpandAlarmProfiles` (5 cases) |
| `cityToIANA` | 78.3% | `TestCityToIANA` (16 cases) | Update unknown case to check error |
| `normalizeDateTimeInput` | 85.7% | None (indirect only) | New `TestNormalizeDateTimeInput` (5 cases) |
| `printOK` | 100% | Indirect | New test using captured stdout |
| `printErr` | 100% | Indirect | New test using captured stdout |
| `promptAlarmField` | 54.1% | None | Key existence test across locales |
| `parseTimedEventTimes` | 58.3% | Indirect | Add normalized-input cases |
| `parseAllDayTimes` | 0% | None | Add slash-separator cases |
| `addBatchAlarms` | 88.9% | Indirect | Add profile-not-found error case |

### Test Files

| File | Content |
|------|---------|
| `main_utils_test.go` | BUG-01 tests (stripEmoji, addEmojiToSummary), REF-06 tests (printOK, printErr) |
| `main_alarm_test.go` | BUG-03 tests (expandAlarmProfiles), alarm integration |
| `main_test.go` | BUG-04 tests (cityToIANA update), BUG-05 tests (parseTimedEventTimes, parseAllDayTimes), normalizeDateTimeInput unit tests |
| `locales/*_test.go` or `internal/i18n/*_test.go` | BUG-02: verify all 13 keys exist in all 4 locales |

## Implementation Order

Recommended order to minimize merge conflicts and build testing infrastructure first:

1. **REF-06: var stdout io.Writer** (no functional change, enables better tests for everything else)
2. **BUG-01: stripEmoji + addEmojiToSummary** (simplest fix, 2 functions, highest user-facing impact)
3. **BUG-03: expandAlarmProfiles error return** (error handling, changes addBatchAlarms signature)
4. **BUG-04: cityToIANA error return** (error handling, same pattern as BUG-03)
5. **BUG-02: promptAlarmField i18n** (most files touched -- 4 locale files + main.go + i18n constants)
6. **BUG-05 / REF-02: normalizeDateTimeInput in create path** (last -- touches parsing functions)

**Rationale:** REF-06 first gives test infrastructure. BUG-01 is independent. BUG-03/04 are a natural pair (same error-handling pattern). BUG-02 touches most files but is isolated. BUG-05 is last because it modifies parsing functions that could interact with other test changes.

## Common Pitfalls

### Pitfall 1: Unicode Category Precision
**What goes wrong:** Using `unicode.Is(unicode.So, r)` alone may miss some emoji (emoji modifiers, ZWJ sequences, flags are not all in Symbol/Other).
**Why it happens:** Emoji span multiple unicode categories (So, Sk, Mn for modifiers, regional indicators).
**How to avoid:** For `stripEmoji`, the scope is "first character is an emoji prefix" -- `unicode.So` covers the common emoji used in this codebase (pill, checkmark, cross, etc.). For robustness, also check `unicode.Is(unicode.Sk, r)` (Symbol, Modifier). Test with actual emoji from the `addEmojiToSummary` map.
**Warning signs:** Test with flag emoji, skin tone modifiers, ZWJ sequences if broader support needed.

### Pitfall 2: addBatchAlarms Error Propagation
**What goes wrong:** `addBatchAlarms` returns void. Changing it to return error requires updating its caller chain. If the caller is inside a loop processing multiple batch records, an early fatal could lose partial progress.
**Why it happens:** Per D-01, profile-not-found is fatal. But in batch mode, should it fail the entire batch or skip that event?
**How to avoid:** D-01 is explicit -- "abort execution". Propagate error up from `addBatchAlarms` to the batch loop, which should abort.
**Warning signs:** Existing tests that process multiple batch records may break if one has an invalid profile.

### Pitfall 3: Existing Test Assertions on Buggy Behavior
**What goes wrong:** Tests pass today but assert incorrect behavior. Fixing the bug breaks tests.
**Why it happens:** `TestStripEmoji` case `"leading high unicode"` expects `"¡Hola"` -> `"Hola"`. `TestCityToIANA` expects `""` for unknown cities.
**How to avoid:** Update test expectations in the same commit as the fix. List all test cases that need updating before implementing.
**Warning signs:** Test failures after fix are EXPECTED -- the bug was "tested" as correct behavior.

### Pitfall 4: printDryRunSummary Has Additional fmt.Printf Calls
**What goes wrong:** REF-06 might only update `printOK`/`printErr` but miss the direct `fmt.Printf` calls in `printDryRunSummary` and `writeBatchOutput`.
**Why it happens:** The variable name change is grep-friendly for `printOK`/`printErr` but misses inline `fmt.Printf`.
**How to avoid:** Also update `printDryRunSummary` (L750-765) -- 4 `fmt.Printf` calls. Consider `writeBatchOutput` (L769-773) for completeness.
**Warning signs:** Output not captured in tests for batch dry-run mode.

### Pitfall 5: addEventAlarms (create path) Doesn't Expand Profiles
**What goes wrong:** D-01 says "Same behavior in both create and batch -- no permissive mode." But `addEventAlarms()` (L523) does NOT call `expandAlarmProfiles()`. Only `addBatchAlarms()` (L1219) does.
**Why it happens:** Create path handles alarms differently from batch path.
**How to avoid:** Either add `expandAlarmProfiles()` call to `addEventAlarms()`, or document that profile expansion is batch-only. D-01 implies both paths need it.
**Warning signs:** `tempus create --alarm "profile:adhd-triple"` silently fails to expand.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None needed (Go convention) |
| Quick run command | `go test ./... -count=1 -short` |
| Full suite command | `go test ./... -count=1 -race` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| BUG-01 | Accented chars preserved by stripEmoji/addEmojiToSummary | unit | `go test -run TestStripEmoji -v` | Yes (update needed) |
| BUG-02 | promptAlarmField uses i18n keys | unit | `go test -run TestAlarmPromptI18nKeys -v` | No -- Wave 0 |
| BUG-03 | expandAlarmProfiles returns error for missing profile | unit | `go test -run TestExpandAlarmProfiles -v` | No -- Wave 0 |
| BUG-04 | cityToIANA returns error for unknown city | unit | `go test -run TestCityToIANA -v` | Yes (update needed) |
| BUG-05 | parseTimedEventTimes accepts slash dates, missing zeros | unit | `go test -run TestParseTimedEventTimes -v` | No -- Wave 0 |
| REF-06 | printOK/printErr write to var stdout | unit | `go test -run TestPrintOK -v` | Partial (update needed) |
| REF-02 | Same as BUG-05 | -- | -- | -- |

### Sampling Rate
- **Per task commit:** `go test ./... -count=1 -short`
- **Per wave merge:** `go test ./... -count=1 -race`
- **Phase gate:** Full suite green + `go test ./... -coverprofile=cover.out && go tool cover -func=cover.out | grep total`

### Wave 0 Gaps
- [ ] `TestExpandAlarmProfiles` in `main_alarm_test.go` -- covers BUG-03
- [ ] `TestNormalizeDateTimeInput` in `main_test.go` -- covers BUG-05
- [ ] `TestAlarmPromptI18nKeys` in locale test file -- covers BUG-02
- [ ] Update `TestStripEmoji` expectations -- covers BUG-01
- [ ] Update `TestCityToIANA` expectations for error return -- covers BUG-04

## Project Constraints (from CLAUDE.md)

- No comments unless explicitly requested
- Python: type hints, f-strings, no bare excepts (N/A -- Go project)
- Shell: `set -euo pipefail` (for any scripts)
- Always verify work: run lint/test/validate before declaring done
- Prefer idempotent solutions
- Never hardcode secrets
- Coverage gate: 79% minimum enforced every phase
- No breaking changes to CLI API (flags and outputs must remain compatible)
- All ICS output must remain RFC 5545 valid

## Sources

### Primary (HIGH confidence)
- Direct code inspection of `main.go` -- all target functions read and analyzed
- Direct code inspection of `internal/i18n/i18n.go` -- translator API confirmed
- Direct code inspection of `locales/en.json` -- existing key inventory
- Direct code inspection of test files -- coverage gaps identified
- `go test ./... -coverprofile` -- function-level coverage numbers verified

### Secondary (MEDIUM confidence)
- Go stdlib `unicode` package -- `unicode.So` category for emoji detection (well-established, stable API)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- Go stdlib only, no new dependencies
- Architecture: HIGH -- all changes within existing monolith, patterns clear
- Pitfalls: HIGH -- code paths fully traced, caller chains verified

**Research date:** 2026-03-29
**Valid until:** 2026-04-28 (stable codebase, no external dependency changes)
