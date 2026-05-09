package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.ExecuteActivity(ctx, "hashOrEncodeOrCompile")
	return nil
}
