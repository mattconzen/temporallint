// Package runtime defines the contract every Temporal runtime check must
// satisfy. Static analyzers (the rest of temporallint) catch mistakes at
// compile time; runtime checks catch the ones that only show up by
// inspecting live Temporal server state — history size, payload size,
// missing workflow timeouts, etc.
//
// The package deliberately exposes a NARROW WorkflowAPI interface rather
// than the full go.temporal.io/sdk/client.Client. That keeps each check's
// dependencies small, makes test doubles trivial to write, and isolates
// the proto-heavy SDK types behind one adapter at the CLI boundary.
package runtime

import (
	"context"
	"time"

	"github.com/mattconzen/temporallint/all"
)

// Severity orders findings by how badly they should fail a CI run.
type Severity string

const (
	SeverityOK   Severity = "ok"
	SeverityWarn Severity = "warn"
	SeverityFail Severity = "fail"
)

// Rank gives Severity a comparable integer ordering. Higher = worse.
func (s Severity) Rank() int {
	switch s {
	case SeverityFail:
		return 2
	case SeverityWarn:
		return 1
	}
	return 0
}

// Finding is one check result. Subject identifies the Temporal entity
// that triggered the finding (workflowID/runID, or namespace=default for
// scope-wide findings); Message is human-readable; URL points back at
// the source-of-truth doc in 100-temporal-mistakes.
type Finding struct {
	Check    string   `json:"check"`
	Severity Severity `json:"severity"`
	Subject  string   `json:"subject"`
	Message  string   `json:"message"`
	URL      string   `json:"url,omitempty"`
}

// WorkflowAPI is the narrow slice of Temporal SDK functionality that
// every Tier-A runtime check needs. Real production code uses an adapter
// over client.Client; tests use a synthetic implementation.
type WorkflowAPI interface {
	// ListWorkflows returns workflows started within the time window.
	// The implementation paginates internally and yields one batch.
	// PageToken handling lives behind the adapter; checks that need
	// large windows iterate via repeated calls.
	ListWorkflows(ctx context.Context, q ListQuery) (ListResult, error)

	// DescribeWorkflow returns history bytes / length / executionTimeout
	// for a single execution.
	DescribeWorkflow(ctx context.Context, workflowID, runID string) (WorkflowDescription, error)

	// SampleHistory pulls up to maxEvents from the history of a single
	// execution, returning a per-event summary (size, type). Used by
	// the individual-payload check; checks that just need totals call
	// DescribeWorkflow instead.
	SampleHistory(ctx context.Context, workflowID, runID string, maxEvents int) ([]EventSummary, error)
}

// ListQuery is the input to WorkflowAPI.ListWorkflows.
type ListQuery struct {
	Namespace string
	TaskQueue string    // empty = no filter
	Since     time.Time // workflows started at-or-after this point
	PageSize  int       // 0 = adapter default
	PageToken []byte    // empty = first page
}

// ListResult is one page of results.
type ListResult struct {
	Executions    []WorkflowExec
	NextPageToken []byte
}

// WorkflowExec is the narrow projection of a workflow listing entry.
type WorkflowExec struct {
	WorkflowID       string
	RunID            string
	TaskQueue        string
	StartTime        time.Time
	ExecutionTimeout time.Duration // 0 means "unset" (server default applies)
	RunTimeout       time.Duration
	Status           string
}

// WorkflowDescription is the narrow projection of DescribeWorkflowExecution.
type WorkflowDescription struct {
	WorkflowID       string
	RunID            string
	HistoryLength    int64
	HistorySizeBytes int64
	ExecutionTimeout time.Duration
	RunTimeout       time.Duration
}

// EventSummary is one history event as we care about it.
type EventSummary struct {
	EventID     int64
	EventType   string
	PayloadSize int64
}

// VersionMarker is the projection of a workflow.GetVersion marker
// recorded in workflow history. The cleanup-versions subcommand uses
// these to verify whether a GetVersion call is safe to remove.
type VersionMarker struct {
	ChangeID string
	Version  int
}

// VersionAPI extends WorkflowAPI with the operations needed by the
// cleanup-versions subcommand. Implementations may also satisfy
// WorkflowAPI; tests can satisfy only the methods they need.
type VersionAPI interface {
	OpenWorkflows(ctx context.Context, q ListQuery) (ListResult, error)
	VersionMarkers(ctx context.Context, workflowID, runID string) ([]VersionMarker, error)
}

// Deps is what every Check.Run receives.
type Deps struct {
	API        WorkflowAPI
	Namespace  string
	TaskQueue  string
	Since      time.Time
	Thresholds Thresholds
}

// Thresholds are loaded from --threshold-config (YAML) with package
// defaults. Each check reads only its own field.
type Thresholds struct {
	HistoryBytesWarn       int64         `yaml:"history_bytes_warn"`
	HistoryBytesFail       int64         `yaml:"history_bytes_fail"`
	HistoryEventsWarn      int64         `yaml:"history_events_warn"`
	HistoryEventsFail      int64         `yaml:"history_events_fail"`
	IndividualPayloadWarn  int64         `yaml:"individual_payload_warn"`
	IndividualPayloadFail  int64         `yaml:"individual_payload_fail"`
	NoWorkflowTimeoutAfter time.Duration `yaml:"no_workflow_timeout_after"`
	SampleHistoryEvents    int           `yaml:"sample_history_events"`
}

// Check is the runtime analogue of *analysis.Analyzer. Each check is
// registered in runtime/all/all.go and runs against the same Deps shared
// across the suite.
type Check interface {
	Name() string
	Category() all.Category
	URL() string
	Doc() string
	Run(ctx context.Context, deps Deps) ([]Finding, error)
}
