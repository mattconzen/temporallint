package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	ch := workflow.GetSignalChannel(ctx, "sig")
	sel := workflow.NewSelector(ctx)
	sel.AddReceive(ch, func(c workflow.Channel, more bool) {
		var msg string
		c.Receive(ctx, &msg)
	})
	sel.Select(ctx)
	for sel.HasPending() {
		sel.Select(ctx)
	}
	return nil
}

func NoSignals(ctx workflow.Context) error {
	return nil
}
