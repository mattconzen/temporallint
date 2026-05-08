package violation

import (
	"go.temporal.io/sdk/workflow"
)

func MyActivity(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error {
	_ = workflow.RegisterSignalHandler(ctx, "go", func(arg int) {
		_ = workflow.ExecuteActivity(ctx, MyActivity).Get(ctx, nil) // want `synchronously waits on workflow.ExecuteActivity`
	})
	return nil
}
