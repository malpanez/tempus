# Phase 1: Bug Fixes & Test Infrastructure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-29
**Phase:** 01-bug-fixes-test-infrastructure
**Areas discussed:** Error handling style, REF-02 scope, REF-06 io.Writer pattern

---

## Error Handling Style

### Alarm profile not found (BUG-03)

| Option | Description | Selected |
|--------|-------------|----------|
| Error fatal + lista de profiles | Abortar con mensaje que lista profiles disponibles | ✓ |
| Warning + continuar sin alarma | Generar ICS sin alarma y avisar | |
| Error fatal en create, warning en batch | Comportamiento diferente según contexto | |

### Ciudad desconocida en timezone (BUG-04)

| Option | Description | Selected |
|--------|-------------|----------|
| Error fatal + sugerencia | Abortar con sugerencia de `tempus timezone list --search` | ✓ |
| Warning + usar timezone por defecto | Usar config timezone y avisar | |
| Igual que alarm profile | Consistencia total con BUG-03 | |

---

## REF-02 Scope en Phase 1

| Option | Description | Selected |
|--------|-------------|----------|
| Solo BUG-05: aplicar normalización en create | 2-line change, unificación completa a Phase 3 | ✓ |
| Unificación completa en Phase 1 | Crear internal/parsing.Parse(ParseOptions) ahora | |

---

## REF-06 io.Writer Pattern

| Option | Description | Selected |
|--------|-------------|----------|
| Variable de paquete (var stdout io.Writer) | Mínimo cambio, 0 refactor de callers, tests sobreescriben | ✓ |
| Parámetro en cada función | printOK(w io.Writer, ...) — actualizar ~40 callers | |
