// Package noworkflowtimeout flags workflows whose WorkflowExecutionTimeout
// is unset (the server's 10-year default). Without a bounded timeout, a
// stuck workflow can sit in history forever and consume resources.
package noworkflowtimeout

import (
	"context"
	"fmt"

	"github.com/mattconzen/temporallint/all"
	"github.com/mattconzen/temporallint/runtime"
)

const (
	name = "runtime-no-workflow-timeout"
	url  = "https://github.com/jlegrone/100-temporal-mistakes#not-setting-a-workflow-timeout"
	doc  = "Flags workflows whose WorkflowExecutionTimeout is unset (server defaults to 10 years)."
)

type Check struct{}

func New() *Check                       { return &Check{} }
func (c *Check) Name() string           { return name }
func (c *Check) URL() string            { return url }
func (c *Check) Doc() string            { return doc }
func (c *Check) Category() all.Category { return all.CategoryTimeouts }

func (c *Check) Run(ctx context.Context, deps runtime.Deps) ([]runtime.Finding, error) {
	var findings []runtime.Finding
	q := runtime.ListQuery{Namespace: deps.Namespace, TaskQueue: deps.TaskQueue, Since: deps.Since}
	for {
		page, err := deps.API.ListWorkflows(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("list workflows: %w", err)
		}
		for _, exec := range page.Executions {
			if f, ok := evaluate(exec, deps.Thresholds); ok {
				findings = append(findings, f)
			}
		}
		if len(page.NextPageToken) == 0 {
			return findings, nil
		}
		q.PageToken = page.NextPageToken
	}
}

func evaluate(exec runtime.WorkflowExec, t runtime.Thresholds) (runtime.Finding, bool) {
	// We treat ExecutionTimeout==0 (unset) AND any value greater than the
	// configured ceiling as findings. The 10y default surfaces as 0 from
	// the adapter when the workflow never set an explicit value.
	if exec.ExecutionTimeout == 0 {
		return runtime.Finding{
			Check:    name,
			Severity: runtime.SeverityWarn,
			Subject:  exec.WorkflowID + "/" + exec.RunID,
			Message:  "WorkflowExecutionTimeout is unset; server default of 10 years applies",
			URL:      url,
		}, true
	}
	if t.NoWorkflowTimeoutAfter > 0 && exec.ExecutionTimeout > t.NoWorkflowTimeoutAfter {
		return runtime.Finding{
			Check:    name,
			Severity: runtime.SeverityFail,
			Subject:  exec.WorkflowID + "/" + exec.RunID,
			Message:  fmt.Sprintf("WorkflowExecutionTimeout %s exceeds configured ceiling %s", exec.ExecutionTimeout, t.NoWorkflowTimeoutAfter),
			URL:      url,
		}, true
	}
	return runtime.Finding{}, false
}
