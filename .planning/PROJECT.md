# Tempus

## What This Is

Tempus es una CLI en Go para generar ficheros ICS (RFC 5545) diseñada específicamente para usuarios neurodivergentes (ADHD, ASD, Dislexia). Permite crear eventos de calendario únicos, en batch desde CSV/JSON/YAML, o a través de lenguaje natural, con features como spell-checking, alarmas adaptadas, detección de conflictos y soporte multiidioma. Los ficheros generados son compatibles con cualquier cliente de calendario estándar.

## Core Value

Un usuario neurodivergente puede crear un evento de calendario correcto con el mínimo fricción posible — sin conocer la sintaxis ICS, sin lidiar con timezones complejos, y con recordatorios configurados automáticamente.

## Requirements

### Validated

<!-- Funcionalidades ya implementadas y que funcionan -->

- ✓ Creación de eventos ICS individuales (`tempus create`) — existing
- ✓ Creación batch desde CSV/JSON/YAML (`tempus batch`) — existing
- ✓ Validación ICS (`tempus lint`) — existing
- ✓ Gestión de configuración (`tempus config`) — existing
- ✓ Templates de eventos (`tempus template`) — existing
- ✓ Explorador de timezones (`tempus timezone`) — existing
- ✓ Helper interactivo para RRULE (`tempus rrule`) — existing
- ✓ Soporte multiidioma (en, es, pt, ga) — existing
- ✓ Spell-checking de títulos y categorías — existing
- ✓ Alarmas configurables con perfiles reutilizables — existing
- ✓ Detección de conflictos y días sobrecargados — existing
- ✓ Prep time auto-generado (`--add-prep-time`) — existing
- ✓ Auto-emoji por categoría — existing
- ✓ Normalización de inputs de fecha/hora (en batch) — existing
- ✓ B1: `stripEmoji()` preserva acentuados — `unicode.Is(unicode.So, r)` — Validated in Phase 1
- ✓ B2: `promptAlarmField()` usa i18n en los 4 locales — Validated in Phase 1
- ✓ B3: Alarm profile no encontrado → error explícito con lista de profiles disponibles — Validated in Phase 1
- ✓ B4: `cityToIANA()` falla con error útil + sugerencia `tempus timezone list --search` — Validated in Phase 1
- ✓ B5: Normalización de inputs en `create` (slash dates, missing zeros) — Validated in Phase 1
- ✓ R6: `printOK`/`printErr` escriben a `var stdout io.Writer` — testable sin capturar stdout global — Validated in Phase 1

### Active

<!-- Mejoras del ciclo actual -->

- [ ] F1: Implementar `tempus init` — wizard de primer uso interactivo
- [ ] F2: Implementar `--interactive` en `create` — modo paso a paso con survey
- [ ] F3: Implementar env vars (`TEMPUS_TIMEZONE`, `TEMPUS_LANGUAGE`, `TEMPUS_OUTPUT_DIR`)
- [ ] F4: Conflict resolution guidance — sugerir horarios ajustados al detectar solapamiento
- [ ] F5: Prep time events personalizables — `prep_time_prefix` en config + `--prep-label` en batch
- [ ] F6: Validación en `config set` — timezone IANA válido, output_dir escribible
- [ ] R1: Refactor `main.go` → packages (`internal/commands/`, `internal/parsing/`, `internal/nd/`, `internal/output/`)
- [ ] R2: Unificar 5 funciones de parsing de fechas en una sola
- [ ] R3: Centralizar smart defaults (duración, alarms, emoji) en un punto
- [ ] R4: Optimizar detección de conflictos O(n²) → O(n log n)
- [ ] R5: Cache para Levenshtein distance en batch processing
- [ ] R6: Abstraer output (`printOK`, `printErr`) para testabilidad

### Out of Scope

- `tempus edit` / `tempus parse` — editar/leer ICS existentes — complejidad alta, no en este ciclo
- Integración con Google Calendar / Outlook API — vendor lock-in, fuera del scope de herramienta offline
- Base de datos de eventos — Tempus es stateless por diseño (genera ficheros)
- GUI / TUI interactiva completa — CLI es el formato primario

## Context

**Codebase actual:**
- `main.go` monolítico de ~3,900 líneas — toda la lógica CLI en un solo fichero
- Tests existentes con ~79% coverage distribuidos en 7 ficheros `main_*_test.go`
- 4 locales embebidos (en, es, pt, ga) en `internal/i18n/`
- Packages internos: `calendar/`, `config/`, `constants/`, `i18n/`, `normalizer/`, `prompts/`, `templates/`, `timezone/`, `utils/`
- El refactor va en paralelo con features: mover código a packages a medida que se toca

**Bugs conocidos afectan usuarios hispanohablantes:**
- `stripEmoji()` con `rune > 127` elimina caracteres acentuados — bug crítico dado el idioma objetivo
- `promptAlarmField()` hardcodeado en español aunque el resto del CLI esté en otro idioma

**Gap principal docs vs implementación:**
- `--interactive` flag existe pero retorna "not yet implemented"
- Env vars documentadas pero Viper no las pickea
- Normalización de inputs (`2025/12/16`, `09:00`) solo funciona en batch, no en `create`

## Constraints

- **Compatibilidad**: No romper la API de comandos existente — flags y outputs deben mantenerse
- **Tests**: Coverage no debe bajar del 79% — cada cambio debe mantener o mejorar tests
- **Go modules**: Mantener `go.mod` limpio — no añadir dependencias innecesarias (`survey` ya está como dep)
- **RFC 5545**: Todos los ICS generados deben seguir siendo válidos — `tempus lint` como gate
- **Sin vendor lock-in**: Herramienta offline, sin deps de APIs externas de calendario

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Refactor en paralelo con features | Evitar una fase de refactor puro sin valor visible para usuarios | — Pending |
| Usar `survey` para `--interactive` | Ya es dependencia existente — sin dep nueva | — Pending |
| Usar Viper `AutomaticEnv()` para env vars | Patrón estándar en el stack Cobra+Viper ya usado | — Pending |
| Mantener main.go como entry point limpio | No eliminar main.go, reducirlo a ~100 líneas de wiring | — Pending |

## Evolution

Este documento evoluciona en las transiciones de fase y milestones.

**Después de cada fase:**
1. ¿Requirements invalidados? → Mover a Out of Scope con razón
2. ¿Requirements validados? → Mover a Validated con referencia de fase
3. ¿Requirements nuevos emergidos? → Añadir a Active
4. ¿Decisiones que registrar? → Añadir a Key Decisions
5. ¿"What This Is" sigue siendo preciso? → Actualizar si ha derivado

**Después de cada milestone:**
1. Revisión completa de todas las secciones
2. Core Value check — ¿sigue siendo la prioridad correcta?
3. Auditar Out of Scope — ¿las razones siguen siendo válidas?

---
*Last updated: 2026-03-30 after Phase 1 completion — 5 bugs fixed, testable output, 79.7% coverage*
