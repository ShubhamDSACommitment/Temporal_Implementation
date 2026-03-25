package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// RegisterWithGateway POSTs the registration request to the gateway with retries.
func RegisterWithGateway(gatewayURL string, req RegisterActivitiesRequest) error {

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal registration: %w", err)
	}

	var lastErr error
	for i := range 10 {
		resp, err := http.Post(gatewayURL+"/api/activities/register", "application/json", bytes.NewReader(body))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				slog.Info("Registered activities with gateway", "service", req.ServiceName, "count", len(req.Activities))
				return nil
			}
			lastErr = fmt.Errorf("gateway returned status %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		wait := time.Duration(min(i+1, 5)) * time.Second
		slog.Warn("Failed to register with gateway, retrying...", "attempt", i+1, "error", lastErr, "retry_in", wait)
		time.Sleep(wait)
	}
	return fmt.Errorf("register with gateway after 10 attempts: %w", lastErr)
}

// StartHeartbeat re-registers activities at the given interval to keep them fresh.
func StartHeartbeat(gatewayURL string, req RegisterActivitiesRequest, interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			slog.Info("Heartbeat stopped", "service", req.ServiceName)
			return
		case <-ticker.C:
			if err := RegisterWithGateway(gatewayURL, req); err != nil {
				slog.Error("Heartbeat registration failed", "service", req.ServiceName, "error", err)
			}
		}
	}
}
