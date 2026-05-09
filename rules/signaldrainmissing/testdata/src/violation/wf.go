package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error { // want `signals delivered between the last loop iteration and return are lost`
	ch := workflow.GetSignalChannel(ctx, "sig")
	var msg string
	ch.Receive(ctx, &msg)
	return nil
}
