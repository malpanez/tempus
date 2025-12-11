---
title: Guia para criar modelos do Tempus
language: pt
---

[Voltar à documentação em inglês](../examples.md)

## O que é um modelo

Tempus suporta modelos "data-driven" para gerar eventos ICS configuráveis sem escrever código Go. Os arquivos JSON ou YAML descrevem campos, valores padrão e como montar o evento final.

## Diretórios suportados

1. **Usuário** (carregado automaticamente):
   - Linux/macOS: `~/.config/tempus/templates/`
   - Windows: `%AppData%\Tempus\templates\`
2. **Personalizado**: informe `--templates-dir=/caminho/para/modelos` em qualquer subcomando `tempus template`.
3. **Repositório**: `internal/templates/json/` contém exemplos distribuídos com o projeto.

## Campos obrigatórios

```yaml
schema_version: 1        # Por enquanto, apenas a versão 1 é aceita
name: identificador      # Nome único do modelo
fields:                  # Lista de perguntas exibidas ao usuário
output:                  # Configuração do evento resultante
```

### Definição de `fields`

| Chave        | Tipo     | Descrição                                   |
|--------------|----------|---------------------------------------------|
| `key`        | string   | Identificador interno (único).              |
| `name`       | string   | Rótulo exibido ao usuário.                  |
| `type`       | string   | `text`, `datetime`, `timezone`, `email`, etc.|
| `required`   | bool     | `true` se o campo é obrigatório.            |
| `default`    | string   | Valor padrão (opcional).                    |
| `description`| string   | Texto de ajuda (opcional).                  |
| `options`    | []string | Lista de opções (opcional).                 |

### Bloco `output`

| Chave              | Descrição                                                            |
|--------------------|----------------------------------------------------------------------|
| `start_field`      | Campo que contém a data/hora inicial (obrigatório).                  |
| `end_field`        | Campo para data/hora final (opcional se houver `duration_field`).    |
| `duration_field`   | Campo de duração (`45m`, `1h30m`, etc.).                             |
| `start_tz_field`   | Campo com o fuso horário de início.                                  |
| `end_tz_field`     | Campo com o fuso horário de término.                                 |
| `summary_tmpl`     | Template de texto (mustache) para o resumo (obrigatório).            |
| `location_tmpl`    | Template para localização.                                           |
| `description_tmpl` | Template para descrição.                                             |
| `categories`       | Lista de categorias (`["Saúde", "Pessoal"]`).                      |
| `priority`         | Prioridade numérica (1-9).                                           |

Os templates (`*_tmpl`) aceitam:
- `{{campo}}`
- `{{slug campo}}` (minúsculas com hífens)
- `{{#campo}}...{{/campo}}` (renderiza o bloco somente se houver valor)

## Exemplos

### YAML

```yaml
schema_version: 1
name: vacina
description: Lembrete de vacinação
fields:
  - key: paciente
    name: Paciente
    type: text
    required: true
  - key: start_time
    name: Data e hora
    type: datetime
    required: true
  - key: timezone
    name: Fuso horário
    type: timezone
    default: Europe/Lisbon
output:
  start_field: start_time
  start_tz_field: timezone
  summary_tmpl: "Vacina para {{paciente}}"
  description_tmpl: "{{#timezone}}Zona: {{timezone}}{{/timezone}}"
  categories: ["Saúde"]
```

### JSON

```json
{
  "schema_version": 1,
  "name": "checkup",
  "fields": [
    { "key": "patient", "name": "Paciente", "type": "text", "required": true },
    { "key": "start_time", "name": "Início", "type": "datetime", "required": true },
    { "key": "duration", "name": "Duração", "type": "text", "default": "45m" },
    { "key": "timezone", "name": "Fuso horário", "type": "timezone", "default": "Europe/Lisbon" }
  ],
  "output": {
    "start_field": "start_time",
    "duration_field": "duration",
    "start_tz_field": "timezone",
    "summary_tmpl": "Consulta: {{patient}}",
    "categories": ["Saúde", "Pessoal"]
  }
}
```

## Validação e uso

1. Salve o arquivo em um dos diretórios suportados.
2. Execute `tempus template list` e confirme que o modelo aparece.
3. Use `tempus template create <nome>` para gerar o evento.
4. Caso exista algum erro (por exemplo, campo referenciado inexistente), Tempus exibirá uma mensagem explicando o problema.

> Dica: Enquanto estiver ajustando um modelo, utilize `tempus template create --templates-dir /caminho/do/projeto <nome>` para testar sem mover arquivos definitivos.

## Comandos úteis

- `tempus template describe <nome>` mostra os campos e o bloco `output`.
- `tempus template validate` revisa os arquivos e aponta erros de estrutura.
- `tempus template init meu-modelo --lang pt --format yaml` gera um esqueleto pronto para edição.
- `tempus locale list` lista os idiomas embutidos e os overrides detectados em disco.
