# Pitfalls Research

**Project:** Tempus (Go CLI for ICS generation)
**Researched:** 2026-03-29
**Overall confidence:** HIGH (based on Go stdlib docs, codebase analysis, Viper official docs)

---

## Unicode Emoji Detection in Go

**Confidence:** HIGH (verified against Go `unicode` package docs at pkg.go.dev)

### The Bug

Current `stripEmoji()` uses `rune > 127` to detect emoji. This is wrong because:
- Latin Extended characters (e, n, u, a, i, o with accents) are all above 127
- `e` = U+00E9 (233), `n` = U+00F1 (241), `u` = U+00FC (252)
- The function strips the first character of "Reunion de equipo" turning it into "eunion de equipo"

### Correct Approach: `unicode.In()` with Category Tables

Use Go's `unicode` package to check if a rune belongs to emoji-related Unicode categories. The key insight: emojis are primarily in `unicode.So` (Symbol, Other) and `unicode.Sk` (Symbol, Modifier), while accented Latin characters are in `unicode.Letter` / `unicode.Latin`.

```go
func isEmoji(r rune) bool {
    return unicode.Is(unicode.So, r) || unicode.Is(unicode.Sk, r) ||
        unicode.Is(unicode.Mn, r) && !unicode.Is(unicode.Latin, r)
}

func stripEmoji(s string) string {
    s = strings.TrimSpace(s)
    runes := []rune(s)
    // Strip leading emoji runes (some emoji are multi-rune sequences)
    i := 0
    for i < len(runes) && isEmoji(runes[i]) {
        i++
    }
    if i > 0 && i < len(runes) {
        return strings.TrimSpace(string(runes[i:]))
    }
    return s
}
```

### Pitfalls to Avoid

1. **Do NOT use a regex like `[\x{1F600}-\x{1F64F}]`** -- emoji ranges are scattered across Unicode and updated with every Unicode version. Hardcoded ranges become stale.

2. **Do NOT use `rivo/uniseg` for this** -- it provides grapheme cluster segmentation, not emoji classification. Overkill for stripping a leading emoji prefix, and it is not currently in go.mod.

3. **Multi-codepoint emoji sequences**: Flag emoji (U+1F1E6..U+1F1FF pairs), skin tone modifiers (U+1F3FB..U+1F3FF), ZWJ sequences. The `unicode.So` approach handles the base codepoints correctly. The combining characters (variation selectors U+FE0E/U+FE0F, ZWJ U+200D) fall under `unicode.Cf` (Format, Other) -- also strip those in the leading sequence.

4. **Edge case: digits with emoji presentation** (e.g., "1" U+0031 + U+FE0F + U+20E3 = keycap 1). The base character is ASCII. Handle by checking for variation selector followers, or accept that keycap emoji at string start is extremely rare for calendar event titles.

5. **Test with actual Tempus data**: "Reunion de equipo" must survive intact. "Cita medica" must survive. Test ALL four supported locales (en, es, pt, ga) -- Galician has accented chars too.

### Recommended Test Cases

```
Input                    | Expected Output
"Meeting"                | "Meeting"           (no emoji, ASCII only)
"Reunion"                | "Reunion"           (accented, must NOT strip)
"Cafe con leche"         | "Cafe con leche"    (accented, must NOT strip)
"Preparacao"             | "Preparacao"        (Portuguese cedilla)
"Team Standup"           | "Team Standup"      (emoji stripped)
"Birthday"               | "Birthday"          (multi-rune emoji stripped)
"Hola"                   | "Hola"              (inverted exclamation, NOT an emoji)
""                       | ""                  (empty string)
"Solo emoji"             | "" or "Solo emoji"  (decide: return empty or original)
```

---

## i18n in Go CLI Prompts -- Pitfalls

**Confidence:** HIGH (based on codebase analysis of existing i18n system)

### The Bug

`promptAlarmField()` has 10+ Spanish hardcoded strings ("Recordatorios sugeridos", "Pulsa Enter para mantenerlos", "Anade hasta 4 recordatorios", etc.) while the rest of the CLI uses `i18n.Translator.T()`.

### The Existing i18n System

Tempus already has a working i18n system (`internal/i18n/`) with:
- `Translator.T(key, args...)` for translation lookup
- English fallback for missing keys
- 4 embedded locales: en, es, pt, ga
- JSON locale files with flat key-value structure

The fix is straightforward: add translation keys to locale files, pass `Translator` to `promptAlarmField()`.

### Pitfalls When Adding i18n to Existing Prompts

1. **Missing the Translator reference**: `promptAlarmField()` currently takes `(label, defaultValue string)`. It needs access to a `*i18n.Translator`. Three options:
   - Pass it as parameter (cleanest, most testable)
   - Use a package-level variable (easy but couples to global state)
   - Create a struct that holds both prompt functions and translator

   **Recommendation:** Pass as parameter. Refactoring to a struct is R1 work and should not block B2.

2. **Hardcoded format strings with interpolation**: Lines like `fmt.Sprintf("Recordatorio #%d (...)", len(specs)+1)` need translation keys with `%d` placeholders. The `T()` function already supports `fmt.Sprintf` args. Make sure the placeholder position is consistent across locales (it usually is for simple numbered items, but verify with Portuguese/Galician translators).

3. **Forgetting to add keys to ALL locale files**: The system falls back to English if a key is missing. This is safe but produces a mixed-language UI. Add keys to all 4 locales (en, es, pt, ga) or at minimum en + es (the two most tested).

4. **User-facing strings in `promptInput()` calls**: The label passed to `promptInput()` inside `promptAlarmField()` is also hardcoded Spanish. Grep for ALL `promptInput()` calls across main.go to find other hardcoded strings -- there may be more beyond `promptAlarmField()`.

5. **Help text ("?") also hardcoded**: The examples shown when user types "?" are in Spanish ("15 minutos antes", "5 minutos despues"). These need translation too, or they become confusing in English mode.

6. **Pluralization trap**: Go's `text/message` package handles plural forms, but Tempus's i18n is simpler (flat key-value). For "Recordatorio #1" vs "Recordatorio #2" this is fine (cardinal number, no plural change in the template). Do NOT introduce `golang.org/x/text/message` for this -- overkill for current needs.

7. **Comparison operators in prompts**: `keep == "s" || keep == "si"` is Spanish-specific. English users should respond "y" or "yes". The accepted responses must be locale-aware too.

---

## Error vs Warn in CLI Tools

**Confidence:** HIGH (established CLI design patterns)

### The Bugs

- B3: `expandAlarmProfiles()` silently keeps `"profile:adhd-triple"` as literal text when profile not found. This string passes through to ICS output as an invalid alarm spec.
- B4: `cityToIANA()` returns `""` for unknown cities, caller continues silently.

### When to Error (Exit Non-Zero)

Use `error` (fail loudly) when:
- **Data corruption would occur**: B3 is this case. A literal "profile:name" string in ICS output is corrupt data. The user thinks they have alarms but they do not.
- **The user explicitly requested something**: If `--alarm profile:adhd-triple` was typed, failing silently is a lie.
- **No reasonable default exists**: An unknown city has no valid timezone fallback.

### When to Warn (Print Warning, Continue)

Use `warning` (print to stderr, continue) when:
- **A reasonable default exists**: e.g., timezone defaults to config value when city lookup fails.
- **The operation can partially succeed**: e.g., batch processing where 1 of 50 rows has a bad city -- warn per row, continue with others.
- **The feature is supplementary**: e.g., spell-check suggestions ("Did you mean 'meeting'?").

### Recommended Fixes

**B3 (alarm profile not found):**
```
Error: alarm profile "adhd-triple" not found
Available profiles: adhd-default, adhd-countdown, medication, single, none
Hint: run 'tempus alarm list' to see all profiles
```
Return an error from `expandAlarmProfiles()` -- change signature to `([]string, error)`. Caller decides whether to abort (single create) or warn-and-skip (batch).

**B4 (cityToIANA unknown city):**
```
Warning: city "Timbuktu" not recognized, using configured timezone (Europe/Madrid)
Hint: run 'tempus timezone search Timbuktu' to find the correct timezone
```
Return `("", ErrCityNotFound)` and let the caller either fall back to config timezone with a warning, or error if no fallback is configured.

### Pitfalls in Error Handling Patterns

1. **Comment lies**: Line 1766 says `"will error later"` but `calendar.ParseAlarmSpecs` does NOT error on a `"profile:name"` string -- it either silently produces garbage or is silently ignored. Verify the actual downstream behavior before assuming another layer catches it.

2. **Batch vs single asymmetry**: In batch mode, one bad row should not abort the entire batch. Use a collected-errors pattern: process all rows, collect warnings/errors, report at end. In single `create` mode, fail immediately.

3. **Error messages must be actionable**: "profile not found" is useless. "profile 'adhd-triple' not found. Available: adhd-default, medication, single" is actionable. Always include: what failed, what was expected, what to do next.

4. **Stderr vs stdout**: Warnings go to stderr (`fmt.Fprintf(os.Stderr, ...)`). Normal output goes to stdout. This allows piping (`tempus create ... > event.ics`) while still seeing warnings.

5. **Exit codes**: Use distinct exit codes for different failure modes if possible. At minimum: 0 = success, 1 = user error (bad input), 2 = system error (can't write file). Cobra handles this partially but verify.

---

## Go Monolith Split -- Common Mistakes

**Confidence:** HIGH (Go project structure is well-documented territory)

### The Problem

`main.go` is 3,906 lines. Target (R1): split into `internal/commands/`, `internal/parsing/`, `internal/nd/`, `internal/output/`.

### Pitfall 1: Import Cycles

Go forbids circular imports. This is the number one cause of failed refactors.

**How it happens:**
- You move `stripEmoji()` to `internal/utils/` and `expandAlarmProfiles()` to `internal/commands/`
- `commands` imports `utils` for `stripEmoji()`
- `utils` imports `config` for alarm profiles
- `config` imports `i18n` for validation
- If `commands` also imports `i18n`, no cycle. But if `i18n` ever imports `commands` -- cycle.

**Prevention:**
- **Direction rule**: Dependencies flow downward: `commands` -> `parsing` / `nd` / `output` -> `config` / `i18n` / `utils`. Never upward.
- **Interface at the boundary**: If a lower package needs behavior from a higher one, define an interface in the lower package and have the higher package implement it.
- **Draw the dependency graph before moving code.** Use `go mod graph` or mentally trace imports.

**Detection:**
```bash
go build ./...
```
Go will refuse to compile with a clear "import cycle" error. Run this after every move.

### Pitfall 2: Unexported Functions Becoming Inaccessible

When `stripEmoji()` is in `main.go`, it is lowercase (unexported) but accessible to all test files in package `main`. When you move it to `internal/utils/`, it must be exported (`StripEmoji()`) or tests must move into the same package.

**Prevention:** When moving a function to a new package:
- Export it (capitalize) if it is used from outside the package
- Move its tests to the new package, or write new tests there
- Keep the old test as an integration test calling the new exported function

### Pitfall 3: Global State (Viper, Config)

`main.go` uses package-level variables and calls `config.Load()` in multiple places. When split across packages:
- Each package that calls `config.Load()` independently may get a different Viper state
- Viper uses a global singleton by default -- this actually helps (shared state) but makes testing hard

**Prevention:**
- Load config ONCE in `main()` or the root Cobra command's `PersistentPreRun`
- Pass `*config.Config` as a parameter to functions that need it
- Do NOT have multiple packages independently calling `config.Load()`

### Pitfall 4: Moving Too Much at Once

Attempting to restructure the entire 3,906-line file in one PR guarantees merge conflicts, broken tests, and reviewer fatigue.

**Prevention:** Move one logical group at a time:
1. First: utility functions (`stripEmoji`, `generateUID`, etc.) to `internal/utils/`
2. Second: parsing functions to `internal/parsing/`
3. Third: neurodivergent features to `internal/nd/`
4. Fourth: output formatting to `internal/output/`
5. Last: command definitions to `internal/commands/`

Each move is one commit. Tests pass after each commit.

### Pitfall 5: Package Naming

- Do NOT name packages `util`, `common`, `helpers`, or `misc` -- they become dumping grounds
- `internal/nd/` is good (neurodivergent features, clear domain)
- `internal/parsing/` is good (date/time parsing, clear responsibility)
- `internal/output/` is good (ICS formatting, print helpers)

### Pitfall 6: Test File Coupling

The 8 test files (`main_*_test.go`) are all in `package main`. They directly call unexported functions. When those functions move to internal packages:
- Tests break immediately (unexported function no longer in scope)
- Must either: (a) export the function, (b) move the test, or (c) create a thin wrapper in main

**Recommendation:** (a) Export and move tests. Creates proper package-level test coverage.

---

## Test Coverage Protection During Refactor

**Confidence:** HIGH (standard Go tooling)

### Current State

- 79% coverage across 353 test functions in 20 files
- Tests tightly coupled to `package main` (call unexported functions directly)

### Pitfall 1: Coverage Drops When Moving Functions

When `stripEmoji` moves from `main.go` to `internal/utils/emoji.go`:
- `main_utils_test.go::TestStripEmoji` no longer compiles (function not in scope)
- If you delete the test, coverage drops
- If you update the test to call `utils.StripEmoji()`, it now counts toward `internal/utils` coverage, not `main` coverage

**Prevention:** Track coverage at the module level, not per-package:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
```
This gives total coverage across all packages. Set a CI gate: `if total < 79%, fail`.

### Pitfall 2: Dead Code Masking

Moving code reveals dead code. Functions that were "covered" because they were called by other tested functions in the same file may have no direct test. When isolated in a new package, coverage for that function drops to 0%.

**Prevention:** Before moving a function, check its individual coverage:
```bash
go test -coverprofile=c.out -run "TestStripEmoji" . && go tool cover -func=c.out | grep stripEmoji
```

### Pitfall 3: Test Helpers in Wrong Package

`internal/testutil/` exists. Good. But test helpers defined in `main_test.go` (like setup/teardown functions) can not be imported by tests in `internal/` packages.

**Prevention:** Move shared test helpers to `internal/testutil/` FIRST, before moving the code they support.

### Pitfall 4: Integration vs Unit Test Confusion

After the split, `main_test.go` should become integration tests (test the CLI commands end-to-end), while `internal/*/` packages have unit tests. Do not duplicate test logic in both places.

### Coverage Protection Strategy

1. **Before refactor**: Run `go test -coverprofile=baseline.out ./...` and record the number
2. **CI gate**: Add a step that fails if coverage drops below baseline minus 1% (allow small fluctuations)
3. **Per-PR**: Show coverage diff in PR description
4. **After refactor**: The new package structure should have HIGHER coverage (better isolated tests)

---

## Viper AutomaticEnv Gotchas

**Confidence:** HIGH (verified against Viper official docs at pkg.go.dev)

### The Bug

F3 calls for env var support (`TEMPUS_TIMEZONE`, `TEMPUS_LANGUAGE`, `TEMPUS_OUTPUT_DIR`). The plan is to use `AutomaticEnv()`. Current `config.go` does NOT call `AutomaticEnv()`.

### Pitfall 1: Missing `SetEnvPrefix`

Without `SetEnvPrefix("TEMPUS")`, `AutomaticEnv()` maps `viper.Get("timezone")` to env var `TIMEZONE`, not `TEMPUS_TIMEZONE`.

**Fix:**
```go
viper.SetEnvPrefix("TEMPUS")
viper.AutomaticEnv()
```

Now `viper.Get("timezone")` checks `TEMPUS_TIMEZONE`.

### Pitfall 2: Underscore Key Mapping

Config keys use underscores: `output_dir`, `date_format`. Viper converts dots to underscores by default with `SetEnvKeyReplacer`, but underscored keys map directly. So `output_dir` with prefix `TEMPUS` maps to `TEMPUS_OUTPUT_DIR`. This works correctly without a replacer for this codebase.

### Pitfall 3: Precedence Surprise

Viper precedence: `Set()` > flags > env > config file > defaults.

This means `TEMPUS_TIMEZONE=America/New_York` OVERRIDES the config file value. This is usually desired, but document it clearly. Users who set an env var and then wonder why `tempus config set timezone X` seems to not work will be confused.

**Prevention:** In `tempus config list`, show the effective value AND its source:
```
timezone: America/New_York (from env: TEMPUS_TIMEZONE)
```

### Pitfall 4: `AutomaticEnv` Must Be Called Before `Get`

`AutomaticEnv()` must be called before any `viper.Get()` calls. In the current `Load()` function, add it before `ReadInConfig()`.

### Pitfall 5: Empty Env Vars

By default, Viper treats empty env vars as unset. If a user sets `TEMPUS_TIMEZONE=""`, Viper ignores it and uses the config file or default value.

If you want empty to mean "use default explicitly", this is fine (current behavior). If you want empty to mean "clear the setting", call `viper.AllowEmptyEnv(true)`. For Tempus, the default behavior (ignore empty) is correct -- an empty timezone is never valid.

### Pitfall 6: `Unmarshal` After `AutomaticEnv`

The current code does:
```go
viper.ReadInConfig()
viper.Unmarshal(&cfg)
```

With `AutomaticEnv()`, `Unmarshal` correctly picks up env var overrides because it calls `Get()` internally for each field. This works. But there is a subtle issue: `mapstructure` tags must match the Viper key names exactly. Current tags (`mapstructure:"timezone"`) match the Viper keys. No issue here.

### Pitfall 7: Viper Global Singleton in Tests

Viper's default instance is a global singleton. Tests that call `config.Load()` pollute each other. This is already a risk in the current codebase, but adding `AutomaticEnv()` makes it worse because tests that set `os.Setenv("TEMPUS_TIMEZONE", ...)` leak into other tests.

**Prevention:**
- Use `t.Setenv()` in tests (automatically cleaned up)
- Or use `viper.New()` to create isolated Viper instances per test (requires refactoring `config.Load()` to accept a Viper instance)

---

## Phase Mapping

| Pitfall | Affects Bug/Feature | Recommended Phase | Rationale |
|---------|-------------------|-------------------|-----------|
| Unicode emoji detection | B1 (`stripEmoji`) | Phase 1 (Bug Fixes) | Critical bug, breaks Spanish/Portuguese users. Self-contained fix. |
| i18n prompt hardcoding | B2 (`promptAlarmField`) | Phase 1 (Bug Fixes) | Must add ~15 translation keys to 4 locale files. Depends on existing i18n system. |
| Alarm profile silent failure | B3 | Phase 1 (Bug Fixes) | Data corruption risk. Change function signature to return error. |
| cityToIANA silent failure | B4 | Phase 1 (Bug Fixes) | Change to return error, caller decides fallback vs fail. |
| Input normalization in create | B5 | Phase 1 (Bug Fixes) | Reuse existing normalizer package -- it already works for batch. |
| Viper AutomaticEnv setup | F3 (env vars) | Phase 2 (Features) | Requires `SetEnvPrefix` + `AutomaticEnv()` in config.Load(). Low risk. |
| Import cycle prevention | R1 (monolith split) | Phase 3 (Refactor) | Must plan dependency graph before moving code. |
| Test coverage protection | R1-R6 (all refactors) | Phase 3 (Refactor) | CI gate needed BEFORE starting refactor work. |
| Viper global singleton in tests | R1 + F3 | Phase 3 (Refactor) | Address when refactoring config package. |
| Moving functions + tests together | R1 | Phase 3 (Refactor) | Each function move must include its test migration. |

### Phase Ordering Rationale

1. **Phase 1: Bug Fixes** -- All 5 bugs are independent of each other and do not require structural changes. Fix them in the current monolithic structure. This delivers immediate user value and reduces the risk of the refactor (fewer bugs to carry forward).

2. **Phase 2: Features** -- F1-F6 build on the bug-fixed codebase. Env var support (F3) is the riskiest due to Viper gotchas. `--interactive` (F2) depends on B2 being fixed (i18n in prompts).

3. **Phase 3: Refactor** -- R1-R6 happen after bugs are fixed and features are in. The refactor moves stable, tested code into packages. Coverage gate ensures nothing regresses.

### Research Flags

- **Phase 1 needs NO further research** -- all bug fixes are well-understood from this analysis.
- **Phase 2 may need research** for F1 (`tempus init` wizard) -- survey library patterns with Cobra.
- **Phase 3 may need research** for R4 (conflict detection O(n log n)) -- interval tree vs sorting approach.
