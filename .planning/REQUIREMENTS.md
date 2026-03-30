# Requirements: Tempus

**Defined:** 2026-03-29
**Core Value:** Un usuario neurodivergente puede crear un evento de calendario correcto con el mínimo de fricción — sin conocer la sintaxis ICS, sin lidiar con timezones, y con recordatorios configurados automáticamente.

## v1 Requirements

### Bug Fixes

- [x] **BUG-01**: El usuario con nombre de evento acentuado (é, ñ, ü) no pierde caracteres al usar auto-emoji — `stripEmoji()` usa `unicode.Is(unicode.So, r)` en lugar de `rune > 127`
- [x] **BUG-02**: El prompt interactivo de alarmas respeta el idioma configurado (`--language`) — `promptAlarmField()` usa el sistema i18n existente con ~15 claves nuevas en los 4 locales
- [x] **BUG-03**: Cuando un alarm profile referenciado no existe, el usuario recibe un error claro con sugerencias — `expandAlarmProfiles()` retorna `([]string, error)` en lugar de pasar el literal `"profile:name"`
- [x] **BUG-04**: Cuando el usuario escribe una ciudad no reconocida como timezone, recibe un error útil con instrucciones — `cityToIANA()` retorna error explícito con sugerencia de `tempus timezone list --search <name>`
- [x] **BUG-05**: El usuario puede usar `2025/12/16`, `2025-1-5`, y `09:00` en `tempus create` igual que en batch — `normalizeDateTimeInput()` se aplica en `parseCreateTimes()`

### UX — Primera Experiencia

- [ ] **UX-01**: El usuario nuevo puede ejecutar `tempus init` y quedar configurado con timezone, idioma, directorio de salida y perfil de alarmas — wizard paso a paso con survey
- [x] **UX-02**: `tempus create --interactive` guía al usuario paso a paso con progreso visible ("Paso 2/7") hasta generar el evento — implementado con charmbracelet/huh (reemplaza survey/v2 archivado)
- [x] **UX-03**: En modo batch, cuando `--check-conflicts` detecta eventos que se solapan entre sí dentro del mismo fichero, el usuario ve exactamente qué eventos colisionan (nombres, horas) y cuánto tiempo solapan — facilitando decidir cuál mover antes de importar al calendario
- [ ] **UX-04**: El nombre del evento de prep time es personalizable mediante `prep_time_prefix` en config y `--prep-label` en batch — default: "Preparation"

### Configuración

- [x] **CONF-01**: El usuario puede configurar timezone, idioma y directorio con variables de entorno (`TEMPUS_TIMEZONE`, `TEMPUS_LANGUAGE`, `TEMPUS_OUTPUT_DIR`) — Viper `SetEnvPrefix("TEMPUS")` + `AutomaticEnv()`
- [x] **CONF-02**: `tempus config set timezone Europe/Madrid` valida que el identificador IANA existe antes de guardar
- [x] **CONF-03**: `tempus config set output_dir /ruta` valida que el directorio existe y es escribible antes de guardar

### Templates Prácticos

- [ ] **TMPL-01**: `tempus batch template school-event` genera plantilla para fechas escolares (trimestres, eventos, vacaciones) con campos: nombre, fecha inicio, fecha fin, categoría, notas
- [ ] **TMPL-02**: `tempus batch template recruiter-meeting` genera plantilla para entrevistas y llamadas con recruiters — incluye prep time automático, alarmas triple ADHD, campo de empresa y rol
- [ ] **TMPL-03**: `tempus batch template travel-day` genera plantilla para días de viaje con soporte multi-timezone (origen/destino), vuelo, alojamiento, actividades

### Refactor — Mantenibilidad

- [x] **REF-01**: El código de cada comando Cobra vive en `internal/cli/<command>.go` con un `App` struct que provee config/translator via `PersistentPreRunE` — `main.go` queda en ~100 líneas
- [x] **REF-02**: Las 13 funciones de parsing de fechas se unifican en `internal/parsing.Parse(ParseOptions)` — mismo comportamiento, un punto de entrada
- [ ] **REF-03**: Las features neurodivergentes (spellcheck, conflictos, prep time, emoji) viven en `internal/nd/` — extraídas con tests migrados sin pérdida de cobertura
- [ ] **REF-04**: La detección de conflictos usa algoritmo sweep-line O(n log n) en lugar de O(n²)
- [ ] **REF-05**: El spell checking en batch precalcula la matriz de distancias una vez, no por cada registro (~100x menos comparaciones)
- [x] **REF-06**: `printOK`, `printErr` y similares aceptan `io.Writer` como parámetro — testables sin capturar stdout global

## v2 Requirements

### Edición de ICS existentes

- **EDIT-01**: Usuario puede editar un evento en un fichero ICS existente (`tempus edit`)
- **EDIT-02**: Usuario puede ver los eventos en un ICS existente (`tempus parse`)

### Mejoras de Dependencias

- **DEP-01**: Migrar de survey/v2 (archivado 2023) a charmbracelet/huh para prompts interactivos
- **DEP-02**: Evaluar reemplazo de olebedev/when (posiblemente no mantenido) para NLP en `tempus quick`

### Discoverability

- **DISC-01**: GitHub Topics configurados (`ics`, `calendar`, `adhd`, `neurodivergent`, `cli`, `golang`)
- **DISC-02**: README con más ejemplos de flujos completos (desde `tempus create` hasta importar en Google Calendar)
- **DISC-03**: Ejemplos en `examples/` usando sintaxis de separadores documentada (`|`, `;`)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Google Calendar / Outlook API | Vendor lock-in; Tempus es offline por diseño |
| Base de datos de eventos | Stateless por diseño — genera ficheros, no los gestiona |
| GUI / TUI completa | CLI es el formato primario; complejidad alta por poco valor añadido |
| Mobile app | Fuera del scope de herramienta CLI |
| `tempus edit` / `tempus parse` | Complejidad de ICS parsing alta; no en este ciclo |
| NLP completo multi-idioma | olebedev/when solo soporta inglés bien; scope creep para este ciclo |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| BUG-01 | Phase 1: Bug Fixes & Test Infrastructure | Complete |
| BUG-02 | Phase 1: Bug Fixes & Test Infrastructure | Complete |
| BUG-03 | Phase 1: Bug Fixes & Test Infrastructure | Complete |
| BUG-04 | Phase 1: Bug Fixes & Test Infrastructure | Complete |
| BUG-05 | Phase 1: Bug Fixes & Test Infrastructure | Complete |
| REF-02 | Phase 1: Bug Fixes & Test Infrastructure | Complete |
| REF-06 | Phase 1: Bug Fixes & Test Infrastructure | Complete |
| CONF-01 | Phase 2: First-Run Experience | Complete |
| CONF-02 | Phase 2: First-Run Experience | Complete |
| CONF-03 | Phase 2: First-Run Experience | Complete |
| UX-01 | Phase 2: First-Run Experience | Pending |
| TMPL-01 | Phase 2: First-Run Experience | Pending |
| TMPL-02 | Phase 2: First-Run Experience | Pending |
| TMPL-03 | Phase 2: First-Run Experience | Pending |
| UX-02 | Phase 3: Interactive Mode & CLI Structure | Complete |
| REF-01 | Phase 3: Interactive Mode & CLI Structure | Complete |
| UX-03 | Phase 4: UX Polish | Complete |
| UX-04 | Phase 4: UX Polish | Pending |
| REF-03 | Phase 5: ND Extraction & Performance | Pending |
| REF-04 | Phase 5: ND Extraction & Performance | Pending |
| REF-05 | Phase 5: ND Extraction & Performance | Pending |

**Coverage:**
- v1 requirements: 21 total
- Mapped to phases: 21
- Unmapped: 0

---
*Requirements defined: 2026-03-29*
*Last updated: 2026-03-29 after roadmap creation*
