# Architecture Research

**Project:** Tempus CLI Refactor
**Researched:** 2026-03-29
**Overall confidence:** HIGH (based on direct codebase analysis + established Go patterns)

## Proposed Package Structure Assessment

The proposed structure (`commands/`, `parsing/`, `nd/`, `output/`) is a good start but needs refinement. After analyzing the 176 functions in `main.go` and their dependency graph, here is the recommended structure:

### Recommended Structure

```
cmd/tempus/
  main.go              <- Thin entry point (~30 lines): calls root.Execute()

internal/
  cli/
    root.go            <- newRootCmd(), App struct with shared deps
    create.go          <- newCreateCmd(), runCreate, createOptions
    quick.go           <- newQuickCmd(), runQuick, quickParsedEvent
    batch.go           <- newBatchCmd(), batch processing pipeline
    lint.go            <- newLintCmd(), lintState, lint logic
    template.go        <- newTemplateCmd(), newTemplateInitCmd(), newBatchTemplateCmd()
    rrule.go           <- newRRuleHelperCmd()
    timezone.go        <- newTimezoneCmd()
    locale.go          <- newLocaleCmd()
    config.go          <- newConfigCmd()
    version.go         <- newVersionCmd()
    output.go          <- printOK, printErr with io.Writer injection

  parsing/
    datetime.go        <- Unified parser (replaces 13 parse* functions)
    duration.go        <- parseDuration, fmtDurationHuman
    batch.go           <- batchRecord parsing, format detection, CSV/JSON/YAML loaders

  nd/
    spellcheck.go      <- correctCategory, levenshteinDistance
    conflicts.go       <- detectEventConflicts (sweep line)
    overwhelm.go       <- detectOverwhelmDays
    preptime.go        <- generatePrepTimeEvents
    emoji.go           <- stripEmoji (fixed), autoEmoji, emojiForCategory

  calendar/            <- Already exists, keep as-is
  config/              <- Already exists, keep as-is
  constants/           <- Already exists, keep as-is
  i18n/                <- Already exists, keep as-is
  normalizer/          <- Already exists, keep as-is
  prompts/             <- Already exists, keep as-is
  templates/           <- Already exists, keep as-is
  timezone/            <- Already exists, keep as-is
  utils/               <- Already exists, keep as-is
```

### Why This Differs From the Proposal

1. **`internal/cli/` instead of `internal/commands/`**. The `cli` name is more idiomatic in Go (see: `kubectl`, `gh`, `terraform`). It also clarifies that this package owns CLI wiring, not just command definitions.

2. **Keep `main.go` at repo root (or move to `cmd/tempus/`)**.
   - Option A: Keep `main.go` at root. Simpler, no import path changes. All tests stay in `package main` for now.
   - Option B: Move to `cmd/tempus/main.go`. Standard Go project layout. Requires moving test files.
   - **Recommendation: Option A for now.** Moving to `cmd/` is a separate, later cleanup. The priority is extracting logic, not reorganizing the build. During the refactor, `main.go` shrinks as functions move to `internal/cli/` and other packages. Tests stay in `package main` and gradually get rewritten as package-level tests in the target packages.

3. **`nd/` package is correct**. Neurodivergent features (spellcheck, conflicts, overwhelm, preptime, emoji) form a cohesive domain. This is a natural bounded context. Keep the name short; document the abbreviation in the package doc comment.

4. **`output.go` stays inside `internal/cli/`, not a separate package**. `printOK`/`printErr` are 5-line functions. A whole package is over-engineering. They belong in `cli/output.go` as package-level helpers with `io.Writer` injection via the `App` struct.

5. **`parsing/batch.go` moves batch format detection/loading into `parsing/`**. The batch loading logic (CSV/JSON/YAML detection, record normalization) is parsing, not command logic. `cli/batch.go` should call `parsing.LoadBatchRecords()`.

## Shared State Pattern (Config/Translator across commands)

### Current Problem

`config.Load()` is called 6+ times independently across commands. Each call reads from disk and creates a new Config. There is no shared translator -- `newTranslator()` is called per-command. This works but is inefficient and means there is no single place to override config (e.g., for env vars).

### Recommended Pattern: App Struct with Dependency Injection

```go
// internal/cli/root.go
package cli

import (
    "io"
    "tempus/internal/config"
    "tempus/internal/i18n"
)

// App holds shared dependencies for all commands.
// Created once in NewRootCmd, passed to subcommands via closures.
type App struct {
    Config     *config.Config
    Translator *i18n.Translator
    Stdout     io.Writer
    Stderr     io.Writer
}

func NewRootCmd() *cobra.Command {
    app := &App{
        Stdout: os.Stdout,
        Stderr: os.Stderr,
    }

    cmd := &cobra.Command{
        Use:   "tempus",
        Short: "A multilingual ICS calendar file generator",
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load()
            if err != nil {
                cfg = config.Default()
            }
            app.Config = cfg

            lang := resolveLang(cmd, cfg)
            tr, err := i18n.NewTranslator(lang)
            if err != nil {
                tr, _ = i18n.NewTranslator("en")
            }
            app.Translator = tr
            return nil
        },
        SilenceUsage: true,
    }

    cmd.PersistentFlags().StringP("language", "l", "", "Language for output")
    cmd.PersistentFlags().StringP("timezone", "t", "", "Default timezone")
    cmd.PersistentFlags().StringP("config", "c", "", "Config file path")

    cmd.AddCommand(
        newCreateCmd(app),
        newBatchCmd(app),
        newLintCmd(app),
        // ...
    )

    return cmd
}
```

### Why This Pattern

- **Single init point**: Config loaded once in `PersistentPreRunE`, available to all subcommands.
- **Testable**: Tests create `App{Stdout: &buf, Config: testConfig}` -- no disk I/O, no global state.
- **No global variables**: The current `scanner` global var disappears. `App` holds everything.
- **Cobra-idiomatic**: This is exactly how `kubectl`, `gh`, and `hugo` handle shared state. Subcommand constructors receive `*App` and close over it.

### Do NOT use `context.Context` for this

Storing config/translator in `context.Context` is an anti-pattern in Go. Context is for cancellation and request-scoped values, not dependency injection. Use a struct.

## Incremental Refactoring Strategy

The key constraint: refactor in parallel with bug fixes and features, never breaking the test suite.

### Step-by-Step Approach

**Phase 0: Scaffolding (1 PR, no logic changes)**

1. Create the empty package directories: `internal/cli/`, `internal/parsing/`, `internal/nd/`.
2. Create `internal/cli/app.go` with the `App` struct (fields only, no methods).
3. Create `internal/cli/output.go` with `printOK`/`printErr` that take `io.Writer`.
4. All tests still pass because nothing moved yet.

**Phase 1: Extract leaf functions first (bottom-up)**

Leaf functions have no dependencies on other `main.go` functions. Extract them first because they can be tested independently.

Priority order:
1. `levenshteinDistance` -> `nd/spellcheck.go` (pure function, zero deps)
2. `stripEmoji` -> `nd/emoji.go` (pure function, fix the bug while moving)
3. `fmtDurationHuman`, `atoiSafe` -> `parsing/duration.go` or `utils/`
4. `detectEventConflicts` -> `nd/conflicts.go` (depends only on `calendar.Event`)
5. `detectOverwhelmDays` -> `nd/overwhelm.go` (depends only on `calendar.Event`)
6. `generatePrepTimeEvents` -> `nd/preptime.go` (depends on `calendar.Event`)
7. `correctCategory` -> `nd/spellcheck.go` (depends on `levenshteinDistance`)

For each extraction:
```
1. Copy function to target package.
2. Export it (capitalize first letter).
3. In main.go, replace the body with a call to the new package:
     func levenshteinDistance(s1, s2 string) int {
         return nd.LevenshteinDistance(s1, s2)
     }
4. Run tests. They pass because the old function signature is unchanged.
5. In a later PR, inline the call sites and remove the wrapper.
```

This "wrapper then inline" pattern is critical. It means every intermediate commit compiles and passes tests.

**Phase 2: Extract parsing functions**

After leaf functions are out, extract the datetime parsing cluster:
1. Create `parsing.ParseDateTime()` unified function (see next section).
2. Wire `parseDateTimeWithTZ`, `parseAllDayTimes`, `parseBatchTimes`, etc. as thin wrappers calling `parsing.ParseDateTime()`.
3. Gradually replace call sites.

**Phase 3: Extract commands (one at a time)**

Order by independence (least coupled first):
1. `version.go` -- zero dependencies, 20 lines. Proof of concept.
2. `lint.go` -- self-contained, depends only on ICS parsing.
3. `locale.go` -- depends on config + i18n.
4. `timezone.go` -- depends on timezone package only.
5. `rrule.go` -- self-contained interactive helper.
6. `config.go` -- depends on config package.
7. `template.go` -- moderate coupling.
8. `create.go` -- high coupling (parsing, nd, output).
9. `quick.go` -- depends on create patterns.
10. `batch.go` -- highest coupling (parsing, nd, output, templates).

For each command extraction:
1. Create `internal/cli/xxx.go` with `newXxxCmd(app *App)`.
2. Move the `RunE` function and its helpers.
3. In `main.go`, replace `newXxxCmd()` with `cli.NewXxxCmd(app)`.
4. Run tests. Fix imports.

**Phase 4: Shrink main.go**

After all commands are extracted, `main.go` becomes:
```go
package main

import (
    "os"
    "tempus/internal/cli"
)

func main() {
    if err := cli.NewRootCmd().Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Test Migration Strategy

Tests in `main_test.go` and `main_*_test.go` are in `package main`. They test unexported functions directly. This is the main friction point.

**Approach: Copy-then-delete, not move**

1. When extracting `func foo()` to `internal/nd/`, write a NEW test in `internal/nd/foo_test.go` that covers the same cases.
2. Keep the old test in `main_test.go` (it now tests the wrapper).
3. Once all call sites are inlined and the wrapper is removed, delete the old test.
4. Coverage only goes UP because you now have two test paths temporarily.

**Never delete a test before the new one exists.** The coverage gate (79%) enforces this.

## Unified DateTime Parser Design

### Current Problem

There are 13 parse functions in `main.go` that handle datetime parsing with overlapping logic:
- `parseCreateTimes`, `parseAllDayTimes`, `parseTimedEventTimes`, `parseEndTime`, `parseDurationEnd` (create path)
- `parseBatchTimes`, `parseBatchAllDayTimes`, `parseBatchTimedEventTimes`, `parseBatchEndTime`, `parseBatchExplicitEnd`, `parseBatchDurationEnd` (batch path)
- `parseDateTimeWithTZ` (shared helper)
- `normalizeTimeInput` (normalization)

The create and batch paths are nearly identical but duplicated.

### Recommended Design

```go
// internal/parsing/datetime.go
package parsing

import "time"

// TimeSpec represents a parsed datetime with its timezone context.
type TimeSpec struct {
    Time   time.Time
    TZ     string       // IANA timezone name
    AllDay bool
}

// ParseOptions controls parsing behavior.
type ParseOptions struct {
    StartDate    string   // "2025-12-16" or "2025/12/16" or "tomorrow"
    StartTime    string   // "09:00" or "" for all-day
    EndDate      string   // optional
    EndTime      string   // optional
    Duration     string   // "1h30m" or "" -- alternative to end time
    StartTZ      string   // IANA timezone for start
    EndTZ        string   // IANA timezone for end (defaults to StartTZ)
    Normalize    bool     // Apply input normalization (slash->dash, etc.)
}

// ParseResult holds the resolved start and end times.
type ParseResult struct {
    Start  TimeSpec
    End    TimeSpec
    AllDay bool
}

// Parse resolves start/end times from flexible input formats.
// This replaces all 13 parse functions in main.go.
func Parse(opts ParseOptions) (ParseResult, error) {
    if opts.Normalize {
        opts = normalize(opts)
    }
    if opts.EndTZ == "" {
        opts.EndTZ = opts.StartTZ
    }

    if opts.StartTime == "" && opts.EndTime == "" && opts.Duration == "" {
        return parseAllDay(opts)
    }
    return parseTimed(opts)
}

func normalize(opts ParseOptions) ParseOptions {
    // Delegate to normalizer package for date/time normalization
    // Handles: 2025/12/16 -> 2025-12-16, 9:00 -> 09:00, etc.
    return opts
}

func parseAllDay(opts ParseOptions) (ParseResult, error) {
    // Parse start date, default end = start + 1 day
    // ...
}

func parseTimed(opts ParseOptions) (ParseResult, error) {
    // Parse start datetime
    // Resolve end: explicit end time > duration > default 1h
    // ...
}
```

### Migration Path

1. Implement `parsing.Parse()` with full test coverage.
2. Rewrite `parseDateTimeWithTZ` to call `parsing.Parse()`.
3. Rewrite `parseCreateTimes` to build `ParseOptions` from `createOptions` and call `parsing.Parse()`.
4. Rewrite `parseBatchTimes` to build `ParseOptions` from `batchRecord` and call `parsing.Parse()`.
5. Remove the 10+ intermediate functions.

### Why Not a Strategy Pattern

A strategy pattern (interface with multiple implementations) is over-engineering here. The 13 functions do not represent different strategies -- they represent the same strategy copy-pasted with minor variations for create vs batch context. A single function with an options struct collapses all the duplication. The `ParseOptions` struct IS the variation point.

## Performance Optimizations

### Levenshtein Distance Caching

**Current behavior:** `correctCategory()` calls `levenshteinDistance()` for every category against a map of ~25 known categories. In batch processing with 1000 events, this runs 25,000 comparisons.

**Optimization: Precomputed lookup + cache**

```go
// internal/nd/spellcheck.go
package nd

import "sync"

// CategoryCorrector holds precomputed correction state.
type CategoryCorrector struct {
    known     map[string]string // lowercase -> canonical
    cache     sync.Map          // input -> corrected (thread-safe)
    threshold int
}

func NewCategoryCorrector(known map[string]string, threshold int) *CategoryCorrector {
    return &CategoryCorrector{
        known:     known,
        threshold: threshold,
    }
}

func (c *CategoryCorrector) Correct(category string) string {
    lower := strings.ToLower(category)

    // Check exact match first (O(1))
    if canonical, ok := c.known[lower]; ok {
        return canonical
    }

    // Check cache (already computed)
    if cached, ok := c.cache.Load(lower); ok {
        return cached.(string)
    }

    // Compute Levenshtein and cache result
    best := category
    bestDist := c.threshold + 1
    for known, canonical := range c.known {
        dist := LevenshteinDistance(lower, known)
        if dist <= c.threshold && dist < bestDist {
            bestDist = dist
            best = canonical
        }
    }

    c.cache.Store(lower, best)
    return best
}
```

**Impact:** For batch processing, the cache means each unique category string is computed once. With typical batch data having 5-10 unique categories across 1000 events, this reduces from 25,000 to ~250 Levenshtein computations. That is a 100x reduction.

**Additional optimization (if needed):** Use a BK-tree for the known categories dictionary. For 25 entries it is overkill, but if the dictionary grows to hundreds (e.g., with user spell corrections from config), a BK-tree reduces comparisons from O(n) to O(n^0.6) on average. Not needed now.

### Levenshtein Algorithm Optimization

The current implementation allocates a full `len(s1) x len(s2)` matrix. For short strings (category names, typically 5-15 chars), this is fine. But if you want to optimize:

```go
// Two-row optimization: O(min(m,n)) space instead of O(m*n)
func LevenshteinDistance(s1, s2 string) int {
    if len(s1) < len(s2) {
        s1, s2 = s2, s1 // ensure s1 is longer
    }
    prev := make([]int, len(s2)+1)
    curr := make([]int, len(s2)+1)
    for j := range prev {
        prev[j] = j
    }
    for i := 1; i <= len(s1); i++ {
        curr[0] = i
        for j := 1; j <= len(s2); j++ {
            cost := 0
            if s1[i-1] != s2[j-1] {
                cost = 1
            }
            curr[j] = min(
                prev[j]+1,
                curr[j-1]+1,
                prev[j-1]+cost,
            )
        }
        prev, curr = curr, prev
    }
    return prev[len(s2)]
}
```

The real win is the cache, not the algorithm. Both are O(m*n) time, but the two-row version avoids GC pressure from the matrix allocation.

### Conflict Detection: O(n^2) to O(n log n)

**Current implementation** (lines 1560-1587): Nested loop comparing every pair of events. For 1000 events, that is 500,000 comparisons.

**Recommended: Sweep line algorithm**

```go
// internal/nd/conflicts.go
package nd

import (
    "fmt"
    "sort"
    "tempus/internal/calendar"
)

// DetectConflicts finds overlapping events using a sweep line algorithm.
// Time complexity: O(n log n) due to sorting.
func DetectConflicts(events []calendar.Event) []string {
    // Filter out all-day events
    timed := make([]calendar.Event, 0, len(events))
    for _, ev := range events {
        if !ev.AllDay {
            timed = append(timed, ev)
        }
    }

    if len(timed) < 2 {
        return nil
    }

    // Sort by start time
    sort.Slice(timed, func(i, j int) bool {
        return timed[i].StartTime.Before(timed[j].StartTime)
    })

    var conflicts []string

    // Sweep: track the latest-ending event seen so far
    // For each event, compare only with events that could overlap
    // (those whose end time is after our start time)
    //
    // Simple approach: since events are sorted by start, we only need
    // to check against "active" events (those not yet ended).
    // Use a max-end tracker.
    for i := 1; i < len(timed); i++ {
        for j := i - 1; j >= 0; j-- {
            // Since sorted by start, if event j ended before event i starts,
            // all events before j also ended before i starts.
            if !timed[j].EndTime.After(timed[i].StartTime) {
                break
            }
            conflicts = append(conflicts, fmt.Sprintf(
                "%s (%s-%s) overlaps with %s (%s-%s)",
                timed[j].Summary,
                timed[j].StartTime.Format("15:04"),
                timed[j].EndTime.Format("15:04"),
                timed[i].Summary,
                timed[i].StartTime.Format("15:04"),
                timed[i].EndTime.Format("15:04"),
            ))
        }
    }

    return conflicts
}
```

**Complexity analysis:**
- Sorting: O(n log n)
- Sweep: O(n + k) where k is the number of conflicts
- Total: O(n log n + k) -- much better than O(n^2) for typical calendar data where most events do NOT overlap

**Important caveat:** The inner loop looks O(n^2) in the worst case (all events overlap). But for calendar data, this does not happen -- events are generally disjoint. The `break` on sorted data means the inner loop typically runs 0-2 iterations. For pathological input (all events at the same time), you would need an interval tree, but that is overkill for calendar data.

**Alternative (if truly needed): Interval tree**

If you want guaranteed O(n log n + k), use an augmented interval tree. But for Tempus's use case (batch processing of hundreds, maybe low thousands of events), the sorted sweep line above is sufficient and much simpler.

## Build Order (Dependencies Between Components)

Based on the dependency analysis of the 176 functions in `main.go`:

```
Layer 0 (no deps on main.go functions):
  internal/nd/spellcheck.go     <- levenshteinDistance, correctCategory
  internal/nd/emoji.go          <- stripEmoji (+ fix bug B1)
  internal/cli/output.go        <- printOK, printErr
  internal/parsing/duration.go  <- fmtDurationHuman, atoiSafe

Layer 1 (depends on Layer 0 + existing packages):
  internal/parsing/datetime.go  <- unified parser
  internal/nd/conflicts.go      <- detectEventConflicts
  internal/nd/overwhelm.go      <- detectOverwhelmDays
  internal/nd/preptime.go       <- generatePrepTimeEvents (uses stripEmoji)

Layer 2 (depends on Layer 1):
  internal/parsing/batch.go     <- batch loading, format detection
  internal/cli/app.go           <- App struct, PersistentPreRunE

Layer 3 (depends on Layer 2):
  internal/cli/version.go       <- simplest command
  internal/cli/lint.go          <- self-contained
  internal/cli/locale.go        <- depends on i18n
  internal/cli/timezone.go      <- depends on timezone
  internal/cli/rrule.go         <- self-contained interactive
  internal/cli/config.go        <- depends on config

Layer 4 (depends on Layers 0-3):
  internal/cli/create.go        <- depends on parsing, nd, output
  internal/cli/template.go      <- depends on templates, parsing

Layer 5 (depends on everything):
  internal/cli/quick.go         <- depends on create patterns
  internal/cli/batch.go         <- depends on parsing, nd, templates
  internal/cli/root.go          <- wires everything together
```

### Recommended Build Order for PRs

Each PR should be reviewable independently and leave the codebase green.

| PR | What | Lines Moved | Risk |
|----|------|-------------|------|
| 1 | Scaffold: create empty packages, App struct, output.go | ~30 new | None |
| 2 | Extract nd/spellcheck.go + nd/emoji.go (fix B1 here) | ~80 | Low |
| 3 | Extract nd/conflicts.go (O(n log n)) + nd/overwhelm.go + nd/preptime.go | ~150 | Low |
| 4 | Extract parsing/datetime.go (unified parser, addresses R2) | ~200 | Medium |
| 5 | Extract parsing/batch.go (format detection, loaders) | ~150 | Medium |
| 6 | Extract cli/version.go + cli/lint.go + cli/config.go | ~300 | Low |
| 7 | Extract cli/locale.go + cli/timezone.go + cli/rrule.go | ~400 | Low |
| 8 | Extract cli/create.go + cli/template.go | ~500 | Medium |
| 9 | Extract cli/quick.go + cli/batch.go | ~600 | High |
| 10 | Final: slim main.go to ~30 lines, remove wrappers | ~50 | Low |

**Total: ~2,460 lines moved across 10 PRs** (some lines deleted as duplication is removed).

PRs 2-3 can happen in parallel with bug fix PRs (B1-B5) because the extraction addresses the bugs directly. PR 4 addresses R2 (unify parsing). PR 3 addresses R4 (O(n log n) conflicts). Levenshtein caching (R5) fits naturally in PR 2.

### Key Risk: PR 9 (batch.go)

Batch processing is the most coupled command (~1,200 lines from `newBatchCmd` at line 571 to related helpers ending around line 1780). It touches parsing, nd features, templates, and output. Extract it last when all its dependencies are already in separate packages.

## Sources

- Direct codebase analysis of `/home/malpanez/repos/tempus/main.go` (3,906 lines, 176 functions)
- Go standard project layout patterns (based on `kubectl`, `gh`, `hugo`, `terraform` CLI architectures)
- Cobra library documentation for `PersistentPreRunE` and subcommand patterns
- Sweep line algorithm for interval overlap detection (computational geometry standard)
- Levenshtein distance optimization (two-row space optimization, standard dynamic programming)

**Confidence levels:**
- Package structure: HIGH (standard Go patterns, verified against existing codebase)
- Shared state pattern: HIGH (Cobra `PersistentPreRunE` + App struct is the canonical approach)
- Incremental refactoring: HIGH (wrapper-then-inline is battle-tested)
- Unified parser: HIGH (options struct pattern, direct analysis of the 13 parse functions)
- Performance optimizations: HIGH (standard algorithms, appropriate for the data scale)
- Build order: MEDIUM (dependency analysis is sound, but PR sizing may need adjustment based on actual coupling discovered during extraction)
