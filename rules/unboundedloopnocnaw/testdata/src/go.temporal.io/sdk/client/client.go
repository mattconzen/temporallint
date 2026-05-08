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
}
