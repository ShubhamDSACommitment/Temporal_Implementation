package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"order-service/internal/activities"
	"order-service/internal/config"
	"order-service/internal/workflow"

	"temporal-shared"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	// Connect to Temporal with retry backoff
	var c client.Client
	var err error
	for i := range 30 {
		c, err = client.Dial(client.Options{
			HostPort:  cfg.TemporalAddress,
			Namespace: cfg.TemporalNamespace,
		})
		if err == nil {
			break
		}
		wait := time.Duration(min(i+1, 5)) * time.Second
		slog.Warn("Failed to connect to Temporal, retrying...", "attempt", i+1, "error", err, "retry_in", wait)
		time.Sleep(wait)
	}
	if err != nil {
		slog.Error("Unable to connect to Temporal server", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, shared.OrderTaskQueue, worker.Options{})

	w.RegisterWorkflow(workflow.OrderWorkflow)
	w.RegisterActivity(activities.ValidateOrder)
	w.RegisterActivity(activities.ShipOrder)
	w.RegisterActivity(activities.SendNotification)

	// Register activities with the gateway
	regReq := shared.RegisterActivitiesRequest{
		ServiceName: "order-service",
		Activities: []shared.ActivityRegistration{
			{
				ActivityName: "ValidateOrder",
				DisplayName:  "Validate Order",
				Description:  "Validates order details and inventory",
				TaskQueue:    shared.OrderTaskQueue,
				Category:     "Order",
				InputSchema: []shared.InputFieldSchema{
					{Name: "order_id", Type: "string", Required: true, Description: "Order identifier"},
					{Name: "customer_id", Type: "string", Required: true, Description: "Customer identifier"},
					{Name: "item", Type: "string", Required: true, Description: "Item name"},
					{Name: "quantity", Type: "number", Required: true, Description: "Quantity ordered"},
					{Name: "price", Type: "number", Required: true, Description: "Item price"},
				},
				DefaultMapping: map[string]string{
					"order_id":    "$input.order_id",
					"customer_id": "$input.customer_id",
					"item":        "$input.item",
					"quantity":    "$input.quantity",
					"price":       "$input.price",
				},
			},
			{
				ActivityName: "ShipOrder",
				DisplayName:  "Ship Order",
				Description:  "Ships the order and returns tracking info",
				TaskQueue:    shared.OrderTaskQueue,
				Category:     "Order",
				InputSchema: []shared.InputFieldSchema{
					{Name: "order_id", Type: "string", Required: true, Description: "Order identifier"},
					{Name: "customer_id", Type: "string", Required: true, Description: "Customer identifier"},
					{Name: "item", Type: "string", Required: true, Description: "Item name"},
					{Name: "quantity", Type: "number", Required: true, Description: "Quantity ordered"},
					{Name: "price", Type: "number", Required: true, Description: "Item price"},
				},
				DefaultMapping: map[string]string{
					"order_id":    "$input.order_id",
					"customer_id": "$input.customer_id",
					"item":        "$input.item",
					"quantity":    "$input.quantity",
					"price":       "$input.price",
				},
			},
			{
				ActivityName: "SendNotification",
				DisplayName:  "Send Notification",
				Description:  "Sends a notification to the customer",
				TaskQueue:    shared.OrderTaskQueue,
				Category:     "Order",
				InputSchema: []shared.InputFieldSchema{
					{Name: "order_id", Type: "string", Required: true, Description: "Order identifier"},
					{Name: "message", Type: "string", Required: true, Description: "Notification message"},
				},
				DefaultMapping: map[string]string{
					"order_id": "$input.order_id",
					"message":  "Order update",
				},
			},
		},
	}

	go func() {
		if err := shared.RegisterWithGateway(cfg.GatewayURL, regReq); err != nil {
			slog.Error("Initial registration failed", "error", err)
		}
	}()

	heartbeatStop := make(chan struct{})
	go shared.StartHeartbeat(cfg.GatewayURL, regReq, 2*time.Minute, heartbeatStop)

	// Health check endpoint
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	healthServer := &http.Server{Addr: ":" + cfg.HealthPort, Handler: healthMux}

	go func() {
		slog.Info("Health check server starting", "port", cfg.HealthPort)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Health check server failed", "error", err)
		}
	}()

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down...")
		close(heartbeatStop)
		healthServer.Shutdown(context.Background())
		w.Stop()
	}()

	slog.Info("Order service worker starting", "taskQueue", shared.OrderTaskQueue)
	if err := w.Run(worker.InterruptCh()); err != nil {
		slog.Error("Worker failed", "error", err)
		os.Exit(1)
	}
}
