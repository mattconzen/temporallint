package clean

import "context"

type Worker struct{}

func (w *Worker) RegisterActivity(fn interface{}) {}

func GoodActivity(ctx context.Context, input string) error {
	return nil
}

func register() {
	w := &Worker{}
	w.RegisterActivity(GoodActivity)
}

// Helper functions that aren't registered are not flagged even without context.
func helper(s string) string { return s }
