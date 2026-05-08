package clean

import (
	"time"

	"go.temporal.io/sdk/client"
)

var _ = client.StartWorkflowOptions{
	ID:                       "wf-1",
	TaskQueue:                "tq",
	WorkflowExecutionTimeout: time.Hour,
}
