---
title: Treoir Cruthú Teimpléad Tempus
language: ga
---

[English version](../en/template-guide.md) | [Ver en español](../es/guia-plantillas.md) | [Ver em português](../pt/guia-modelos.md)

## Cad é Teimpléad?

Ceadaíonn Tempus duit teimpléid "sonraí-thiomáinte" a chur leis chun imeachtaí ICS a ghiniúint gan cód Go a scríobh. Déanann comhaid JSON nó YAML cur síos ar na réimsí, na luachanna réamhshocraithe agus na rialacha chun an t-imeacht a thógáil.

## Eolairí Tacaithe

1. **Úsáideoir** (luchtaithe go huathoibríoch):
   - Linux/macOS: `~/.config/tempus/templates/`
   - Windows: `%AppData%\Tempus\templates\`
2. **Saincheaptha**: cuir `--templates-dir=/cosán/go/teimpléid` le haon fho-ordú `tempus template`.
3. **Stór**: Tá samplaí ionsuite sa tionscadal in `internal/templates/json/`.

## Réimsí Riachtanacha

```yaml
schema_version: 1         # Ní thacaítear ach le leagan 1 faoi láthair
name: aitheantóir         # Ainm uathúil an teimpléid
fields:                   # Liosta réimsí a iarrfaidh Tempus
output:                   # Conas a thógtar an t-imeacht deiridh
```

### Sainmhíniú Réimse (`fields`)

Leanann gach eilimint an struchtúr `Field` (féach `internal/templates/templates.go`):

| Eochair      | Cineál   | Cur Síos                                     |
|--------------|----------|----------------------------------------------|
| `key`        | string   | Aitheantóir inmheánach (uathúil).            |
| `name`       | string   | Lipéad a thaispeántar don úsáideoir.         |
| `type`       | string   | `text`, `datetime`, `timezone`, `email`, etc.|
| `required`   | bool     | `true` má tá sé riachtanach.                 |
| `default`    | string   | Luach réamhshocraithe (roghnach).            |
| `description`| string   | Téacs cabhrach breise (roghnach).            |
| `options`    | []string | Roghanna atá ar fáil (roghnach).             |

### Bloc `output`

| Eochair           | Cur Síos                                                           |
|-------------------|--------------------------------------------------------------------|
| `start_field`     | Réimse ina bhfuil dáta/am tosaigh (riachtanach).                   |
| `end_field`       | Réimse le dáta/am deiridh (roghnach má úsáidtear `duration_field`).|
| `duration_field`  | Réimse faide (`45m`, `1h30m`, etc.).                               |
| `start_tz_field`  | Réimse le crios ama tosaigh.                                       |
| `end_tz_field`    | Réimse le crios ama deiridh.                                       |
| `summary_tmpl`    | Teimpléad téacs (mustache) don achoimre (riachtanach).             |
| `location_tmpl`   | Teimpléad don suíomh.                                              |
| `description_tmpl`| Teimpléad don tuairisc.                                            |
| `categories`      | Liosta catagóirí (`["Sláinte", "Obair"]`).                         |
| `priority`        | Tosaíocht uimhriúil (1-9).                                         |

Tacaíonn teimpléid (`*_tmpl`) le:
- `{{réimse}}`
- `{{slug réimse}}` (tiontaíonn go cás íochtair agus cuireann fleiscíní in ionad spásanna)
- `{{#réimse}}...{{/réimse}}` (ní léiríonn an bloc ach amháin má tá luach ann)

## Samplaí

### YAML

```yaml
schema_version: 1
name: vacsaín
description: Meabhrúchán vacsaínithe
fields:
  - key: othar
    name: Othar
    type: text
    required: true
  - key: start_time
    name: Dáta agus am
    type: datetime
    required: true
  - key: timezone
    name: Crios ama
    type: timezone
    default: Europe/Dublin
output:
  start_field: start_time
  start_tz_field: timezone
  summary_tmpl: "Vacsaín do {{othar}}"
  description_tmpl: "{{#timezone}}Crios ama: {{timezone}}{{/timezone}}"
  categories: ["Sláinte"]
```

### JSON

```json
{
  "schema_version": 1,
  "name": "seiceáil",
  "fields": [
    { "key": "othar", "name": "Othar", "type": "text", "required": true },
    { "key": "start_time", "name": "Tosach", "type": "datetime", "required": true },
    { "key": "duration", "name": "Fad", "type": "text", "default": "45m" },
    { "key": "timezone", "name": "Crios ama", "type": "timezone", "default": "Europe/Dublin" }
  ],
  "output": {
    "start_field": "start_time",
    "duration_field": "duration",
    "start_tz_field": "timezone",
    "summary_tmpl": "Seiceáil leighis: {{othar}}",
    "categories": ["Sláinte", "Pearsanta"]
  }
}
```

## Bailíochtú agus Úsáid

1. Sábháil an comhad i gceann de na heolairí tacaithe.
2. Rith `tempus template list` chun a fhíorú go bhfeictear an teimpléad.
3. Úsáid `tempus template create <ainm>` agus líon isteach na réimsí.
4. Má tá earráidí sa chomhad (m.sh. pointe `start_field` chuig eochair nach ann), taispeánfaidh Tempus teachtaireacht earráide.

> Leid: Agus tú ag forbairt teimpléid, is féidir leat é a choinneáil in eolaire oibre agus `tempus template create --templates-dir /cosán/go/tionscadal <ainm>` a úsáid le tástáil gan comhaid a bhogadh.

## Orduithe Úsáideacha

- `tempus template describe <ainm>` taispeánann na réimsí agus an bloc `output`.
- `tempus template validate` seiceálann teimpléid agus tuairiscíonn earráidí struchtúir.
- `tempus template init mo-théama --lang ga --format yaml` gineann creatlach réidh le heagartha.
- `tempus locale list` liostálann teangacha leabaithe agus aistriúcháin saincheaptha a braitheadh ar an diosca.
