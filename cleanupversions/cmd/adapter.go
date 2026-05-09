package cmd

import (
	"context"
	"fmt"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"github.com/mattconzen/temporallint/runtime"
)

// versionAdapter implements runtime.VersionAPI on top of a real
// client.Client. It is independent of the runtime/internal/temporaladapter
// package because that package is `internal/` and can't be imported from
// outside the runtime subtree.
type versionAdapter struct {
	c         client.Client
	namespace string
}

func newAdapter(c client.Client, namespace string) *versionAdapter {
	return &versionAdapter{c: c, namespace: namespace}
}

// OpenWorkflows lists workflows whose ExecutionStatus is 'Running'.
func (a *versionAdapter) OpenWorkflows(ctx context.Context, q runtime.ListQuery) (runtime.ListResult, error) {
	resp, err := a.c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Namespace:     a.namespace,
		PageSize:      100,
		NextPageToken: q.PageToken,
		Query:         "ExecutionStatus = 'Running'",
	})
	if err != nil {
		return runtime.ListResult{}, fmt.Errorf("list open workflows: %w", err)
	}
	out := runtime.ListResult{NextPageToken: resp.NextPageToken}
	for _, info := range resp.Executions {
		out.Executions = append(out.Executions, runtime.WorkflowExec{
			WorkflowID: info.GetExecution().GetWorkflowId(),
			RunID:      info.GetExecution().GetRunId(),
			Status:     info.GetStatus().String(),
			TaskQueue:  info.GetTaskQueue(),
		})
	}
	return out, nil
}

// VersionMarkers walks a workflow's history and counts MarkerRecorded
// events whose name is "Version". Each entry returned is a stub: the
// (changeID, version) tuple is encoded in the marker's details payload
// using the SDK's converter pipeline, which we don't reconstruct here.
//
// This means the live adapter only knows that a marker was recorded,
// not which change ID — so production usage of cleanup-versions
// against a real server gives Indeterminate verdicts for every
// candidate. The fakeapi-based test path exercises the full classifier;
// integrating real marker decoding is tracked as a follow-up.
func (a *versionAdapter) VersionMarkers(ctx context.Context, workflowID, runID string) ([]runtime.VersionMarker, error) {
	iter := a.c.GetWorkflowHistory(ctx, workflowID, runID, false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var out []runtime.VersionMarker
	for iter.HasNext() {
		ev, err := iter.Next()
		if err != nil {
			return out, fmt.Errorf("history next: %w", err)
		}
		mr := ev.GetMarkerRecordedEventAttributes()
		if mr == nil || mr.GetMarkerName() != "Version" {
			continue
		}
		out = append(out, runtime.VersionMarker{ChangeID: "", Version: 0})
	}
	return out, nil
}
