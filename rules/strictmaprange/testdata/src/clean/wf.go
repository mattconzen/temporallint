package clean

import (
	"sort"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	m := map[string]int{"a": 1, "b": 2}

	// Correct pattern: extract keys, sort, then range over the slice.
	keys := sortedKeys(m)
	for _, k := range keys {
		_ = m[k]
	}

	// Slices and arrays are deterministic — no diagnostic expected.
	for _, x := range []int{1, 2, 3} {
		_ = x
	}
	return nil
}

// sortedKeys runs at workflow scope but the for-range inside is over a
// SLICE (the result of `for k := range m` is treated as a map iteration —
// which we DO want to flag, but only inside workflow funcs. This helper
// is not a workflow.Context-receiving function, so the analyzer skips it.)
func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
