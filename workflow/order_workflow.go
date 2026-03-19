package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"temporal-learning/activity"
	"temporal-learning/shared"
)

// OrderWorkflow orchestrates the full order processing pipeline.
//
// Steps:
//  1. Validate the order
//  2. Charge payment
//  3. Ship the order
//  4. Send notification to customer
//
// If any step fails, the workflow fails with an error.
func OrderWorkflow(ctx workflow.Context, order shared.OrderInput) (shared.OrderResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderWorkflow started", "orderID", order.OrderID)

	// Activity options: retry up to 3 times, timeout after 10 seconds
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Validate Order
	var validationStatus string
	err := workflow.ExecuteActivity(ctx, activity.ValidateOrder, order).Get(ctx, &validationStatus)
	if err != nil {
		logger.Error("Order validation failed", "orderID", order.OrderID, "error", err)
		return shared.OrderResult{}, fmt.Errorf("validation failed: %w", err)
	}
	logger.Info("Order validated", "orderID", order.OrderID, "status", validationStatus)

	// Step 2: Charge Payment
	var payment shared.PaymentResult
	err = workflow.ExecuteActivity(ctx, activity.ChargePayment, order).Get(ctx, &payment)
	if err != nil {
		logger.Error("Payment failed", "orderID", order.OrderID, "error", err)
		return shared.OrderResult{}, fmt.Errorf("payment failed: %w", err)
	}
	logger.Info("Payment charged", "orderID", order.OrderID, "transactionID", payment.TransactionID)

	// Step 3: Ship Order
	var shipment shared.ShipmentResult
	err = workflow.ExecuteActivity(ctx, activity.ShipOrder, order).Get(ctx, &shipment)
	if err != nil {
		logger.Error("Shipment failed", "orderID", order.OrderID, "error", err)
		return shared.OrderResult{}, fmt.Errorf("shipment failed: %w", err)
	}
	logger.Info("Order shipped", "orderID", order.OrderID, "trackingNumber", shipment.TrackingNumber)

	// Step 4: Send Notification
	message := fmt.Sprintf("Your order %s has been shipped! Tracking: %s", order.OrderID, shipment.TrackingNumber)
	err = workflow.ExecuteActivity(ctx, activity.SendNotification, order.OrderID, message).Get(ctx, nil)
	if err != nil {
		// Notification failure is non-critical, log and continue
		logger.Warn("Notification failed, but order is complete", "orderID", order.OrderID, "error", err)
	}

	result := shared.OrderResult{
		OrderID:        order.OrderID,
		Status:         "COMPLETED",
		TotalCharged:   payment.AmountCharged,
		TrackingNumber: shipment.TrackingNumber,
	}

	logger.Info("OrderWorkflow completed", "orderID", order.OrderID, "result", result)
	return result, nil
}
