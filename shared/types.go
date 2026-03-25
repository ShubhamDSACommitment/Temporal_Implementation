package shared

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
	OrderID        string  `json:"order_id"`
	Status         string  `json:"status"`
	TotalCharged   float64 `json:"total_charged"`
	TrackingNumber string  `json:"tracking_number"`
}

// PaymentRequest is the payload sent cross-service to the payment worker.
type PaymentRequest struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
}

// PaymentResult is returned by the charge/refund activities.
type PaymentResult struct {
	TransactionID string  `json:"transaction_id"`
	AmountCharged float64 `json:"amount_charged"`
}

// ShipmentResult is returned by the ship activity.
type ShipmentResult struct {
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier"`
}

// InputFieldSchema describes one expected input field for an activity.
type InputFieldSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "number", "boolean"
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// ActivityRegistration is what a service sends to register its activities.
type ActivityRegistration struct {
	ActivityName   string            `json:"activity_name"`
	DisplayName    string            `json:"display_name"`
	Description    string            `json:"description"`
	TaskQueue      string            `json:"task_queue"`
	ServiceName    string            `json:"service_name"`
	Category       string            `json:"category"`
	InputSchema    []InputFieldSchema `json:"input_schema"`
	DefaultMapping map[string]string `json:"default_mapping"`
}

// RegisterActivitiesRequest is the payload a service POSTs to register its activities.
type RegisterActivitiesRequest struct {
	ServiceName string                 `json:"service_name"`
	Activities  []ActivityRegistration `json:"activities"`
}
