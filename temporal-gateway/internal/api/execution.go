package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"temporal-gateway/internal/store"
	"temporal-gateway/internal/workflow"
	"temporal-shared"
)

type ExecutionHandler struct {
	store          *store.Store
	temporalClient client.Client
}

func NewExecutionHandler(s *store.Store, tc client.Client) *ExecutionHandler {
	return &ExecutionHandler{store: s, temporalClient: tc}
}

type StartRequest struct {
	DefinitionID string                 `json:"definition_id"`
	Input        map[string]interface{} `json:"input"`
}

type StartResponse struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}

func (h *ExecutionHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	def, err := h.store.Get(req.DefinitionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if def == nil {
		writeError(w, http.StatusNotFound, "definition not found")
		return
	}

	workflowID := fmt.Sprintf("dynamic-%s-%s", def.Name, uuid.New().String()[:8])

	dwInput := workflow.DynamicWorkflowInput{
		Definition: *def,
		Input:      req.Input,
	}

	run, err := h.temporalClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: shared.PlatformTaskQueue,
		},
		"DynamicWorkflow",
		dwInput,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start workflow: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, StartResponse{
		WorkflowID: run.GetID(),
		RunID:      run.GetRunID(),
	})
}

type StatusResponse struct {
	WorkflowID string      `json:"workflow_id"`
	RunID      string      `json:"run_id"`
	Status     string      `json:"status"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func (h *ExecutionHandler) Status(w http.ResponseWriter, r *http.Request) {
	workflowID := r.PathValue("id")

	desc, err := h.temporalClient.DescribeWorkflowExecution(context.Background(), workflowID, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to describe workflow: "+err.Error())
		return
	}

	status := desc.WorkflowExecutionInfo.Status.String()

	resp := StatusResponse{
		WorkflowID: workflowID,
		RunID:      desc.WorkflowExecutionInfo.Execution.RunId,
		Status:     status,
	}

	// If completed or failed, try to get the result / error
	if status == "Completed" || status == "Failed" {
		run := h.temporalClient.GetWorkflow(context.Background(), workflowID, "")
		var result map[string]interface{}
		if err := run.Get(context.Background(), &result); err != nil {
			resp.Error = err.Error()
		} else {
			resp.Result = result
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
