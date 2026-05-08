package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.ChildWorkflowOptions{ // want `neither WorkflowExecutionTimeout nor WorkflowRunTimeout`
		WorkflowID: "child-1",
		TaskQueue:  "tq",
	}
	return nil
}
