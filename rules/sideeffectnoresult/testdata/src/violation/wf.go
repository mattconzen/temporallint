package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	workflow.SideEffect(ctx, func(workflow.Context) interface{} { // want `result discarded`
		return 42
	})
	return nil
}
