package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"temporal-gateway/internal/store"
	"temporal-shared"
)

type DefinitionsHandler struct {
	store *store.Store
}

func NewDefinitionsHandler(s *store.Store) *DefinitionsHandler {
	return &DefinitionsHandler{store: s}
}

func (h *DefinitionsHandler) List(w http.ResponseWriter, r *http.Request) {
	defs, err := h.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if defs == nil {
		defs = []shared.WorkflowDefinition{}
	}
	writeJSON(w, http.StatusOK, defs)
}

func (h *DefinitionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	def, err := h.store.Get(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if def == nil {
		writeError(w, http.StatusNotFound, "definition not found")
		return
	}
	writeJSON(w, http.StatusOK, def)
}

func (h *DefinitionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var def shared.WorkflowDefinition
	if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if def.ID == "" {
		def.ID = uuid.New().String()
	}
	if def.Version == 0 {
		def.Version = 1
	}
	if err := h.store.Create(&def); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, def)
}

func (h *DefinitionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var def shared.WorkflowDefinition
	if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	def.ID = id
	if err := h.store.Update(&def); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, def)
}

func (h *DefinitionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
