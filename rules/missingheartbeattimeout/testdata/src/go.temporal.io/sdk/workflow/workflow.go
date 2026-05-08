// Stub implementation of go.temporal.io/sdk/workflow used by analysistest
// fixtures across every rule's testdata. The deploy script
// (scripts/sync-testdata-stubs.sh, or the Bash loop below) copies this
// file into each rule's testdata/src/go.temporal.io/sdk/workflow/. Edit
// here, redeploy.
package workflow

import "time"

type Context interface {
	Done() <-chan struct{}
}

type ActivityOptions struct {
	StartToCloseTimeout    time.Duration
	ScheduleToCloseTimeout time.Duration
	HeartbeatTimeout       time.Duration
	TaskQueue              string
	RetryPolicy            *RetryPolicyAlias
}

// RetryPolicyAlias mirrors go.temporal.io/sdk/temporal.RetryPolicy. The
// real SDK uses temporal.RetryPolicy; the alias here lets ActivityOptions
// reference it without importing across stub packages.
type RetryPolicyAlias = struct {
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaximumInterval    time.Duration
	MaximumAttempts    int32
}

type LocalActivityOptions struct {
	StartToCloseTimeout    time.Duration
	ScheduleToCloseTimeout time.Duration
}

type ChildWorkflowOptions struct {
	WorkflowID                string
	WorkflowExecutionTimeout  time.Duration
	WorkflowRunTimeout        time.Duration
	WorkflowTaskTimeout       time.Duration
	TaskQueue                 string
	ParentClosePolicy         int
}

type Channel interface {
	Send(ctx Context, v interface{})
	Receive(ctx Context, valuePtr interface{}) (more bool)
}

type Future interface {
	Get(ctx Context, valuePtr interface{}) error
	IsReady() bool
}

func WithActivityOptions(ctx Context, options ActivityOptions) Context { return ctx }
func WithChildWorkflowOptions(ctx Context, options ChildWorkflowOptions) Context {
	return ctx
}
func ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future {
	return nil
}
func ExecuteChildWorkflow(ctx Context, child interface{}, args ...interface{}) Future {
	return nil
}
func Now(ctx Context) time.Time                { return time.Time{} }
func Sleep(ctx Context, d time.Duration) error { return nil }
func NewChannel(ctx Context) Channel            { return nil }
func Go(ctx Context, f func(Context))           {}
func SideEffect(ctx Context, f func(Context) interface{}) interface{} {
	return nil
}
func NewTimer(ctx Context, d time.Duration) Future                { return nil }
func GetVersion(ctx Context, changeID string, min, max int) int    { return 0 }
func GetSignalChannel(ctx Context, name string) Channel            { return nil }
func ContinueAsNew(ctx Context, fn interface{}, args ...interface{}) error {
	return nil
}

const DefaultVersion = -1
