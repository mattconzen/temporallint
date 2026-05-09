package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context, a string, b int) error { // want `takes 3 parameters; use a single input struct`
	return nil
}

func WF2(ctx workflow.Context, a, b string) error { // want `takes 3 parameters; use a single input struct`
	return nil
}
