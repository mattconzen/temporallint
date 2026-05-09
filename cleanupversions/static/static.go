// Package static implements Phase 1 of the cleanup-versions
// subcommand: walking Go source to find workflow.GetVersion calls and
// classifying the surrounding branching shape.
package static

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattconzen/temporallint/cleanupversions"
)

// Discover walks files under root (recursively), parses every .go
// file, and returns a Candidate for every workflow.GetVersion call.
// Skip-class candidates carry a Reason explaining why the rewrite
// path didn't engage.
//
// Test files (_test.go) are skipped — fixtures aren't production code.
func Discover(root string) ([]cleanupversions.Candidate, error) {
	var out []cleanupversions.Candidate
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || d.Name() == "vendor" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fc, ferr := discoverFile(path)
		if ferr != nil {
			return ferr
		}
		out = append(out, fc...)
		return nil
	})
	return out, err
}

// DiscoverString is the test-friendly variant: parses a single source
// string from a virtual filename and returns the candidates. Used by
// the static_test.go fixtures.
func DiscoverString(filename, src string) ([]cleanupversions.Candidate, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return collect(fset, f, []byte(src), filename), nil
}

func discoverFile(path string) ([]cleanupversions.Candidate, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return collect(fset, f, src, path), nil
}

func collect(fset *token.FileSet, file *ast.File, src []byte, path string) []cleanupversions.Candidate {
	var out []cleanupversions.Candidate
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		ifs, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		c, ok := classifyIf(fset, ifs, src, path)
		if ok {
			out = append(out, c)
		}
		return true
	})
	// Also flag GetVersion calls that AREN'T in an if-init expression.
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isGetVersion(call) {
			return true
		}
		// Skip if this is the call inside an if-init we already handled.
		if isInsideIfInit(call, file) {
			return true
		}
		out = append(out, cleanupversions.Candidate{
			ChangeID: extractChangeID(call),
			File:     path,
			Pos:      fset.Position(call.Pos()),
			Reason:   "unrecognised branching shape (GetVersion not in canonical if-init)",
		})
		return true
	})
	return out
}

// classifyIf returns a Candidate when ifs has the canonical
// `if v := workflow.GetVersion(ctx, "id", DefaultVersion, max); v == workflow.DefaultVersion { OLD } else { NEW }`
// shape. The Edit replaces the entire IfStmt with the body of the else
// branch, so the workflow keeps only the new behaviour.
func classifyIf(fset *token.FileSet, ifs *ast.IfStmt, src []byte, path string) (cleanupversions.Candidate, bool) {
	if ifs.Init == nil {
		return cleanupversions.Candidate{}, false
	}
	as, ok := ifs.Init.(*ast.AssignStmt)
	if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
		return cleanupversions.Candidate{}, false
	}
	call, ok := as.Rhs[0].(*ast.CallExpr)
	if !ok || !isGetVersion(call) {
		return cleanupversions.Candidate{}, false
	}
	changeID := extractChangeID(call)
	pos := fset.Position(call.Pos())
	if changeID == "" {
		return cleanupversions.Candidate{
			ChangeID: "",
			File:     path,
			Pos:      pos,
			Reason:   "non-literal change ID; cannot reason about safety",
		}, true
	}
	// Condition must be a simple comparison against workflow.DefaultVersion.
	cond, ok := ifs.Cond.(*ast.BinaryExpr)
	if !ok || cond.Op.String() != "==" {
		return cleanupversions.Candidate{
			ChangeID: changeID, File: path, Pos: pos,
			Reason: "non-canonical condition (expected `v == workflow.DefaultVersion`)",
		}, true
	}
	if !isDefaultVersion(cond.X) && !isDefaultVersion(cond.Y) {
		return cleanupversions.Candidate{
			ChangeID: changeID, File: path, Pos: pos,
			Reason: "condition does not reference workflow.DefaultVersion",
		}, true
	}
	// Else branch must exist and be a block.
	elseBlock, ok := ifs.Else.(*ast.BlockStmt)
	if !ok {
		return cleanupversions.Candidate{
			ChangeID: changeID, File: path, Pos: pos,
			Reason: "missing or non-block else branch",
		}, true
	}
	// Compute the edit: replace `ifs` with the contents of the else
	// block (without the surrounding `{` `}`). We extract the source
	// bytes between the two braces, exclusive.
	elseInner := bodySource(src, elseBlock)
	startOff := fset.Position(ifs.Pos()).Offset
	endOff := fset.Position(ifs.End()).Offset
	return cleanupversions.Candidate{
		ChangeID: changeID,
		File:     path,
		Pos:      pos,
		Edit: cleanupversions.Edit{
			StartOffset: startOff,
			EndOffset:   endOff,
			NewBytes:    elseInner,
		},
	}, true
}

func bodySource(src []byte, b *ast.BlockStmt) []byte {
	if b == nil || b.Lbrace == 0 || b.Rbrace == 0 {
		return nil
	}
	// Keep the inner body without the braces. Trim a leading newline
	// for cleanliness.
	start := int(b.Lbrace) // 1-indexed
	end := int(b.Rbrace) - 1
	if start <= 0 || end <= start || end >= len(src) {
		return nil
	}
	inner := src[start:end]
	// trim outer whitespace conservatively
	inner = trimLeadingNewline(inner)
	inner = trimTrailingTabs(inner)
	return inner
}

func trimLeadingNewline(b []byte) []byte {
	for len(b) > 0 && (b[0] == '\n' || b[0] == '\r') {
		b = b[1:]
	}
	return b
}

func trimTrailingTabs(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\t' || b[len(b)-1] == ' ') {
		b = b[:len(b)-1]
	}
	return b
}

func isGetVersion(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "GetVersion" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "workflow"
}

func isDefaultVersion(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "DefaultVersion" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "workflow"
}

func extractChangeID(call *ast.CallExpr) string {
	if len(call.Args) < 2 {
		return ""
	}
	bl, ok := call.Args[1].(*ast.BasicLit)
	if !ok || bl.Kind.String() != "STRING" {
		return ""
	}
	// Strip surrounding quotes.
	if len(bl.Value) >= 2 && bl.Value[0] == '"' && bl.Value[len(bl.Value)-1] == '"' {
		return bl.Value[1 : len(bl.Value)-1]
	}
	return bl.Value
}

func isInsideIfInit(target *ast.CallExpr, file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if found || n == nil {
			return false
		}
		ifs, ok := n.(*ast.IfStmt)
		if !ok || ifs.Init == nil {
			return true
		}
		as, ok := ifs.Init.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range as.Rhs {
			if rhs == target {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
