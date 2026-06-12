# Stack Research

**Project:** Tempus (Go CLI for ICS generation)
**Researched:** 2026-03-29
**Note:** Web search tools were unavailable. Findings based on Go ecosystem knowledge (stable patterns) plus codebase analysis. Confidence adjusted accordingly.

---

## Go CLI Project Structure (2025)

The standard Go project layout for a CLI tool of Tempus's size (~4K LOC, 10 commands) follows this pattern, established by projects like `gh`, `kubectl`, `hugo`, and codified in `golang-standards/project-layout`:

```
tempus/
  cmd/
    tempus/
      main.go          # ~20 lines: wiring only
  internal/
    cli/               # Command definitions (Cobra wiring)
      root.go          # Root command + persistent flags
      create.go        # Each command = one file
      batch.go
      lint.go
      config.go
      template.go
      timezone.go
      rrule.go
      quick.go
      locale.go
      version.go
    event/             # Core domain logic (ICS generation)
      event.go
      builder.go
      conflict.go
      prep_time.go
    parsing/           # Date/time parsing, input normalization
      dates.go
      natural.go       # Natural language (olebedev/when wrapper)
    nd/                # Neurodivergent-specific features
      spellcheck.go
      emoji.go
      smart_defaults.go
    output/            # Output formatting, file writing
      writer.go
      printer.go       # printOK/printErr abstraction
    calendar/          # (existing) ICS calendar assembly
    config/            # (existing) Viper config
    constants/         # (existing)
    i18n/              # (existing) Translation system
    normalizer/        # (existing) Input normalization
    prompts/           # (existing) User input
    templates/         # (existing) Event templates
    timezone/          # (existing) TZ handling
    utils/             # (existing) String utilities
  locales/             # Embedded translation files
  docs/
  examples/
```

**Key principles for Tempus specifically:**

1. **Move `main.go` to `cmd/tempus/main.go`** -- standard Go convention for a single-binary project. The current root `main.go` becomes just:
   ```go
   package main

   import "tempus/internal/cli"

   func main() {
       cli.Execute()
   }
   ```

2. **One file per command in `internal/cli/`** -- each `newXxxCmd()` function moves to its own file. The `RunE` function stays in the same file. This is exactly how `gh` CLI does it (see `pkg/cmd/` in cli/cli).

3. **Business logic must NOT live in command files** -- command files parse flags and call domain packages. The 3,906-line monolith has parsing, validation, ICS assembly, conflict detection, spell checking all interleaved with Cobra wiring. These must separate.

4. **`internal/` not `pkg/`** -- Tempus is an end-user CLI, not a library. Everything goes in `internal/` to prevent accidental API promises. The project already follows this correctly.

**What NOT to do:**
- Do NOT use `cobra-cli init` scaffolding -- it generates a `cmd/` package (not `cmd/appname/`) with a flat structure that does not scale well.
- Do NOT create a `pkg/` directory -- this is for libraries, not CLIs.
- Do NOT split too granularly -- a package with one 30-line file is over-engineering. Group by domain, not by file count.

**Migration path from current monolith:**

The project doc (R1) already suggests the right target: `internal/commands/`, `internal/parsing/`, `internal/nd/`, `internal/output/`. I'd rename `commands` to `cli` to match ecosystem convention, but the decomposition is sound. The refactor should happen incrementally:

1. Extract `newRootCmd()` + command registration to `internal/cli/root.go`
2. Move each `newXxxCmd()` + its `runXxx` to `internal/cli/xxx.go`
3. Extract shared business logic (date parsing, conflict detection, etc.) into domain packages
4. Leave `main.go` (or `cmd/tempus/main.go`) as a thin entry point

---

## Cobra Command Packaging Patterns

### Pattern: Command Factory Functions

The mature pattern (used by `gh`, `kubectl`, `eksctl`) is command factory functions that accept dependencies:

```go
// internal/cli/create.go
package cli

import (
    "tempus/internal/config"
    "tempus/internal/event"
    "github.com/spf13/cobra"
)

func newCreateCmd(cfg *config.Config) *cobra.Command {
    opts := &createOptions{}
    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new ICS event",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runCreate(opts, cfg)
        },
    }
    cmd.Flags().StringVarP(&opts.title, "title", "T", "", "Event title")
    // ... more flags
    return cmd
}

type createOptions struct {
    title    string
    date     string
    // ... all flags as struct fields
}

func runCreate(opts *createOptions, cfg *config.Config) error {
    // Business logic here, calling domain packages
}
```

**Why this pattern:**
- Testable: you can call `runCreate()` directly in tests with fake options
- No globals: config is injected, not imported from a singleton
- Flag parsing is co-located with the command definition
- `gh` CLI uses this exact pattern at scale (~100 commands)

### Pattern: Root Command Assembly

```go
// internal/cli/root.go
package cli

import (
    "os"
    "tempus/internal/config"
    "github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
    cfg, _ := config.Load()

    cmd := &cobra.Command{
        Use:          "tempus",
        Short:        "A multilingual ICS calendar file generator",
        SilenceUsage: true,
    }

    cmd.PersistentFlags().StringP("language", "l", "", "Language")
    cmd.PersistentFlags().StringP("timezone", "t", "", "Timezone")
    cmd.PersistentFlags().StringP("config", "c", "", "Config file")

    cmd.AddCommand(
        newCreateCmd(cfg),
        newQuickCmd(cfg),
        newBatchCmd(cfg),
        newLintCmd(),
        newConfigCmd(cfg),
        newVersionCmd(),
        newTemplateCmd(cfg),
        newLocaleCmd(),
        newTimezoneCmd(),
        newRRuleHelperCmd(),
    )

    return cmd
}

func Execute() {
    if err := NewRootCmd().Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Pattern: Subcommand Groups (not needed yet)

For CLIs with 20+ commands, group them: `tempus config get`, `tempus config set`. Tempus already does this with `config`. No need to add more nesting at this scale.

### Migration without breaking public API

Critical constraint: all existing `tempus create --title "foo"` invocations must keep working.

1. **Do not change `Use`, `Short`, flag names, or flag shorthand** -- these are the public API
2. **Move functions, do not rename commands** -- the refactor is about file organization, not UX changes
3. **Test the command tree after each move** -- run `go build && ./tempus --help` and verify every command appears
4. **Keep `ldflags` injection working** -- the `version`, `commit`, `date` vars must remain accessible. If moving main to `cmd/tempus/`, update the Makefile `-X` paths accordingly (they currently target `main.*`, which still works if `cmd/tempus/main.go` is `package main`)

---

## i18n Patterns in Go CLIs

### Current state in Tempus

Tempus has a solid i18n foundation: embedded JSON locales in `internal/i18n/`, `Translator` struct with English fallback, key-based lookups. This is already better than most Go CLIs.

### Standard patterns in the ecosystem

**1. `go-i18n` (nicksnyder/go-i18n) -- most popular Go i18n library**
- Message catalogs in TOML/JSON/YAML
- Pluralization support (CLDR rules)
- `go generate` for extracting translation keys
- Used by Hugo, Mattermost
- Tempus does NOT need this -- the current homegrown system is sufficient for a CLI with ~50-80 message keys

**2. `golang.org/x/text` -- stdlib-adjacent**
- Low-level: message catalogs, plural forms, number/date formatting
- Good for complex localization (currency, dates, collation)
- Overkill for Tempus's use case (translating CLI messages)

**3. Key-based lookup (current Tempus approach)**
- Simple, explicit, easy to audit
- No magic, no code generation
- Limitation: no pluralization, no interpolation beyond `fmt.Sprintf`
- Perfectly adequate for a CLI tool

### Recommendation for Tempus

**Keep the current i18n system.** It works, it's simple, and the project constraint says "no unnecessary dependencies." The improvements needed are:

1. **Fix the hardcoded Spanish in `promptAlarmField()`** -- this is a bug (B2), not an architecture problem. Pass the Translator to prompt functions.
2. **Ensure all user-facing strings go through the Translator** -- audit `main.go` for `fmt.Println` with hardcoded English/Spanish strings.
3. **Add interpolation support** -- for messages like "Event '%s' created in %s", use `Translator.Format(key, args...)` that wraps `fmt.Sprintf(t.Get(key), args...)`. This may already exist.

Do NOT migrate to `go-i18n` -- it would add complexity and a dependency for marginal benefit at this scale.

---

## Viper + Env Vars Pattern

### The problem

Tempus documents env vars (`TEMPUS_TIMEZONE`, `TEMPUS_LANGUAGE`, `TEMPUS_OUTPUT_DIR`) but never calls `viper.AutomaticEnv()` or `viper.SetEnvPrefix()`. Confirmed: grep for `AutomaticEnv|SetEnvPrefix|BindEnv` returns zero matches.

### The correct pattern

Add to `config.Load()`, before `viper.ReadInConfig()`:

```go
viper.SetEnvPrefix("TEMPUS")
viper.AutomaticEnv()
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
```

**How it works:**
- `SetEnvPrefix("TEMPUS")` makes Viper look for `TEMPUS_*` env vars
- `AutomaticEnv()` enables automatic env var binding
- `SetEnvKeyReplacer` maps config keys like `output_dir` to `TEMPUS_OUTPUT_DIR`
- Env vars override config file values but NOT command-line flags

**Precedence order (Viper's built-in):**
1. Explicit `Set()` calls (highest)
2. Command-line flags (via `BindPFlag`)
3. Environment variables
4. Config file
5. Key/value store
6. Default values (lowest)

**Important: BindPFlag for flag override**

For the persistent flags (`--language`, `--timezone`) to properly override env vars, bind them:

```go
// In root command setup
viper.BindPFlag("language", cmd.PersistentFlags().Lookup("language"))
viper.BindPFlag("timezone", cmd.PersistentFlags().Lookup("timezone"))
```

This ensures `--timezone America/New_York` beats `TEMPUS_TIMEZONE=UTC`.

**Testing env vars:**

```go
func TestEnvVarOverridesConfig(t *testing.T) {
    t.Setenv("TEMPUS_TIMEZONE", "America/New_York")
    cfg, err := config.Load()
    require.NoError(t, err)
    assert.Equal(t, "America/New_York", cfg.Timezone)
}
```

**Nested config keys:**

For `alarm_profiles`, Viper maps `TEMPUS_ALARM_PROFILES` but cannot represent a `map[string][]string` in a single env var. This is fine -- alarm profiles should only be configurable via the config file, not env vars. Document this limitation.

### What to expose as env vars

Only simple string values that make sense to override per-session:

| Env Var | Config Key | Type |
|---------|-----------|------|
| `TEMPUS_TIMEZONE` | `timezone` | string |
| `TEMPUS_LANGUAGE` | `language` | string |
| `TEMPUS_OUTPUT_DIR` | `output_dir` | string |
| `TEMPUS_DATE_FORMAT` | `date_format` | string |
| `TEMPUS_TIME_FORMAT` | `time_format` | string |

Do NOT expose: `alarm_profiles`, `spell_corrections` (complex types, config-file only).

---

## Interactive Prompts (Survey/alternatives)

### AlecAivazis/survey/v2 -- current dependency

- **Status:** The repository was archived in late 2023. No new releases since v2.3.7 (the version in `go.mod`).
- **Works fine for now:** It compiles, it functions, terminal support is adequate.
- **Risk:** No security patches, no bug fixes, no new terminal support. The `golang.org/x/term` dependency is pinned to an old version (visible in go.mod: `v0.0.0-20210927222741`).
- **Current usage in Tempus:** Minimal -- only one `survey.Confirm` call in `main.go` line 168. The `internal/prompts/` package uses raw `bufio.Scanner`, not survey at all.

### Alternatives

**1. charmbracelet/huh -- recommended replacement**
- Active development (Charm ecosystem: bubbletea, lipgloss, huh)
- Modern terminal UI, accessibility features
- Form-based API that maps well to Tempus's `--interactive` needs
- Theming support (relevant for neurodivergent-friendly high contrast)
- Example:
  ```go
  form := huh.NewForm(
      huh.NewGroup(
          huh.NewInput().Title("Event title").Value(&title),
          huh.NewInput().Title("Date (YYYY-MM-DD)").Value(&date),
          huh.NewSelect[string]().
              Title("Alarm profile").
              Options(
                  huh.NewOption("ADHD Default", "adhd-default"),
                  huh.NewOption("Countdown", "adhd-countdown"),
                  huh.NewOption("Medication", "medication"),
              ).
              Value(&alarmProfile),
      ),
  )
  err := form.Run()
  ```

**2. charmbracelet/bubbletea -- lower level**
- Full TUI framework (Elm architecture)
- Overkill for forms/prompts, but the foundation `huh` builds on
- Use directly only if you need custom interactive widgets

**3. manifoldco/promptui -- legacy**
- Still works, minimal maintenance
- Less capable than both survey and huh
- No reason to choose this over huh

### Recommendation

**Replace survey/v2 with charmbracelet/huh** for the `--interactive` implementation (F2) and `tempus init` wizard (F1).

Rationale:
- survey is archived with known terminal issues
- huh is actively maintained by the Charm team (well-funded, Go-focused)
- huh's form API maps directly to Tempus's need: multi-field input for event creation
- Accessibility features align with neurodivergent target audience
- The Charm ecosystem (lipgloss for styling, huh for forms) is the clear community standard for Go CLI UX in 2025

**Migration cost is low:**
- Only one `survey.Confirm` call exists -- replace with `huh.NewConfirm()`
- The `internal/prompts/` package already abstracts user input -- add huh behind this interface
- Remove `survey/v2` from `go.mod` after migration

**Dependency change:**
```bash
go get github.com/charmbracelet/huh@latest
# After migration:
go mod tidy  # removes survey/v2
```

---

## Confidence Levels

| Section | Confidence | Rationale |
|---------|-----------|-----------|
| Go CLI Project Structure | **HIGH** | Patterns are well-established (gh, kubectl, cobra docs). Go project layout has been stable since 2019. Verified against actual codebase structure. |
| Cobra Command Packaging | **HIGH** | Standard patterns used by gh CLI (open source, auditable). Factory function pattern is the dominant approach in every major Go CLI. |
| i18n Patterns | **MEDIUM** | Recommendation to keep current system is sound. Could not verify latest go-i18n features via web. The "don't migrate" recommendation holds regardless. |
| Viper + Env Vars | **HIGH** | Viper's API is stable and well-documented. Confirmed no `AutomaticEnv` call exists in codebase (verified via grep). Pattern is standard Cobra+Viper. |
| Interactive Prompts (survey) | **MEDIUM** | Survey archival is well-known (happened 2023). Huh recommendation is based on Charm ecosystem dominance in Go TUI space. Could not verify latest huh release version via web -- confirm before adding dependency. |

### Verification needed before implementation

1. **huh latest version** -- run `go list -m -versions github.com/charmbracelet/huh` to confirm latest
2. **huh accessibility** -- verify huh supports screen readers / high contrast before committing to it for ND users
3. **Go 1.23 compatibility** -- project uses Go 1.23, verify huh minimum Go version requirement
