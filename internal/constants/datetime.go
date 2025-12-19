package constants

// Date and time format constants used throughout the application.
// These follow Go's reference time format: Mon Jan 2 15:04:05 MST 2006
const (
	// Standard date formats
	DateFormatISO      = "2006-01-02" // ISO 8601 date
	DateFormatCompact  = "20060102"   // Compact date (no separators)
	DateFormatSlash    = "02/01/2006" // European format
	DateFormatDDMMYYYY = "DD/MM/YYYY" // Display format string

	// Standard time formats
	TimeFormatHHMM   = "15:04"    // 24-hour time (HH:MM)
	TimeFormatHHMMSS = "15:04:05" // 24-hour time with seconds

	// Combined date-time formats
	DateTimeFormatISO        = "2006-01-02 15:04"    // ISO date + time
	DateTimeFormatISOSeconds = "2006-01-02 15:04:05" // ISO with seconds

	// ICS/iCalendar specific formats (RFC 5545)
	ICSFormatUTC      = "20060102T150405Z" // UTC time in ICS format
	ICSFormatLocal    = "20060102T150405"  // Local time in ICS format
	ICSFormatDateOnly = "20060102"         // Date-only in ICS format
)
