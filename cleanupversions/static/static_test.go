package static_test

import (
	"strings"
	"testing"

	"github.com/mattconzen/temporallint/cleanupversions"
	"github.com/mattconzen/temporallint/cleanupversions/static"
)

// TestCanonicalShape exercises the happy path: a canonical
// if-init GetVersion call produces a Candidate with a non-empty Edit
// and no Reason.
func TestCanonicalShape(t *testing.T) {
	src := `package wf

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
	got, err := static.DiscoverString("wf.go", src)
	if err != nil {
		t.Fatalf("DiscoverString: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 candidate, got %d", len(got))
	}
	c := got[0]
	if c.ChangeID != "swap-a" {
		t.Fatalf("ChangeID: got %q", c.ChangeID)
	}
	if c.Reason != "" {
		t.Fatalf("expected no Reason, got %q", c.Reason)
	}
	if c.Edit.NewBytes == nil {
		t.Fatal("expected non-empty Edit")
	}
	if !strings.Contains(string(c.Edit.NewBytes), `_ = "new"`) {
		t.Fatalf("Edit.NewBytes should contain the new branch; got %q", c.Edit.NewBytes)
	}
}

func TestNonLiteralChangeID(t *testing.T) {
	src := `package wf

import "go.temporal.io/sdk/workflow"

func WF(ctx workflow.Context, id string) error {
	if v := workflow.GetVersion(ctx, id, workflow.DefaultVersion, 1); v == workflow.DefaultVersion {
		_ = 1
	} else {
		_ = 2
	}
	return nil
}
`
	got, err := static.DiscoverString("wf.go", src)
	if err != nil {
		t.Fatalf("DiscoverString: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 candidate, got %d", len(got))
	}
	if got[0].ChangeID != "" || !strings.Contains(got[0].Reason, "non-literal change ID") {
		t.Fatalf("expected non-literal-change-ID skip, got %+v", got[0])
	}
}

func TestUnrecognisedShape(t *testing.T) {
	src := `package wf

import "go.temporal.io/sdk/workflow"

func WF(ctx workflow.Context) error {
	v := workflow.GetVersion(ctx, "swap-a", workflow.DefaultVersion, 1)
	if v == 0 {
		_ = 1
	}
	return nil
}
`
	got, err := static.DiscoverString("wf.go", src)
	if err != nil {
		t.Fatalf("DiscoverString: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 candidate, got %d", len(got))
	}
	if !strings.Contains(got[0].Reason, "unrecognised branching shape") {
		t.Fatalf("expected unrecognised-shape skip, got %+v", got[0])
	}
}

// TestNoElse covers the case where there's no else branch — the rule
// can't tell which branch is "new" vs "old", so it skips.
func TestNoElse(t *testing.T) {
	src := `package wf

import "go.temporal.io/sdk/workflow"

func WF(ctx workflow.Context) error {
	if v := workflow.GetVersion(ctx, "swap-a", workflow.DefaultVersion, 1); v == workflow.DefaultVersion {
		_ = 1
	}
	return nil
}
`
	got, err := static.DiscoverString("wf.go", src)
	if err != nil {
		t.Fatalf("DiscoverString: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 candidate, got %d", len(got))
	}
	if !strings.Contains(got[0].Reason, "missing or non-block else branch") {
		t.Fatalf("expected missing-else skip, got %+v", got[0])
	}
}

func TestNoMatch(t *testing.T) {
	src := `package wf

import "go.temporal.io/sdk/workflow"

func WF(ctx workflow.Context) error {
	_ = workflow.Now(ctx)
	return nil
}
`
	got, err := static.DiscoverString("wf.go", src)
	if err != nil {
		t.Fatalf("DiscoverString: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want no candidates, got %+v", got)
	}
}

// silence the cleanupversions import when no symbols are referenced
var _ cleanupversions.Decision = cleanupversions.DecisionSafe
