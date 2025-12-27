package timezone

// Import tzdata to embed timezone database for platforms that don't have it installed.
// This blank import ensures the timezone database is available in the compiled binary.
import _ "time/tzdata"
