package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.ChildWorkflowOptions{ // want `no ParentClosePolicy`
		WorkflowID:               "child-1",
		WorkflowExecutionTimeout: time.Hour,
	}
	return nil
}
