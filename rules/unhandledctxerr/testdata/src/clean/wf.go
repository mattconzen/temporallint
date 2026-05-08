package clean

import (
	"go.temporal.io/sdk/workflow"
)

func MyActivity(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error {
	f := workflow.ExecuteActivity(ctx, MyActivity)
	if err := f.Get(ctx, nil); err != nil {
		if cerr := ctx.Err(); cerr != nil {
			return cerr
		}
		return err
	}
	return nil
}
