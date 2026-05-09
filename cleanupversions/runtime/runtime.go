// Package runtime implements Phase 2 of the cleanup-versions
// subcommand: querying the Temporal server for open workflows that
// might still replay through the GetVersion call we want to remove.
package runtime

import (
	"context"
	"fmt"
	"sort"

	"github.com/mattconzen/temporallint/cleanupversions"
	rt "github.com/mattconzen/temporallint/runtime"
)

// Verifier classifies each Candidate as Safe / Unsafe / Indeterminate
// using a runtime.VersionAPI.
type Verifier struct {
	API       rt.VersionAPI
	Namespace string
	MaxOpen   int // 0 = unlimited
}

// Verify pages through OpenWorkflows, fetches every workflow's version
// markers, and produces a SafetyReport per Candidate. Skip candidates
// (no edit possible) bypass the runtime check entirely.
func (v *Verifier) Verify(ctx context.Context, candidates []cleanupversions.Candidate) ([]cleanupversions.SafetyReport, error) {
	openByChange, scanned, err := v.scanOpenMarkers(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan open workflows: %w", err)
	}

	reports := make([]cleanupversions.SafetyReport, 0, len(candidates))
	for _, c := range candidates {
		// Skip-class candidates pass through with their static reason.
		if c.Reason != "" {
			reports = append(reports, cleanupversions.SafetyReport{
				Candidate: c, Decision: cleanupversions.DecisionSkip,
				Detail: c.Reason,
			})
			continue
		}
		decision, detail, sample := classify(c.ChangeID, openByChange, scanned)
		reports = append(reports, cleanupversions.SafetyReport{
			Candidate:     c,
			Decision:      decision,
			Detail:        detail,
			OpenWorkflows: sample,
		})
	}
	return reports, nil
}

// markerSummary holds, per change ID, the set of open workflows that
// referenced it and the worst (lowest) recorded version.
type markerSummary struct {
	workflows map[string]struct{}
	minVer    int
	hasZero   bool
}

func (v *Verifier) scanOpenMarkers(ctx context.Context) (map[string]*markerSummary, int, error) {
	out := map[string]*markerSummary{}
	scanned := 0
	q := rt.ListQuery{Namespace: v.Namespace}
	for {
		page, err := v.API.OpenWorkflows(ctx, q)
		if err != nil {
			return nil, scanned, err
		}
		for _, exec := range page.Executions {
			if v.MaxOpen > 0 && scanned >= v.MaxOpen {
				return out, scanned, fmt.Errorf("open-workflow scan exceeded MaxOpen=%d", v.MaxOpen)
			}
			scanned++
			markers, err := v.API.VersionMarkers(ctx, exec.WorkflowID, exec.RunID)
			if err != nil {
				return nil, scanned, fmt.Errorf("markers %s/%s: %w", exec.WorkflowID, exec.RunID, err)
			}
			for _, m := range markers {
				summary, ok := out[m.ChangeID]
				if !ok {
					summary = &markerSummary{workflows: map[string]struct{}{}, minVer: m.Version}
					out[m.ChangeID] = summary
				}
				summary.workflows[exec.WorkflowID+"/"+exec.RunID] = struct{}{}
				if m.Version == 0 {
					summary.hasZero = true
				}
				if m.Version < summary.minVer {
					summary.minVer = m.Version
				}
			}
		}
		if len(page.NextPageToken) == 0 {
			return out, scanned, nil
		}
		q.PageToken = page.NextPageToken
	}
}

func classify(changeID string, openByChange map[string]*markerSummary, scanned int) (cleanupversions.Decision, string, []string) {
	summary, ok := openByChange[changeID]
	if !ok {
		// No open workflow recorded this change ID at all. Two cases:
		//  - all workflows have already passed the new branch: SAFE
		//  - workflows haven't reached the GetVersion yet: UNSAFE
		// We can't distinguish from outside, so return Indeterminate.
		// When scanned == 0 we also can't tell — same verdict.
		return cleanupversions.DecisionIndeterminate,
			fmt.Sprintf("no open workflow has recorded change ID %q (scanned %d)", changeID, scanned),
			nil
	}
	if summary.hasZero {
		return cleanupversions.DecisionUnsafe,
			fmt.Sprintf("%d open workflow(s) recorded change ID %q at version 0 (old branch)",
				countZeroWorkflows(summary), changeID),
			sampleWorkflows(summary)
	}
	return cleanupversions.DecisionSafe,
		fmt.Sprintf("%d open workflow(s) referenced change ID %q; all at version >= 1",
			len(summary.workflows), changeID), nil
}

func countZeroWorkflows(s *markerSummary) int {
	// We don't track per-workflow versions, only minVer. If hasZero is
	// true, at least one workflow is on version 0; we can't break it
	// down further without retaining more state. Return the size of
	// the workflows set as an upper bound.
	return len(s.workflows)
}

func sampleWorkflows(s *markerSummary) []string {
	const limit = 5
	out := make([]string, 0, limit)
	for k := range s.workflows {
		out = append(out, k)
		if len(out) == limit {
			break
		}
	}
	sort.Strings(out)
	return out
}
