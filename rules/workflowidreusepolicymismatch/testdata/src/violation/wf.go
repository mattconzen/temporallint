package violation

import (
	"time"

	"go.temporal.io/sdk/client"
)

var _ = client.StartWorkflowOptions{ // want `no WorkflowIDReusePolicy`
	ID:                       "wf-1",
	TaskQueue:                "tq",
	WorkflowExecutionTimeout: time.Hour,
}
