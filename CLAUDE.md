<!-- GSD:project-start source:PROJECT.md -->
## Project

**Tempus**

Tempus es una CLI en Go para generar ficheros ICS (RFC 5545) diseñada específicamente para usuarios neurodivergentes (ADHD, ASD, Dislexia). Permite crear eventos de calendario únicos, en batch desde CSV/JSON/YAML, o a través de lenguaje natural, con features como spell-checking, alarmas adaptadas, detección de conflictos y soporte multiidioma. Los ficheros generados son compatibles con cualquier cliente de calendario estándar.

**Core Value:** Un usuario neurodivergente puede crear un evento de calendario correcto con el mínimo fricción posible — sin conocer la sintaxis ICS, sin lidiar con timezones complejos, y con recordatorios configurados automáticamente.

### Constraints

- **Compatibilidad**: No romper la API de comandos existente — flags y outputs deben mantenerse
- **Tests**: Coverage no debe bajar del 79% — cada cambio debe mantener o mejorar tests
- **Go modules**: Mantener `go.mod` limpio — no añadir dependencias innecesarias (`survey` ya está como dep)
- **RFC 5545**: Todos los ICS generados deben seguir siendo válidos — `tempus lint` como gate
- **Sin vendor lock-in**: Herramienta offline, sin deps de APIs externas de calendario
<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->
## Technology Stack

## Go CLI Project Structure (2025)
- Do NOT use `cobra-cli init` scaffolding -- it generates a `cmd/` package (not `cmd/appname/`) with a flat structure that does not scale well.
- Do NOT create a `pkg/` directory -- this is for libraries, not CLIs.
- Do NOT split too granularly -- a package with one 30-line file is over-engineering. Group by domain, not by file count.
## Cobra Command Packaging Patterns
### Pattern: Command Factory Functions
- Testable: you can call `runCreate()` directly in tests with fake options
- No globals: config is injected, not imported from a singleton
- Flag parsing is co-located with the command definition
- `gh` CLI uses this exact pattern at scale (~100 commands)
### Pattern: Root Command Assembly
### Pattern: Subcommand Groups (not needed yet)
### Migration without breaking public API
## i18n Patterns in Go CLIs
### Current state in Tempus
### Standard patterns in the ecosystem
- Message catalogs in TOML/JSON/YAML
- Pluralization support (CLDR rules)
- `go generate` for extracting translation keys
- Used by Hugo, Mattermost
- Tempus does NOT need this -- the current homegrown system is sufficient for a CLI with ~50-80 message keys
- Low-level: message catalogs, plural forms, number/date formatting
- Good for complex localization (currency, dates, collation)
- Overkill for Tempus's use case (translating CLI messages)
- Simple, explicit, easy to audit
- No magic, no code generation
- Limitation: no pluralization, no interpolation beyond `fmt.Sprintf`
- Perfectly adequate for a CLI tool
### Recommendation for Tempus
## Viper + Env Vars Pattern
### The problem
### The correct pattern
- `SetEnvPrefix("TEMPUS")` makes Viper look for `TEMPUS_*` env vars
- `AutomaticEnv()` enables automatic env var binding
- `SetEnvKeyReplacer` maps config keys like `output_dir` to `TEMPUS_OUTPUT_DIR`
- Env vars override config file values but NOT command-line flags
### What to expose as env vars
| Env Var | Config Key | Type |
|---------|-----------|------|
| `TEMPUS_TIMEZONE` | `timezone` | string |
| `TEMPUS_LANGUAGE` | `language` | string |
| `TEMPUS_OUTPUT_DIR` | `output_dir` | string |
| `TEMPUS_DATE_FORMAT` | `date_format` | string |
| `TEMPUS_TIME_FORMAT` | `time_format` | string |
## Interactive Prompts (Survey/alternatives)
### AlecAivazis/survey/v2 -- current dependency
- **Status:** The repository was archived in late 2023. No new releases since v2.3.7 (the version in `go.mod`).
- **Works fine for now:** It compiles, it functions, terminal support is adequate.
- **Risk:** No security patches, no bug fixes, no new terminal support. The `golang.org/x/term` dependency is pinned to an old version (visible in go.mod: `v0.0.0-20210927222741`).
- **Current usage in Tempus:** Minimal -- only one `survey.Confirm` call in `main.go` line 168. The `internal/prompts/` package uses raw `bufio.Scanner`, not survey at all.
### Alternatives
- Active development (Charm ecosystem: bubbletea, lipgloss, huh)
- Modern terminal UI, accessibility features
- Form-based API that maps well to Tempus's `--interactive` needs
- Theming support (relevant for neurodivergent-friendly high contrast)
- Example:
- Full TUI framework (Elm architecture)
- Overkill for forms/prompts, but the foundation `huh` builds on
- Use directly only if you need custom interactive widgets
- Still works, minimal maintenance
- Less capable than both survey and huh
- No reason to choose this over huh
### Recommendation
- survey is archived with known terminal issues
- huh is actively maintained by the Charm team (well-funded, Go-focused)
- huh's form API maps directly to Tempus's need: multi-field input for event creation
- Accessibility features align with neurodivergent target audience
- The Charm ecosystem (lipgloss for styling, huh for forms) is the clear community standard for Go CLI UX in 2025
- Only one `survey.Confirm` call exists -- replace with `huh.NewConfirm()`
- The `internal/prompts/` package already abstracts user input -- add huh behind this interface
- Remove `survey/v2` from `go.mod` after migration
# After migration:
## Confidence Levels
| Section | Confidence | Rationale |
|---------|-----------|-----------|
| Go CLI Project Structure | **HIGH** | Patterns are well-established (gh, kubectl, cobra docs). Go project layout has been stable since 2019. Verified against actual codebase structure. |
| Cobra Command Packaging | **HIGH** | Standard patterns used by gh CLI (open source, auditable). Factory function pattern is the dominant approach in every major Go CLI. |
| i18n Patterns | **MEDIUM** | Recommendation to keep current system is sound. Could not verify latest go-i18n features via web. The "don't migrate" recommendation holds regardless. |
| Viper + Env Vars | **HIGH** | Viper's API is stable and well-documented. Confirmed no `AutomaticEnv` call exists in codebase (verified via grep). Pattern is standard Cobra+Viper. |
| Interactive Prompts (survey) | **MEDIUM** | Survey archival is well-known (happened 2023). Huh recommendation is based on Charm ecosystem dominance in Go TUI space. Could not verify latest huh release version via web -- confirm before adding dependency. |
### Verification needed before implementation
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
