package runtime_test

import (
	"context"
	"testing"

	"github.com/mattconzen/monorepo/tools/temporallint/cleanupversions"
	cuvruntime "github.com/mattconzen/monorepo/tools/temporallint/cleanupversions/runtime"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/fakeapi"
)

func TestSafe(t *testing.T) {
	api := fakeapi.New().
		WithOpenExec(runtime.WorkflowExec{WorkflowID: "wf-1", RunID: "run-1"}).
		WithVersionMarkers("wf-1", "run-1", []runtime.VersionMarker{
			{ChangeID: "swap-a", Version: 1}, // already on new
		})

	v := &cuvruntime.Verifier{API: api, Namespace: "default"}
	reports, err := v.Verify(context.Background(), []cleanupversions.Candidate{
		{ChangeID: "swap-a", File: "x.go"},
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(reports) != 1 || reports[0].Decision != cleanupversions.DecisionSafe {
		t.Fatalf("expected SAFE, got %+v", reports)
	}
}

func TestUnsafe(t *testing.T) {
	api := fakeapi.New().
		WithOpenExec(runtime.WorkflowExec{WorkflowID: "wf-2", RunID: "run-2"}).
		WithVersionMarkers("wf-2", "run-2", []runtime.VersionMarker{
			{ChangeID: "swap-a", Version: 0}, // still on old
		})

	v := &cuvruntime.Verifier{API: api, Namespace: "default"}
	reports, err := v.Verify(context.Background(), []cleanupversions.Candidate{
		{ChangeID: "swap-a", File: "x.go"},
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if reports[0].Decision != cleanupversions.DecisionUnsafe {
		t.Fatalf("expected UNSAFE, got %+v", reports[0])
	}
	if len(reports[0].OpenWorkflows) == 0 {
		t.Fatal("expected non-empty OpenWorkflows sample on Unsafe verdict")
	}
}

func TestIndeterminate(t *testing.T) {
	api := fakeapi.New().
		WithOpenExec(runtime.WorkflowExec{WorkflowID: "wf-3", RunID: "run-3"})
	// No version markers registered → no workflow has recorded the change ID.
	v := &cuvruntime.Verifier{API: api, Namespace: "default"}
	reports, err := v.Verify(context.Background(), []cleanupversions.Candidate{
		{ChangeID: "swap-a", File: "x.go"},
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if reports[0].Decision != cleanupversions.DecisionIndeterminate {
		t.Fatalf("expected INDETERMINATE, got %+v", reports[0])
	}
}

func TestSkipPassthrough(t *testing.T) {
	api := fakeapi.New()
	v := &cuvruntime.Verifier{API: api, Namespace: "default"}
	reports, err := v.Verify(context.Background(), []cleanupversions.Candidate{
		{ChangeID: "", File: "x.go", Reason: "non-literal change ID"},
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if reports[0].Decision != cleanupversions.DecisionSkip {
		t.Fatalf("expected SKIP, got %+v", reports[0])
	}
}
