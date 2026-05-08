package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.UpsertSearchAttributes(ctx, map[string]interface{}{ // want `map\[string\]interface\{\}`
		"orderID": "abc",
		"qty":     42,
	})
	return nil
}
