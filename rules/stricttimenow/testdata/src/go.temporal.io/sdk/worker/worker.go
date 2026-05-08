// Stub implementation of go.temporal.io/sdk/worker for analysistest
// fixtures.
package worker

type Options struct {
	MaxConcurrentActivityExecutionSize int
	WorkerStopTimeout                  int
}

type Worker interface {
	RegisterWorkflow(fn interface{})
	RegisterActivity(fn interface{})
	Run(stopCh <-chan interface{}) error
	Start() error
	Stop()
}

func New(c interface{}, taskQueue string, options Options) Worker { return nil }
func InterruptCh() <-chan interface{}                             { return nil }
