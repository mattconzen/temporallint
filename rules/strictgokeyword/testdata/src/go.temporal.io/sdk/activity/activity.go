// Stub implementation of go.temporal.io/sdk/activity for analysistest
// fixtures.
package activity

import "context"

func RecordHeartbeat(ctx context.Context, details ...interface{}) {}

func GetInfo(ctx context.Context) Info { return Info{} }

type Info struct {
	WorkflowType string
	ActivityID   string
	Attempt      int32
}
