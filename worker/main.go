package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"temporal-learning/activity"
	"temporal-learning/shared"
	"temporal-learning/workflow"
)

func main() {
	// Connect to Temporal server (default: localhost:7233)
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("Unable to connect to Temporal server: %v", err)
	}
	defer c.Close()

	// Create a worker that listens on the task queue
	w := worker.New(c, shared.TaskQueueName, worker.Options{})

	// Register workflow
	w.RegisterWorkflow(workflow.OrderWorkflow)

	// Register activities
	w.RegisterActivity(activity.ValidateOrder)
	w.RegisterActivity(activity.ChargePayment)
	w.RegisterActivity(activity.ShipOrder)
	w.RegisterActivity(activity.SendNotification)

	log.Println("Starting worker... listening on task queue:", shared.TaskQueueName)

	// Start the worker (blocks until interrupted)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
