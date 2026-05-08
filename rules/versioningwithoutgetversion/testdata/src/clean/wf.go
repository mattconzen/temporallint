package clean

import (
	"go.temporal.io/sdk/workflow"
)

func OldA(workflow.Context) error { return nil }
func NewA(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error {
	if v := workflow.GetVersion(ctx, "swap-a", workflow.DefaultVersion, 1); v == workflow.DefaultVersion {
		_ = workflow.ExecuteActivity(ctx, OldA)
	} else {
		_ = workflow.ExecuteActivity(ctx, NewA)
	}
	return nil
}
