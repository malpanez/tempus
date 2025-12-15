---
title: Tempus Template Creation Guide
language: en
---

[Ver en español](../es/guia-plantillas.md) | [Ver em português](../pt/guia-modelos.md) | [Féach i nGaeilge](../ga/guia-modelos.md)

## What is a Template?

Tempus allows you to add "data-driven" templates for generating ICS events without writing Go code. JSON or YAML files describe the fields, default values, and rules for building the event.

## Supported Directories

1. **User** (automatically loaded):
   - Linux/macOS: `~/.config/tempus/templates/`
   - Windows: `%AppData%\Tempus\templates\`
2. **Custom**: pass `--templates-dir=/path/to/templates` to any `tempus template` subcommand.
3. **Repository**: `internal/templates/json/` contains built-in examples included in the project.

## Required Fields

```yaml
schema_version: 1         # Currently only version 1 is supported
name: identifier          # Unique template name
fields:                   # List of fields Tempus will prompt for
output:                   # How the final event is built
```

### Field Definition (`fields`)

Each element follows the `Field` structure (see `internal/templates/templates.go`):

| Key           | Type     | Description                                   |
|---------------|----------|-----------------------------------------------|
| `key`         | string   | Internal identifier (unique).                 |
| `name`        | string   | Label shown to user.                          |
| `type`        | string   | `text`, `datetime`, `timezone`, `email`, etc. |
| `required`    | bool     | `true` if mandatory.                          |
| `default`     | string   | Default value (optional).                     |
| `description` | string   | Additional help text (optional).              |
| `options`     | []string | Available options (optional).                 |

### Output Block

| Key                | Description                                                             |
|--------------------|-------------------------------------------------------------------------|
| `start_field`      | Field containing start date/time (required).                            |
| `end_field`        | Field with end date/time (optional if `duration_field` is used).        |
| `duration_field`   | Duration field (`45m`, `1h30m`, etc.).                                  |
| `start_tz_field`   | Field with start timezone.                                              |
| `end_tz_field`     | Field with end timezone.                                                |
| `summary_tmpl`     | Text template (mustache) for event summary (required).                  |
| `location_tmpl`    | Template for location.                                                  |
| `description_tmpl` | Template for description.                                               |
| `categories`       | List of categories (`["Health", "Work"]`).                              |
| `priority`         | Numeric priority (1-9).                                                 |

Templates (`*_tmpl`) support:
- `{{field}}`
- `{{slug field}}` (converts to lowercase and replaces spaces with hyphens)
- `{{#field}}...{{/field}}` (renders block only if value exists)

## Examples

### YAML

```yaml
schema_version: 1
name: vaccination
description: Vaccination reminder
fields:
  - key: patient
    name: Patient
    type: text
    required: true
  - key: start_time
    name: Date and time
    type: datetime
    required: true
  - key: timezone
    name: Timezone
    type: timezone
    default: Europe/Madrid
output:
  start_field: start_time
  start_tz_field: timezone
  summary_tmpl: "Vaccination for {{patient}}"
  description_tmpl: "{{#timezone}}Timezone: {{timezone}}{{/timezone}}"
  categories: ["Health"]
```

### JSON

```json
{
  "schema_version": 1,
  "name": "checkup",
  "fields": [
    { "key": "patient", "name": "Patient", "type": "text", "required": true },
    { "key": "start_time", "name": "Start", "type": "datetime", "required": true },
    { "key": "duration", "name": "Duration", "type": "text", "default": "45m" },
    { "key": "timezone", "name": "Timezone", "type": "timezone", "default": "Europe/Madrid" }
  ],
  "output": {
    "start_field": "start_time",
    "duration_field": "duration",
    "start_tz_field": "timezone",
    "summary_tmpl": "Medical checkup: {{patient}}",
    "categories": ["Health", "Personal"]
  }
}
```

## Validation and Usage

1. Save the file in one of the supported directories.
2. Run `tempus template list` to verify the template appears.
3. Use `tempus template create <name>` and fill in the fields.
4. If the file has errors (e.g., `start_field` points to a non-existent key), Tempus will show an error message.

> **Tip**: While developing a template, you can keep it in a working directory and use `tempus template create --templates-dir /path/to/project <name>` to test without moving files.

## Useful Commands

- `tempus template describe <name>` shows fields and output block.
- `tempus template validate` checks templates and reports structure errors.
- `tempus template init my-theme --lang en --format yaml` generates a skeleton ready to edit.
- `tempus locale list` lists embedded languages and custom translations detected on disk.
