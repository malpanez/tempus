package calendar

import (
	"strings"
	"testing"
)

func TestGenerateVTZDeterministic(t *testing.T) {
	a := generateVTZ("Europe/Madrid")
	b := generateVTZ("Europe/Madrid")
	if a != b || a == "" {
		t.Fatal("generateVTZ must be deterministic and non-empty for valid zones")
	}
}

func TestGenerateVTZMatchesLegacyMadridBlock(t *testing.T) {
	got := knownVTZ("Europe/Madrid")
	for _, want := range []string{
		"TZID:Europe/Madrid\r\n",
		"TZOFFSETFROM:+0100\r\n",
		"TZOFFSETTO:+0200\r\n",
		"TZNAME:CEST\r\n",
		"DTSTART:19700329T020000\r\n",
		"RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=-1SU\r\n",
		"TZNAME:CET\r\n",
		"DTSTART:19701025T030000\r\n",
		"RRULE:FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU\r\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Madrid VTIMEZONE missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestGenerateVTZNoDSTZone(t *testing.T) {
	got := knownVTZ("Asia/Tokyo")
	if strings.Contains(got, "BEGIN:DAYLIGHT") {
		t.Error("Asia/Tokyo has no DST, must emit only a STANDARD block")
	}
	for _, want := range []string{
		"BEGIN:STANDARD\r\n",
		"TZOFFSETFROM:+0900\r\n",
		"TZOFFSETTO:+0900\r\n",
		"TZNAME:JST\r\n",
		"DTSTART:19700101T000000\r\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Tokyo VTIMEZONE missing %q\ngot:\n%s", want, got)
		}
	}
	if strings.Contains(got, "RRULE") {
		t.Error("no-DST zone must not carry an RRULE")
	}
}

func TestGenerateVTZUSDSTRules(t *testing.T) {
	got := knownVTZ("America/New_York")
	for _, want := range []string{
		"RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU\r\n",
		"RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU\r\n",
		"TZNAME:EDT\r\n",
		"TZNAME:EST\r\n",
		"TZOFFSETFROM:-0500\r\n",
		"TZOFFSETTO:-0400\r\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("New_York VTIMEZONE missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestGenerateVTZSouthernHemisphereOrder(t *testing.T) {
	got := knownVTZ("Australia/Sydney")
	std := strings.Index(got, "BEGIN:STANDARD")
	dst := strings.Index(got, "BEGIN:DAYLIGHT")
	if std == -1 || dst == -1 {
		t.Fatalf("Sydney must have both blocks:\n%s", got)
	}
	if std > dst {
		t.Error("Sydney's first reference-year transition is the April STANDARD change; blocks must be chronological")
	}
}

func TestGenerateVTZInvalidAndUTC(t *testing.T) {
	if knownVTZ("Invalid/Zone") != "" {
		t.Error("invalid zone must return empty")
	}
	if knownVTZ("UTC") != "" {
		t.Error("UTC needs no VTIMEZONE")
	}
}
