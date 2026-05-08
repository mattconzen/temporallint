package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	sel := workflow.NewSelector(ctx)
	ch := workflow.GetSignalChannel(ctx, "tick")
	sel.AddReceive(ch, func(c workflow.Channel, more bool) {
		var v int
		c.Receive(ctx, &v)
	})
	sel.Select(ctx)
	return nil
}
