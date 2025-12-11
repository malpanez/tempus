package google_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"tempus/internal/calendar"
	gsync "tempus/internal/sync/google"
)

func TestClientImportICSInsertsEvents(t *testing.T) {
	t.Parallel()

	cal := calendar.NewCalendar()
	start1 := time.Date(2025, 3, 1, 10, 0, 0, 0, time.FixedZone("CET", 3600))
	end1 := time.Date(2025, 3, 1, 12, 0, 0, 0, time.FixedZone("GMT", 0))
	ev1 := calendar.NewEvent("Flight MAD-DUB", start1, end1)
	ev1.SetStartTimezone("Europe/Madrid")
	ev1.SetEndTimezone("Europe/Dublin")
	ev1.Description = "Ryanair FR1234"
	ev1.Location = "Madrid T1"
	ev1.RRule = "FREQ=DAILY;COUNT=2"
	ev1.Alarms = []calendar.Alarm{
		{
			Action:            "DISPLAY",
			Description:       "Boarding",
			TriggerIsRelative: true,
			TriggerDuration:   -15 * time.Minute,
		},
	}
	cal.AddEvent(ev1)

	start2 := time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC)
	end2 := start2.AddDate(0, 0, 1)
	ev2 := calendar.NewEvent("Neuro Day", start2, end2)
	ev2.AllDay = true
	cal.AddEvent(ev2)

	ics := cal.ToICS()

	var mu sync.Mutex
	var requests []map[string]interface{}

	server := mustHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/token"):
			_ = r.ParseForm()
			if got := r.FormValue("grant_type"); got != "refresh_token" {
				t.Fatalf("unexpected grant_type %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"new-token","expires_in":3600,"token_type":"Bearer"}`))
		case strings.Contains(r.URL.Path, "/calendars/primary/events"):
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if want := "Bearer new-token"; r.Header.Get("Authorization") != want {
				t.Fatalf("expected Authorization %q, got %q", want, r.Header.Get("Authorization"))
			}
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			mu.Lock()
			requests = append(requests, payload)
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"evt"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")
	expired := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano)
	if err := os.WriteFile(tokenPath, []byte(`{"access_token":"old-token","refresh_token":"refresh","token_type":"Bearer","expiry":"`+expired+`"}`), 0600); err != nil {
		t.Fatalf("write token: %v", err)
	}

	opts := gsync.Options{
		ClientID:        "client-id",
		ClientSecret:    "secret",
		TokenFile:       tokenPath,
		DeviceEndpoint:  server.URL + "/device",
		TokenEndpoint:   server.URL + "/token",
		CalendarBaseURL: server.URL,
		HTTPClient:      server.Client(),
	}

	client, err := gsync.NewClient(context.Background(), opts)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	if err := client.ImportICS(context.Background(), "primary", ics); err != nil {
		t.Fatalf("ImportICS error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 2 {
		t.Fatalf("expected 2 event insert requests, got %d", len(requests))
	}

	first := requests[0]
	if got := first["summary"]; got != "Flight MAD-DUB" {
		t.Fatalf("summary mismatch: %v", got)
	}
	start := first["start"].(map[string]interface{})
	if _, ok := start["dateTime"]; !ok {
		t.Fatalf("expected dateTime in start %#v", start)
	}
	if tz := start["timeZone"]; tz != "Europe/Madrid" {
		t.Fatalf("expected start timezone Europe/Madrid, got %#v", tz)
	}
	end := first["end"].(map[string]interface{})
	if _, ok := end["dateTime"]; !ok {
		t.Fatalf("expected dateTime in end %#v", end)
	}
	recurrence := first["recurrence"].([]interface{})
	if len(recurrence) != 1 || recurrence[0] != "RRULE:FREQ=DAILY;COUNT=2" {
		t.Fatalf("unexpected recurrence %#v", recurrence)
	}
	rem := first["reminders"].(map[string]interface{})
	if rem["useDefault"].(bool) {
		t.Fatalf("expected useDefault false in reminders")
	}
	overrides := rem["overrides"].([]interface{})
	if len(overrides) != 1 {
		t.Fatalf("expected one reminder override got %d", len(overrides))
	}
	override := overrides[0].(map[string]interface{})
	if override["method"] != "popup" || override["minutes"].(float64) != 15 {
		t.Fatalf("unexpected reminder override %#v", override)
	}

	second := requests[1]
	start2Payload := second["start"].(map[string]interface{})
	if got := start2Payload["date"]; got != "2025-03-05" {
		t.Fatalf("expected all-day start date, got %#v", got)
	}
	end2Payload := second["end"].(map[string]interface{})
	if got := end2Payload["date"]; got != "2025-03-06" {
		t.Fatalf("expected all-day end date, got %#v", got)
	}
}

func TestImportICSConvertsAbsoluteTriggers(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var payloads []map[string]interface{}

	server := mustHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/calendars/primary/events") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		mu.Lock()
		payloads = append(payloads, payload)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"evt"}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")
	expiry := time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)
	token := `{"access_token":"valid","refresh_token":"refresh","token_type":"Bearer","expiry":"` + expiry + `"}`
	if err := os.WriteFile(tokenPath, []byte(token), 0o600); err != nil {
		t.Fatalf("write token: %v", err)
	}

	opts := gsync.Options{
		ClientID:        "client-id",
		ClientSecret:    "secret",
		TokenFile:       tokenPath,
		CalendarBaseURL: server.URL,
		HTTPClient:      server.Client(),
	}

	client, err := gsync.NewClient(context.Background(), opts)
	// ensure client loads token without contacting endpoints
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	ics := strings.Join([]string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//Tempus//Test//EN",
		"X-WR-TIMEZONE:Europe/Madrid",
		"BEGIN:VEVENT",
		"UID:abs-1",
		"SUMMARY:Absolute Alarm Event",
		"DTSTART;TZID=Europe/Madrid:20250301T093000",
		"DTEND;TZID=Europe/Madrid:20250301T103000",
		"BEGIN:VALARM",
		"ACTION:DISPLAY",
		"TRIGGER;VALUE=DATE-TIME;TZID=Europe/Madrid:20250301T090000",
		"DESCRIPTION:Reminder",
		"END:VALARM",
		"END:VEVENT",
		"END:VCALENDAR",
	}, "\r\n")

	if err := client.ImportICS(context.Background(), "primary", ics); err != nil {
		t.Fatalf("ImportICS error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(payloads) != 1 {
		t.Fatalf("expected 1 payload, got %d", len(payloads))
	}
	reminders, ok := payloads[0]["reminders"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected reminders block, got %#v", payloads[0])
	}
	if reminders["useDefault"].(bool) {
		t.Fatalf("expected custom reminders block")
	}
	overrides, ok := reminders["overrides"].([]interface{})
	if !ok || len(overrides) != 1 {
		t.Fatalf("expected single override, got %#v", overrides)
	}
	override := overrides[0].(map[string]interface{})
	if minutes := override["minutes"].(float64); minutes != 30 {
		t.Fatalf("expected 30 minute reminder, got %v", minutes)
	}
}

func TestClientPerformsDeviceFlowWhenNoToken(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var deviceShown bool
	var pollCount int
	var importCalled bool

	server := mustHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/device"):
			deviceShown = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"device_code":"device123",
				"user_code":"ABCD-EFGH",
				"verification_url":"https://example.com/device",
				"expires_in":1800,
				"interval":1
			}`))
		case strings.HasSuffix(r.URL.Path, "/token"):
			pollCount++
			_ = r.ParseForm()
			if pollCount < 2 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"fresh","refresh_token":"r1","expires_in":3600,"token_type":"Bearer"}`))
		case strings.Contains(r.URL.Path, "/calendars/primary/events"):
			importCalled = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"evt"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	opts := gsync.Options{
		ClientID:        "client-id",
		ClientSecret:    "secret",
		TokenFile:       tokenPath,
		DeviceEndpoint:  server.URL + "/device",
		TokenEndpoint:   server.URL + "/token",
		CalendarBaseURL: server.URL,
		HTTPClient:      server.Client(),
	}

	client, err := gsync.NewClient(context.Background(), opts)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cal := calendar.NewCalendar()
	cal.AddEvent(calendar.NewEvent("Demo", time.Now().UTC(), time.Now().Add(time.Hour).UTC()))
	ics := cal.ToICS()

	if err := client.ImportICS(ctx, "primary", ics); err != nil {
		t.Fatalf("ImportICS error: %v", err)
	}

	if !deviceShown {
		t.Fatal("expected device authorization to be triggered")
	}
	if pollCount < 2 {
		t.Fatalf("expected at least two polling attempts, got %d", pollCount)
	}
	if !importCalled {
		t.Fatal("expected events import to be executed after authorization")
	}

	mu.Lock()
	defer mu.Unlock()
	if _, err := os.Stat(tokenPath); err != nil {
		t.Fatalf("expected token file to be written: %v", err)
	}
}

func mustHTTPServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprint(r)
			if strings.Contains(msg, "operation not permitted") {
				t.Skipf("skipping test: network sandbox prevented httptest server (%s)", msg)
			}
			panic(r)
		}
	}()
	return httptest.NewServer(handler)
}
