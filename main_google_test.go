package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"tempus/internal/calendar"
)

func findGoogleImportCmd() *cobra.Command {
	root := newGoogleCmd()
	for _, cmd := range root.Commands() {
		if strings.HasPrefix(cmd.Use, "import") {
			return cmd
		}
	}
	return nil
}

func TestGoogleImportCommand(t *testing.T) {
	importCmd := findGoogleImportCmd()
	if importCmd == nil {
		t.Fatalf("google import command not found")
	}

	// set up mock Google endpoints
	var inserted bool
	var pollCount int
	var deviceShown bool
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
			if pollCount == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
				return
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse form: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"fresh","refresh_token":"r1","expires_in":3600,"token_type":"Bearer"}`))
		case strings.Contains(r.URL.Path, "/calendars/primary/events"):
			inserted = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"evt"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	cal := calendar.NewCalendar()
	start := time.Now().Add(time.Hour).UTC()
	end := start.Add(time.Hour)
	ev := calendar.NewEvent("CLI Test", start, end)
	cal.AddEvent(ev)
	icsPath := filepath.Join(tmpDir, "event.ics")
	if err := os.WriteFile(icsPath, []byte(cal.ToICS()), 0o644); err != nil {
		t.Fatalf("failed to write ICS: %v", err)
	}

	set := func(name, value string) {
		if err := importCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set flag %s: %v", name, err)
		}
	}

	set("input", icsPath)
	set("calendar", "primary")
	set("client-id", "client-id")
	set("client-secret", "secret")
	set("token-file", tokenPath)
	set("device-endpoint", server.URL+"/device")
	set("token-endpoint", server.URL+"/token")
	set("api-base", server.URL)

	if err := runGoogleImport(importCmd, nil); err != nil {
		t.Fatalf("runGoogleImport returned error: %v", err)
	}

	if !inserted {
		t.Fatalf("expected event insert request to be sent")
	}
	if !deviceShown || pollCount < 2 {
		t.Fatalf("expected device authorization flow to run; deviceShown=%v pollCount=%d", deviceShown, pollCount)
	}

	if _, err := os.Stat(tokenPath); err != nil {
		t.Fatalf("expected token file to be created: %v", err)
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
