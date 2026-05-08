// Package all is the central registry for every analyzer shipped by
// temporallint. Each rule registers an Entry here so that the binary,
// the golangci-lint plugin, and the RULES.md generator all consume the
// same source of truth.
package all

import (
	"sort"

	"golang.org/x/tools/go/analysis"

	wfcheck "go.temporal.io/sdk/contrib/tools/workflowcheck/workflow"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/childworkflownotimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/maxattemptsone"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingheartbeattimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingretrypolicy"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingstarttoclosetimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingworkflowtimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/preventretriesbytimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictglobalmutation"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictgokeyword"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictmakechan"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictmaprange"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictmathrand"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictnethttp"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictosexit"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictosgetenv"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictsyncprimitives"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/stricttimeafter"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/stricttimenow"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/stricttimesleep"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/unboundednoceiling"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/workflowretrypolicy"
)

// Status describes whether a documented mistake can be enforced statically.
type Status string

const (
	// StatusImplemented means a static Analyzer is wired up and emits diagnostics.
	StatusImplemented Status = "Implemented"
	// StatusRuntimeImplemented means a runtime check (under tools/temporallint/runtime)
	// is wired up. Static analysis cannot detect the mistake; the runtime
	// subcommand verifies it against a live Temporal server.
	StatusRuntimeImplemented Status = "RuntimeImplemented"
	// StatusPlanned means the rule is on the roadmap but not yet implemented.
	StatusPlanned Status = "Planned"
	// StatusRuntimeOnly means the mistake is only visible at runtime and no
	// runtime check has been written yet (most often because it requires
	// metrics scraping rather than SDK calls).
	StatusRuntimeOnly Status = "RuntimeOnly"
	// StatusDocOnly means the mistake is a design/process concern surfaced for
	// awareness but not enforced.
	StatusDocOnly Status = "DocOnly"
)

// Category groups rules in the generated RULES.md table.
type Category string

const (
	CategoryWorkflowLimits Category = "Workflow Limits"
	CategoryReplay         Category = "Workflow Replay"
	CategoryTimeouts       Category = "Timeouts & Retries"
	CategoryCancellation   Category = "Cancellation"
	CategoryDesign         Category = "Software Design"
	CategoryOperations     Category = "Operations"
	CategoryOther          Category = "Other"
)

// Entry is registry metadata for one rule. The Analyzer field may be nil
// for Planned / RuntimeOnly / DocOnly rules — those still appear in
// RULES.md so engineers can find them.
type Entry struct {
	Name       string
	Category   Category
	Status     Status
	MistakeURL string
	Summary    string
	Analyzer   *analysis.Analyzer
}

func registry() []Entry {
	return append(implemented(), planned()...)
}

func implemented() []Entry {
	return []Entry{
		{
			Name:       "workflowcheck",
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-using-static-analysissandboxed-sdk",
			Summary:    "Wraps Temporal's upstream call-graph non-determinism checker.",
			Analyzer:   wfcheck.NewChecker(wfcheck.Config{}).NewAnalyzer(),
		},
		{
			Name:       missingstarttoclosetimeout.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#preventing-activity-retries",
			Summary:    "ActivityOptions must set StartToCloseTimeout (or ScheduleToCloseTimeout).",
			Analyzer:   missingstarttoclosetimeout.Analyzer,
		},
		{
			Name:       stricttimenow.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#using-system-time-instead-of-workflow-time",
			Summary:    "Bans time.Now() inside workflow code; use workflow.Now(ctx).",
			Analyzer:   stricttimenow.Analyzer,
		},
		{
			Name:       strictosgetenv.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#reading-environment-variables-in-workflow-code",
			Summary:    "Bans os.Getenv inside workflow code.",
			Analyzer:   strictosgetenv.Analyzer,
		},
		{
			Name:       strictmathrand.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
			Summary:    "Bans math/rand inside workflow code; use workflow.SideEffect.",
			Analyzer:   strictmathrand.Analyzer,
		},
		{
			Name:       strictnethttp.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
			Summary:    "Bans net/http calls inside workflow code; use activities for I/O.",
			Analyzer:   strictnethttp.Analyzer,
		},
		{
			Name:       strictosexit.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
			Summary:    "Bans os.Exit / log.Fatal inside workflow code.",
			Analyzer:   strictosexit.Analyzer,
		},
		{
			Name:       strictgokeyword.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
			Summary:    "Bans bare `go` statements in workflow code; use workflow.Go.",
			Analyzer:   strictgokeyword.Analyzer,
		},
		{
			Name:       strictmakechan.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
			Summary:    "Bans make(chan T) in workflow code; use workflow.NewChannel.",
			Analyzer:   strictmakechan.Analyzer,
		},
		{
			Name:       stricttimesleep.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#using-system-time-instead-of-workflow-time",
			Summary:    "Bans time.Sleep in workflow code; use workflow.Sleep.",
			Analyzer:   stricttimesleep.Analyzer,
		},
		{
			Name:       stricttimeafter.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#using-system-time-instead-of-workflow-time",
			Summary:    "Bans time.After / time.Tick in workflow code; use workflow.NewTimer.",
			Analyzer:   stricttimeafter.Analyzer,
		},
		{
			Name:       strictsyncprimitives.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
			Summary:    "Bans sync.Mutex / WaitGroup / Once in workflow code.",
			Analyzer:   strictsyncprimitives.Analyzer,
		},
		{
			Name:       strictmaprange.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
			Summary:    "Flags `for range` over a map in workflow code (iteration order is random).",
			Analyzer:   strictmaprange.Analyzer,
		},
		// --- Batch 3: Timeouts & retries ---
		{
			Name:       missingretrypolicy.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#preventing-activity-retries",
			Summary:    "ActivityOptions should set RetryPolicy explicitly.",
			Analyzer:   missingretrypolicy.Analyzer,
		},
		{
			Name:       maxattemptsone.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#preventing-activity-retries",
			Summary:    "RetryPolicy.MaximumAttempts == 1 silently disables retry.",
			Analyzer:   maxattemptsone.Analyzer,
		},
		{
			Name:       unboundednoceiling.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#preventing-activity-retries",
			Summary:    "Unbounded RetryPolicy needs MaximumInterval ceiling.",
			Analyzer:   unboundednoceiling.Analyzer,
		},
		{
			Name:       preventretriesbytimeout.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#preventing-activity-retries",
			Summary:    "ScheduleToCloseTimeout shorter than RetryPolicy.InitialInterval prevents retries.",
			Analyzer:   preventretriesbytimeout.Analyzer,
		},
		{
			Name:       missingheartbeattimeout.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-sending-heartbeats-from-activities",
			Summary:    "Heartbeating activities require HeartbeatTimeout (heuristic).",
			Analyzer:   missingheartbeattimeout.Analyzer,
		},
		{
			Name:       missingworkflowtimeout.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-setting-a-workflow-timeout",
			Summary:    "StartWorkflowOptions should set a workflow-level timeout.",
			Analyzer:   missingworkflowtimeout.Analyzer,
		},
		{
			Name:       childworkflownotimeout.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-setting-a-workflow-timeout",
			Summary:    "ChildWorkflowOptions should set a workflow-level timeout.",
			Analyzer:   childworkflownotimeout.Analyzer,
		},
		{
			Name:       workflowretrypolicy.Analyzer.Name,
			Category:   CategoryTimeouts,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#using-workflow-retries",
			Summary:    "Avoid workflow-level RetryPolicy; retry inside activities.",
			Analyzer:   workflowretrypolicy.Analyzer,
		},
		{
			Name:       strictglobalmutation.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#modifying-shared-state-in-workflow-code",
			Summary:    "Flags writes to package-level variables from workflow code.",
			Analyzer:   strictglobalmutation.Analyzer,
		},
	}
}

// Analyzers returns the deduplicated list of live analyzers, sorted by name.
// This is what cmd/temporallint and the golangci-lint plugin consume.
func Analyzers() []*analysis.Analyzer {
	seen := map[string]struct{}{}
	var out []*analysis.Analyzer
	for _, e := range implemented() {
		if e.Analyzer == nil {
			continue
		}
		if _, ok := seen[e.Analyzer.Name]; ok {
			continue
		}
		seen[e.Analyzer.Name] = struct{}{}
		out = append(out, e.Analyzer)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Registry returns every entry — implemented and planned — for documentation
// generation. The order is stable: implemented rules first (by name), then
// planned/runtime/doc-only rules (by name).
func Registry() []Entry {
	all := registry()
	sort.SliceStable(all, func(i, j int) bool {
		if rank(all[i].Status) != rank(all[j].Status) {
			return rank(all[i].Status) < rank(all[j].Status)
		}
		return all[i].Name < all[j].Name
	})
	return all
}

func rank(s Status) int {
	switch s {
	case StatusImplemented:
		return 0
	case StatusRuntimeImplemented:
		return 1
	case StatusPlanned:
		return 2
	case StatusRuntimeOnly:
		return 3
	case StatusDocOnly:
		return 4
	}
	return 5
}
