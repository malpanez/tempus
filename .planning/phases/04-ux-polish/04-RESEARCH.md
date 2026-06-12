# Phase 4: UX Polish - Research

**Researched:** 2026-03-30
**Domain:** Conflict detection UX + prep time customization in Go CLI
**Confidence:** HIGH

## Summary

Phase 4 is a focused UX improvement phase touching two existing features in `internal/cli/nd.go`: conflict detection output and prep time event naming. Both features already work functionally -- the changes are about making their output more useful (UX-03) and making naming configurable (UX-04).

The codebase is well-structured after Phase 3's refactor. The `App` struct provides config/writer injection, `PrintOK`/`PrintErr` handle consistent output, and existing test patterns in `nd_test.go` and `batch_test.go` cover both features. The changes are surgical: enhance `DetectEventConflicts` return data, add a config field + flag for prep time prefix, and thread the prefix through `GeneratePrepTimeEvents`.

**Primary recommendation:** Two plans -- (1) UX-03: enhance conflict detection with overlap duration and suggested fix times, (2) UX-04: add `prep_time_prefix` config key + `--prep-label` flag. Both are low-risk, code-only changes confined to `nd.go`, `batch.go`, and `config.go`.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UX-03 | Conflict detection shows event names, times, overlap duration to help user decide which to move | `DetectEventConflicts` at nd.go:263 returns `[]string` -- needs richer data (overlap minutes, suggestion). `collectBatchWarnings` at batch.go:182 formats output. All data (StartTime, EndTime, Summary) already available on `calendar.Event`. |
| UX-04 | Prep time event prefix customizable via `prep_time_prefix` config + `--prep-label` flag | `determinePrepTime` at nd.go:364 returns description string "Preparation". `createPrepEventIfNeeded` at nd.go:339 concatenates it into Summary. Config struct at config.go:16 needs new field. `batchOptions` at batch.go:47 needs new field. |
</phase_requirements>

## Architecture Patterns

### Current Code Flow (Conflict Detection)

```
runBatch() -> buildBatchCalendar() -> collectBatchWarnings() -> DetectEventConflicts()
                                                             -> prints warnings via fmt.Fprintln
```

**Key locations:**

| Function | File | Line | Signature | What it does |
|----------|------|------|-----------|--------------|
| `DetectEventConflicts` | nd.go | 263 | `(events []calendar.Event) []string` | O(n^2) pairwise check, returns formatted conflict strings |
| `collectBatchWarnings` | batch.go | 182 | `(events []calendar.Event, opts *batchOptions) []string` | Calls DetectEventConflicts + DetectOverwhelmDays, formats with bullet points |
| `writeBatchOutput` | batch.go | 251 | `(app *App, cal *calendar.Calendar, warnings []string, output string, eventCount int) error` | Prints warnings BEFORE writing file |
| `handleDryRun` | batch.go | 208 | `(app *App, validationErrors, warnings []string, records []batchRecord, input, output string) error` | Prints warnings in dry-run mode |

**Current conflict output format (nd.go:275-281):**
```
Event 1 (10:00-11:00) overlaps with Event 2 (10:30-12:00)
```

**What UX-03 needs to add:**
- Overlap duration: calculate `min(ev1.EndTime, ev2.EndTime) - max(ev1.StartTime, ev2.StartTime)`
- Suggested move time: `ev1.EndTime` (move ev2 to start after ev1 ends)
- Date context: include date when events span multiple days

### Current Code Flow (Prep Time)

```
runBatch() -> buildBatchCalendar() -> GeneratePrepTimeEvents() -> createPrepEventIfNeeded()
                                                                -> createTransitionEventIfNeeded()
```

**Key locations:**

| Function | File | Line | Signature | What it does |
|----------|------|------|-----------|--------------|
| `GeneratePrepTimeEvents` | nd.go | 290 | `(events []calendar.Event) []*calendar.Event` | Iterates events, creates prep/transition events |
| `createPrepEventIfNeeded` | nd.go | 331 | `(ev calendar.Event) *calendar.Event` | Calls `determinePrepTime`, builds event with hardcoded prefix |
| `determinePrepTime` | nd.go | 364 | `(summary string) (time.Duration, string)` | Returns duration + description ("Preparation" or "Travel & arrival buffer") |
| `createTransitionEventIfNeeded` | nd.go | 311 | `(ev calendar.Event) *calendar.Event` | Hardcoded "Transition: " prefix (out of scope for UX-04) |

**Hardcoded strings in nd.go:**
- Line 339: `"\u23f0 " + description + ": " + StripEmoji(ev.Summary)` -- description comes from `determinePrepTime`
- Line 372: `return 15 * time.Minute, "Preparation"` -- the string that UX-04 must make configurable
- Line 368: `return 20 * time.Minute, "Travel & arrival buffer"` -- medical prep, separate description
- Line 318: `"\U0001f504 Transition: "` -- transition events (not in scope for UX-04)
- Line 345: `Categories: []string{"Preparation"}` -- category string (keep as-is, UX-04 is about the summary prefix only)

### Pattern: Config Field + Flag Override

Existing pattern for config-backed flags (from Phase 2):

```go
// In config.go Config struct:
type Config struct {
    // ... existing fields ...
    PrepTimePrefix string `mapstructure:"prep_time_prefix" json:"prep_time_prefix"`
}

// In defaultConfig:
var defaultConfig = Config{
    // ... existing ...
    PrepTimePrefix: "Preparation",
}

// In Load() -- add viper.SetDefault:
viper.SetDefault("prep_time_prefix", defaultConfig.PrepTimePrefix)
```

```go
// In batch.go batchOptions:
type batchOptions struct {
    // ... existing ...
    prepLabel string
}

// In NewBatchCmd:
cmd.Flags().String("prep-label", "", "Custom prefix for prep time events (overrides config)")

// In parseBatchFlags:
opts.prepLabel, _ = cmd.Flags().GetString("prep-label")
```

**Priority chain:** flag > config > default ("Preparation")

```go
// Resolution in buildBatchCalendar or before calling GeneratePrepTimeEvents:
func resolvePrepLabel(flagValue string, cfg *config.Config) string {
    if flagValue != "" {
        return flagValue
    }
    if cfg != nil && cfg.PrepTimePrefix != "" {
        return cfg.PrepTimePrefix
    }
    return "Preparation"
}
```

### Pattern: Threading Label Through GeneratePrepTimeEvents

`GeneratePrepTimeEvents` currently takes only `[]calendar.Event`. To pass the label:

**Option A: Add parameter** (simplest, recommended)
```go
func GeneratePrepTimeEvents(events []calendar.Event, prepLabel string) []*calendar.Event
```

**Option B: Options struct** (over-engineering for one field)

Use Option A. It requires updating:
1. `nd.go:290` -- function signature
2. `nd.go:331` `createPrepEventIfNeeded` -- accept label param, use it instead of `determinePrepTime`'s returned description
3. `batch.go:173` -- call site in `buildBatchCalendar`
4. `nd_test.go:204` -- existing test call sites
5. `main.go` -- if any thin wrapper calls this (check needed)

**Important nuance:** `determinePrepTime` returns TWO different descriptions: "Preparation" (meetings) and "Travel & arrival buffer" (medical). UX-04 only customizes the generic "Preparation" prefix. When `determinePrepTime` returns "Travel & arrival buffer", that should remain unchanged regardless of the custom prefix. The label parameter should only override when the description is "Preparation".

### Anti-Patterns to Avoid
- **Breaking the determinePrepTime return contract:** Don't remove the description return -- it distinguishes medical prep from meeting prep. Only override when description == "Preparation".
- **Global config access in nd.go:** `GeneratePrepTimeEvents` must NOT call `config.Load()` directly. Pass the label as a parameter (keeps functions pure and testable).
- **Changing transition event prefix:** UX-04 scope is prep time only. "Transition:" stays hardcoded.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Overlap duration calculation | Time range intersection library | `min(end1,end2) - max(start1,start2)` inline | Two lines of Go; no library needed |
| Suggested move time | Complex scheduling algorithm | `ev1.EndTime.Format("15:04")` as suggestion | Simple "move after" is sufficient per requirements |

## Existing Infrastructure to Reuse

| What | Where | How to Use |
|------|-------|------------|
| `PrintOK` / `PrintErr` | helpers.go:230-237 | Output formatting with emoji prefixes |
| `App.Stdout` | app.go:16 | Writer injection for testability |
| `App.Config` | app.go:14 | Access `PrepTimePrefix` after config loads |
| `batchOptions` struct | batch.go:47 | Add `prepLabel` field |
| `testutil` constants | internal/testutil/ | Event title constants, timezone constants |
| `collectBatchWarnings` | batch.go:182 | Already formats conflict output -- enhance here |
| `viper.SetDefault` pattern | config.go:88-96 | Add `prep_time_prefix` default |

## Common Pitfalls

### Pitfall 1: Overlap Duration Calculation Edge Case
**What goes wrong:** Events on different dates but same clock times produce wrong overlap duration.
**Why it happens:** Using only `Format("15:04")` ignores the date component.
**How to avoid:** Always use full `time.Time` values for overlap calculation, not formatted strings. The current code already uses `ev1.EndTime.After(ev2.StartTime)` with full timestamps.

### Pitfall 2: Prep Label vs. Medical Description
**What goes wrong:** Custom prep label overrides "Travel & arrival buffer" for doctor appointments.
**Why it happens:** Blindly replacing all descriptions with the custom label.
**How to avoid:** Only replace when `determinePrepTime` returns "Preparation" as the description. Keep "Travel & arrival buffer" unchanged.

### Pitfall 3: Empty Prep Label
**What goes wrong:** User sets `--prep-label ""` and gets events with ": Event Name" summary.
**How to avoid:** Treat empty string as "not set" -- fall back to config, then default.

### Pitfall 4: Test Coverage Regression
**What goes wrong:** New code paths added without tests drops below 79%.
**How to avoid:** Add test cases for: overlap duration in conflict output, custom prep label, flag > config > default priority, empty label fallback.

## Code Examples

### UX-03: Enhanced Conflict Detection Output

Current output (nd.go:275-281):
```
Event 1 (10:00-11:00) overlaps with Event 2 (10:30-12:00)
```

Target output:
```
Event 1 (10:00-11:00) overlaps with Event 2 (10:30-12:00) by 30min. Suggestion: move Event 2 to 11:00
```

Implementation in `DetectEventConflicts`:
```go
func DetectEventConflicts(events []calendar.Event) []string {
    var conflicts []string
    for i := 0; i < len(events); i++ {
        for j := i + 1; j < len(events); j++ {
            ev1, ev2 := events[i], events[j]
            if ev1.AllDay || ev2.AllDay {
                continue
            }
            if ev1.EndTime.After(ev2.StartTime) && ev2.EndTime.After(ev1.StartTime) {
                overlapStart := ev1.StartTime
                if ev2.StartTime.After(overlapStart) {
                    overlapStart = ev2.StartTime
                }
                overlapEnd := ev1.EndTime
                if ev2.EndTime.Before(overlapEnd) {
                    overlapEnd = ev2.EndTime
                }
                overlapDuration := overlapEnd.Sub(overlapStart)
                suggestion := ev1.EndTime.Format("15:04")

                conflict := fmt.Sprintf("%s (%s-%s) overlaps with %s (%s-%s) by %s. Suggestion: move %s to %s",
                    ev1.Summary,
                    ev1.StartTime.Format("15:04"),
                    ev1.EndTime.Format("15:04"),
                    ev2.Summary,
                    ev2.StartTime.Format("15:04"),
                    ev2.EndTime.Format("15:04"),
                    formatDuration(overlapDuration),
                    ev2.Summary,
                    suggestion)
                conflicts = append(conflicts, conflict)
            }
        }
    }
    return conflicts
}

func formatDuration(d time.Duration) string {
    minutes := int(d.Minutes())
    if minutes < 60 {
        return fmt.Sprintf("%dm", minutes)
    }
    hours := minutes / 60
    remaining := minutes % 60
    if remaining == 0 {
        return fmt.Sprintf("%dh", hours)
    }
    return fmt.Sprintf("%dh%dm", hours, remaining)
}
```

### UX-04: Configurable Prep Label

Config addition (config.go):
```go
type Config struct {
    // ... existing fields ...
    PrepTimePrefix string `mapstructure:"prep_time_prefix" json:"prep_time_prefix"`
}

// In defaultConfig:
PrepTimePrefix: "Preparation",

// In Load():
viper.SetDefault("prep_time_prefix", defaultConfig.PrepTimePrefix)
```

Signature change (nd.go):
```go
func GeneratePrepTimeEvents(events []calendar.Event, prepLabel string) []*calendar.Event

func createPrepEventIfNeeded(ev calendar.Event, prepLabel string) *calendar.Event {
    duration, description := determinePrepTime(ev.Summary)
    if duration == 0 {
        return nil
    }
    if description == "Preparation" && prepLabel != "" {
        description = prepLabel
    }
    // ... rest unchanged
}
```

Flag + resolution (batch.go):
```go
// In NewBatchCmd:
cmd.Flags().String("prep-label", "", "Custom prefix for prep time events (overrides config)")

// In parseBatchFlags:
opts.prepLabel, _ = cmd.Flags().GetString("prep-label")

// In buildBatchCalendar, before calling GeneratePrepTimeEvents:
prepLabel := resolvePrepLabel(opts.prepLabel, app.Config)
prepEvents := GeneratePrepTimeEvents(cal.Events, prepLabel)
```

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none (stdlib) |
| Quick run command | `go test ./internal/cli/ -run "TestDetectEventConflicts\|TestGeneratePrepTimeEvents\|TestCollectBatchWarnings" -v` |
| Full suite command | `go test ./...` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UX-03 | Conflict output includes overlap duration | unit | `go test ./internal/cli/ -run TestDetectEventConflicts -v` | Exists (nd_test.go:107) -- needs new assertions |
| UX-03 | Conflict output includes move suggestion | unit | `go test ./internal/cli/ -run TestDetectEventConflicts -v` | Exists -- needs new assertions |
| UX-04 | Custom prep label via config | unit | `go test ./internal/cli/ -run TestGeneratePrepTimeEvents -v` | Exists (nd_test.go:194) -- needs new test cases |
| UX-04 | Custom prep label via --prep-label flag | unit | `go test ./internal/cli/ -run TestCollectBatchWarnings -v` | Exists (batch_test.go:670) -- needs new test cases |
| UX-04 | Flag overrides config | unit | `go test ./internal/cli/ -run TestPrepLabel -v` | New test needed |
| UX-04 | Medical prep ignores custom label | unit | `go test ./internal/cli/ -run TestGeneratePrepTimeEvents -v` | Exists -- needs new assertion |

### Sampling Rate
- **Per task commit:** `go test ./internal/cli/ -v -count=1`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green + coverage >= 79%

### Wave 0 Gaps
- [ ] `TestDetectEventConflicts` needs assertions for overlap duration string and suggestion string
- [ ] `TestGeneratePrepTimeEvents` needs test cases with custom label parameter
- [ ] New `TestResolvePrepLabel` for flag > config > default priority
- [ ] `TestCollectBatchWarnings` may need update for new conflict output format

## Exact Changes Required

### File: `internal/config/config.go`
1. **Line 26 (Config struct):** Add `PrepTimePrefix string \`mapstructure:"prep_time_prefix" json:"prep_time_prefix"\``
2. **Line 36 (defaultConfig):** Add `PrepTimePrefix: "Preparation",`
3. **Line ~96 (Load func):** Add `viper.SetDefault("prep_time_prefix", defaultConfig.PrepTimePrefix)`

### File: `internal/cli/nd.go`
4. **Line 263 (DetectEventConflicts):** Enhance conflict string to include overlap duration and suggestion
5. **New function:** `formatDuration(d time.Duration) string` -- human-readable duration
6. **Line 290 (GeneratePrepTimeEvents):** Add `prepLabel string` parameter
7. **Line 331 (createPrepEventIfNeeded):** Add `prepLabel string` parameter, use it when description == "Preparation"

### File: `internal/cli/batch.go`
8. **Line 47 (batchOptions):** Add `prepLabel string` field
9. **Line 62 (NewBatchCmd):** Add `--prep-label` flag
10. **Line 120 (parseBatchFlags):** Parse `--prep-label` flag
11. **Line 172-177 (buildBatchCalendar):** Resolve prep label, pass to `GeneratePrepTimeEvents`
12. **buildBatchCalendar needs access to App.Config:** Currently only receives `(records, opts)` -- needs config for prep label resolution. Either pass config or resolve label in `runBatch` and store in opts.

### File: `internal/cli/nd_test.go`
13. **Line 128, 146, 151, 156:** Update `DetectEventConflicts` call assertions for new output format
14. **Line 204, 233, 248, 262, 273, 278:** Update `GeneratePrepTimeEvents` calls to pass label parameter
15. **New test cases:** Custom label, empty label, medical-keeps-description

### File: `internal/cli/batch_test.go`
16. **Line 747 (collectBatchWarnings tests):** Update assertions for new conflict output format

## Threading Config Into buildBatchCalendar

`buildBatchCalendar` currently receives `(records []batchRecord, opts *batchOptions)` but not config. Two options:

**Option A (recommended): Resolve in runBatch, store in opts**
```go
// In runBatch, after parseBatchFlags:
if opts.prepLabel == "" && app.Config != nil {
    opts.prepLabel = app.Config.PrepTimePrefix
}
```
This keeps `buildBatchCalendar` signature unchanged and uses `opts.prepLabel` (already resolved).

**Option B: Pass app to buildBatchCalendar** -- more invasive, changes signature.

Use Option A. The resolution is: flag (non-empty) > config > default "Preparation".

## Sources

### Primary (HIGH confidence)
- Direct codebase inspection of nd.go, batch.go, config.go, nd_test.go, batch_test.go
- Existing test patterns and `App` struct DI pattern

### Secondary (MEDIUM confidence)
- Viper config patterns verified against working Phase 2 implementation (config.go:88-100)

## Metadata

**Confidence breakdown:**
- UX-03 (conflict enhancement): HIGH -- existing code is clear, change is additive
- UX-04 (prep label config): HIGH -- follows established Viper pattern from Phase 2
- Test patterns: HIGH -- existing tests in nd_test.go and batch_test.go provide clear patterns

**Research date:** 2026-03-30
**Valid until:** 2026-04-30 (stable domain, no external dependencies)
