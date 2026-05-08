package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	val := workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		return 42
	})
	_ = val
	return nil
}
