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

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"payment-service/internal/activities"
	"payment-service/internal/config"

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

	// Create worker on PaymentTaskQueue
	w := worker.New(c, shared.PaymentTaskQueue, worker.Options{})

	// Register activities with explicit names for cross-service invocation
	w.RegisterActivityWithOptions(activities.ChargePayment, activity.RegisterOptions{
		Name: "ChargePayment",
	})
	w.RegisterActivityWithOptions(activities.RefundPayment, activity.RegisterOptions{
		Name: "RefundPayment",
	})

	// Register activities with the gateway
	regReq := shared.RegisterActivitiesRequest{
		ServiceName: "payment-service",
		Activities: []shared.ActivityRegistration{
			{
				ActivityName: "ChargePayment",
				DisplayName:  "Charge Payment",
				Description:  "Charges the customer for the order",
				TaskQueue:    shared.PaymentTaskQueue,
				Category:     "Payment",
				InputSchema: []shared.InputFieldSchema{
					{Name: "order_id", Type: "string", Required: true, Description: "Order identifier"},
					{Name: "customer_id", Type: "string", Required: true, Description: "Customer identifier"},
					{Name: "amount", Type: "number", Required: true, Description: "Amount to charge"},
				},
				DefaultMapping: map[string]string{
					"order_id":    "$input.order_id",
					"customer_id": "$input.customer_id",
					"amount":      "$input.price",
				},
			},
			{
				ActivityName: "RefundPayment",
				DisplayName:  "Refund Payment",
				Description:  "Refunds the customer for the order",
				TaskQueue:    shared.PaymentTaskQueue,
				Category:     "Payment",
				InputSchema: []shared.InputFieldSchema{
					{Name: "order_id", Type: "string", Required: true, Description: "Order identifier"},
					{Name: "customer_id", Type: "string", Required: true, Description: "Customer identifier"},
					{Name: "amount", Type: "number", Required: true, Description: "Amount to refund"},
				},
				DefaultMapping: map[string]string{
					"order_id":    "$input.order_id",
					"customer_id": "$input.customer_id",
					"amount":      "$input.price",
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

	slog.Info("Payment service worker starting", "taskQueue", shared.PaymentTaskQueue)
	if err := w.Run(worker.InterruptCh()); err != nil {
		slog.Error("Worker failed", "error", err)
		os.Exit(1)
	}
}
