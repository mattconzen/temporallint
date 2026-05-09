// Package fakeapi is a test double that satisfies runtime.WorkflowAPI by
// returning canned responses. It exists so per-check tests can express
// "given this server state, the check should emit this finding" with no
// proto types or mock-generation infrastructure. Production code never
// imports this package.
package fakeapi

import (
	"context"

	"github.com/mattconzen/temporallint/runtime"
)

// API is a configurable, single-namespace fake implementation of
// runtime.WorkflowAPI plus runtime.VersionAPI. Builder methods (`With*`)
// return the receiver so they chain.
type API struct {
	pages         [][]runtime.WorkflowExec
	openPages     [][]runtime.WorkflowExec
	descriptions  map[string]runtime.WorkflowDescription
	histories     map[string][]runtime.EventSummary
	versionMarks  map[string][]runtime.VersionMarker
	listError     error
	describeError error
	sampleError   error
	openError     error
	markersError  error
}

func New() *API {
	return &API{
		descriptions: map[string]runtime.WorkflowDescription{},
		histories:    map[string][]runtime.EventSummary{},
		versionMarks: map[string][]runtime.VersionMarker{},
	}
}

func (a *API) WithOpenExec(e runtime.WorkflowExec) *API {
	if len(a.openPages) == 0 {
		a.openPages = [][]runtime.WorkflowExec{{}}
	}
	a.openPages[0] = append(a.openPages[0], e)
	return a
}

func (a *API) WithVersionMarkers(workflowID, runID string, markers []runtime.VersionMarker) *API {
	a.versionMarks[workflowID+"/"+runID] = markers
	return a
}

func (a *API) WithOpenError(err error) *API    { a.openError = err; return a }
func (a *API) WithMarkersError(err error) *API { a.markersError = err; return a }

// WithExec adds a single-page exec to the fake's ListWorkflows response.
// Subsequent calls accumulate onto the same page; use WithPagedExecs to
// emit multi-page responses.
func (a *API) WithExec(e runtime.WorkflowExec) *API {
	if len(a.pages) == 0 {
		a.pages = [][]runtime.WorkflowExec{{}}
	}
	a.pages[0] = append(a.pages[0], e)
	return a
}

// WithPagedExecs replaces the entire page sequence with the given slices.
func (a *API) WithPagedExecs(pages ...[]runtime.WorkflowExec) *API {
	a.pages = append([][]runtime.WorkflowExec(nil), pages...)
	return a
}

// WithDescription registers a DescribeWorkflowExecution response keyed
// by workflowID/runID.
func (a *API) WithDescription(d runtime.WorkflowDescription) *API {
	a.descriptions[d.WorkflowID+"/"+d.RunID] = d
	return a
}

// WithHistory registers a SampleHistory response keyed by workflowID/runID.
func (a *API) WithHistory(workflowID, runID string, events []runtime.EventSummary) *API {
	a.histories[workflowID+"/"+runID] = events
	return a
}

// WithListError makes ListWorkflows return err on the next call.
func (a *API) WithListError(err error) *API     { a.listError = err; return a }
func (a *API) WithDescribeError(err error) *API { a.describeError = err; return a }
func (a *API) WithSampleError(err error) *API   { a.sampleError = err; return a }

func (a *API) ListWorkflows(_ context.Context, q runtime.ListQuery) (runtime.ListResult, error) {
	if a.listError != nil {
		return runtime.ListResult{}, a.listError
	}
	idx := 0
	if len(q.PageToken) > 0 {
		idx = int(q.PageToken[0])
	}
	if idx >= len(a.pages) {
		return runtime.ListResult{}, nil
	}
	res := runtime.ListResult{Executions: a.pages[idx]}
	if idx+1 < len(a.pages) {
		res.NextPageToken = []byte{byte(idx + 1)}
	}
	return res, nil
}

func (a *API) DescribeWorkflow(_ context.Context, workflowID, runID string) (runtime.WorkflowDescription, error) {
	if a.describeError != nil {
		return runtime.WorkflowDescription{}, a.describeError
	}
	return a.descriptions[workflowID+"/"+runID], nil
}

func (a *API) SampleHistory(_ context.Context, workflowID, runID string, _ int) ([]runtime.EventSummary, error) {
	if a.sampleError != nil {
		return nil, a.sampleError
	}
	return a.histories[workflowID+"/"+runID], nil
}

// OpenWorkflows mirrors ListWorkflows but draws from a separate page
// list so tests can configure them independently. If no Open pages are
// configured, it falls back to ListWorkflows' pages.
func (a *API) OpenWorkflows(ctx context.Context, q runtime.ListQuery) (runtime.ListResult, error) {
	if a.openError != nil {
		return runtime.ListResult{}, a.openError
	}
	if len(a.openPages) == 0 {
		return a.ListWorkflows(ctx, q)
	}
	idx := 0
	if len(q.PageToken) > 0 {
		idx = int(q.PageToken[0])
	}
	if idx >= len(a.openPages) {
		return runtime.ListResult{}, nil
	}
	res := runtime.ListResult{Executions: a.openPages[idx]}
	if idx+1 < len(a.openPages) {
		res.NextPageToken = []byte{byte(idx + 1)}
	}
	return res, nil
}

// VersionMarkers returns the markers registered for the given
// workflow/run. Empty slice when none are registered.
func (a *API) VersionMarkers(_ context.Context, workflowID, runID string) ([]runtime.VersionMarker, error) {
	if a.markersError != nil {
		return nil, a.markersError
	}
	return a.versionMarks[workflowID+"/"+runID], nil
}
