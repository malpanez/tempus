---
phase: quick
plan: 260403-skg
type: execute
wave: 1
depends_on: []
files_modified:
  - sonar-project.properties
autonomous: true
must_haves:
  truths:
    - "SonarCloud CPD duplication drops below 3%"
    - "Coverage reporting still works via profile.out"
  artifacts:
    - path: "sonar-project.properties"
      provides: "SonarCloud config without test file analysis"
      contains: "sonar.go.coverage.reportPaths"
  key_links:
    - from: "sonar-project.properties"
      to: "SonarCloud analysis"
      via: "sonar.cpd.exclusions now covers all _test.go files since they are only seen as sources"
      pattern: "sonar\\.cpd\\.exclusions=\\*\\*\\/\\*_test\\.go"
---

<objective>
Fix SonarCloud CPD duplication from 5.1% to below 3% by removing sonar.tests and sonar.test.inclusions from sonar-project.properties.

Purpose: When sonar.tests=. is set, SonarCloud analyzes test files as "test code" in a separate pass where sonar.cpd.exclusions does not apply. Removing sonar.tests makes all _test.go files only visible as source files, where sonar.cpd.exclusions=**/*_test.go already excludes them from CPD. Coverage still works because sonar.go.coverage.reportPaths reads profile.out directly — it does not depend on sonar.tests.

Output: Updated sonar-project.properties
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@sonar-project.properties
</context>

<tasks>

<task type="auto">
  <name>Task 1: Remove sonar.tests config to fix CPD exclusion scope</name>
  <files>sonar-project.properties</files>
  <action>
Remove these two lines from sonar-project.properties:
- `sonar.tests=.`
- `sonar.test.inclusions=**/*_test.go`

Keep everything else unchanged. The `sonar.cpd.exclusions=**/*_test.go` line already exists and will now correctly exclude test files from CPD since they will only be seen as source files (where cpd.exclusions applies), not test files (where it does not).

The `sonar.go.coverage.reportPaths=.coverage/profile.out` line handles coverage independently — it reads the Go coverage profile directly and does NOT depend on sonar.tests being set.

Also keep `sonar.exclusions=**/*_test.go,**/vendor/**,...` which excludes test files from source analysis rules (bugs, code smells) but NOT from the file listing that cpd.exclusions filters against.

Wait — if sonar.exclusions already excludes *_test.go from sources, and we remove sonar.tests, then test files are excluded from BOTH source and test analysis. That means sonar.cpd.exclusions is redundant but harmless. The net effect: test files are invisible to SonarCloud entirely, CPD drops to 0% for test code, coverage still works via profile.out.
  </action>
  <verify>
    <automated>grep -c "sonar.tests" sonar-project.properties | grep -q "^0$" && grep -q "sonar.go.coverage.reportPaths" sonar-project.properties && echo "PASS" || echo "FAIL"</automated>
  </verify>
  <done>sonar-project.properties has no sonar.tests or sonar.test.inclusions lines; sonar.go.coverage.reportPaths is preserved; sonar.cpd.exclusions is preserved as defense-in-depth</done>
</task>

</tasks>

<verification>
- `grep "sonar.tests" sonar-project.properties` returns nothing
- `grep "sonar.test.inclusions" sonar-project.properties` returns nothing
- `grep "sonar.go.coverage.reportPaths" sonar-project.properties` returns the profile.out path
- `grep "sonar.cpd.exclusions" sonar-project.properties` returns the test file exclusion (defense-in-depth)
</verification>

<success_criteria>
- sonar-project.properties no longer contains sonar.tests or sonar.test.inclusions
- Coverage path preserved
- CPD exclusions preserved
- After PR merge, SonarCloud CPD should report below 3% (verified on next CI run)
</success_criteria>

<output>
After completion, create `.planning/quick/260403-skg-fix-sonarcloud-cpd-5-1-to-below-3-by-fix/260403-skg-SUMMARY.md`
</output>
