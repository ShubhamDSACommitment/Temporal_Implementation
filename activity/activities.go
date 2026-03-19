package activity

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"

	"temporal-learning/shared"
)

// ValidateOrder checks if the order is valid (item exists, quantity > 0, etc).
func ValidateOrder(ctx context.Context, order shared.OrderInput) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Validating order", "orderID", order.OrderID)

	// Simulate validation logic
	if order.Quantity <= 0 {
		return "", fmt.Errorf("invalid quantity: %d", order.Quantity)
	}
	if order.Price <= 0 {
		return "", fmt.Errorf("invalid price: %.2f", order.Price)
	}

	time.Sleep(500 * time.Millisecond) // simulate some work
	logger.Info("Order validated successfully", "orderID", order.OrderID)
	return "VALIDATED", nil
}

// ChargePayment processes the payment for the order.
func ChargePayment(ctx context.Context, order shared.OrderInput) (shared.PaymentResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Charging payment", "orderID", order.OrderID, "amount", order.Price*float64(order.Quantity))

	// Simulate payment processing
	time.Sleep(1 * time.Second)

	totalAmount := order.Price * float64(order.Quantity)
	txnID := fmt.Sprintf("TXN-%d", rand.Intn(100000))

	logger.Info("Payment charged", "transactionID", txnID, "amount", totalAmount)
	return shared.PaymentResult{
		TransactionID: txnID,
		AmountCharged: totalAmount,
	}, nil
}

// ShipOrder creates a shipment for the order.
func ShipOrder(ctx context.Context, order shared.OrderInput) (shared.ShipmentResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Shipping order", "orderID", order.OrderID)

	// Simulate shipment creation
	time.Sleep(1 * time.Second)

	trackingNum := fmt.Sprintf("TRACK-%d", rand.Intn(100000))

	logger.Info("Order shipped", "trackingNumber", trackingNum)
	return shared.ShipmentResult{
		TrackingNumber: trackingNum,
		Carrier:        "FedEx",
	}, nil
}

// SendNotification sends a notification to the customer.
func SendNotification(ctx context.Context, orderID string, message string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending notification", "orderID", orderID, "message", message)

	// Simulate sending email/SMS
	time.Sleep(300 * time.Millisecond)

	logger.Info("Notification sent", "orderID", orderID)
	return nil
}
