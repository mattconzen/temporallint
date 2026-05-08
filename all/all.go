// Package all is the central registry for every analyzer shipped by
// temporallint. Each rule registers an Entry here so that the binary,
// the golangci-lint plugin, and the RULES.md generator all consume the
// same source of truth.
package all

import (
	"sort"

	"golang.org/x/tools/go/analysis"

	wfcheck "go.temporal.io/sdk/contrib/tools/workflowcheck/workflow"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/activitycallsexecuteactivity"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/activityignoresctxdone"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/activitymissingcontext"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/awaitnotimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/childworkflownotimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingdisconnectedcontextcleanup"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/maxattemptsone"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingheartbeattimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingrecordheartbeat"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingretrypolicy"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingstarttoclosetimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingworkflowtimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/nogracefuldrain"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/oversizedpayloadreturn"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/payloadanderror"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/pollingloopwithsleep"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/preventretriesbytimeout"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/queryhandlerwithsideeffects"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/searchattributetyping"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/sideeffectnoresult"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/signalchanneloutsideselector"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/signalhandlerblocksonactivity"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/startworkflowbadtaskqueue"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/startworkflowfromactivity"
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
	"github.com/mattconzen/monorepo/tools/temporallint/rules/terminatevscancel"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/toomanyactivitytypes"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/unboundedloopnocnaw"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/unboundednoceiling"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/unhandledctxerr"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/versioningwithoutgetversion"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/workernotaskqueue"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/workflowidreusepolicymismatch"
	"github.com/mattconzen/monorepo/tools/temporallint/rules/workflowretrypolicy"
)

// Status describes whether a documented mistake can be enforced statically.
type Status string

const (
	// StatusImplemented means an Analyzer is wired up and emits diagnostics.
	StatusImplemented Status = "Implemented"
	// StatusPlanned means the rule is on the roadmap but not yet implemented.
	StatusPlanned Status = "Planned"
	// StatusRuntimeOnly means the mistake is only visible at runtime
	// (history size, sync-match rate, STSL, etc.) so no static rule is possible.
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
		// --- Batch 7: Operations + soft-flag promotions ---
		{Name: nogracefuldrain.Analyzer.Name, Category: CategoryOperations, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-draining-activity-tasks-before-shutdown",
			Summary:    "worker.Run(nil) skips graceful drain on shutdown.",
			Analyzer:   nogracefuldrain.Analyzer},
		{Name: workernotaskqueue.Analyzer.Name, Category: CategoryOperations, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#starting-workflows-on-wrong-task-queue",
			Summary:    "worker.New requires a non-empty task queue.",
			Analyzer:   workernotaskqueue.Analyzer},
		{Name: startworkflowbadtaskqueue.Analyzer.Name, Category: CategoryOther, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#starting-workflows-on-wrong-task-queue",
			Summary:    "StartWorkflowOptions.TaskQueue should match a worker.New task queue (package-local).",
			Analyzer:   startworkflowbadtaskqueue.Analyzer},

		// --- Batch 6: Software design ---
		{Name: versioningwithoutgetversion.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-using-workflow-versioningpatching",
			Summary:    "Branching workflows that swap activities should use workflow.GetVersion (heuristic).",
			Analyzer:   versioningwithoutgetversion.Analyzer},
		{Name: signalhandlerblocksonactivity.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#assuming-signalsupdates-receive-in-order",
			Summary:    "Signal handler synchronously waits on ExecuteActivity.",
			Analyzer:   signalhandlerblocksonactivity.Analyzer},
		{Name: queryhandlerwithsideeffects.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#querying-closed-workflows",
			Summary:    "Query handlers must be pure.",
			Analyzer:   queryhandlerwithsideeffects.Analyzer},
		{Name: toomanyactivitytypes.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#doing-too-many-things-in-one-workflow",
			Summary:    "Workflow references too many distinct activity types (heuristic).",
			Analyzer:   toomanyactivitytypes.Analyzer},
		{Name: workflowidreusepolicymismatch.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-properly-scoping-semantic-workflow-ids",
			Summary:    "StartWorkflowOptions should set WorkflowIDReusePolicy explicitly (heuristic).",
			Analyzer:   workflowidreusepolicymismatch.Analyzer},
		{Name: searchattributetyping.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-properly-scoping-semantic-workflow-ids",
			Summary:    "UpsertSearchAttributes should use typed keys, not map[string]interface{}.",
			Analyzer:   searchattributetyping.Analyzer},
		{Name: terminatevscancel.Analyzer.Name, Category: CategoryOperations, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#terminating-rather-than-canceling",
			Summary:    "TerminateWorkflow skips graceful cancellation.",
			Analyzer:   terminatevscancel.Analyzer},
		{Name: startworkflowfromactivity.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#starting-workflows-from-activities",
			Summary:    "Activities should not start workflows.",
			Analyzer:   startworkflowfromactivity.Analyzer},

		// --- Batch 5: Cancellation & control flow ---
		{
			Name: unboundedloopnocnaw.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-using-continueasnew",
			Summary:    "Workflow with unbounded for{} loop must call ContinueAsNew.",
			Analyzer:   unboundedloopnocnaw.Analyzer,
		},
		{
			Name: pollingloopwithsleep.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#writing-polling-loops-in-workflow-code",
			Summary:    "Polling loops should use signals/timers, not workflow.Sleep (heuristic).",
			Analyzer:   pollingloopwithsleep.Analyzer,
		},
		{
			Name: awaitnotimeout.Analyzer.Name, Category: CategoryCancellation, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#deadlocking-when-workflow-canceled",
			Summary:    "workflow.Await without paired timer can stall forever.",
			Analyzer:   awaitnotimeout.Analyzer,
		},
		{
			Name: signalchanneloutsideselector.Analyzer.Name, Category: CategoryDesign, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#assuming-signalsupdates-receive-in-order",
			Summary:    "Direct GetSignalChannel.Receive should be inside a NewSelector (heuristic).",
			Analyzer:   signalchanneloutsideselector.Analyzer,
		},
		{
			Name: missingdisconnectedcontextcleanup.Analyzer.Name, Category: CategoryCancellation, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-using-disconnected-context-for-cleanup",
			Summary:    "Deferred ExecuteActivity must use NewDisconnectedContext (heuristic).",
			Analyzer:   missingdisconnectedcontextcleanup.Analyzer,
		},
		{
			Name: unhandledctxerr.Analyzer.Name, Category: CategoryCancellation, Status: StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#deadlocking-when-workflow-canceled",
			Summary:    "Workflow with Future.Get should check ctx.Err() (heuristic).",
			Analyzer:   unhandledctxerr.Analyzer,
		},

		// --- Batch 4: Activity shape ---
		{
			Name:       payloadanderror.Analyzer.Name,
			Category:   CategoryDesign,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#returning-both-payload-and-error",
			Summary:    "Activity returns must not provide both a non-zero payload and a non-nil error.",
			Analyzer:   payloadanderror.Analyzer,
		},
		{
			Name:       activitycallsexecuteactivity.Analyzer.Name,
			Category:   CategoryDesign,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#starting-workflows-from-activities",
			Summary:    "Activities cannot orchestrate other activities.",
			Analyzer:   activitycallsexecuteactivity.Analyzer,
		},
		{
			Name:       activitymissingcontext.Analyzer.Name,
			Category:   CategoryDesign,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-making-activities-idempotent",
			Summary:    "Activities must accept context.Context as first parameter.",
			Analyzer:   activitymissingcontext.Analyzer,
		},
		{
			Name:       activityignoresctxdone.Analyzer.Name,
			Category:   CategoryDesign,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-sending-heartbeats-from-activities",
			Summary:    "Activity loops must select on ctx.Done() or heartbeat (heuristic).",
			Analyzer:   activityignoresctxdone.Analyzer,
		},
		{
			Name:       missingrecordheartbeat.Analyzer.Name,
			Category:   CategoryDesign,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-sending-heartbeats-from-activities",
			Summary:    "Long activity loops should call activity.RecordHeartbeat (heuristic).",
			Analyzer:   missingrecordheartbeat.Analyzer,
		},
		{
			Name:       oversizedpayloadreturn.Analyzer.Name,
			Category:   CategoryDesign,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#passing-too-much-information-from-activities",
			Summary:    "Activity return type has too many fields (heuristic).",
			Analyzer:   oversizedpayloadreturn.Analyzer,
		},
		{
			Name:       sideeffectnoresult.Analyzer.Name,
			Category:   CategoryReplay,
			Status:     StatusImplemented,
			MistakeURL: "https://github.com/jlegrone/100-temporal-mistakes#not-using-return-value-in-side-effects",
			Summary:    "workflow.SideEffect result must be captured.",
			Analyzer:   sideeffectnoresult.Analyzer,
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
	case StatusPlanned:
		return 1
	case StatusRuntimeOnly:
		return 2
	case StatusDocOnly:
		return 3
	}
	return 4
}
