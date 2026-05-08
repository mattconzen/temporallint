package clean

import (
	"go.temporal.io/sdk/workflow"
)

func MyActivity(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error {
	jobs := workflow.NewChannel(ctx)
	_ = workflow.RegisterSignalHandler(ctx, "go", func(arg int) {
		jobs.Send(ctx, arg) // dispatch via channel
	})
	workflow.Go(ctx, func(c workflow.Context) {
		var arg int
		jobs.Receive(c, &arg)
		_ = workflow.ExecuteActivity(c, MyActivity).Get(c, nil)
	})
	return nil
}
