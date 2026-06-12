package cli

import (
	"strings"
	"tempus/internal/testutil"
	"testing"
)

func TestNewLocaleCmd(t *testing.T) {
	app := TestApp()
	cmd := NewLocaleCmd(app)
	if cmd == nil {
		t.Fatal("NewLocaleCmd() returned nil")
	}
	if cmd.Use != "locale" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "locale")
	}

	subcommands := cmd.Commands()
	if len(subcommands) != 1 {
		t.Errorf("expected 1 subcommand, got %d", len(subcommands))
	}

	listCmd := subcommands[0]
	if !strings.HasPrefix(listCmd.Use, "list") {
		t.Error("locale command should have 'list' subcommand")
	}
}
