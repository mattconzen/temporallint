package clean

import (
	"go.temporal.io/sdk/workflow"
)

func A1(workflow.Context) error { return nil }
func A2(workflow.Context) error { return nil }
func A3(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error {
	_ = workflow.ExecuteActivity(ctx, A1)
	_ = workflow.ExecuteActivity(ctx, A2)
	_ = workflow.ExecuteActivity(ctx, A3)
	return nil
}
