package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/seldomhappy/vibe_architecture/internal/domain/models"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/usecase/task"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	useCase *task.UseCase
	log     logger.ILogger
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(useCase *task.UseCase, log logger.ILogger) *TaskHandler {
	return &TaskHandler{
		useCase: useCase,
		log:     log,
	}
}

// HandleTasks handles GET /tasks and POST /tasks
func (h *TaskHandler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	default:
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleTaskByID handles GET /tasks/{id}, PUT /tasks/{id}, DELETE /tasks/{id}
func (h *TaskHandler) HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	parts := strings.Split(path, "/")
	
	if len(parts) == 0 || parts[0] == "" {
		h.respondError(w, http.StatusBadRequest, "Task ID required")
		return
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, id)
	case http.MethodPut:
		h.updateTask(w, r, id)
	case http.MethodDelete:
		h.deleteTask(w, r, id)
	default:
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		h.respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	task, err := h.useCase.CreateTask(r.Context(), req)
	if err != nil {
		h.log.Error("Failed to create task: %v", err)
		h.respondError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	h.respondJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 10
	}

	tasks, err := h.useCase.ListTasks(r.Context(), limit, offset)
	if err != nil {
		h.log.Error("Failed to list tasks: %v", err)
		h.respondError(w, http.StatusInternalServerError, "Failed to list tasks")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"tasks":  tasks,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *TaskHandler) getTask(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	task, err := h.useCase.GetTask(r.Context(), id)
	if err != nil {
		h.log.Error("Failed to get task: %v", err)
		h.respondError(w, http.StatusNotFound, "Task not found")
		return
	}

	h.respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) updateTask(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	task, err := h.useCase.UpdateTask(r.Context(), id, req)
	if err != nil {
		h.log.Error("Failed to update task: %v", err)
		h.respondError(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	h.respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) deleteTask(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if err := h.useCase.DeleteTask(r.Context(), id); err != nil {
		h.log.Error("Failed to delete task: %v", err)
		h.respondError(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error("Failed to encode response: %v", err)
	}
}

func (h *TaskHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
