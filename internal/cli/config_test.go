package cli

import (
	"strings"
	"tempus/internal/testutil"
	"testing"
)

func TestNewConfigCmd(t *testing.T) {
	app := TestApp()
	cmd := NewConfigCmd(app)
	if cmd == nil {
		t.Fatal("NewConfigCmd() returned nil")
	}
	if cmd.Use != "config" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "config")
	}

	subcommands := cmd.Commands()
	if len(subcommands) != 3 {
		t.Errorf("expected 3 subcommands, got %d", len(subcommands))
	}

	var hasSet, hasList, hasAlarmProfiles bool
	for _, sub := range subcommands {
		if strings.HasPrefix(sub.Use, "set") {
			hasSet = true
		}
		if strings.HasPrefix(sub.Use, "list") {
			hasList = true
		}
		if strings.HasPrefix(sub.Use, "alarm-profiles") {
			hasAlarmProfiles = true
		}
	}
	if !hasSet {
		t.Error("config command missing 'set' subcommand")
	}
	if !hasList {
		t.Error("config command missing 'list' subcommand")
	}
	if !hasAlarmProfiles {
		t.Error("config command missing 'alarm-profiles' subcommand")
	}
}

func TestRunConfigSet(t *testing.T) {
	app := TestApp()
	cmd := NewConfigCmd(app)
	setCmd := findSubcommand(cmd, "set")
	if setCmd == nil {
		t.Fatal("set subcommand not found")
	}

	if setCmd.Args == nil {
		t.Error("set command should have Args validator")
	}
}

func TestRunConfigList(t *testing.T) {
	app := TestApp()
	cmd := NewConfigCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}
	if listCmd.RunE == nil {
		t.Error("list command should have RunE function")
	}
}
