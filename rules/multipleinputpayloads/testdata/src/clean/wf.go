package clean

import (
	"go.temporal.io/sdk/workflow"
)

type Input struct {
	A string
	B int
}

func WF(ctx workflow.Context, in Input) error {
	return nil
}

func NoArg(ctx workflow.Context) error {
	return nil
}
