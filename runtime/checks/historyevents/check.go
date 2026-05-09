// Package historyevents flags workflows whose history event count
// exceeds configured thresholds. Temporal's server hard cap is 51200
// events; running close to that risks workflow termination on the next
// event append.
package historyevents

import (
	"context"
	"fmt"

	"github.com/mattconzen/temporallint/all"
	"github.com/mattconzen/temporallint/runtime"
)

const (
	name = "runtime-history-events"
	url  = "https://github.com/jlegrone/100-temporal-mistakes#overflowing-workflow-history-length"
	doc  = "Flags workflows whose history event count is approaching the server hard cap (51200)."
)

type Check struct{}

func New() *Check                       { return &Check{} }
func (c *Check) Name() string           { return name }
func (c *Check) URL() string            { return url }
func (c *Check) Doc() string            { return doc }
func (c *Check) Category() all.Category { return all.CategoryWorkflowLimits }

func (c *Check) Run(ctx context.Context, deps runtime.Deps) ([]runtime.Finding, error) {
	var findings []runtime.Finding
	q := runtime.ListQuery{Namespace: deps.Namespace, TaskQueue: deps.TaskQueue, Since: deps.Since}
	for {
		page, err := deps.API.ListWorkflows(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("list workflows: %w", err)
		}
		for _, exec := range page.Executions {
			desc, err := deps.API.DescribeWorkflow(ctx, exec.WorkflowID, exec.RunID)
			if err != nil {
				return nil, fmt.Errorf("describe %s/%s: %w", exec.WorkflowID, exec.RunID, err)
			}
			if f, ok := evaluate(desc, deps.Thresholds); ok {
				findings = append(findings, f)
			}
		}
		if len(page.NextPageToken) == 0 {
			return findings, nil
		}
		q.PageToken = page.NextPageToken
	}
}

func evaluate(desc runtime.WorkflowDescription, t runtime.Thresholds) (runtime.Finding, bool) {
	switch {
	case desc.HistoryLength >= t.HistoryEventsFail:
		return runtime.Finding{
			Check:    name,
			Severity: runtime.SeverityFail,
			Subject:  desc.WorkflowID + "/" + desc.RunID,
			Message:  fmt.Sprintf("history length %d events exceeds fail threshold %d (server hard cap is 51200)", desc.HistoryLength, t.HistoryEventsFail),
			URL:      url,
		}, true
	case desc.HistoryLength >= t.HistoryEventsWarn:
		return runtime.Finding{
			Check:    name,
			Severity: runtime.SeverityWarn,
			Subject:  desc.WorkflowID + "/" + desc.RunID,
			Message:  fmt.Sprintf("history length %d events exceeds warn threshold %d", desc.HistoryLength, t.HistoryEventsWarn),
			URL:      url,
		}, true
	}
	return runtime.Finding{}, false
}
