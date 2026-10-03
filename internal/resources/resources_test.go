package resources

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirFromEnvironment(t *testing.T) {
	t.Setenv("RECUERDO_DATA", "/somewhere")
	if Dir() != "/somewhere" {
		t.Errorf("Dir() = %q", Dir())
	}
}

func TestDirFindsWorkingDirectory(t *testing.T) {
	t.Setenv("RECUERDO_DATA", "")
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "data", "maps"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)
	if got, _ := filepath.EvalSymlinks(Dir()); got != mustEval(t, tmp) {
		t.Errorf("Dir() = %q, want %q", Dir(), tmp)
	}
}

func TestDirFallsBackToDot(t *testing.T) {
	t.Setenv("RECUERDO_DATA", "")
	t.Chdir(t.TempDir())
	if Dir() != "." {
		t.Errorf("Dir() = %q, want .", Dir())
	}
}

func mustEval(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
