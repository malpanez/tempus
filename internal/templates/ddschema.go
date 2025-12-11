package templates

// Data-driven template schema. We reuse the existing Field type from templates.go
// to avoid duplicate declarations.

type OutputTemplate struct {
	// Event shape
	AllDay     bool     `json:"all_day,omitempty" yaml:"all_day,omitempty"`
	Categories []string `json:"categories,omitempty" yaml:"categories,omitempty"`
	Priority   int      `json:"priority,omitempty" yaml:"priority,omitempty"`

	// Field mappings (names refer to keys in Fields / user inputs)
	StartField    string `json:"start_field,omitempty" yaml:"start_field,omitempty"`
	EndField      string `json:"end_field,omitempty" yaml:"end_field,omitempty"`
	DurationField string `json:"duration_field,omitempty" yaml:"duration_field,omitempty"` // optional "duration" field (e.g., 60, 45m, 1h30m)

	StartTZField string `json:"start_tz_field,omitempty" yaml:"start_tz_field,omitempty"`
	EndTZField   string `json:"end_tz_field,omitempty" yaml:"end_tz_field,omitempty"`
	RRuleField   string `json:"rrule_field,omitempty" yaml:"rrule_field,omitempty"`
	ExDatesField string `json:"exdates_field,omitempty" yaml:"exdates_field,omitempty"`
	AlarmsField  string `json:"alarms_field,omitempty" yaml:"alarms_field,omitempty"` // comma-separated relative alarms

	// Text templates (mustache-lite)
	SummaryTmpl     string `json:"summary_tmpl,omitempty" yaml:"summary_tmpl,omitempty"`
	LocationTmpl    string `json:"location_tmpl,omitempty" yaml:"location_tmpl,omitempty"`
	DescriptionTmpl string `json:"description_tmpl,omitempty" yaml:"description_tmpl,omitempty"`
}

type DataDrivenTemplate struct {
	SchemaVersion    int            `json:"schema_version,omitempty" yaml:"schema_version,omitempty"`
	Name             string         `json:"name" yaml:"name"`
	Description      string         `json:"description,omitempty" yaml:"description,omitempty"`
	FilenameTemplate string         `json:"filename_tmpl,omitempty" yaml:"filename_tmpl,omitempty"`
	Fields           []Field        `json:"fields" yaml:"fields"`
	Output           OutputTemplate `json:"output" yaml:"output"`
	Source           string         `json:"-" yaml:"-"` // path where this template was loaded from
}
