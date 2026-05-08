// Package temporaladapter wraps go.temporal.io/sdk/client.Client so it
// satisfies runtime.WorkflowAPI. This is the only place in temporallint
// that depends on the proto-heavy Temporal types — checks see only our
// narrow domain types.
package temporaladapter

import (
	"context"
	"fmt"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"github.com/mattconzen/monorepo/tools/temporallint/runtime"
)

// Adapter is a runtime.WorkflowAPI implementation backed by a real
// Temporal client.
type Adapter struct {
	Client    client.Client
	Namespace string
	PageSize  int32
}

// New constructs an Adapter. Callers retain ownership of c — closing it
// is their responsibility.
func New(c client.Client, namespace string) *Adapter {
	return &Adapter{Client: c, Namespace: namespace, PageSize: 100}
}

// ListWorkflows runs ListWorkflowExecutions with a query that filters
// to executions started within q.Since. TaskQueue is appended as an AND
// clause when set.
func (a *Adapter) ListWorkflows(ctx context.Context, q runtime.ListQuery) (runtime.ListResult, error) {
	query := fmt.Sprintf("StartTime >= '%s'", q.Since.UTC().Format(time.RFC3339))
	if q.TaskQueue != "" {
		query += fmt.Sprintf(" AND TaskQueue = '%s'", q.TaskQueue)
	}
	pageSize := a.PageSize
	if q.PageSize > 0 {
		pageSize = int32(q.PageSize)
	}
	resp, err := a.Client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Namespace:     a.Namespace,
		PageSize:      pageSize,
		NextPageToken: q.PageToken,
		Query:         query,
	})
	if err != nil {
		return runtime.ListResult{}, fmt.Errorf("list workflow: %w", err)
	}
	out := runtime.ListResult{NextPageToken: resp.NextPageToken}
	for _, info := range resp.Executions {
		out.Executions = append(out.Executions, toExec(info))
	}
	return out, nil
}

// DescribeWorkflow returns history bytes / length / timeouts for a
// single execution.
func (a *Adapter) DescribeWorkflow(ctx context.Context, workflowID, runID string) (runtime.WorkflowDescription, error) {
	resp, err := a.Client.DescribeWorkflowExecution(ctx, workflowID, runID)
	if err != nil {
		return runtime.WorkflowDescription{}, fmt.Errorf("describe workflow execution: %w", err)
	}
	desc := runtime.WorkflowDescription{
		WorkflowID: workflowID,
		RunID:      runID,
	}
	if info := resp.GetWorkflowExecutionInfo(); info != nil {
		desc.HistoryLength = info.GetHistoryLength()
		desc.HistorySizeBytes = info.GetHistorySizeBytes()
		if d := info.GetExecutionDuration(); d != nil {
			desc.RunTimeout = d.AsDuration()
		}
	}
	if cfg := resp.GetExecutionConfig(); cfg != nil {
		if d := cfg.GetWorkflowExecutionTimeout(); d != nil {
			desc.ExecutionTimeout = d.AsDuration()
		}
		if d := cfg.GetWorkflowRunTimeout(); d != nil {
			desc.RunTimeout = d.AsDuration()
		}
	}
	return desc, nil
}

// SampleHistory pulls events one-by-one until the iterator is exhausted
// or maxEvents is reached. Each event's payload size is approximated as
// the protobuf-encoded byte length so we don't decode user data.
func (a *Adapter) SampleHistory(ctx context.Context, workflowID, runID string, maxEvents int) ([]runtime.EventSummary, error) {
	iter := a.Client.GetWorkflowHistory(ctx, workflowID, runID, false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var out []runtime.EventSummary
	for iter.HasNext() && (maxEvents == 0 || len(out) < maxEvents) {
		ev, err := iter.Next()
		if err != nil {
			return out, fmt.Errorf("history next: %w", err)
		}
		out = append(out, runtime.EventSummary{
			EventID:     ev.GetEventId(),
			EventType:   ev.GetEventType().String(),
			PayloadSize: int64(payloadBytes(ev)),
		})
	}
	return out, nil
}

// payloadBytes is a conservative size estimator: it serializes the event
// to its proto-wire representation and returns the byte count. We avoid
// poking at typed payload fields (which vary by event type) and accept
// that the size includes some envelope overhead — for threshold purposes
// that's acceptable.
func payloadBytes(ev interface{ Marshal() ([]byte, error) }) int {
	if ev == nil {
		return 0
	}
	b, err := ev.Marshal()
	if err != nil {
		return 0
	}
	return len(b)
}

func toExec(info *workflowpb.WorkflowExecutionInfo) runtime.WorkflowExec {
	out := runtime.WorkflowExec{
		WorkflowID: info.GetExecution().GetWorkflowId(),
		RunID:      info.GetExecution().GetRunId(),
		TaskQueue:  info.GetTaskQueue(),
		Status:     info.GetStatus().String(),
	}
	if t := info.GetStartTime(); t != nil {
		out.StartTime = t.AsTime()
	}
	if d := info.GetExecutionDuration(); d != nil {
		out.RunTimeout = d.AsDuration()
	}
	return out
}
