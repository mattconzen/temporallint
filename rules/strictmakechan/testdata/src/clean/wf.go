package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	ch := workflow.NewChannel(ctx)
	_ = ch
	return nil
}
