package handler

import (
	"encoding/json"
	"task-scheduler/internal/model"
	"time"
)

type TaskRes struct {
	ID        int             `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
	Attempts  int             `json:"attempts"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type CreateTaskReq struct {
	Type    model.TaskType  `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type TaskStatsRes struct {
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}
