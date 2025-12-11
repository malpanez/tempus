package timezone

import (
	"bufio"
	"bytes"
	"embed"
	"strings"
)

// Pin the import so overzealous formatters don’t drop it before //go:embed runs.
var _ embed.FS

// Keep this import here; it is used by the //go:embed directive below.
//
//go:embed data/zone1970.tab
var zone1970Tab []byte

type tabRow struct {
	CC      string // country code(s), comma separated
	TZ      string // IANA timezone (Area/City)
	Comment string // optional comment
}

func parseZone1970Tab() []tabRow {
	if len(zone1970Tab) == 0 {
		return nil
	}
	r := bytes.NewReader(zone1970Tab)
	sc := bufio.NewScanner(r)
	rows := make([]tabRow, 0, 800)

	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Format: CC<TAB>coordinates<TAB>TZ<TAB>comments?
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		cc := strings.TrimSpace(parts[0])
		tz := strings.TrimSpace(parts[2])
		comment := ""
		if len(parts) >= 4 {
			comment = strings.TrimSpace(parts[3])
		}
		if cc == "" || tz == "" {
			continue
		}
		rows = append(rows, tabRow{CC: cc, TZ: tz, Comment: comment})
	}
	return rows
}
