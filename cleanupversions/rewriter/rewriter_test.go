package rewriter_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattconzen/temporallint/cleanupversions"
	"github.com/mattconzen/temporallint/cleanupversions/rewriter"
	"github.com/mattconzen/temporallint/cleanupversions/static"
)

const wfSrc = `package wf

import "go.temporal.io/sdk/workflow"

func WF(ctx workflow.Context) error {
	if v := workflow.GetVersion(ctx, "swap-a", workflow.DefaultVersion, 1); v == workflow.DefaultVersion {
		_ = "old"
	} else {
		_ = "new"
	}
	return nil
}
`

func TestApplyHappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wf.go")
	if err := os.WriteFile(path, []byte(wfSrc), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	candidates, err := static.DiscoverString(path, wfSrc)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("want 1 candidate, got %d", len(candidates))
	}
	reports := []cleanupversions.SafetyReport{{
		Candidate: candidates[0],
		Decision:  cleanupversions.DecisionSafe,
	}}

	touched, err := rewriter.Apply(reports)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(touched) != 1 || touched[0] != path {
		t.Fatalf("expected one touched file %q, got %v", path, touched)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	out := string(got)
	// The rewritten file must keep the new branch and drop the GetVersion call.
	if strings.Contains(out, "GetVersion") {
		t.Fatalf("expected GetVersion to be removed; got:\n%s", out)
	}
	if !strings.Contains(out, `_ = "new"`) {
		t.Fatalf("expected new branch body to be present; got:\n%s", out)
	}
	if strings.Contains(out, `_ = "old"`) {
		t.Fatalf("expected old branch body to be gone; got:\n%s", out)
	}
}

func TestApplySkipsNonSafe(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wf.go")
	if err := os.WriteFile(path, []byte(wfSrc), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	candidates, err := static.DiscoverString(path, wfSrc)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	reports := []cleanupversions.SafetyReport{{
		Candidate: candidates[0],
		Decision:  cleanupversions.DecisionUnsafe,
	}}
	touched, err := rewriter.Apply(reports)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(touched) != 0 {
		t.Fatalf("expected no files touched; got %v", touched)
	}
	got, _ := os.ReadFile(path)
	if string(got) != wfSrc {
		t.Fatal("file was modified despite Unsafe verdict")
	}
}
