---
phase: quick
plan: 260405-ixt
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/cli/batch.go
  - internal/cli/create.go
  - internal/cli/helpers.go
  - internal/cli/batch_test.go
  - internal/cli/coverage_gaps_test.go
  - internal/cli/create_test.go
autonomous: true
requirements: []
must_haves:
  truths:
    - "SonarCloud CPD duplication drops below 3%"
    - "All existing tests pass without modification to assertions"
    - "No exported function signatures change"
  artifacts:
    - path: "internal/cli/helpers.go"
      provides: "Unified parseDurationEnd function"
      contains: "func parseDurationEnd"
    - path: "internal/cli/batch.go"
      provides: "loadBatchFromStructured helper, no inline TZ block in configureBatchEvent"
      contains: "loadBatchFromStructured"
  key_links:
    - from: "internal/cli/batch.go"
      to: "internal/cli/helpers.go"
      via: "parseDurationEnd call (same package)"
      pattern: "parseDurationEnd\\(startTime"
    - from: "internal/cli/batch.go"
      to: "internal/cli/create.go"
      via: "setEventTimezones call (same package)"
      pattern: "setEventTimezones\\(event"
---

<objective>
Eliminate remaining SonarCloud CPD duplication in batch.go and create.go by extracting shared logic and reusing existing helpers.

Purpose: Reduce CPD from 4.8% to below 3% threshold.
Output: Refactored batch.go, create.go, helpers.go with zero behavioral changes.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/cli/batch.go
@internal/cli/create.go
@internal/cli/helpers.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Extract loadBatchFromStructured and replace inline TZ block</name>
  <files>internal/cli/batch.go</files>
  <action>
Fix 1 - In batch.go, add a new unexported function above loadBatchFromJSON:

```go
func loadBatchFromStructured(path string, unmarshal func([]byte, interface{}) error) ([]batchRecord, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}
	var raw []map[string]interface{}
	if err := unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return parseMapsToRecords(raw), nil
}
```

Then replace loadBatchFromJSON body with:
```go
func loadBatchFromJSON(path string) ([]batchRecord, error) {
	return loadBatchFromStructured(path, json.Unmarshal)
}
```

And loadBatchFromYAML body with:
```go
func loadBatchFromYAML(path string) ([]batchRecord, error) {
	return loadBatchFromStructured(path, yaml.Unmarshal)
}
```

Fix 2 - In configureBatchEvent (line ~598-608), replace the inline timezone block:
```go
if startTZ != "" {
    event.SetStartTimezone(startTZ)
}
if endTZ != "" {
    event.SetEndTimezone(endTZ)
} else if startTZ != "" {
    event.SetEndTimezone(startTZ)
}
```
With:
```go
setEventTimezones(event, startTZ, endTZ)
```

setEventTimezones already exists in create.go (same package, no import needed).
  </action>
  <verify>
    <automated>cd /home/malpanez/repos/tempus && go build ./... && go test ./internal/cli/ -run "TestLoadBatch|TestConfigureBatch|TestBatch" -count=1 -timeout 60s</automated>
  </verify>
  <done>loadBatchFromJSON and loadBatchFromYAML are 1-liner wrappers. configureBatchEvent uses setEventTimezones. All batch tests pass.</done>
</task>

<task type="auto">
  <name>Task 2: Unify parseDurationEnd into helpers.go and update all call sites and tests</name>
  <files>internal/cli/helpers.go, internal/cli/create.go, internal/cli/batch.go, internal/cli/batch_test.go, internal/cli/coverage_gaps_test.go, internal/cli/create_test.go</files>
  <action>
Step 1 - Move parseDurationEnd to helpers.go. Add to helpers.go:

```go
func parseDurationEnd(startTime time.Time, durStr string) (time.Time, error) {
	dur, err := calendar.ParseHumanDuration(durStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration %q: %v", durStr, err)
	}
	if dur <= 0 {
		return time.Time{}, fmt.Errorf(testutil.ErrMsgDurationGreaterThanZero)
	}
	return startTime.Add(dur), nil
}
```

This uses the %q format from batch's version. Add `"tempus/internal/calendar"` and `"tempus/internal/testutil"` to helpers.go imports (check if already present first).

Step 2 - In create.go: delete the parseDurationEnd function (lines 202-211). The call site at line 171 (`parseDurationEnd(startTime, durStr)`) keeps the same signature so no change needed there. Verify create.go still compiles.

Step 3 - In batch.go: delete parseBatchDurationEnd function (lines 587-596). Update call site at line 562 from `parseBatchDurationEnd(rec.Duration, startTime)` to `parseDurationEnd(startTime, rec.Duration)` (note: argument order swap).

Step 4 - Update tests. The error message format changed from "invalid duration: %v" to "invalid duration %q: %v" for the create.go path. In test files:
- batch_test.go: rename all `parseBatchDurationEnd` calls to `parseDurationEnd` and swap argument order from `(durStr, startTime)` to `(startTime, durStr)`
- coverage_gaps_test.go: same rename and arg swap for parseBatchDurationEnd calls. The existing parseDurationEnd calls already use the correct signature. Check if any test asserts on the exact error message string "invalid duration: %v" — if so, update to match new format "invalid duration %q: %v".
- create_test.go: check if any assertion depends on the old error format. If tests only check `wantErr` bool, no change needed.

Step 5 - Clean up imports. In batch.go, verify all imports still needed after removing parseBatchDurationEnd (calendar import likely still used elsewhere). In create.go, verify calendar import still needed (used in parseEndTime). In helpers.go, add calendar and testutil imports if not present.
  </action>
  <verify>
    <automated>cd /home/malpanez/repos/tempus && go build ./... && go test ./internal/cli/ -count=1 -timeout 120s && go vet ./internal/cli/</automated>
  </verify>
  <done>parseDurationEnd lives in helpers.go. No parseBatchDurationEnd exists. No parseDurationEnd in create.go. All tests pass. go vet clean.</done>
</task>

</tasks>

<verification>
```bash
cd /home/malpanez/repos/tempus && go build ./... && go test ./... -count=1 -timeout 180s
```
Full test suite passes. No compilation errors.

```bash
cd /home/malpanez/repos/tempus && grep -c "parseBatchDurationEnd" internal/cli/*.go
```
Returns 0 for all files (function fully removed).

```bash
cd /home/malpanez/repos/tempus && grep -c "loadBatchFromStructured" internal/cli/batch.go
```
Returns 3+ (definition + 2 call sites).
</verification>

<success_criteria>
- loadBatchFromJSON and loadBatchFromYAML are 1-liner delegations to loadBatchFromStructured
- configureBatchEvent uses setEventTimezones instead of inline TZ block
- parseDurationEnd exists only in helpers.go, used by both create.go and batch.go paths
- parseBatchDurationEnd no longer exists anywhere
- All tests pass (`go test ./...`)
- No exported API changes
</success_criteria>

<output>
After completion, create `.planning/quick/260405-ixt-fix-remaining-sonarcloud-cpd-duplication/260405-ixt-SUMMARY.md`
</output>
