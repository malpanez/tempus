# Research Summary -- Tempus

**Project:** Tempus (Go CLI for ICS calendar generation)
**Domain:** CLI tool / productivity / neurodivergent UX
**Researched:** 2026-03-29
**Confidence:** HIGH (codebase analysis + established Go patterns; no live web verification)

---

## TL;DR

1. **Fix 5 bugs first** -- B1 (stripEmoji breaks Spanish/Portuguese), B3 (alarm profiles silently corrupt ICS), and B4 (city lookup swallows errors) are data-integrity issues that erode trust. ADHD users abandon tools that feel broken.
2. **Replace survey/v2 with charmbracelet/huh** -- survey is archived (2023), huh is the Go ecosystem standard for interactive forms. Migration cost is trivial (one `survey.Confirm` call). Do this when building `--interactive` and `tempus init`.
3. **`tempus init` and `--interactive` are make-or-break** -- the user's real workflow (recruiter meetings, school dates, travel planning, Google Calendar via ICS) depends on fast event creation. The init wizard + interactive mode + practical templates (school-event, recruiter-meeting, travel) are the critical adoption path for both the user and future open-source contributors.
4. **Refactor incrementally, never big-bang** -- the 3,906-line `main.go` splits into `internal/cli/`, `internal/parsing/`, `internal/nd/` using the "wrapper then inline" pattern. Each PR compiles and passes tests. Coverage gate at 79% minimum.
5. **App struct for dependency injection** -- load config once in `PersistentPreRunE`, pass `*App` to all subcommands. Eliminates 6+ redundant `config.Load()` calls and enables testability.

---

## Stack Decisions

**Core stack (keep as-is):**
- **Go + Cobra + Viper** -- standard CLI stack, well-tested in Tempus already. No changes needed.
- **olebedev/when** -- natural language date parsing. Adequate for "tomorrow", "next tuesday". Do not expand scope.
- **Homegrown i18n** -- Tempus's flat JSON key-value translator is sufficient for ~80 message keys. Do NOT migrate to `go-i18n` or `golang.org/x/text`.

**One change: charmbracelet/huh replaces survey/v2.**
- survey/v2 is archived since 2023 with no security patches.
- Only one `survey.Confirm` call exists -- migration is a 10-line change.
- huh provides form-based API that maps directly to `--interactive` needs, with accessibility features relevant to the ND target audience.
- Verify before adding: `go list -m -versions github.com/charmbracelet/huh` for Go 1.23 compatibility.

**Project structure target:**
- `cmd/tempus/main.go` (thin entry, ~30 lines) -- defer this move; keep root `main.go` during refactor.
- `internal/cli/` (command definitions + App struct), `internal/parsing/` (unified datetime parser), `internal/nd/` (neurodivergent features).
- Existing `internal/` packages (calendar, config, constants, i18n, normalizer, prompts, templates, timezone, utils) stay as-is.

**Viper env var activation (missing):**
- Add `viper.SetEnvPrefix("TEMPUS")` + `viper.AutomaticEnv()` in `config.Load()`. This is a 3-line fix that enables documented but broken `TEMPUS_TIMEZONE`, `TEMPUS_LANGUAGE`, `TEMPUS_OUTPUT_DIR` support.

## Feature Priorities

**Table stakes (must ship -- users expect these):**
- `tempus init` wizard -- auto-detect timezone/language from system, 4 questions max, "next steps" hint. Critical for first-run experience and open-source adoption.
- `--interactive` mode for `create` -- step-by-step prompted input with defaults, step counter ("2/7"), immediate parsing feedback, summary before confirm. The flagship ADHD UX feature.
- Input normalization in `create` -- already works in batch, must work in single create too. Consistency bug.
- Env var support (F3) -- 3-line Viper fix, unblocks CI/CD and power user workflows.

**Differentiators (already implemented, protect these):**
- ADHD alarm profiles, prep time buffers, overwhelm detection, smart duration defaults, auto-emoji, spell-checking, batch from CSV/JSON/YAML. No competitor has these.

**User-specific templates (critical for adoption):**
- `school-event` -- kids' school dates tracking
- `recruiter-meeting` -- work meetings with recruiters
- `travel` -- travel planning events
- These are the user's actual daily workflow. They also serve as excellent onboarding examples for open-source contributors.

**Defer to v2+:**
- Full TUI calendar view (khal/calcurse territory, scope creep)
- CalDAV / Google Calendar sync (offline-first is the value prop)
- Event editing / `tempus edit` (ICS parsing complexity)
- Plugin system (premature abstraction)
- Notification daemon (calendar apps handle this)

## Architecture Approach

**Incremental refactor of `main.go` (3,906 lines, 176 functions) into domain packages:**

The refactor follows a strict bottom-up extraction order based on dependency analysis:

1. **Layer 0 (leaf functions, zero deps):** `nd/spellcheck.go`, `nd/emoji.go`, `cli/output.go`, `parsing/duration.go`
2. **Layer 1 (depends on Layer 0):** `parsing/datetime.go` (unified parser replacing 13 functions), `nd/conflicts.go`, `nd/overwhelm.go`, `nd/preptime.go`
3. **Layer 2:** `parsing/batch.go`, `cli/app.go` (App struct + PersistentPreRunE)
4. **Layer 3 (simple commands):** version, lint, locale, timezone, rrule, config
5. **Layer 4 (complex commands):** create, template, quick, batch

**Key patterns:**
- **App struct with DI** -- single `config.Load()` in `PersistentPreRunE`, `*App` passed to all subcommand constructors via closures. Exactly how `gh`, `kubectl`, `hugo` do it.
- **Command factory functions** -- `newCreateCmd(app *App) *cobra.Command` with `createOptions` struct. `RunE` calls domain packages. Business logic never lives in command files.
- **Wrapper-then-inline migration** -- copy function to target package, leave a thin wrapper in `main.go` calling the new location, remove wrapper in a later PR. Every intermediate commit compiles and passes tests.
- **Unified datetime parser** -- `parsing.Parse(ParseOptions) (ParseResult, error)` replaces all 13 parse functions. Options struct is the variation point, not a strategy pattern.

**10 PRs planned, ~2,460 lines moved.** PR 9 (batch.go) is highest risk due to coupling.

## Critical Pitfalls to Avoid

### 1. Unicode emoji stripping breaks Spanish/Portuguese (B1)

`stripEmoji()` uses `rune > 127` which strips accented characters (e=233, n=241). Fix: use `unicode.Is(unicode.So, r)` to detect actual emoji. Test with all 4 locales (en, es, pt, ga).

**Prevention:** Replace `rune > 127` with `unicode.So` / `unicode.Sk` category checks. Add test cases for "Reunion de equipo", "Cafe con leche", "Preparacao".

### 2. Alarm profile silent data corruption (B3)

`expandAlarmProfiles()` keeps literal `"profile:adhd-triple"` string when profile not found. This passes through to ICS output as invalid alarm data -- user thinks they have alarms but they do not.

**Prevention:** Change signature to `([]string, error)`. Return actionable error: "profile 'adhd-triple' not found. Available: adhd-default, adhd-countdown, medication, single, none".

### 3. Import cycles during monolith split (R1)

Go forbids circular imports. Moving functions to new packages without planning the dependency graph causes compile failures.

**Prevention:** Enforce strict downward dependency flow: `cli/` -> `parsing/` / `nd/` -> `config/` / `i18n/` / `utils/`. Run `go build ./...` after every move. If a lower package needs higher-package behavior, define an interface at the boundary.

### 4. Test coverage regression during refactor

353 tests in `package main` directly call unexported functions. Moving functions to internal packages breaks these tests immediately.

**Prevention:** Copy-then-delete approach -- write new test in target package first, keep old test as integration test, delete old test only after new one exists. CI gate at 79% minimum coverage across `./...`.

### 5. Viper global singleton pollution in tests

`AutomaticEnv()` makes env vars leak between tests. Tests that set `TEMPUS_TIMEZONE` affect other tests.

**Prevention:** Use `t.Setenv()` (auto-cleaned) in all tests. Consider `viper.New()` for isolated instances during config package refactor.

## Bug Fix Specifics

| Bug | Root Cause | Fix Approach | Complexity |
|-----|-----------|--------------|------------|
| **B1: stripEmoji** | `rune > 127` catches accented Latin chars | Replace with `unicode.Is(unicode.So, r)` check. Strip `unicode.Sk` and `unicode.Cf` (ZWJ, variation selectors) in leading sequence. | Low |
| **B2: promptAlarmField i18n** | 10+ Spanish strings hardcoded, bypasses Translator | Add ~15 translation keys to all 4 locale JSON files. Pass `*i18n.Translator` as parameter to `promptAlarmField()`. Also fix locale-aware yes/no responses ("s"/"si" vs "y"/"yes"). | Low-Med |
| **B3: alarm profile error** | `expandAlarmProfiles()` returns literal string on miss | Change signature to `([]string, error)`. Caller decides: abort in single create, warn-and-skip in batch. | Low |
| **B4: cityToIANA error** | Returns `""` for unknown city, caller continues | Return `("", ErrCityNotFound)`. Caller falls back to config timezone with warning, or errors if no fallback. | Low |
| **B5: normalize in create** | Normalization runs in batch path but not create path | Reuse existing `normalizer` package in `runCreate()`. Set `ParseOptions.Normalize = true` when building from create flags. | Low |

## Recommended Build Order

### Phase 1: Bug Fixes (B1-B5)

**Rationale:** Broken things before new things. All 5 bugs are independent, require no structural changes, and deliver immediate user value. B1 and B3 are data-integrity issues that actively harm users.

**Delivers:** Reliable ICS output for Spanish/Portuguese users, working alarm profiles, proper error messages, consistent normalization across create and batch paths.

**Features:** B1-B5 (all bugs)

**Pitfalls addressed:** Unicode emoji detection, alarm silent failure, cityToIANA silent failure

**Research needed:** None -- all fixes are well-understood.

### Phase 2: Foundation Features (F3 env vars + F1 init + practical templates)

**Rationale:** F3 (env vars) is a 3-line Viper fix that unblocks F1. F1 (init wizard) is the first-run experience gate -- new users (and open-source contributors) need guided setup. Practical templates (school-event, recruiter-meeting, travel) ship alongside init as "next steps" examples.

**Delivers:** Working env var overrides, first-run wizard with system auto-detection, templates for the user's actual workflow (recruiter meetings, school dates, travel planning).

**Features:** F3, F1, practical template additions

**Pitfalls addressed:** Viper AutomaticEnv gotchas (SetEnvPrefix, precedence, empty vars)

**Research needed:** Minimal -- init wizard patterns are well-documented (npm init, gh auth login).

### Phase 3: Interactive Mode (F2 --interactive)

**Rationale:** Depends on B2 (i18n in prompts) and B5 (normalization) being fixed. This is the flagship ADHD UX feature. Replace survey/v2 with charmbracelet/huh here.

**Delivers:** Step-by-step event creation with defaults, parsing feedback, step counter, summary confirmation. The "zero-flags" path that makes Tempus accessible to users who forget flag names.

**Features:** F2 (--interactive), survey-to-huh migration

**Pitfalls addressed:** Archived dependency risk (survey), i18n in prompts

**Research needed:** Verify huh Go 1.23 compatibility and accessibility features before committing.

### Phase 4: Polish Features (F4 conflict guidance, F5 prep time, F6 config validation)

**Rationale:** These build on the stable, bug-fixed, interactive-capable base. Conflict resolution guidance (F4) completes the conflict detection story. Prep time customization (F5) and config validation (F6) are polish.

**Delivers:** Actionable conflict suggestions with transition buffer options, customizable prep time, validated config writes.

**Features:** F4, F5, F6

**Pitfalls addressed:** Error vs warn patterns for batch vs single mode

**Research needed:** F4 may benefit from `/gsd:research-phase` for CLI-specific conflict resolution UX patterns.

### Phase 5: Refactor (R1-R6)

**Rationale:** Refactor stable, tested, feature-complete code. Moving code that is still being actively changed creates merge conflicts. The wrapper-then-inline pattern means this can be done across 10 incremental PRs.

**Delivers:** Clean package structure, testable architecture, App struct DI, unified datetime parser, O(n log n) conflict detection, Levenshtein cache.

**Features:** R1 (monolith split), R2 (unified parser), R3 (centralize defaults), R4 (conflict optimization), R5 (Levenshtein cache), R6 (abstract output)

**Pitfalls addressed:** Import cycles, test coverage regression, global state, test file coupling

**Research needed:** R4 (sweep line vs interval tree) has standard patterns -- skip research.

### Phase Ordering Rationale

- **Bugs before features** because broken alarm profiles (B3) and stripped accented characters (B1) actively harm the user's real workflow (Spanish-language recruiter meetings, school events).
- **Init + templates before interactive** because first-run experience gates adoption -- both for the user and for open-source contributors trying Tempus for the first time.
- **Interactive before polish** because `--interactive` is the single highest-leverage UX improvement for ADHD users.
- **Refactor last** because moving stable code is safer than moving in-flux code, and the user gets value from every phase before the refactor.
- **Weave R1-R6 into earlier phases opportunistically** -- when touching a function for a bug fix or feature, extract it to the target package if the extraction is clean.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Go CLI patterns are stable. Cobra/Viper well-documented. huh recommendation based on Charm ecosystem dominance. |
| Features | MEDIUM-HIGH | ADHD UX patterns well-established. Competitor landscape slow-moving. Interactive mode patterns standard (gh, npm, cargo). |
| Architecture | HIGH | Direct codebase analysis of 176 functions. Dependency graph verified. Extraction order based on actual coupling. |
| Pitfalls | HIGH | All pitfalls verified against codebase. Unicode, Viper, import cycle patterns are well-documented Go territory. |

**Overall confidence:** HIGH

### Gaps to Address

- **huh library version and Go 1.23 compatibility** -- verify before Phase 3 with `go list -m -versions github.com/charmbracelet/huh`. Blocking for Phase 3 only.
- **huh accessibility features** -- verify screen reader / high contrast support before committing to it for ND users. If insufficient, fall back to raw bubbletea or keep survey temporarily.
- **Template content for school-event, recruiter-meeting, travel** -- the user knows exactly what fields matter for these. Gather specifics during Phase 2 planning.
- **Open-source contributor onboarding** -- init wizard and templates serve double duty. Consider CONTRIBUTING.md updates in Phase 2.

## Sources

### Primary (HIGH confidence)
- Direct codebase analysis: `main.go` (3,906 lines, 176 functions), `go.mod`, `internal/` packages
- Go `unicode` package documentation (emoji detection via `unicode.So`)
- Viper official documentation (AutomaticEnv, SetEnvPrefix, precedence)
- Cobra subcommand patterns (kubectl, gh, hugo source code)

### Secondary (MEDIUM confidence)
- survey/v2 archival status (confirmed archived 2023, no releases since v2.3.7)
- charmbracelet/huh as successor (Charm ecosystem dominance in Go TUI, 2024-2025)
- ADHD/cognitive accessibility UX patterns (W3C WCAG, ADDitude magazine)
- ICS CLI competitor analysis (khal, calcurse, gcalcli -- slow-moving landscape)

### Tertiary (needs validation)
- huh latest version and Go 1.23 minimum requirement
- huh accessibility features (screen reader, high contrast support)

---
*Research completed: 2026-03-29*
*Ready for roadmap: yes*
