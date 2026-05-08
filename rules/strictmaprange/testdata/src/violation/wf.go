package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	m := map[string]int{"a": 1, "b": 2}
	for k, v := range m { // want `for range over a map in workflow code`
		_ = k
		_ = v
	}
	return nil
}
