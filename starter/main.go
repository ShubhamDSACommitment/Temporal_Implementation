package main

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"

	"temporal-learning/shared"
	"temporal-learning/workflow"
)

func main() {
	// Connect to Temporal server
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("Unable to connect to Temporal server: %v", err)
	}
	defer c.Close()

	// Create an order
	order := shared.OrderInput{
		OrderID:    "ORD-001",
		CustomerID: "CUST-123",
		Item:       "Mechanical Keyboard",
		Quantity:   2,
		Price:      79.99,
	}

	// Start the workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        "order-workflow-" + order.OrderID,
		TaskQueue: shared.TaskQueueName,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflow.OrderWorkflow, order)
	if err != nil {
		log.Fatalf("Unable to start workflow: %v", err)
	}

	log.Printf("Workflow started! WorkflowID: %s, RunID: %s\n", we.GetID(), we.GetRunID())

	// Wait for the workflow to complete and get the result
	var result shared.OrderResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	fmt.Println("\n========== ORDER RESULT ==========")
	fmt.Printf("Order ID:        %s\n", result.OrderID)
	fmt.Printf("Status:          %s\n", result.Status)
	fmt.Printf("Total Charged:   $%.2f\n", result.TotalCharged)
	fmt.Printf("Tracking Number: %s\n", result.TrackingNumber)
	fmt.Println("==================================")
}
