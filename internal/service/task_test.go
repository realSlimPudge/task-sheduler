package service

import (
	"context"
	"errors"
	"task-scheduler/internal/apperrors"
	"task-scheduler/internal/model"
	"testing"
)

type mockRepository struct {
	TaskRepository

	taskToReturn *model.Task
	errToReturn  error
}

func (m *mockRepository) GetTaskByID(ctx context.Context, id int) (*model.Task, error) {
	return m.taskToReturn, m.errToReturn
}

func TestInvalidID(t *testing.T) {
	svc := NewTaskService(&mockRepository{})

	tests := []struct {
		name string
		ID   int
	}{
		{
			name: "zero ID",
			ID:   0,
		},
		{
			name: "negative ID",
			ID:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetTaskByID(context.Background(), tt.ID)

			if !errors.Is(err, apperrors.ErrValidation) {
				t.Errorf("Invalid task ID error = %v", err)
			}
		})
	}
}

func TestValidID(t *testing.T) {
	expectedTask := &model.Task{
		ID:   1,
		Type: model.TypeEmail,
	}
	mock := &mockRepository{
		taskToReturn: expectedTask,
		errToReturn:  nil,
	}
	svc := NewTaskService(mock)

	tests := []struct {
		name string
		ID   int
	}{
		{
			name: "correct ID",
			ID:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := svc.GetTaskByID(context.Background(), tt.ID)

			if err != nil {
				t.Errorf("unexpected err, error: %v", err)
			}

			if task == nil {
				t.Fatal("task is nil")
			}

			if task.ID != expectedTask.ID {
				t.Errorf("got ID %d, want %d", task.ID, expectedTask.ID)
			}
		})
	}
}
