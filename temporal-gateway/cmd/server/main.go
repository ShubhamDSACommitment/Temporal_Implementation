package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"temporal-gateway/internal/api"
	"temporal-gateway/internal/config"
	"temporal-gateway/internal/store"
	dynamicwf "temporal-gateway/internal/workflow"
	"temporal-shared"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	// Connect to Temporal
	tc, err := client.Dial(client.Options{
		HostPort: cfg.TemporalAddress,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Temporal: %v", err)
	}
	defer tc.Close()
	log.Println("Connected to Temporal")

	// Start Temporal worker for DynamicWorkflow
	w := worker.New(tc, shared.PlatformTaskQueue, worker.Options{})
	w.RegisterWorkflowWithOptions(dynamicwf.DynamicWorkflow, workflow.RegisterOptions{
		Name: "DynamicWorkflow",
	})
	go func() {
		if err := w.Run(worker.InterruptCh()); err != nil {
			log.Fatalf("Worker failed: %v", err)
		}
	}()
	log.Println("Temporal worker started on", shared.PlatformTaskQueue)

	// Start HTTP server
	router := api.NewRouter(db, tc, cfg.StaticDir)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down...")
	w.Stop()
}
