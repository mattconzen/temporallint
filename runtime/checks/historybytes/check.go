// Package historybytes flags workflows whose history size in bytes
// exceeds configured thresholds. Large histories slow down replays,
// stretch worker memory, and ultimately hit Temporal's hard size limit.
package historybytes

import (
	"context"
	"fmt"

	"github.com/mattconzen/temporallint/all"
	"github.com/mattconzen/temporallint/runtime"
)

const (
	name = "runtime-history-bytes"
	url  = "https://github.com/jlegrone/100-temporal-mistakes#overflowing-workflow-history-bytes"
	doc  = "Flags workflows whose history size in bytes exceeds the warn / fail thresholds."
)

// Check is the runtime check for history-byte overflow.
type Check struct{}

func New() *Check                    { return &Check{} }
func (c *Check) Name() string        { return name }
func (c *Check) URL() string         { return url }
func (c *Check) Doc() string         { return doc }
func (c *Check) Category() all.Category {
	return all.CategoryWorkflowLimits
}

// Run lists workflows in the time window and DescribeWorkflowExecution's
// each one to read HistorySizeBytes. Workflows below the warn threshold
// are silently OK; otherwise a Finding is emitted.
func (c *Check) Run(ctx context.Context, deps runtime.Deps) ([]runtime.Finding, error) {
	var findings []runtime.Finding
	q := runtime.ListQuery{
		Namespace: deps.Namespace,
		TaskQueue: deps.TaskQueue,
		Since:     deps.Since,
	}
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
	case desc.HistorySizeBytes >= t.HistoryBytesFail:
		return runtime.Finding{
			Check:    name,
			Severity: runtime.SeverityFail,
			Subject:  desc.WorkflowID + "/" + desc.RunID,
			Message:  fmt.Sprintf("history size %d bytes exceeds fail threshold %d", desc.HistorySizeBytes, t.HistoryBytesFail),
			URL:      url,
		}, true
	case desc.HistorySizeBytes >= t.HistoryBytesWarn:
		return runtime.Finding{
			Check:    name,
			Severity: runtime.SeverityWarn,
			Subject:  desc.WorkflowID + "/" + desc.RunID,
			Message:  fmt.Sprintf("history size %d bytes exceeds warn threshold %d", desc.HistorySizeBytes, t.HistoryBytesWarn),
			URL:      url,
		}, true
	}
	return runtime.Finding{}, false
}
