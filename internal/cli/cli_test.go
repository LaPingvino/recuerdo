package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/ocr"
)

func run(t *testing.T, stdin string, args ...string) (string, int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Run(args, strings.NewReader(stdin), &out, &errOut)
	return out.String() + errOut.String(), code
}

func TestNewViewReverseConvertMerge(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "dieren.otwd")
	if out, code := run(t, "hond = dog\nkat = cat\n", "new-word-list", "-t", "Dieren", "-q", "Dutch", "-a", "English", list); code != 0 {
		t.Fatalf("new: %d %s", code, out)
	}
	if out, _ := run(t, "", "view-word-list", list); out != "hond  dog\nkat   cat\n" {
		t.Errorf("view list: %q", out)
	}
	if out, _ := run(t, "", "view-word-list", "-p", "question-lang", list); out != "Dutch\n" {
		t.Errorf("view language: %q", out)
	}

	reversed := filepath.Join(dir, "reversed.otwd")
	if out, code := run(t, "", "reverse-list", list, reversed); code != 0 {
		t.Fatalf("reverse: %s", out)
	}
	if out, _ := run(t, "", "view-word-list", reversed); out != "dog  hond\ncat  kat\n" {
		t.Errorf("reversed: %q", out)
	}

	if out, code := run(t, "", "convert", "-f", "csv", list); code != 0 || !strings.Contains(out, "dieren.csv") {
		t.Fatalf("convert: %d %s", code, out)
	}
	csv, err := lesson.NewFileLoader().LoadFile(filepath.Join(dir, "dieren.csv"))
	if err != nil || len(csv.List.Items) != 2 {
		t.Fatalf("converted file: %v", err)
	}

	more := filepath.Join(dir, "more.otwd")
	run(t, "vis = fish\n", "new-word-list", more, "-")
	merged := filepath.Join(dir, "merged.otwd")
	if out, code := run(t, "", "merge", merged, list, more); code != 0 {
		t.Fatalf("merge: %s", out)
	}
	if out, _ := run(t, "", "view-word-list", merged); strings.Count(out, "\n") != 3 || !strings.Contains(out, "vis") {
		t.Errorf("merged: %q", out)
	}
}

func TestPractise(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "dieren.otwd")
	run(t, "hond = dog\nkat = cat\n", "new-word-list", list)
	out, code := run(t, "dog\nmouse\n", "practise-word-list", list)
	if code != 0 || !strings.Contains(out, "[1/2] hond: Right!") || !strings.Contains(out, "Wrong, the answer is: cat") ||
		!strings.HasSuffix(out, "1 of 2 right.\n") {
		t.Errorf("practise: %d %q", code, out)
	}
	// an empty line stops early
	if out, _ := run(t, "\n", "practise-word-list", list); !strings.HasSuffix(out, "0 of 0 right.\n") {
		t.Errorf("stop: %q", out)
	}
}

func TestOtherCommands(t *testing.T) {
	if out, code := run(t, "", "authors", "-c", "nonexistent role"); code != 0 || out != "" {
		t.Errorf("authors filter: %q", out)
	}
	if out, _ := run(t, "", "authors"); !strings.Contains(out, "Marten de Vries") {
		t.Errorf("authors: %q", out[:min(len(out), 200)])
	}
	if out, code := run(t, "", "nope"); code != 2 || !strings.Contains(out, "unknown command") {
		t.Errorf("unknown: %d %q", code, out)
	}
	if out, code := run(t, "", "merge", "only-one"); code != 1 || !strings.Contains(out, "need an output file") {
		t.Errorf("missing args: %d %q", code, out)
	}
	if out, _ := run(t, "", "help"); !strings.Contains(out, "practise-word-list") || !IsCommand("convert") || IsCommand("nope") {
		t.Errorf("help: %q", out)
	}
	dir := t.TempDir()
	topo := filepath.Join("..", "..", "testdata", "legacy_files", "application_x-openteachingtopography.openteacher3x.ottp")
	if out, code := run(t, "", "reverse-list", topo, filepath.Join(dir, "x.ottp")); code != 1 || !strings.Contains(out, "topography") {
		t.Errorf("reverse topo: %d %q", code, out)
	}
	// OCR, with Tesseract when it is there
	pic := filepath.Join("..", "ocr", "testdata", "list.png")
	out, code := run(t, "", "ocr-word-list", pic, filepath.Join(dir, "ocr.otwd"))
	if strings.Contains(out, ocr.ErrNoTesseract.Error()) {
		t.Skip("Tesseract is not installed")
	}
	if code != 0 || !strings.Contains(out, "5 word pairs read") {
		t.Errorf("ocr: %d %q", code, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "ocr.otwd")); errors.Is(err, os.ErrNotExist) {
		t.Error("ocr wrote no file")
	}
}
