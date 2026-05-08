package violation

import (
	"go.temporal.io/sdk/workflow"
)

func Cleanup(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error {
	defer func() { // want `deferred workflow.ExecuteActivity uses the parent context`
		_ = workflow.ExecuteActivity(ctx, Cleanup)
	}()
	return nil
}
