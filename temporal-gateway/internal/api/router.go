package api

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"go.temporal.io/sdk/client"
	"temporal-gateway/internal/store"
)

func NewRouter(s *store.Store, tc client.Client, staticDir string) http.Handler {
	mux := http.NewServeMux()

	defs := NewDefinitionsHandler(s)
	exec := NewExecutionHandler(s, tc)
	reg := NewRegistryHandler(s)

	// Definition CRUD
	mux.HandleFunc("GET /api/definitions", defs.List)
	mux.HandleFunc("GET /api/definitions/{id}", defs.Get)
	mux.HandleFunc("POST /api/definitions", defs.Create)
	mux.HandleFunc("PUT /api/definitions/{id}", defs.Update)
	mux.HandleFunc("DELETE /api/definitions/{id}", defs.Delete)

	// Workflow execution
	mux.HandleFunc("POST /api/workflows/start", exec.Start)
	mux.HandleFunc("GET /api/workflows/{id}/status", exec.Status)

	// Activity registry
	mux.HandleFunc("POST /api/activities/register", reg.Register)
	mux.HandleFunc("GET /api/activities", reg.List)
	mux.HandleFunc("DELETE /api/activities/service/{name}", reg.Deregister)

	// Serve Vue SPA static files
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		mux.Handle("GET /", spaHandler(staticDir))
	}

	return corsMiddleware(mux)
}

// spaHandler serves static files and falls back to index.html for Vue Router.
func spaHandler(staticDir string) http.Handler {
	fileServer := http.FileServer(http.Dir(staticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := filepath.Join(staticDir, r.URL.Path)
		if _, err := fs.Stat(os.DirFS(staticDir), r.URL.Path[1:]); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fall back to index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(path[:0], staticDir, "index.html"))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
