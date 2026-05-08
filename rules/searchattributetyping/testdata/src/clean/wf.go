package clean

import (
	"go.temporal.io/sdk/workflow"
)

// Stub typed-key helper to mimic temporal.NewSearchAttributeKey*.
type typedKey struct{ Name string }

func WF(ctx workflow.Context) error {
	// Real production code would use typed setters. The stub doesn't
	// model them yet; the point of this fixture is just that the linter
	// doesn't flag a non-untyped-map call.
	_ = workflow.Now(ctx)
	return nil
}
