package constants

// ICS/iCalendar constants (RFC 5545)
const (
	// Event status values
	StatusConfirmed  = "CONFIRMED"
	StatusTentative  = "TENTATIVE"
	StatusCancelled  = "CANCELLED"

	// Alarm action types
	AlarmActionDisplay = "DISPLAY"
	AlarmActionEmail   = "EMAIL"
	AlarmActionAudio   = "AUDIO"

	// Recurrence rule frequencies
	RRuleFreqDaily   = "FREQ=DAILY"
	RRuleFreqWeekly  = "FREQ=WEEKLY"
	RRuleFreqMonthly = "FREQ=MONTHLY"
	RRuleFreqYearly  = "FREQ=YEARLY"

	// iCalendar line folding limit (RFC 5545)
	ICalMaxLineLength = 75
)
