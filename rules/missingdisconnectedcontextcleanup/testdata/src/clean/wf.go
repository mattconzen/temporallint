package clean

import (
	"go.temporal.io/sdk/workflow"
)

func Cleanup(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error {
	defer func() {
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		_ = workflow.ExecuteActivity(newCtx, Cleanup)
	}()
	return nil
}
