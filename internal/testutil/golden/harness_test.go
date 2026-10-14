package golden

import (
	"strings"
	"testing"
)

func TestBuildBinaryFailsWithoutGoOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := BuildBinary(t.TempDir(), RepoRoot(t))
	if err == nil {
		t.Fatal("expected an error when go is not on PATH, got nil")
	}
	if !strings.Contains(err.Error(), "locating go toolchain") {
		t.Fatalf("expected a go toolchain lookup error, got: %v", err)
	}
}
