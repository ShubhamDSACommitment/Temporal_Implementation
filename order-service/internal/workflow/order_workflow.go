package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"order-service/internal/activities"

	"temporal-shared"
)

// OrderWorkflow orchestrates the full order processing pipeline.
//
// It uses two task queues:
//   - OrderTaskQueue for local activities (validate, ship, notify)
//   - PaymentTaskQueue for cross-service payment activities
//
// The payment activity is invoked by string name because the payment-service
// code is not imported — Temporal carries the payload between services.
func OrderWorkflow(ctx workflow.Context, order shared.OrderInput) (shared.OrderResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderWorkflow started", "orderID", order.OrderID)

	// Context for local order-service activities
	orderCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           shared.OrderTaskQueue,
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	})

	// Context for cross-service payment activities
	paymentCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           shared.PaymentTaskQueue,
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	})

	// Convert order to map for activities
	orderMap := map[string]interface{}{
		"order_id":    order.OrderID,
		"customer_id": order.CustomerID,
		"item":        order.Item,
		"quantity":    float64(order.Quantity),
		"price":       order.Price,
	}

	// Step 1: Validate Order (local)
	var validationResult map[string]interface{}
	err := workflow.ExecuteActivity(orderCtx, activities.ValidateOrder, orderMap).Get(ctx, &validationResult)
	if err != nil {
		logger.Error("Order validation failed", "orderID", order.OrderID, "error", err)
		return shared.OrderResult{}, fmt.Errorf("validation failed: %w", err)
	}
	logger.Info("Order validated", "orderID", order.OrderID, "status", validationResult["status"])

	// Step 2: Charge Payment (cross-service via string name)
	paymentMap := map[string]interface{}{
		"order_id":    order.OrderID,
		"customer_id": order.CustomerID,
		"amount":      order.Price * float64(order.Quantity),
	}

	var payment map[string]interface{}
	err = workflow.ExecuteActivity(paymentCtx, "ChargePayment", paymentMap).Get(ctx, &payment)
	if err != nil {
		logger.Error("Payment failed", "orderID", order.OrderID, "error", err)
		return shared.OrderResult{}, fmt.Errorf("payment failed: %w", err)
	}
	txnID, _ := payment["transaction_id"].(string)
	amountCharged, _ := payment["amount_charged"].(float64)
	logger.Info("Payment charged", "orderID", order.OrderID, "transactionID", txnID)

	// Step 3: Ship Order (local)
	var shipment map[string]interface{}
	err = workflow.ExecuteActivity(orderCtx, activities.ShipOrder, orderMap).Get(ctx, &shipment)
	if err != nil {
		logger.Error("Shipment failed", "orderID", order.OrderID, "error", err)
		return shared.OrderResult{}, fmt.Errorf("shipment failed: %w", err)
	}
	trackingNumber, _ := shipment["tracking_number"].(string)
	logger.Info("Order shipped", "orderID", order.OrderID, "trackingNumber", trackingNumber)

	// Step 4: Send Notification (local, non-critical)
	notifInput := map[string]interface{}{
		"order_id": order.OrderID,
		"message":  fmt.Sprintf("Your order %s has been shipped! Tracking: %s", order.OrderID, trackingNumber),
	}
	var notifResult map[string]interface{}
	err = workflow.ExecuteActivity(orderCtx, activities.SendNotification, notifInput).Get(ctx, &notifResult)
	if err != nil {
		logger.Warn("Notification failed, but order is complete", "orderID", order.OrderID, "error", err)
	}

	result := shared.OrderResult{
		OrderID:        order.OrderID,
		Status:         "COMPLETED",
		TotalCharged:   amountCharged,
		TrackingNumber: trackingNumber,
	}

	logger.Info("OrderWorkflow completed", "orderID", order.OrderID, "result", result)
	return result, nil
}
