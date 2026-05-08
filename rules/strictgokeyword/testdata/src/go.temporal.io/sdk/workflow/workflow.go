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
}

type LocalActivityOptions struct {
	StartToCloseTimeout    time.Duration
	ScheduleToCloseTimeout time.Duration
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
func ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future {
	return nil
}
func Now(ctx Context) time.Time          { return time.Time{} }
func Sleep(ctx Context, d time.Duration) error { return nil }
func NewChannel(ctx Context) Channel     { return nil }
func Go(ctx Context, f func(Context))    {}
func SideEffect(ctx Context, f func(Context) interface{}) interface{} {
	return nil
}
func NewTimer(ctx Context, d time.Duration) Future { return nil }
