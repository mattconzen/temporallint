// Stub implementation of go.temporal.io/sdk/temporal for analysistest
// fixtures.
package temporal

import "time"

type RetryPolicy struct {
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaximumInterval    time.Duration
	MaximumAttempts    int32
	NonRetryableErrorTypes []string
}

type ApplicationError struct {
	Message string
	Type    string
}

func NewNonRetryableApplicationError(message, errType string, cause error, details ...interface{}) error {
	return &ApplicationError{Message: message, Type: errType}
}

func (e *ApplicationError) Error() string { return e.Message }
