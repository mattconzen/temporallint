package noworkflowtimeout_test

import (
	"context"
	"testing"
	"time"

	"github.com/mattconzen/monorepo/tools/temporallint/runtime"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/checks/noworkflowtimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/internal/fakeapi"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime/thresholds"
)

func TestViolationUnset(t *testing.T) {
	api := fakeapi.New().WithExec(runtime.WorkflowExec{
		WorkflowID:       "wf-1",
		RunID:            "run-1",
		ExecutionTimeout: 0, // unset
	})

	got, err := noworkflowtimeout.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 || got[0].Severity != runtime.SeverityWarn {
		t.Fatalf("want Warn finding for unset timeout, got %+v", got)
	}
}

func TestViolationCeiling(t *testing.T) {
	api := fakeapi.New().WithExec(runtime.WorkflowExec{
		WorkflowID:       "wf-2",
		RunID:            "run-2",
		ExecutionTimeout: 2 * 365 * 24 * time.Hour, // 2 years; default ceiling is 1 year
	})

	got, err := noworkflowtimeout.New().Run(context.Background(), runtime.Deps{
		API: api, Namespace: "default", Since: time.Now().Add(-24 * time.Hour),
		Thresholds: thresholds.Defaults(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 || got[0].Severity != runtime.SeverityFail {
		t.Fatalf("want Fail finding for over-ceiling timeout, got %+v", got)
	}
}

func TestClean(t *testing.T) {
	api := fakeapi.New().WithExec(runtime.WorkflowExec{
		WorkflowID:       "wf-ok",
		RunID:            "run-ok",
		ExecutionTimeout: 30 * time.Minute,
	})

	got, err := noworkflowtimeout.New().Run(context.Background(), runtime.Deps{
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
