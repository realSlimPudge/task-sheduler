package repository

import (
	"context"
	"encoding/json"
	"task-scheduler/internal/model"
	"time"
)

type InMemoryRepo struct {
	tasks []model.Task
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		tasks: []model.Task{{
			ID:        1,
			Type:      model.TypeImage,
			Payload:   json.RawMessage{},
			Status:    model.StatusPending,
			Result:    json.RawMessage(nil),
			Error:     "",
			Attempts:  1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}},
	}
}

func (r *InMemoryRepo) GetTasks(ctx context.Context) ([]model.Task, error) {
	return append([]model.Task(nil), r.tasks...), nil
}
