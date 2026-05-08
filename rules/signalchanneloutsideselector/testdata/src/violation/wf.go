package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	var v int
	workflow.GetSignalChannel(ctx, "tick").Receive(ctx, &v) // want `outside a workflow.NewSelector`
	return nil
}
