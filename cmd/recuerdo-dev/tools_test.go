package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, text string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755)
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func tree(t *testing.T) string {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module example.org/app\n\ngo 1.26\n")
	write(t, dir, "a/a.go", "package a\n\nimport \"example.org/app/b\"\n\nfunc A() int { return b.B() }\n")
	write(t, dir, "b/b.go", `package b

import "fmt"

func B() int {
	n := 0
	for i := 0; i < 3; i++ {
		if i > 0 && i < 2 || i == 5 {
			n++
		}
	}
	switch n {
	case 1:
		fmt.Println("one")
	case 2:
	default:
	}
	return n
}
`)
	write(t, dir, "legacy/x.go", "package x\n\nimport \"example.org/app/a\"\n")
	return dir
}

func TestModuleGraph(t *testing.T) {
	var out bytes.Buffer
	if err := moduleGraph(&out, tree(t)); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, `"a" -> "b";`) || strings.Contains(s, "fmt") || strings.Contains(s, "legacy") || !strings.HasPrefix(s, "digraph") {
		t.Errorf("graph:\n%s", s)
	}
}

func TestComplexity(t *testing.T) {
	f, _ := parser.ParseFile(token.NewFileSet(), "b.go", `package b
func B(i int) int {
	for i < 3 { if i > 0 && i < 2 || i == 5 { i++ } }
	switch i { case 1: case 2: default: }
	return i
}`, 0)
	// 1 + for + if + && + || + 2 cases
	if c := Complexity(f.Decls[0].(*ast.FuncDecl)); c != 7 {
		t.Errorf("complexity %d", c)
	}
	var out bytes.Buffer
	complexity(&out, tree(t), 3)
	if !strings.Contains(out.String(), "B  b/b.go:5") || !strings.Contains(out.String(), "1 of 2 functions above 3") {
		t.Errorf("report:\n%s", out.String())
	}
}

func TestTranslators(t *testing.T) {
	dir := t.TempDir()
	po := func(name, who string) string {
		return "msgid \"\"\nmsgstr \"\"\n\"Last-Translator: " + who + "\\n\"\n"
	}
	write(t, dir, "m1/nl.po", po("nl", "Marten de Vries <m@example.org>"))
	write(t, dir, "m2/nl.po", po("nl", "Michael Tel <t@example.org>"))
	write(t, dir, "m2/de.po", po("de", "FULL NAME <EMAIL@ADDRESS>"))
	write(t, dir, "m3/fr.po", po("fr", "Ubuntu"))
	var out bytes.Buffer
	if err := translators(&out, dir); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "nl: Marten de Vries, Michael Tel\n") || strings.Contains(s, "de:") || strings.Contains(s, "fr:") {
		t.Errorf("translators:\n%s", s)
	}
}
