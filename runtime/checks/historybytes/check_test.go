package historybytes_test

import (
	"context"
	"testing"
	"time"

	"github.com/mattconzen/temporallint/runtime"
	"github.com/mattconzen/temporallint/runtime/checks/historybytes"
	"github.com/mattconzen/temporallint/runtime/fakeapi"
	"github.com/mattconzen/temporallint/runtime/thresholds"
)

// TestViolation is the TDD-red half: synthetic API returns a workflow
// whose history is above the fail threshold. The check MUST emit a
// Severity=Fail finding.
func TestViolation(t *testing.T) {
	api := fakeapi.New().
		WithExec(runtime.WorkflowExec{WorkflowID: "wf-1", RunID: "run-1"}).
		WithDescription(runtime.WorkflowDescription{
			WorkflowID:       "wf-1",
			RunID:            "run-1",
			HistorySizeBytes: 300 << 20, // 300 MiB — above 200 MiB fail
		})

	got, err := historybytes.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %d: %+v", len(got), got)
	}
	if got[0].Severity != runtime.SeverityFail {
		t.Fatalf("want Severity=fail, got %q", got[0].Severity)
	}
	if got[0].Subject != "wf-1/run-1" {
		t.Fatalf("want Subject=wf-1/run-1, got %q", got[0].Subject)
	}
}

// TestWarn checks the warn-tier threshold.
func TestWarn(t *testing.T) {
	api := fakeapi.New().
		WithExec(runtime.WorkflowExec{WorkflowID: "wf-1", RunID: "run-1"}).
		WithDescription(runtime.WorkflowDescription{
			WorkflowID:       "wf-1",
			RunID:            "run-1",
			HistorySizeBytes: 60 << 20, // 60 MiB — between 50 warn and 200 fail
		})

	got, err := historybytes.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %d", len(got))
	}
	if got[0].Severity != runtime.SeverityWarn {
		t.Fatalf("want Severity=warn, got %q", got[0].Severity)
	}
}

// TestClean is the no-false-positive half: synthetic API returns a
// workflow whose history size is well below the warn threshold. The
// check MUST emit zero findings.
func TestClean(t *testing.T) {
	api := fakeapi.New().
		WithExec(runtime.WorkflowExec{WorkflowID: "wf-ok", RunID: "run-ok"}).
		WithDescription(runtime.WorkflowDescription{
			WorkflowID:       "wf-ok",
			RunID:            "run-ok",
			HistorySizeBytes: 1 << 20, // 1 MiB
		})

	got, err := historybytes.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want no findings, got %+v", got)
	}
}

// TestPagination verifies the check follows NextPageToken until the
// adapter signals end-of-stream.
func TestPagination(t *testing.T) {
	api := fakeapi.New().
		WithPagedExecs(
			[]runtime.WorkflowExec{{WorkflowID: "p1", RunID: "r1"}},
			[]runtime.WorkflowExec{{WorkflowID: "p2", RunID: "r2"}},
		).
		WithDescription(runtime.WorkflowDescription{WorkflowID: "p1", RunID: "r1", HistorySizeBytes: 1}).
		WithDescription(runtime.WorkflowDescription{WorkflowID: "p2", RunID: "r2", HistorySizeBytes: 300 << 20})

	got, err := historybytes.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 || got[0].Subject != "p2/r2" {
		t.Fatalf("expected p2/r2 to be flagged across pages; got %+v", got)
	}
}
