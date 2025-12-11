package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/seldomhappy/vibe_architecture/internal/domain"
	"github.com/seldomhappy/vibe_architecture/internal/usecase/task"
	"github.com/seldomhappy/vibe_architecture/logger"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	useCase task.UseCase
	logger  logger.ILogger
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(uc task.UseCase, log logger.ILogger) *TaskHandler {
	return &TaskHandler{
		useCase: uc,
		logger:  log,
	}
}

// CreateTaskRequest represents a request to create a task
type CreateTaskRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Priority    domain.Priority `json:"priority"`
	CreatedBy   int64           `json:"created_by"`
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	Status      *domain.TaskStatus  `json:"status,omitempty"`
	Priority    *domain.Priority    `json:"priority,omitempty"`
}

// AssignTaskRequest represents a request to assign a task
type AssignTaskRequest struct {
	UserID int64 `json:"user_id"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateTask handles POST /tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validateCreateTaskRequest(req); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := task.CreateTaskInput{
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		CreatedBy:   req.CreatedBy,
	}

	createdTask, err := h.useCase.CreateTask(r.Context(), input)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, createdTask)
}

// GetTask handles GET /tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := h.extractIDFromPath(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.useCase.GetTask(r.Context(), id)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, task)
}

// ListTasks handles GET /tasks
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	filter := task.ListTasksFilter{
		Limit:  50,
		Offset: 0,
	}

	if status := query.Get("status"); status != "" {
		s := domain.TaskStatus(status)
		filter.Status = &s
	}

	if priority := query.Get("priority"); priority != "" {
		p := domain.Priority(priority)
		filter.Priority = &p
	}

	if assignedTo := query.Get("assigned_to"); assignedTo != "" {
		id, err := strconv.ParseInt(assignedTo, 10, 64)
		if err == nil {
			filter.AssignedTo = &id
		}
	}

	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	if offset := query.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	tasks, err := h.useCase.ListTasks(r.Context(), filter)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, tasks)
}

// UpdateTask handles PUT /tasks/{id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := h.extractIDFromPath(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := task.UpdateTaskInput{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
	}

	updatedTask, err := h.useCase.UpdateTask(r.Context(), id, input)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, updatedTask)
}

// DeleteTask handles DELETE /tasks/{id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := h.extractIDFromPath(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.useCase.DeleteTask(r.Context(), id); err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignTask handles POST /tasks/{id}/assign
func (h *TaskHandler) AssignTask(w http.ResponseWriter, r *http.Request) {
	id, err := h.extractIDFromPath(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req AssignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID <= 0 {
		h.respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.useCase.AssignTask(r.Context(), id, req.UserID); err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "task assigned successfully"})
}

// CompleteTask handles POST /tasks/{id}/complete
func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := h.extractIDFromPath(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.useCase.CompleteTask(r.Context(), id); err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "task completed successfully"})
}

// Health handles GET /health
func (h *TaskHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Helper methods

func (h *TaskHandler) extractIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid path")
	}
	
	// Find the ID after /tasks/
	for i, part := range parts {
		if part == "tasks" && i+1 < len(parts) {
			return strconv.ParseInt(parts[i+1], 10, 64)
		}
	}
	
	return 0, fmt.Errorf("task id not found in path")
}

func (h *TaskHandler) validateCreateTaskRequest(req CreateTaskRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("name is too long (max 255 characters)")
	}
	if req.Priority == "" {
		return fmt.Errorf("priority is required")
	}
	if !req.Priority.IsValid() {
		return fmt.Errorf("invalid priority (allowed: low, medium, high)")
	}
	if req.CreatedBy <= 0 {
		return fmt.Errorf("created_by is required")
	}
	return nil
}

func (h *TaskHandler) handleUseCaseError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrTaskNotFound:
		h.respondError(w, http.StatusNotFound, err.Error())
	case domain.ErrEmptyTaskName, domain.ErrTaskNameTooLong, domain.ErrInvalidInput:
		h.respondError(w, http.StatusBadRequest, err.Error())
	case domain.ErrUnauthorized:
		h.respondError(w, http.StatusUnauthorized, err.Error())
	default:
		h.respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *TaskHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response: %v", err)
	}
}

func (h *TaskHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{Error: message})
}
