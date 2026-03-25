package activities

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// ValidateOrder checks if the order is valid.
func ValidateOrder(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	orderID, _ := input["order_id"].(string)
	logger.Info("Validating order", "orderID", orderID)

	quantity, _ := input["quantity"].(float64)
	price, _ := input["price"].(float64)

	if quantity <= 0 {
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("invalid quantity: %v", quantity),
			"VALIDATION_ERROR", nil,
		)
	}
	if price <= 0 {
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("invalid price: %.2f", price),
			"VALIDATION_ERROR", nil,
		)
	}

	time.Sleep(500 * time.Millisecond)
	logger.Info("Order validated successfully", "orderID", orderID)
	return map[string]interface{}{
		"status":   "VALIDATED",
		"order_id": orderID,
	}, nil
}

// ShipOrder creates a shipment for the order.
func ShipOrder(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	orderID, _ := input["order_id"].(string)
	logger.Info("Shipping order", "orderID", orderID)

	time.Sleep(1 * time.Second)

	trackingNum := fmt.Sprintf("TRACK-%d", rand.Intn(100000))

	logger.Info("Order shipped", "trackingNumber", trackingNum)
	return map[string]interface{}{
		"tracking_number": trackingNum,
		"carrier":         "FedEx",
	}, nil
}

// SendNotification sends a notification to the customer.
func SendNotification(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	orderID, _ := input["order_id"].(string)
	message, _ := input["message"].(string)
	if message == "" {
		message = "Order update"
	}
	logger.Info("Sending notification", "orderID", orderID, "message", message)

	time.Sleep(300 * time.Millisecond)

	logger.Info("Notification sent", "orderID", orderID)
	return map[string]interface{}{
		"status":   "SENT",
		"order_id": orderID,
	}, nil
}
