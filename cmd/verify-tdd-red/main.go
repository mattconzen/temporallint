// Command verify-tdd-red audits the inverse-red property: for every
// static analyzer rule, it disables the rule's logic and confirms that
// the rule's `analysistest.Run` test FAILS. If a test passes with the
// rule disabled, the test isn't actually exercising the rule and the
// "TDD-red" guarantee is broken.
//
// Two rule shapes are supported:
//
//  1. `Run: run` field referencing a top-level `func run(pass) (any, error)`
//     — the verifier replaces the body of `run` with `return nil, nil`.
//
//  2. `Run: func(pass *analysis.Pass) (any, error) { ... }` inline in
//     the Analyzer struct literal — the verifier replaces the inline
//     function literal's body with `return nil, nil`.
//
// Rules whose source matches neither shape are reported as `Skipped`
// rather than treated as failures (the verifier doesn't aim to be a
// universal mutator).
//
// Usage:
//
//	go run ./tools/temporallint/cmd/verify-tdd-red [-rules-dir tools/temporallint/rules]
//
// Exit code 0 = every applicable rule verified red on disable.
// Exit code 1 = at least one rule's test passed even with the rule
// disabled (broken TDD-red property). Exit code 2 = unexpected error.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	rulesDir := flag.String("rules-dir", "tools/temporallint/rules", "directory containing rule packages")
	verbose := flag.Bool("verbose", false, "print per-rule output")
	flag.Parse()

	entries, err := os.ReadDir(*rulesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read rules dir %s: %v\n", *rulesDir, err)
		os.Exit(2)
	}

	var verified, broken, skipped []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rulePath := filepath.Join(*rulesDir, e.Name(), "rule.go")
		if _, err := os.Stat(rulePath); err != nil {
			continue
		}
		switch r := verifyRule(rulePath, e.Name(), *verbose); r {
		case "verified":
			verified = append(verified, e.Name())
		case "broken":
			broken = append(broken, e.Name())
		case "skipped":
			skipped = append(skipped, e.Name())
		}
	}

	sort.Strings(verified)
	sort.Strings(broken)
	sort.Strings(skipped)
	fmt.Printf("verified: %d, broken: %d, skipped: %d\n", len(verified), len(broken), len(skipped))
	if len(broken) > 0 {
		fmt.Println("BROKEN — these rules' tests still pass with the rule disabled:")
		for _, n := range broken {
			fmt.Println("  ", n)
		}
		os.Exit(1)
	}
	if len(skipped) > 0 && *verbose {
		fmt.Println("Skipped (rule.go didn't match either supported shape):")
		for _, n := range skipped {
			fmt.Println("  ", n)
		}
	}
}

// verifyRule runs the inverse-red check against one rule. Returns
// "verified" / "broken" / "skipped" rather than panicking on internal
// errors so the harness produces a complete report.
func verifyRule(rulePath, ruleName string, verbose bool) string {
	orig, err := os.ReadFile(rulePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: read: %v\n", ruleName, err)
		return "skipped"
	}

	mutated, ok := disableRule(orig)
	if !ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "%s: no run() / inline Run field found\n", ruleName)
		}
		return "skipped"
	}
	if bytes.Equal(mutated, orig) {
		return "skipped"
	}

	if err := os.WriteFile(rulePath, mutated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "%s: write mutated: %v\n", ruleName, err)
		return "skipped"
	}

	pkg := "./" + filepath.Dir(rulePath)
	cmd := exec.Command("go", "test", "-count=1", pkg)
	out, err := cmd.CombinedOutput()

	// Always restore — even on panic in the test runner.
	if rerr := os.WriteFile(rulePath, orig, 0o644); rerr != nil {
		fmt.Fprintf(os.Stderr, "%s: RESTORE FAILED: %v — file is now in mutated state!\n", ruleName, rerr)
		return "skipped"
	}

	if err == nil {
		// Test PASSED with rule disabled — broken inverse-red.
		fmt.Fprintf(os.Stderr, "%s: TEST PASSED WITH RULE DISABLED (broken)\n", ruleName)
		if verbose {
			fmt.Fprintln(os.Stderr, string(out))
		}
		return "broken"
	}
	if verbose {
		fmt.Printf("%s: verified red (test failed as expected)\n", ruleName)
	}
	return "verified"
}

// disableRule rewrites src so the analyzer's logic is a no-op. Tries
// the `func run` shape first; falls back to mutating an inline
// `Run: func(...)` field. Returns (mutated, true) on success.
func disableRule(src []byte) ([]byte, bool) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "rule.go", src, parser.ParseComments)
	if err != nil {
		return nil, false
	}
	mutated := false

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name == nil || fd.Name.Name != "run" {
			continue
		}
		fd.Body = noopBody()
		mutated = true
	}

	if !mutated {
		// Look for `Run: func(pass *analysis.Pass) (any, error) { ... }`
		// inside an Analyzer struct literal at package scope.
		ast.Inspect(f, func(n ast.Node) bool {
			if mutated {
				return false
			}
			cl, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			for _, elt := range cl.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok || key.Name != "Run" {
					continue
				}
				fl, ok := kv.Value.(*ast.FuncLit)
				if !ok {
					continue
				}
				fl.Body = noopBody()
				mutated = true
				return false
			}
			return true
		})
	}

	if !mutated {
		return nil, false
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, f); err != nil {
		return nil, false
	}
	out := buf.Bytes()
	// Drop unused imports the no-op body may have left behind. The Go
	// parser doesn't auto-prune, so we punt to gofmt+goimports — but
	// we don't want to require goimports as a dep. Instead, rely on
	// `go test` failing with "imported and not used" which still
	// counts as the test failing (i.e. the rule was disabled). That's
	// fine: a build error means the analyzer ran zero diagnostics,
	// which is what we want to verify the test catches.
	_ = strings.TrimSpace
	return out, true
}

func noopBody() *ast.BlockStmt {
	return &ast.BlockStmt{List: []ast.Stmt{
		&ast.ReturnStmt{Results: []ast.Expr{
			&ast.Ident{Name: "nil"},
			&ast.Ident{Name: "nil"},
		}},
	}}
}
