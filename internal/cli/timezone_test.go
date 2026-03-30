package cli

import (
	"strings"
	"tempus/internal/testutil"
	"testing"
)

func TestCityToIANA(t *testing.T) {
	tests := []struct {
		city    string
		want    string
		wantErr bool
	}{
		// Spain
		{"madrid", testutil.TZEuropeMadrid, false},
		{"barcelona", testutil.TZEuropeMadrid, false},
		{"melilla", testutil.TZAfricaCeuta, false},
		{"ceuta", testutil.TZAfricaCeuta, false},
		{"canarias", testutil.TZAtlanticCanary, false},
		{"tenerife", testutil.TZAtlanticCanary, false},

		// Brazil
		{"pelotas", testutil.TZAmericaSaoPaulo, false},
		{"porto alegre", testutil.TZAmericaSaoPaulo, false},
		{"campo grande", testutil.TZAmericaCampoGrande, false},
		{"manaus", "America/Manaus", false},
		{"rio", testutil.TZAmericaSaoPaulo, false},
		{"sao paulo", testutil.TZAmericaSaoPaulo, false},

		// Ireland/UK
		{"dublin", testutil.TZEuropeDublin, false},
		{"london", testutil.TZEuropeLondon, false},

		// Unknown -- should error
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.city, func(t *testing.T) {
			got, err := cityToIANA(tt.city)
			if tt.wantErr {
				if err == nil {
					t.Errorf("cityToIANA(%q) expected error, got nil", tt.city)
				} else {
					if !strings.Contains(err.Error(), "Unknown city") {
						t.Errorf("cityToIANA(%q) error = %q, want substring %q", tt.city, err.Error(), "Unknown city")
					}
					if !strings.Contains(err.Error(), "tempus timezone list --search") {
						t.Errorf("cityToIANA(%q) error = %q, want substring %q", tt.city, err.Error(), "tempus timezone list --search")
					}
				}
			} else {
				if err != nil {
					t.Errorf("cityToIANA(%q) unexpected error: %v", tt.city, err)
				}
				if got != tt.want {
					t.Errorf("cityToIANA(%q) = %q, want %q", tt.city, got, tt.want)
				}
			}
		})
	}
}

func TestNewTimezoneCmd(t *testing.T) {
	app := TestApp()
	cmd := NewTimezoneCmd(app)
	if cmd == nil {
		t.Fatal("NewTimezoneCmd() returned nil")
	}
	if cmd.Use != "timezone" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "timezone")
	}

	subcommands := cmd.Commands()
	if len(subcommands) != 2 {
		t.Errorf("expected 2 subcommands, got %d", len(subcommands))
	}

	var hasList, hasInfo bool
	for _, sub := range subcommands {
		if strings.HasPrefix(sub.Use, "list") {
			hasList = true
		}
		if strings.HasPrefix(sub.Use, "info") {
			hasInfo = true
		}
	}
	if !hasList {
		t.Error("timezone command missing 'list' subcommand")
	}
	if !hasInfo {
		t.Error("timezone command missing 'info' subcommand")
	}
}
