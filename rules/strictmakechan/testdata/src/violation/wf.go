package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	ch := make(chan int, 1) // want `make\(chan T\) in workflow code`
	_ = ch
	return nil
}
