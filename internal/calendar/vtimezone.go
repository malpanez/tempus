package calendar

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// vtzReferenceYear anchors transition discovery so generated VTIMEZONE
// blocks are deterministic regardless of when the binary runs. DTSTART
// values are then projected back onto 1970, matching the long-standing
// iCalendar convention (and the blocks tempus used to hardcode).
const vtzReferenceYear = 2030

var (
	vtzCacheMu sync.Mutex
	vtzCache   = map[string]string{}
)

// knownVTZ returns a VTIMEZONE block for any loadable IANA zone, generated
// from the embedded tzdata. Zones without DST get a single STANDARD block;
// zones with the common two-transition pattern get DAYLIGHT+STANDARD blocks
// with FREQ=YEARLY rules derived from the reference year. Unresolvable
// zones (and UTC, which needs no VTIMEZONE) return "".
func knownVTZ(tzid string) string {
	tzid = strings.TrimSpace(tzid)
	if tzid == "" || tzid == "UTC" {
		return ""
	}

	vtzCacheMu.Lock()
	defer vtzCacheMu.Unlock()
	if v, ok := vtzCache[tzid]; ok {
		return v
	}
	v := generateVTZ(tzid)
	vtzCache[tzid] = v
	return v
}

// HasVTZDefinition reports whether a VTIMEZONE definition can be emitted
// for the given TZID — true for every loadable IANA zone.
func HasVTZDefinition(tzid string) bool {
	return knownVTZ(tzid) != ""
}

type vtzTransition struct {
	at         time.Time // UTC instant the new offset takes effect
	fromOffset int       // seconds
	toOffset   int       // seconds
	name       string    // abbreviation after the transition
	isDST      bool
}

func generateVTZ(tzid string) string {
	loc, err := time.LoadLocation(tzid)
	if err != nil {
		return ""
	}

	transitions := findTransitions(loc, vtzReferenceYear)

	var b strings.Builder
	b.WriteString("BEGIN:VTIMEZONE\r\n")
	b.WriteString("TZID:" + tzid + "\r\n")
	b.WriteString("X-LIC-LOCATION:" + tzid + "\r\n")

	if len(transitions) != 2 {
		// No DST, or an irregular year (>2 transitions): emit a single
		// STANDARD block with the January offset. Clients fall back to
		// the IANA TZID for anything more exotic.
		jan := time.Date(vtzReferenceYear, 1, 1, 12, 0, 0, 0, loc)
		name, offset := jan.Zone()
		writeVTZBlock(&b, "STANDARD", offset, offset, name, "DTSTART:19700101T000000\r\n", "")
		b.WriteString("END:VTIMEZONE\r\n")
		return b.String()
	}

	for _, tr := range transitions {
		kind := "STANDARD"
		if tr.isDST {
			kind = "DAYLIGHT"
		}
		dtstart, rrule := vtzRecurrence(tr)
		writeVTZBlock(&b, kind, tr.fromOffset, tr.toOffset, tr.name, dtstart, rrule)
	}

	b.WriteString("END:VTIMEZONE\r\n")
	return b.String()
}

func writeVTZBlock(b *strings.Builder, kind string, fromOffset, toOffset int, name, dtstart, rrule string) {
	b.WriteString("BEGIN:" + kind + "\r\n")
	b.WriteString("TZOFFSETFROM:" + formatUTCOffset(fromOffset) + "\r\n")
	b.WriteString("TZOFFSETTO:" + formatUTCOffset(toOffset) + "\r\n")
	if name != "" {
		b.WriteString("TZNAME:" + name + "\r\n")
	}
	b.WriteString(dtstart)
	if rrule != "" {
		b.WriteString(rrule)
	}
	b.WriteString("END:" + kind + "\r\n")
}

// findTransitions scans the reference year for UTC-offset changes, refined
// to minute precision.
func findTransitions(loc *time.Location, year int) []vtzTransition {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	var out []vtzTransition
	prev := start
	_, prevOffset := prev.In(loc).Zone()

	for t := start.Add(time.Hour); !t.After(end); t = t.Add(time.Hour) {
		_, offset := t.In(loc).Zone()
		if offset == prevOffset {
			prev = t
			prevOffset = offset
			continue
		}
		at := refineTransition(loc, prev, t)
		after := at.In(loc)
		name, toOffset := after.Zone()
		out = append(out, vtzTransition{
			at:         at,
			fromOffset: prevOffset,
			toOffset:   toOffset,
			name:       name,
			isDST:      after.IsDST(),
		})
		prev = t
		prevOffset = offset
	}
	return out
}

// refineTransition binary-searches the exact minute the offset changes
// inside (lo, hi].
func refineTransition(loc *time.Location, lo, hi time.Time) time.Time {
	_, loOffset := lo.In(loc).Zone()
	for hi.Sub(lo) > time.Minute {
		mid := lo.Add(hi.Sub(lo) / 2).Truncate(time.Minute)
		if mid.Equal(lo) {
			break
		}
		if _, o := mid.In(loc).Zone(); o == loOffset {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi
}

// vtzRecurrence derives the DTSTART (projected onto 1970) and the YEARLY
// RRULE from a reference-year transition. The wall-clock time is expressed
// in the pre-transition offset, per RFC 5545 §3.6.5.
func vtzRecurrence(tr vtzTransition) (dtstart, rrule string) {
	wall := tr.at.UTC().Add(time.Duration(tr.fromOffset) * time.Second)
	month := wall.Month()
	weekday := wall.Weekday()

	ordinal := (wall.Day()-1)/7 + 1
	lastDay := time.Date(wall.Year(), month+1, 0, 12, 0, 0, 0, time.UTC).Day()
	isLast := wall.Day()+7 > lastDay

	byday := fmt.Sprintf("%d%s", ordinal, icsWeekday(weekday))
	if isLast {
		byday = "-1" + icsWeekday(weekday)
	}

	day1970 := matchingDay1970(month, weekday, ordinal, isLast)
	dtstart = fmt.Sprintf("DTSTART:1970%02d%02dT%02d%02d%02d\r\n",
		int(month), day1970, wall.Hour(), wall.Minute(), wall.Second())
	rrule = fmt.Sprintf("RRULE:FREQ=YEARLY;BYMONTH=%d;BYDAY=%s\r\n", int(month), byday)
	return dtstart, rrule
}

// matchingDay1970 finds the day-of-month in 1970 matching the same
// month/weekday/ordinal rule, so DTSTART is itself an instance of the RRULE.
func matchingDay1970(month time.Month, weekday time.Weekday, ordinal int, isLast bool) int {
	if isLast {
		last := time.Date(1970, month+1, 0, 12, 0, 0, 0, time.UTC)
		for d := last; ; d = d.AddDate(0, 0, -1) {
			if d.Weekday() == weekday {
				return d.Day()
			}
		}
	}
	count := 0
	first := time.Date(1970, month, 1, 12, 0, 0, 0, time.UTC)
	for d := first; d.Month() == month; d = d.AddDate(0, 0, 1) {
		if d.Weekday() == weekday {
			count++
			if count == ordinal {
				return d.Day()
			}
		}
	}
	return 1
}

func icsWeekday(w time.Weekday) string {
	return [...]string{"SU", "MO", "TU", "WE", "TH", "FR", "SA"}[w]
}

func formatUTCOffset(seconds int) string {
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	return fmt.Sprintf("%s%02d%02d", sign, h, m)
}
