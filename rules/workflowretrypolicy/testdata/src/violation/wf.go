package violation

import (
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var _ = client.StartWorkflowOptions{
	ID:                       "wf-1",
	TaskQueue:                "tq",
	WorkflowExecutionTimeout: time.Hour,
	RetryPolicy: &temporal.RetryPolicy{ // want `RetryPolicy retries the entire workflow`
		MaximumAttempts: 5,
	},
}
