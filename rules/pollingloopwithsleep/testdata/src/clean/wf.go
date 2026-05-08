package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	ch := workflow.GetSignalChannel(ctx, "tick")
	for i := 0; i < 100; i++ {
		var v int
		ch.Receive(ctx, &v) // signal-driven, no Sleep in loop
	}
	return nil
}
