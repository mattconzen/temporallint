package clean

import (
	"time"

	"go.temporal.io/sdk/temporal"
)

var _ = temporal.RetryPolicy{
	InitialInterval: time.Second,
	MaximumAttempts: 10, // bounded
}

var _ = temporal.RetryPolicy{
	InitialInterval: time.Second,
	MaximumInterval: time.Minute, // unbounded but capped
}
