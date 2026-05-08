package individualpayload_test

import (
	"context"
	"testing"
	"time"

	"github.com/mattconzen/monorepo/tools/temporallint/runtime"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/checks/individualpayload"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/internal/fakeapi"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/thresholds"
)

func TestViolation(t *testing.T) {
	api := fakeapi.New().
		WithExec(runtime.WorkflowExec{WorkflowID: "wf-1", RunID: "run-1"}).
		WithHistory("wf-1", "run-1", []runtime.EventSummary{
			{EventID: 1, EventType: "WorkflowExecutionStarted", PayloadSize: 1024},
			{EventID: 17, EventType: "ActivityTaskScheduled", PayloadSize: 5 << 20}, // 5 MiB — above 4 MiB fail
		})

	got, err := individualpayload.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 || got[0].Severity != runtime.SeverityFail {
		t.Fatalf("want one Fail finding, got %+v", got)
	}
	if got[0].Subject != "wf-1/run-1 event=17 (ActivityTaskScheduled)" {
		t.Fatalf("subject mismatch: %q", got[0].Subject)
	}
}

func TestClean(t *testing.T) {
	api := fakeapi.New().
		WithExec(runtime.WorkflowExec{WorkflowID: "wf-ok", RunID: "run-ok"}).
		WithHistory("wf-ok", "run-ok", []runtime.EventSummary{
			{EventID: 1, EventType: "WorkflowExecutionStarted", PayloadSize: 256},
			{EventID: 2, EventType: "ActivityTaskScheduled", PayloadSize: 1024},
		})

	got, err := individualpayload.New().Run(context.Background(), runtime.Deps{
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
