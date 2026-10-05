// Command recuerdo-dev is Recuerdo's developer tooling, after
// OpenTeacher's developer profiles, as small command line tools:
//
//	recuerdo-dev module-graph [dir]      the packages and their imports, as Graphviz (dot)
//	recuerdo-dev complexity [-over n] [dir]  the most complex functions (cyclomatic complexity)
//	recuerdo-dev translators [legacy dir]  OpenTeacher's translators per language, for data/translators.txt
//
// dir is the source tree (default "."). The translation tooling itself is
// scripts/extract_strings.py and scripts/merge_translations.py.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "module-graph":
		err = moduleGraph(os.Stdout, arg(2, "."))
	case "complexity":
		fs := flag.NewFlagSet("complexity", flag.ExitOnError)
		over := fs.Int("over", 15, "list functions with a complexity above this")
		fs.Parse(os.Args[2:])
		dir := "."
		if fs.NArg() > 0 {
			dir = fs.Arg(0)
		}
		err = complexity(os.Stdout, dir, *over)
	case "translators":
		err = translators(os.Stdout, arg(2, "legacy/modules"))
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func arg(i int, def string) string {
	if len(os.Args) > i {
		return os.Args[i]
	}
	return def
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: recuerdo-dev module-graph [dir]
       recuerdo-dev complexity [-over n] [dir]
       recuerdo-dev translators [legacy modules dir]`)
	os.Exit(2)
}
