package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"task-scheduler/internal/apperrors"
	"task-scheduler/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.GetTasks(r.Context())
	if err != nil {
		slog.Error("Get tasks", "err:", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	res := make([]TaskRes, 0, len(tasks))

	for _, t := range tasks {
		var task TaskRes

		task.ID = t.ID
		task.Type = string(t.Type)
		task.Payload = t.Payload
		task.Status = string(t.Status)
		task.Result = t.Result
		task.Error = t.Error
		task.Attempts = t.Attempts
		task.CreatedAt = t.CreatedAt
		task.UpdatedAt = t.UpdatedAt

		res = append(res, task)
	}

	respondJSON(w, http.StatusOK, res)
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.service.GetTaskByID(r.Context(), id)

	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Error("Get tasks", "err:", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	res := TaskRes{
		ID:        task.ID,
		Type:      string(task.Type),
		Payload:   task.Payload,
		Status:    string(task.Status),
		Result:    task.Result,
		Error:     task.Error,
		Attempts:  task.Attempts,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}

	respondJSON(w, http.StatusOK, res)
}

func (h *TaskHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		slog.Error("get stats error", "err", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	statsRes := TaskStatsRes{
		Pending:   stats.Pending,
		Running:   stats.Running,
		Completed: stats.Completed,
		Failed:    stats.Failed,
	}

	respondJSON(w, http.StatusOK, statsRes)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.service.CreateTask(r.Context(), req.Type, req.Payload)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		slog.Error("Create task:", "err", err.Error())
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	res := TaskRes{
		ID:        created.ID,
		Type:      string(created.Type),
		Payload:   created.Payload,
		Status:    string(created.Status),
		Result:    created.Result,
		Error:     created.Error,
		Attempts:  created.Attempts,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}

	respondJSON(w, http.StatusCreated, res)
}
