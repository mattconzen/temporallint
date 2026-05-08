package violation

import (
	"go.temporal.io/sdk/client"
)

var _ = client.StartWorkflowOptions{ // want `neither WorkflowExecutionTimeout nor WorkflowRunTimeout`
	ID:        "wf-1",
	TaskQueue: "tq",
}
