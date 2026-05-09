package clean

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.ChildWorkflowOptions{
		WorkflowID:               "child-1",
		WorkflowExecutionTimeout: time.Hour,
		ParentClosePolicy:        1,
	}
	return nil
}
