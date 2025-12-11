---
title: Guía para crear plantillas de Tempus
language: es
---

[Volver a la documentación en inglés](../examples.md)

## Qué es una plantilla

Tempus permite añadir plantillas "data-driven" para generar eventos ICS sin escribir código Go. Los archivos JSON o YAML describen los campos, valores por defecto y las reglas para construir el evento.

## Directorios soportados

1. **Usuario** (carga automática):
   - Linux/macOS: `~/.config/tempus/templates/`
   - Windows: `%AppData%\Tempus\templates\`
2. **Personalizado**: pasa `--templates-dir=/ruta/a/plantillas` a cualquier subcomando `tempus template`.
3. **Repositorio**: `internal/templates/json/` contiene ejemplos incluidos en el proyecto.

## Campos obligatorios

```yaml
schema_version: 1         # Actualmente solo se acepta la versión 1
name: identificador       # Nombre único de la plantilla
fields:                   # Lista de campos que Tempus preguntará
output:                   # Cómo se construye el evento final
```

### Definición de campos (`fields`)

Cada elemento sigue la estructura `Field` (ver `internal/templates/templates.go`):

| Clave        | Tipo     | Descripción                                  |
|--------------|----------|----------------------------------------------|
| `key`        | string   | Identificador interno (único).               |
| `name`       | string   | Etiqueta mostrada al usuario.                |
| `type`       | string   | `text`, `datetime`, `timezone`, `email`, etc.|
| `required`   | bool     | `true` si es obligatorio.                    |
| `default`    | string   | Valor por defecto (opcional).                |
| `description`| string   | Ayuda adicional (opcional).                  |
| `options`    | []string | Opciones disponibles (opcional).             |

### Bloque `output`

| Clave             | Descripción                                                        |
|-------------------|--------------------------------------------------------------------|
| `start_field`     | Campo que contiene la fecha/hora de inicio (obligatorio).          |
| `end_field`       | Campo con fecha/hora de fin (opcional si se usa `duration_field`).|
| `duration_field`  | Campo de duración (`45m`, `1h30m`, etc.).                          |
| `start_tz_field`  | Campo con el huso horario de inicio.                               |
| `end_tz_field`    | Campo con el huso horario de fin.                                  |
| `summary_tmpl`    | Plantilla de texto (mustache) para el resumen (obligatorio).       |
| `location_tmpl`   | Plantilla para la ubicación.                                       |
| `description_tmpl`| Plantilla para la descripción.                                     |
| `categories`      | Lista de categorías (`["Salud", "Trabajo"]`).                   |
| `priority`        | Prioridad numérica (1-9).                                          |

Las plantillas (`*_tmpl`) aceptan:
- `{{campo}}`
- `{{slug campo}}` (convierte a minúsculas y reemplaza espacios por guiones)
- `{{#campo}}...{{/campo}}` (renderiza el bloque solo si el valor existe)

## Ejemplos

### YAML

```yaml
schema_version: 1
name: vacuna
description: Recordatorio de vacunación
fields:
  - key: paciente
    name: Paciente
    type: text
    required: true
  - key: start_time
    name: Fecha y hora
    type: datetime
    required: true
  - key: timezone
    name: Zona horaria
    type: timezone
    default: Europe/Madrid
output:
  start_field: start_time
  start_tz_field: timezone
  summary_tmpl: "Vacuna para {{paciente}}"
  description_tmpl: "{{#timezone}}Zona: {{timezone}}{{/timezone}}"
  categories: ["Salud"]
```

### JSON

```json
{
  "schema_version": 1,
  "name": "checkup",
  "fields": [
    { "key": "patient", "name": "Paciente", "type": "text", "required": true },
    { "key": "start_time", "name": "Inicio", "type": "datetime", "required": true },
    { "key": "duration", "name": "Duración", "type": "text", "default": "45m" },
    { "key": "timezone", "name": "Zona horaria", "type": "timezone", "default": "Europe/Madrid" }
  ],
  "output": {
    "start_field": "start_time",
    "duration_field": "duration",
    "start_tz_field": "timezone",
    "summary_tmpl": "Chequeo médico: {{patient}}",
    "categories": ["Salud", "Personal"]
  }
}
```

## Validación y uso

1. Guarda el archivo en uno de los directorios soportados.
2. Ejecuta `tempus template list` para verificar que la plantilla aparece.
3. Usa `tempus template create <nombre>` y completa los campos.
4. Si el archivo tiene errores (por ejemplo, `start_field` apunta a una clave inexistente), Tempus mostrará un mensaje indicando el problema.

> Consejo: mientras desarrollas una plantilla, puedes mantenerla en un directorio de trabajo y usar `tempus template create --templates-dir /ruta/al/proyecto <nombre>` para probar sin mover archivos definitivos.

## Comandos útiles

- `tempus template describe <nombre>` muestra los campos y el bloque `output`.
- `tempus template validate` revisa las plantillas y reporta errores de estructura.
- `tempus template init mi-tema --lang es --format yaml` genera un esqueleto listo para editar.
- `tempus locale list` lista los idiomas embebidos y las traducciones personalizadas detectadas en disco.
