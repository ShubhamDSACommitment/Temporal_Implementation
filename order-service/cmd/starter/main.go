package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.temporal.io/sdk/client"

	"order-service/internal/config"
	"order-service/internal/workflow"

	"temporal-shared"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	c, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalAddress,
		Namespace: cfg.TemporalNamespace,
	})
	if err != nil {
		slog.Error("Unable to connect to Temporal server", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	order := shared.OrderInput{
		OrderID:    "ORD-001",
		CustomerID: "CUST-123",
		Item:       "Mechanical Keyboard",
		Quantity:   2,
		Price:      79.99,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        "order-workflow-" + order.OrderID,
		TaskQueue: shared.OrderTaskQueue,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflow.OrderWorkflow, order)
	if err != nil {
		slog.Error("Unable to start workflow", "error", err)
		os.Exit(1)
	}

	slog.Info("Workflow started", "workflowID", we.GetID(), "runID", we.GetRunID())

	var result shared.OrderResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		slog.Error("Workflow failed", "error", err)
		os.Exit(1)
	}

	fmt.Println("\n========== ORDER RESULT ==========")
	fmt.Printf("Order ID:        %s\n", result.OrderID)
	fmt.Printf("Status:          %s\n", result.Status)
	fmt.Printf("Total Charged:   $%.2f\n", result.TotalCharged)
	fmt.Printf("Tracking Number: %s\n", result.TrackingNumber)
	fmt.Println("==================================")
}
