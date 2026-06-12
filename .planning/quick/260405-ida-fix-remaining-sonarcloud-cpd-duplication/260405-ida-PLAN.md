---
phase: quick
plan: 260405-ida
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/cli/batch.go
  - internal/nd/nd.go
autonomous: true
requirements: []
must_haves:
  truths:
    - "SonarCloud CPD duplication reduced by eliminating duplicated struct literal blocks"
    - "All existing tests pass with no behavioral changes"
    - "No exported function signatures changed"
  artifacts:
    - path: "internal/cli/batch.go"
      provides: "parseMapsToRecords helper eliminates loadBatchFromJSON/loadBatchFromYAML duplication"
      contains: "func parseMapsToRecords"
    - path: "internal/nd/nd.go"
      provides: "newGeneratedEvent helper eliminates createTransitionEventIfNeeded/createPrepEventIfNeeded duplication"
      contains: "func newGeneratedEvent"
  key_links: []
---

<objective>
Extract shared helpers in batch.go and nd.go to eliminate CPD-detected duplication blocks.

Purpose: Reduce SonarCloud CPD from 5.1% to below 3% on PR #3.
Output: Two refactored files with no behavioral changes.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/cli/batch.go
@internal/nd/nd.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Extract parseMapsToRecords in batch.go</name>
  <files>internal/cli/batch.go</files>
  <action>
Extract the 14-line record-mapping loop (lines 418-436 / 454-472) into a new unexported function:

```go
func parseMapsToRecords(raw []map[string]interface{}) []batchRecord {
    records := make([]batchRecord, 0, len(raw))
    for _, item := range raw {
        rec := batchRecord{
            Summary:     ValueAsString(item["summary"]),
            Start:       ValueAsString(item["start"]),
            End:         ValueAsString(item["end"]),
            Duration:    ValueAsString(item["duration"]),
            StartTZ:     ValueAsString(item["start_tz"]),
            EndTZ:       ValueAsString(item["end_tz"]),
            Location:    ValueAsString(item["location"]),
            Description: ValueAsString(item["description"]),
            RRule:       ValueAsString(item["rrule"]),
            AllDay:      ValueAsBool(item["all_day"]),
            ExDates:     ValueAsStringSlice(item["exdate"]),
            Categories:  ValueAsStringSlice(item["categories"]),
            Alarms:      ValueAsAlarmSlice(item["alarms"]),
        }
        records = append(records, rec)
    }
    return records
}
```

Replace the duplicated loop in both loadBatchFromJSON and loadBatchFromYAML with:
```go
return parseMapsToRecords(raw), nil
```

No new imports needed. No exported signatures change.
  </action>
  <verify>
    <automated>cd /home/malpanez/repos/tempus && go build ./... && go test ./internal/cli/ -count=1 -timeout 60s</automated>
  </verify>
  <done>loadBatchFromJSON and loadBatchFromYAML each call parseMapsToRecords instead of duplicating the mapping loop. All batch tests pass.</done>
</task>

<task type="auto">
  <name>Task 2: Extract newGeneratedEvent in nd.go</name>
  <files>internal/nd/nd.go</files>
  <action>
Extract the common Event struct fields into a new unexported function:

```go
func newGeneratedEvent(summary string, start, end time.Time, ev calendar.Event, categories []string) *calendar.Event {
    return &calendar.Event{
        UID:        GenerateUID(),
        Summary:    summary,
        StartTime:  start,
        EndTime:    end,
        StartTZ:    ev.StartTZ,
        EndTZ:      ev.EndTZ,
        AllDay:     false,
        Categories: categories,
        Status:     "CONFIRMED",
        Created:    time.Now().UTC(),
        LastMod:    time.Now().UTC(),
    }
}
```

Rewrite createTransitionEventIfNeeded to:
```go
func createTransitionEventIfNeeded(ev calendar.Event) *calendar.Event {
    if !needsFocusTransition(ev.Summary) {
        return nil
    }
    return newGeneratedEvent(
        "\U0001f504 Transition: "+StripEmoji(ev.Summary),
        ev.EndTime,
        ev.EndTime.Add(5*time.Minute),
        ev,
        []string{"Transition"},
    )
}
```

Rewrite createPrepEventIfNeeded to:
```go
func createPrepEventIfNeeded(ev calendar.Event, prepLabel string) *calendar.Event {
    duration, description := determinePrepTime(ev.Summary)
    if duration == 0 {
        return nil
    }
    if description == "Preparation" && prepLabel != "" {
        description = prepLabel
    }
    return newGeneratedEvent(
        "\u23f0 "+description+": "+StripEmoji(ev.Summary),
        ev.StartTime.Add(-duration),
        ev.StartTime,
        ev,
        []string{"Preparation"},
    )
}
```

No new imports needed. No exported signatures change.
  </action>
  <verify>
    <automated>cd /home/malpanez/repos/tempus && go build ./... && go test ./internal/nd/ -count=1 -timeout 60s</automated>
  </verify>
  <done>createTransitionEventIfNeeded and createPrepEventIfNeeded both use newGeneratedEvent. All nd tests pass.</done>
</task>

</tasks>

<verification>
```bash
cd /home/malpanez/repos/tempus && go build ./... && go test ./... -count=1 -timeout 120s
```
All tests pass. `go vet ./...` clean.
</verification>

<success_criteria>
- parseMapsToRecords exists in batch.go, called by both JSON and YAML loaders
- newGeneratedEvent exists in nd.go, called by both transition and prep event creators
- `go build ./...` succeeds
- `go test ./...` passes with no failures
- No exported function signatures changed
</success_criteria>

<output>
After completion, create `.planning/quick/260405-ida-fix-remaining-sonarcloud-cpd-duplication/260405-ida-SUMMARY.md`
</output>
