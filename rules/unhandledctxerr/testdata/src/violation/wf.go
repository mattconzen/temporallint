package violation

import (
	"go.temporal.io/sdk/workflow"
)

func MyActivity(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error {
	f := workflow.ExecuteActivity(ctx, MyActivity)
	if err := f.Get(ctx, nil); err != nil { // want `never checks ctx.Err\(\)`
		return err
	}
	return nil
}
