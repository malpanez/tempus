package constants

import "time"

// Time conversion constants
const (
	// Seconds
	SecondsPerMinute = 60
	SecondsPerHour   = 3600
	SecondsPerDay    = 86400

	// Minutes
	MinutesPerHour = 60
	MinutesPerDay  = 1440

	// Hours
	HoursPerDay = 24
	DaysPerWeek = 7
)

// Duration constants for common time periods
const (
	OneMinute = time.Minute
	OneHour   = time.Hour
	OneDay    = 24 * time.Hour
	OneWeek   = 7 * 24 * time.Hour
)
