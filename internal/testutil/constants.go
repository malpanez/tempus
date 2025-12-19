package testutil

// Common test constants to avoid duplication across test files
const (
	// Common timezone names used in tests
	TZAmericaNewYork     = "America/New_York"
	TZEuropeMadrid       = "Europe/Madrid"
	TZEuropeLondon       = "Europe/London"
	TZEuropeDublin       = "Europe/Dublin"
	TZEuropeParis        = "Europe/Paris"
	TZEuropeBerlin       = "Europe/Berlin"
	TZAtlanticCanary     = "Atlantic/Canary"
	TZAfricaCeuta        = "Africa/Ceuta"
	TZAmericaSaoPaulo    = "America/Sao_Paulo"
	TZAmericaCampoGrande = "America/Campo_Grande"
	TZAsiaTokyо          = "Asia/Tokyo"
	TZInvalid            = "Invalid/Timezone"

	// Test email addresses
	EmailAlice = "alice@example.com"
	EmailBob   = "bob@example.com"

	// Test person names
	NameJohnDoe = "John Doe"

	// Common date strings
	Date20251115 = "2025-11-15"
	Date20251201 = "2025-12-01"
	Date20251215 = "2025-12-15"
	Date20251220 = "2025-12-20"
	Date20251227 = "2025-12-27"
	Date20250501 = "2025-05-01"
	Date20250503 = "2025-05-03"
	Date20250615 = "2025-06-15"

	// Common datetime strings
	DateTime20251115_1000 = "2025-11-15 10:00:00"
	DateTime20251201_0800 = "2025-12-01 08:00"
	DateTime20251201_0900 = "2025-12-01 09:00"
	DateTime20251201_1000 = "2025-12-01 10:00"
	DateTime20251201_1130 = "2025-12-01 11:30"
	DateTime20251201_1400 = "2025-12-01 14:00"
	DateTime20251201_1430 = "2025-12-01 14:30"
	DateTime20250501_1000 = "2025-05-01 10:00"
	DateTime20250501_1100 = "2025-05-01 11:00"
	DateTime20250501_1400 = "2025-05-01 14:00"
	DateTime20250615_1000 = "2025-06-15 10:00"
	DateTime20251216_1030 = "2025-12-16 10:30"

	// Common test names/labels
	TestNameEmptyString = "empty string"
	TestNameEmptySlice  = "empty slice"
	TestNameEmptyFile   = "empty file"
	TestNameWithSpaces  = "with spaces"
	TestNameFullDatetime = "full datetime"
	TestNameDateOnly    = "date only"
	TestNameWithTimezone = "with timezone"

	// Common event titles
	EventTitleTeamMeeting = "Team Meeting"
	EventTitleBadMeeting  = "Bad Meeting"
	EventTitleTestEvent   = "Test Event"
	EventTitleHelloWorld  = "Hello World"
	EventTitleEvent1      = "Event 1"
	EventTitleEvent2      = "Event 2"

	// Common airline/flight data
	AirlineAmerican = "American Airlines"

	// Common template/file names
	TemplateCustomEvent = "custom-event"
	TemplateDateTest    = "date-test"
	TemplateHelloWorld  = "hello-world"
	TemplatesDir        = "templates-dir"
	FilenameEventICS    = "event.ics"
	FilenameEventsCSV   = "events.csv"
	FilenameEventsTXT   = "events.txt"
	FilenameDataCSV     = "data.csv"
	FilenameDataTXT     = "data.txt"
	FilenameTestCSV     = "test.csv"
	FilenameTestJSON    = "test.json"

	// Common template strings
	TemplatePlaceholderTitle = "{{title}}"

	// Common timezone abbreviations
	TZAbbrevCETCEST = "CET/CEST"

	// Common location strings
	LocationNewYork = "New York"

	// Common description strings
	DescriptionMeetingNotes = "Meeting notes"
	DescriptionDeepWork     = "deep work"

	// Common recurrence rules
	RRuleDaily5Count = "FREQ=DAILY;COUNT=5"

	// Common error message formats
	ErrMsgEventIsNil                = "event is nil"
	ErrMsgInvalidStartTime          = "invalid start time"
	ErrMsgBadTime                   = "bad-time"
	ErrMsgNotADate                  = "not-a-date"
	ErrMsgExpectedErrorGotNil       = "expected error, got nil"
	ErrMsgUnexpectedError           = "unexpected error: %v"
	ErrMsgFailedToGetTemplate       = "failed to get template: %v"
	ErrMsgGeneratorFailed           = "generator failed: %v"
	ErrMsgLocationMismatch          = "location = %q, want %q"
	ErrMsgEndDateAfterStart         = "end date must be on or after start date"
	ErrMsgDurationGreaterThanZero   = "duration must be greater than zero"
	ErrMsgRowFormat                 = "row %d: %w"
	ErrMsgFailedToWriteTestJSON     = "Failed to write test JSON: %v"
	ErrMsgLoadJSONDirError          = "LoadJSONDir() unexpected error: %v"
	ErrMsgFailedToGetLoadedZone     = "Failed to get loaded zone: %v"
	ErrMsgLoadJSONFileError         = "loadJSONFile() unexpected error: %v"
	ErrMsgFailedToWriteFile         = "failed to write file: %v\n"
	ErrMsgCreatedFormat             = "Created: %s\n"
	ErrMsgFailedToCreateTestFile    = "failed to create test file: %v"
	ErrMsgFailedToWriteTestFile     = "failed to write test file: %v"
	ErrMsgBuildEventFromBatchError  = "buildEventFromBatch() error = %v"
	ErrMsgUseMismatch               = "Use = %q, want %q"
	ErrMsgRequiresInternalStructures = "requires internal template structures"

	// Common test zone names
	TestZone1 = "Test/Zone1"

	// Common test strings
	TestStringEmptyString = "Empty string"

	// Common date formats (for test validation)
	TestDateFormatRFC         = "Mon, 02 Jan 2006 15:04 MST"
	TestDateFormatDate        = "2006-01-02"
	TestDateFormatDateTime    = "2006-01-02 15:04"
	TestDateFormatDateTimeHMS = "2006-01-02 15:04:05"

	// Common format strings
	FormatKeyValuePair = "%s: %s\n"

	// Common test strings and labels
	StringHelloWorld           = "Hello World"
	StringLine1Line2Escaped    = "Line1\\nLine2"
	StringLine1Line2           = "Line1\nLine2"
	StringABC                  = "A\nB\nC"
	StringWithPlus             = "with plus"
	StringZeroDuration         = "zero duration"
	StringInvalidDate          = "invalid date"
	StringInvalidStartTime     = "invalid start time"
	StringRequiredFieldMissing = "required field missing"
	StringOptionB              = "Option B"
	StringOptionC              = "Option C"
	StringConferenceRoomA      = "Conference Room A"
	StringFocusBlock           = "focus-block"
	StringNonExistentTemplate  = "non-existent template"

	// Common ICS fields
	ICSStatusConfirmed         = "STATUS:CONFIRMED"
	ICSBeginVTimezone          = "BEGIN:VTIMEZONE"
	ICSRRuleWeeklyMonday       = "FREQ=WEEKLY;BYDAY=MO"
	ICSPromptStartTime         = "Start Time (YYYY-MM-DD HH:MM)"
	ICSMustacheNameTemplate    = "{{#name}}Hello {{name}}{{/name}}"
	ICSCountryUnitedStates     = "United States"

	// Alarm-specific test strings
	AlarmExpected1Alarm              = "Expected 1 alarm, got %d"
	AlarmParseError                  = "ParseAlarmsFromString(%q) error: %v"
	AlarmTriggerDurationMismatch     = "TriggerDuration = %v, want -15m"
	AlarmTriggerIsRelativeShouldTrue = "TriggerIsRelative should be true"
	AlarmResultMismatch              = "Result = %v, want %v"
	AlarmErrorFormat                 = "Error: %v"
	AlarmInvalidStartTimeFormat      = "invalid start time: %w"
	AlarmInvalidDurationFormat       = "invalid duration: %w"
)
