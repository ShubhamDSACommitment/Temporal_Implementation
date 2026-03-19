package shared

// TaskQueueName is the task queue all workers listen on.
const TaskQueueName = "order-task-queue"

// OrderInput is the payload passed into the workflow.
type OrderInput struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Item       string  `json:"item"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

// OrderResult is the final output of the workflow.
type OrderResult struct {
	OrderID       string  `json:"order_id"`
	Status        string  `json:"status"`
	TotalCharged  float64 `json:"total_charged"`
	TrackingNumber string `json:"tracking_number"`
}

// PaymentResult is returned by the charge activity.
type PaymentResult struct {
	TransactionID string  `json:"transaction_id"`
	AmountCharged float64 `json:"amount_charged"`
}

// ShipmentResult is returned by the ship activity.
type ShipmentResult struct {
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier"`
}
