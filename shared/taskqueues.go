package shared

const (
	// OrderTaskQueue is the task queue for order-service activities and workflow.
	OrderTaskQueue = "order-task-queue"

	// PaymentTaskQueue is the task queue for payment-service activities.
	PaymentTaskQueue = "payment-task-queue"

	// PlatformTaskQueue is the task queue for the dynamic workflow engine.
	PlatformTaskQueue = "platform-task-queue"
)
