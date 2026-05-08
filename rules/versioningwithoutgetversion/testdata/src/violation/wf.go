package violation

import (
	"go.temporal.io/sdk/workflow"
)

func OldA(workflow.Context) error { return nil }
func NewA(workflow.Context) error { return nil }

func WF(ctx workflow.Context, useNew bool) error {
	if useNew { // want `branches call different activities without workflow.GetVersion`
		_ = workflow.ExecuteActivity(ctx, NewA)
	} else {
		_ = workflow.ExecuteActivity(ctx, OldA)
	}
	return nil
}
