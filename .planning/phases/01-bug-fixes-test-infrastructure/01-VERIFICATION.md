---
phase: 01-bug-fixes-test-infrastructure
verified: 2026-03-29T23:45:00Z
status: passed
score: 10/10 must-haves verified
re_verification:
  previous_status: gaps_found
  previous_score: 9/10
  gaps_closed:
    - "Coverage gate >= 79%: now 79.7% (was 78.7%)"
  gaps_remaining: []
  regressions: []
---

# Phase 01: Bug Fixes & Test Infrastructure Verification Report

**Phase Goal:** Users get correct ICS output for all locales, proper error messages, and consistent input normalization across create and batch paths
**Verified:** 2026-03-29T23:45:00Z
**Status:** passed
**Re-verification:** Yes — after gap closure (plan 01-04)

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | stripEmoji preserves accented Latin characters (e-acute, n-tilde, u-umlaut, inverted punctuation) and only strips actual emoji | VERIFIED | `unicode.Is(unicode.So, firstRune)` at main.go:1461; TestStripEmoji/accented_e, n_tilde, umlaut, leading_high_unicode all PASS |
| 2 | addEmojiToSummary adds emoji to summaries starting with accented characters instead of skipping them | VERIFIED | `unicode.Is(unicode.So, firstRune)` at main.go:1729; TestAddEmojiToSummary/accented_start_gets_emoji PASS |
| 3 | printOK, printErr, printDryRunSummary, writeBatchOutput write to var stdout io.Writer | VERIFIED | `var stdout io.Writer = os.Stdout` at main.go:37; all 4 functions use `fmt.Fprintf(stdout, ...)` / `fmt.Fprintln(stdout, ...)`; zero residual `fmt.Printf` in those functions |
| 4 | Tests can capture output from printOK/printErr by overriding the stdout variable | VERIFIED | TestPrintOK and TestPrintErr in main_utils_test.go assign `bytes.Buffer` to `stdout`, defer restore; both PASS |
| 5 | Missing alarm profile aborts with descriptive error listing available profiles | VERIFIED | `expandAlarmProfiles` returns `([]string, error)` at main.go:1774; error message: `"profile '%s' not found. Available: %s"`; TestExpandAlarmProfiles/unknown_profile and unknown_profile_lists_available PASS |
| 6 | Unknown city aborts with error suggesting tempus timezone list --search (no fuzzy fallback) | VERIFIED | `cityToIANA` returns `(string, error)` at main.go:3856; error contains "Unknown city" and "tempus timezone list --search"; caller at main.go:3803 returns immediately on cityErr before fuzzy fallback; TestCityToIANA/unknown and #00 (empty) PASS |
| 7 | promptAlarmField displays all user-facing strings in configured language | VERIFIED | Signature `func promptAlarmField(label, defaultValue string, t *i18n.Translator)` at main.go:3438; all 13 strings replaced with `t.T("alarm_prompt_*")` calls; zero hardcoded Spanish strings remain; TestAlarmPromptI18nKeys verifies all 13 keys in all 4 locales PASS |
| 8 | Profile expansion works in both create and batch paths | VERIFIED | `addEventAlarms` (create path, main.go:538) calls `expandAlarmProfiles` and returns error; `addBatchAlarms` (batch path, main.go:1245) calls `expandAlarmProfiles` and returns error; both propagated to callers |
| 9 | User can use slash dates (2025/12/16), missing zeros (2025-1-5), colon-less times (0900) in tempus create | VERIFIED | `normalizeDateTimeInput` called at 4 insertion points: parseAllDayTimes startStr (L385), parseAllDayTimes endStr (L394), parseTimedEventTimes startStr (L413), parseEndTime endStr (L447); TestParseTimedEventTimesNormalization/slash_separators and missing_leading_zeros PASS; TestParseAllDayTimesNormalization/slash_separators and missing_leading_zeros PASS |
| 10 | Coverage gate >= 79% satisfied | VERIFIED | Coverage is 79.7% (1478 tests passing across 11 packages). Gate satisfied with +0.7pp margin. |

**Score:** 10/10 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `main.go` | var stdout, fixed stripEmoji/addEmojiToSummary, expandAlarmProfiles ([]string,error), cityToIANA (string,error), promptAlarmField with Translator, normalizeDateTimeInput in create path | VERIFIED | All patterns confirmed via grep and code read |
| `main_utils_test.go` | TestStripEmoji updated, TestPrintOK, TestPrintErr | VERIFIED | All 3 functions present and passing |
| `main_alarm_test.go` | TestExpandAlarmProfiles with error cases | VERIFIED | Function exists; 4 sub-cases including error cases pass |
| `main_test.go` | TestCityToIANA updated, TestAlarmPromptI18nKeys, TestNormalizeDateTimeInput, TestParseTimedEventTimesNormalization, TestParseAllDayTimesNormalization | VERIFIED | All 5 functions present; all pass |
| `locales/en.json` | 13 alarm_prompt_* keys | VERIFIED | 13 keys confirmed with English translations |
| `locales/es.json` | 13 alarm_prompt_* keys | VERIFIED | 13 keys confirmed with Spanish translations |
| `locales/pt.json` | 13 alarm_prompt_* keys | VERIFIED | 13 keys confirmed with Portuguese translations |
| `locales/ga.json` | 13 alarm_prompt_* keys | VERIFIED | 13 keys confirmed with Galician translations |
| `internal/config/config_test.go` | Tests for GetAlarmProfile and ListAlarmProfiles | VERIFIED | Added by plan 01-04; pushed total coverage above 79% gate |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| main.go:printOK | main.go:stdout | `fmt.Fprintf(stdout,` | WIRED | main.go:3916 confirmed |
| main.go:printErr | main.go:stdout | `fmt.Fprintf(stdout,` | WIRED | main.go:3921 confirmed |
| main.go:stripEmoji | unicode.Is(unicode.So, r) | unicode category check | WIRED | main.go:1461 confirmed |
| main.go:addEmojiToSummary | unicode.Is(unicode.So, r) | unicode category check | WIRED | main.go:1729 confirmed |
| main.go:addBatchAlarms | main.go:expandAlarmProfiles | error propagation up to configureBatchEvent -> buildEventFromBatch | WIRED | configureBatchEvent at L1192 returns error; `if err := configureBatchEvent(...); err != nil` at L1067 |
| main.go:addEventAlarms | main.go:expandAlarmProfiles | profile expansion in create path | WIRED | main.go:538; caller at L499: `if err := addEventAlarms(...); err != nil` |
| main.go:promptAlarmField | i18n.Translator.T | translator parameter for all user-facing strings | WIRED | 13 `t.T("alarm_prompt_*")` calls confirmed; caller at L2777 passes `tr` |
| main.go:parseTimedEventTimes | main.go:normalizeDateTimeInput | normalize startStr before time.Parse | WIRED | main.go:413 |
| main.go:parseAllDayTimes | main.go:normalizeDateTimeInput | normalize startStr before time.Parse | WIRED | main.go:385 (startStr), L394 (endStr) |
| main.go:parseEndTime | main.go:normalizeDateTimeInput | normalize endStr before time.Parse | WIRED | main.go:447 |

### Data-Flow Trace (Level 4)

Not applicable — this phase produces no components that render dynamic data. All artifacts are pure functions, test infrastructure, and locale data files.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| TestStripEmoji passes with accented chars | `go test -run TestStripEmoji -v .` | 9 sub-cases PASS | PASS |
| TestPrintOK/TestPrintErr capture output via stdout override | `go test -run "TestPrintOK\|TestPrintErr" -v .` | Both PASS | PASS |
| TestExpandAlarmProfiles returns error for unknown profile | `go test -run TestExpandAlarmProfiles -v .` | 4 sub-cases PASS including "not found" and "Available:" | PASS |
| TestCityToIANA returns error with search suggestion | `go test -run TestCityToIANA -v .` | 16 sub-cases PASS | PASS |
| TestAlarmPromptI18nKeys verifies all 13 keys in 4 locales | `go test -run TestAlarmPromptI18nKeys -v .` | PASS | PASS |
| TestParseTimedEventTimesNormalization accepts slash/zero-padded dates | `go test -run TestParseTimedEventTimesNormalization -v .` | 3 sub-cases PASS | PASS |
| Full test suite passes without regressions | `go test ./... -count=1` | 1478 tests PASS, 11 packages | PASS |
| Coverage gate >= 79% | `go test ./... -coverprofile=cover.out && go tool cover -func=cover.out \| grep total` | **79.7%** | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| BUG-01 | 01-01 | stripEmoji uses unicode.Is(unicode.So) not rune > 127 | SATISFIED | main.go:1461,1729; TestStripEmoji/accented_e,n_tilde,umlaut PASS |
| BUG-02 | 01-02 | promptAlarmField uses i18n system | SATISFIED | 13 t.T() calls in promptAlarmField; TestAlarmPromptI18nKeys PASS for en/es/pt/ga |
| BUG-03 | 01-02 | expandAlarmProfiles returns ([]string, error) | SATISFIED | Signature confirmed; error with profile listing; TestExpandAlarmProfiles PASS |
| BUG-04 | 01-02 | cityToIANA returns error with search suggestion | SATISFIED | Signature confirmed; error message with "tempus timezone list --search"; TestCityToIANA PASS |
| BUG-05 | 01-03 | normalizeDateTimeInput applied in parseCreateTimes path | SATISFIED | 4 insertion points confirmed; TestParseTimedEventTimesNormalization, TestParseAllDayTimesNormalization PASS |
| REF-02 | 01-03 | Date parsing unified — normalizeDateTimeInput single entry point for create path | SATISFIED | create path now uses same normalizeDateTimeInput as batch path |
| REF-06 | 01-01 | printOK, printErr accept io.Writer via package-level var | SATISFIED | var stdout io.Writer at main.go:37; TestPrintOK, TestPrintErr PASS |

All 7 requirements claimed by the phase plans are SATISFIED. No orphaned requirements detected.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| main.go | 294 | `return fmt.Errorf("interactive mode not yet implemented")` | Info | Pre-existing stub for runInteractive — not related to this phase's scope; UX-02 is Phase 3 work |

No blockers or warnings found in files modified by this phase.

### Human Verification Required

None — all acceptance criteria are programmatically verifiable.

### Gaps Summary

No gaps remain. The sole gap from the initial verification (coverage gate) was closed by plan 01-04, which added tests for `GetAlarmProfile` and `ListAlarmProfiles` in `internal/config`, pushing total coverage from 78.7% to 79.7%.

All functional requirements (BUG-01 through BUG-05, REF-02, REF-06) are fully implemented, tested, and wired. The phase goal of correct ICS output, proper error messages, and consistent input normalization is achieved.

---

_Verified: 2026-03-29T23:45:00Z_
_Verifier: Claude (gsd-verifier)_
