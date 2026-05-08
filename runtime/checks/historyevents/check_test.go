package historyevents_test

import (
	"context"
	"testing"
	"time"

	"github.com/mattconzen/monorepo/tools/temporallint/runtime"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/checks/historyevents"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/internal/fakeapi"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/thresholds"
)

func TestViolation(t *testing.T) {
	api := fakeapi.New().
		WithExec(runtime.WorkflowExec{WorkflowID: "wf-1", RunID: "run-1"}).
		WithDescription(runtime.WorkflowDescription{
			WorkflowID:    "wf-1",
			RunID:         "run-1",
			HistoryLength: 50_001, // above 50k fail
		})

	got, err := historyevents.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 || got[0].Severity != runtime.SeverityFail {
		t.Fatalf("want one Fail finding, got %+v", got)
	}
}

func TestWarn(t *testing.T) {
	api := fakeapi.New().
		WithExec(runtime.WorkflowExec{WorkflowID: "wf-1", RunID: "run-1"}).
		WithDescription(runtime.WorkflowDescription{
			WorkflowID:    "wf-1",
			RunID:         "run-1",
			HistoryLength: 12_000, // between 10k warn and 50k fail
		})

	got, err := historyevents.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 || got[0].Severity != runtime.SeverityWarn {
		t.Fatalf("want one Warn finding, got %+v", got)
	}
}

func TestClean(t *testing.T) {
	api := fakeapi.New().
		WithExec(runtime.WorkflowExec{WorkflowID: "wf-ok", RunID: "run-ok"}).
		WithDescription(runtime.WorkflowDescription{
			WorkflowID:    "wf-ok",
			RunID:         "run-ok",
			HistoryLength: 42, // tiny
		})

	got, err := historyevents.New().Run(context.Background(), runtime.Deps{
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
