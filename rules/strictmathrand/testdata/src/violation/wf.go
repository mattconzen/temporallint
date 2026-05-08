package violation

import (
	"math/rand"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = rand.Intn(10)    // want `math/rand in workflow code`
	_ = rand.Float64()   // want `math/rand in workflow code`
	return nil
}
