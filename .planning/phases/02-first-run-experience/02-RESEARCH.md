# Phase 2: First-Run Experience - Research

**Researched:** 2026-03-29
**Domain:** Go CLI first-run wizard, Viper env var integration, config validation, batch templates
**Confidence:** HIGH

## Summary

Phase 2 adds `tempus init` (interactive wizard using survey/v2), Viper env var support (CONF-01), config validation for timezone and output_dir (CONF-02/CONF-03), and 3 new batch templates with `--format` flag for CSV/YAML output (TMPL-01/02/03).

The existing codebase provides strong foundations: `config.Load()` already uses Viper with defaults, `ValidateTimezone()` and `ValidateLanguage()` already exist, `getBatchTemplateContent()` has a clean switch-case pattern for adding templates, and survey/v2 is already imported with a working `survey.Confirm` call at L171 of main.go. The main work is wiring these together and adding the missing Viper env var calls.

**Primary recommendation:** Add `SetEnvPrefix`/`AutomaticEnv` to `config.Load()` first (enables all env var tests), then build `tempus init` wizard, then add config validation to `config set`, then add templates with `--format` flag.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- D-01: `tempus init` auto-detects timezone from OS and language from LANG env var, presents for confirmation. Wizard always in English.
- D-02: Existing config file prompts "Overwrite? [y/N]" before running wizard.
- D-03: Completion output shows summary table + next-step hint.
- D-04: Interactive only -- no `--yes` flag.
- D-05: Config validation errors follow Phase 1 error style (fatal + actionable guidance).
- D-06: `tempus config set` shows before+after on success.
- D-07: Viper `SetEnvPrefix("TEMPUS")` + `AutomaticEnv()` + `SetEnvKeyReplacer` for 5 env vars.
- D-08: school-event template with specific CSV columns.
- D-09: recruiter-meeting template with specific CSV columns.
- D-10: travel-day template with specific CSV columns.
- D-11: `--format` flag on `tempus batch template` (csv default | yaml).

### Claude's Discretion
- Timezone auto-detection implementation (timedatectl / /etc/localtime symlink / go-tzlocal)
- Survey prompt style (select vs input vs confirm)
- Config file path detection (reuse existing logic)
- YAML template formatting details

### Deferred Ideas (OUT OF SCOPE)
- Non-interactive `tempus init --yes` flag
- `charmbracelet/huh` for the init wizard (Phase 3)
- Config validation for fields beyond timezone and output_dir
- `tempus init --update` to selectively update one field
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CONF-01 | Env vars via Viper AutomaticEnv | Viper pattern documented; pitfalls identified (SetEnvKeyReplacer bug in CONTEXT, test isolation) |
| CONF-02 | `config set timezone` validates IANA | `ValidateTimezone()` already exists in config.go L243; wire into `runConfigSet` |
| CONF-03 | `config set output_dir` validates exists+writable | New `ValidateOutputDir()` needed; stdlib only (os.Stat + write test) |
| UX-01 | `tempus init` wizard | survey/v2 pattern from L171; 4 steps: timezone, language, output_dir, alarm profile |
| TMPL-01 | school-event template | CSV columns defined in D-08; add to `getBatchTemplateContent()` switch |
| TMPL-02 | recruiter-meeting template | CSV columns defined in D-09; add to switch |
| TMPL-03 | travel-day template | CSV columns defined in D-10; add to switch |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

- No breaking the existing command API (flags and outputs preserved)
- Test coverage must not drop below 79% (currently 79.7%)
- Keep `go.mod` clean -- survey/v2 is already a dependency, no new deps needed
- All ICS output must remain RFC 5545 valid
- Offline tool -- no external API dependencies

## Standard Stack

### Core (already in go.mod)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| spf13/viper | v1.18.2 | Config + env var management | Already used; adding AutomaticEnv is 3 lines |
| spf13/cobra | v1.8.0 | CLI framework | Already used; new commands follow existing pattern |
| AlecAivazis/survey/v2 | v2.3.7 | Interactive prompts for init wizard | Already a dependency; migration to huh is Phase 3 |

### Supporting (stdlib only)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| time (stdlib) | go1.23 | `time.LoadLocation()` for timezone validation | CONF-02 validation |
| os (stdlib) | go1.23 | `os.Stat()` + dir writability check | CONF-03 validation |
| gopkg.in/yaml.v3 | v3.0.1 | YAML template output | Already in go.mod for batch YAML |

### No New Dependencies Needed
The phase can be implemented entirely with existing dependencies plus stdlib. No new entries in `go.mod`.

## Architecture Patterns

### Where New Code Lives

```
main.go
  newRootCmd() L55         -- add newInitCmd() to cmd.AddCommand()
  newInitCmd()             -- NEW: factory function for tempus init
  runInit()                -- NEW: wizard logic using survey/v2
  newConfigCmd() L2352     -- add validation to runConfigSet
  runConfigSet() L2380     -- add ValidateTimezone/ValidateOutputDir calls
  newBatchTemplateCmd() L1810 -- add --format flag
  runBatchTemplate() L1839    -- handle --format flag
  getBatchTemplateContent() L1862 -- add 3 new cases + format param

internal/config/config.go
  Load() L74               -- add SetEnvPrefix, AutomaticEnv, SetEnvKeyReplacer
  ValidateTimezone() L243  -- ALREADY EXISTS
  ValidateLanguage() L254  -- ALREADY EXISTS
  ValidateOutputDir()      -- NEW: os.Stat + writability check
  DetectTimezone()         -- NEW: /etc/localtime symlink resolution
  DetectLanguage()         -- NEW: LANG env var parsing
```

### Pattern: survey/v2 Usage (from existing code L171)

Existing pattern in main.go:
```go
confirmPrompt := &survey.Confirm{
    Message: "Does this look correct?",
    Default: true,
}
var confirmed bool
if err := survey.AskOne(confirmPrompt, &confirmed); err != nil {
    return false
}
```

For `tempus init`, use `survey.AskOne` for each step (not `survey.Ask` with a question list) to keep the wizard sequential and allow early exit. Use:
- `survey.Input` for timezone (pre-filled with detected value)
- `survey.Select` for language (from supported list: en, es, pt, ga)
- `survey.Input` for output_dir (pre-filled with ".")
- `survey.Select` for alarm profile (from available profile names)
- `survey.Confirm` for overwrite check (existing config scenario)

### Pattern: Viper Env Var Integration

Add to `config.Load()` BEFORE `viper.ReadInConfig()`:
```go
viper.SetEnvPrefix("TEMPUS")
viper.AutomaticEnv()
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
```

### Pattern: Config Validation in runConfigSet

```go
func runConfigSet(_ *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return err
    }

    key, value := args[0], args[1]
    oldValue, _ := cfg.Get(key)

    switch key {
    case "timezone":
        if err := config.ValidateTimezone(value); err != nil {
            return fmt.Errorf("Invalid timezone: '%s'. Use 'tempus timezone list --search <name>' to find a valid IANA identifier.", value)
        }
    case "output_dir":
        if err := config.ValidateOutputDir(value); err != nil {
            return err
        }
    }

    if err := cfg.Set(key, value); err != nil {
        return err
    }
    printOK("%s: %s -> %s\n", key, oldValue, value)
    return nil
}
```

### Pattern: Template with --format Flag

Extend `getBatchTemplateContent` to accept a format parameter:
```go
func getBatchTemplateContent(templateType, format string) (string, error) {
    switch templateType {
    case "school-event":
        if format == "yaml" {
            return getSchoolEventTemplateYAML(), nil
        }
        return getSchoolEventTemplateCSV(), nil
    // ... existing cases unchanged (they only support their native format)
    }
}
```

### Anti-Patterns to Avoid
- **Do not use `survey.Ask` with multiple questions at once** -- it removes ability to validate/transform between steps. Use `survey.AskOne` per step.
- **Do not call `viper.Reset()` in production code** -- only in tests. It clears all state including defaults.
- **Do not create a separate `internal/init/` package** -- the wizard is a single command; keep it in main.go like all other commands until Phase 3 refactor.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Timezone validation | Custom IANA database | `time.LoadLocation()` | Stdlib uses system tzdata; always current |
| Language validation | Custom language list | `i18n.IsSupportedLanguage()` | Already exists, uses actual translation catalog |
| Config file path | Custom XDG logic | `config.ConfigDir()` | Already exists, handles XDG_CONFIG_HOME + fallback |
| Env var binding | Manual os.Getenv calls | Viper AutomaticEnv | Handles precedence automatically |
| YAML marshaling | Manual string building | `yaml.v3` Marshal | Already a dependency, handles escaping |

## Common Pitfalls

### Pitfall 1: SetEnvKeyReplacer Direction in CONTEXT.md D-07

**What goes wrong:** CONTEXT.md D-07 specifies `strings.NewReplacer("_", "-")` which replaces underscores with hyphens in config keys. This would make `output_dir` look up `TEMPUS_OUTPUT-DIR` instead of `TEMPUS_OUTPUT_DIR`.
**Why it happens:** The replacer direction was inverted in the discussion. STACK.md research correctly identifies `strings.NewReplacer(".", "_", "-", "_")` (dots/hyphens to underscores).
**How to avoid:** Use `strings.NewReplacer(".", "_", "-", "_")` as documented in STACK.md and PITFALLS.md. Since the current config keys only use underscores (no dots or hyphens), this replacer is a no-op safety net. The keys `output_dir`, `date_format`, `time_format` already contain underscores which map directly to `TEMPUS_OUTPUT_DIR` etc. without needing replacement.
**Warning signs:** `TEMPUS_OUTPUT_DIR` env var is ignored in tests.
**Confidence:** HIGH -- verified against PITFALLS.md analysis and Viper documentation.

### Pitfall 2: Viper Global Singleton in Tests

**What goes wrong:** `viper.Reset()` in one test affects another. Adding `AutomaticEnv()` means `t.Setenv("TEMPUS_TIMEZONE", ...)` leaks between tests.
**Why it happens:** Viper uses a global singleton by default. Config tests already call `viper.Reset()` before each test (verified in config_test.go).
**How to avoid:** Continue using `viper.Reset()` at the start of each test function. Use `t.Setenv()` for env var tests (auto-cleaned up by Go test framework). Never use `os.Setenv()` directly in tests.
**Warning signs:** Flaky tests that pass individually but fail in batch.

### Pitfall 3: Timezone Auto-Detection Portability

**What goes wrong:** `/etc/localtime` symlink resolution works on Linux/macOS but not on all systems. Some systems use a copy instead of a symlink.
**Why it happens:** Different distributions handle timezone differently.
**How to avoid:** Use a fallback chain: (1) Read `/etc/localtime` symlink target, extract IANA name from path; (2) Read `/etc/timezone` file content; (3) Fall back to "UTC" and let user correct.
**Warning signs:** Empty or invalid timezone detected on CI or containers.

### Pitfall 4: survey/v2 Terminal Issues

**What goes wrong:** survey/v2 is archived (2023) and has known terminal rendering issues with some modern terminals.
**Why it happens:** No maintenance since 2023. The `golang.org/x/term` dependency is pinned to an old version.
**How to avoid:** Keep survey usage minimal (only this wizard + the existing L171 confirm). Phase 3 migrates to charmbracelet/huh. Do not add complex survey patterns (multi-select, editor, etc.).
**Warning signs:** Garbled terminal output after wizard completes.

### Pitfall 5: Config Set Before+After When Env Var Overrides

**What goes wrong:** If `TEMPUS_TIMEZONE` is set as env var, `config set timezone X` changes the config file but the effective value is still the env var. The before+after output would show the config file change, but the user might be confused.
**Why it happens:** Viper precedence: env vars override config file.
**How to avoid:** Document this in the config set output or show a warning. Out of scope for this phase per deferred ideas, but the planner should be aware.
**Warning signs:** User sets config, but `config list` shows the env var value.

### Pitfall 6: Coverage Gate at 79%

**What goes wrong:** New code without tests drops coverage below 79.7%.
**Why it happens:** Adding template functions and wizard logic adds lines without corresponding test coverage.
**How to avoid:** Every new function needs a test. Template content functions are trivially testable (call and check non-empty). Wizard logic needs at least a unit test for the auto-detection functions (DetectTimezone, DetectLanguage). The survey interaction itself is hard to test (needs terminal mock) -- focus tests on the logic, not the UI.
**Warning signs:** `go test -cover ./...` shows < 79%.

## Code Examples

### Timezone Auto-Detection (stdlib only)

```go
func DetectTimezone() string {
    // Method 1: /etc/localtime symlink
    if target, err := os.Readlink("/etc/localtime"); err == nil {
        if idx := strings.Index(target, "zoneinfo/"); idx != -1 {
            tz := target[idx+len("zoneinfo/"):]
            if _, err := time.LoadLocation(tz); err == nil {
                return tz
            }
        }
    }

    // Method 2: /etc/timezone file
    if data, err := os.ReadFile("/etc/timezone"); err == nil {
        tz := strings.TrimSpace(string(data))
        if _, err := time.LoadLocation(tz); err == nil {
            return tz
        }
    }

    return "UTC"
}
```

**Confidence:** HIGH -- both methods are standard Linux approaches. Verified `/etc/localtime` symlink exists on this system (`/usr/share/zoneinfo/Europe/Dublin`). `/etc/timezone` also exists and contains `Europe/Dublin`.

### Language Detection from LANG

```go
func DetectLanguage() string {
    lang := os.Getenv("LANG")
    if lang == "" {
        return "en"
    }
    // LANG format: "es_ES.UTF-8" or "en_US" or "C.UTF-8"
    lang = strings.Split(lang, ".")[0] // strip encoding
    lang = strings.Split(lang, "_")[0] // strip country
    lang = strings.ToLower(lang)

    if lang == "c" || lang == "posix" || lang == "" {
        return "en"
    }
    if i18n.IsSupportedLanguage(lang) {
        return lang
    }
    return "en"
}
```

### ValidateOutputDir

```go
func ValidateOutputDir(dir string) error {
    dir = strings.TrimSpace(dir)
    if dir == "" {
        return fmt.Errorf("output directory cannot be empty")
    }
    info, err := os.Stat(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("Directory '%s' does not exist or is not writable.", dir)
        }
        return fmt.Errorf("cannot access directory '%s': %w", dir, err)
    }
    if !info.IsDir() {
        return fmt.Errorf("'%s' is not a directory", dir)
    }
    // Test writability
    f, err := os.CreateTemp(dir, ".tempus-test-*")
    if err != nil {
        return fmt.Errorf("Directory '%s' does not exist or is not writable.", dir)
    }
    f.Close()
    os.Remove(f.Name())
    return nil
}
```

### Init Wizard Structure

```go
func newInitCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "init",
        Short: "Configure Tempus interactively",
        RunE:  runInit,
    }
}

func runInit(_ *cobra.Command, _ []string) error {
    configDir, _ := config.ConfigDir()
    configFile := filepath.Join(configDir, "config.yaml")

    if _, err := os.Stat(configFile); err == nil {
        var overwrite bool
        survey.AskOne(&survey.Confirm{
            Message: fmt.Sprintf("Config already exists at %s. Overwrite?", configFile),
            Default: false,
        }, &overwrite)
        if !overwrite {
            fmt.Fprintf(stdout, "Config unchanged. Use 'tempus config set <key> <value>' for individual changes.\n")
            return nil
        }
    }

    // Step 1: Timezone
    detectedTZ := config.DetectTimezone()
    var timezone string
    survey.AskOne(&survey.Input{
        Message: "Timezone:",
        Default: detectedTZ,
    }, &timezone, survey.WithValidator(func(ans interface{}) error {
        return config.ValidateTimezone(ans.(string))
    }))

    // Step 2: Language
    detectedLang := config.DetectLanguage()
    var language string
    survey.AskOne(&survey.Select{
        Message: "Language:",
        Options: []string{"en", "es", "pt", "ga"},
        Default: detectedLang,
    }, &language)

    // Step 3: Output directory
    var outputDir string
    survey.AskOne(&survey.Input{
        Message: "Output directory:",
        Default: ".",
    }, &outputDir, survey.WithValidator(func(ans interface{}) error {
        return config.ValidateOutputDir(ans.(string))
    }))

    // Step 4: Alarm profile
    var alarmProfile string
    survey.AskOne(&survey.Select{
        Message: "Default alarm profile:",
        Options: []string{"adhd-default", "adhd-countdown", "medication", "single", "none"},
        Default: "adhd-default",
    }, &alarmProfile)

    // Save config
    // ... create Config, save, print summary
}
```

### Template --format Flag Integration

```go
func newBatchTemplateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "template [type]",
        // ... existing Short/Long ...
        Args: cobra.ExactArgs(1),
        RunE: runBatchTemplate,
    }

    cmd.Flags().StringP("output", "o", "", "Output file path (required)")
    _ = cmd.MarkFlagRequired("output")
    cmd.Flags().StringP("format", "f", "csv", "Template format: csv or yaml")

    return cmd
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| survey/v2 for prompts | charmbracelet/huh | 2023 (survey archived) | Phase 2 keeps survey; Phase 3 migrates |
| Manual env var reading | Viper AutomaticEnv | Stable since Viper v1.0 | 3-line addition to Load() |
| No config validation | Validate on set | This phase | Prevents invalid config from being saved |

## Open Questions

1. **SetEnvKeyReplacer argument mismatch**
   - What we know: CONTEXT.md D-07 says `NewReplacer("_", "-")` but STACK.md/PITFALLS.md correctly identify `NewReplacer(".", "_", "-", "_")`. The D-07 version would break `TEMPUS_OUTPUT_DIR` lookup.
   - Recommendation: Use the STACK.md version. The intent of D-07 is clearly to enable env var support; the specific replacer argument is a typo. Flag this to the user during planning.

2. **Alarm profile storage in config**
   - What we know: The wizard collects a "default alarm profile" but the Config struct has no `DefaultAlarmProfile` field -- it has `AlarmProfiles` (the map of all profiles).
   - What's unclear: Where should the selected default profile be stored? A new config key `default_alarm_profile` would need to be added to the Config struct, Set(), Get(), and List().
   - Recommendation: Add a `default_alarm_profile` string field to Config. It stores the profile name, not the triggers. Commands that need alarms look up `config.AlarmProfiles[config.DefaultAlarmProfile]`.

3. **survey.AskOne error handling in wizard**
   - What we know: The existing L171 usage swallows the error (returns false). For the wizard, Ctrl+C should abort cleanly.
   - Recommendation: Check `survey.AskOne` error -- if it's terminal interrupt, return `nil` (clean exit). Otherwise propagate the error.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None (Go convention) |
| Quick run command | `go test ./internal/config/ -v -run TestName` |
| Full suite command | `go test -cover ./...` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CONF-01 | Env vars override config file values | unit | `go test ./internal/config/ -run TestEnvVar -v` | Wave 0 |
| CONF-02 | `config set timezone` validates IANA | unit | `go test ./internal/config/ -run TestValidateTimezone -v` | Exists (config_test.go L256) |
| CONF-03 | `config set output_dir` validates dir | unit | `go test ./internal/config/ -run TestValidateOutputDir -v` | Wave 0 |
| UX-01 | tempus init auto-detection functions | unit | `go test ./internal/config/ -run TestDetect -v` | Wave 0 |
| TMPL-01 | school-event template content | unit | `go test . -run TestSchoolEventTemplate -v` | Wave 0 |
| TMPL-02 | recruiter-meeting template content | unit | `go test . -run TestRecruiterMeetingTemplate -v` | Wave 0 |
| TMPL-03 | travel-day template content | unit | `go test . -run TestTravelDayTemplate -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test -cover ./...` (must stay >= 79%)
- **Per wave merge:** Full suite + coverage check
- **Phase gate:** Full suite green before verification

### Wave 0 Gaps
- [ ] `internal/config/config_test.go` -- add TestEnvVarOverride, TestValidateOutputDir, TestDetectTimezone, TestDetectLanguage
- [ ] `main_test.go` -- add TestSchoolEventTemplate, TestRecruiterMeetingTemplate, TestTravelDayTemplate, TestBatchTemplateFormat

## Sources

### Primary (HIGH confidence)
- Codebase inspection: `internal/config/config.go` -- existing Load(), Set(), ValidateTimezone(), ValidateLanguage()
- Codebase inspection: `main.go` L171 -- survey.Confirm pattern, L1810-1880 -- batch template pattern, L2352-2390 -- config set pattern
- `.planning/research/PITFALLS.md` L300-378 -- Viper AutomaticEnv gotchas (7 pitfalls documented)
- `.planning/research/STACK.md` L245-265 -- Viper env var correct pattern

### Secondary (MEDIUM confidence)
- System inspection: `/etc/localtime` -> `/usr/share/zoneinfo/Europe/Dublin`, `/etc/timezone` = `Europe/Dublin` -- timezone detection verified on target system

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all libraries already in go.mod, no new dependencies
- Architecture: HIGH -- all patterns verified against existing codebase
- Pitfalls: HIGH -- documented in prior research, verified against code
- Templates: HIGH -- CSV columns specified in CONTEXT.md decisions; pattern clear from existing templates

**Research date:** 2026-03-29
**Valid until:** 2026-04-28 (stable -- no fast-moving dependencies)
