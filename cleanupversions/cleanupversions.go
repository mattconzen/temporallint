// Package cleanupversions implements the `temporallint cleanup-versions`
// subcommand: a hybrid static + runtime tool that finds obsolete
// workflow.GetVersion calls and (optionally) auto-removes them.
//
// Three phases:
//
//  1. Static discovery: walk Go packages, find every workflow.GetVersion
//     call, classify the surrounding branching shape, and capture an
//     edit plan if the shape is the canonical
//     `if v := workflow.GetVersion(...); v == workflow.DefaultVersion {
//          // old branch
//      } else {
//          // new branch
//      }`
//     pattern. Other shapes are reported as Skip.
//
//  2. Runtime verification: for each candidate change ID, query the
//     Temporal server for open workflows that have recorded a marker
//     for that ID. A change is Safe iff zero open workflows recorded
//     version 0 (the old branch).
//
//  3. Reporting and (optional) rewrite: emit a text/JSON report. When
//     --apply is set AND every candidate is Safe, atomically rewrite
//     the source files.
package cleanupversions

import "go/token"

// Decision is the safety verdict for a single candidate.
type Decision string

const (
	// DecisionSafe means no open workflow recorded the marker at version 0.
	DecisionSafe Decision = "safe"
	// DecisionUnsafe means at least one open workflow recorded version 0.
	DecisionUnsafe Decision = "unsafe"
	// DecisionIndeterminate means we couldn't reach a verdict (server
	// scan partial, no markers seen, etc.). Treated as Unsafe by --apply.
	DecisionIndeterminate Decision = "indeterminate"
	// DecisionSkip means the static phase didn't recognise the shape.
	DecisionSkip Decision = "skip"
)

// Candidate is one workflow.GetVersion call discovered by the static
// phase, plus enough metadata to (a) report on it and (b) rewrite the
// source if --apply is set.
type Candidate struct {
	ChangeID string
	File     string
	Pos      token.Position
	Reason   string // populated when Decision is Skip
	// Edit captures the byte-range and replacement for the rewrite.
	// Empty when shape isn't supported.
	Edit Edit
}

// Edit is a byte-level replacement to apply if the candidate is Safe
// AND --apply is set. The static phase produces these; the rewriter
// applies them.
type Edit struct {
	StartOffset int    // 0-indexed byte offset in source file
	EndOffset   int    // exclusive
	NewBytes    []byte // replacement bytes (often the kept branch's body)
}

// SafetyReport is the per-candidate verdict combining the static
// discovery result with the runtime verification result.
type SafetyReport struct {
	Candidate Candidate `json:"candidate"`
	Decision  Decision  `json:"decision"`
	Detail    string    `json:"detail,omitempty"`
	// OpenWorkflows is a sample of workflow IDs that referenced the
	// change ID, useful for "why is this Unsafe?" debugging.
	OpenWorkflows []string `json:"open_workflows,omitempty"`
}

// Report is the top-level JSON structure emitted by the subcommand.
type Report struct {
	Reports []SafetyReport `json:"reports"`
	Applied bool           `json:"applied"`
}
