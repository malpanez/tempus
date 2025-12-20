package templates

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/i18n"
	"tempus/internal/testutil"
)

// Template represents an event template (built-in or data-driven wrapper)
type Template struct {
	Name        string
	Description string
	Fields      []Field
	Generator   func(data map[string]string, translator *i18n.Translator) (*calendar.Event, error)
}

// Field represents a template field
type Field struct {
	Key         string   `json:"key" yaml:"key"`
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"` // text, datetime, timezone, email, number, etc.
	Required    bool     `json:"required" yaml:"required"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Options     []string `json:"options,omitempty" yaml:"options,omitempty"`
}

// TemplateManager manages event templates
type TemplateManager struct {
	templates   map[string]*Template
	ddTemplates map[string]DataDrivenTemplate // parsed DD templates (for filename/metadata)
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates:   make(map[string]*Template),
		ddTemplates: make(map[string]DataDrivenTemplate),
	}
	// Register built-in templates as fallback
	tm.registerBuiltinTemplates()
	return tm
}

// GetTemplate returns a template by name
func (tm *TemplateManager) GetTemplate(name string) (*Template, error) {
	if t, ok := tm.templates[name]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("template not found: %s", name)
}

// ListTemplates returns all available templates
func (tm *TemplateManager) ListTemplates() map[string]*Template {
	return tm.templates
}

// GenerateEvent generates an event from a template with given data
func (tm *TemplateManager) GenerateEvent(templateName string, data map[string]string, translator *i18n.Translator) (*calendar.Event, error) {
	t, err := tm.GetTemplate(templateName)
	if err != nil {
		return nil, err
	}
	// Validate required fields
	for _, f := range t.Fields {
		if f.Required {
			if v := strings.TrimSpace(data[f.Key]); v == "" {
				return nil, fmt.Errorf("required field missing: %s", f.Key)
			}
		}
	}
	return t.Generator(data, translator)
}

// ----------------------
// Data-driven templates
// ----------------------

// LoadDDDir loads JSON DD templates from a directory and registers them.
// This matches the usage in main.go where errors are non-fatal/ignored.
func (tm *TemplateManager) LoadDDDir(dir string) {
	m, err := LoadDDTemplates(dir)
	if err != nil {
		return // silent: caller treats it as optional
	}
	for _, dd := range m {
		tm.RegisterDDTemplate(dd)
	}
}

// RegisterDDTemplate converts and registers a data-driven template.
func (tm *TemplateManager) RegisterDDTemplate(dd DataDrivenTemplate) {
	// Convert fields
	fields := make([]Field, 0, len(dd.Fields))
	for _, f := range dd.Fields {
		fields = append(fields, Field{
			Key:         f.Key,
			Name:        f.Name,
			Type:        f.Type,
			Required:    f.Required,
			Default:     f.Default,
			Description: f.Description,
			Options:     f.Options,
		})
	}

	// Generator uses the data-driven renderer with duration support.
	gen := func(data map[string]string, tr *i18n.Translator) (*calendar.Event, error) {
		ddCopy := dd // avoid capturing pointer to loop var
		return tm.renderDDToEvent(&ddCopy, data, tr)
	}

	// Register into both maps
	tmpl := &Template{
		Name:        dd.Name,
		Description: dd.Description,
		Fields:      fields,
		Generator:   gen,
	}
	tm.templates[dd.Name] = tmpl
	tm.ddTemplates[dd.Name] = dd
}

// FilenameTemplate returns a dd filename template if present.
func (tm *TemplateManager) FilenameTemplate(name string) (string, bool) {
	if dd, ok := tm.ddTemplates[name]; ok && strings.TrimSpace(dd.FilenameTemplate) != "" {
		return dd.FilenameTemplate, true
	}
	return "", false
}

// DataTemplate returns the raw data-driven template definition if available.
func (tm *TemplateManager) DataTemplate(name string) (DataDrivenTemplate, bool) {
	dd, ok := tm.ddTemplates[name]
	return dd, ok
}

// GuessStartDate tries to extract a YYYY-MM-DD date from values based on the dd StartField.
func (tm *TemplateManager) GuessStartDate(name string, values map[string]string) (string, bool) {
	dd, ok := tm.ddTemplates[name]
	if !ok {
		return "", false
	}
	field := strings.TrimSpace(dd.Output.StartField)
	if field == "" {
		return "", false
	}
	v := strings.TrimSpace(values[field])
	if v == "" {
		return "", false
	}
	// Expect "YYYY-MM-DD" or "YYYY-MM-DD HH:MM"
	if len(v) >= 10 {
		return v[:10], true
	}
	return "", false
}

// ----------------------
// Built-in templates
// ----------------------

func (tm *TemplateManager) registerBuiltinTemplates() {
	// Flight template
	tm.templates["flight"] = &Template{
		Name:        "flight",
		Description: "Flight itinerary",
		Fields: []Field{
			{Key: "flight_number", Name: "Flight Number", Type: "text", Required: true},
			{Key: "from", Name: "From", Type: "text", Required: true},
			{Key: "to", Name: "To", Type: "text", Required: true},
			{Key: "departure_time", Name: "Departure Time (YYYY-MM-DD HH:MM)", Type: "datetime", Required: true},
			{Key: "arrival_time", Name: "Arrival Time (YYYY-MM-DD HH:MM)", Type: "datetime", Required: true},
			{Key: "departure_tz", Name: "Departure Timezone", Type: "timezone", Required: false, Default: "UTC"},
			{Key: "arrival_tz", Name: "Arrival Timezone", Type: "timezone", Required: false, Default: "UTC"},
			{Key: "airline", Name: "Airline", Type: "text", Required: false},
			{Key: "seat", Name: "Seat", Type: "text", Required: false},
			{Key: "gate", Name: "Gate", Type: "text", Required: false},
		},
		Generator: generateFlightEvent,
	}

	// Meeting template
	tm.templates["meeting"] = &Template{
		Name:        "meeting",
		Description: "Business meeting",
		Fields: []Field{
			{Key: "title", Name: "Meeting Title", Type: "text", Required: true},
			{Key: "start_time", Name: "Start Time (YYYY-MM-DD HH:MM)", Type: "datetime", Required: true},
			{Key: "end_time", Name: "End Time (YYYY-MM-DD HH:MM)", Type: "datetime", Required: false},
			{Key: "duration", Name: "Duration (e.g. 30, 45m, 1h, 1h30m)", Type: "text", Required: false, Default: "60m"},
			{Key: "location", Name: "Location", Type: "text", Required: false},
			{Key: "timezone", Name: "Timezone", Type: "timezone", Required: false, Default: "UTC"},
			{Key: "attendees", Name: "Attendees (comma-separated emails)", Type: "email", Required: false},
			{Key: "agenda", Name: "Agenda", Type: "text", Required: false},
			{Key: "meeting_url", Name: "Meeting URL", Type: "text", Required: false},
		},
		Generator: generateMeetingEvent,
	}

	// Holiday (all-day) template
	tm.templates["holiday"] = &Template{
		Name:        "holiday",
		Description: "Holiday/vacation period",
		Fields: []Field{
			{Key: "destination", Name: "Destination", Type: "text", Required: true},
			{Key: "start_date", Name: "Start Date (YYYY-MM-DD)", Type: "datetime", Required: true},
			{Key: "end_date", Name: "End Date (YYYY-MM-DD)", Type: "datetime", Required: true},
			{Key: "timezone", Name: "Timezone", Type: "timezone", Required: false, Default: "UTC"},
			{Key: "accommodation", Name: "Accommodation", Type: "text", Required: false},
			{Key: "notes", Name: "Notes", Type: "text", Required: false},
		},
		Generator: generateHolidayEvent,
	}

	// Focus Block template (ADHD-friendly)
	tm.templates["focus-block"] = &Template{
		Name:        "focus-block",
		Description: "Deep focus time block (ADHD-friendly)",
		Fields: []Field{
			{Key: "task", Name: "Task/Project", Type: "text", Required: true},
			{Key: "start_time", Name: "Start Time (YYYY-MM-DD HH:MM)", Type: "datetime", Required: true},
			{Key: "duration", Name: "Duration (e.g. 30, 45m, 1h, 2h)", Type: "text", Required: false, Default: "90m"},
			{Key: "timezone", Name: "Timezone", Type: "timezone", Required: false, Default: "UTC"},
			{Key: "notes", Name: "Notes/Subtasks", Type: "text", Required: false},
		},
		Generator: generateFocusBlockEvent,
	}

	// Medication Reminder template (ADHD-friendly)
	tm.templates["medication"] = &Template{
		Name:        "medication",
		Description: "Medication reminder (ADHD-friendly)",
		Fields: []Field{
			{Key: "medication_name", Name: "Medication Name", Type: "text", Required: true},
			{Key: "time", Name: "Time (YYYY-MM-DD HH:MM)", Type: "datetime", Required: true},
			{Key: "dosage", Name: "Dosage (e.g. 20mg, 1 pill)", Type: "text", Required: true},
			{Key: "timezone", Name: "Timezone", Type: "timezone", Required: false, Default: "UTC"},
			{Key: "instructions", Name: "Instructions (e.g. with food)", Type: "text", Required: false},
			{Key: "recurrence", Name: "Recurrence (RRULE)", Type: "text", Required: false},
		},
		Generator: generateMedicationEvent,
	}

	// Appointment template (ADHD-friendly with travel time)
	tm.templates["appointment"] = &Template{
		Name:        "appointment",
		Description: "Appointment with travel time (ADHD-friendly)",
		Fields: []Field{
			{Key: "title", Name: "Appointment Type (e.g. Doctor, Therapy)", Type: "text", Required: true},
			{Key: "provider", Name: "Provider Name", Type: "text", Required: false},
			{Key: "start_time", Name: "Appointment Time (YYYY-MM-DD HH:MM)", Type: "datetime", Required: true},
			{Key: "duration", Name: "Duration (e.g. 30, 45m, 1h)", Type: "text", Required: false, Default: "30m"},
			{Key: "travel_time", Name: "Travel Time Before (e.g. 15m, 30m)", Type: "text", Required: false, Default: "15m"},
			{Key: "location", Name: "Location/Address", Type: "text", Required: false},
			{Key: "timezone", Name: "Timezone", Type: "timezone", Required: false, Default: "UTC"},
			{Key: "notes", Name: "Notes", Type: "text", Required: false},
		},
		Generator: generateAppointmentEvent,
	}

	// Transition time template (ADHD-friendly)
	tm.templates["transition"] = &Template{
		Name:        "transition",
		Description: "Transition/buffer time (ADHD-friendly)",
		Fields: []Field{
			{Key: "from_activity", Name: "From Activity", Type: "text", Required: true},
			{Key: "to_activity", Name: "To Activity", Type: "text", Required: true},
			{Key: "start_time", Name: "Start Time (YYYY-MM-DD HH:MM)", Type: "datetime", Required: true},
			{Key: "duration", Name: "Duration (e.g. 10m, 15m, 30m)", Type: "text", Required: false, Default: "15m"},
			{Key: "timezone", Name: "Timezone", Type: "timezone", Required: false, Default: "UTC"},
		},
		Generator: generateTransitionEvent,
	}

	// Deadline template (ADHD-friendly with countdown reminders)
	tm.templates["deadline"] = &Template{
		Name:        "deadline",
		Description: "Deadline with countdown reminders (ADHD-friendly)",
		Fields: []Field{
			{Key: "task", Name: "Task/Project", Type: "text", Required: true},
			{Key: "due_date", Name: "Due Date (YYYY-MM-DD)", Type: "datetime", Required: true},
			{Key: "priority", Name: "Priority (1-9, 1=highest)", Type: "number", Required: false, Default: "5"},
			{Key: "timezone", Name: "Timezone", Type: "timezone", Required: false, Default: "UTC"},
			{Key: "notes", Name: "Notes", Type: "text", Required: false},
		},
		Generator: generateDeadlineEvent,
	}
}

// ----- Built-in generators -----

func generateFlightEvent(data map[string]string, translator *i18n.Translator) (*calendar.Event, error) {
	flightNumber := data["flight_number"]
	from := data["from"]
	to := data["to"]

	// Parse times
	departureTime, err := time.Parse("2006-01-02 15:04", data["departure_time"])
	if err != nil {
		return nil, fmt.Errorf("invalid departure time: %w", err)
	}
	arrivalTime, err := time.Parse("2006-01-02 15:04", data["arrival_time"])
	if err != nil {
		return nil, fmt.Errorf("invalid arrival time: %w", err)
	}

	// Create event
	summary := translator.T(i18n.KeyFlightTemplate, fmt.Sprintf("%s: %s -> %s", flightNumber, from, to))
	event := calendar.NewEvent(summary, departureTime, arrivalTime)

	// Set timezones
	if depTZ := data["departure_tz"]; depTZ != "" {
		event.SetStartTimezone(depTZ)
	}
	if arrTZ := data["arrival_tz"]; arrTZ != "" {
		event.SetEndTimezone(arrTZ)
	}

	// Build description
	var description string
	description += fmt.Sprintf("%s: %s\n", translator.T(i18n.KeyFlightNumber), flightNumber)
	description += fmt.Sprintf("%s: %s\n", translator.T(i18n.KeyFlightFrom), from)
	description += fmt.Sprintf("%s: %s\n", translator.T(i18n.KeyFlightTo), to)

	if airline := data["airline"]; airline != "" {
		description += fmt.Sprintf("%s: %s\n", translator.T("airline"), airline)
	}
	if seat := data["seat"]; seat != "" {
		description += fmt.Sprintf("%s: %s\n", translator.T("seat"), seat)
	}
	if gate := data["gate"]; gate != "" {
		description += fmt.Sprintf("%s: %s\n", translator.T("gate"), gate)
	}

	event.Description = description
	event.AddCategory("Travel")
	event.AddCategory("Flight")

	return event, nil
}

func generateMeetingEvent(data map[string]string, translator *i18n.Translator) (*calendar.Event, error) {
	title := data["title"]

	// Parse start
	startTime, err := time.Parse("2006-01-02 15:04", data["start_time"])
	if err != nil {
		return nil, fmt.Errorf(testutil.ErrMsgInvalidStartTimeFormat, err)
	}

	// End can be explicit or computed from duration
	var endTime time.Time
	if endStr := strings.TrimSpace(data["end_time"]); endStr != "" {
		endTime, err = time.Parse("2006-01-02 15:04", endStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end time: %w", err)
		}
	} else if durStr := strings.TrimSpace(data["duration"]); durStr != "" {
		dur, derr := parseHumanDuration(durStr)
		if derr != nil {
			return nil, fmt.Errorf(testutil.ErrMsgInvalidDurationFormat, derr)
		}
		endTime = startTime.Add(dur)
	} else {
		endTime = startTime.Add(time.Hour)
	}

	if !endTime.After(startTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	// Create event
	summary := translator.T(i18n.KeyMeetingTemplate, title)
	event := calendar.NewEvent(summary, startTime, endTime)

	// Set timezone
	if tz := data["timezone"]; tz != "" {
		event.SetTimezone(tz)
	}

	// Set location
	if location := data["location"]; location != "" {
		event.Location = location
	}

	// Add attendees
	if attendees := data["attendees"]; attendees != "" {
		for _, attendee := range splitAndTrim(attendees, ",") {
			event.AddAttendee(attendee)
		}
	}

	// Build description
	var description string
	if agenda := data["agenda"]; agenda != "" {
		description += fmt.Sprintf("%s: %s\n", translator.T(i18n.KeyMeetingTopic), agenda)
	}
	if meetingURL := data["meeting_url"]; meetingURL != "" {
		description += fmt.Sprintf("%s: %s\n", translator.T("meeting_url"), meetingURL)
	}

	event.Description = description
	event.AddCategory("Meeting")
	event.AddCategory("Work")

	return event, nil
}

func generateHolidayEvent(data map[string]string, translator *i18n.Translator) (*calendar.Event, error) {
	destination := data["destination"]

	// Parse dates
	startDate, err := time.Parse("2006-01-02", data["start_date"])
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", data["end_date"])
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Create all-day event (DTEND exclusive)
	summary := translator.T(i18n.KeyHolidayTemplate, destination)
	event := calendar.NewEvent(summary, startDate, endDate.AddDate(0, 0, 1))
	event.AllDay = true

	// Set timezone
	if tz := data["timezone"]; tz != "" {
		event.SetTimezone(tz)
	}

	event.Location = destination

	// Build description
	var description string
	description += fmt.Sprintf("%s: %s\n", translator.T(i18n.KeyHolidayDestination), destination)

	if accommodation := data["accommodation"]; accommodation != "" {
		description += fmt.Sprintf("%s: %s\n", translator.T("accommodation"), accommodation)
	}
	if notes := data["notes"]; notes != "" {
		description += fmt.Sprintf("%s: %s\n", translator.T("notes"), notes)
	}

	event.Description = description
	event.AddCategory("Vacation")
	event.AddCategory("Personal")

	return event, nil
}

// Helper function to split and trim strings
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// ----- ADHD-friendly template generators -----

func generateFocusBlockEvent(data map[string]string, _ *i18n.Translator) (*calendar.Event, error) {
	task := data["task"]

	// Parse start time
	startTime, err := time.Parse("2006-01-02 15:04", data["start_time"])
	if err != nil {
		return nil, fmt.Errorf(testutil.ErrMsgInvalidStartTimeFormat, err)
	}

	// Parse duration
	durStr := strings.TrimSpace(data["duration"])
	if durStr == "" {
		durStr = "90m"
	}
	dur, err := parseHumanDuration(durStr)
	if err != nil {
		return nil, fmt.Errorf(testutil.ErrMsgInvalidDurationFormat, err)
	}
	endTime := startTime.Add(dur)

	// Create event
	summary := fmt.Sprintf("🎯 Focus: %s", task)
	event := calendar.NewEvent(summary, startTime, endTime)

	// Set timezone
	if tz := data["timezone"]; tz != "" {
		event.SetTimezone(tz)
	}

	// Build description
	var description string
	description += "Deep focus block - Do Not Disturb recommended\n\n"
	if notes := data["notes"]; notes != "" {
		description += fmt.Sprintf("Notes:\n%s\n", notes)
	}
	description += "\n💡 Tips:\n"
	description += "- Close unnecessary apps and tabs\n"
	description += "- Put phone on Do Not Disturb\n"
	description += "- Have water and snacks ready\n"
	description += "- Set clear goal for this session\n"

	event.Description = description
	event.AddCategory("Focus")
	event.AddCategory("Work")

	// Add reminders: 5min before to prepare, at start
	event.Alarms = append(event.Alarms,
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       "Focus session starting in 5 minutes - prepare",
			TriggerIsRelative: true,
			TriggerDuration:   -5 * time.Minute,
		},
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       "Focus session starting now",
			TriggerIsRelative: true,
			TriggerDuration:   0,
		},
	)

	return event, nil
}

func generateMedicationEvent(data map[string]string, _ *i18n.Translator) (*calendar.Event, error) {
	medName := data["medication_name"]
	dosage := data["dosage"]

	// Parse time
	medTime, err := time.Parse("2006-01-02 15:04", data["time"])
	if err != nil {
		return nil, fmt.Errorf("invalid time: %w", err)
	}

	// Medication reminders are short events (15min window)
	endTime := medTime.Add(15 * time.Minute)

	// Create event
	summary := fmt.Sprintf("💊 %s - %s", medName, dosage)
	event := calendar.NewEvent(summary, medTime, endTime)

	// Set timezone
	if tz := data["timezone"]; tz != "" {
		event.SetTimezone(tz)
	}

	// Build description
	var description string
	description += fmt.Sprintf("Medication: %s\n", medName)
	description += fmt.Sprintf("Dosage: %s\n", dosage)
	if instructions := data["instructions"]; instructions != "" {
		description += fmt.Sprintf("Instructions: %s\n", instructions)
	}

	event.Description = description
	event.AddCategory("Health")
	event.AddCategory("Medication")

	// Set recurrence if provided
	if rrule := data["recurrence"]; rrule != "" {
		event.RRule = rrule
	}

	// Multiple reminders for medication (critical!)
	event.Alarms = append(event.Alarms,
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       fmt.Sprintf("Take %s in 10 minutes", medName),
			TriggerIsRelative: true,
			TriggerDuration:   -10 * time.Minute,
		},
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       fmt.Sprintf("Take %s NOW - %s", medName, dosage),
			TriggerIsRelative: true,
			TriggerDuration:   0,
		},
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       fmt.Sprintf("Did you take %s?", medName),
			TriggerIsRelative: true,
			TriggerDuration:   5 * time.Minute,
		},
	)

	return event, nil
}

func generateAppointmentEvent(data map[string]string, _ *i18n.Translator) (*calendar.Event, error) {
	title := data["title"]

	// Parse appointment time
	apptTime, err := time.Parse("2006-01-02 15:04", data["start_time"])
	if err != nil {
		return nil, fmt.Errorf(testutil.ErrMsgInvalidStartTimeFormat, err)
	}

	// Parse duration
	durStr := strings.TrimSpace(data["duration"])
	if durStr == "" {
		durStr = "30m"
	}
	dur, err := parseHumanDuration(durStr)
	if err != nil {
		return nil, fmt.Errorf(testutil.ErrMsgInvalidDurationFormat, err)
	}
	endTime := apptTime.Add(dur)

	// Create event
	summary := title
	if provider := data["provider"]; provider != "" {
		summary = fmt.Sprintf("%s - %s", title, provider)
	}
	event := calendar.NewEvent(summary, apptTime, endTime)

	// Set timezone
	if tz := data["timezone"]; tz != "" {
		event.SetTimezone(tz)
	}

	// Set location
	if location := data["location"]; location != "" {
		event.Location = location
	}

	// Build description
	var description string
	if provider := data["provider"]; provider != "" {
		description += fmt.Sprintf("Provider: %s\n", provider)
	}
	if location := data["location"]; location != "" {
		description += fmt.Sprintf("Location: %s\n", location)
	}
	if notes := data["notes"]; notes != "" {
		description += fmt.Sprintf("\nNotes:\n%s\n", notes)
	}

	event.Description = description
	event.AddCategory("Appointment")

	// Add reminders including travel time
	travelStr := strings.TrimSpace(data["travel_time"])
	if travelStr == "" {
		travelStr = "15m"
	}
	travelTime, err := parseHumanDuration(travelStr)
	if err == nil && travelTime > 0 {
		// Reminder for when to leave (travel time before)
		event.Alarms = append(event.Alarms, calendar.Alarm{
			Action:            "DISPLAY",
			Description:       "Time to leave!",
			TriggerIsRelative: true,
			TriggerDuration:   -travelTime,
		})
	}
	// Standard reminders
	event.Alarms = append(event.Alarms,
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       "Appointment in 1 hour",
			TriggerIsRelative: true,
			TriggerDuration:   -1 * time.Hour,
		},
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       "Appointment in 10 minutes",
			TriggerIsRelative: true,
			TriggerDuration:   -10 * time.Minute,
		},
	)

	return event, nil
}

func generateTransitionEvent(data map[string]string, _ *i18n.Translator) (*calendar.Event, error) {
	fromActivity := data["from_activity"]
	toActivity := data["to_activity"]

	// Parse start time
	startTime, err := time.Parse("2006-01-02 15:04", data["start_time"])
	if err != nil {
		return nil, fmt.Errorf(testutil.ErrMsgInvalidStartTimeFormat, err)
	}

	// Parse duration
	durStr := strings.TrimSpace(data["duration"])
	if durStr == "" {
		durStr = "15m"
	}
	dur, err := parseHumanDuration(durStr)
	if err != nil {
		return nil, fmt.Errorf(testutil.ErrMsgInvalidDurationFormat, err)
	}
	endTime := startTime.Add(dur)

	// Create event
	summary := fmt.Sprintf("🔄 Transition: %s → %s", fromActivity, toActivity)
	event := calendar.NewEvent(summary, startTime, endTime)

	// Set timezone
	if tz := data["timezone"]; tz != "" {
		event.SetTimezone(tz)
	}

	// Build description
	description := fmt.Sprintf("Buffer time between activities\n\nFrom: %s\nTo: %s\n\n", fromActivity, toActivity)
	description += "Use this time to:\n"
	description += "- Wrap up previous task\n"
	description += "- Take a short break\n"
	description += "- Prepare for next activity\n"
	description += "- Mentally switch context\n"

	event.Description = description
	event.AddCategory("Transition")

	// Single reminder at start
	event.Alarms = append(event.Alarms, calendar.Alarm{
		Action:            "DISPLAY",
		Description:       "Transition time - wrap up and prepare",
		TriggerIsRelative: true,
		TriggerDuration:   0,
	})

	return event, nil
}

func generateDeadlineEvent(data map[string]string, _ *i18n.Translator) (*calendar.Event, error) {
	task := data["task"]

	// Parse due date
	dueDate, err := time.Parse("2006-01-02", data["due_date"])
	if err != nil {
		return nil, fmt.Errorf("invalid due date: %w", err)
	}

	// Create all-day event
	summary := fmt.Sprintf("⏰ DEADLINE: %s", task)
	event := calendar.NewEvent(summary, dueDate, dueDate.AddDate(0, 0, 1))
	event.AllDay = true

	// Set timezone
	if tz := data["timezone"]; tz != "" {
		event.SetTimezone(tz)
	}

	// Set priority
	if priorityStr := data["priority"]; priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil && priority >= 1 && priority <= 9 {
			event.Priority = priority
		}
	}

	// Build description
	var description string
	description += fmt.Sprintf("Task: %s\n", task)
	description += fmt.Sprintf("Due: %s\n\n", dueDate.Format("2006-01-02"))
	if notes := data["notes"]; notes != "" {
		description += fmt.Sprintf("Notes:\n%s\n\n", notes)
	}
	description += "⚡ Reminders set:\n"
	description += "- 1 week before\n"
	description += "- 3 days before\n"
	description += "- 1 day before\n"
	description += "- Morning of deadline\n"

	event.Description = description
	event.AddCategory("Deadline")

	// Countdown reminders
	event.Alarms = append(event.Alarms,
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       fmt.Sprintf("Deadline in 1 week: %s", task),
			TriggerIsRelative: true,
			TriggerDuration:   -7 * 24 * time.Hour,
		},
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       fmt.Sprintf("Deadline in 3 days: %s", task),
			TriggerIsRelative: true,
			TriggerDuration:   -3 * 24 * time.Hour,
		},
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       fmt.Sprintf("Deadline TOMORROW: %s", task),
			TriggerIsRelative: true,
			TriggerDuration:   -1 * 24 * time.Hour,
		},
		calendar.Alarm{
			Action:            "DISPLAY",
			Description:       fmt.Sprintf("DEADLINE TODAY: %s", task),
			TriggerIsRelative: true,
			TriggerDuration:   -9 * time.Hour, // 9 AM on the day
		},
	)

	return event, nil
}
