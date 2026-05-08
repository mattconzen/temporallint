package main

import (
	"go.temporal.io/sdk/worker"
)

func main() {
	w := worker.New(nil, "tq", worker.Options{})
	_ = w.Run(nil) // want `no interrupt channel`
}
