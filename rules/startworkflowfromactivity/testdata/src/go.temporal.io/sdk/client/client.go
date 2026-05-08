// Stub implementation of go.temporal.io/sdk/client for analysistest
// fixtures.
package client

import "time"

type StartWorkflowOptions struct {
	ID                       string
	TaskQueue                string
	WorkflowExecutionTimeout time.Duration
	WorkflowRunTimeout       time.Duration
	WorkflowTaskTimeout      time.Duration
	WorkflowIDReusePolicy    int
	RetryPolicy              interface{}
}

type Client interface {
	Close()
	TerminateWorkflow(ctx interface{}, workflowID, runID, reason string, details ...interface{}) error
	CancelWorkflow(ctx interface{}, workflowID, runID string) error
	ExecuteWorkflow(ctx interface{}, options StartWorkflowOptions, workflow interface{}, args ...interface{}) (interface{}, error)
	SignalWithStartWorkflow(ctx interface{}, workflowID, signalName string, signalArg interface{}, options StartWorkflowOptions, workflow interface{}, args ...interface{}) (interface{}, error)
}

const (
	WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE             = 0
	WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE            = 1
	WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY = 2
)
