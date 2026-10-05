package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// goModule is the module path of the tree at dir (from go.mod).
func goModule(dir string) (string, error) {
	f, err := os.Open(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if m, ok := strings.CutPrefix(sc.Text(), "module "); ok {
			return strings.TrimSpace(m), nil
		}
	}
	return "", fmt.Errorf("no module in %s/go.mod", dir)
}

// goFiles calls fn for every non-test Go file under dir (legacy, vendor
// and hidden directories left out).
func goFiles(dir string, fn func(path string, f *ast.File, fset *token.FileSet)) error {
	fset := token.NewFileSet()
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != dir && (strings.HasPrefix(name, ".") || name == "legacy" || name == "vendor" || name == "testdata" || name == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil // not this tool's business
		}
		fn(path, f, fset)
		return nil
	})
}

// moduleGraph writes the module's packages and their imports of each
// other as a Graphviz graph (OpenTeacher's moduleGraph drew its modules'
// "requires"): recuerdo-dev module-graph | dot -Tsvg > modules.svg
func moduleGraph(w io.Writer, dir string) error {
	mod, err := goModule(dir)
	if err != nil {
		return err
	}
	edges := map[[2]string]bool{}
	nodes := map[string]bool{}
	err = goFiles(dir, func(path string, f *ast.File, _ *token.FileSet) {
		rel, _ := filepath.Rel(dir, filepath.Dir(path))
		from := filepath.ToSlash(rel)
		nodes[from] = true
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if to, ok := strings.CutPrefix(p, mod+"/"); ok {
				edges[[2]string{from, to}] = true
				nodes[to] = true
			}
		}
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "digraph recuerdo {\n\trankdir=LR;\n\tnode [shape=box, fontname=sans, fontsize=10];")
	for _, n := range sorted(nodes) {
		fmt.Fprintf(w, "\t%q;\n", n)
	}
	var es []string
	for e := range edges {
		es = append(es, fmt.Sprintf("\t%q -> %q;", e[0], e[1]))
	}
	sort.Strings(es)
	fmt.Fprintln(w, strings.Join(es, "\n"))
	fmt.Fprintln(w, "}")
	return nil
}

func sorted(set map[string]bool) []string {
	var out []string
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Complexity is a function's cyclomatic complexity: 1 plus its branches
// (if, for, case, &&, ||), as gocyclo counts.
func Complexity(fn *ast.FuncDecl) int {
	c := 1
	ast.Inspect(fn, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			c++
		case *ast.CaseClause:
			if n.List != nil { // not default
				c++
			}
		case *ast.CommClause:
			if n.Comm != nil {
				c++
			}
		case *ast.BinaryExpr:
			if n.Op == token.LAND || n.Op == token.LOR {
				c++
			}
		}
		return true
	})
	return c
}

// complexity lists the functions above a complexity, most complex first
// (OpenTeacher's codeComplexity profile did this for its Python).
func complexity(w io.Writer, dir string, over int) error {
	type fn struct {
		name, pos string
		c         int
	}
	var fns []fn
	err := goFiles(dir, func(path string, f *ast.File, fset *token.FileSet) {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Body != nil {
				name := fd.Name.Name
				if fd.Recv != nil && len(fd.Recv.List) > 0 {
					name = recvName(fd.Recv.List[0].Type) + "." + name
				}
				p := fset.Position(fd.Pos())
				rel, _ := filepath.Rel(dir, p.Filename)
				fns = append(fns, fn{name, fmt.Sprintf("%s:%d", filepath.ToSlash(rel), p.Line), Complexity(fd)})
			}
		}
	})
	if err != nil {
		return err
	}
	sort.Slice(fns, func(i, j int) bool { return fns[i].c > fns[j].c || (fns[i].c == fns[j].c && fns[i].pos < fns[j].pos) })
	total, n := 0, 0
	for _, f := range fns {
		total += f.c
		if f.c > over {
			fmt.Fprintf(w, "%4d  %s  %s\n", f.c, f.name, f.pos)
			n++
		}
	}
	if len(fns) > 0 {
		fmt.Fprintf(w, "%d of %d functions above %d; average complexity %.1f\n", n, len(fns), over, float64(total)/float64(len(fns)))
	}
	return nil
}

func recvName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return recvName(t.X)
	case *ast.IndexExpr:
		return recvName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return "?"
}

var lastTranslator = regexp.MustCompile(`"Last-Translator: ([^<"\\]+?)\s*(<[^>]*>)?\\n"`)

// translators lists who translated OpenTeacher, per language, from the
// Last-Translator lines of its .po files (OpenTeacher asked Launchpad,
// which no longer serves them): "nl: Marten de Vries, Michael Tel".
// recuerdo-dev translators > data/translators.txt
func translators(w io.Writer, dir string) error {
	byLang := map[string]map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".po") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		m := lastTranslator.FindSubmatch(data)
		if m == nil {
			return nil
		}
		name := strings.TrimSpace(string(m[1]))
		if name == "" || name == "Ubuntu" || strings.Contains(strings.ToUpper(name), "FULL NAME") { // Ubuntu: Launchpad's imports
			return nil
		}
		lang := strings.TrimSuffix(filepath.Base(path), ".po")
		if byLang[lang] == nil {
			byLang[lang] = map[string]bool{}
		}
		byLang[lang][name] = true
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "# Who translated OpenTeacher (from its .po files' Last-Translator lines),")
	fmt.Fprintln(w, "# made by: recuerdo-dev translators > data/translators.txt")
	for _, lang := range sortedKeys(byLang) {
		fmt.Fprintf(w, "%s: %s\n", lang, strings.Join(sorted(byLang[lang]), ", "))
	}
	return nil
}

func sortedKeys(m map[string]map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
