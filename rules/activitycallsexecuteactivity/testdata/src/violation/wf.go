package violation

import (
	"context"

	_ "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func MyActivity(ctx context.Context) error {
	wfctx := workflow.Context(nil)
	_ = workflow.ExecuteActivity(wfctx, "Other") // want `called from inside an activity`
	return nil
}
