package violation

import (
	"go.temporal.io/sdk/workflow"
)

func A1(workflow.Context) error { return nil }
func A2(workflow.Context) error { return nil }
func A3(workflow.Context) error { return nil }
func A4(workflow.Context) error { return nil }
func A5(workflow.Context) error { return nil }
func A6(workflow.Context) error { return nil }
func A7(workflow.Context) error { return nil }
func A8(workflow.Context) error { return nil }
func A9(workflow.Context) error { return nil }

func WF(ctx workflow.Context) error { // want `references 9 distinct activity types`
	_ = workflow.ExecuteActivity(ctx, A1)
	_ = workflow.ExecuteActivity(ctx, A2)
	_ = workflow.ExecuteActivity(ctx, A3)
	_ = workflow.ExecuteActivity(ctx, A4)
	_ = workflow.ExecuteActivity(ctx, A5)
	_ = workflow.ExecuteActivity(ctx, A6)
	_ = workflow.ExecuteActivity(ctx, A7)
	_ = workflow.ExecuteActivity(ctx, A8)
	_ = workflow.ExecuteActivity(ctx, A9)
	return nil
}
