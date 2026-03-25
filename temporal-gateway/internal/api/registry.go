package api

import (
	"encoding/json"
	"net/http"

	"temporal-gateway/internal/store"
	"temporal-shared"
)

type RegistryHandler struct {
	store *store.Store
}

func NewRegistryHandler(s *store.Store) *RegistryHandler {
	return &RegistryHandler{store: s}
}

func (h *RegistryHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req shared.RegisterActivitiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ServiceName == "" || len(req.Activities) == 0 {
		writeError(w, http.StatusBadRequest, "service_name and activities are required")
		return
	}

	// Stamp service name onto each activity
	for i := range req.Activities {
		req.Activities[i].ServiceName = req.ServiceName
	}

	if err := h.store.BulkUpsertActivities(req.Activities); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to register activities")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

type activityCategoryResponse struct {
	Name       string                      `json:"name"`
	Activities []shared.ActivityRegistration `json:"activities"`
}

func (h *RegistryHandler) List(w http.ResponseWriter, r *http.Request) {
	activities, err := h.store.ListActivities()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list activities")
		return
	}

	// Group by category
	catMap := make(map[string][]shared.ActivityRegistration)
	catOrder := []string{}
	for _, a := range activities {
		cat := a.Category
		if cat == "" {
			cat = "General"
		}
		if _, exists := catMap[cat]; !exists {
			catOrder = append(catOrder, cat)
		}
		catMap[cat] = append(catMap[cat], a)
	}

	categories := make([]activityCategoryResponse, 0, len(catOrder))
	for _, name := range catOrder {
		categories = append(categories, activityCategoryResponse{
			Name:       name,
			Activities: catMap[name],
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"categories": categories})
}

func (h *RegistryHandler) Deregister(w http.ResponseWriter, r *http.Request) {
	serviceName := r.PathValue("name")
	if serviceName == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}
	if err := h.store.DeleteActivitiesByService(serviceName); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to deregister activities")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deregistered"})
}
