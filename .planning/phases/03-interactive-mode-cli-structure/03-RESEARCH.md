# Phase 3: Interactive Mode & CLI Structure - Research

**Researched:** 2026-03-30
**Domain:** Go CLI architecture (cobra), interactive forms (charmbracelet/huh), monolith decomposition
**Confidence:** HIGH

## Summary

Phase 3 has two major workstreams: (1) implementing `tempus create --interactive` as a 7-step wizard using charmbracelet/huh, and (2) splitting the 4,223-line `main.go` monolith into `internal/cli/` with an `App` struct and `PersistentPreRunE` dependency injection. Both are well-understood patterns in the Go CLI ecosystem.

The critical blocker from STATE.md -- huh's Go 1.23 compatibility -- is resolved. **huh v1.0.0** (published 2026-02-23) requires `go 1.23.0`, which is compatible with the project's `go 1.23` directive. The installed Go toolchain is 1.24.9. huh v2.0.0+ requires Go 1.25.8 and must NOT be used. survey/v2 has 9 direct+indirect dependencies that will be cleaned up by `go mod tidy` after removal.

The monolith split is the highest-risk task due to 4,395 lines of tests in `package main` that call unexported functions directly. When functions move to `internal/cli/`, they must be exported (e.g., `runCreate` -> `RunCreate` or wrapped via factory pattern), and test files must move alongside their source files into `internal/cli/` package.

**Primary recommendation:** Use charmbracelet/huh v1.0.0 (NOT v2). Split monolith command-by-command with tests moving in lockstep. Inject `*App` via closures in command factories.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- D-01: Interactive wizard has 7 steps: Summary, Date/Time/Duration, Timezone, Alarms, Categories, Details, Confirm
- D-02: Alarms step uses profile selector + optional custom offsets (comma-separated)
- D-03: Categories step uses MultiSelect with preset list + "Other" custom entry
- D-04: huh replaces survey everywhere: new wizard, runInit, promptAlarmField, confirmQuickEvent
- D-05: All commands move to `internal/cli/<command>.go`; main.go becomes ~100 lines
- D-06: App struct with Config, Translator, Stdout, Stderr; PersistentPreRunE wiring
- D-07: internal/parsing full unification with `Parse(ParseOptions)` entry point

### Claude's Discretion
- Step progress display style (huh API choice for "Step N/7" header)
- huh version (researcher verifies -- RESOLVED: v1.0.0)
- `internal/parsing.ParseOptions` struct fields
- `confirmQuickEvent` migration approach (huh vs simple y/N)
- App injection method (closure vs cobra context)

### Deferred Ideas (OUT OF SCOPE)
- huh theming (colors, high-contrast) -- Phase 4
- --rrule in interactive wizard -- out of scope
- Attendees/priority fields in wizard -- not in 7-step scope
- internal/cli/ sub-packages per domain -- only split by command file
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UX-02 | `tempus create --interactive` guides user step-by-step with progress ("Paso 2/7") to generate event | huh v1.0.0 Form with 7 Groups, each Group.Title() shows "Step N/7 - Label"; Note field for confirm summary |
| REF-01 | Command code lives in `internal/cli/<command>.go` with App struct and PersistentPreRunE | Factory function pattern (gh CLI), App struct with closure injection, tests move with source |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charmbracelet/huh | v1.0.0 | Interactive forms (wizard, init, confirms) | Charm ecosystem standard; replaces archived survey/v2; Go 1.23 compatible |
| spf13/cobra | v1.8.0 | CLI framework (already in use) | Industry standard; PersistentPreRunE for DI |
| spf13/viper | v1.18.2 | Config management (already in use) | Already wired with cobra |

### Supporting (transitive, via huh)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| charmbracelet/bubbletea | v1.3.6 | TUI runtime (huh dependency) | Not used directly |
| charmbracelet/lipgloss | v1.1.0 | Styling (huh dependency) | Not used directly |
| charmbracelet/bubbles | v0.21.x | UI components (huh dependency) | Not used directly |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| huh v1.0.0 | huh v2.0.0+ | v2 requires Go 1.25.8 -- incompatible with project's go 1.23 |
| huh | promptui | Still works but less capable, no form grouping, less active |
| huh | raw bubbletea | Full control but 10x more code for forms |

**Installation:**
```bash
go get github.com/charmbracelet/huh@v1.0.0
```

**After survey removal:**
```bash
# Remove survey import from all files, then:
go mod tidy
```
This will clean up: `survey/v2`, `go-shellquote`, `go-colorable` (if no other dep needs it), `mgutz/ansi`, and the pinned `golang.org/x/term v0.0.0-20210927222741` (huh brings a modern version).

**Version verification:**
- huh v1.0.0: Published 2026-02-23, requires go 1.23.0. Verified via `go list -m -json`.
- huh v2.0.3: Latest but requires Go 1.25.8. NOT compatible.

## Architecture Patterns

### Recommended Project Structure (after Phase 3)
```
main.go                          # ~100 lines: main(), newRootCmd()
internal/
  cli/
    app.go                       # App struct, PersistentPreRunE setup
    create.go                    # NewCreateCmd, RunCreate, runInteractive
    create_test.go               # Tests for create command
    batch.go                     # NewBatchCmd, RunBatch, loadBatchInput
    batch_test.go
    quick.go                     # NewQuickCmd, RunQuick
    quick_test.go
    lint.go                      # NewLintCmd, lintICSFile
    lint_test.go
    config.go                    # NewConfigCmd, RunConfigSet, RunConfigList
    config_test.go
    init.go                      # NewInitCmd, RunInit (huh migration)
    init_test.go
    template.go                  # NewTemplateCmd + subcommands
    template_test.go
    timezone.go                  # NewTimezoneCmd
    timezone_test.go
    rrule.go                     # NewRRuleHelperCmd
    rrule_test.go
    version.go                   # NewVersionCmd
    locale.go                    # NewLocaleCmd
    helpers.go                   # Shared helpers: parseBoolish, slugify, etc.
    helpers_test.go
    coverage_test.go             # Coverage gap tests (from main_coverage_test.go)
    utils_test.go                # Utility tests (from main_utils_test.go)
  parsing/
    parsing.go                   # Parse(ParseOptions), all date/time functions
    parsing_test.go              # Tests migrated from main_test.go
  calendar/                      # (existing)
  config/                        # (existing)
  i18n/                          # (existing)
  prompts/                       # (existing -- keep for batch non-interactive input)
  ...
```

### Pattern 1: App Struct with Closure Injection (RECOMMENDED)

**What:** Each command factory receives `*App` and closes over it in RunE.
**When to use:** Always -- this is the gh CLI pattern at scale.

```go
// internal/cli/app.go
package cli

import (
    "io"
    "tempus/internal/config"
    "tempus/internal/i18n"
)

type App struct {
    Config     *config.Config
    Translator *i18n.Translator
    Stdout     io.Writer
    Stderr     io.Writer
}
```

```go
// internal/cli/create.go
package cli

import "github.com/spf13/cobra"

func NewCreateCmd(app *App) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create <summary>",
        Short: "Create a single ICS event",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runCreate(app, cmd, args)
        },
    }
    cmd.Flags().BoolP("interactive", "i", false, "Interactive wizard")
    // ... other flags
    return cmd
}

func runCreate(app *App, cmd *cobra.Command, args []string) error {
    interactive, _ := cmd.Flags().GetBool("interactive")
    if interactive {
        return runInteractive(app, cmd)
    }
    // ... existing logic using app.Config, app.Stdout, app.Translator
}
```

```go
// main.go (~100 lines)
package main

import (
    "os"
    "tempus/internal/cli"
    "tempus/internal/config"
    "tempus/internal/i18n"
    "github.com/spf13/cobra"
)

var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    if err := newRootCmd().Execute(); err != nil {
        os.Exit(1)
    }
}

func newRootCmd() *cobra.Command {
    app := &cli.App{
        Stdout: os.Stdout,
        Stderr: os.Stderr,
    }

    cmd := &cobra.Command{
        Use:          "tempus",
        Short:        "A multilingual ICS calendar file generator",
        SilenceUsage: true,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load()
            if err != nil {
                cfg = config.DefaultConfig()
            }
            // Override from flags
            if lang, _ := cmd.Flags().GetString("language"); lang != "" {
                cfg.Language = lang
            }
            if tz, _ := cmd.Flags().GetString("timezone"); tz != "" {
                cfg.Timezone = tz
            }
            app.Config = cfg
            app.Translator = i18n.NewTranslator(cfg.Language)
            return nil
        },
    }

    cmd.PersistentFlags().StringP("language", "l", "", "Language")
    cmd.PersistentFlags().StringP("timezone", "t", "", "Timezone")
    cmd.PersistentFlags().StringP("config", "c", "", "Config file")

    cmd.AddCommand(
        cli.NewCreateCmd(app),
        cli.NewQuickCmd(app),
        cli.NewBatchCmd(app),
        // ... all other commands
    )

    return cmd
}
```

**Why closure over cobra.Context:** Closures are type-safe, no casting needed, zero boilerplate. cobra.Context uses `context.WithValue` which requires type assertions and is error-prone. The gh CLI uses closures exclusively.

### Pattern 2: huh Multi-Step Wizard with Progress

**What:** Each wizard step is a `huh.Group` with a title showing "Step N/7 - Label".

```go
func runInteractive(app *App, cmd *cobra.Command) error {
    var (
        summary    string
        startDate  string
        startTime  string
        duration   string
        allDay     bool
        timezone   string
        alarmProf  string
        categories []string
        location   string
        description string
        confirmed  bool
    )

    timezone = app.Config.Timezone
    alarmProf = app.Config.DefaultAlarmProfile

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Event name").
                Value(&summary).
                Validate(func(s string) error {
                    if strings.TrimSpace(s) == "" {
                        return fmt.Errorf("summary is required")
                    }
                    return nil
                }),
        ).Title("Step 1/7 - Summary"),

        huh.NewGroup(
            huh.NewInput().Title("Start date (YYYY-MM-DD)").Value(&startDate),
            huh.NewInput().Title("Start time (HH:MM)").Value(&startTime),
            huh.NewInput().Title("Duration (e.g., 1h30m)").Value(&duration),
            huh.NewConfirm().Title("All-day event?").Value(&allDay),
        ).Title("Step 2/7 - Date, Time & Duration"),

        huh.NewGroup(
            huh.NewInput().
                Title("Timezone").
                Value(&timezone).
                Validate(func(s string) error {
                    return config.ValidateTimezone(s)
                }),
        ).Title("Step 3/7 - Timezone"),

        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Alarm profile").
                Options(
                    huh.NewOption("ADHD Default (-2h, -30m, -5m)", "adhd-default"),
                    huh.NewOption("ADHD Countdown (-1h, -45m, -30m, -15m, -5m)", "adhd-countdown"),
                    huh.NewOption("Medication (-30m, -5m)", "medication"),
                    huh.NewOption("Single (-15m)", "single"),
                    huh.NewOption("None", "none"),
                    huh.NewOption("Custom...", "custom"),
                ).
                Value(&alarmProf),
        ).Title("Step 4/7 - Alarms"),

        huh.NewGroup(
            huh.NewMultiSelect[string]().
                Title("Categories (optional)").
                Options(
                    huh.NewOption("Work", "work"),
                    huh.NewOption("Health", "health"),
                    huh.NewOption("Personal", "personal"),
                    huh.NewOption("Travel", "travel"),
                    huh.NewOption("School", "school"),
                    huh.NewOption("Finance", "finance"),
                ).
                Value(&categories),
        ).Title("Step 5/7 - Categories"),

        huh.NewGroup(
            huh.NewInput().Title("Location (optional)").Value(&location),
            huh.NewInput().Title("Description (optional)").Value(&description),
        ).Title("Step 6/7 - Details"),

        huh.NewGroup(
            huh.NewNote().
                Title("Event Summary").
                Description("...formatted summary..."),
            huh.NewConfirm().
                Title("Create this event?").
                Value(&confirmed),
        ).Title("Step 7/7 - Confirm"),
    ).WithAccessible(false)

    if err := form.Run(); err != nil {
        return nil // User cancelled (Ctrl+C)
    }

    if !confirmed {
        fmt.Fprintf(app.Stdout, "Event creation cancelled.\n")
        return nil
    }

    // Build and write ICS using existing createCalendarWithEvent logic
    // ...
    return nil
}
```

### Pattern 3: Parsing Unification

**What:** All 17 date/time/duration parsing functions consolidated into `internal/parsing/`.

```go
// internal/parsing/parsing.go
package parsing

type ParseOptions struct {
    StartDate string
    StartTime string
    EndDate   string
    EndTime   string
    Duration  string
    Timezone  string
    EndTZ     string
    AllDay    bool
}

type ParseResult struct {
    Start time.Time
    End   time.Time
}

func Parse(opts ParseOptions) (ParseResult, error) {
    if opts.AllDay {
        return parseAllDay(opts)
    }
    return parseTimed(opts)
}
```

The 17 functions in main.go to migrate:
1. `normalizeTimeInput` (L479)
2. `parseCreateTimes` (L486)
3. `parseAllDayTimes` (L493)
4. `parseTimedEventTimes` (L521)
5. `parseEndTime` (L548)
6. `parseDurationEnd` (L564)
7. `parseBatchTimes` (L1206)
8. `parseBatchAllDayTimes` (L1213)
9. `parseBatchTimedEventTimes` (L1237)
10. `parseBatchEndTime` (L1258)
11. `parseBatchExplicitEnd` (L1271)
12. `parseBatchDurationEnd` (L1290)
13. `normalizeDateTimeInput` (L1409)
14. `normalizeClockOnlyDateTimes` (L3792)
15. `normalizeEndTimeFromDuration` (L3815)
16. `parseDateTimeWithTZ` (L3910)
17. `parseExDateValues` (L3943)

### Anti-Patterns to Avoid
- **Moving code without tests in lockstep:** If `runCreate` moves to `internal/cli/create.go` but its tests stay in `main_test.go`, coverage drops to 0% on the moved code. Tests MUST move alongside source.
- **Exporting everything:** Only export what `main.go` needs to call (command factories, App struct). Keep `runCreate`, helpers as unexported within `internal/cli/`.
- **Using cobra.Context for DI:** Type-unsafe, verbose. Use closures instead.
- **Mixing huh and survey:** Phase 3 must remove ALL survey usage. No partial migration.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Interactive forms | Raw bufio.Scanner multi-step wizard | huh.Form with Groups | Terminal handling, cursor control, validation, accessibility |
| Step progress display | Custom "Step N/7" print statements | huh Group.Title("Step N/7 - Label") | Built-in, consistent rendering, no manual clear/redraw |
| Select/MultiSelect | Custom numbered-list chooser | huh.Select / huh.MultiSelect | Arrow key navigation, search filtering, keyboard handling |
| Config file loading race | Manual file read + parse | viper via PersistentPreRunE (once) | Already in the project; singleton load, flag override |

**Key insight:** huh handles terminal raw mode, cursor positioning, screen clearing, and keyboard events. Building this manually is ~500+ lines of fragile code.

## Common Pitfalls

### Pitfall 1: Coverage Drop During Monolith Split
**What goes wrong:** Moving 4,223 lines from `main.go` to `internal/cli/` without moving the 4,395 lines of tests causes coverage to plummet because the moved code is now in a different package.
**Why it happens:** Tests in `package main` can access unexported symbols in `main.go`. When code moves to `package cli`, those tests can't compile.
**How to avoid:** Move each command file AND its corresponding test file together. Run `go test -cover ./...` after each move. The test files (main_test.go, main_batch_test.go, etc.) must be split to match the new source files.
**Warning signs:** `go test ./...` fails to compile; coverage drops below 79%.

### Pitfall 2: PersistentPreRunE Skipped for init/version Commands
**What goes wrong:** `tempus init` and `tempus version` don't need config loaded, but PersistentPreRunE runs for ALL subcommands including these.
**Why it happens:** PersistentPreRunE is inherited by all children.
**How to avoid:** Make PersistentPreRunE tolerant of missing config (use defaults). Or check `cmd.Name()` and skip loading for `init`/`version`. The tolerant approach is cleaner.
**Warning signs:** `tempus init` fails on fresh install because no config exists yet.

### Pitfall 3: huh Forms Not Testable Without Terminal
**What goes wrong:** `form.Run()` needs a terminal (TTY). Tests in CI have no terminal.
**Why it happens:** huh uses bubbletea which reads/writes to the terminal directly.
**How to avoid:** Extract the form-building logic into a testable function that returns field values. Test the ICS generation from those values separately. For the form itself, test the struct/config, not the interactive run. Use `huh.WithAccessible(true)` in test mode if needed (renders as simple prompts).
**Warning signs:** Tests hang or panic in CI.

### Pitfall 4: survey Import Left Behind
**What goes wrong:** After migration, `survey` import remains in go.mod as indirect (or a stale import in a test file).
**Why it happens:** Incomplete grep-and-replace. survey is used in 6 locations across main.go.
**How to avoid:** After all migration: grep entire codebase for `survey`, then `go mod tidy`. Verify `go.sum` no longer references survey.
**Warning signs:** `go mod tidy` doesn't remove survey.

### Pitfall 5: Parsing Tests Assume Package-Level Access
**What goes wrong:** Tests like `TestParseAllDayTimes` call `parseAllDayTimes()` directly. After move to `internal/parsing`, function name may change or need exporting.
**Why it happens:** Parsing functions are currently unexported in package main. Moving to a new package requires either exporting them or keeping tests in the same package.
**How to avoid:** Use `package parsing` (not `package parsing_test`) for test files, so they can access unexported functions. Or export the unified `Parse()` entry point and test through it.
**Warning signs:** Tests don't compile after move.

## Code Examples

### huh Confirm (replacing survey.Confirm)

```go
// Before (survey):
confirmPrompt := &survey.Confirm{
    Message: "Does this look correct?",
    Default: true,
}
var confirmed bool
if err := survey.AskOne(confirmPrompt, &confirmed); err != nil {
    return false
}

// After (huh):
var confirmed bool
form := huh.NewForm(
    huh.NewGroup(
        huh.NewConfirm().
            Title("Does this look correct?").
            Affirmative("Yes").
            Negative("No").
            Value(&confirmed),
    ),
)
if err := form.Run(); err != nil {
    return false // Ctrl+C or error
}
```

### huh Input with Validation (replacing survey.Input)

```go
// Before (survey):
if err := survey.AskOne(
    &survey.Input{Message: "Timezone:", Default: config.DetectTimezone()},
    &timezone,
    survey.WithValidator(func(ans interface{}) error {
        return config.ValidateTimezone(ans.(string))
    }),
); err != nil {
    // ...
}

// After (huh):
huh.NewInput().
    Title("Timezone").
    Value(&timezone).  // pre-filled with default
    Validate(func(s string) error {
        return config.ValidateTimezone(s)
    })
```

### huh Select (replacing survey.Select)

```go
// Before (survey):
survey.AskOne(
    &survey.Select{Message: "Language:", Options: []string{"en", "es", "pt", "ga"}, Default: detectedLang},
    &language,
)

// After (huh):
huh.NewSelect[string]().
    Title("Language").
    Options(
        huh.NewOption("English", "en"),
        huh.NewOption("Spanish", "es"),
        huh.NewOption("Portuguese", "pt"),
        huh.NewOption("Galician", "ga"),
    ).
    Value(&language)  // pre-filled with default
```

### Test Pattern for CLI Commands (post-split)

```go
// internal/cli/create_test.go
package cli

import (
    "bytes"
    "testing"
    "tempus/internal/config"
    "tempus/internal/i18n"
)

func testApp() *App {
    return &App{
        Config:     config.DefaultConfig(),
        Translator: i18n.NewTranslator("en"),
        Stdout:     &bytes.Buffer{},
        Stderr:     &bytes.Buffer{},
    }
}

func TestRunCreate(t *testing.T) {
    app := testApp()
    cmd := NewCreateCmd(app)
    cmd.SetArgs([]string{"Meeting", "-s", "2025-01-15 10:00", "-d", "1h"})
    err := cmd.Execute()
    // ... assertions
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| survey/v2 for prompts | charmbracelet/huh | survey archived 2023, huh v1.0.0 Feb 2026 | huh has accessibility, theming, form groups |
| Flat main.go monolith | internal/cli/ with App DI | Established pattern since gh CLI (~2020) | Testable, maintainable, scalable |
| Per-command config loading | PersistentPreRunE once | Standard cobra pattern | Single load point, consistent state |

**Deprecated/outdated:**
- survey/v2: Archived 2023, pinned to old golang.org/x/term. Must be removed.
- Package-level `var stdout` pattern: Replaced by `App.Stdout` struct field.

## Open Questions

1. **"Other" category follow-up in Step 5**
   - What we know: D-03 says "Other" lets user type a custom category
   - What's unclear: huh MultiSelect doesn't natively support "if selected, show input". Need a conditional group or two-step group.
   - Recommendation: Use `WithHideFunc` on a follow-up input group that appears only when "other" is in categories slice. Or handle "other" as a second group with `huh.NewInput` that is shown conditionally.

2. **Custom alarm offsets follow-up in Step 4**
   - What we know: D-02 says if user picks "custom", show input for comma-separated offsets
   - What's unclear: Same pattern as above -- conditional follow-up within a form
   - Recommendation: Add a second group for custom input with `WithHideFunc` that checks if `alarmProf == "custom"`.

3. **internal/parsing scope: 17 functions vs unified API**
   - What we know: CONTEXT.md D-07 says "single `Parse(ParseOptions)` entry point"
   - What's unclear: Whether ALL 17 functions collapse into one, or if `Parse()` is the public API with helpers unexported in the package
   - Recommendation: `Parse(ParseOptions)` is the exported entry point. Keep individual functions as unexported helpers within the package. This preserves existing logic while unifying the interface.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None (stdlib, no config needed) |
| Quick run command | `go test ./... -count=1` |
| Full suite command | `go test ./... -count=1 -coverprofile=coverage.out && go tool cover -func=coverage.out \| tail -1` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UX-02 | Interactive wizard creates valid ICS from form values | unit | `go test ./internal/cli/ -run TestRunInteractive -count=1` | Wave 0 |
| UX-02 | Step progress visible ("Step N/7") in group titles | unit | `go test ./internal/cli/ -run TestInteractiveFormStructure -count=1` | Wave 0 |
| UX-02 | Wizard respects config defaults (timezone, alarm profile) | unit | `go test ./internal/cli/ -run TestInteractiveDefaults -count=1` | Wave 0 |
| UX-02 | Cancel (Ctrl+C) exits cleanly without writing file | unit | `go test ./internal/cli/ -run TestInteractiveCancel -count=1` | Wave 0 |
| REF-01 | All existing commands work identically after split | integration | `go test ./... -count=1` | Existing (move) |
| REF-01 | main.go is ~100 lines | manual | `wc -l main.go` | N/A |
| REF-01 | Coverage >= 79% | metric | `go test ./... -coverprofile=c.out && go tool cover -func=c.out \| tail -1` | Existing |
| REF-01 | App struct injects config/translator | unit | `go test ./internal/cli/ -run TestAppWiring -count=1` | Wave 0 |
| REF-01 | survey/v2 fully removed from go.mod | manual | `grep survey go.mod` | N/A |

### Sampling Rate
- **Per task commit:** `go test ./... -count=1`
- **Per wave merge:** `go test ./... -count=1 -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1`
- **Phase gate:** Full suite green + coverage >= 79% before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/cli/create_test.go` -- covers UX-02 interactive wizard tests
- [ ] `internal/cli/app_test.go` -- covers REF-01 App struct and PersistentPreRunE wiring
- [ ] `internal/parsing/parsing_test.go` -- covers migrated parsing tests
- [ ] Test helper `testApp()` function in `internal/cli/` for constructing test App instances
- [ ] All existing test files in `main_*_test.go` must be split and moved to `internal/cli/`

## Discretion Recommendations

### Step Progress Display: Use Group.Title()
**Recommendation:** Static `Group.Title("Step N/7 - Label")` on each group. No need for `TitleFunc` since step numbers are fixed. Clean, simple, guaranteed to render.

### confirmQuickEvent: Migrate to huh
**Recommendation:** Use `huh.NewConfirm()` wrapped in a `huh.NewForm()`. This keeps a single prompt library. The current survey.Confirm is 6 lines; huh equivalent is ~8 lines. Worth it for consistency and survey removal.

### App Injection: Closures (not cobra.Context)
**Recommendation:** Pass `*App` to each command factory function. RunE closures capture it. Type-safe, zero boilerplate, matches gh CLI pattern exactly. See Architecture Pattern 1 above.

### ParseOptions Fields
**Recommendation:** Based on the 17 existing functions, the unified struct needs:
```go
type ParseOptions struct {
    StartDate  string  // "2025-01-15" or "2025/01/15"
    StartTime  string  // "10:00" or "10:00:00"
    EndDate    string  // optional
    EndTime    string  // optional
    Duration   string  // "1h30m" or "90m"
    Timezone   string  // IANA timezone
    EndTZ      string  // optional, for cross-timezone events
    AllDay     bool
    Summary    string  // for error messages (batch uses this)
}
```

## Sources

### Primary (HIGH confidence)
- `go list -m -json github.com/charmbracelet/huh@v1.0.0` -- GoVersion: 1.23.0, published 2026-02-23
- `go list -m -versions github.com/charmbracelet/huh` -- v1.0.0 is latest v1 line
- `go list -m -versions github.com/charmbracelet/huh/v2` -- v2.0.0-v2.0.3 exist, require Go 1.25.8
- [huh v1.0.0 go.mod](https://raw.githubusercontent.com/charmbracelet/huh/v1.0.0/go.mod) -- verified go 1.23.0 requirement
- [huh pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/huh) -- API reference: Form, Group, Input, Select, MultiSelect, Confirm, Note
- [huh releases](https://github.com/charmbracelet/huh/releases) -- release history and changelogs
- Codebase analysis: main.go (4,223 lines), 6 survey usages at L102/106/113/128/139/153/280/285, 17 parsing functions, 4,395 lines of tests

### Secondary (MEDIUM confidence)
- [huh README](https://github.com/charmbracelet/huh) -- API examples, Group.Title(), WithAccessible(), theming

### Tertiary (LOW confidence)
- None -- all findings verified against code or official sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- huh v1.0.0 compatibility verified against go.mod and Go toolchain
- Architecture: HIGH -- patterns established in gh CLI, cobra docs, and existing Tempus codebase
- Pitfalls: HIGH -- derived from direct codebase analysis (test count, function signatures, survey locations)
- Parsing unification: MEDIUM -- function list verified but exact ParseOptions fields are a recommendation

**Research date:** 2026-03-30
**Valid until:** 2026-04-30 (stable ecosystem, huh v1 is likely final v1 release)
