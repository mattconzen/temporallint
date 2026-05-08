package main

import (
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var _ = worker.New(nil, "orders-tq", worker.Options{})

var _ = client.StartWorkflowOptions{
	ID:                       "wf-1",
	TaskQueue:                "shipments-tq", // want `does not match any worker.New`
	WorkflowExecutionTimeout: time.Hour,
}
