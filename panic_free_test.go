package policydsl_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoPanicsInNonTestSource is the in-repo regression guard for the panic-free
// contract documented in AGENTS.md. erraudit ./... verifies the same property
// at the tool level, but this test fails at PR time independent of whether
// erraudit is wired into CI — so a re-introduced panic( is caught before merge.
//
// It parses every non-test .go file in the package directory and fails if any
// contains a `panic(` call expression. Using go/parser (not string matching)
// avoids false positives from comments and string literals that mention "panic".
func TestNoPanicsInNonTestSource(t *testing.T) {
	t.Parallel()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not determine working directory: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("could not read package directory: %v", err)
	}

	var offenders []string

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		path := filepath.Join(dir, name)

		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			t.Fatalf("could not parse %s: %v", name, parseErr)
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			ident, ok := call.Fun.(*ast.Ident)

			if !ok || ident.Name != "panic" {
				return true
			}

			pos := fset.Position(call.Pos())
			offenders = append(offenders, pos.String())

			return true
		})
	}

	if len(offenders) > 0 {
		t.Errorf("panic-free contract violated — panic( found in non-test source:\n  %s",
			strings.Join(offenders, "\n  "))
	}
}
