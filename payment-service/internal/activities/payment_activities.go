package activities

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// ChargePayment processes a payment for an order.
// Registered with explicit name "ChargePayment" so the order-service
// can invoke it by string without importing this package.
func ChargePayment(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	orderID, _ := input["order_id"].(string)
	amount, _ := input["amount"].(float64)
	logger.Info("Charging payment", "orderID", orderID, "amount", amount)

	if amount <= 0 {
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("invalid payment amount: %.2f", amount),
			"PAYMENT_VALIDATION_ERROR", nil,
		)
	}

	// Simulate payment gateway call
	time.Sleep(1 * time.Second)

	txnID := fmt.Sprintf("TXN-%d", rand.Intn(100000))

	logger.Info("Payment charged successfully", "transactionID", txnID, "amount", amount)
	return map[string]interface{}{
		"transaction_id": txnID,
		"amount_charged": amount,
	}, nil
}

// RefundPayment refunds a previously charged payment.
// Registered with explicit name "RefundPayment".
func RefundPayment(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	orderID, _ := input["order_id"].(string)
	amount, _ := input["amount"].(float64)
	logger.Info("Refunding payment", "orderID", orderID, "amount", amount)

	time.Sleep(800 * time.Millisecond)

	txnID := fmt.Sprintf("REFUND-%d", rand.Intn(100000))

	logger.Info("Payment refunded successfully", "transactionID", txnID, "amount", amount)
	return map[string]interface{}{
		"transaction_id": txnID,
		"amount_charged": -amount,
	}, nil
}
