package violation

type Worker struct{}

func (w *Worker) RegisterActivity(fn interface{}) {}

func BadActivity(input string) error { // want `must accept context.Context as its first parameter`
	return nil
}

func register() {
	w := &Worker{}
	w.RegisterActivity(BadActivity)
}
