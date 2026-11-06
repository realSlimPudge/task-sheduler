package service

import (
	"context"
	"encoding/json"
	"fmt"
	"task-scheduler/internal/apperrors"
	"task-scheduler/internal/model"
)

type TaskRepository interface {
	GetTasks(ctx context.Context) ([]model.Task, error)
	GetTaskByID(ctx context.Context, id int) (*model.Task, error)
	GetStats(ctx context.Context) (*model.TaskStats, error)
	CreateTask(ctx context.Context, taskType model.TaskType, payload json.RawMessage) (*model.Task, error)
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{
		repo: repo,
	}
}

func (s *TaskService) GetTasks(ctx context.Context) ([]model.Task, error) {
	return s.repo.GetTasks(ctx)
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int) (*model.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrValidation, apperrors.ErrInvalidID)
	}

	task, err := s.repo.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetStats(ctx context.Context) (*model.TaskStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *TaskService) CreateTask(ctx context.Context, taskType model.TaskType, payload json.RawMessage) (*model.Task, error) {
	if err := taskType.IsValid(); err != nil {
		return nil, err
	}

	if err := model.ValidatePayload(taskType, payload); err != nil {
		return nil, err
	}

	created, err := s.repo.CreateTask(ctx, taskType, payload)
	if err != nil {
		return nil, err
	}
	return created, nil
}
