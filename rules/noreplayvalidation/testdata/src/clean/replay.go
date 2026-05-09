package clean

import (
	"go.temporal.io/sdk/worker"
)

// VerifyReplay exercises a recorded history through the replay worker.
// Many repos run this from a TestMain or a CI-only entry point so it
// lives in a regular .go file rather than _test.go; either is fine —
// the rule looks for the replay API call anywhere in the package.
func VerifyReplay() error {
	r := worker.NewWorkflowReplayer()
	r.RegisterWorkflow(WF)
	return r.ReplayWorkflowHistory(nil, nil)
}
