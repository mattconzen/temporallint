// Package individualpayload flags individual history events whose
// payload size exceeds configured thresholds. Temporal enforces a
// per-event size limit (4 MiB by default); approaching it risks event
// rejection and workflow stalls.
package individualpayload

import (
	"context"
	"fmt"

	"github.com/mattconzen/monorepo/tools/temporallint/all"
	"github.com/mattconzen/monorepo/tools/temporallint/runtime"
)

const (
	name = "runtime-individual-payload"
	url  = "https://github.com/jlegrone/100-temporal-mistakes#overflowing-maximum-individual-payload-size"
	doc  = "Samples workflow histories and flags any single event whose payload exceeds the warn / fail thresholds."
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
	maxEvents := deps.Thresholds.SampleHistoryEvents
	for {
		page, err := deps.API.ListWorkflows(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("list workflows: %w", err)
		}
		for _, exec := range page.Executions {
			events, err := deps.API.SampleHistory(ctx, exec.WorkflowID, exec.RunID, maxEvents)
			if err != nil {
				return nil, fmt.Errorf("sample history %s/%s: %w", exec.WorkflowID, exec.RunID, err)
			}
			for _, ev := range events {
				if f, ok := evaluate(exec, ev, deps.Thresholds); ok {
					findings = append(findings, f)
				}
			}
		}
		if len(page.NextPageToken) == 0 {
			return findings, nil
		}
		q.PageToken = page.NextPageToken
	}
}

func evaluate(exec runtime.WorkflowExec, ev runtime.EventSummary, t runtime.Thresholds) (runtime.Finding, bool) {
	switch {
	case ev.PayloadSize >= t.IndividualPayloadFail:
		return runtime.Finding{
			Check:    name,
			Severity: runtime.SeverityFail,
			Subject:  fmt.Sprintf("%s/%s event=%d (%s)", exec.WorkflowID, exec.RunID, ev.EventID, ev.EventType),
			Message:  fmt.Sprintf("payload %d bytes exceeds fail threshold %d", ev.PayloadSize, t.IndividualPayloadFail),
			URL:      url,
		}, true
	case ev.PayloadSize >= t.IndividualPayloadWarn:
		return runtime.Finding{
			Check:    name,
			Severity: runtime.SeverityWarn,
			Subject:  fmt.Sprintf("%s/%s event=%d (%s)", exec.WorkflowID, exec.RunID, ev.EventID, ev.EventType),
			Message:  fmt.Sprintf("payload %d bytes exceeds warn threshold %d", ev.PayloadSize, t.IndividualPayloadWarn),
			URL:      url,
		}, true
	}
	return runtime.Finding{}, false
}
