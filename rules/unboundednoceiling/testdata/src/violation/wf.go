package violation

import (
	"time"

	"go.temporal.io/sdk/temporal"
)

var _ = temporal.RetryPolicy{ // want `unbounded \(no MaximumAttempts\) and has no MaximumInterval`
	InitialInterval:    time.Second,
	BackoffCoefficient: 2.0,
}
