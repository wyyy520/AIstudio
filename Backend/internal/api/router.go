package api

import (
	"encoding/json"
	"net/http"

	"github.com/aistudio/backend/internal/workflow"
)

type Handler struct {
	engine *workflow.Engine
}

func NewHandler(engine *workflow.Engine) *Handler {
	return &Handler{engine: engine}
}

func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/workflow/run", h.handleRunWorkflow)
	mux.HandleFunc("/api/workflow/task/", h.handleGetTask)
	mux.HandleFunc("/api/workflow/nodes", h.handleListNodeTypes)
	mux.HandleFunc("/api/health", h.handleHealth)
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleRunWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Code:    -1,
			Message: "method not allowed, use POST",
		})
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Code:    -1,
			Message: "invalid JSON body: " + err.Error(),
		})
		return
	}

	workflowJSON, err := json.Marshal(body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Code:    -1,
			Message: "failed to re-encode JSON: " + err.Error(),
		})
		return
	}

	result, err := h.engine.Run(r.Context(), workflowJSON)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Code:    -1,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"task_id": result.TaskID,
			"status":  result.Status,
		},
	})
}

func (h *Handler) handleListNodeTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Code:    -1,
			Message: "method not allowed, use GET",
		})
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	nodes := h.engine.Registry().List()
	type nodeInfo struct {
		Type        string          `json:"type"`
		Plugin      string          `json:"plugin"`
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Inputs      []workflow.Port `json:"inputs"`
		Outputs     []workflow.Port `json:"outputs"`
	}
	items := make([]nodeInfo, 0, len(nodes))
	for _, def := range nodes {
		items = append(items, nodeInfo{
			Type:        def.Type,
			Plugin:      def.Plugin,
			Name:        def.Name,
			Description: def.Description,
			Inputs:      def.Inputs,
			Outputs:     def.Outputs,
		})
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    items,
	})
}

func (h *Handler) handleGetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Code:    -1,
			Message: "method not allowed, use GET",
		})
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	taskID := r.URL.Path[len("/api/workflow/task/"):]
	if taskID == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Code:    -1,
			Message: "task_id is required",
		})
		return
	}

	result, ok := h.engine.GetTask(taskID)
	if !ok {
		writeJSON(w, http.StatusNotFound, APIResponse{
			Code:    -1,
			Message: "task not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "AIStudio Workflow Engine is running",
	})
}
